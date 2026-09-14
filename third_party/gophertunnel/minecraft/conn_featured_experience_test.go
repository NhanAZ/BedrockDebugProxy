package minecraft

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestHandlePlayStatusCompletesFeaturedExperienceSpawn(t *testing.T) {
	conn := &Conn{
		ctx:   context.Background(),
		log:   slog.Default(),
		proto: DefaultProtocol,
		spawn: make(chan struct{}),
		enc:   packet.NewEncoder(io.Discard),
		hdr:   &packet.Header{},
		gameData: GameData{
			EntityRuntimeID: 42,
		},
	}

	if err := conn.handlePlayStatus(&packet.PlayStatus{Status: packet.PlayStatusPlayerSpawn}); err != nil {
		t.Fatalf("handle PlayStatus: %v", err)
	}
	if !conn.loggedIn {
		t.Fatal("featured-experience spawn did not mark the connection logged in")
	}
	select {
	case <-conn.spawn:
	default:
		t.Fatal("featured-experience spawn did not close the spawn signal")
	}
}

func TestHandleDrainsResourcePacksInfoDeferredBeforePlayStatus(t *testing.T) {
	conn := &Conn{
		ctx:   context.Background(),
		log:   slog.Default(),
		proto: DefaultProtocol,
		pool:  DefaultProtocol.Packets(false),
		spawn: make(chan struct{}),
		enc:   packet.NewEncoder(io.Discard),
		hdr:   &packet.Header{},
	}
	conn.expectedIDs.Store([]uint32{packet.IDPlayStatus})

	resourcePacksInfo := wirePacketData(t, &packet.ResourcePacksInfo{})
	if err := conn.handle(resourcePacksInfo); err != nil {
		t.Fatalf("defer ResourcePacksInfo: %v", err)
	}
	if len(conn.deferredPackets) != 1 {
		t.Fatalf("deferred packet count after out-of-order packet = %d, want 1", len(conn.deferredPackets))
	}

	if err := conn.handle(wirePacketData(t, &packet.PlayStatus{Status: packet.PlayStatusLoginSuccess})); err != nil {
		t.Fatalf("handle PlayStatus: %v", err)
	}
	if len(conn.deferredPackets) != 0 {
		t.Fatalf("deferred packet count after state transition = %d, want 0", len(conn.deferredPackets))
	}
	if !conn.isExpectedPacket(packet.IDResourcePackStack) {
		t.Fatalf("login state did not advance after draining ResourcePacksInfo: %#v", conn.expectedIDs.Load())
	}
}

func TestHandleDrainsExpectedDeferredPacketBehindLaterStatePacket(t *testing.T) {
	conn := &Conn{
		ctx:   context.Background(),
		log:   slog.Default(),
		proto: DefaultProtocol,
		pool:  DefaultProtocol.Packets(false),
		spawn: make(chan struct{}),
		enc:   packet.NewEncoder(io.Discard),
		hdr:   &packet.Header{},
	}
	conn.expectedIDs.Store([]uint32{packet.IDPlayStatus})

	// ResourcePackStack belongs after ResourcePacksInfo. If it arrives first, it must not
	// prevent the earlier ResourcePacksInfo from advancing the login state.
	if err := conn.handle(wirePacketData(t, &packet.ResourcePackStack{})); err != nil {
		t.Fatalf("defer ResourcePackStack: %v", err)
	}
	if err := conn.handle(wirePacketData(t, &packet.ResourcePacksInfo{})); err != nil {
		t.Fatalf("defer ResourcePacksInfo: %v", err)
	}

	if err := conn.handle(wirePacketData(t, &packet.PlayStatus{Status: packet.PlayStatusLoginSuccess})); err != nil {
		t.Fatalf("handle PlayStatus: %v", err)
	}
	if len(conn.deferredPackets) != 0 {
		t.Fatalf("deferred packet count after reordering = %d, want 0", len(conn.deferredPackets))
	}
	if !conn.isExpectedPacket(packet.IDDimensionData) || !conn.isExpectedPacket(packet.IDStartGame) {
		t.Fatalf("login state did not advance through reordered packets: %#v", conn.expectedIDs.Load())
	}
}

func TestHandleAcceptsResourcePackLoginPermutations(t *testing.T) {
	packets := map[string]packet.Packet{
		"play_status": &packet.PlayStatus{Status: packet.PlayStatusLoginSuccess},
		"packs_info":  &packet.ResourcePacksInfo{},
		"pack_stack":  &packet.ResourcePackStack{},
	}
	permutations := [][]string{
		{"play_status", "packs_info", "pack_stack"},
		{"play_status", "pack_stack", "packs_info"},
		{"packs_info", "play_status", "pack_stack"},
		{"packs_info", "pack_stack", "play_status"},
		{"pack_stack", "play_status", "packs_info"},
		{"pack_stack", "packs_info", "play_status"},
	}

	for _, order := range permutations {
		order := order
		t.Run(strings.Join(order, "_"), func(t *testing.T) {
			conn := newLoginTestConn()
			conn.expectedIDs.Store([]uint32{packet.IDPlayStatus})
			for _, name := range order {
				if err := conn.handle(wirePacketData(t, packets[name])); err != nil {
					t.Fatalf("handle %s: %v", name, err)
				}
			}
			if len(conn.deferredPackets) != 0 {
				t.Fatalf("deferred packet count after order %v = %d, want 0", order, len(conn.deferredPackets))
			}
			if !conn.isExpectedPacket(packet.IDDimensionData) || !conn.isExpectedPacket(packet.IDStartGame) {
				t.Fatalf("login state did not reach the world phase after order %v: %#v", order, conn.expectedIDs.Load())
			}
		})
	}
}

func TestHandleKeepsLaterPhasePacketsUntilTheirState(t *testing.T) {
	conn := newLoginTestConn()
	conn.expectedIDs.Store([]uint32{packet.IDPlayStatus})

	// These packets belong to later phases. They must remain lossless while the
	// deferred queue drains the packet that advances the current phase.
	for _, pk := range []packet.Packet{
		&packet.ResourcePackStack{},
		&packet.ItemRegistry{},
		&packet.ResourcePacksInfo{},
	} {
		if err := conn.handle(wirePacketData(t, pk)); err != nil {
			t.Fatalf("defer %T: %v", pk, err)
		}
	}
	if err := conn.handle(wirePacketData(t, &packet.PlayStatus{Status: packet.PlayStatusLoginSuccess})); err != nil {
		t.Fatalf("handle PlayStatus: %v", err)
	}

	// ResourcePacksInfo and ResourcePackStack are now valid and must be
	// consumed, but ItemRegistry is still a later-phase packet and must stay
	// queued until StartGame advances the state machine.
	if len(conn.deferredPackets) != 1 {
		t.Fatalf("deferred packet count after first transition = %d, want 1", len(conn.deferredPackets))
	}
	if !conn.isExpectedPacket(packet.IDDimensionData) || !conn.isExpectedPacket(packet.IDStartGame) {
		t.Fatalf("unexpected state after resource-pack drain: %#v", conn.expectedIDs.Load())
	}

	if err := conn.handle(wirePacketData(t, &packet.StartGame{})); err != nil {
		t.Fatalf("handle StartGame: %v", err)
	}
	if len(conn.deferredPackets) != 0 {
		t.Fatalf("deferred packet count after second transition = %d, want 0", len(conn.deferredPackets))
	}
	if !conn.isExpectedPacket(packet.IDChunkRadiusUpdated) || !conn.isExpectedPacket(packet.IDPlayStatus) {
		t.Fatalf("unexpected state after ItemRegistry drain: %#v", conn.expectedIDs.Load())
	}
}

func newLoginTestConn() *Conn {
	return &Conn{
		ctx:   context.Background(),
		log:   slog.Default(),
		proto: DefaultProtocol,
		pool:  DefaultProtocol.Packets(false),
		spawn: make(chan struct{}),
		enc:   packet.NewEncoder(io.Discard),
		hdr:   &packet.Header{},
	}
}

func wirePacketData(t *testing.T, pk packet.Packet) *packetData {
	t.Helper()
	var full bytes.Buffer
	header := packet.Header{PacketID: pk.ID()}
	if err := header.Write(&full); err != nil {
		t.Fatal(err)
	}
	pk.Marshal(DefaultProtocol.NewWriter(&full, 0))
	data, err := parseData(full.Bytes(), nil)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
