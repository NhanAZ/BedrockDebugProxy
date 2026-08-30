package main

import (
	"bytes"
	"context"
	"os"
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

func TestRunRejectsExperienceWithoutDeviceAuthentication(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"run", "--upstream", "experience:The Hive", "--auth", "none"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "Experience targets require --auth device") {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestCaptureInspectionAnalysisExplanationAndExportCommands(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "command-test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		Kind: "packet.decoded", Direction: capture.DirectionClientToServer, Channel: "bridge",
		Packet: &capture.PacketInfo{ID: 9, Name: "Text", DecodeStatus: "decoded"},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if code := run([]string{"inspect", "--kind", "packet.decoded", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("inspect code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind":"packet.decoded"`) {
		t.Fatalf("inspect stdout = %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"analyze", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("analyze code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"capture_id": "command-test"`) || !strings.Contains(stdout.String(), `"name": "Text"`) {
		t.Fatalf("analyze stdout = %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"explain", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("explain code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "command-test") || !strings.Contains(stdout.String(), "verifier found no") {
		t.Fatalf("explain stdout = %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	output := filepath.Join(t.TempDir(), "command-test.bdpcap")
	if code := run([]string{"export", root, output}, &stdout, &stderr); code != 0 {
		t.Fatalf("export code = %d, stderr = %s", code, stderr.String())
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatalf("export output = %v, %v", info, err)
	}
	if !strings.Contains(stdout.String(), `"entries": 2`) {
		t.Fatalf("export stdout = %s", stdout.String())
	}
}

func TestExportWarnsAboutDecryptedResourcePackArtifacts(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "decrypted-export-test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), capture.Record{
		Event: capture.Event{Kind: "resource_pack.contents_manifest"},
		Raw:   []byte(`{"content":[]}`),
	}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	output := filepath.Join(t.TempDir(), "decrypted-export-test.bdpcap")
	if code := run([]string{"export", root, output}, &stdout, &stderr); code != 0 {
		t.Fatalf("export code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"contains_decrypted_resource_packs": true`) {
		t.Fatalf("export stdout = %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "ownership and redistribution rights") {
		t.Fatalf("export stderr = %s", stderr.String())
	}
}
