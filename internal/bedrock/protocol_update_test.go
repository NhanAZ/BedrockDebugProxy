package bedrock

import (
	"bytes"
	"testing"

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
