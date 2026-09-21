package proxy

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/artifacts"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

func TestUpstreamDialTimeoutErrorPreservesCause(t *testing.T) {
	err := &upstreamDialTimeoutError{timeout: time.Minute, err: context.DeadlineExceeded}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("timeout error did not preserve context deadline cause")
	}
	if !strings.Contains(err.Error(), "upstream login and resource-pack exchange timed out after 1m0s") {
		t.Fatalf("timeout error = %q", err)
	}
}

func TestDecodeRequestChunkRadius(t *testing.T) {
	got, err := decodeRequestChunkRadius([]byte{0x30, 0x1c})
	if err != nil {
		t.Fatalf("decode request chunk radius: %v", err)
	}
	if got != (packet.RequestChunkRadius{ChunkRadius: 24, MaxChunkRadius: 28}) {
		t.Fatalf("decoded request = %#v", got)
	}
	if _, err := decodeRequestChunkRadius([]byte{0x30, 0x1c, 0x00}); err == nil {
		t.Fatal("decode accepted trailing bytes")
	}
	if _, err := decodeRequestChunkRadius([]byte{0x30}); err == nil {
		t.Fatal("decode accepted truncated payload")
	}
}

type chunkRadiusPacketWriter struct {
	packet packet.Packet
	err    error
}

func (w *chunkRadiusPacketWriter) WritePacket(pk packet.Packet) error {
	w.packet = pk
	return w.err
}

func TestForwardRequestChunkRadiusPreservesTypedValues(t *testing.T) {
	writer := new(chunkRadiusPacketWriter)
	want := packet.RequestChunkRadius{ChunkRadius: 24, MaxChunkRadius: 28}
	if err := forwardRequestChunkRadius(writer, want); err != nil {
		t.Fatalf("forward request chunk radius: %v", err)
	}
	got, ok := writer.packet.(*packet.RequestChunkRadius)
	if !ok {
		t.Fatalf("forwarded packet type = %T, want *packet.RequestChunkRadius", writer.packet)
	}
	if *got != want {
		t.Fatalf("forwarded request = %#v, want %#v", *got, want)
	}
}

func TestForwardRequestChunkRadiusReturnsWriteError(t *testing.T) {
	wantErr := errors.New("upstream closed")
	writer := &chunkRadiusPacketWriter{err: wantErr}
	if err := forwardRequestChunkRadius(writer, packet.RequestChunkRadius{ChunkRadius: 24, MaxChunkRadius: 28}); !errors.Is(err, wantErr) {
		t.Fatalf("forward error = %v, want %v", err, wantErr)
	}
}

type listeningAddressWriter struct {
	mu      sync.Mutex
	buffer  strings.Builder
	address chan string
	sent    bool
}

func newListeningAddressWriter() *listeningAddressWriter {
	return &listeningAddressWriter{address: make(chan string, 1)}
}

func TestLinkedConnectionContextCancelsWhenClientDisconnects(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	connection, cancelConnection := context.WithCancel(context.Background())
	linked, release := linkedConnectionContext(parent, connection)
	defer release()

	cancelConnection()
	select {
	case <-linked.Done():
	case <-time.After(time.Second):
		t.Fatal("linked context was not canceled after downstream disconnect")
	}
	if err := linked.Err(); err != context.Canceled {
		t.Fatalf("linked context error = %v, want context.Canceled", err)
	}
}

func integrationPlayerSkin() *packet.PlayerSkin {
	return &packet.PlayerSkin{
		UUID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Skin: protocol.Skin{
			SkinID:                    "integration-skin",
			SkinResourcePatch:         []byte{},
			SkinImageWidth:            1,
			SkinImageHeight:           1,
			SkinData:                  []byte{1, 2, 3, 4},
			Animations:                []protocol.SkinAnimation{},
			CapeImageWidth:            1,
			CapeImageHeight:           1,
			CapeData:                  []byte{5, 6, 7, 8},
			SkinGeometry:              []byte(`{"geometry":{"default":"integration"}}`),
			AnimationData:             []byte{},
			GeometryDataEngineVersion: []byte{},
			CapeID:                    "integration-cape",
			PersonaPieces:             []protocol.PersonaPiece{},
			PieceTintColours:          []protocol.PersonaPieceTintColour{},
			OverrideAppearance:        true,
		},
	}
}

func isIntegrationPlayerSkin(skin *packet.PlayerSkin) bool {
	return reflect.DeepEqual(skin, integrationPlayerSkin())
}

func integrationClientData() login.ClientData {
	return login.ClientData{
		DeviceOS:         protocol.DeviceAndroid,
		GameVersion:      "1.26.51",
		LanguageCode:     "en_GB",
		DeviceID:         login.DeviceID("0123456789abcdef0123456789abcdef"),
		SelfSignedID:     "22222222-2222-4222-8222-222222222222",
		PlayFabID:        "0123456789abcdef",
		CurrentInputMode: 2,
		DefaultInputMode: 2,
		SkinID:           "integration-login-skin",
		SkinImageWidth:   1,
		SkinImageHeight:  1,
		SkinData:         base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4}),
		CapeID:           "integration-login-cape",
		CapeImageWidth:   1,
		CapeImageHeight:  1,
		CapeData:         base64.StdEncoding.EncodeToString([]byte{5, 6, 7, 8}),
		SkinGeometry:     base64.StdEncoding.EncodeToString([]byte(`{"geometry":{"default":"integration"}}`)),
	}
}

func integrationRawPacket(t *testing.T, pk packet.Packet) []byte {
	t.Helper()
	var full bytes.Buffer
	header := packet.Header{PacketID: pk.ID()}
	if err := header.Write(&full); err != nil {
		t.Fatal(err)
	}
	pk.Marshal(minecraft.DefaultProtocol.NewWriter(&full, 0))
	return full.Bytes()
}

func integrationEntityMetadata() protocol.EntityMetadata {
	metadata := protocol.NewEntityMetadata()
	metadata[protocol.EntityDataKeyName] = "integration-entity"
	metadata[protocol.EntityDataKeyScale] = float32(1.25)
	return metadata
}

func integrationItemInstance() protocol.ItemInstance {
	return protocol.ItemInstance{
		StackNetworkID: 7,
		Stack: protocol.ItemStack{
			ItemType:       protocol.ItemType{NetworkID: 1, MetadataValue: 2},
			BlockRuntimeID: 42,
			Count:          2,
			NBTData:        map[string]any{"marker": "integration-inventory"},
			CanBePlacedOn:  []string{"minecraft:stone"},
			CanBreak:       []string{"minecraft:dirt"},
		},
	}
}

func integrationInventoryTransaction() *packet.InventoryTransaction {
	return &packet.InventoryTransaction{TransactionData: &protocol.ReleaseItemTransactionData{
		ActionType:   protocol.ReleaseItemActionRelease,
		HotBarSlot:   3,
		HeldItem:     integrationItemInstance(),
		HeadPosition: mgl32.Vec3{4, 6, 6},
	}}
}

func integrationCreativeContent() *packet.CreativeContent {
	return &packet.CreativeContent{
		Groups: []protocol.CreativeGroup{{
			Category: protocol.CreativeCategoryNature,
			Name:     "integration-creative-group",
		}},
		Items: []protocol.CreativeItem{{
			CreativeItemNetworkID: 7,
			Item: protocol.ItemStack{
				ItemType:      protocol.ItemType{NetworkID: 1},
				Count:         1,
				CanBePlacedOn: []string{},
				CanBreak:      []string{},
			},
			GroupIndex: 0,
		}},
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

	integrationResourcePackStack := &packet.ResourcePackStack{
		TexturePackRequired:          true,
		BaseGameVersion:              "1.26.51-integration-base",
		Experiments:                  []protocol.ExperimentData{{Name: "integration-experiment", Enabled: true}},
		ExperimentsPreviouslyToggled: true,
		IncludeEditorPacks:           true,
	}
	integrationResourcePacksInfo := &packet.ResourcePacksInfo{
		TexturePackRequired:        false,
		HasAddons:                  true,
		HasScripts:                 true,
		ForceDisableVibrantVisuals: true,
		WorldTemplateUUID:          uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		WorldTemplateVersion:       "integration-template",
	}
	integrationItemRegistry := &packet.ItemRegistry{Items: []protocol.ItemEntry{{
		Name:           "integration:item",
		RuntimeID:      42,
		ComponentBased: true,
		Version:        1,
		Data: map[string]any{
			"components": map[string]any{
				"integration:first":  "one",
				"integration:second": "two",
			},
		},
	}}}
	integrationItemRegistryRaw := integrationRawPacket(t, integrationItemRegistry)
	integrationStartGame := &packet.StartGame{
		WorldName:                "BedrockDebugProxy complete integration",
		EntityUniqueID:           1,
		EntityRuntimeID:          1,
		Generator:                2,
		EducationFeaturesEnabled: false,
		MultiPlayerGame:          false,
		LANBroadcastEnabled:      false,
		CommandsEnabled:          false,
		LevelID:                  "integration-level",
		EnchantmentSeed:          17,
		GameVersion:              "1.26.51",
		ServerAuthoritativeSound: true,
		ServerID:                 "integration-server",
	}
	integrationStartGameRaw := integrationRawPacket(t, integrationStartGame)
	integrationJigsaw := &packet.JigsawStructureData{StructureData: map[string]any{
		"jigsaws": []any{map[string]any{"identifier": "integration:jigsaw"}},
	}}
	integrationJigsawRaw := integrationRawPacket(t, integrationJigsaw)
	integrationVoxelShapes := &packet.VoxelShapes{
		NameMap:          []protocol.VoxelShapeNameEntry{{Name: "integration:shape", ID: 7}},
		CustomShapeCount: 7,
	}
	integrationVoxelShapesRaw := integrationRawPacket(t, integrationVoxelShapes)
	upstreamChunkRadiusRequests := make(chan packet.RequestChunkRadius, 4)
	upstreamListener, err := (minecraft.ListenConfig{
		AuthenticationDisabled: true,
		AllowUnknownPackets:    true,
		AllowInvalidPackets:    true,
		PacketFunc: func(header packet.Header, payload []byte, _, _ net.Addr) {
			if header.PacketID != packet.IDRequestChunkRadius {
				return
			}
			request, decodeErr := decodeRequestChunkRadius(payload)
			if decodeErr != nil {
				return
			}
			select {
			case upstreamChunkRadiusRequests <- request:
			default:
			}
		},
		FetchResourcePackStack: func(login.IdentityData, login.ClientData) *packet.ResourcePackStack {
			return integrationResourcePackStack
		},
		FetchResourcePacksInfo: func(login.IdentityData, login.ClientData) *packet.ResourcePacksInfo {
			return integrationResourcePacksInfo
		},
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
		identity := serverConn.IdentityData()
		if identity.Identity != "11111111-1111-1111-1111-111111111111" || identity.DisplayName != "Integration" {
			serverErrors <- fmt.Errorf("upstream identity = %#v, want downstream identity", identity)
			return
		}
		clientData := serverConn.ClientData()
		if clientData.DeviceOS != protocol.DeviceAndroid || clientData.GameVersion != "1.26.51" ||
			clientData.LanguageCode != "en_GB" || clientData.DeviceID != login.DeviceID("0123456789abcdef0123456789abcdef") ||
			clientData.SelfSignedID != "22222222-2222-4222-8222-222222222222" || clientData.PlayFabID != "0123456789abcdef" ||
			clientData.CurrentInputMode != 2 || clientData.DefaultInputMode != 2 ||
			clientData.SkinID != "integration-login-skin" || clientData.SkinData != base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4}) ||
			clientData.CapeID != "integration-login-cape" || clientData.CapeData != base64.StdEncoding.EncodeToString([]byte{5, 6, 7, 8}) {
			serverErrors <- fmt.Errorf("upstream client data did not preserve downstream metadata: %#v", clientData)
			return
		}
		if startErr := serverConn.StartGame(minecraft.GameData{
			WorldName:                        "BedrockDebugProxy integration",
			EntityUniqueID:                   1,
			EntityRuntimeID:                  1,
			SuppressGeneratedCreativeContent: true,
			RawProtocolID:                    serverConn.Proto().ID(),
			RawItemRegistry:                  append([]byte(nil), integrationItemRegistryRaw...),
			RawStartGame:                     append([]byte(nil), integrationStartGameRaw...),
			RawJigsawStructureData:           append([]byte(nil), integrationJigsawRaw...),
			RawVoxelShapes:                   append([]byte(nil), integrationVoxelShapesRaw...),
			StartGamePacket:                  integrationStartGame,
			JigsawStructureData:              integrationJigsaw,
			VoxelShapes:                      integrationVoxelShapes,
		}); startErr != nil {
			serverErrors <- fmt.Errorf("start upstream game: %w", startErr)
			return
		}
		for index, want := range []packet.RequestChunkRadius{
			{ChunkRadius: 16, MaxChunkRadius: 16},
			{ChunkRadius: 24, MaxChunkRadius: 28},
		} {
			select {
			case got := <-upstreamChunkRadiusRequests:
				if got != want {
					serverErrors <- fmt.Errorf("upstream request chunk radius %d = %#v, want %#v", index+1, got, want)
					return
				}
			case <-time.After(5 * time.Second):
				serverErrors <- fmt.Errorf("timed out waiting for upstream request chunk radius %d", index+1)
				return
			}
		}
		if writeErr := serverConn.WritePacket(integrationCreativeContent()); writeErr != nil {
			serverErrors <- fmt.Errorf("write upstream creative content: %w", writeErr)
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
				if typed.EntityRuntimeID != 1 || typed.Position != (mgl32.Vec3{4, 5, 6}) || typed.Mode != packet.MoveModeNormal || !typed.OnGround {
					serverErrors <- fmt.Errorf("upstream received changed player movement: %#v", typed)
					return
				}
				expectedClientPackets["MovePlayer"] = true
			case *packet.SubChunkRequest:
				if typed.Dimension != 0 || typed.Position != (protocol.SubChunkPos{0, 0, 0}) || !reflect.DeepEqual(typed.Offsets, []protocol.SubChunkOffset{{0, 0, 0}}) {
					serverErrors <- fmt.Errorf("upstream received changed sub-chunk request: %#v", typed)
					return
				}
				expectedClientPackets["SubChunkRequest"] = true
			case *packet.InventoryTransaction:
				release, ok := typed.TransactionData.(*protocol.ReleaseItemTransactionData)
				if !ok {
					serverErrors <- fmt.Errorf("upstream received inventory transaction type %T", typed.TransactionData)
					return
				}
				if release.ActionType != protocol.ReleaseItemActionRelease || release.HotBarSlot != 3 || release.HeadPosition != (mgl32.Vec3{4, 6, 6}) || !reflect.DeepEqual(release.HeldItem, integrationItemInstance()) {
					serverErrors <- fmt.Errorf("upstream received changed inventory transaction: release=%#v item=%#v", release, release.HeldItem)
					return
				}
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
				EntityMetadata:  integrationEntityMetadata(),
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
				Layer:             1,
			},
			&packet.InventoryContent{
				WindowID: 0,
				Content:  []protocol.ItemInstance{integrationItemInstance()},
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
	views, err := artifacts.Start(root, protocol.CurrentProtocol, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recorder.Close("closed", nil); _, _ = views.Finish() })
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
		RequestChunkRadiusProvider: func() (packet.RequestChunkRadius, bool) {
			return packet.RequestChunkRadius{ChunkRadius: 24, MaxChunkRadius: 28}, true
		},
	}).DialContext(ctx, "raknet", proxyAddress)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.DoSpawnContext(ctx); err != nil {
		_ = client.Close()
		t.Fatal(err)
	}
	startGamePacket := client.GameData().StartGamePacket
	if startGamePacket == nil || startGamePacket.WorldName != "BedrockDebugProxy complete integration" ||
		startGamePacket.Generator != 2 || startGamePacket.EducationFeaturesEnabled || startGamePacket.MultiPlayerGame ||
		startGamePacket.LANBroadcastEnabled || startGamePacket.CommandsEnabled || startGamePacket.LevelID != "integration-level" ||
		startGamePacket.EnchantmentSeed != 17 || startGamePacket.GameVersion != "1.26.51" ||
		!startGamePacket.ServerAuthoritativeSound || startGamePacket.ServerID != "integration-server" {
		_ = client.Close()
		t.Fatalf("downstream StartGame did not preserve complete packet: %#v", startGamePacket)
	}
	if clientData := client.GameData(); clientData.JigsawStructureData == nil ||
		!reflect.DeepEqual(clientData.JigsawStructureData.StructureData, map[string]any{
			"jigsaws": []any{map[string]any{"identifier": "integration:jigsaw"}},
		}) || clientData.VoxelShapes == nil ||
		!reflect.DeepEqual(clientData.VoxelShapes.NameMap, []protocol.VoxelShapeNameEntry{{Name: "integration:shape", ID: 7}}) ||
		clientData.VoxelShapes.CustomShapeCount != 7 {
		_ = client.Close()
		t.Fatalf("downstream startup world packets were not preserved: jigsaw=%#v voxel=%#v", clientData.JigsawStructureData, clientData.VoxelShapes)
	}
	stack := client.GameData().ResourcePackStack
	stackPacksEqual := stack != nil && len(stack.TexturePacks) == len(integrationResourcePackStack.TexturePacks) &&
		(reflect.DeepEqual(stack.TexturePacks, integrationResourcePackStack.TexturePacks) || len(stack.TexturePacks) == 0)
	if stack == nil || !stackPacksEqual ||
		stack.TexturePackRequired != integrationResourcePackStack.TexturePackRequired ||
		stack.BaseGameVersion != integrationResourcePackStack.BaseGameVersion ||
		!reflect.DeepEqual(stack.Experiments, integrationResourcePackStack.Experiments) ||
		stack.ExperimentsPreviouslyToggled != integrationResourcePackStack.ExperimentsPreviouslyToggled ||
		stack.IncludeEditorPacks != integrationResourcePackStack.IncludeEditorPacks {
		_ = client.Close()
		t.Fatalf("downstream resource-pack stack was not preserved: %#v", stack)
	}
	info := client.GameData().ResourcePacksInfo
	if info == nil || info.TexturePackRequired != integrationResourcePacksInfo.TexturePackRequired ||
		info.HasAddons != integrationResourcePacksInfo.HasAddons || info.HasScripts != integrationResourcePacksInfo.HasScripts ||
		info.ForceDisableVibrantVisuals != integrationResourcePacksInfo.ForceDisableVibrantVisuals ||
		info.WorldTemplateUUID != integrationResourcePacksInfo.WorldTemplateUUID ||
		info.WorldTemplateVersion != integrationResourcePacksInfo.WorldTemplateVersion {
		_ = client.Close()
		t.Fatalf("downstream resource-pack offer metadata was not preserved: %#v", info)
	}
	if !bytes.Equal(client.GameData().RawItemRegistry, integrationItemRegistryRaw) {
		_ = client.Close()
		t.Fatal("downstream ItemRegistry wire bytes were not preserved")
	}
	if !bytes.Equal(client.GameData().RawStartGame, integrationStartGameRaw) ||
		!bytes.Equal(client.GameData().RawJigsawStructureData, integrationJigsawRaw) ||
		!bytes.Equal(client.GameData().RawVoxelShapes, integrationVoxelShapesRaw) {
		_ = client.Close()
		t.Fatal("downstream startup packet wire bytes were not preserved")
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
		integrationInventoryTransaction(),
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
	case serverErr := <-serverErrors:
		_ = client.Close()
		if serverErr == nil {
			t.Fatal("upstream server stopped before observing representative client packets")
		}
		t.Fatal(serverErr)
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
		"CreativeContent":  false,
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
		case *packet.JigsawStructureData, *packet.VoxelShapes:
			_ = client.Close()
			t.Fatalf("downstream received duplicate startup packet after spawn: %T", typed)
		case *packet.CreativeContent:
			if !reflect.DeepEqual(typed, integrationCreativeContent()) {
				_ = client.Close()
				t.Fatalf("downstream creative content was changed: %#v", typed)
			}
			expectedServerPackets["CreativeContent"] = true
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
			if typed.EntityUniqueID != 2 || typed.EntityRuntimeID != 2 || typed.EntityType != "minecraft:pig" || typed.Position != (mgl32.Vec3{1, 2, 3}) || !reflect.DeepEqual(typed.EntityMetadata, integrationEntityMetadata()) {
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
			if typed.Position != (protocol.BlockPos{1, 64, 2}) || typed.NewBlockRuntimeID != 42 || typed.Flags != packet.BlockUpdateNetwork || typed.Layer != 1 {
				_ = client.Close()
				t.Fatalf("downstream block update was changed: %#v", typed)
			}
			expectedServerPackets["UpdateBlock"] = true
		case *packet.InventoryContent:
			if typed.WindowID != 0 || len(typed.Content) != 1 || !reflect.DeepEqual(typed.Content[0], integrationItemInstance()) {
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
	viewStatus, viewErr := views.Finish()
	if viewErr != nil || viewStatus.State != "complete" || viewStatus.LastSequence != verification.Events {
		viewErrors, _ := os.ReadFile(filepath.Join(root, "artifacts/errors.jsonl"))
		t.Fatalf("artifact view = %+v, error = %v, details = %s", viewStatus, viewErr, viewErrors)
	}
	for _, path := range []string{"skins/index.jsonl", "observations/entities.jsonl", "observations/world.jsonl", "observations/blocks.jsonl", "observations/inventory.jsonl"} {
		info, err := os.Stat(filepath.Join(root, "artifacts", path))
		if err != nil || info.Size() == 0 {
			t.Fatalf("missing live artifact %s: %v", path, err)
		}
	}
	var negotiated, spawned, closed bool
	var negotiatedClientData bool
	var connectionViews, loginSkinMetadataViews, gameDataViews int
	type transportMetadataState struct {
		encryptionEnabled     bool
		compressionNegotiated bool
	}
	transportMetadataObserved := map[string]transportMetadataState{}
	earlyTransportMetadataObserved := map[string]bool{}
	type decodedPacketKey struct {
		direction capture.Direction
		name      string
	}
	expectedDecoded := map[decodedPacketKey][][]byte{
		{capture.DirectionClientToServer, "Emote"}:                {[]byte("client-to-server")},
		{capture.DirectionClientToServer, "PlayerSkin"}:           {[]byte("integration-skin"), []byte(`"preview_hex":"01020304"`), []byte(`"preview_hex":"05060708"`)},
		{capture.DirectionClientToServer, "MovePlayer"}:           {[]byte(`"Position":[4,5,6]`)},
		{capture.DirectionClientToServer, "SubChunkRequest"}:      nil,
		{capture.DirectionClientToServer, "InventoryTransaction"}: {[]byte("integration-inventory"), []byte(`"HotBarSlot":3`)},
		{capture.DirectionServerToClient, "Text"}:                 {[]byte("server-to-client")},
		{capture.DirectionServerToClient, "CreativeContent"}:      {[]byte("integration-creative-group")},
		{capture.DirectionServerToClient, "PlayerSkin"}:           {[]byte("integration-skin"), []byte(`"preview_hex":"01020304"`), []byte(`"preview_hex":"05060708"`)},
		{capture.DirectionServerToClient, "AddActor"}:             {[]byte("integration-entity")},
		{capture.DirectionServerToClient, "LevelChunk"}:           {[]byte(`"preview_hex":"010203"`)},
		{capture.DirectionServerToClient, "UpdateBlock"}:          {[]byte(`"NewBlockRuntimeID":42`), []byte(`"Layer":1`)},
		{capture.DirectionServerToClient, "InventoryContent"}:     {[]byte("integration-inventory"), []byte("minecraft:stone"), []byte("minecraft:dirt")},
	}
	decodedObserved := make(map[decodedPacketKey]bool, len(expectedDecoded))
	timingObserved := make(map[decodedPacketKey]bool)
	type rawPacketKey struct {
		direction capture.Direction
		id        uint32
	}
	expectedRaw := map[rawPacketKey]struct{}{
		{capture.DirectionClientToServer, packet.IDEmote}:                {},
		{capture.DirectionClientToServer, packet.IDPlayerSkin}:           {},
		{capture.DirectionClientToServer, packet.IDMovePlayer}:           {},
		{capture.DirectionClientToServer, packet.IDSubChunkRequest}:      {},
		{capture.DirectionClientToServer, packet.IDInventoryTransaction}: {},
		{capture.DirectionServerToClient, packet.IDResourcePacksInfo}:    {},
		{capture.DirectionServerToClient, packet.IDResourcePackStack}:    {},
		{capture.DirectionServerToClient, packet.IDStartGame}:            {},
		{capture.DirectionServerToClient, packet.IDItemRegistry}:         {},
		{capture.DirectionServerToClient, packet.IDJigsawStructureData}:  {},
		{capture.DirectionServerToClient, packet.IDVoxelShapes}:          {},
		{capture.DirectionServerToClient, packet.IDCreativeContent}:      {},
		{capture.DirectionServerToClient, packet.IDText}:                 {},
		{capture.DirectionServerToClient, packet.IDPlayerSkin}:           {},
		{capture.DirectionServerToClient, packet.IDAddActor}:             {},
		{capture.DirectionServerToClient, packet.IDLevelChunk}:           {},
		{capture.DirectionServerToClient, packet.IDUpdateBlock}:          {},
		{capture.DirectionServerToClient, packet.IDInventoryContent}:     {},
	}
	rawHashes := make(map[rawPacketKey]map[string]map[string]struct{}, len(expectedRaw))
	var downstreamPacksPolicyPreserved bool
	expectedClientData := integrationClientData()
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		switch event.Kind {
		case "transport.payload":
			var data struct {
				EncryptionState  string `json:"bedrock_encryption_state"`
				CompressionState string `json:"bedrock_compression_state"`
			}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return err
			}
			state := transportMetadataObserved[event.Channel]
			state.encryptionEnabled = state.encryptionEnabled || data.EncryptionState == "enabled"
			state.compressionNegotiated = state.compressionNegotiated || data.CompressionState == "negotiated"
			transportMetadataObserved[event.Channel] = state
			if !negotiated && data.CompressionState == "negotiated" {
				earlyTransportMetadataObserved[event.Channel] = true
			}
		case "session.negotiated":
			negotiated = true
			type networkSettingsView struct {
				CompressionThreshold    uint16  `json:"compression_threshold"`
				CompressionAlgorithm    uint16  `json:"compression_algorithm"`
				ClientThrottle          bool    `json:"client_throttle"`
				ClientThrottleThreshold uint8   `json:"client_throttle_threshold"`
				ClientThrottleScalar    float32 `json:"client_throttle_scalar"`
			}
			var data struct {
				DownstreamVersion         string              `json:"downstream_client_data_game_version"`
				UpstreamVersion           string              `json:"upstream_client_data_game_version"`
				VersionMatch              bool                `json:"client_data_game_version_match"`
				DownstreamOS              protocol.DeviceOS   `json:"downstream_client_data_device_os"`
				UpstreamOS                protocol.DeviceOS   `json:"upstream_client_data_device_os"`
				OSMatch                   bool                `json:"client_data_device_os_match"`
				DownstreamNetworkSettings networkSettingsView `json:"downstream_network_settings"`
				UpstreamNetworkSettings   networkSettingsView `json:"upstream_network_settings"`
				NetworkSettingsAvailable  bool                `json:"network_settings_available"`
				NetworkSettingsMatch      bool                `json:"network_settings_match"`
			}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return err
			}
			if data.DownstreamVersion != expectedClientData.GameVersion || data.UpstreamVersion != expectedClientData.GameVersion || !data.VersionMatch || data.DownstreamOS != expectedClientData.DeviceOS || data.UpstreamOS != expectedClientData.DeviceOS || !data.OSMatch || !data.NetworkSettingsAvailable || !data.NetworkSettingsMatch || data.DownstreamNetworkSettings != data.UpstreamNetworkSettings {
				return fmt.Errorf("session negotiated client data = %+v", data)
			}
			negotiatedClientData = true
		case "session.connection_metadata":
			connectionViews++
			if bytes.Contains(event.Data, []byte(expectedClientData.SkinID)) && bytes.Contains(event.Data, []byte(expectedClientData.SkinData)) &&
				bytes.Contains(event.Data, []byte(expectedClientData.CapeID)) && bytes.Contains(event.Data, []byte(expectedClientData.CapeData)) &&
				bytes.Contains(event.Data, []byte(expectedClientData.SkinGeometry)) {
				loginSkinMetadataViews++
			}
		case "session.game_data":
			gameDataViews++
		case "session.spawned":
			spawned = true
		case "session.close":
			closed = true
		case "packet.raw":
			if event.Packet != nil {
				if event.Channel == "downstream" && event.Direction == capture.DirectionServerToClient && event.Packet.ID == packet.IDResourcePacksInfo {
					if event.Blob == nil {
						return fmt.Errorf("downstream ResourcePacksInfo has no raw payload")
					}
					payload, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(event.Blob.Path)))
					if err != nil {
						return err
					}
					info, err := decodeResourcePacksInfo(payload)
					if err != nil {
						return err
					}
					if info.TexturePackRequired != integrationResourcePacksInfo.TexturePackRequired ||
						info.HasAddons != integrationResourcePacksInfo.HasAddons ||
						info.HasScripts != integrationResourcePacksInfo.HasScripts ||
						info.ForceDisableVibrantVisuals != integrationResourcePacksInfo.ForceDisableVibrantVisuals ||
						info.WorldTemplateUUID != integrationResourcePacksInfo.WorldTemplateUUID ||
						info.WorldTemplateVersion != integrationResourcePacksInfo.WorldTemplateVersion {
						return fmt.Errorf("downstream ResourcePacksInfo metadata = %#v, want %#v", info, integrationResourcePacksInfo)
					}
					downstreamPacksPolicyPreserved = true
				}
				key := rawPacketKey{event.Direction, event.Packet.ID}
				if _, ok := expectedRaw[key]; ok {
					if event.Blob == nil || event.Blob.Size == 0 || event.Blob.Representation != "bedrock_packet_payload" {
						return fmt.Errorf("representative raw packet %d in %s has invalid payload blob", event.Packet.ID, event.Direction)
					}
					if event.Channel != "downstream" && event.Channel != "upstream" {
						return fmt.Errorf("representative raw packet %d has unexpected channel %q", event.Packet.ID, event.Channel)
					}
					if rawHashes[key] == nil {
						rawHashes[key] = map[string]map[string]struct{}{}
					}
					if rawHashes[key][event.Channel] == nil {
						rawHashes[key][event.Channel] = map[string]struct{}{}
					}
					rawHashes[key][event.Channel][event.Blob.SHA256] = struct{}{}
				}
			}
		case "packet.decoded":
			if event.Packet != nil {
				key := decodedPacketKey{event.Direction, event.Packet.Name}
				if markers, ok := expectedDecoded[key]; ok {
					matches := true
					for _, marker := range markers {
						matches = matches && bytes.Contains(event.Data, marker)
					}
					decodedObserved[key] = decodedObserved[key] || matches
				}
			}
		case "bridge.forward_timing":
			if event.Packet == nil || event.ParentSequence == nil || !json.Valid(event.Data) {
				return fmt.Errorf("forward timing event has incomplete ancestry or data: %#v", event)
			}
			timingObserved[decodedPacketKey{event.Direction, event.Packet.Name}] = true
		case "packet.decode_error", "capture.view_error":
			return fmt.Errorf("unexpected %s event", event.Kind)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !negotiated || !negotiatedClientData || !spawned || !closed || connectionViews != 2 || loginSkinMetadataViews != 2 || gameDataViews != 1 {
		t.Fatalf("session evidence negotiated=%t negotiated_client_data=%t spawned=%t closed=%t connection_views=%d login_skin_metadata_views=%d game_data_views=%d", negotiated, negotiatedClientData, spawned, closed, connectionViews, loginSkinMetadataViews, gameDataViews)
	}
	if downstream := transportMetadataObserved["downstream"]; !downstream.encryptionEnabled || !downstream.compressionNegotiated {
		t.Fatalf("downstream transport metadata = %#v", downstream)
	}
	if upstream := transportMetadataObserved["upstream"]; !upstream.encryptionEnabled || !upstream.compressionNegotiated {
		t.Fatalf("upstream transport metadata = %#v", upstream)
	}
	for _, channel := range []string{"downstream", "upstream"} {
		if !earlyTransportMetadataObserved[channel] {
			t.Errorf("%s transport metadata did not expose negotiated compression before session.negotiated", channel)
		}
	}
	if !downstreamPacksPolicyPreserved {
		t.Error("downstream resource-pack offer policy was not captured")
	}
	for key := range expectedDecoded {
		if !decodedObserved[key] {
			t.Errorf("decoded representative packet %s in %s direction did not preserve its expected fields", key.name, key.direction)
		}
	}
	for _, key := range []decodedPacketKey{
		{capture.DirectionClientToServer, "MovePlayer"},
		{capture.DirectionServerToClient, "LevelChunk"},
		{capture.DirectionServerToClient, "UpdateBlock"},
	} {
		if !timingObserved[key] {
			t.Errorf("forward timing was not recorded for %s in %s direction", key.name, key.direction)
		}
	}
	for key := range expectedRaw {
		channels := rawHashes[key]
		downstreamHashes := channels["downstream"]
		upstreamHashes := channels["upstream"]
		if len(downstreamHashes) == 0 || len(upstreamHashes) == 0 {
			t.Errorf("raw representative packet %d in %s direction was not captured on both proxy channels", key.id, key.direction)
			continue
		}
		matched := false
		for hash := range downstreamHashes {
			if _, ok := upstreamHashes[hash]; ok {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("raw representative packet %d in %s direction changed between proxy channels", key.id, key.direction)
		}
	}
}

func TestRunnerPreservesMixedResourcePackOfferOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	firstUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	middleUUID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	lastUUID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	archives := map[string][]byte{
		"/first.zip":  testURLResourcePackArchiveWith(t, firstUUID, "1.0.0"),
		"/middle.zip": testURLResourcePackArchiveWith(t, middleUUID, "2.0.0"),
		"/last.zip":   testURLResourcePackArchiveWith(t, lastUUID, "3.0.0"),
	}
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		archive, ok := archives[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(archive)
	}))
	defer httpServer.Close()

	firstPack, err := resource.ReadURL(httpServer.URL + "/first.zip")
	if err != nil {
		t.Fatal(err)
	}
	middlePack, err := resource.Read(bytes.NewReader(archives["/middle.zip"]))
	if err != nil {
		t.Fatal(err)
	}
	lastPack, err := resource.ReadURL(httpServer.URL + "/last.zip")
	if err != nil {
		t.Fatal(err)
	}
	packs := []*resource.Pack{firstPack, middlePack, lastPack}
	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{
		{UUID: firstPack.UUID(), Version: firstPack.Version(), Size: uint64(len(archives["/first.zip"])), ContentKey: "first-content-key", ContentIdentity: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", SubPackName: "first-subpack", HasScripts: true, AddonPack: true, RTXEnabled: true, DownloadURL: firstPack.DownloadURL()},
		{UUID: middlePack.UUID(), Version: middlePack.Version(), Size: uint64(len(archives["/middle.zip"])), ContentKey: "middle-content-key", ContentIdentity: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", SubPackName: "middle-subpack", HasScripts: true},
		{UUID: lastPack.UUID(), Version: lastPack.Version(), Size: uint64(len(archives["/last.zip"])), ContentIdentity: "cccccccc-cccc-4ccc-8ccc-cccccccccccc", SubPackName: "last-subpack", RTXEnabled: true, DownloadURL: lastPack.DownloadURL()},
	}}
	stack := &packet.ResourcePackStack{
		BaseGameVersion: "1.26.51-mixed-pack-test",
		TexturePacks: []protocol.StackResourcePack{
			{UUID: firstUUID.String(), Version: firstPack.Version()},
			{UUID: middleUUID.String(), Version: middlePack.Version()},
			{UUID: lastUUID.String(), Version: lastPack.Version()},
		},
	}
	upstreamListener, err := (minecraft.ListenConfig{
		AuthenticationDisabled: true,
		AllowUnknownPackets:    true,
		AllowInvalidPackets:    true,
		ResourcePacks:          packs,
		FetchResourcePacks: func(login.IdentityData, login.ClientData, []*resource.Pack) []*resource.Pack {
			return packs
		},
		FetchResourcePacksInfo: func(login.IdentityData, login.ClientData) *packet.ResourcePacksInfo {
			return info
		},
		FetchResourcePackStack: func(login.IdentityData, login.ClientData) *packet.ResourcePackStack {
			return stack
		},
	}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = upstreamListener.Close() }()

	serverErrors := make(chan error, 1)
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		accepted, acceptErr := upstreamListener.Accept()
		if acceptErr != nil {
			serverErrors <- acceptErr
			return
		}
		serverConn, ok := accepted.(*minecraft.Conn)
		if !ok {
			_ = accepted.Close()
			serverErrors <- fmt.Errorf("accepted unexpected connection type %T", accepted)
			return
		}
		defer func() { _ = serverConn.Close() }()
		serverErrors <- serverConn.StartGame(mixedResourcePackGameData())
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
	defer func() {
		cancel()
		select {
		case <-runnerDone:
		case <-time.After(5 * time.Second):
			t.Errorf("proxy runner did not stop during cleanup")
		}
		_ = recorder.Close("closed", nil)
	}()

	var proxyAddress string
	select {
	case proxyAddress = <-output.address:
	case <-ctx.Done():
		t.Fatalf("wait for proxy listener: %v", ctx.Err())
	}
	client, err := (minecraft.Dialer{
		IdentityData: login.IdentityData{Identity: "11111111-1111-1111-1111-111111111111", DisplayName: "MixedPackTest"},
		ClientData:   integrationClientData(),
		DownloadResourcePack: func(uuid.UUID, string, int, int) bool {
			return false
		},
	}).DialContext(ctx, "raknet", proxyAddress)
	if err != nil {
		output.mu.Lock()
		proxyOutput := output.buffer.String()
		output.mu.Unlock()
		select {
		case serverErr := <-serverErrors:
			t.Fatalf("%v; upstream error: %v; proxy output: %s; capture: %s", err, serverErr, proxyOutput, root)
		case runnerErr := <-runnerErrors:
			t.Fatalf("%v; proxy error: %v; proxy output: %s; capture: %s", err, runnerErr, proxyOutput, root)
		case <-time.After(time.Second):
			t.Fatalf("%v; proxy output: %s; capture: %s", err, proxyOutput, root)
		}
	}
	defer func() { _ = client.Close() }()
	if err := client.DoSpawnContext(ctx); err != nil {
		t.Fatal(err)
	}

	gotInfo := client.GameData().ResourcePacksInfo
	if gotInfo == nil || len(gotInfo.TexturePacks) != len(info.TexturePacks) {
		t.Fatalf("downstream mixed offer = %#v, want %d entries", gotInfo, len(info.TexturePacks))
	}
	for index, want := range info.TexturePacks {
		got := gotInfo.TexturePacks[index]
		if got.UUID != want.UUID || got.Version != want.Version || got.Size != want.Size || got.DownloadURL != "" {
			t.Fatalf("downstream offer entry %d = %#v, want UUID/version/size %s/%s/%d with empty URL", index, got, want.UUID, want.Version, want.Size)
		}
		if got.ContentKey != want.ContentKey || got.ContentIdentity != want.ContentIdentity || got.SubPackName != want.SubPackName || got.HasScripts != want.HasScripts || got.AddonPack != want.AddonPack || got.RTXEnabled != want.RTXEnabled {
			t.Fatalf("downstream offer metadata %d = %#v, want key/identity/subpack/flags from %#v", index, got, want)
		}
	}
	gotStack := client.GameData().ResourcePackStack
	if gotStack == nil || !reflect.DeepEqual(gotStack.TexturePacks, stack.TexturePacks) || gotStack.BaseGameVersion != stack.BaseGameVersion {
		t.Fatalf("downstream mixed stack = %#v, want %#v", gotStack, stack)
	}

	cancel()
	select {
	case err := <-runnerErrors:
		if err != nil {
			t.Fatalf("proxy run error: %v", err)
		}
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			t.Fatalf("wait for proxy shutdown: %v", ctx.Err())
		}
	}
	select {
	case err := <-serverErrors:
		if err != nil {
			t.Fatalf("upstream mixed-pack session: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("upstream mixed-pack session did not finish")
	}
	select {
	case <-runnerDone:
	case <-time.After(5 * time.Second):
		t.Fatal("proxy runner did not finish shutting down")
	}
	select {
	case <-serverDone:
	case <-time.After(5 * time.Second):
		t.Fatal("upstream mixed-pack connection did not close")
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
}

func mixedResourcePackGameData() minecraft.GameData {
	return minecraft.GameData{WorldName: "Mixed resource-pack integration", EntityUniqueID: 1, EntityRuntimeID: 1}
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
