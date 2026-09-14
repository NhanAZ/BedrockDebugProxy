package minecraft

import (
	"bytes"
	"context"
	"io"
	"log/slog"
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
