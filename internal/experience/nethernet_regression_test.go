package experience

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/df-mc/go-nethernet"
)

func TestNetherNetOfferDoesNotDuplicateSDPCandidatePrefix(t *testing.T) {
	client, server := newInProcessNetherNetSignalingPair("client", "server")
	t.Cleanup(client.close)
	t.Cleanup(server.close)

	listener, err := (nethernet.ListenConfig{
		AllowAnonymous:    true,
		DisableTrickleICE: true,
	}).Listen(server)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	accepted := make(chan net.Conn, 1)
	acceptErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			acceptErr <- err
			return
		}
		accepted <- conn
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := (nethernet.Dialer{DisableTrickleICE: true}).DialContext(ctx, server.NetworkID(), client)
	if err != nil {
		t.Fatalf("DialContext() error = %v", err)
	}
	defer func() { _ = conn.Close() }()

	select {
	case offer := <-client.offers:
		if strings.Contains(offer, "a=candidate:candidate:") {
			t.Fatal("NetherNet SDP offer duplicated the candidate prefix")
		}
		if !strings.Contains(offer, "a=candidate:") {
			t.Fatal("NetherNet SDP offer did not contain an ICE candidate")
		}
	case <-ctx.Done():
		t.Fatal("NetherNet did not publish an SDP offer")
	}

	select {
	case serverConn := <-accepted:
		_ = serverConn.Close()
	case err := <-acceptErr:
		t.Fatalf("Accept() error = %v", err)
	case <-ctx.Done():
		t.Fatal("NetherNet listener did not accept the connection")
	}
}

type inProcessNetherNetSignaling struct {
	id     string
	peer   *inProcessNetherNetSignaling
	offers chan string
	ctx    context.Context
	cancel context.CancelFunc

	mu         sync.Mutex
	notifiers  map[int]nethernet.Notifier
	nextNotify int
}

func newInProcessNetherNetSignalingPair(clientID, serverID string) (*inProcessNetherNetSignaling, *inProcessNetherNetSignaling) {
	clientCtx, clientCancel := context.WithCancel(context.Background())
	serverCtx, serverCancel := context.WithCancel(context.Background())
	client := &inProcessNetherNetSignaling{
		id:        clientID,
		offers:    make(chan string, 1),
		ctx:       clientCtx,
		cancel:    clientCancel,
		notifiers: make(map[int]nethernet.Notifier),
	}
	server := &inProcessNetherNetSignaling{
		id:        serverID,
		ctx:       serverCtx,
		cancel:    serverCancel,
		notifiers: make(map[int]nethernet.Notifier),
	}
	client.peer = server
	server.peer = client
	return client, server
}

func (s *inProcessNetherNetSignaling) Signal(ctx context.Context, signal *nethernet.Signal) error {
	if signal.Type == nethernet.SignalTypeOffer {
		s.offers <- signal.Data
	}

	s.peer.mu.Lock()
	notifiers := make([]nethernet.Notifier, 0, len(s.peer.notifiers))
	for _, notifier := range s.peer.notifiers {
		notifiers = append(notifiers, notifier)
	}
	s.peer.mu.Unlock()

	delivered := *signal
	delivered.NetworkID = s.id
	for _, notifier := range notifiers {
		if !notifier.NotifySignal(&delivered) {
			return net.ErrClosed
		}
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.peer.ctx.Done():
		return net.ErrClosed
	default:
		return nil
	}
}

func (s *inProcessNetherNetSignaling) Notify(notifier nethernet.Notifier) func() {
	s.mu.Lock()
	id := s.nextNotify
	s.nextNotify++
	s.notifiers[id] = notifier
	s.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			s.mu.Lock()
			delete(s.notifiers, id)
			s.mu.Unlock()
		})
	}
}

func (s *inProcessNetherNetSignaling) Context() context.Context { return s.ctx }

func (*inProcessNetherNetSignaling) Credentials(context.Context) (*nethernet.Credentials, error) {
	return nil, nil
}

func (s *inProcessNetherNetSignaling) NetworkID() string { return s.id }

func (*inProcessNetherNetSignaling) PongData([]byte) {}

func (*inProcessNetherNetSignaling) DisableTrickleICE() bool { return true }

func (s *inProcessNetherNetSignaling) close() { s.cancel() }
