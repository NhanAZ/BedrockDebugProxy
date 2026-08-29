package capturearchive

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

func TestExportProducesDeterministicVerifiedArchive(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "archive-test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{Kind: "packet.raw"}, Raw: []byte{1, 2, 3}}); err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{Kind: "packet.raw"}, Raw: []byte{1, 2, 3}}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	firstPath := filepath.Join(t.TempDir(), "first.bdpcap")
	secondPath := filepath.Join(t.TempDir(), "second.bdpcap")
	first, err := Export(root, firstPath)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Export(root, secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if first.SHA256 != second.SHA256 || first.Bytes != second.Bytes || first.Entries != 3 {
		t.Fatalf("exports = %#v %#v", first, second)
	}
	firstBytes, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("exports are not byte-for-byte deterministic")
	}
	reader, err := zip.OpenReader(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()
	if reader.Comment != capture.SchemaVersion {
		t.Fatalf("archive comment = %q", reader.Comment)
	}
}

func TestExportRejectsOpenCaptureAndExistingOutput(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "capture.bdpcap")
	if _, err := Export(root, output); err == nil {
		t.Fatal("Export() accepted an open capture")
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Export(root, output); err == nil {
		t.Fatal("Export() replaced an existing output")
	}
}

func TestExportRejectsBlobSymlinkOutsideCapture(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	recorded, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{Kind: "packet.raw"}, Raw: []byte("outside")})
	if err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	blobPath := filepath.Join(root, filepath.FromSlash(recorded.Blob.Path))
	outsidePath := filepath.Join(t.TempDir(), "outside.bin")
	if err := os.WriteFile(outsidePath, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(blobPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsidePath, blobPath); err != nil {
		t.Skipf("symbolic links are unavailable: %v", err)
	}
	if _, err := Export(root, filepath.Join(t.TempDir(), "capture.bdpcap")); err == nil {
		t.Fatal("Export() followed a blob symlink outside the capture")
	}
}
