package capturearchive

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

var archiveTime = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

type Result struct {
	Path                           string `json:"path"`
	SHA256                         string `json:"sha256"`
	Bytes                          int64  `json:"bytes"`
	Entries                        int    `json:"entries"`
	ContainsDecryptedResourcePacks bool   `json:"contains_decrypted_resource_packs"`
}

func Export(root, output string) (Result, error) {
	verification, err := capture.Verify(root)
	if err != nil {
		return Result{}, err
	}
	if len(verification.Issues) != 0 {
		return Result{}, fmt.Errorf("capture verification found %d issue(s)", len(verification.Issues))
	}
	if verification.Manifest.Status == "open" {
		return Result{}, errors.New("cannot export a capture whose manifest status is open")
	}
	if strings.TrimSpace(output) == "" {
		return Result{}, errors.New("output path is empty")
	}
	inside, err := pathInside(root, output)
	if err != nil {
		return Result{}, err
	}
	if inside {
		return Result{}, errors.New("archive output must be outside the capture directory")
	}
	paths := []string{"manifest.json", verification.Manifest.EventsPath}
	seen := map[string]struct{}{"manifest.json": {}, verification.Manifest.EventsPath: {}}
	containsDecryptedResourcePacks := false
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind == "resource_pack.contents_manifest" || event.Kind == "resource_pack.decrypted_archive" {
			containsDecryptedResourcePacks = true
		}
		if event.Blob == nil {
			return nil
		}
		if _, ok := seen[event.Blob.Path]; !ok {
			seen[event.Blob.Path] = struct{}{}
			paths = append(paths, event.Blob.Path)
		}
		return nil
	}); err != nil {
		return Result{}, err
	}
	sort.Strings(paths[2:])
	if err := os.MkdirAll(filepath.Dir(output), 0o700); err != nil {
		return Result{}, fmt.Errorf("create archive parent directory: %w", err)
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return Result{}, fmt.Errorf("create archive: %w", err)
	}
	removeOutput := true
	defer func() {
		_ = f.Close()
		if removeOutput {
			_ = os.Remove(output)
		}
	}()
	writer := zip.NewWriter(f)
	if err := writer.SetComment(capture.SchemaVersion); err != nil {
		_ = writer.Close()
		return Result{}, fmt.Errorf("set archive comment: %w", err)
	}
	for _, path := range paths {
		if err := addFile(writer, root, path); err != nil {
			_ = writer.Close()
			return Result{}, err
		}
	}
	if err := writer.Close(); err != nil {
		return Result{}, fmt.Errorf("close archive stream: %w", err)
	}
	if err := f.Sync(); err != nil {
		return Result{}, fmt.Errorf("sync archive: %w", err)
	}
	if err := f.Close(); err != nil {
		return Result{}, fmt.Errorf("close archive: %w", err)
	}
	result, err := inspectArchive(output, len(paths))
	if err != nil {
		return Result{}, err
	}
	result.ContainsDecryptedResourcePacks = containsDecryptedResourcePacks
	removeOutput = false
	return result, nil
}

func addFile(writer *zip.Writer, root, relative string) error {
	clean := filepath.Clean(filepath.FromSlash(relative))
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("unsafe archive entry path %q", relative)
	}
	path := filepath.Join(root, clean)
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("resolve capture root: %w", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("resolve archive entry %s: %w", relative, err)
	}
	inside, err := pathInside(resolvedRoot, resolvedPath)
	if err != nil {
		return err
	}
	if !inside {
		return fmt.Errorf("archive entry %s resolves outside the capture directory", relative)
	}
	source, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open archive entry %s: %w", relative, err)
	}
	defer func() { _ = source.Close() }()
	info, err := source.Stat()
	if err != nil {
		return fmt.Errorf("inspect archive entry %s: %w", relative, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("archive entry %s is not a regular file", relative)
	}
	header := &zip.FileHeader{Name: filepath.ToSlash(clean), Method: zip.Store, Modified: archiveTime}
	header.SetMode(0o600)
	destination, err := writer.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("create archive entry %s: %w", relative, err)
	}
	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("write archive entry %s: %w", relative, err)
	}
	return nil
}

func inspectArchive(path string, expectedEntries int) (Result, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return Result{}, fmt.Errorf("reopen archive: %w", err)
	}
	defer func() { _ = reader.Close() }()
	if len(reader.File) != expectedEntries {
		return Result{}, fmt.Errorf("archive contains %d entries, expected %d", len(reader.File), expectedEntries)
	}
	for _, entry := range reader.File {
		if entry.Method != zip.Store {
			return Result{}, fmt.Errorf("archived entry %s uses unexpected compression method %d", entry.Name, entry.Method)
		}
		content, err := entry.Open()
		if err != nil {
			return Result{}, fmt.Errorf("open archived entry %s: %w", entry.Name, err)
		}
		// #nosec G110 -- Export creates only uncompressed entries from already verified local files.
		_, copyErr := io.Copy(io.Discard, content)
		closeErr := content.Close()
		if copyErr != nil || closeErr != nil {
			return Result{}, fmt.Errorf("verify archived entry %s: %w", entry.Name, errors.Join(copyErr, closeErr))
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return Result{}, fmt.Errorf("open completed archive: %w", err)
	}
	defer func() { _ = f.Close() }()
	hash := sha256.New()
	size, err := io.Copy(hash, f)
	if err != nil {
		return Result{}, fmt.Errorf("hash completed archive: %w", err)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return Result{}, fmt.Errorf("resolve archive path: %w", err)
	}
	return Result{Path: absolute, SHA256: hex.EncodeToString(hash.Sum(nil)), Bytes: size, Entries: len(reader.File)}, nil
}

func pathInside(root, candidate string) (bool, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return false, fmt.Errorf("resolve capture root: %w", err)
	}
	absoluteCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return false, fmt.Errorf("resolve archive output: %w", err)
	}
	relative, err := filepath.Rel(absoluteRoot, absoluteCandidate)
	if err != nil {
		return false, fmt.Errorf("compare capture and archive paths: %w", err)
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}
