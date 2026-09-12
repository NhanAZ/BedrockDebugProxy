package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

func TestExpectedShutdownErrorRequiresCoordinatedShutdown(t *testing.T) {
	wrappedCanceled := fmt.Errorf("write packet: %w", context.Canceled)
	wrappedClosed := fmt.Errorf("read packet: %w", net.ErrClosed)
	var shuttingDown atomic.Bool

	if expectedShutdownError(wrappedCanceled, &shuttingDown) {
		t.Fatal("context cancellation was suppressed before shutdown")
	}
	shuttingDown.Store(true)
	for _, err := range []error{wrappedCanceled, wrappedClosed} {
		if !expectedShutdownError(err, &shuttingDown) {
			t.Fatalf("expected coordinated shutdown error %v to be suppressed", err)
		}
	}
	if expectedShutdownError(errors.New("protocol violation"), &shuttingDown) {
		t.Fatal("unrelated error was suppressed during shutdown")
	}
}

func TestFinishForwardDoesNotRecordExpectedShutdownWriteError(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	runner := &Runner{config: Config{Recorder: recorder}}
	var shuttingDown atomic.Bool
	shuttingDown.Store(true)
	operationErr := fmt.Errorf("write packet: %w", context.Canceled)
	if err := runner.finishForward("bridge.write_error", capture.DirectionServerToClient, "upstream", "upstream-1", "write packet", operationErr, &shuttingDown); !errors.Is(err, context.Canceled) {
		t.Fatalf("finishForward() error = %v, want context cancellation", err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	var bridgeErrors int
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind == "bridge.write_error" {
			bridgeErrors++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if bridgeErrors != 0 {
		t.Fatalf("recorded %d expected shutdown bridge errors", bridgeErrors)
	}
}
