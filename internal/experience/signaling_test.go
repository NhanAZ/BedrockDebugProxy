package experience

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/df-mc/go-nethernet"
)

func TestSignalingClosePublishesCauseBeforeWakingCredentials(t *testing.T) {
	for _, protocol := range []string{"legacy", "jsonrpc"} {
		t.Run(protocol, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := websocket.Accept(w, r, nil)
				if err != nil {
					return
				}
				defer func() { _ = conn.CloseNow() }()
				_, _, _ = conn.Read(r.Context())
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			// websocket.Dial owns and closes its handshake response body.
			conn, _, err := websocket.Dial(ctx, server.URL, nil) //nolint:bodyclose
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = conn.CloseNow() }()
			signalContext, cancelSignal := context.WithCancelCause(context.Background())
			defer cancelSignal(nil)
			closed := make(chan struct{})
			var mu *sync.Mutex
			var closeSignaling func() error
			var credentials func(context.Context) (*nethernet.Credentials, error)
			switch protocol {
			case "legacy":
				s := &legacySignaling{conn: conn, ctx: signalContext, cancel: cancelSignal, closed: closed}
				mu, closeSignaling, credentials = &s.notifierMu, s.Close, s.Credentials
			case "jsonrpc":
				s := &jsonRPCSignaling{conn: conn, ctx: signalContext, cancel: cancelSignal, closed: closed}
				mu, closeSignaling, credentials = &s.notifierMu, s.Close, s.Credentials
			}
			// Hold notifier cleanup so the close notification is observed
			// before Close can proceed to the WebSocket handshake.
			mu.Lock()
			closeResult := make(chan error, 1)
			go func() { closeResult <- closeSignaling() }()
			select {
			case <-closed:
			case <-ctx.Done():
				mu.Unlock()
				t.Fatal("signaling close did not notify credentials")
			}
			got, credentialsErr := credentials(ctx)
			mu.Unlock()
			<-closeResult
			if got != nil || !errors.Is(credentialsErr, net.ErrClosed) {
				t.Fatalf("Credentials() after close = %v, %v; want nil, net.ErrClosed", got, credentialsErr)
			}
		})
	}
}
