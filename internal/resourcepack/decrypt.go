package resourcepack

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	contentsHeaderSize      = 256
	maxArchiveEntries       = 100_000
	maxContentsManifests    = 1_024
	maxContentsManifestSize = 16 << 20
	maxContentsPlaintext    = 64 << 20
	maxDecryptedPayloadSize = 1 << 30
)

var (
	contentsMagic = [4]byte{0xfc, 0xb9, 0xcf, 0x9b}
	archiveTime   = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
)

type Report struct {
	Algorithm             string            `json:"algorithm"`
	Authenticated         bool              `json:"authenticated"`
	ContentID             string            `json:"content_id"`
	InputEntries          int               `json:"input_entries"`
	OutputEntries         int               `json:"output_entries"`
	DecryptedFiles        int               `json:"decrypted_files"`
	CopiedFiles           int               `json:"copied_files"`
	Directories           int               `json:"directories"`
	DecryptedPayloadBytes int64             `json:"decrypted_payload_bytes"`
	Contents              []ContentsSummary `json:"contents"`
}

type ContentsSummary struct {
	Path           string `json:"path"`
	Entries        int    `json:"entries"`
	EncryptedFiles int    `json:"encrypted_files"`
	PlainFiles     int    `json:"plain_files"`
}

type ContentsManifest struct {
	ContentsSummary
	Plaintext []byte `json:"-"`
}

type Result struct {
	Report    Report
	Manifests []ContentsManifest
}

type contentsDocument struct {
	Content []contentsEntry `json:"content"`
}

type contentsEntry struct {
	Path string `json:"path"`
	Key  string `json:"key"`
}

type cfb8Decrypter struct {
	block    cipher.Block
	register [aes.BlockSize]byte
	scratch  [aes.BlockSize]byte
}

// DecryptArchive decodes the documented pack format into a derived ZIP. The
// caller owns key provenance and must preserve the source archive separately.
// The shipped production caller supplies only a key received on its current
// upstream Bedrock connection.
func DecryptArchive(ctx context.Context, source io.ReaderAt, archiveSize int64, contentKey string, destination io.Writer) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if source == nil {
		return Result{}, errors.New("resource pack source is nil")
	}
	if destination == nil {
		return Result{}, errors.New("resource pack destination is nil")
	}
	if archiveSize < 0 {
		return Result{}, errors.New("resource pack archive size is negative")
	}
	key := []byte(contentKey)
	if len(key) != 32 {
		return Result{}, fmt.Errorf("resource pack content key must be 32 bytes, got %d", len(key))
	}
	reader, err := zip.NewReader(source, archiveSize)
	if err != nil {
		return Result{}, fmt.Errorf("open resource pack archive: %w", err)
	}
	if len(reader.File) > maxArchiveEntries {
		return Result{}, fmt.Errorf("resource pack has %d entries, limit is %d", len(reader.File), maxArchiveEntries)
	}

	files := make(map[string]*zip.File, len(reader.File))
	orderedNames := make([]string, 0, len(reader.File))
	manifestPaths := make([]string, 0, 1)
	for _, file := range reader.File {
		name, err := cleanArchivePath(file.Name, file.FileInfo().IsDir())
		if err != nil {
			return Result{}, err
		}
		if _, exists := files[name]; exists {
			return Result{}, fmt.Errorf("resource pack contains duplicate entry %q", name)
		}
		files[name] = file
		orderedNames = append(orderedNames, name)
		if !file.FileInfo().IsDir() && path.Base(name) == "manifest.json" {
			manifestPaths = append(manifestPaths, name)
		}
	}
	if len(manifestPaths) == 0 {
		return Result{}, errors.New("resource pack manifest.json was not found")
	}
	sort.Slice(manifestPaths, func(i, j int) bool {
		leftDepth := strings.Count(manifestPaths[i], "/")
		rightDepth := strings.Count(manifestPaths[j], "/")
		if leftDepth != rightDepth {
			return leftDepth < rightDepth
		}
		return manifestPaths[i] < manifestPaths[j]
	})
	rootManifest := manifestPaths[0]
	root := path.Dir(rootManifest)
	if root == "." {
		root = ""
	}
	rootContents := joinArchivePath(root, "contents.json")
	if _, ok := files[rootContents]; !ok {
		return Result{}, fmt.Errorf("encrypted resource pack does not contain %s", rootContents)
	}

	contentsPaths := []string{rootContents}
	subpackPrefix := joinArchivePath(root, "subpacks") + "/"
	for name, file := range files {
		if file.FileInfo().IsDir() || name == rootContents || !strings.HasPrefix(name, subpackPrefix) {
			continue
		}
		relative := strings.TrimPrefix(name, subpackPrefix)
		parts := strings.Split(relative, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] == "contents.json" {
			contentsPaths = append(contentsPaths, name)
		}
	}
	sort.Strings(contentsPaths[1:])
	if len(contentsPaths) > maxContentsManifests {
		return Result{}, fmt.Errorf("resource pack has %d contents manifests, limit is %d", len(contentsPaths), maxContentsManifests)
	}

	declarations := make(map[string]string)
	contentsSet := make(map[string]struct{}, len(contentsPaths))
	totalDeclarations := 0
	remainingContentsPlaintext := int64(maxContentsPlaintext)
	result := Result{Report: Report{
		Algorithm:     "AES-256-CFB8",
		Authenticated: false,
		InputEntries:  len(reader.File),
	}}
	for _, manifestPath := range contentsPaths {
		manifest, contentID, entries, err := decryptContentsManifest(ctx, files[manifestPath], manifestPath, key)
		if err != nil {
			return Result{}, err
		}
		if int64(len(manifest.Plaintext)) > remainingContentsPlaintext {
			return Result{}, fmt.Errorf("decrypted contents manifests exceed the %d-byte plaintext limit", maxContentsPlaintext)
		}
		remainingContentsPlaintext -= int64(len(manifest.Plaintext))
		if len(entries) > maxArchiveEntries-totalDeclarations {
			return Result{}, fmt.Errorf("resource pack contents declare more than %d total entries", maxArchiveEntries)
		}
		totalDeclarations += len(entries)
		if result.Report.ContentID == "" {
			result.Report.ContentID = contentID
		} else if result.Report.ContentID != contentID {
			return Result{}, fmt.Errorf("contents manifest %s has content ID %q, expected %q", manifestPath, contentID, result.Report.ContentID)
		}
		contentsSet[manifestPath] = struct{}{}
		base := path.Dir(manifestPath)
		if base == "." {
			base = ""
		}
		for _, entry := range entries {
			relative, err := cleanContentPath(entry.Path)
			if err != nil {
				return Result{}, fmt.Errorf("contents manifest %s: %w", manifestPath, err)
			}
			fullPath := joinArchivePath(base, relative)
			if _, exists := declarations[fullPath]; exists {
				return Result{}, fmt.Errorf("resource pack contents declare %s more than once", fullPath)
			}
			if entry.Key != "" && len([]byte(entry.Key)) != 32 {
				return Result{}, fmt.Errorf("resource pack entry %s has a %d-byte key, expected 32", fullPath, len([]byte(entry.Key)))
			}
			if _, isContents := contentsSet[fullPath]; !isContents {
				if _, exists := files[fullPath]; !exists {
					return Result{}, fmt.Errorf("resource pack contents reference missing entry %s", fullPath)
				}
			}
			declarations[fullPath] = entry.Key
			if entry.Key == "" {
				manifest.PlainFiles++
			} else {
				manifest.EncryptedFiles++
			}
		}
		result.Manifests = append(result.Manifests, manifest)
		result.Report.Contents = append(result.Report.Contents, manifest.ContentsSummary)
	}

	writer := zip.NewWriter(destination)
	if err := writer.SetComment("BedrockDebugProxy decrypted resource pack"); err != nil {
		_ = writer.Close()
		return Result{}, fmt.Errorf("set decrypted archive comment: %w", err)
	}
	remaining := int64(maxDecryptedPayloadSize)
	for _, name := range orderedNames {
		if _, skip := contentsSet[name]; skip {
			continue
		}
		file := files[name]
		mode := file.Mode()
		if mode&os.ModeSymlink != 0 {
			_ = writer.Close()
			return Result{}, fmt.Errorf("resource pack entry %s is a symbolic link", name)
		}
		directory := file.FileInfo().IsDir()
		if !directory && !mode.IsRegular() {
			_ = writer.Close()
			return Result{}, fmt.Errorf("resource pack entry %s is not a regular file", name)
		}
		header := &zip.FileHeader{Name: name, Method: zip.Store, Modified: archiveTime}
		if directory {
			header.SetMode(0o700 | os.ModeDir)
		} else {
			header.SetMode(0o600)
		}
		destinationEntry, err := writer.CreateHeader(header)
		if err != nil {
			_ = writer.Close()
			return Result{}, fmt.Errorf("create decrypted archive entry %s: %w", name, err)
		}
		result.Report.OutputEntries++
		if directory {
			result.Report.Directories++
			continue
		}
		if file.UncompressedSize64 > uint64(maxDecryptedPayloadSize) {
			_ = writer.Close()
			return Result{}, fmt.Errorf("resource pack entry %s exceeds the %d-byte decrypted payload limit", name, maxDecryptedPayloadSize)
		}
		sourceEntry, err := file.Open()
		if err != nil {
			_ = writer.Close()
			return Result{}, fmt.Errorf("open resource pack entry %s: %w", name, err)
		}
		input := io.Reader(sourceEntry)
		if entryKey := declarations[name]; entryKey != "" {
			stream, streamErr := newCFB8Decrypter([]byte(entryKey), []byte(entryKey)[:aes.BlockSize])
			if streamErr != nil {
				_ = sourceEntry.Close()
				_ = writer.Close()
				return Result{}, fmt.Errorf("create decryptor for %s: %w", name, streamErr)
			}
			input = &cipher.StreamReader{S: stream, R: input}
			result.Report.DecryptedFiles++
		} else {
			result.Report.CopiedFiles++
		}
		written, copyErr := copyWithBudget(ctx, destinationEntry, input, &remaining)
		closeErr := sourceEntry.Close()
		if copyErr != nil || closeErr != nil {
			_ = writer.Close()
			return Result{}, fmt.Errorf("write decrypted archive entry %s: %w", name, errors.Join(copyErr, closeErr))
		}
		expectedSize := int64(file.UncompressedSize64)
		if written != expectedSize {
			_ = writer.Close()
			return Result{}, fmt.Errorf("resource pack entry %s produced %d bytes, expected %d", name, written, file.UncompressedSize64)
		}
		result.Report.DecryptedPayloadBytes += written
	}
	if err := writer.Close(); err != nil {
		return Result{}, fmt.Errorf("close decrypted resource pack archive: %w", err)
	}
	return result, nil
}

func decryptContentsManifest(ctx context.Context, file *zip.File, manifestPath string, key []byte) (ContentsManifest, string, []contentsEntry, error) {
	data, err := readLimitedEntry(ctx, file)
	if err != nil {
		return ContentsManifest{}, "", nil, fmt.Errorf("read contents manifest %s: %w", manifestPath, err)
	}
	if len(data) < contentsHeaderSize {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s is %d bytes, shorter than its %d-byte header", manifestPath, len(data), contentsHeaderSize)
	}
	header := data[:contentsHeaderSize]
	if version := binary.LittleEndian.Uint32(header[:4]); version != 0 {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s uses unsupported header version %d", manifestPath, version)
	}
	if !bytes.Equal(header[4:8], contentsMagic[:]) {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s has magic %x, expected %x", manifestPath, header[4:8], contentsMagic)
	}
	if !allZero(header[8:16]) {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s has non-zero reserved header bytes", manifestPath)
	}
	contentIDLength := int(header[16])
	contentIDEnd := 17 + contentIDLength
	if contentIDLength == 0 || contentIDEnd > contentsHeaderSize {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s has invalid content ID length %d", manifestPath, contentIDLength)
	}
	contentID := string(header[17:contentIDEnd])
	if !utf8.ValidString(contentID) {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s has a non-UTF-8 content ID", manifestPath)
	}
	if !allZero(header[contentIDEnd:]) {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s has non-zero header padding", manifestPath)
	}
	stream, err := newCFB8Decrypter(key, key[:aes.BlockSize])
	if err != nil {
		return ContentsManifest{}, "", nil, err
	}
	plaintext := make([]byte, len(data)-contentsHeaderSize)
	stream.XORKeyStream(plaintext, data[contentsHeaderSize:])
	plaintext = bytes.TrimRight(plaintext, "\x00 \t\r\n")
	if len(plaintext) == 0 || !json.Valid(plaintext) {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s did not decrypt to valid JSON", manifestPath)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(plaintext, &envelope); err != nil {
		return ContentsManifest{}, "", nil, fmt.Errorf("decode contents manifest %s: %w", manifestPath, err)
	}
	rawContent, ok := envelope["content"]
	trimmedContent := bytes.TrimSpace(rawContent)
	if !ok || len(trimmedContent) == 0 || trimmedContent[0] != '[' {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s has no content array", manifestPath)
	}
	var document contentsDocument
	if err := json.Unmarshal(rawContent, &document.Content); err != nil {
		return ContentsManifest{}, "", nil, fmt.Errorf("decode contents entries in %s: %w", manifestPath, err)
	}
	if len(document.Content) > maxArchiveEntries {
		return ContentsManifest{}, "", nil, fmt.Errorf("contents manifest %s has %d entries, limit is %d", manifestPath, len(document.Content), maxArchiveEntries)
	}
	manifest := ContentsManifest{
		ContentsSummary: ContentsSummary{Path: manifestPath, Entries: len(document.Content)},
		Plaintext:       append([]byte(nil), plaintext...),
	}
	return manifest, contentID, document.Content, nil
}

func newCFB8Decrypter(key, iv []byte) (*cfb8Decrypter, error) {
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("AES-CFB8 IV must be %d bytes, got %d", aes.BlockSize, len(iv))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := &cfb8Decrypter{block: block}
	copy(stream.register[:], iv)
	return stream, nil
}

func (s *cfb8Decrypter) XORKeyStream(destination, source []byte) {
	if len(destination) < len(source) {
		panic("resourcepack: output smaller than input")
	}
	for index, ciphertext := range source {
		s.block.Encrypt(s.scratch[:], s.register[:])
		destination[index] = ciphertext ^ s.scratch[0]
		copy(s.register[:aes.BlockSize-1], s.register[1:])
		s.register[aes.BlockSize-1] = ciphertext
	}
}

func readLimitedEntry(ctx context.Context, file *zip.File) ([]byte, error) {
	if file.UncompressedSize64 > maxContentsManifestSize {
		return nil, fmt.Errorf("entry is %d bytes, limit is %d", file.UncompressedSize64, maxContentsManifestSize)
	}
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(io.LimitReader(&contextReader{ctx: ctx, source: reader}, maxContentsManifestSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxContentsManifestSize {
		return nil, fmt.Errorf("entry exceeds %d bytes", maxContentsManifestSize)
	}
	return data, nil
}

func copyWithBudget(ctx context.Context, destination io.Writer, source io.Reader, remaining *int64) (int64, error) {
	if *remaining < 0 {
		return 0, errors.New("decrypted resource pack payload limit exceeded")
	}
	limited := &io.LimitedReader{R: &contextReader{ctx: ctx, source: source}, N: *remaining + 1}
	written, err := io.Copy(destination, limited)
	if err != nil {
		return written, err
	}
	if written > *remaining {
		return written, fmt.Errorf("decrypted resource pack payload exceeds %d bytes", maxDecryptedPayloadSize)
	}
	*remaining -= written
	return written, nil
}

type contextReader struct {
	ctx    context.Context
	source io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.source.Read(buffer)
	}
}

func cleanArchivePath(name string, directory bool) (string, error) {
	if name == "" || !utf8.ValidString(name) || strings.Contains(name, "\\") || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("resource pack contains unsafe entry path %q", name)
	}
	clean := path.Clean(name)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("resource pack contains unsafe entry path %q", name)
	}
	if directory {
		clean += "/"
	}
	if clean != name {
		return "", fmt.Errorf("resource pack contains non-canonical entry path %q", name)
	}
	return clean, nil
}

func cleanContentPath(name string) (string, error) {
	if name == "" || !utf8.ValidString(name) || strings.Contains(name, "\\") || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("unsafe content path %q", name)
	}
	directory := strings.HasSuffix(name, "/")
	clean := path.Clean(name)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("unsafe content path %q", name)
	}
	if directory {
		clean += "/"
	}
	if clean != name {
		return "", fmt.Errorf("non-canonical content path %q", name)
	}
	return clean, nil
}

func joinArchivePath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "/" + name
}

func allZero(data []byte) bool {
	for _, value := range data {
		if value != 0 {
			return false
		}
	}
	return true
}
