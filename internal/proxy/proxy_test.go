package proxy

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type listeningAddressWriter struct {
	mu      sync.Mutex
	buffer  strings.Builder
	address chan string
	sent    bool
}

func newListeningAddressWriter() *listeningAddressWriter {
	return &listeningAddressWriter{address: make(chan string, 1)}
}

func integrationPlayerSkin() *packet.PlayerSkin {
	return &packet.PlayerSkin{
		UUID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Skin: protocol.Skin{
			SkinID:             "integration-skin",
			SkinImageWidth:     1,
			SkinImageHeight:    1,
			SkinData:           []byte{1, 2, 3, 4},
			CapeImageWidth:     1,
			CapeImageHeight:    1,
			CapeData:           []byte{5, 6, 7, 8},
			SkinGeometry:       []byte(`{"geometry":{"default":"integration"}}`),
			CapeID:             "integration-cape",
			OverrideAppearance: true,
		},
	}
}

func isIntegrationPlayerSkin(skin *packet.PlayerSkin) bool {
	if skin == nil {
		return false
	}
	want := integrationPlayerSkin()
	return skin.UUID == want.UUID &&
		skin.Skin.SkinID == want.Skin.SkinID &&
		bytes.Equal(skin.Skin.SkinData, want.Skin.SkinData) &&
		bytes.Equal(skin.Skin.CapeData, want.Skin.CapeData) &&
		bytes.Equal(skin.Skin.SkinGeometry, want.Skin.SkinGeometry) &&
		skin.Skin.CapeID == want.Skin.CapeID
}

func integrationClientData() login.ClientData {
	return login.ClientData{
		SkinID:          "integration-login-skin",
		SkinImageWidth:  1,
		SkinImageHeight: 1,
		SkinData:        base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4}),
		CapeID:          "integration-login-cape",
		CapeImageWidth:  1,
		CapeImageHeight: 1,
		CapeData:        base64.StdEncoding.EncodeToString([]byte{5, 6, 7, 8}),
		SkinGeometry:    base64.StdEncoding.EncodeToString([]byte(`{"geometry":{"default":"integration"}}`)),
	}
}

func (w *listeningAddressWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, _ = w.buffer.Write(p)
	if w.sent {
		return len(p), nil
	}
	const prefix = "Listening on "
	const delimiter = " and forwarding"
	text := w.buffer.String()
	start := strings.Index(text, prefix)
	if start == -1 {
		return len(p), nil
	}
	start += len(prefix)
	end := strings.Index(text[start:], delimiter)
	if end == -1 {
		return len(p), nil
	}
	w.address <- text[start : start+end]
	w.sent = true
	return len(p), nil
}

func TestRunnerStopsCleanlyWhenCancelledBeforeAccept(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := New(Config{
		ListenAddress:   "127.0.0.1:0",
		UpstreamAddress: "127.0.0.1:19133",
		Recorder:        recorder,
		Output:          io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := runner.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	verification, err := capture.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) != 0 {
		t.Fatalf("capture issues = %v", verification.Issues)
	}
	if verification.Events < 2 {
		t.Fatalf("event count = %d, want at least 2 lifecycle events", verification.Events)
	}
}

func TestRunnerForwardsBidirectionalPacketsAndCapturesSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	upstreamListener, err := (minecraft.ListenConfig{
		AuthenticationDisabled: true,
		AllowUnknownPackets:    true,
		AllowInvalidPackets:    true,
	}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = upstreamListener.Close() }()

	clientReceived := make(chan struct{})
	serverObserved := make(chan *packet.Emote, 1)
	serverErrors := make(chan error, 1)
	go func() {
		accepted, acceptErr := upstreamListener.Accept()
		if acceptErr != nil {
			serverErrors <- fmt.Errorf("accept proxy connection: %w", acceptErr)
			return
		}
		serverConn, ok := accepted.(*minecraft.Conn)
		if !ok {
			_ = accepted.Close()
			serverErrors <- fmt.Errorf("accepted unexpected connection type %T", accepted)
			return
		}
		defer func() { _ = serverConn.Close() }()
		if startErr := serverConn.StartGame(minecraft.GameData{
			WorldName:       "BedrockDebugProxy integration",
			EntityUniqueID:  1,
			EntityRuntimeID: 1,
		}); startErr != nil {
			serverErrors <- fmt.Errorf("start upstream game: %w", startErr)
			return
		}
		expectedClientPackets := map[string]bool{
			"Emote":                false,
			"PlayerSkin":           false,
			"MovePlayer":           false,
			"SubChunkRequest":      false,
			"InventoryTransaction": false,
		}
		var observedEmote *packet.Emote
		for {
			received, readErr := serverConn.ReadPacket()
			if readErr != nil {
				serverErrors <- fmt.Errorf("read client packet: %w", readErr)
				return
			}
			switch typed := received.(type) {
			case *packet.Emote:
				expectedClientPackets["Emote"] = true
				observedEmote = typed
			case *packet.PlayerSkin:
				if !isIntegrationPlayerSkin(typed) {
					serverErrors <- fmt.Errorf("upstream received unexpected player skin: %#v", typed)
					return
				}
				expectedClientPackets["PlayerSkin"] = true
			case *packet.MovePlayer:
				expectedClientPackets["MovePlayer"] = true
			case *packet.SubChunkRequest:
				expectedClientPackets["SubChunkRequest"] = true
			case *packet.InventoryTransaction:
				expectedClientPackets["InventoryTransaction"] = true
			}
			complete := true
			for _, seen := range expectedClientPackets {
				complete = complete && seen
			}
			if complete {
				break
			}
		}
		if observedEmote == nil || observedEmote.EmoteID != "client-to-server" {
			serverErrors <- fmt.Errorf("upstream did not observe expected emote: %#v", observedEmote)
			return
		}
		serverObserved <- observedEmote
		if writeErr := serverConn.WritePacket(&packet.Text{TextType: packet.TextTypeRaw, Message: "server-to-client"}); writeErr != nil {
			serverErrors <- fmt.Errorf("write server packet: %w", writeErr)
			return
		}
		for _, outgoing := range []packet.Packet{
			integrationPlayerSkin(),
			&packet.AddActor{
				EntityUniqueID:  2,
				EntityRuntimeID: 2,
				EntityType:      "minecraft:pig",
				Position:        mgl32.Vec3{1, 2, 3},
				EntityMetadata:  protocol.NewEntityMetadata(),
			},
			&packet.LevelChunk{
				Position:      protocol.ChunkPos{1, 2},
				Dimension:     0,
				SubChunkCount: 0,
				RawPayload:    []byte{0x01, 0x02, 0x03},
			},
			&packet.UpdateBlock{
				Position:          protocol.BlockPos{1, 64, 2},
				NewBlockRuntimeID: 42,
				Flags:             packet.BlockUpdateNetwork,
			},
			&packet.InventoryContent{
				WindowID: 0,
				Content: []protocol.ItemInstance{{
					StackNetworkID: 7,
					Stack: protocol.ItemStack{
						ItemType: protocol.ItemType{NetworkID: 1},
						Count:    2,
					},
				}},
			},
		} {
			if writeErr := serverConn.WritePacket(outgoing); writeErr != nil {
				serverErrors <- fmt.Errorf("write representative server packet %T: %w", outgoing, writeErr)
				return
			}
		}
		if flushErr := serverConn.Flush(); flushErr != nil {
			serverErrors <- fmt.Errorf("flush server packet: %w", flushErr)
			return
		}
		select {
		case <-clientReceived:
			serverErrors <- nil
		case <-ctx.Done():
			serverErrors <- fmt.Errorf("wait for client receipt: %w", ctx.Err())
		}
	}()

	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{SyncEachEvent: true})
	if err != nil {
		t.Fatal(err)
	}
	output := newListeningAddressWriter()
	runner, err := New(Config{
		ListenAddress:              "127.0.0.1:0",
		UpstreamAddress:            upstreamListener.Addr().String(),
		AllowUnauthenticatedClient: true,
		Recorder:                   recorder,
		Output:                     output,
	})
	if err != nil {
		t.Fatal(err)
	}
	runnerErrors := make(chan error, 1)
	runnerDone := make(chan struct{})
	go func() {
		runnerErrors <- runner.Run(ctx)
		close(runnerDone)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-runnerDone:
		case <-time.After(5 * time.Second):
			t.Errorf("proxy runner did not stop during cleanup")
		}
		_ = recorder.Close("failed", ctx.Err())
	})

	var proxyAddress string
	select {
	case proxyAddress = <-output.address:
	case <-ctx.Done():
		t.Fatalf("wait for proxy listener: %v", ctx.Err())
	}
	client, err := (minecraft.Dialer{
		IdentityData: login.IdentityData{
			Identity:    "11111111-1111-1111-1111-111111111111",
			DisplayName: "Integration",
		},
		ClientData: integrationClientData(),
	}).DialContext(ctx, "raknet", proxyAddress)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.DoSpawnContext(ctx); err != nil {
		_ = client.Close()
		t.Fatal(err)
	}
	for _, outgoing := range []packet.Packet{
		integrationPlayerSkin(),
		&packet.MovePlayer{
			EntityRuntimeID: client.GameData().EntityRuntimeID,
			Position:        mgl32.Vec3{4, 5, 6},
			Mode:            packet.MoveModeNormal,
			OnGround:        true,
		},
		&packet.SubChunkRequest{
			Dimension: 0,
			Offsets:   []protocol.SubChunkOffset{{0, 0, 0}},
			Position:  protocol.SubChunkPos{0, 0, 0},
		},
		&packet.InventoryTransaction{TransactionData: &protocol.NormalTransactionData{}},
		&packet.Emote{
			EntityRuntimeID: client.GameData().EntityRuntimeID,
			EmoteID:         "client-to-server",
		},
	} {
		if err := client.WritePacket(outgoing); err != nil {
			_ = client.Close()
			t.Fatalf("write representative client packet %T: %v", outgoing, err)
		}
	}
	if err := client.Flush(); err != nil {
		_ = client.Close()
		t.Fatal(err)
	}
	select {
	case emote := <-serverObserved:
		if emote.EmoteID != "client-to-server" {
			t.Fatalf("upstream emote ID = %q", emote.EmoteID)
		}
	case <-ctx.Done():
		_ = client.Close()
		t.Fatalf("wait for upstream packet: %v", ctx.Err())
	}
	if err := client.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		_ = client.Close()
		t.Fatal(err)
	}
	expectedServerPackets := map[string]bool{
		"Text":             false,
		"PlayerSkin":       false,
		"AddActor":         false,
		"LevelChunk":       false,
		"UpdateBlock":      false,
		"InventoryContent": false,
	}
	for {
		received, readErr := client.ReadPacket()
		if readErr != nil {
			_ = client.Close()
			t.Fatal(readErr)
		}
		switch typed := received.(type) {
		case *packet.Text:
			if typed.Message != "server-to-client" {
				_ = client.Close()
				t.Fatalf("downstream text = %q", typed.Message)
			}
			expectedServerPackets["Text"] = true
		case *packet.PlayerSkin:
			if !isIntegrationPlayerSkin(typed) {
				_ = client.Close()
				t.Fatalf("downstream player skin did not preserve skin/cape payload: %#v", typed)
			}
			expectedServerPackets["PlayerSkin"] = true
		case *packet.AddActor:
			if typed.EntityRuntimeID != 2 || typed.EntityType != "minecraft:pig" || typed.Position != (mgl32.Vec3{1, 2, 3}) {
				_ = client.Close()
				t.Fatalf("downstream entity was changed: %#v", typed)
			}
			expectedServerPackets["AddActor"] = true
		case *packet.LevelChunk:
			if typed.Position != (protocol.ChunkPos{1, 2}) || !bytes.Equal(typed.RawPayload, []byte{0x01, 0x02, 0x03}) {
				_ = client.Close()
				t.Fatalf("downstream chunk payload was changed: %#v", typed)
			}
			expectedServerPackets["LevelChunk"] = true
		case *packet.UpdateBlock:
			if typed.Position != (protocol.BlockPos{1, 64, 2}) || typed.NewBlockRuntimeID != 42 || typed.Flags != packet.BlockUpdateNetwork {
				_ = client.Close()
				t.Fatalf("downstream block update was changed: %#v", typed)
			}
			expectedServerPackets["UpdateBlock"] = true
		case *packet.InventoryContent:
			if typed.WindowID != 0 || len(typed.Content) != 1 || typed.Content[0].StackNetworkID != 7 || typed.Content[0].Stack.Count != 2 {
				_ = client.Close()
				t.Fatalf("downstream inventory content was changed: %#v", typed)
			}
			expectedServerPackets["InventoryContent"] = true
		}
		complete := true
		for _, seen := range expectedServerPackets {
			complete = complete && seen
		}
		if complete {
			break
		}
	}
	close(clientReceived)
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case serverErr := <-serverErrors:
		if serverErr != nil {
			t.Fatal(serverErr)
		}
	case <-ctx.Done():
		t.Fatalf("wait for upstream shutdown: %v", ctx.Err())
	}
	select {
	case runErr := <-runnerErrors:
		if runErr != nil {
			t.Fatalf("Run() error = %v", runErr)
		}
	case <-ctx.Done():
		t.Fatalf("wait for proxy shutdown: %v", ctx.Err())
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	verification, err := capture.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) != 0 {
		t.Fatalf("capture issues = %v", verification.Issues)
	}
	var negotiated, spawned, closed bool
	var connectionViews, loginSkinMetadataViews, gameDataViews int
	var decodedClientToServer, decodedServerToClient bool
	var rawClientToServer, rawServerToClient bool
	expectedDecoded := map[capture.Direction]map[string]bool{
		capture.DirectionClientToServer: {
			"Emote":                false,
			"PlayerSkin":           false,
			"MovePlayer":           false,
			"SubChunkRequest":      false,
			"InventoryTransaction": false,
		},
		capture.DirectionServerToClient: {
			"Text":             false,
			"PlayerSkin":       false,
			"AddActor":         false,
			"LevelChunk":       false,
			"UpdateBlock":      false,
			"InventoryContent": false,
		},
	}
	expectedRaw := map[capture.Direction]map[uint32]bool{
		capture.DirectionClientToServer: {
			packet.IDEmote:                false,
			packet.IDPlayerSkin:           false,
			packet.IDMovePlayer:           false,
			packet.IDSubChunkRequest:      false,
			packet.IDInventoryTransaction: false,
		},
		capture.DirectionServerToClient: {
			packet.IDText:             false,
			packet.IDPlayerSkin:       false,
			packet.IDAddActor:         false,
			packet.IDLevelChunk:       false,
			packet.IDUpdateBlock:      false,
			packet.IDInventoryContent: false,
		},
	}
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		switch event.Kind {
		case "session.negotiated":
			negotiated = true
		case "session.connection_metadata":
			connectionViews++
			if bytes.Contains(event.Data, []byte("integration-login-skin")) && bytes.Contains(event.Data, []byte("integration-login-cape")) {
				loginSkinMetadataViews++
			}
		case "session.game_data":
			gameDataViews++
		case "session.spawned":
			spawned = true
		case "session.close":
			closed = true
		case "packet.raw":
			rawClientToServer = rawClientToServer || event.Direction == capture.DirectionClientToServer
			rawServerToClient = rawServerToClient || event.Direction == capture.DirectionServerToClient
			if event.Packet != nil {
				if packets := expectedRaw[event.Direction]; packets != nil {
					if _, ok := packets[event.Packet.ID]; ok {
						if event.Blob == nil || event.Blob.Size == 0 {
							return fmt.Errorf("representative raw packet %d has no payload blob", event.Packet.ID)
						}
						packets[event.Packet.ID] = true
					}
				}
			}
		case "packet.decoded":
			if event.Packet != nil {
				if packets := expectedDecoded[event.Direction]; packets != nil {
					if _, ok := packets[event.Packet.Name]; ok {
						packets[event.Packet.Name] = true
					}
				}
			}
			if event.Packet != nil && event.Packet.Name == "Emote" && event.Direction == capture.DirectionClientToServer {
				decodedClientToServer = true
			}
			if event.Packet != nil && event.Packet.Name == "Text" && event.Direction == capture.DirectionServerToClient {
				decodedServerToClient = true
			}
		case "packet.decode_error", "capture.view_error":
			return fmt.Errorf("unexpected %s event", event.Kind)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !negotiated || !spawned || !closed || connectionViews != 2 || loginSkinMetadataViews != 2 || gameDataViews != 1 {
		t.Fatalf("session evidence negotiated=%t spawned=%t closed=%t connection_views=%d login_skin_metadata_views=%d game_data_views=%d", negotiated, spawned, closed, connectionViews, loginSkinMetadataViews, gameDataViews)
	}
	if !decodedClientToServer || !decodedServerToClient || !rawClientToServer || !rawServerToClient {
		t.Fatalf("traffic evidence decoded_c2s=%t decoded_s2c=%t raw_c2s=%t raw_s2c=%t", decodedClientToServer, decodedServerToClient, rawClientToServer, rawServerToClient)
	}
	for direction, packets := range expectedDecoded {
		for name, seen := range packets {
			if !seen {
				t.Errorf("decoded representative packet %s in %s direction was not captured", name, direction)
			}
		}
	}
	for direction, packets := range expectedRaw {
		for id, seen := range packets {
			if !seen {
				t.Errorf("raw representative packet %d in %s direction was not captured", id, direction)
			}
		}
	}
}

func TestRecordStructuredViewIndexesBinaryFields(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	runner := &Runner{config: Config{Recorder: recorder, DecodedBinaryPreviewBytes: 2}}
	value := struct {
		Name string
		Data []byte
	}{Name: "snapshot", Data: []byte{1, 2, 3, 4}}
	if err := runner.recordStructuredView(
		"session.test_snapshot", value, capture.DirectionInternal, "connection-test", "test", "decoded_test", nil,
	); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind != "session.test_snapshot" || event.ConnectionID != "connection-test" {
			t.Fatalf("snapshot event = %#v", event)
		}
		if !json.Valid(event.Data) || !bytes.Contains(event.Data, []byte(`"size":4`)) ||
			!bytes.Contains(event.Data, []byte(`"preview_hex":"0102"`)) || !bytes.Contains(event.Data, []byte(`"omitted_bytes":2`)) {
			t.Fatalf("snapshot data = %s", event.Data)
		}
		if strings.Contains(string(event.Data), "01020304") {
			t.Fatalf("snapshot data embedded the complete binary field: %s", event.Data)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestStructuredConnectionMetadataKeepsDerivedIdentityFields(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	runner := &Runner{config: Config{Recorder: recorder}}
	value := connectionMetadata{
		Role: "downstream", Authenticated: true, ProtocolID: 900, GameVersion: "1.2.3",
		Identity: identityMetadata{
			XUID: "123", Identity: "identity", DisplayName: "Player",
			PlayFabTitleID: "title", PlayFabID: "player",
		},
	}
	if err := runner.recordStructuredView(
		"session.connection_metadata", value, capture.DirectionInternal, "downstream-1", "downstream", "decoded_login_state", nil,
	); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		text := string(event.Data)
		for _, expected := range []string{`"role":"downstream"`, `"protocol_id":900`, `"playfab_title_id":"title"`, `"playfab_id":"player"`} {
			if !strings.Contains(text, expected) {
				t.Fatalf("connection metadata does not contain %s: %s", expected, text)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
