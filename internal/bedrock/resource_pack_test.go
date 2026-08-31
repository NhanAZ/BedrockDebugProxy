package bedrock

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/NhanAZ/BedrockDebugProxy/internal/artifacts"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

func TestRecordResourcePacksPreservesArchiveMetadataAndContentKey(t *testing.T) {
	archive := testResourcePackArchive(t)
	pack, err := resource.Read(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	pack = pack.WithContentKey("test-content-key")
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	server := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 19132}
	startPackArtifactCheck(t, root, recorder, 1)
	proxy := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 50000}
	if err := RecordResourcePacks(context.Background(), recorder, "session-test", "upstream-test", 1, server, proxy, []*resource.Pack{pack}, ResourcePackCaptureOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind != "resource_pack.archive" || event.Blob == nil {
			t.Fatalf("event = %#v", event)
		}
		if event.Source == nil || event.Source.Address != server.String() || event.Destination == nil || event.Destination.Address != proxy.String() {
			t.Fatalf("endpoints = %#v -> %#v", event.Source, event.Destination)
		}
		stored, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(event.Blob.Path)))
		if err != nil {
			return err
		}
		if !bytes.Equal(stored, archive) {
			t.Fatal("recorded resource pack differs from the downloaded archive")
		}
		var data resourcePackData
		if err := json.Unmarshal(event.Data, &data); err != nil {
			return err
		}
		if data.UUID != "11111111-1111-1111-1111-111111111111" || data.Version != "1.0.0" || data.Name != "Test Pack" {
			t.Fatalf("metadata = %#v", data)
		}
		if !data.Encrypted || data.ContentKey != "test-content-key" || !data.HasTextures {
			t.Fatalf("pack flags = %#v", data)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	verification, err := capture.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) != 0 {
		t.Fatalf("capture issues = %v", verification.Issues)
	}
}

func TestRecordResourcePacksStoresDerivedDecryptionArtifacts(t *testing.T) {
	const contentKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"
	const fileKey = "ZYXWVUTSRQPONMLKJIHGFEDCBA654321"
	archive := testEncryptedResourcePackArchive(t, contentKey, fileKey)
	pack, err := resource.Read(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	pack = pack.WithContentKey(contentKey)
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	server := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 19132}
	startPackArtifactCheck(t, root, recorder, 2)
	proxy := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 50000}
	if err := RecordResourcePacks(context.Background(), recorder, "session-test", "upstream-test", 1, server, proxy, []*resource.Pack{pack}, ResourcePackCaptureOptions{Decrypt: true}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	var kinds []string
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		kinds = append(kinds, event.Kind)
		if event.Sequence > 1 && (event.ParentSequence == nil || *event.ParentSequence != 1) {
			t.Fatalf("event %d parent = %v, want 1", event.Sequence, event.ParentSequence)
		}
		if event.Kind != "resource_pack.archive" && bytes.Contains(event.Data, []byte(fileKey)) {
			t.Fatalf("derived event %s exposes a file key in metadata", event.Kind)
		}
		if event.Kind != "resource_pack.decrypted_archive" {
			return nil
		}
		stored, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(event.Blob.Path)))
		if err != nil {
			return err
		}
		files := readTestZip(t, stored)
		if string(files["textures/secret.txt"]) != "decrypted fixture" {
			t.Fatalf("decrypted file = %q", files["textures/secret.txt"])
		}
		if _, exists := files["contents.json"]; exists {
			t.Fatal("derived archive retained encrypted contents.json")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	wantKinds := []string{"resource_pack.archive", "resource_pack.contents_manifest", "resource_pack.decrypted_archive"}
	if !slicesEqual(kinds, wantKinds) {
		t.Fatalf("event kinds = %v, want %v", kinds, wantKinds)
	}
	verification, err := capture.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) != 0 {
		t.Fatalf("capture issues = %v", verification.Issues)
	}
}

func TestRecordResourcePacksKeepsRawArchiveWhenDecryptionFails(t *testing.T) {
	const contentKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"
	const wrongKey = "654321ZYXWVUTSRQPONMLKJIHGFEDCBA"
	archive := testEncryptedResourcePackArchive(t, contentKey, "ZYXWVUTSRQPONMLKJIHGFEDCBA654321")
	pack, err := resource.Read(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	pack = pack.WithContentKey(wrongKey)
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := RecordResourcePacks(context.Background(), recorder, "session-test", "upstream-test", 1, nil, nil, []*resource.Pack{pack}, ResourcePackCaptureOptions{Decrypt: true}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	manifest := recorder.Manifest()
	if manifest.Completeness.Complete || len(manifest.Completeness.Limitations) != 1 {
		t.Fatalf("capture completeness = %#v", manifest.Completeness)
	}
	var events []capture.Event
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Kind != "resource_pack.archive" || events[1].Kind != "resource_pack.decrypt_error" {
		t.Fatalf("events = %#v", events)
	}
	if events[1].Error == nil || events[1].ParentSequence == nil || *events[1].ParentSequence != events[0].Sequence {
		t.Fatalf("decryption error event = %#v", events[1])
	}
	stored, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(events[0].Blob.Path)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stored, archive) {
		t.Fatal("failed decryption changed or discarded the raw archive")
	}
}

func startPackArtifactCheck(t *testing.T, root string, recorder *capture.Recorder, archives int) {
	t.Helper()
	views, err := artifacts.Start(root, protocol.CurrentProtocol, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = recorder.Close("closed", nil)
		status, err := views.Finish()
		if err != nil || status.State != "complete" {
			t.Errorf("pack view = %+v, %v", status, err)
			return
		}
		data, err := os.ReadFile(filepath.Join(root, "artifacts/packs/index.jsonl"))
		if err != nil {
			t.Error(err)
			return
		}
		lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
		if len(lines) != archives {
			t.Errorf("pack views = %d, want %d", len(lines), archives)
		}
		for _, line := range lines {
			var entry struct {
				Path   string          `json:"path"`
				Source capture.BlobRef `json:"source_blob"`
			}
			if err := json.Unmarshal(line, &entry); err != nil {
				t.Error(err)
				continue
			}
			original, err := os.ReadFile(filepath.Join(root, entry.Source.Path))
			if err != nil {
				t.Error(err)
				continue
			}
			copyData, err := os.ReadFile(filepath.Join(root, entry.Path))
			if err != nil || !bytes.Equal(original, copyData) {
				t.Errorf("pack copy differs: %v", err)
			}
			_ = readTestZip(t, copyData)
		}
	})
}

func testResourcePackArchive(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	manifest, err := writer.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = manifest.Write([]byte(`{"format_version":2,"header":{"name":"Test Pack","description":"Synthetic fixture","uuid":"11111111-1111-1111-1111-111111111111","version":[1,0,0],"min_engine_version":[1,20,0]},"modules":[{"type":"resources","uuid":"22222222-2222-2222-2222-222222222222","version":[1,0,0]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	texture, err := writer.Create("textures/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := texture.Write([]byte("fixture")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func testEncryptedResourcePackArchive(t *testing.T, contentKey, fileKey string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	entries := []struct {
		name string
		data []byte
	}{
		{
			name: "manifest.json",
			data: []byte(`{"format_version":2,"header":{"name":"Encrypted Test Pack","description":"Synthetic fixture","uuid":"11111111-1111-1111-1111-111111111111","version":[1,0,0],"min_engine_version":[1,20,0]},"modules":[{"type":"resources","uuid":"22222222-2222-2222-2222-222222222222","version":[1,0,0]}]}`),
		},
		{name: "textures/secret.txt", data: testEncryptCFB8(t, []byte(fileKey), []byte("decrypted fixture"))},
		{name: "contents.json", data: testEncryptedContents(t, contentKey, fileKey)},
	}
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
	return buffer.Bytes()
}

func testEncryptedContents(t *testing.T, contentKey, fileKey string) []byte {
	t.Helper()
	const contentID = "11111111-1111-1111-1111-111111111111"
	plaintext := []byte(`{"content":[{"path":"textures/secret.txt","key":"` + fileKey + `"}]}`)
	header := make([]byte, 256)
	binary.LittleEndian.PutUint32(header[:4], 0)
	copy(header[4:8], []byte{0xfc, 0xb9, 0xcf, 0x9b})
	header[16] = 36
	copy(header[17:], contentID)
	ciphertext := testEncryptCFB8(t, []byte(contentKey), plaintext)
	result := make([]byte, len(header)+len(ciphertext))
	copy(result, header)
	copy(result[len(header):], ciphertext)
	return result
}

func testEncryptCFB8(t *testing.T, key, plaintext []byte) []byte {
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

func readTestZip(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][]byte, len(reader.File))
	for _, file := range reader.File {
		contents, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		value, readErr := io.ReadAll(contents)
		closeErr := contents.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal(readErr, closeErr)
		}
		files[file.Name] = value
	}
	return files
}

func slicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
