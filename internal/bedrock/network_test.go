package bedrock

import (
	"context"
	"errors"
	"io"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

func TestObservedConnPreservesPacketReadsContextAndLatency(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	source := &fakePacketConn{ctx: ctx, readPayload: []byte{1, 2, 3}, latency: 25 * time.Millisecond}
	failures := &FailureSink{}
	conn := &observedConn{Conn: source, network: Network{
		Recorder:       recorder,
		Failures:       failures,
		SessionID:      "session-test",
		ConnectionID:   "connection-test",
		Channel:        "upstream",
		Hop:            1,
		ReadDirection:  capture.DirectionServerToClient,
		WriteDirection: capture.DirectionClientToServer,
	}}

	read, err := conn.ReadPacket()
	if err != nil || string(read) != string([]byte{1, 2, 3}) {
		t.Fatalf("ReadPacket() = %x, %v", read, err)
	}
	if written, err := conn.Write([]byte{4, 5}); err != nil || written != 2 {
		t.Fatalf("Write() = %d, %v", written, err)
	}
	if conn.Context() != ctx {
		t.Fatal("Context() did not preserve the transport context")
	}
	if conn.Latency() != 25*time.Millisecond {
		t.Fatalf("Latency() = %v", conn.Latency())
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := failures.Err(); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	var events []capture.Event
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Fatalf("event count = %d, want 3", len(events))
	}
	if events[0].Kind != "transport.payload" || events[0].Direction != capture.DirectionServerToClient || events[0].Blob == nil || events[0].Blob.Size != 3 {
		t.Fatalf("read event = %#v", events[0])
	}
	if events[1].Kind != "transport.payload" || events[1].Direction != capture.DirectionClientToServer || events[1].Blob == nil || events[1].Blob.Size != 2 {
		t.Fatalf("write event = %#v", events[1])
	}
	if events[2].Kind != "transport.connection_close" {
		t.Fatalf("close event = %#v", events[2])
	}
}

func TestNetworkWrapReportsTransportContext(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	transportContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	source := &fakePacketConn{ctx: transportContext}
	failures := &FailureSink{}
	observer := NewObserver(recorder, failures, "session-test", 1, nil)
	reported := make(chan context.Context, 1)
	network := Network{
		Recorder:              recorder,
		Observer:              observer,
		Failures:              failures,
		SessionID:             "session-test",
		ConnectionID:          "connection-test",
		Channel:               "downstream",
		ConnectionContextFunc: func(ctx context.Context) { reported <- ctx },
	}
	wrapped, err := network.wrap(source)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-reported:
		if got != transportContext {
			t.Fatal("ConnectionContextFunc received a different context")
		}
	case <-time.After(time.Second):
		t.Fatal("ConnectionContextFunc was not called")
	}
	if err := wrapped.Close(); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
}

func TestObservedConnPreservesTransportCapabilities(t *testing.T) {
	conn := &observedConn{Conn: &fakeCapabilityConn{fakePacketConn: fakePacketConn{}}}
	if got := conn.BatchHeader(); got != nil {
		t.Fatalf("BatchHeader() = %x, want nil", got)
	}
	if !conn.DisableEncryption() {
		t.Fatal("DisableEncryption() = false")
	}
}

func TestObservedConnReportsTransferredBytesWhenCaptureFails(t *testing.T) {
	recorder, err := capture.New(filepath.Join(t.TempDir(), "capture"), capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	failures := &FailureSink{}
	conn := &observedConn{Conn: &fakePacketConn{}, network: Network{
		Recorder:       recorder,
		Failures:       failures,
		WriteDirection: capture.DirectionClientToServer,
	}}
	written, err := conn.Write([]byte{1, 2, 3})
	if written != 3 || err == nil {
		t.Fatalf("Write() = %d, %v", written, err)
	}
	if failures.Err() == nil {
		t.Fatal("capture failure was not retained")
	}
}

func TestFailureSinkNotifiesOnce(t *testing.T) {
	notifications := 0
	sink := NewFailureSink(func() { notifications++ })
	first := errors.New("first")
	sink.Set(first)
	sink.Set(errors.New("second"))
	if !errors.Is(sink.Err(), first) {
		t.Fatalf("Err() = %v", sink.Err())
	}
	if notifications != 1 {
		t.Fatalf("notifications = %d, want 1", notifications)
	}
}

type fakePacketConn struct {
	ctx         context.Context
	readPayload []byte
	latency     time.Duration
	closed      bool
}

type fakeCapabilityConn struct {
	fakePacketConn
}

func (*fakeCapabilityConn) BatchHeader() []byte     { return nil }
func (*fakeCapabilityConn) DisableEncryption() bool { return true }

func (c *fakePacketConn) Read(buffer []byte) (int, error) {
	if c.closed {
		return 0, net.ErrClosed
	}
	if len(c.readPayload) == 0 {
		return 0, io.EOF
	}
	return copy(buffer, c.readPayload), nil
}

func (c *fakePacketConn) ReadPacket() ([]byte, error) {
	if c.closed {
		return nil, net.ErrClosed
	}
	if c.readPayload == nil {
		return nil, io.EOF
	}
	return append([]byte(nil), c.readPayload...), nil
}

func (c *fakePacketConn) Write(payload []byte) (int, error) {
	if c.closed {
		return 0, net.ErrClosed
	}
	return len(payload), nil
}

func (c *fakePacketConn) Close() error {
	if c.closed {
		return errors.New("already closed")
	}
	c.closed = true
	return nil
}

func (c *fakePacketConn) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 19132}
}

func (c *fakePacketConn) RemoteAddr() net.Addr {
	return &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 19132}
}

func (c *fakePacketConn) SetDeadline(time.Time) error      { return nil }
func (c *fakePacketConn) SetReadDeadline(time.Time) error  { return nil }
func (c *fakePacketConn) SetWriteDeadline(time.Time) error { return nil }
func (c *fakePacketConn) Latency() time.Duration           { return c.latency }

func (c *fakePacketConn) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}
