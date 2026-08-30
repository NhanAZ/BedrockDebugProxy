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
	recorder, err := capture.New(root, capture.Options{
		CaptureID: "analysis-test", Version: "test", Commit: "0123456789abcdef0123456789abcdef01234567",
		Values:      map[string]string{"protocol_id": "900", "game_version": "1.2.3", "decrypt_resource_packs": "true"},
		Limitations: []string{"synthetic limitation"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		Kind: "packet.decoded", Direction: capture.DirectionClientToServer, Channel: "bridge",
		Packet: &capture.PacketInfo{ID: 9, Name: "Text", DecodeStatus: "decoded"},
	}}); err != nil {
		t.Fatal(err)
	}
	negotiatedData := []byte(`{"downstream_protocol_id":900,"downstream_game_version":"1.2.3","upstream_protocol_id":900,"upstream_game_version":"1.2.3"}`)
	for _, event := range []capture.Event{
		{Kind: "upstream.connected", Direction: capture.DirectionInternal},
		{Kind: "session.negotiated", Direction: capture.DirectionInternal, Data: negotiatedData},
		{Kind: "session.connection_metadata", Direction: capture.DirectionInternal},
		{Kind: "session.game_data", Direction: capture.DirectionServerToClient},
		{Kind: "session.spawned", Direction: capture.DirectionInternal},
		{Kind: "transport.payload", Direction: capture.DirectionServerToClient},
		{Kind: "packet.raw", Direction: capture.DirectionServerToClient},
		{Kind: "session.close", Direction: capture.DirectionInternal},
	} {
		if _, err := recorder.Record(context.Background(), capture.Record{Event: event}); err != nil {
			t.Fatal(err)
		}
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
	parent := uint64(11)
	decryptionData := []byte(`{"uuid":"11111111-1111-1111-1111-111111111111","version":"1.0.0","report":{"algorithm":"AES-256-CFB8","authenticated":false,"decrypted_files":2}}`)
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		Kind: "resource_pack.decrypted_archive", Direction: capture.DirectionServerToClient, Data: decryptionData, ParentSequence: &parent,
	}, Raw: []byte("decrypted archive"), Representation: "minecraft_resource_pack_decrypted_archive"}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	summary, err := Analyze(root)
	if err != nil {
		t.Fatal(err)
	}
	if summary.CaptureID != "analysis-test" || summary.ObservedEvents != 12 || len(summary.Packets) != 1 || len(summary.Errors) != 1 || len(summary.ResourcePacks) != 1 || len(summary.PackDecryptions) != 1 {
		t.Fatalf("summary = %#v", summary)
	}
	if summary.Build.Commit != "0123456789abcdef0123456789abcdef01234567" || !summary.Protocol.PackDecryptEnabled || summary.Protocol.DownstreamID != 900 {
		t.Fatalf("build and protocol summary = %#v %#v", summary.Build, summary.Protocol)
	}
	if !summary.Session.UpstreamConnected || !summary.Session.Negotiated || !summary.Session.Spawned || !summary.Session.Closed || summary.Session.GameDataViews != 1 || summary.Session.RawPacketEvents != 1 || summary.Session.ErrorEvents != 1 {
		t.Fatalf("session summary = %#v", summary.Session)
	}
	if !summary.ResourcePacks[0].HasContentKey || summary.ResourcePacks[0].BlobSHA256 == "" {
		t.Fatalf("resource pack = %#v", summary.ResourcePacks[0])
	}
	if summary.PackDecryptions[0].Algorithm != "AES-256-CFB8" || summary.PackDecryptions[0].Authenticated || summary.PackDecryptions[0].DecryptedFiles != 2 {
		t.Fatalf("resource pack decryption = %#v", summary.PackDecryptions[0])
	}
	explanation := Explain(summary)
	for _, expected := range []string{"analysis-test", "synthetic limitation", "Test Pack", "AES-256-CFB8", "synthetic failure", "real Minecraft client"} {
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
