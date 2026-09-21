package minecraft

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
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

func TestHandleNetworkSettingsRetainsNegotiatedValues(t *testing.T) {
	conn := newLoginTestConn()
	conn.dec = packet.NewDecoder(bytes.NewReader(nil))
	want := &packet.NetworkSettings{
		CompressionThreshold:    17,
		CompressionAlgorithm:    packet.CompressionAlgorithmFlate,
		ClientThrottle:          true,
		ClientThrottleThreshold: 9,
		ClientThrottleScalar:    0.5,
	}
	if err := conn.handleNetworkSettings(want); err != nil {
		t.Fatalf("handle NetworkSettings: %v", err)
	}
	got, ok := conn.NetworkSettings()
	if !ok || !reflect.DeepEqual(&got, want) {
		t.Fatalf("retained NetworkSettings = %#v, available=%t, want %#v", got, ok, want)
	}
}

func TestConnEncryptionStateReportsTransportBoundary(t *testing.T) {
	conn := newLoginTestConn()
	if conn.EncryptionEnabled() {
		t.Fatal("new connection reports enabled encryption")
	}
	if conn.EncryptionDisabled() {
		t.Fatal("new connection reports disabled encryption")
	}
	conn.disableEncryption = true
	if !conn.EncryptionDisabled() {
		t.Fatal("disabled transport boundary was not reported")
	}
	if conn.EncryptionEnabled() {
		t.Fatal("disabled transport boundary reports enabled encryption")
	}
}

func TestHandleRequestChunkRadiusCanSuppressGeneratedCreativeContent(t *testing.T) {
	var output bytes.Buffer
	conn := newLoginTestConn()
	conn.enc = packet.NewEncoder(&output)
	conn.gameData.EntityRuntimeID = 42
	conn.gameData.SuppressGeneratedCreativeContent = true

	if err := conn.handleRequestChunkRadius(&packet.RequestChunkRadius{ChunkRadius: 8}); err != nil {
		t.Fatalf("handle RequestChunkRadius: %v", err)
	}
	if err := conn.Flush(); err != nil {
		t.Fatalf("flush spawn packets: %v", err)
	}
	decoded, err := packet.NewDecoder(bytes.NewReader(output.Bytes())).Decode()
	if err != nil {
		t.Fatalf("decode spawn packets: %v", err)
	}
	for _, raw := range decoded {
		data, err := parseData(raw, nil)
		if err != nil {
			t.Fatalf("parse spawn packet: %v", err)
		}
		if data.h.PacketID == packet.IDCreativeContent {
			t.Fatal("suppressed listener still emitted CreativeContent")
		}
	}
}

func TestHandleItemRegistryUsesRequestChunkRadiusProvider(t *testing.T) {
	var output bytes.Buffer
	conn := newLoginTestConn()
	conn.enc = packet.NewEncoder(&output)
	conn.requestChunkRadiusProvider = func() (packet.RequestChunkRadius, bool) {
		return packet.RequestChunkRadius{ChunkRadius: 24, MaxChunkRadius: 28}, true
	}

	if err := conn.handleItemRegistry(&packet.ItemRegistry{}, nil); err != nil {
		t.Fatalf("handle ItemRegistry: %v", err)
	}
	if err := conn.Flush(); err != nil {
		t.Fatalf("flush chunk-radius request: %v", err)
	}
	decoded, err := packet.NewDecoder(bytes.NewReader(output.Bytes())).Decode()
	if err != nil {
		t.Fatalf("decode chunk-radius request: %v", err)
	}
	if len(decoded) != 1 {
		t.Fatalf("decoded packet count = %d, want 1", len(decoded))
	}
	data, err := parseData(decoded[0], nil)
	if err != nil {
		t.Fatalf("parse chunk-radius request: %v", err)
	}
	if data.h.PacketID != packet.IDRequestChunkRadius {
		t.Fatalf("packet ID = %d, want %d", data.h.PacketID, packet.IDRequestChunkRadius)
	}
	got := &packet.RequestChunkRadius{}
	got.Marshal(protocol.NewReader(data.payload, 0, true))
	if got.ChunkRadius != 24 || got.MaxChunkRadius != 28 {
		t.Fatalf("chunk-radius request = %#v, want radius 24 and max 28", got)
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

func TestHandleStartGamePreservesCompletePacketForBridge(t *testing.T) {
	conn := newLoginTestConn()
	conn.expectedIDs.Store([]uint32{packet.IDStartGame})
	original := &packet.StartGame{
		WorldName:                    "complete-start-game",
		Generator:                    3,
		AchievementsDisabled:         false,
		EducationFeaturesEnabled:     false,
		MultiPlayerGame:              false,
		LANBroadcastEnabled:          false,
		CommandsEnabled:              false,
		ExperimentsPreviouslyToggled: true,
		BonusChestEnabled:            true,
		StartWithMapEnabled:          true,
		ServerChunkTickRadius:        7,
		HasLockedBehaviourPack:       true,
		LevelID:                      "level-id",
		TemplateContentIdentity:      "template-id",
		Trial:                        true,
		EnchantmentSeed:              17,
		MultiPlayerCorrelationID:     "correlation-id",
		GameVersion:                  "1.26.51",
		ServerAuthoritativeSound:     true,
		ServerID:                     "server-id",
		ScenarioID:                   "scenario-id",
		WorldID:                      "world-id",
		OwnerID:                      "owner-id",
	}

	raw := wirePacketData(t, original)
	if err := conn.handle(raw); err != nil {
		t.Fatalf("handle StartGame: %v", err)
	}
	got := conn.GameData().StartGamePacket
	if got == nil {
		t.Fatal("GameData.StartGamePacket is nil")
	}
	if !bytes.Equal(conn.GameData().RawStartGame, raw.full) {
		t.Fatal("retained StartGame bytes differ from input")
	}
	for name, value := range map[string][2]any{
		"WorldName":                    {got.WorldName, original.WorldName},
		"Generator":                    {got.Generator, original.Generator},
		"EducationFeaturesEnabled":     {got.EducationFeaturesEnabled, original.EducationFeaturesEnabled},
		"MultiPlayerGame":              {got.MultiPlayerGame, original.MultiPlayerGame},
		"LANBroadcastEnabled":          {got.LANBroadcastEnabled, original.LANBroadcastEnabled},
		"CommandsEnabled":              {got.CommandsEnabled, original.CommandsEnabled},
		"ExperimentsPreviouslyToggled": {got.ExperimentsPreviouslyToggled, original.ExperimentsPreviouslyToggled},
		"BonusChestEnabled":            {got.BonusChestEnabled, original.BonusChestEnabled},
		"StartWithMapEnabled":          {got.StartWithMapEnabled, original.StartWithMapEnabled},
		"ServerChunkTickRadius":        {got.ServerChunkTickRadius, original.ServerChunkTickRadius},
		"HasLockedBehaviourPack":       {got.HasLockedBehaviourPack, original.HasLockedBehaviourPack},
		"LevelID":                      {got.LevelID, original.LevelID},
		"TemplateContentIdentity":      {got.TemplateContentIdentity, original.TemplateContentIdentity},
		"Trial":                        {got.Trial, original.Trial},
		"EnchantmentSeed":              {got.EnchantmentSeed, original.EnchantmentSeed},
		"MultiPlayerCorrelationID":     {got.MultiPlayerCorrelationID, original.MultiPlayerCorrelationID},
		"GameVersion":                  {got.GameVersion, original.GameVersion},
		"ServerAuthoritativeSound":     {got.ServerAuthoritativeSound, original.ServerAuthoritativeSound},
		"ServerID":                     {got.ServerID, original.ServerID},
		"ScenarioID":                   {got.ScenarioID, original.ScenarioID},
		"WorldID":                      {got.WorldID, original.WorldID},
		"OwnerID":                      {got.OwnerID, original.OwnerID},
	} {
		if value[0] != value[1] {
			t.Fatalf("preserved StartGame field %s = %v, want %v", name, value[0], value[1])
		}
	}
}

func TestHandlePreservesStartupWorldPacketsForBridge(t *testing.T) {
	conn := newLoginTestConn()
	conn.expectedIDs.Store([]uint32{packet.IDJigsawStructureData, packet.IDVoxelShapes})
	jigsaw := &packet.JigsawStructureData{StructureData: map[string]any{
		"jigsaws": []any{map[string]any{"identifier": "integration:jigsaw"}},
	}}
	voxelShapes := &packet.VoxelShapes{
		NameMap:          []protocol.VoxelShapeNameEntry{{Name: "integration:shape", ID: 7}},
		CustomShapeCount: 7,
	}

	jigsawRaw := wirePacketData(t, jigsaw)
	if err := conn.handle(jigsawRaw); err != nil {
		t.Fatalf("handle JigsawStructureData: %v", err)
	}
	voxelRaw := wirePacketData(t, voxelShapes)
	if err := conn.handle(voxelRaw); err != nil {
		t.Fatalf("handle VoxelShapes: %v", err)
	}
	data := conn.GameData()
	if data.JigsawStructureData == nil || !reflect.DeepEqual(data.JigsawStructureData.StructureData, jigsaw.StructureData) {
		t.Fatalf("preserved JigsawStructureData = %#v, want %#v", data.JigsawStructureData, jigsaw)
	}
	if data.VoxelShapes == nil || !reflect.DeepEqual(data.VoxelShapes.NameMap, voxelShapes.NameMap) || data.VoxelShapes.CustomShapeCount != voxelShapes.CustomShapeCount {
		t.Fatalf("preserved VoxelShapes = %#v, want %#v", data.VoxelShapes, voxelShapes)
	}
	if !bytes.Equal(data.RawJigsawStructureData, jigsawRaw.full) {
		t.Fatal("retained JigsawStructureData bytes differ from input")
	}
	if !bytes.Equal(data.RawVoxelShapes, voxelRaw.full) {
		t.Fatal("retained VoxelShapes bytes differ from input")
	}
}

func TestHandlePreservesResourcePackStackForBridge(t *testing.T) {
	conn := newLoginTestConn()
	conn.expectedIDs.Store([]uint32{packet.IDResourcePackStack})
	conn.ignoredResourcePacks = []exemptedResourcePack{{uuid: "integration-pack", version: "2"}}
	original := &packet.ResourcePackStack{
		TexturePackRequired:          true,
		TexturePacks:                 []protocol.StackResourcePack{{UUID: "integration-pack", Version: "2"}},
		BaseGameVersion:              "1.26.51-integration-base",
		Experiments:                  []protocol.ExperimentData{{Name: "integration-experiment", Enabled: true}},
		ExperimentsPreviouslyToggled: true,
		IncludeEditorPacks:           true,
	}

	raw := wirePacketData(t, original)
	if err := conn.handle(raw); err != nil {
		t.Fatalf("handle ResourcePackStack: %v", err)
	}
	got := conn.GameData().ResourcePackStack
	if got == nil || !reflect.DeepEqual(got, original) {
		t.Fatalf("preserved ResourcePackStack = %#v, want %#v", got, original)
	}
	if !bytes.Equal(conn.GameData().RawResourcePackStack, raw.full) {
		t.Fatal("retained ResourcePackStack bytes differ from input")
	}

	original.TexturePacks[0].Version = "mutated-after-handle"
	original.Experiments[0].Enabled = false
	if got.TexturePacks[0].Version != "2" || !got.Experiments[0].Enabled {
		t.Fatalf("preserved ResourcePackStack aliases input slices: %#v", got)
	}
}

func TestHandlePreservesRawStartupPacketsForBridge(t *testing.T) {
	conn := newLoginTestConn()
	dimension := &packet.DimensionData{Definitions: []protocol.DimensionDefinition{{
		Name:          "integration:dimension",
		MinimumY:      -64,
		HeightRange:   384,
		Generator:     protocol.GeneratorOverworld,
		DimensionType: 1000,
		DefaultBiome:  "integration:biome",
	}}}
	conn.expectedIDs.Store([]uint32{packet.IDDimensionData})
	dimensionRaw := wirePacketData(t, dimension)
	if err := conn.handle(dimensionRaw); err != nil {
		t.Fatalf("handle DimensionData: %v", err)
	}

	itemRegistry := &packet.ItemRegistry{Items: []protocol.ItemEntry{{
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
	conn.expectedIDs.Store([]uint32{packet.IDItemRegistry})
	itemRaw := wirePacketData(t, itemRegistry)
	if err := conn.handle(itemRaw); err != nil {
		t.Fatalf("handle ItemRegistry: %v", err)
	}

	data := conn.GameData()
	if data.RawProtocolID != DefaultProtocol.ID() {
		t.Fatalf("raw protocol ID = %d, want %d", data.RawProtocolID, DefaultProtocol.ID())
	}
	if !bytes.Equal(data.RawDimensionData, dimensionRaw.full) {
		t.Fatal("retained DimensionData bytes differ from input")
	}
	if !bytes.Equal(data.RawItemRegistry, itemRaw.full) {
		t.Fatal("retained ItemRegistry bytes differ from input")
	}
}

func TestStartGameWritesRetainedItemRegistryBytes(t *testing.T) {
	itemRegistry := &packet.ItemRegistry{Items: []protocol.ItemEntry{{
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
	raw := wirePacketData(t, itemRegistry)
	var output bytes.Buffer
	conn := &Conn{
		ctx:   context.Background(),
		log:   slog.Default(),
		proto: DefaultProtocol,
		enc:   packet.NewEncoder(&output),
		hdr:   &packet.Header{},
		gameData: GameData{
			RawProtocolID:   DefaultProtocol.ID(),
			RawItemRegistry: append([]byte(nil), raw.full...),
		},
	}

	conn.startGame()
	decoded, err := packet.NewDecoder(bytes.NewReader(output.Bytes())).Decode()
	if err != nil {
		t.Fatalf("decode startup packets: %v", err)
	}
	if len(decoded) != 4 {
		t.Fatalf("startup packet count = %d, want 4", len(decoded))
	}
	if !bytes.Equal(decoded[len(decoded)-1], raw.full) {
		t.Fatal("StartGame did not write retained ItemRegistry bytes")
	}
}

func TestResourcePackResponseWritesRetainedStackBytes(t *testing.T) {
	stack := &packet.ResourcePackStack{
		TexturePackRequired: true,
		TexturePacks:        []protocol.StackResourcePack{{UUID: "integration-stack", Version: "1"}},
		BaseGameVersion:     "1.26.51-integration-base",
		Experiments:         []protocol.ExperimentData{{Name: "integration-experiment", Enabled: true}},
	}
	raw := wirePacketData(t, stack)
	rawBytes := append(append([]byte(nil), raw.full...), 0x7f)
	var output bytes.Buffer
	conn := &Conn{
		ctx:   context.Background(),
		log:   slog.Default(),
		proto: DefaultProtocol,
		enc:   packet.NewEncoder(&output),
		hdr:   &packet.Header{},
		gameData: GameData{
			RawProtocolID: DefaultProtocol.ID(),
		},
		resourcePackStack:              cloneResourcePackStack(stack),
		rawResourcePackStack:           rawBytes,
		rawResourcePackStackProtocolID: DefaultProtocol.ID(),
	}

	if err := conn.handleResourcePackClientResponse(&packet.ResourcePackClientResponse{Response: packet.PackResponseAllPacksDownloaded}); err != nil {
		t.Fatalf("handle all-packs-downloaded response: %v", err)
	}
	if err := conn.Flush(); err != nil {
		t.Fatal(err)
	}
	decoded, err := packet.NewDecoder(bytes.NewReader(output.Bytes())).Decode()
	if err != nil {
		t.Fatalf("decode resource-pack stack: %v", err)
	}
	if len(decoded) != 1 || !bytes.Equal(decoded[0], rawBytes) {
		t.Fatalf("resource-pack stack bytes = %x, want %x", decoded, rawBytes)
	}
}

func TestStartGameWritesRetainedStartupPacketBytes(t *testing.T) {
	jigsaw := &packet.JigsawStructureData{StructureData: map[string]any{
		"jigsaws": []any{map[string]any{"identifier": "integration:jigsaw"}},
	}}
	voxelShapes := &packet.VoxelShapes{
		NameMap:          []protocol.VoxelShapeNameEntry{{Name: "integration:shape", ID: 7}},
		CustomShapeCount: 7,
	}
	startGame := &packet.StartGame{
		WorldName:       "retained-start-game",
		EntityUniqueID:  1,
		EntityRuntimeID: 1,
		GameVersion:     "1.26.51",
	}
	jigsawRaw := wirePacketData(t, jigsaw)
	voxelRaw := wirePacketData(t, voxelShapes)
	startGameRaw := wirePacketData(t, startGame)
	var output bytes.Buffer
	conn := &Conn{
		ctx:   context.Background(),
		log:   slog.Default(),
		proto: DefaultProtocol,
		enc:   packet.NewEncoder(&output),
		hdr:   &packet.Header{},
		gameData: GameData{
			RawProtocolID:          DefaultProtocol.ID(),
			RawJigsawStructureData: append([]byte(nil), jigsawRaw.full...),
			RawStartGame:           append([]byte(nil), startGameRaw.full...),
			RawVoxelShapes:         append([]byte(nil), voxelRaw.full...),
		},
	}

	conn.startGame()
	decoded, err := packet.NewDecoder(bytes.NewReader(output.Bytes())).Decode()
	if err != nil {
		t.Fatalf("decode startup packets: %v", err)
	}
	if len(decoded) != 4 {
		t.Fatalf("startup packet count = %d, want 4", len(decoded))
	}
	for index, want := range [][]byte{jigsawRaw.full, voxelRaw.full, startGameRaw.full} {
		if !bytes.Equal(decoded[index], want) {
			t.Fatalf("startup packet %d differs from retained bytes", index)
		}
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
