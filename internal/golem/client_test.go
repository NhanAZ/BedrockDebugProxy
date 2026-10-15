package golem

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func mailbox(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"requests", "responses"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "ready.json"), []byte(`{"schema":"`+Schema+`","boot":"test-boot"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCallRoundTripAndIdentity(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "wrong-boot"}[corrupt], func(t *testing.T) {
			dir := mailbox(t)
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = root.Close() }()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				ticker := time.NewTicker(time.Millisecond)
				defer ticker.Stop()
				for {
					files, err := filepath.Glob(filepath.Join(dir, "requests", "*.json"))
					if err != nil {
						done <- err
						return
					}
					if len(files) != 0 {
						data, err := readBounded(root, filepath.Join("requests", filepath.Base(files[0])))
						if err != nil {
							done <- err
							return
						}
						var req envelope
						if err := json.Unmarshal(data, &req); err != nil {
							done <- err
							return
						}
						if req.Request.Op != "spawn" || req.Deadline <= time.Now().UnixMilli() {
							done <- errors.New("invalid request envelope")
							return
						}
						boot := req.Boot
						if corrupt {
							boot = "another-boot"
						}
						data, err = json.Marshal(Response{Schema: Schema, Boot: boot, TransportID: req.TransportID, ID: req.Request.ID, OK: true, Result: json.RawMessage(`{"online":true}`)})
						if err == nil {
							err = atomicWrite(root, filepath.Join("responses", req.TransportID+".json"), data)
						}
						done <- err
						return
					}
					select {
					case <-ctx.Done():
						done <- ctx.Err()
						return
					case <-ticker.C:
					}
				}
			}()
			response, err := Call(ctx, dir, Request{ID: "step-1", Op: "spawn", Args: map[string]any{"name": "AgentOne"}})
			if corrupt {
				if err == nil || !strings.Contains(err.Error(), "identity mismatch") {
					t.Fatalf("unexpected error %v", err)
				}
			} else if err != nil || !response.OK || response.ID != "step-1" {
				t.Fatalf("response=%+v error=%v", response, err)
			}
			if err := <-done; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCallCancellationRetainsRequestAndMarksCancel(t *testing.T) {
	dir := mailbox(t)
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer cancel()
	_, err := Call(ctx, dir, Request{Op: "wait"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error=%v", err)
	}
	files, err := filepath.Glob(filepath.Join(dir, "requests", "*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("requests=%v error=%v", files, err)
	}
	if _, err := os.Stat(strings.TrimSuffix(files[0], ".json") + ".cancel"); err != nil {
		t.Fatal(err)
	}
}

func TestCallRejectsMissingDeadlineAndOversizeBeforePublishing(t *testing.T) {
	dir := mailbox(t)
	if _, err := Call(context.Background(), dir, Request{Op: "spawn"}); err == nil {
		t.Fatal("missing deadline accepted")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := Call(ctx, dir, Request{Op: "chat", Args: map[string]any{"message": strings.Repeat("x", MaxMessageBytes)}}); err == nil {
		t.Fatal("oversized request accepted")
	}
	files, err := os.ReadDir(filepath.Join(dir, "requests"))
	if err != nil || len(files) != 0 {
		t.Fatalf("requests=%v error=%v", files, err)
	}
}
