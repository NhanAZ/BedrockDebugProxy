package bedrock

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
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
	proxy := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 50000}
	if err := RecordResourcePacks(context.Background(), recorder, "session-test", "upstream-test", 1, server, proxy, []*resource.Pack{pack}); err != nil {
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
