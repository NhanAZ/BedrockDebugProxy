package capture

import (
	"context"
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
