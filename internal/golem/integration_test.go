package golem

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Set BDP_GOLEM_TEST_DATA only for an operator-selected disposable server.
// The fixture resets its simulated players and changes its flat test world.
func TestDevelopmentServerIntegration(t *testing.T) {
	directory := os.Getenv("BDP_GOLEM_TEST_DATA")
	if directory == "" {
		t.Skip("set BDP_GOLEM_TEST_DATA to a disposable GolemAgent data directory for server integration")
	}
	source, err := os.ReadFile(filepath.Join("..", "..", "dev", "golem-agent", "tests", "smoke.php"))
	if err != nil {
		t.Fatal(err)
	}
	name := "integration-" + uuid.NewString() + ".php"
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	if err := root.WriteFile(filepath.Join("scripts", name), source, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	response, err := Call(ctx, directory, Request{Op: "script", Args: map[string]any{"file": name}})
	cancel()
	if err != nil || !response.OK {
		t.Fatalf("server fixture: %v, %+v", err, response)
	}
	var report struct {
		Passed int `json:"passed"`
	}
	if err := json.Unmarshal(response.Result, &report); err != nil || report.Passed < 74 {
		t.Fatalf("incomplete fixture report: %s, %v", response.Result, err)
	}
	t.Logf("server simulation passed %d checks", report.Passed)
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	_, err = Call(ctx, directory, Request{Op: "wait", Args: map[string]any{"ticks": 20}})
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected request deadline, got %v", err)
	}
	started := time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	response, err = Call(ctx, directory, Request{Op: "wait", Args: map[string]any{"ticks": 40}})
	cancel()
	if err != nil || !response.OK || time.Since(started) < 1800*time.Millisecond {
		t.Fatalf("expired callback completed another request: %v, %+v, %s", err, response, time.Since(started))
	}
}
