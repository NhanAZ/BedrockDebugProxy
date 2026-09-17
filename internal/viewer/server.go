package viewer

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

// LocalServer serves one capture viewer on an automatically selected IPv4
// loopback port. The unpredictable URL path is an access boundary for other
// local browser tabs, not a substitute for operating-system access control.
type LocalServer struct {
	URL    string
	server *http.Server
	done   chan struct{}
	mu     sync.Mutex
	err    error
}

func StartLocal(view *Viewer) (*LocalServer, error) {
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	token, err := NewAccessToken()
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	local := &LocalServer{
		URL:  "http://" + listener.Addr().String() + "/" + token + "/",
		done: make(chan struct{}),
	}
	local.server = &http.Server{
		Handler:           view.Handler(token),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	go func() {
		err := local.server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		local.mu.Lock()
		local.err = err
		local.mu.Unlock()
		close(local.done)
	}()
	return local, nil
}

func (s *LocalServer) Done() <-chan struct{} {
	return s.done
}

func (s *LocalServer) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

func (s *LocalServer) Close(ctx context.Context) error {
	shutdownErr := s.server.Shutdown(ctx)
	if shutdownErr != nil {
		_ = s.server.Close()
	}
	select {
	case <-s.done:
		return errors.Join(shutdownErr, s.Err())
	case <-ctx.Done():
		_ = s.server.Close()
		return errors.Join(shutdownErr, ctx.Err())
	}
}
