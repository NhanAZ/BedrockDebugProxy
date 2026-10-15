package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/golem"
)

// Exercise an agent choosing each next command after observing the last result.
// No complete scenario is supplied and stdin stays open between requests.
func TestGolemInteractiveServerControl(t *testing.T) {
	directory := os.Getenv("BDP_GOLEM_TEST_DATA")
	if directory == "" {
		t.Skip("set BDP_GOLEM_TEST_DATA to a disposable GolemAgent data directory")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	input, send := io.Pipe()
	receive, output := io.Pipe()
	defer func() { _ = send.Close(); _ = input.Close(); _ = receive.Close(); _ = output.Close() }()
	stop := context.AfterFunc(ctx, func() { _ = input.CloseWithError(ctx.Err()); _ = output.CloseWithError(ctx.Err()) })
	defer stop()
	var diagnostics bytes.Buffer
	done := make(chan int, 1)
	go func() {
		done <- runGolemInput(ctx, directory, input, output, &diagnostics, 10*time.Second)
	}()
	encoder, decoder := json.NewEncoder(send), json.NewDecoder(receive)
	type observation struct {
		Name     string `json:"name"`
		Online   bool   `json:"online"`
		Position struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"position"`
		Movement struct {
			Status string `json:"status"`
		} `json:"movement"`
	}
	call := func(request golem.Request) observation {
		t.Helper()
		if err := encoder.Encode(request); err != nil {
			t.Fatal(err)
		}
		var response golem.Response
		if err := decoder.Decode(&response); err != nil || !response.OK {
			t.Fatalf("interactive response: %+v, %v", response, err)
		}
		var result observation
		if request.Op == "state" || request.Op == "move" || request.Op == "stop" {
			if err := json.Unmarshal(response.Result, &result); err != nil {
				t.Fatal(err)
			}
		}
		return result
	}
	call(golem.Request{Op: "reset"})
	call(golem.Request{Op: "spawn", Args: map[string]any{"name": "LiveAgent", "seed": 7}})
	state := call(golem.Request{Op: "state", Player: "LiveAgent"})
	position := state.Position
	move := call(golem.Request{Op: "move", Player: "LiveAgent", Args: map[string]any{"x": position.X + 8, "y": position.Y, "z": position.Z, "speed": 0.1}})
	if move.Movement.Status != "pending" {
		t.Fatalf("move did not return immediately: %+v", move)
	}
	state = call(golem.Request{Op: "state", Player: "LiveAgent"})
	if !state.Online {
		t.Fatal("moving golem is not online")
	}
	stopped := call(golem.Request{Op: "stop", Player: "LiveAgent"})
	if stopped.Movement.Status != "cancelled" {
		t.Fatalf("stop did not cancel the observed movement: %+v", stopped)
	}
	call(golem.Request{Op: "chat", Player: "LiveAgent", Args: map[string]any{"message": "Observed " + state.Name}})
	_ = send.Close()
	if code := <-done; code != 0 {
		t.Fatalf("interactive CLI exit=%d, %s", code, diagnostics.String())
	}
	// Closing the CLI does not remove the server-owned player.
	var out, diagnostic bytes.Buffer
	code := runGolemInput(ctx, directory, strings.NewReader("{\"op\":\"state\",\"player\":\"LiveAgent\"}\n{\"op\":\"reset\"}\n"), &out, &diagnostic, 10*time.Second)
	if code != 0 || !strings.Contains(out.String(), "Observed LiveAgent") {
		t.Fatalf("reattach failed: code=%d output=%s diagnostic=%s", code, out.String(), diagnostic.String())
	}
	t.Log("interactive spawn, observe, move, stop, chat and reattach passed")
}
