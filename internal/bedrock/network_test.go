package bedrock

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft"
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

func TestObservedConnAddsPayloadMetadataWithoutReplacingRawPayload(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	conn := &observedConn{Conn: &fakePacketConn{readPayload: []byte{1, 2, 3}}, network: Network{
		Recorder:       recorder,
		Failures:       &FailureSink{},
		SessionID:      "session-test",
		Channel:        "upstream",
		ReadDirection:  capture.DirectionServerToClient,
		WriteDirection: capture.DirectionClientToServer,
		PayloadMetadataFunc: func() map[string]any {
			return map[string]any{
				"bedrock_encryption_state":      "enabled",
				"bedrock_compression_state":     "negotiated",
				"bedrock_compression_threshold": uint16(17),
			}
		},
	}}
	if _, err := conn.ReadPacket(); err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind != "transport.payload" {
			return nil
		}
		var data map[string]any
		if err := json.Unmarshal(event.Data, &data); err != nil {
			return err
		}
		if data["bedrock_encryption_state"] != "enabled" || data["bedrock_compression_state"] != "negotiated" || data["bedrock_compression_threshold"] != float64(17) {
			t.Fatalf("payload metadata = %#v", data)
		}
		if event.Blob == nil || event.Blob.Size != 3 || event.Blob.Representation != "bedrock_transport_payload" {
			t.Fatalf("transport payload blob = %#v", event.Blob)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
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
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	conn := &observedConn{Conn: &fakeCapabilityConn{fakePacketConn: fakePacketConn{}, publicKey: &key.PublicKey}}
	if got := conn.BatchHeader(); got != nil {
		t.Fatalf("BatchHeader() = %x, want nil", got)
	}
	if !conn.DisableEncryption() {
		t.Fatal("DisableEncryption() = false")
	}
	if got := conn.PublicKey(); got != &key.PublicKey {
		t.Fatal("PublicKey() did not preserve the transport key")
	}
}

func TestRecordConnectionCapturesTransportIdentityMetadata(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	network := Network{
		Recorder:     recorder,
		Failures:     &FailureSink{},
		SessionID:    "session-test",
		ConnectionID: "connection-test",
		Channel:      "upstream",
		Hop:          2,
	}
	conn := &fakeCapabilityConn{fakePacketConn: fakePacketConn{}, publicKey: &key.PublicKey, clientGUID: 1234}
	if err := network.recordConnection("transport.connection_open", conn, nil); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind != "transport.connection_open" {
			return nil
		}
		var data map[string]any
		if err := json.Unmarshal(event.Data, &data); err != nil {
			return err
		}
		if data["raknet_client_guid"] != float64(1234) || data["transport_public_key_capability"] != true || data["transport_public_key_present"] != true {
			t.Fatalf("connection metadata = %#v", data)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestNetworkPingCapturesResponseBytes(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	response := []byte("MCPE;local;1;2;3;4;5;6;7;19132;")
	network := Network{
		Transport:    &fakePingNetwork{response: response},
		Recorder:     recorder,
		Failures:     &FailureSink{},
		SessionID:    "session-test",
		ConnectionID: "connection-test",
		Channel:      "upstream",
		Hop:          2,
	}
	got, err := network.PingContext(context.Background(), "example.test:19132")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(response) {
		t.Fatalf("ping response = %q, want %q", got, response)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	var event capture.Event
	if err := capture.ScanEvents(root, func(candidate capture.Event) error {
		if candidate.Kind == "transport.ping" {
			event = candidate
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if event.Kind != "transport.ping" || event.Direction != capture.DirectionInternal || event.Channel != "upstream" || event.Hop != 2 {
		t.Fatalf("ping event = %#v", event)
	}
	if event.Blob == nil || event.Blob.Size != int64(len(response)) || event.Blob.Representation != "bedrock_transport_ping" {
		t.Fatalf("ping blob = %#v", event.Blob)
	}
	var data map[string]any
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data["address"] != "example.test:19132" || data["response_bytes"] != float64(len(response)) {
		t.Fatalf("ping data = %#v", data)
	}
}

func TestNetworkPingCapturesFailure(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	pingErr := errors.New("ping timeout")
	network := Network{
		Transport: &fakePingNetwork{response: []byte("partial"), err: pingErr},
		Recorder:  recorder,
		Failures:  &FailureSink{},
	}
	response, err := network.PingContext(context.Background(), "example.test:19132")
	if !errors.Is(err, pingErr) {
		t.Fatalf("ping error = %v", err)
	}
	if string(response) != "partial" {
		t.Fatalf("partial response = %q", response)
	}
	if err := recorder.Close("closed", pingErr); err != nil {
		t.Fatal(err)
	}

	var event capture.Event
	if err := capture.ScanEvents(root, func(candidate capture.Event) error {
		if candidate.Kind == "transport.ping" {
			event = candidate
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if event.Severity != capture.SeverityError || event.Error == nil || event.Error.Operation != "ping" {
		t.Fatalf("ping failure event = %#v", event)
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

func TestObservedConnPreservesTransportErrorWhenCaptureFails(t *testing.T) {
	recorder, err := capture.New(filepath.Join(t.TempDir(), "capture"), capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	readErr := errors.New("transport read failed")
	writeErr := errors.New("transport write failed")
	conn := &observedConn{Conn: &fakePacketConn{
		readPayload: []byte{1, 2, 3},
		readErr:     readErr,
		writeErr:    writeErr,
	}, network: Network{
		Recorder:       recorder,
		Failures:       &FailureSink{},
		ReadDirection:  capture.DirectionServerToClient,
		WriteDirection: capture.DirectionClientToServer,
	}}
	buffer := make([]byte, 8)
	read, err := conn.Read(buffer)
	if read != 3 || !errors.Is(err, readErr) || !errors.Is(err, capture.ErrClosed) {
		t.Fatalf("Read() = %d, %v, want transferred bytes and both errors", read, err)
	}
	written, err := conn.Write([]byte{4, 5})
	if written != 2 || !errors.Is(err, writeErr) || !errors.Is(err, capture.ErrClosed) {
		t.Fatalf("Write() = %d, %v, want transferred bytes and both errors", written, err)
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
	readErr     error
	writeErr    error
	latency     time.Duration
	closed      bool
}

type fakeCapabilityConn struct {
	fakePacketConn
	publicKey  *ecdsa.PublicKey
	clientGUID int64
}

func (*fakeCapabilityConn) BatchHeader() []byte           { return nil }
func (*fakeCapabilityConn) DisableEncryption() bool       { return true }
func (c *fakeCapabilityConn) PublicKey() *ecdsa.PublicKey { return c.publicKey }
func (c *fakeCapabilityConn) ClientGUID() int64           { return c.clientGUID }

type fakePingNetwork struct {
	response []byte
	err      error
}

func (*fakePingNetwork) DialContext(context.Context, string) (net.Conn, error) {
	return nil, errors.New("fake ping network does not dial")
}

func (n *fakePingNetwork) PingContext(context.Context, string) ([]byte, error) {
	return append([]byte(nil), n.response...), n.err
}

func (*fakePingNetwork) Listen(string) (minecraft.NetworkListener, error) {
	return nil, errors.New("fake ping network does not listen")
}

func (c *fakePacketConn) Read(buffer []byte) (int, error) {
	if c.closed {
		return 0, net.ErrClosed
	}
	if len(c.readPayload) == 0 {
		return 0, io.EOF
	}
	return copy(buffer, c.readPayload), c.readErr
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
	return len(payload), c.writeErr
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
