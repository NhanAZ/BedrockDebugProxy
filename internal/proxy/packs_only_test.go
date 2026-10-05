package proxy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

func TestRunPacksOnlyCapturesUpstreamWithoutListener(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	archive := testURLResourcePackArchiveWith(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), "1.0.0")
	pack, err := resource.Read(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	packets := make(chan uint32, 64)
	upstream, err := (minecraft.ListenConfig{
		AuthenticationDisabled: true,
		ResourcePacks:          []*resource.Pack{pack},
		PacketFunc: func(header packet.Header, _ []byte, _, _ net.Addr) {
			select {
			case packets <- header.PacketID:
			default:
			}
		},
		FetchResourcePacksInfo: func(login.IdentityData, login.ClientData) *packet.ResourcePacksInfo {
			return &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
				UUID:    pack.UUID(),
				Version: pack.Version(),
				Size:    uint64(len(archive)),
			}}}
		},
		FetchResourcePackStack: func(login.IdentityData, login.ClientData) *packet.ResourcePackStack {
			return &packet.ResourcePackStack{TexturePacks: []protocol.StackResourcePack{{
				UUID:    pack.UUID().String(),
				Version: pack.Version(),
			}}}
		},
	}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = upstream.Close() }()

	serverResult := make(chan error, 1)
	go func() {
		accepted, acceptErr := upstream.Accept()
		if acceptErr != nil {
			serverResult <- acceptErr
			return
		}
		conn, ok := accepted.(*minecraft.Conn)
		if !ok {
			_ = accepted.Close()
			serverResult <- io.ErrUnexpectedEOF
			return
		}
		defer func() { _ = conn.Close() }()
		if startErr := conn.StartGame(minecraft.GameData{
			WorldName:       "Packs-only boundary test",
			EntityUniqueID:  1,
			EntityRuntimeID: 1,
		}); startErr == nil {
			serverResult <- errors.New("packs-only client unexpectedly completed world spawn")
			return
		}
		serverResult <- nil
	}()

	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{Values: map[string]string{
		"operation_mode":         "packs_only",
		"client_metadata_source": "gophertunnel_default",
	}})
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = recorder.Close("failed", ctx.Err())
		}
	}()

	runner, err := New(Config{
		ListenAddress:   "not-an-address",
		PacksOnly:       true,
		UpstreamAddress: upstream.Addr().String(),
		Recorder:        recorder,
		Output:          io.Discard,
	})
	if err != nil {
		t.Fatalf("New() rejected packs-only configuration without a listener: %v", err)
	}
	if err := runner.RunPacksOnly(ctx); err != nil {
		var observed []uint32
		for {
			select {
			case packetID := <-packets:
				observed = append(observed, packetID)
			default:
				t.Fatalf("RunPacksOnly(): %v; upstream packet IDs: %v", err, observed)
			}
		}
	}
	select {
	case serverErr := <-serverResult:
		if serverErr != nil {
			t.Fatalf("upstream session: %v", serverErr)
		}
	case <-ctx.Done():
		t.Fatalf("upstream connection did not close: %v", ctx.Err())
	}
	for {
		select {
		case packetID := <-packets:
			if packetID == packet.IDRequestChunkRadius || packetID == packet.IDSetLocalPlayerAsInitialised {
				t.Fatalf("packs-only connection sent world-spawn packet ID %d", packetID)
			}
		default:
			goto spawnPacketsChecked
		}
	}

spawnPacketsChecked:
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	closed = true

	verification, err := capture.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) != 0 {
		t.Fatalf("capture verification issues = %v", verification.Issues)
	}
	if got := verification.Manifest.Options.Values["operation_mode"]; got != "packs_only" {
		t.Fatalf("capture operation_mode = %q, want packs_only", got)
	}
	if got := verification.Manifest.Options.Values["client_metadata_source"]; got != "gophertunnel_default" {
		t.Fatalf("client_metadata_source = %q, want gophertunnel_default", got)
	}
	if _, exists := verification.Manifest.Options.Values["listen_address"]; exists {
		t.Fatal("packs-only capture unexpectedly records a local listen address")
	}

	var fetchComplete bool
	archiveCount := 0
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		switch event.Kind {
		case "resource_pack.fetch_complete":
			var data struct {
				ResourcePackCount int  `json:"resource_pack_count"`
				WorldSpawned      bool `json:"world_spawned"`
			}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return err
			}
			if data.ResourcePackCount != 1 || data.WorldSpawned {
				t.Errorf("fetch-complete data = %#v, want one offered pack and no world spawn", data)
			}
			fetchComplete = true
		case "resource_pack.archive":
			if event.Blob == nil {
				return io.ErrUnexpectedEOF
			}
			captured, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(event.Blob.Path)))
			if err != nil {
				return err
			}
			if got := sha256.Sum256(captured); got != sha256.Sum256(archive) {
				t.Errorf("captured pack archive SHA-256 = %x, want %x", got, sha256.Sum256(archive))
			}
			if event.Blob.Size != int64(len(archive)) {
				t.Errorf("captured pack archive size = %d, want %d", event.Blob.Size, len(archive))
			}
			archiveCount++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !fetchComplete {
		t.Fatal("capture lacks resource_pack.fetch_complete")
	}
	if archiveCount != 1 {
		t.Fatalf("capture has %d archive events, want one for the offered pack", archiveCount)
	}
}

func TestDialerResourcePackExchangeOnlyStopsBeforeWorldStartup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	packetIDs := make(chan uint32, 64)
	upstream, err := (minecraft.ListenConfig{
		AuthenticationDisabled: true,
		PacketFunc: func(header packet.Header, _ []byte, _, _ net.Addr) {
			select {
			case packetIDs <- header.PacketID:
			default:
			}
		},
		FetchResourcePacksInfo: func(login.IdentityData, login.ClientData) *packet.ResourcePacksInfo {
			return &packet.ResourcePacksInfo{}
		},
		FetchResourcePackStack: func(login.IdentityData, login.ClientData) *packet.ResourcePackStack {
			return &packet.ResourcePackStack{}
		},
	}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = upstream.Close() }()

	serverConnReady := make(chan *minecraft.Conn, 1)
	serverResult := make(chan error, 1)
	go func() {
		accepted, acceptErr := upstream.Accept()
		if acceptErr != nil {
			serverResult <- acceptErr
			return
		}
		conn, ok := accepted.(*minecraft.Conn)
		if !ok {
			_ = accepted.Close()
			serverResult <- io.ErrUnexpectedEOF
			return
		}
		serverConnReady <- conn
		startErr := conn.StartGame(minecraft.GameData{
			WorldName:       "Post-pack startup test",
			EntityUniqueID:  1,
			EntityRuntimeID: 1,
		})
		_ = conn.Close()
		serverResult <- startErr
	}()

	client, err := (minecraft.Dialer{ResourcePackExchangeOnly: true}).DialContext(ctx, "raknet", upstream.Addr().String())
	if err != nil {
		t.Fatalf("dial through resource-pack exchange: %v", err)
	}
	defer func() { _ = client.Close() }()
	select {
	case <-serverConnReady:
	case <-ctx.Done():
		t.Fatalf("upstream did not accept completed pack exchange: %v", ctx.Err())
	}

	startGameSeen := false
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for !startGameSeen {
		select {
		case packetID := <-packetIDs:
			startGameSeen = packetID == packet.IDStartGame
		case <-deadline.C:
			t.Fatal("test server did not send StartGame after the resource-pack exchange")
		case <-ctx.Done():
			t.Fatalf("wait for server startup packet: %v", ctx.Err())
		}
	}
	spawnCheck := time.NewTimer(100 * time.Millisecond)
	defer spawnCheck.Stop()
	for {
		select {
		case packetID := <-packetIDs:
			if packetID == packet.IDRequestChunkRadius || packetID == packet.IDSetLocalPlayerAsInitialised {
				t.Fatalf("resource-pack-only dial sent world-spawn packet ID %d", packetID)
			}
		case <-spawnCheck.C:
			goto spawnCheckComplete
		case <-ctx.Done():
			t.Fatalf("checking post-pack packets: %v", ctx.Err())
		}
	}

spawnCheckComplete:

	if err := client.Close(); err != nil {
		t.Fatalf("close pack-only dial: %v", err)
	}
	select {
	case serverErr := <-serverResult:
		if serverErr == nil {
			t.Fatal("test server completed world spawn after the pack-only dial closed")
		}
	case <-ctx.Done():
		t.Fatalf("test server did not stop after pack-only connection closed: %v", ctx.Err())
	}
}
