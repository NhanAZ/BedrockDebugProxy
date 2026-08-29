package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

func TestVerifyCommand(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "cli-test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{Kind: "test"}}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"verify", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("run() code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"capture_id": "cli-test"`) || !strings.Contains(stdout.String(), `"events": 1`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestNextCapturePathAvoidsCollision(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 8, 30, 12, 34, 56, 0, time.UTC)
	first, err := nextCapturePath(root, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := captureDirectory(first); err != nil {
		t.Fatal(err)
	}
	second, err := nextCapturePath(root, now)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasSuffix(second, "-1") {
		t.Fatalf("paths = %q and %q", first, second)
	}
}

func captureDirectory(path string) error {
	recorder, err := capture.New(path, capture.Options{})
	if err != nil {
		return err
	}
	return recorder.Close("closed", nil)
}

func TestRunRejectsMissingUpstream(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"run"}, &stdout, &stderr); code != 2 {
		t.Fatalf("run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "--upstream is required") {
		t.Fatalf("stderr = %s", stderr.String())
	}
}
