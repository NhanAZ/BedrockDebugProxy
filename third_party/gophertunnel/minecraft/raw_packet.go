package minecraft

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PacketRead contains one application packet before protocol decoding.
// Data includes the packet header and payload exactly as received after
// transport decrypt/decompress processing. Payload starts after the header.
// The raw data remains valid until Release or DecodePacketRead is called.
// Packet is set for packets that were already decoded by the connection's
// login state machine and therefore have no raw representation available.
type PacketRead struct {
	Header  packet.Header
	Data    []byte
	Payload []byte
	Packet  packet.Packet

	data *packetData
}

func packetReadFromData(data *packetData) *PacketRead {
	return &PacketRead{
		Header:  *data.h,
		Data:    data.full,
		Payload: data.payload.Bytes(),
		data:    data,
	}
}

// Release releases the raw ownership held by the read. The connection does
// not pool packetData today, but clearing the references prevents accidental
// use after forwarding and keeps the ownership contract explicit.
func (r *PacketRead) Release() {
	if r == nil {
		return
	}
	r.data = nil
	r.Data = nil
	r.Payload = nil
}

// Clone returns an independently owned raw packet that can be decoded while
// retaining the original for byte-preserving forwarding.
func (r *PacketRead) Clone() (*PacketRead, error) {
	if r == nil {
		return nil, errors.New("clone packet read: nil packet")
	}
	if r.Packet != nil {
		return &PacketRead{Packet: r.Packet}, nil
	}
	if len(r.Data) == 0 {
		return nil, errors.New("clone packet read: empty packet")
	}
	data, err := parseData(append([]byte(nil), r.Data...), nil)
	if err != nil {
		return nil, fmt.Errorf("clone packet read: %w", err)
	}
	return packetReadFromData(data), nil
}

// ReadPacketDataWithTime reads one packet without discarding its wire bytes.
// Packets already decoded by the login state machine are returned in Packet
// and must be forwarded through WritePacket instead.
func (conn *Conn) ReadPacketDataWithTime() (*PacketRead, time.Time, error) {
	if len(conn.additional) > 0 {
		return &PacketRead{Packet: <-conn.additional}, time.Now(), nil
	}
	if data, ok := conn.takeDeferredPacket(); ok {
		return packetReadFromData(data), time.Now(), nil
	}

	select {
	case <-conn.ctx.Done():
		return nil, time.Now(), conn.closeErr("read packet")
	case <-conn.readDeadline:
		return nil, time.Now(), conn.wrap(context.DeadlineExceeded, "read packet")
	case data := <-conn.packets:
		return packetReadFromData(data), time.Now(), nil
	}
}

// DecodePacketRead decodes a raw read and releases its raw ownership. Any
// additional packets produced by protocol conversion are queued for the next
// read, matching ReadPacket behavior.
func (conn *Conn) DecodePacketRead(read *PacketRead) (packet.Packet, error) {
	if read == nil {
		return nil, errors.New("decode packet read: nil packet")
	}
	if read.Packet != nil {
		return read.Packet, nil
	}
	if read.data == nil {
		return nil, errors.New("decode packet read: packet already released")
	}
	data := read.data
	read.data = nil
	read.Data = nil
	read.Payload = nil
	packets, err := data.decode(conn)
	if err != nil {
		return nil, err
	}
	if len(packets) == 0 {
		return nil, nil
	}
	for _, additional := range packets[1:] {
		conn.additional <- additional
	}
	return packets[0], nil
}

// WriteRawPacket forwards packet bytes already expressed in this connection's
// protocol. The bytes are copied before returning, so the PacketRead may be
// released immediately afterwards.
func (conn *Conn) WriteRawPacket(read *PacketRead) error {
	if read == nil || len(read.Data) == 0 {
		return errors.New("write raw packet: empty packet")
	}
	select {
	case <-conn.ctx.Done():
		return conn.closeErr("write raw packet")
	default:
	}
	conn.sendMu.Lock()
	conn.bufferedSend = append(conn.bufferedSend, append([]byte(nil), read.Data...))
	if conn.packetFunc != nil {
		conn.packetFunc(read.Header, read.Payload, conn.LocalAddr(), conn.RemoteAddr())
	}
	conn.sendMu.Unlock()
	return nil
}

// ProtocolID returns the negotiated Bedrock protocol identifier.
func (conn *Conn) ProtocolID() int32 {
	return conn.proto.ID()
}

// Raw forwarding only supports the current protocol. The helper is kept on
// the connection so callers can compare both endpoints before using it.
var _ interface {
	ReadPacketDataWithTime() (*PacketRead, time.Time, error)
	DecodePacketRead(*PacketRead) (packet.Packet, error)
	WriteRawPacket(*PacketRead) error
	ProtocolID() int32
} = (*Conn)(nil)
