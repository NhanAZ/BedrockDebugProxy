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

	if err := conn.handle(wirePacketData(t, original)); err != nil {
		t.Fatalf("handle StartGame: %v", err)
	}
	got := conn.GameData().StartGamePacket
	if got == nil {
		t.Fatal("GameData.StartGamePacket is nil")
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

	if err := conn.handle(wirePacketData(t, jigsaw)); err != nil {
		t.Fatalf("handle JigsawStructureData: %v", err)
	}
	if err := conn.handle(wirePacketData(t, voxelShapes)); err != nil {
		t.Fatalf("handle VoxelShapes: %v", err)
	}
	data := conn.GameData()
	if data.JigsawStructureData == nil || !reflect.DeepEqual(data.JigsawStructureData.StructureData, jigsaw.StructureData) {
		t.Fatalf("preserved JigsawStructureData = %#v, want %#v", data.JigsawStructureData, jigsaw)
	}
	if data.VoxelShapes == nil || !reflect.DeepEqual(data.VoxelShapes.NameMap, voxelShapes.NameMap) || data.VoxelShapes.CustomShapeCount != voxelShapes.CustomShapeCount {
		t.Fatalf("preserved VoxelShapes = %#v, want %#v", data.VoxelShapes, voxelShapes)
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

	if err := conn.handle(wirePacketData(t, original)); err != nil {
		t.Fatalf("handle ResourcePackStack: %v", err)
	}
	got := conn.GameData().ResourcePackStack
	if got == nil || !reflect.DeepEqual(got, original) {
		t.Fatalf("preserved ResourcePackStack = %#v, want %#v", got, original)
	}

	original.TexturePacks[0].Version = "mutated-after-handle"
	original.Experiments[0].Enabled = false
	if got.TexturePacks[0].Version != "2" || !got.Experiments[0].Enabled {
		t.Fatalf("preserved ResourcePackStack aliases input slices: %#v", got)
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
