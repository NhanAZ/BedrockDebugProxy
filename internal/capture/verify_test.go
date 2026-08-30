package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyReportsTamperedBlob(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := New(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	event, err := recorder.Record(context.Background(), Record{Event: Event{Kind: "raw"}, Raw: []byte("original")})
	if err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(event.Blob.Path)), []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	verification, err := Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) != 1 || !strings.Contains(verification.Issues[0], "SHA-256") {
		t.Fatalf("issues = %v", verification.Issues)
	}
}

func TestScanEventsRejectsTruncatedLine(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, eventsName), []byte(`{"sequence":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	err := ScanEvents(root, func(Event) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "not newline terminated") {
		t.Fatalf("ScanEvents() error = %v", err)
	}
}

func TestVerifyRejectsInvalidRepeatedBlobReference(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := New(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := recorder.Record(context.Background(), Record{Event: Event{Kind: "packet.raw"}, Raw: []byte("shared")}); err != nil {
			t.Fatal(err)
		}
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	eventsPath := filepath.Join(root, eventsName)
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSuffix(data, []byte{'\n'}), []byte{'\n'})
	if len(lines) != 2 {
		t.Fatalf("event lines = %d, want 2", len(lines))
	}
	var event Event
	if err := json.Unmarshal(lines[1], &event); err != nil {
		t.Fatal(err)
	}
	event.Blob.Path = "../outside.bin"
	lines[1], err = json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	// #nosec G703 -- eventsPath is the fixed events file under a test-owned temporary capture root.
	if err := os.WriteFile(eventsPath, append(bytes.Join(lines, []byte{'\n'}), '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	verification, err := Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) == 0 || !strings.Contains(strings.Join(verification.Issues, "\n"), "unsafe relative path") {
		t.Fatalf("issues = %v", verification.Issues)
	}
}

func TestVerifyReportsManifestBlobByteMismatch(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := New(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), Record{Event: Event{Kind: "packet.raw"}, Raw: []byte("blob")}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, manifestName)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Counts.BlobBytes++
	data, err = json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	verification, err := Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) == 0 || !strings.Contains(strings.Join(verification.Issues, "\n"), "unique blob bytes") {
		t.Fatalf("issues = %v", verification.Issues)
	}
}

func TestVerifyRejectsBlobSymlinkOutsideCapture(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := New(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	event, err := recorder.Record(context.Background(), Record{Event: Event{Kind: "packet.raw"}, Raw: []byte("outside")})
	if err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	blobPath := filepath.Join(root, filepath.FromSlash(event.Blob.Path))
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
	verification, err := Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) == 0 || !strings.Contains(strings.Join(verification.Issues, "\n"), "outside the capture") {
		t.Fatalf("issues = %v", verification.Issues)
	}
}
