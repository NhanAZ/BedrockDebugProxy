package bedrock

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestProtocolVersion12661(t *testing.T) {
	if got, want := protocol.CurrentProtocol, 2225; got != want {
		t.Fatalf("protocol ID = %d, want %d", got, want)
	}
	if got, want := protocol.CurrentVersion, "1.26.61"; got != want {
		t.Fatalf("game version = %q, want %q", got, want)
	}
}

func TestClientCacheBlobStatusAllowsObservedHashCount(t *testing.T) {
	want := &packet.ClientCacheBlobStatus{
		MissHashes: make([]uint64, 1302),
		HitHashes:  []uint64{0x0102030405060708},
	}
	for i := range want.MissHashes {
		want.MissHashes[i] = uint64(i + 1)
	}

	var encoded bytes.Buffer
	want.Marshal(protocol.NewWriter(&encoded, 0))

	got := &packet.ClientCacheBlobStatus{}
	got.Marshal(protocol.NewReader(bytes.NewReader(encoded.Bytes()), 0, true))
	if len(got.MissHashes) != len(want.MissHashes) {
		t.Fatalf("decoded miss hash count = %d, want %d", len(got.MissHashes), len(want.MissHashes))
	}
	if got.MissHashes[0] != want.MissHashes[0] || got.MissHashes[len(got.MissHashes)-1] != want.MissHashes[len(want.MissHashes)-1] {
		t.Fatalf("decoded miss hash boundaries = %#x, %#x, want %#x, %#x", got.MissHashes[0], got.MissHashes[len(got.MissHashes)-1], want.MissHashes[0], want.MissHashes[len(want.MissHashes)-1])
	}
	if len(got.HitHashes) != 1 || got.HitHashes[0] != want.HitHashes[0] {
		t.Fatalf("decoded hit hashes = %#v, want %#v", got.HitHashes, want.HitHashes)
	}
}

func TestProtocol2225PacketRegistration(t *testing.T) {
	if got, want := packet.IDSetPlayerFurnaceOptions, 351; got != want {
		t.Fatalf("SetPlayerFurnaceOptions packet ID = %d, want %d", got, want)
	}
	if got, want := packet.IDRecordStarted, 352; got != want {
		t.Fatalf("RecordStarted packet ID = %d, want %d", got, want)
	}

	clientPool := packet.NewClientPool()
	clientFactory, ok := clientPool[packet.IDSetPlayerFurnaceOptions]
	if !ok {
		t.Fatal("SetPlayerFurnaceOptions is not registered in the client packet pool")
	}
	if _, ok := clientFactory().(*packet.SetPlayerFurnaceOptions); !ok {
		t.Fatalf("client packet factory returned %T", clientFactory())
	}

	serverPool := packet.NewServerPool()
	for id, expected := range map[uint32]any{
		packet.IDSetPlayerFurnaceOptions: (*packet.SetPlayerFurnaceOptions)(nil),
		packet.IDRecordStarted:           (*packet.RecordStarted)(nil),
	} {
		factory, ok := serverPool[id]
		if !ok {
			t.Fatalf("packet ID %d is not registered in the server packet pool", id)
		}
		switch expected.(type) {
		case *packet.SetPlayerFurnaceOptions:
			if _, ok := factory().(*packet.SetPlayerFurnaceOptions); !ok {
				t.Fatalf("server packet factory for ID %d returned %T", id, factory())
			}
		case *packet.RecordStarted:
			if _, ok := factory().(*packet.RecordStarted); !ok {
				t.Fatalf("server packet factory for ID %d returned %T", id, factory())
			}
		}
	}

	clientTypes := map[uint32]packet.Packet{
		packet.IDServerboundStonecutterSetRecipe: &packet.ServerboundStonecutterSetRecipe{},
		packet.IDServerboundMatchmakingCancel:    &packet.ServerboundMatchmakingCancel{},
		packet.IDServerboundCursorItemDrag:       &packet.ServerboundCursorItemDrag{},
		packet.IDServerboundRegisterAudioContent: &packet.ServerboundRegisterAudioContent{},
	}
	for id, want := range clientTypes {
		factory, ok := clientPool[id]
		if !ok {
			t.Fatalf("client packet pool is missing packet ID %d", id)
		}
		if got := factory(); reflect.TypeOf(got) != reflect.TypeOf(want) {
			t.Fatalf("client packet factory for ID %d returned %T, want %T", id, got, want)
		}
	}

	serverTypes := map[uint32]packet.Packet{
		packet.IDClientboundMatchmakingState:     &packet.ClientboundMatchmakingState{},
		packet.IDClientboundStonecutterSetRecipe: &packet.ClientboundStonecutterSetRecipe{},
		packet.IDSetPassengerOfBlock:             &packet.SetPassengerOfBlock{},
		packet.IDClientboundPlayAudioContent:     &packet.ClientboundPlayAudioContent{},
	}
	for id, want := range serverTypes {
		factory, ok := serverPool[id]
		if !ok {
			t.Fatalf("server packet pool is missing packet ID %d", id)
		}
		if got := factory(); reflect.TypeOf(got) != reflect.TypeOf(want) {
			t.Fatalf("server packet factory for ID %d returned %T, want %T", id, got, want)
		}
	}
}

func TestProtocol2225PacketEncodings(t *testing.T) {
	encode := func(pk packet.Packet) []byte {
		var encoded bytes.Buffer
		pk.Marshal(protocol.NewWriter(&encoded, 0))
		return encoded.Bytes()
	}

	tests := []struct {
		name string
		pk   packet.Packet
		want []byte
	}{
		{
			name: "serverbound cursor item drag",
			pk:   &packet.ServerboundCursorItemDrag{State: packet.CursorItemDragStop},
			want: []byte{0x01},
		},
		{
			name: "serverbound stonecutter recipe",
			pk: &packet.ServerboundStonecutterSetRecipe{
				ContainerID: 4,
				RecipeIndex: -3,
			},
			want: []byte{0x04, 0x05},
		},
		{
			name: "clientbound matchmaking state",
			pk: &packet.ClientboundMatchmakingState{
				State:           packet.MatchmakingStateTimedOut,
				DestinationName: "hub",
			},
			want: []byte{0x07, 0x03, 'h', 'u', 'b', 0x00},
		},
		{
			name: "clientbound stonecutter recipe",
			pk: &packet.ClientboundStonecutterSetRecipe{
				PlayerUniqueID: -2,
				ContainerID:    4,
				RecipeIndex:    -3,
			},
			want: []byte{0x03, 0x04, 0x05},
		},
		{
			name: "set passenger of block",
			pk:   &packet.SetPassengerOfBlock{PassengerUniqueID: 123},
			want: []byte{0xf6, 0x01, 0x00},
		},
		{
			name: "serverbound register audio content",
			pk: &packet.ServerboundRegisterAudioContent{
				Registrations: []protocol.AudioContentRegistrationEntry{{
					AudioContentID:  "id",
					SharedMetadata:  "a",
					ServerContent:   "b",
					PlaybackContent: "c",
				}},
			},
			want: []byte{0x01, 0x02, 'i', 'd', 0x01, 'a', 0x01, 'b', 0x01, 'c'},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := encode(test.pk); !bytes.Equal(got, test.want) {
				t.Fatalf("encoded bytes = %x, want %x", got, test.want)
			}
		})
	}
}

func TestProtocol2225CorrectedPacketLayouts(t *testing.T) {
	t.Run("camera spline instruction", func(t *testing.T) {
		want := protocol.CameraSplineInstruction{
			TotalTime:         1.5,
			SplineType:        2,
			Curve:             []mgl32.Vec3{},
			ProgressKeyFrames: []protocol.CameraProgressOption{},
			RotationOptions:   []protocol.CameraRotationOption{},
			SplineIdentifier:  "spline",
			LoadFromJson:      true,
		}
		var encoded bytes.Buffer
		want.Marshal(protocol.NewWriter(&encoded, 0))

		var got protocol.CameraSplineInstruction
		got.Marshal(protocol.NewReader(bytes.NewReader(encoded.Bytes()), 0, true))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("decoded instruction = %#v, want %#v", got, want)
		}
	})

	t.Run("player video capture start", func(t *testing.T) {
		pk := &packet.PlayerVideoCapture{
			Action:     packet.PlayerVideoCaptureActionStart,
			FrameRate:  30,
			FilePrefix: "frame",
		}
		var encoded bytes.Buffer
		pk.Marshal(protocol.NewWriter(&encoded, 0))
		want := []byte{0x00, 0x1e, 0x00, 0x00, 0x00, 0x05, 'f', 'r', 'a', 'm', 'e'}
		if got := encoded.Bytes(); !bytes.Equal(got, want) {
			t.Fatalf("encoded video capture = %x, want %x", got, want)
		}
	})

	t.Run("structure block filtered name", func(t *testing.T) {
		want := packet.StructureBlockUpdate{FilteredStructureName: protocol.Option("filtered")}
		var encoded bytes.Buffer
		want.Marshal(protocol.NewWriter(&encoded, 0))

		var got packet.StructureBlockUpdate
		got.Marshal(protocol.NewReader(bytes.NewReader(encoded.Bytes()), 0, true))
		filtered, ok := got.FilteredStructureName.Value()
		if !ok || filtered != "filtered" {
			t.Fatalf("filtered structure name = %q, set = %t", filtered, ok)
		}
	})
}

func TestNetherNetClientDataAcceptsSinglePortURL(t *testing.T) {
	base := login.ClientData{
		DeviceOS:          1,
		GameVersion:       "1.26.61",
		LanguageCode:      "en_US",
		SkinResourcePatch: "e30=",
		SkinID:            "synthetic",
	}
	for _, address := range []string{
		"https://example.com:443",
		"https://example.com:443:443",
	} {
		t.Run(address, func(t *testing.T) {
			data := base
			data.ServerAddress = address
			if err := data.Validate(); err != nil {
				t.Fatalf("ClientData.Validate(%q) error = %v", address, err)
			}
		})
	}
}

func TestSetPlayerFurnaceOptionsEncoding(t *testing.T) {
	pk := &packet.SetPlayerFurnaceOptions{
		FurnaceType: packet.FurnaceTypeSmoker,
		FurnaceOptions: protocol.FurnaceOptions{
			LeftFurnaceTab: protocol.FurnaceLeftTabRecipeSearch,
			Filtering:      true,
			Layout:         protocol.FurnaceLayoutDefault,
		},
	}
	var encoded bytes.Buffer
	pk.Marshal(protocol.NewWriter(&encoded, 0))

	want := []byte{0x03, 0x08, 0x01, 0x04}
	if got := encoded.Bytes(); !bytes.Equal(got, want) {
		t.Fatalf("encoded furnace options = %x, want %x", got, want)
	}
}

func TestRecordStartedEncoding(t *testing.T) {
	pk := &packet.RecordStarted{
		Position: protocol.BlockPos{1, 2, -3},
		Handle:   0x0102030405060708,
	}
	var encoded bytes.Buffer
	pk.Marshal(protocol.NewWriter(&encoded, 0))

	want := []byte{0x02, 0x04, 0x05, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	if got := encoded.Bytes(); !bytes.Equal(got, want) {
		t.Fatalf("encoded record start = %x, want %x", got, want)
	}
}

func TestPlayerInventoryActionRoundTripIncludesHand(t *testing.T) {
	want := protocol.UseItemTransactionData{
		LegacyRequestID:     0,
		Actions:             []protocol.InventoryAction{},
		ActionType:          protocol.UseItemActionBreakBlock,
		TriggerType:         protocol.TriggerTypeSimulationTick,
		BlockPosition:       protocol.BlockPos{12, 34, -56},
		BlockFace:           5,
		HotBarSlot:          7,
		Hand:                protocol.HandSlotOffHand,
		Position:            mgl32.Vec3{1.25, 2.5, 3.75},
		ClickedPosition:     mgl32.Vec3{0.25, 0.5, 0.75},
		BlockRuntimeID:      42,
		ClientPrediction:    protocol.ClientPredictionSuccess,
		ClientCooldownState: protocol.ClientCooldownStateOn,
	}

	var encoded bytes.Buffer
	w := protocol.NewWriter(&encoded, 0)
	w.PlayerInventoryAction(&want)

	got := protocol.UseItemTransactionData{}
	raw := bytes.NewReader(encoded.Bytes())
	r := protocol.NewReader(raw, 0, true)
	r.PlayerInventoryAction(&got)
	if raw.Len() != 0 {
		t.Fatalf("PlayerInventoryAction left %d bytes", raw.Len())
	}
	if got.Hand != want.Hand {
		t.Fatalf("hand = %d, want %d", got.Hand, want.Hand)
	}
	if got.Position != want.Position || got.ClickedPosition != want.ClickedPosition {
		t.Fatalf("positions shifted after hand: got position %v/%v, want %v/%v", got.Position, got.ClickedPosition, want.Position, want.ClickedPosition)
	}
}

func TestSubChunkHeightMapRoundTripUsesRows(t *testing.T) {
	var want protocol.HeightMap
	for z := range want {
		for x := range want[z] {
			want[z][x] = int8(z*16 + x - 128)
		}
	}
	wantEntry := protocol.SubChunkEntry{
		Offset:        protocol.SubChunkOffset{1, -2, 3},
		Result:        protocol.SubChunkResultSuccess,
		HeightMapType: protocol.HeightMapDataHasData,
		HeightMapData: protocol.Option(want),
	}

	var encoded bytes.Buffer
	wantEntry.Marshal(protocol.NewWriter(&encoded, 0))
	data := encoded.Bytes()
	if got, want := len(data), 282; got != want {
		t.Fatalf("encoded sub-chunk entry length = %d, want %d", got, want)
	}
	// The height map begins after the offset, result, raw-payload option,
	// height-map selector, and height-map option. Each row starts with its
	// varuint32 length before the 16 signed heights.
	if got, want := data[7], byte(16); got != want {
		t.Fatalf("first height-map row length = %d, want %d", got, want)
	}
	if got, want := data[8], byte(0x80); got != want {
		t.Fatalf("first height-map value byte = %#x, want %#x", got, want)
	}
	if got, want := data[24], byte(16); got != want {
		t.Fatalf("second height-map row length = %d, want %d", got, want)
	}
	if got, want := data[25], byte(0x90); got != want {
		t.Fatalf("second height-map value byte = %#x, want %#x", got, want)
	}

	var gotEntry protocol.SubChunkEntry
	raw := bytes.NewReader(data)
	gotEntry.Marshal(protocol.NewReader(raw, 0, true))
	if raw.Len() != 0 {
		t.Fatalf("sub-chunk entry left %d bytes", raw.Len())
	}
	got, ok := gotEntry.HeightMapData.Value()
	if !ok {
		t.Fatal("decoded height map is not set")
	}
	if got != want {
		t.Fatalf("decoded height map differs from input")
	}
}

func TestSubChunkHeightMapRejectsWrongRowLength(t *testing.T) {
	// Keep enough bytes for the former flat-slice decoder to succeed. The
	// row-aware decoder must reject the first row before consuming its values.
	data := make([]byte, 282)
	data[5] = protocol.HeightMapDataHasData
	data[6] = 1
	data[7] = 15
	data[279] = protocol.HeightMapDataNone

	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("decoding accepted a height-map row whose length was not 16")
		}
	}()
	var entry protocol.SubChunkEntry
	entry.Marshal(protocol.NewReader(bytes.NewReader(data), 0, true))
}
