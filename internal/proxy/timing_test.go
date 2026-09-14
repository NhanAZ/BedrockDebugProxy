package proxy

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestShouldRecordForwardTiming(t *testing.T) {
	tests := []struct {
		name   string
		packet packet.Packet
		want   bool
	}{
		{name: "chunk", packet: &packet.LevelChunk{}, want: true},
		{name: "block update", packet: &packet.UpdateBlock{}, want: true},
		{name: "movement correction", packet: &packet.MovePlayer{}, want: true},
		{name: "transfer", packet: &packet.Transfer{}, want: true},
		{name: "ordinary packet", packet: &packet.Text{}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldRecordForwardTiming(test.packet); got != test.want {
				t.Fatalf("shouldRecordForwardTiming(%T) = %t, want %t", test.packet, got, test.want)
			}
		})
	}
}

func TestRecordForwardTimingPreservesPacketAncestry(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	runner := &Runner{config: Config{Recorder: recorder}}
	parent := capture.Event{
		Sequence: 42,
		Packet:   &capture.PacketInfo{ID: packet.IDLevelChunk, Name: "LevelChunk", DecodeStatus: "decoded"},
	}
	if err := runner.recordForwardTiming(parent, capture.DirectionServerToClient, "upstream-1", 2, 2*time.Millisecond, 3*time.Millisecond, 4*time.Millisecond, 5*time.Millisecond, "explicit", "typed"); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	var timing capture.Event
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind == "bridge.forward_timing" {
			timing = event
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if timing.ParentSequence == nil || *timing.ParentSequence != 42 {
		t.Fatalf("timing parent sequence = %v, want 42", timing.ParentSequence)
	}
	if timing.Packet == nil || timing.Packet.Name != "LevelChunk" || timing.Hop != 2 {
		t.Fatalf("timing packet metadata = %#v, hop = %d", timing.Packet, timing.Hop)
	}
	var fields map[string]any
	if err := json.Unmarshal(timing.Data, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["write_mode"] != "typed" || fields["flush_mode"] != "explicit" || fields["measurement"] == nil {
		t.Fatalf("timing fields = %#v", fields)
	}
	if !bytes.Contains(timing.Data, []byte(`"write_packet_duration_nano":"4000000"`)) {
		t.Fatalf("timing data = %s", timing.Data)
	}
}
