package bedrock

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft"
)

type Network struct {
	Transport        minecraft.Network
	Recorder         *capture.Recorder
	Observer         *Observer
	Failures         *FailureSink
	Logger           *slog.Logger
	SessionID        string
	ConnectionID     string
	ConnectionIDFunc func() string
	Channel          string
	Hop              int
	HopFunc          func() int
	ReadDirection    capture.Direction
	WriteDirection   capture.Direction
	// PayloadMetadataFunc returns bounded state associated with the connection
	// that produced the next transport payload event. It is evaluated for each
	// payload and supplements, but never replaces, the raw payload blob.
	PayloadMetadataFunc func() map[string]any
	// ConnectionContextFunc is called after a transport connection is opened.
	// It exposes the transport lifecycle context so callers can cancel work
	// that is waiting on the corresponding Minecraft connection.
	ConnectionContextFunc func(context.Context)
}

func (n Network) connectionID() string {
	if n.ConnectionIDFunc != nil {
		return n.ConnectionIDFunc()
	}
	return n.ConnectionID
}

func (n Network) hop() int {
	if n.HopFunc != nil {
		return n.HopFunc()
	}
	return n.Hop
}

func (n Network) DialContext(ctx context.Context, address string) (net.Conn, error) {
	conn, err := n.transport().DialContext(ctx, address)
	if err != nil {
		captureErr := n.recordConnection("transport.connection_error", nil, err)
		if captureErr != nil {
			return nil, errors.Join(err, captureErr)
		}
		return nil, err
	}
	return n.wrap(conn)
}

// DialContextIdentity preserves gophertunnel's authenticated transport hook when
// Network observes a transport such as NetherNet.
func (n Network) DialContextIdentity(ctx context.Context, address, token string, key *ecdsa.PrivateKey) (net.Conn, error) {
	transport := n.transport()
	identityTransport, ok := transport.(interface {
		DialContextIdentity(context.Context, string, string, *ecdsa.PrivateKey) (net.Conn, error)
	})
	if !ok {
		return n.DialContext(ctx, address)
	}
	conn, err := identityTransport.DialContextIdentity(ctx, address, token, key)
	if err != nil {
		captureErr := n.recordConnection("transport.connection_error", nil, err)
		if captureErr != nil {
			return nil, errors.Join(err, captureErr)
		}
		return nil, err
	}
	return n.wrap(conn)
}

func (n Network) PingContext(ctx context.Context, address string) ([]byte, error) {
	started := time.Now()
	response, operationErr := n.transport().PingContext(ctx, address)
	data, _ := json.Marshal(map[string]any{
		"address":        address,
		"response_bytes": len(response),
		"duration_nano":  time.Since(started).Nanoseconds(),
	})
	event := capture.Event{
		SessionID:    n.SessionID,
		ConnectionID: n.connectionID(),
		Hop:          n.hop(),
		Kind:         "transport.ping",
		Severity:     capture.SeverityInfo,
		Direction:    capture.DirectionInternal,
		Channel:      n.Channel,
		Stage:        "post_transport_ping",
		Data:         data,
	}
	if operationErr != nil {
		event.Severity = capture.SeverityError
		event.Error = captureError("ping", operationErr)
	}
	_, captureErr := n.Recorder.Record(context.Background(), capture.Record{
		Event:          event,
		Raw:            response,
		MediaType:      "application/octet-stream",
		Representation: "bedrock_transport_ping",
	})
	if captureErr != nil {
		n.Failures.Set(captureErr)
		captureErr = fmt.Errorf("capture transport ping: %w", captureErr)
	}
	return response, errors.Join(operationErr, captureErr)
}

func (n Network) Listen(address string) (minecraft.NetworkListener, error) {
	listener, err := n.transport().Listen(address)
	if err != nil {
		captureErr := n.recordConnection("transport.listen_error", nil, err)
		if captureErr != nil {
			return nil, errors.Join(err, captureErr)
		}
		return nil, err
	}
	return &observedListener{listener: listener, network: n}, nil
}

func (n Network) transport() minecraft.Network {
	if n.Transport != nil {
		return n.Transport
	}
	return rakNetNetwork{logger: n.Logger}
}

type rakNetNetwork struct {
	logger *slog.Logger
}

func (n rakNetNetwork) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return (raknet.Dialer{ErrorLog: n.logger}).DialContext(ctx, address)
}

func (n rakNetNetwork) PingContext(ctx context.Context, address string) ([]byte, error) {
	return (raknet.Dialer{ErrorLog: n.logger}).PingContext(ctx, address)
}

func (n rakNetNetwork) Listen(address string) (minecraft.NetworkListener, error) {
	return (raknet.ListenConfig{ErrorLog: n.logger}).Listen(address)
}

func (n Network) wrap(conn net.Conn) (net.Conn, error) {
	n.Observer.SetFlow(n.Channel, conn.LocalAddr(), conn.RemoteAddr())
	if err := n.recordConnection("transport.connection_open", conn, nil); err != nil {
		_ = conn.Close()
		return nil, err
	}
	observed := &observedConn{Conn: conn, network: n}
	if n.ConnectionContextFunc != nil {
		n.ConnectionContextFunc(observed.Context())
	}
	return observed, nil
}

func (n Network) recordConnection(kind string, conn net.Conn, operationErr error) error {
	var data []byte
	if metadata := transportConnectionMetadata(conn); len(metadata) > 0 {
		data, _ = json.Marshal(metadata)
	}
	event := capture.Event{
		SessionID:    n.SessionID,
		ConnectionID: n.connectionID(),
		Hop:          n.hop(),
		Kind:         kind,
		Severity:     capture.SeverityInfo,
		Direction:    capture.DirectionInternal,
		Channel:      n.Channel,
		Stage:        "post_transport_connection",
		Data:         data,
	}
	if conn != nil {
		event.Source = endpoint(conn.LocalAddr())
		event.Destination = endpoint(conn.RemoteAddr())
	}
	if operationErr != nil {
		event.Severity = capture.SeverityError
		event.Error = captureError("connection", operationErr)
	}
	_, err := n.Recorder.Record(context.Background(), capture.Record{Event: event})
	if err != nil {
		n.Failures.Set(err)
	}
	return err
}

func transportConnectionMetadata(conn net.Conn) map[string]any {
	if conn == nil {
		return nil
	}
	metadata := map[string]any{}
	if source, ok := conn.(interface{ ClientGUID() int64 }); ok {
		if clientGUID := source.ClientGUID(); clientGUID != 0 {
			metadata["raknet_client_guid"] = clientGUID
		}
	}
	if source, ok := conn.(interface{ PublicKey() *ecdsa.PublicKey }); ok {
		metadata["transport_public_key_capability"] = true
		metadata["transport_public_key_present"] = source.PublicKey() != nil
	}
	return metadata
}

type observedListener struct {
	listener minecraft.NetworkListener
	network  Network
}

func (l *observedListener) Accept() (net.Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return l.network.wrap(conn)
}

func (l *observedListener) Close() error {
	return l.listener.Close()
}

func (l *observedListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *observedListener) ID() int64 {
	return l.listener.ID()
}

func (l *observedListener) PongData(data []byte) {
	l.listener.PongData(data)
}

type observedConn struct {
	net.Conn
	network   Network
	closeOnce sync.Once
	closeErr  error
}

func (c *observedConn) Read(buffer []byte) (int, error) {
	n, err := c.Conn.Read(buffer)
	if captureErr := c.recordPayload("read", buffer[:n], n, n, err); captureErr != nil {
		return 0, captureErr
	}
	return n, err
}

func (c *observedConn) ReadPacket() ([]byte, error) {
	reader, ok := c.Conn.(interface {
		ReadPacket() ([]byte, error)
	})
	if !ok {
		return nil, errors.New("observed transport does not implement packet reads")
	}
	payload, err := reader.ReadPacket()
	if captureErr := c.recordPayload("read", payload, len(payload), len(payload), err); captureErr != nil {
		return nil, captureErr
	}
	return payload, err
}

func (c *observedConn) BatchHeader() []byte {
	if source, ok := c.Conn.(interface{ BatchHeader() []byte }); ok {
		return source.BatchHeader()
	}
	return []byte{0xfe}
}

func (c *observedConn) DisableEncryption() bool {
	if source, ok := c.Conn.(interface{ DisableEncryption() bool }); ok {
		return source.DisableEncryption()
	}
	return false
}

// PublicKey preserves the authenticated transport key exposed by NetherNet.
// gophertunnel uses this optional capability to compare the transport identity
// with the login identity before enabling packet encryption.
func (c *observedConn) PublicKey() *ecdsa.PublicKey {
	if source, ok := c.Conn.(interface{ PublicKey() *ecdsa.PublicKey }); ok {
		return source.PublicKey()
	}
	return nil
}

func (c *observedConn) Write(payload []byte) (int, error) {
	n, err := c.Conn.Write(payload)
	if captureErr := c.recordPayload("write", payload, len(payload), n, err); captureErr != nil {
		return n, captureErr
	}
	return n, err
}

func (c *observedConn) Close() error {
	c.closeOnce.Do(func() {
		transportErr := c.Conn.Close()
		captureErr := c.network.recordConnection("transport.connection_close", c.Conn, transportErr)
		c.closeErr = errors.Join(transportErr, captureErr)
	})
	return c.closeErr
}

func (c *observedConn) Latency() time.Duration {
	if source, ok := c.Conn.(interface{ Latency() time.Duration }); ok {
		return source.Latency()
	}
	return 0
}

// ClientGUID preserves the optional RakNet identity hook through the capture
// wrapper. Other transports simply do not implement this method.
func (c *observedConn) ClientGUID() int64 {
	if source, ok := c.Conn.(interface{ ClientGUID() int64 }); ok {
		return source.ClientGUID()
	}
	return 0
}

func (c *observedConn) Context() context.Context {
	if source, ok := c.Conn.(interface{ Context() context.Context }); ok {
		return source.Context()
	}
	return context.Background()
}

func (c *observedConn) recordPayload(operation string, payload []byte, attempted, transferred int, operationErr error) error {
	direction := c.network.ReadDirection
	source := c.RemoteAddr()
	destination := c.LocalAddr()
	if operation == "write" {
		direction = c.network.WriteDirection
		source = c.LocalAddr()
		destination = c.RemoteAddr()
	}
	dataMap := map[string]any{
		"operation":         operation,
		"attempted_bytes":   attempted,
		"transferred_bytes": transferred,
	}
	if c.network.PayloadMetadataFunc != nil {
		for key, value := range c.network.PayloadMetadataFunc() {
			dataMap[key] = value
		}
	}
	data, _ := json.Marshal(dataMap)
	event := capture.Event{
		SessionID:    c.network.SessionID,
		ConnectionID: c.network.connectionID(),
		Hop:          c.network.hop(),
		Kind:         "transport.payload",
		Severity:     capture.SeverityDebug,
		Direction:    direction,
		Channel:      c.network.Channel,
		Stage:        "post_transport_application_payload",
		Source:       endpoint(source),
		Destination:  endpoint(destination),
		Data:         data,
	}
	if operationErr != nil {
		event.Severity = capture.SeverityError
		event.Error = captureError(operation, operationErr)
	}
	var raw []byte
	if payload != nil {
		raw = payload
	}
	_, err := c.network.Recorder.Record(context.Background(), capture.Record{
		Event:          event,
		Raw:            raw,
		MediaType:      "application/octet-stream",
		Representation: "bedrock_transport_payload",
	})
	if err != nil {
		c.network.Failures.Set(err)
		return fmt.Errorf("capture transport %s: %w", operation, err)
	}
	return nil
}

func captureError(operation string, err error) *capture.ErrorInfo {
	if err == nil {
		return nil
	}
	info := &capture.ErrorInfo{Operation: operation, Message: err.Error(), Type: fmt.Sprintf("%T", err)}
	var netErr net.Error
	if errors.As(err, &netErr) {
		timeout := netErr.Timeout()
		info.Timeout = &timeout
	}
	return info
}
