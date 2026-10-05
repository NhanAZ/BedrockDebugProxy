package resourcepack

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

const (
	testContentKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"
	testContentID  = "11111111-1111-1111-1111-111111111111"
)

func TestCFB8DecrypterNISTAES256KnownAnswer(t *testing.T) {
	t.Parallel()

	key := decodeHex(t, "0000000000000000000000000000000000000000000000000000000000000000")
	iv := decodeHex(t, "014730f80ac625fe84f026c60bfd547d")
	ciphertext := decodeHex(t, "5c")
	want := decodeHex(t, "00")

	stream, err := newCFB8Decrypter(key, iv)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]byte, len(ciphertext))
	stream.XORKeyStream(got, ciphertext)
	if !bytes.Equal(got, want) {
		t.Fatalf("decrypt known-answer vector = %x, want %x", got, want)
	}
}

func TestCFB8DecrypterIndependentMultiByteVector(t *testing.T) {
	t.Parallel()

	key := decodeHex(t, "4142434445464748494a4b4c4d4e4f505152535455565758595a313233343536")
	iv := decodeHex(t, "4142434445464748494a4b4c4d4e4f50")
	ciphertext := decodeHex(t, "5ca682b77a6111f54f45d64aad9c96466d28374ad812d916eaaea7dd369f")
	want := decodeHex(t, "426564726f636b446562756750726f787920434642382066697874757265")

	stream, err := newCFB8Decrypter(key, iv)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]byte, len(ciphertext))
	stream.XORKeyStream(got, ciphertext)
	if !bytes.Equal(got, want) {
		t.Fatalf("decrypt independent vector = %x, want %x", got, want)
	}
}

func TestDecryptArchiveRootAndSubpack(t *testing.T) {
	t.Parallel()

	rootKey := "ZYXWVUTSRQPONMLKJIHGFEDCBA654321"
	subpackKey := "1234567890abcdefghijklmnopqrstuv"
	input := buildEncryptedArchive(t, []testArchiveEntry{
		{name: "manifest.json", data: []byte(`{"format_version":2}`)},
		{name: "pack_icon.png", data: []byte("plain icon")},
		{name: "textures/root.txt", data: encryptCFB8(t, []byte(rootKey), []byte("root plaintext"))},
		{name: "plain.txt", data: []byte("root plain file")},
		{name: "contents.json", data: encryptedContents(t, testContentKey, testContentID, []contentsEntry{
			{Path: "textures/root.txt", Key: rootKey},
			{Path: "plain.txt"},
		})},
		{name: "subpacks/hd/file.txt", data: encryptCFB8(t, []byte(subpackKey), []byte("subpack plaintext"))},
		{name: "subpacks/hd/contents.json", data: encryptedContents(t, testContentKey, testContentID, []contentsEntry{
			{Path: "file.txt", Key: subpackKey},
		})},
	})

	var output bytes.Buffer
	result, err := DecryptArchive(context.Background(), bytes.NewReader(input), int64(len(input)), testContentKey, &output)
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.Algorithm != "AES-256-CFB8" || result.Report.Authenticated {
		t.Fatalf("unexpected encryption report: %#v", result.Report)
	}
	if result.Report.ContentID != testContentID || result.Report.DecryptedFiles != 2 || result.Report.CopiedFiles != 3 {
		t.Fatalf("unexpected decryption counts: %#v", result.Report)
	}
	if len(result.Manifests) != 2 || result.Manifests[0].Path != "contents.json" || result.Manifests[1].Path != "subpacks/hd/contents.json" {
		t.Fatalf("unexpected contents manifests: %#v", result.Manifests)
	}
	for _, manifest := range result.Manifests {
		if !json.Valid(manifest.Plaintext) {
			t.Fatalf("manifest %s plaintext is not valid JSON", manifest.Path)
		}
	}

	files := readArchiveFiles(t, output.Bytes())
	want := map[string]string{
		"manifest.json":        `{"format_version":2}`,
		"pack_icon.png":        "plain icon",
		"textures/root.txt":    "root plaintext",
		"plain.txt":            "root plain file",
		"subpacks/hd/file.txt": "subpack plaintext",
	}
	if _, exists := files["contents.json"]; exists {
		t.Fatal("decrypted archive retained encrypted root contents.json")
	}
	if _, exists := files["subpacks/hd/contents.json"]; exists {
		t.Fatal("decrypted archive retained encrypted subpack contents.json")
	}
	if len(files) != len(want) {
		t.Fatalf("decrypted archive has %d files, want %d: %#v", len(files), len(want), files)
	}
	for name, expected := range want {
		if string(files[name]) != expected {
			t.Errorf("decrypted %s = %q, want %q", name, files[name], expected)
		}
	}
}

func TestDecryptArchivePreservesKeyedRootMetadata(t *testing.T) {
	t.Parallel()

	fileKey := "ZYXWVUTSRQPONMLKJIHGFEDCBA654321"
	for _, prefix := range []string{"", "outer/"} {
		t.Run(prefix, func(t *testing.T) {
			plain := map[string][]byte{
				"manifest.json":     []byte(`{"format_version":2}`),
				"pack_icon.png":     []byte("plain pack icon"),
				"bug_pack_icon.png": []byte("plain bug icon"),
			}
			entries := []testArchiveEntry{}
			declarations := []contentsEntry{}
			for _, name := range []string{"manifest.json", "pack_icon.png", "bug_pack_icon.png"} {
				entries = append(entries, testArchiveEntry{name: prefix + name, data: plain[name]})
				declarations = append(declarations, contentsEntry{Path: name, Key: fileKey})
			}
			for _, name := range []string{"assets/manifest.json", "assets/pack_icon.png"} {
				data := []byte("encrypted non-metadata payload")
				entries = append(entries, testArchiveEntry{name: prefix + name, data: encryptCFB8(t, []byte(fileKey), data)})
				declarations = append(declarations, contentsEntry{Path: name, Key: fileKey})
			}
			entries = append(entries, testArchiveEntry{name: prefix + "contents.json", data: encryptedContents(t, testContentKey, testContentID, declarations)})
			input := buildEncryptedArchive(t, entries)
			var output bytes.Buffer
			result, err := DecryptArchive(context.Background(), bytes.NewReader(input), int64(len(input)), testContentKey, &output)
			if err != nil {
				t.Fatal(err)
			}
			files := readArchiveFiles(t, output.Bytes())
			for name, want := range plain {
				if !bytes.Equal(files[prefix+name], want) {
					t.Errorf("root metadata %s changed", prefix+name)
				}
			}
			for _, name := range []string{"assets/manifest.json", "assets/pack_icon.png"} {
				if string(files[prefix+name]) != "encrypted non-metadata payload" {
					t.Errorf("non-metadata entry %s was not decrypted", prefix+name)
				}
			}
			if result.Report.CopiedFiles != 3 || result.Report.DecryptedFiles != 2 {
				t.Errorf("actual copy/decrypt counts = %+v", result.Report)
			}
		})
	}
}

func TestDecryptArchiveAllowsImplicitDirectoryDeclarations(t *testing.T) {
	t.Parallel()

	fileKey := "ZYXWVUTSRQPONMLKJIHGFEDCBA654321"
	input := buildEncryptedArchive(t, []testArchiveEntry{
		{name: "manifest.json", data: []byte(`{"format_version":2}`)},
		{name: "animation_controllers/example.json", data: encryptCFB8(t, []byte(fileKey), []byte("encrypted controller"))},
		{name: "texts/en.lang", data: []byte("plain translation")},
		{name: "contents.json", data: encryptedContents(t, testContentKey, testContentID, []contentsEntry{
			{Path: "animation_controllers"},
			{Path: "animation_controllers/example.json", Key: fileKey},
			{Path: "texts"},
			{Path: "texts/en.lang"},
		})},
	})

	var output bytes.Buffer
	result, err := DecryptArchive(context.Background(), bytes.NewReader(input), int64(len(input)), testContentKey, &output)
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.DecryptedFiles != 1 || result.Report.CopiedFiles != 2 {
		t.Fatalf("directory declarations were counted as files: %#v", result.Report)
	}
	files := readArchiveFiles(t, output.Bytes())
	if string(files["animation_controllers/example.json"]) != "encrypted controller" {
		t.Fatalf("decrypted directory child = %q", files["animation_controllers/example.json"])
	}
	if string(files["texts/en.lang"]) != "plain translation" {
		t.Fatalf("copied directory child = %q", files["texts/en.lang"])
	}
}

func TestDecryptArchiveRejectsMissingDirectoryDeclaration(t *testing.T) {
	t.Parallel()

	input := buildEncryptedArchive(t, []testArchiveEntry{
		{name: "manifest.json", data: []byte(`{"format_version":2}`)},
		{name: "contents.json", data: encryptedContents(t, testContentKey, testContentID, []contentsEntry{{Path: "missing"}})},
	})
	var output bytes.Buffer
	if _, err := DecryptArchive(context.Background(), bytes.NewReader(input), int64(len(input)), testContentKey, &output); err == nil || !strings.Contains(err.Error(), "missing entry missing") {
		t.Fatalf("DecryptArchive error = %v, want missing directory entry", err)
	}
}

func TestDecryptArchiveRejectsKeyedDirectoryDeclaration(t *testing.T) {
	t.Parallel()

	fileKey := "ZYXWVUTSRQPONMLKJIHGFEDCBA654321"
	input := buildEncryptedArchive(t, []testArchiveEntry{
		{name: "manifest.json", data: []byte(`{"format_version":2}`)},
		{name: "animation_controllers/example.json", data: encryptCFB8(t, []byte(fileKey), []byte("encrypted controller"))},
		{name: "contents.json", data: encryptedContents(t, testContentKey, testContentID, []contentsEntry{{Path: "animation_controllers", Key: fileKey}})},
	})
	var output bytes.Buffer
	if _, err := DecryptArchive(context.Background(), bytes.NewReader(input), int64(len(input)), testContentKey, &output); err == nil || !strings.Contains(err.Error(), "directory animation_controllers has an encryption key") {
		t.Fatalf("DecryptArchive error = %v, want keyed directory error", err)
	}
}

func TestDecryptArchiveRejectsWrongContentKey(t *testing.T) {
	t.Parallel()

	input := buildEncryptedArchive(t, []testArchiveEntry{
		{name: "manifest.json", data: []byte("{}")},
		{name: "contents.json", data: encryptedContents(t, testContentKey, testContentID, nil)},
	})
	wrongKey := "654321ZYXWVUTSRQPONMLKJIHGFEDCBA"
	var output bytes.Buffer
	if _, err := DecryptArchive(context.Background(), bytes.NewReader(input), int64(len(input)), wrongKey, &output); err == nil {
		t.Fatal("DecryptArchive accepted a wrong content key")
	}
}

func TestDecryptArchiveRejectsEscapingContentPath(t *testing.T) {
	t.Parallel()

	input := buildEncryptedArchive(t, []testArchiveEntry{
		{name: "manifest.json", data: []byte("{}")},
		{name: "contents.json", data: encryptedContents(t, testContentKey, testContentID, []contentsEntry{{Path: "../escape.txt"}})},
	})
	var output bytes.Buffer
	if _, err := DecryptArchive(context.Background(), bytes.NewReader(input), int64(len(input)), testContentKey, &output); err == nil {
		t.Fatal("DecryptArchive accepted an escaping content path")
	}
}

func TestDecryptArchiveRejectsNullContent(t *testing.T) {
	t.Parallel()

	input := buildEncryptedArchive(t, []testArchiveEntry{
		{name: "manifest.json", data: []byte("{}")},
		{name: "contents.json", data: encryptedContents(t, testContentKey, testContentID, nil)},
	})
	var output bytes.Buffer
	if _, err := DecryptArchive(context.Background(), bytes.NewReader(input), int64(len(input)), testContentKey, &output); err == nil {
		t.Fatal("DecryptArchive accepted a null content value")
	}
}

type testArchiveEntry struct {
	name string
	data []byte
}

func buildEncryptedArchive(t *testing.T, entries []testArchiveEntry) []byte {
	t.Helper()

	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, entry := range entries {
		file, err := writer.CreateHeader(&zip.FileHeader{Name: entry.name, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(entry.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func encryptedContents(t *testing.T, contentKey, contentID string, entries []contentsEntry) []byte {
	t.Helper()

	plaintext, err := json.Marshal(contentsDocument{Content: entries})
	if err != nil {
		t.Fatal(err)
	}
	header := make([]byte, contentsHeaderSize)
	binary.LittleEndian.PutUint32(header[:4], 0)
	copy(header[4:8], contentsMagic[:])
	header[16] = 36
	copy(header[17:], contentID)
	ciphertext := encryptCFB8(t, []byte(contentKey), plaintext)
	result := make([]byte, len(header)+len(ciphertext))
	copy(result, header)
	copy(result[len(header):], ciphertext)
	return result
}

func encryptCFB8(t *testing.T, key, plaintext []byte) []byte {
	t.Helper()

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	register := append([]byte(nil), key[:aes.BlockSize]...)
	scratch := make([]byte, aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))
	for index, value := range plaintext {
		block.Encrypt(scratch, register)
		ciphertext[index] = value ^ scratch[0]
		copy(register, register[1:])
		register[len(register)-1] = ciphertext[index]
	}
	return ciphertext
}

func readArchiveFiles(t *testing.T, data []byte) map[string][]byte {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][]byte, len(reader.File))
	for _, file := range reader.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		contents, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal(readErr, closeErr)
		}
		files[file.Name] = contents
	}
	return files
}

func decodeHex(t *testing.T, value string) []byte {
	t.Helper()

	decoded, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}
