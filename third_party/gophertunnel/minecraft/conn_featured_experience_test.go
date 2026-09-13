package minecraft

import (
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
