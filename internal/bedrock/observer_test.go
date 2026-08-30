package bedrock

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestObserverRecordsRawPacketsWithObservedDirection(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	failures := &FailureSink{}
	observer := NewObserver(recorder, failures, "session-test", 1)
	local := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 19132}
	client := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 50000}
	server := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 19132}
	observer.SetFlow("downstream", local, client)
	observer.SetFlow("upstream", local, server)

	header := packet.Header{PacketID: 42, SenderSubClient: 1, TargetSubClient: 2}
	observer.PacketFunc("downstream", "downstream-test")(header, []byte{0xaa}, client, local)
	observer.PacketFunc("upstream", "upstream-test")(header, []byte{0xbb}, server, local)
	if err := failures.Err(); err != nil {
		t.Fatalf("capture failure = %v", err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	var events []capture.Event
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("event count = %d, want 2", len(events))
	}
	for index, event := range events {
		if event.Kind != "packet.raw" || event.Stage != "post_decrypt_decompress_frame" {
			t.Fatalf("event %d = %#v", index, event)
		}
		if event.Direction != []capture.Direction{capture.DirectionClientToServer, capture.DirectionServerToClient}[index] {
			t.Fatalf("event %d direction = %q", index, event.Direction)
		}
		if event.Packet == nil || event.Packet.ID != 42 || event.Packet.DecodeStatus != "not_yet_decoded" {
			t.Fatalf("event %d packet = %#v", index, event.Packet)
		}
		if event.Blob == nil {
			t.Fatalf("event %d has no blob", index)
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(event.Blob.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != 1 || data[0] != []byte{0xaa, 0xbb}[index] {
			t.Fatalf("event %d blob = %x", index, data)
		}
	}
}

func TestCaptureLogHandlerPreservesAttributeGroupsAndIntegerPrecision(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	failures := &FailureSink{}
	logger := NewCaptureLogHandler(recorder, failures, "session-test", "upstream-test", "upstream", nil)
	handler := logger.WithAttrs([]slog.Attr{slog.Int("root", 1)}).WithGroup("nested").WithAttrs([]slog.Attr{slog.Int("bound", 2)})
	record := slog.NewRecord(time.Unix(0, 0), slog.LevelError, "decode batch failed", 0)
	record.AddAttrs(slog.Group("details", slog.Uint64("count", math.MaxUint64)))
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	if err := failures.Err(); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind != "packet.decode_error" {
			t.Fatalf("kind = %q", event.Kind)
		}
		if event.ConnectionID != "upstream-test" || event.Channel != "upstream" {
			t.Fatalf("log context = %#v", event)
		}
		var envelope struct {
			Attributes map[string]any `json:"attributes"`
		}
		if err := json.Unmarshal(event.Data, &envelope); err != nil {
			return err
		}
		if envelope.Attributes["root"] != "1" {
			t.Fatalf("root attribute = %#v", envelope.Attributes["root"])
		}
		nested, ok := envelope.Attributes["nested"].(map[string]any)
		if !ok || nested["bound"] != "2" {
			t.Fatalf("nested attributes = %#v", envelope.Attributes["nested"])
		}
		details, ok := nested["details"].(map[string]any)
		if !ok || details["count"] != "18446744073709551615" {
			t.Fatalf("details = %#v", nested["details"])
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
