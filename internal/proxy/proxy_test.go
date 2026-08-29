package proxy

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

func TestRunnerStopsCleanlyWhenCancelledBeforeAccept(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := New(Config{
		ListenAddress:   "127.0.0.1:0",
		UpstreamAddress: "127.0.0.1:19133",
		Recorder:        recorder,
		Output:          io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := runner.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	verification, err := capture.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) != 0 {
		t.Fatalf("capture issues = %v", verification.Issues)
	}
	if verification.Events < 2 {
		t.Fatalf("event count = %d, want at least 2 lifecycle events", verification.Events)
	}
}
