package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/NhanAZ/BedrockDebugProxy/internal/analysis"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/NhanAZ/BedrockDebugProxy/internal/golemnet"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestGolemNetworkWritesCanonicalCapture(t *testing.T) {
	listener, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	serverErrors := make(chan error, 1)
	go func() {
		accepted, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverErrors <- acceptErr
			return
		}
		conn, ok := accepted.(*minecraft.Conn)
		if !ok {
			serverErrors <- fmt.Errorf("unexpected server connection %T", accepted)
			_ = accepted.Close()
			return
		}
		defer func() { _ = conn.Close() }()
		if err := conn.StartGame(minecraft.GameData{WorldName: "Canonical capture test", EntityUniqueID: 1, EntityRuntimeID: 1}); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.WritePacket(&packet.Text{TextType: packet.TextTypeRaw, Message: "capture-ready"}); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.Flush(); err != nil {
			serverErrors <- err
			return
		}
		for {
			if _, err := conn.ReadPacket(); err != nil {
				serverErrors <- nil
				return
			}
		}
	}()

	root := t.TempDir()
	captureRoot := filepath.Join(root, "capture")
	scenarioPath := filepath.Join(root, "scenario.jsonl")
	scenario := []map[string]any{
		{"op": "wait", "duration": "100ms"},
		{"op": "quit"},
	}
	var scenarioData bytes.Buffer
	for _, request := range scenario {
		if err := json.NewEncoder(&scenarioData).Encode(request); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(scenarioPath, scenarioData.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	args := []string{
		"golem", "network", "--server", listener.Addr().String(), "--auth", "none",
		"--capture", captureRoot, "--input", scenarioPath, "--timeout", "2s", "--connect-timeout", "5s",
	}
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("golem network exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if err := <-serverErrors; err != nil {
		t.Fatal(err)
	}
	if _, err := capture.Verify(captureRoot); err != nil {
		t.Fatal(err)
	}
	summary, err := analysis.Analyze(captureRoot)
	if err != nil {
		t.Fatal(err)
	}
	if !summary.Session.UpstreamConnected || !summary.Session.Negotiated || !summary.Session.Spawned || !summary.Session.Closed {
		t.Fatalf("session evidence = %+v", summary.Session)
	}
	if summary.Session.RawPacketEvents == 0 || summary.Session.DecodedPacketEvents == 0 || summary.Session.ErrorEvents != 0 {
		t.Fatalf("packet evidence = %+v", summary.Session)
	}
	artifactStatus, err := os.ReadFile(filepath.Join(captureRoot, "artifacts", "status.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(artifactStatus, []byte(`"state": "complete"`)) {
		t.Fatalf("automatic artifacts did not complete: %s", artifactStatus)
	}
}

func TestGolemCaptureWritesCanonicalEvidence(t *testing.T) {
	root := t.TempDir()
	recorder, err := capture.New(root, capture.Options{
		GeneratorName: "test-golem",
		Version:       "test",
		Commit:        "0123456789abcdef0123456789abcdef01234567",
		RawLayers:     []string{"bedrock_packet_payload"},
		Values: map[string]string{
			"protocol_id":            "2225",
			"game_version":           "1.26.62",
			"decrypt_resource_packs": "true",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	evidence := newGolemCapture(recorder, "golem-test", true)
	if err := evidence.onStart(); err != nil {
		t.Fatal(err)
	}
	if err := evidence.record(context.Background(), capture.Event{
		SessionID:    "golem-test",
		ConnectionID: golemConnectionID(1),
		Hop:          1,
		Kind:         "upstream.connected",
		Direction:    capture.DirectionInternal,
		Channel:      "upstream",
		Stage:        "transport_connected",
	}, nil, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := evidence.record(context.Background(), capture.Event{
		SessionID:    "golem-test",
		ConnectionID: golemConnectionID(1),
		Hop:          1,
		Kind:         "packet.raw",
		Direction:    capture.DirectionClientToServer,
		Channel:      "upstream",
		Stage:        "post_decrypt_decompress_frame",
		Packet:       &capture.PacketInfo{ID: 9, Name: "Text", DecodeStatus: "not_yet_decoded"},
	}, []byte{1, 2, 3}, "application/octet-stream", "bedrock_packet_payload"); err != nil {
		t.Fatal(err)
	}
	if err := evidence.recordDecoded(context.Background(), &packet.Text{}, capture.DirectionClientToServer, 1); err != nil {
		t.Fatal(err)
	}
	if err := evidence.recordJSON(context.Background(), "session.negotiated", capture.SeverityInfo, capture.DirectionInternal, "upstream", "negotiated", 1, map[string]any{
		"downstream_protocol_id":  2225,
		"downstream_game_version": "1.26.62",
		"upstream_protocol_id":    2225,
		"upstream_game_version":   "1.26.62",
	}); err != nil {
		t.Fatal(err)
	}
	if err := evidence.recordStructured(context.Background(), "session.game_data", map[string]any{"world_name": "synthetic"}, capture.DirectionServerToClient, 1, "decoded_upstream_game_data", nil); err != nil {
		t.Fatal(err)
	}
	if err := evidence.recordJSON(context.Background(), "session.spawned", capture.SeverityInfo, capture.DirectionInternal, "upstream", "lifecycle", 1, map[string]any{"world_spawned": true}); err != nil {
		t.Fatal(err)
	}
	evidence.onClose(nil, golemnet.Report{Spawned: true, Closed: true, Hops: 1}, 1)
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := capture.Verify(root); err != nil {
		t.Fatal(err)
	}
	summary, err := analysis.Analyze(root)
	if err != nil {
		t.Fatal(err)
	}
	if !summary.Session.UpstreamConnected || !summary.Session.Negotiated || !summary.Session.Spawned || !summary.Session.Closed {
		t.Fatalf("session evidence = %+v", summary.Session)
	}
	if summary.Session.RawPacketEvents != 1 || summary.Session.DecodedPacketEvents != 1 || summary.Session.ErrorEvents != 0 {
		t.Fatalf("packet evidence = %+v", summary.Session)
	}
}
