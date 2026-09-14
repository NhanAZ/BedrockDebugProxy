package minecraft

import (
	"bytes"
	"context"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func testWirePacket(t *testing.T) []byte {
	t.Helper()
	var full bytes.Buffer
	header := packet.Header{PacketID: packet.IDPlayStatus}
	if err := header.Write(&full); err != nil {
		t.Fatal(err)
	}
	(&packet.PlayStatus{Status: packet.PlayStatusPlayerSpawn}).Marshal(DefaultProtocol.NewWriter(&full, 0))
	return full.Bytes()
}

func TestWriteRawPacketPreservesWirePacket(t *testing.T) {
	var output bytes.Buffer
	conn := &Conn{
		ctx:   context.Background(),
		proto: DefaultProtocol,
		enc:   packet.NewEncoder(&output),
	}
	data, err := parseData(testWirePacket(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.WriteRawPacket(packetReadFromData(data)); err != nil {
		t.Fatal(err)
	}
	if err := conn.Flush(); err != nil {
		t.Fatal(err)
	}

	decoder := packet.NewDecoder(bytes.NewReader(output.Bytes()))
	packets, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 || !bytes.Equal(packets[0], testWirePacket(t)) {
		t.Fatal("raw forwarded packet differs from its wire input")
	}
}

func TestReadPacketDataCloneCanBeDecodedAfterOriginalRelease(t *testing.T) {
	data, err := parseData(testWirePacket(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := &Conn{
		ctx:        context.Background(),
		proto:      DefaultProtocol,
		pool:       packet.NewServerPool(),
		packets:    make(chan *packetData, 1),
		additional: make(chan packet.Packet, 8),
	}
	conn.packets <- data

	read, _, err := conn.ReadPacketDataWithTime()
	if err != nil {
		t.Fatal(err)
	}
	clone, err := read.Clone()
	if err != nil {
		t.Fatal(err)
	}
	read.Release()
	decoded, err := conn.DecodePacketRead(clone)
	if err != nil {
		t.Fatal(err)
	}
	status, ok := decoded.(*packet.PlayStatus)
	if !ok || status.Status != packet.PlayStatusPlayerSpawn {
		t.Fatalf("decoded packet = %#v", decoded)
	}
}
