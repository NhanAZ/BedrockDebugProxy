package analysis

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

func TestAnalyzeAndExplainCapture(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "analysis-test", Limitations: []string{"synthetic limitation"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		Kind: "packet.decoded", Direction: capture.DirectionClientToServer, Channel: "bridge",
		Packet: &capture.PacketInfo{ID: 9, Name: "Text", DecodeStatus: "decoded"},
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		Kind: "bridge.read_error", Severity: capture.SeverityError, Direction: capture.DirectionServerToClient,
		Error: &capture.ErrorInfo{Operation: "read packet", Message: "synthetic failure", Type: "*errors.errorString"},
	}}); err != nil {
		t.Fatal(err)
	}
	packData := []byte(`{"uuid":"11111111-1111-1111-1111-111111111111","version":"1.0.0","name":"Test Pack","archive_bytes":7,"checksum_sha256":"abcd","delivery":"raknet","encrypted":true,"content_key":"secret"}`)
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		Kind: "resource_pack.archive", Direction: capture.DirectionServerToClient, Data: packData,
	}, Raw: []byte("archive"), Representation: "minecraft_resource_pack_archive"}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	summary, err := Analyze(root)
	if err != nil {
		t.Fatal(err)
	}
	if summary.CaptureID != "analysis-test" || summary.ObservedEvents != 3 || len(summary.Packets) != 1 || len(summary.Errors) != 1 || len(summary.ResourcePacks) != 1 {
		t.Fatalf("summary = %#v", summary)
	}
	if !summary.ResourcePacks[0].HasContentKey || summary.ResourcePacks[0].BlobSHA256 == "" {
		t.Fatalf("resource pack = %#v", summary.ResourcePacks[0])
	}
	explanation := Explain(summary)
	for _, expected := range []string{"analysis-test", "synthetic limitation", "Test Pack", "synthetic failure", "real Minecraft client"} {
		if !strings.Contains(explanation, expected) {
			t.Fatalf("explanation does not contain %q:\n%s", expected, explanation)
		}
	}
}

func TestExplainSanitizesCapturedMarkdownFields(t *testing.T) {
	explanation := Explain(Summary{
		CaptureID: "capture`\n# injected",
		Schema:    "schema`value",
		Status:    "closed",
		Limitations: []string{
			"line one\n# line two",
		},
		ResourcePacks: []ResourcePack{{Name: "pack`name", Delivery: "raknet\n# injected"}},
	})
	if strings.Contains(explanation, "`\n#") || strings.Contains(explanation, "raknet\n# injected") {
		t.Fatalf("explanation contains captured Markdown structure:\n%s", explanation)
	}
}
