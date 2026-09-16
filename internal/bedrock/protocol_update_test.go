package bedrock

import (
	"bytes"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestProtocolVersion12650(t *testing.T) {
	if got, want := protocol.CurrentProtocol, 2193; got != want {
		t.Fatalf("protocol ID = %d, want %d", got, want)
	}
	if got, want := protocol.CurrentVersion, "1.26.50"; got != want {
		t.Fatalf("game version = %q, want %q", got, want)
	}
}

func TestProtocol12650PacketRegistration(t *testing.T) {
	if got, want := packet.IDSetPlayerFurnaceOptions, 351; got != want {
		t.Fatalf("SetPlayerFurnaceOptions packet ID = %d, want %d", got, want)
	}
	if got, want := packet.IDRecordStarted, 352; got != want {
		t.Fatalf("RecordStarted packet ID = %d, want %d", got, want)
	}

	clientFactory, ok := packet.NewClientPool()[packet.IDSetPlayerFurnaceOptions]
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
