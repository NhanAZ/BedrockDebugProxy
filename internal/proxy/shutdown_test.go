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

func TestFinishForwardCoordinatesRacingShutdownErrors(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	runner := &Runner{config: Config{Recorder: recorder}}
	var shuttingDown atomic.Bool
	readErr := fmt.Errorf("read packet: %w", context.Canceled)
	if err := runner.finishForward("bridge.read_error", capture.DirectionClientToServer, "downstream", "downstream-1", "read packet", readErr, &shuttingDown); !errors.Is(err, context.Canceled) {
		t.Fatalf("finishForward() read error = %v, want context cancellation", err)
	}
	if !shuttingDown.Load() {
		t.Fatal("finishForward() did not claim coordinated shutdown")
	}
	writeErr := fmt.Errorf("write packet: %w", context.Canceled)
	if err := runner.finishForward("bridge.write_error", capture.DirectionServerToClient, "upstream", "upstream-1", "write packet", writeErr, &shuttingDown); !errors.Is(err, context.Canceled) {
		t.Fatalf("finishForward() write error = %v, want context cancellation", err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	var bridgeErrors, readErrors int
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind == "bridge.write_error" {
			bridgeErrors++
		}
		if event.Kind == "bridge.read_error" {
			readErrors++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if bridgeErrors != 0 {
		t.Fatalf("recorded %d racing peer write errors", bridgeErrors)
	}
	if readErrors != 1 {
		t.Fatalf("recorded %d read-side shutdown errors, want 1", readErrors)
	}
}

func TestFinishForwardSuppressesFirstPeerCloseWriteError(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	runner := &Runner{config: Config{Recorder: recorder}}
	var shuttingDown atomic.Bool
	operationErr := fmt.Errorf("write packet: %w", net.ErrClosed)
	if err := runner.finishForward("bridge.write_error", capture.DirectionServerToClient, "upstream", "upstream-1", "write packet", operationErr, &shuttingDown); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("finishForward() error = %v, want closed-connection error", err)
	}
	if !shuttingDown.Load() {
		t.Fatal("finishForward() did not claim coordinated shutdown")
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
		t.Fatalf("recorded %d first peer-close write errors", bridgeErrors)
	}
}
