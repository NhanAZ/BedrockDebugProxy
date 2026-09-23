package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
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

func TestReorderLoginFlagsKeepsAccountSyntaxShort(t *testing.T) {
	got := reorderLoginFlags([]string{"NhanAZ", "--max-age", "7d"})
	want := []string{"--max-age", "7d", "NhanAZ"}
	if !slices.Equal(got, want) {
		t.Fatalf("reorderLoginFlags() = %#v, want %#v", got, want)
	}
}

func TestLoginCachePolicyDefaultsToNoLocalLimit(t *testing.T) {
	var output bytes.Buffer
	maxAge, setMaxAge, err := loginCachePolicy("", strings.NewReader("\n"), &output, true)
	if err != nil || !setMaxAge || maxAge != 0 {
		t.Fatalf("loginCachePolicy() = %v, %v, %v", maxAge, setMaxAge, err)
	}
	if !strings.Contains(output.String(), "Enter for no limit") {
		t.Fatalf("prompt = %q", output.String())
	}
}

func TestPromptAccountNameAllowsDefaultAndSharesInputReader(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("NhanAZ\n7d\n"))
	var output bytes.Buffer
	name, err := promptAccountName(reader, &output, false)
	if err != nil || name != "NhanAZ" {
		t.Fatalf("promptAccountName() = %q, %v", name, err)
	}
	maxAge, setMaxAge, err := loginCachePolicy("", reader, io.Discard, true)
	if err != nil || !setMaxAge || maxAge != 7*24*time.Hour {
		t.Fatalf("loginCachePolicy() after account prompt = %v, %v, %v", maxAge, setMaxAge, err)
	}
	if !strings.Contains(output.String(), "Enter for default") {
		t.Fatalf("account prompt = %q", output.String())
	}
}

func TestPromptAccountNameRequiresFirstAccountLabel(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\n"))
	var output bytes.Buffer
	if name, err := promptAccountName(reader, &output, true); err == nil || name != "" {
		t.Fatalf("promptAccountName() = %q, %v; want required-name error", name, err)
	}
	if !strings.Contains(output.String(), "required") {
		t.Fatalf("required prompt = %q", output.String())
	}
}

func TestRenameCommandRequiresBothNames(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"rename", "default"}, &stdout, &stderr); code != 2 {
		t.Fatalf("rename code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Both old and new account names are required") {
		t.Fatalf("rename stderr = %s", stderr.String())
	}
}

func TestHelpGroupsCommands(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("help code = %d, stderr = %s", code, stderr.String())
	}
	output := stdout.String()
	for _, section := range []string{"Proxy", "Authentication accounts", "Capture inspection", "Information", "help [COMMAND]", "capture packs only"} {
		if !strings.Contains(output, section) {
			t.Fatalf("help output missing %q: %s", section, output)
		}
	}
}

func TestCommandHelpAndUnknownHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"help", "login"}, &stdout, &stderr); code != 0 {
		t.Fatalf("login help code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Usage - bedrock-debug-proxy login") {
		t.Fatalf("login help = %s", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"help", "missing"}, &stdout, &stderr); code != 2 {
		t.Fatalf("unknown help code = %d", code)
	}
	if !strings.Contains(stderr.String(), "Unknown command") || !strings.Contains(stderr.String(), "Authentication accounts") {
		t.Fatalf("unknown help stderr = %s", stderr.String())
	}
}

func TestRunHelpDocumentsPacksOnlyMode(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"help", "run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run help code = %d, stderr = %s", code, stderr.String())
	}
	for _, expected := range []string{"--packs-only", "without a local listener or world spawn", "--follow-transfers"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("run help missing %q: %s", expected, stdout.String())
		}
	}
}

func TestRunPacksOnlyRejectsListenerOnlyOptions(t *testing.T) {
	for _, extra := range [][]string{
		{"--listen", "127.0.0.1:19132"},
		{"--follow-transfers"},
		{"--allow-unauthenticated-client"},
	} {
		args := append([]string{"run", "--upstream", "127.0.0.1:19132", "--packs-only"}, extra...)
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Fatalf("run(%v) code = %d, want 2; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
		}
		if !strings.Contains(stderr.String(), "--packs-only cannot be combined") {
			t.Fatalf("run(%v) stderr = %s", args, stderr.String())
		}
	}
}

func TestLoginCachePolicySupportsNonInteractiveAndFlagForms(t *testing.T) {
	maxAge, setMaxAge, err := loginCachePolicy("", strings.NewReader("7d\n"), io.Discard, false)
	if err != nil || !setMaxAge || maxAge != 0 {
		t.Fatalf("non-interactive policy = %v, %v, %v", maxAge, setMaxAge, err)
	}
	maxAge, setMaxAge, err = loginCachePolicy("7d", strings.NewReader("invalid\n"), io.Discard, true)
	if err != nil || !setMaxAge || maxAge != 7*24*time.Hour {
		t.Fatalf("flag policy = %v, %v, %v", maxAge, setMaxAge, err)
	}
}

func TestViewRejectsMissingCapture(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"view"}, &stdout, &stderr); code != 2 {
		t.Fatalf("run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "bedrock-debug-proxy view") {
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

func TestCaptureLimitationsDescribeTransportStateBoundary(t *testing.T) {
	limitations := strings.Join(captureLimitations(false, false), "\n")
	if strings.Contains(limitations, "cannot mirror an upstream optional-pack policy") {
		t.Fatalf("stale resource-pack limitation = %s", limitations)
	}
	if !strings.Contains(limitations, "Upstream resource-pack offer metadata is preserved") {
		t.Fatalf("resource-pack offer limitation missing = %s", limitations)
	}
	if strings.Contains(limitations, "not yet classified per event") {
		t.Fatalf("stale transport limitation = %s", limitations)
	}
	if !strings.Contains(limitations, "Early transport payloads may not expose compression or encryption state") {
		t.Fatalf("transport boundary limitation missing = %s", limitations)
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
	if code := run([]string{"inspect", "--kind", "packet.decoded", "--packet", "Text", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("inspect code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind":"packet.decoded"`) {
		t.Fatalf("inspect stdout = %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"inspect", "--packet", "LevelChunk", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("inspect absent packet code = %d, stderr = %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("inspect absent packet stdout = %s", stdout.String())
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
