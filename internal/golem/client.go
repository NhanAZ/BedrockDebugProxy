// Package golem controls the optional server-side development runtime.
// Its records are simulation evidence, not Bedrock network captures.
package golem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const Schema = "bedrockdebugproxy.golem.v1"
const MaxMessageBytes = 1 << 20

type Ready struct {
	Schema  string `json:"schema"`
	Boot    string `json:"boot"`
	Runtime string `json:"runtime"`
}

type Request struct {
	ID     string         `json:"id,omitempty"`
	Op     string         `json:"op"`
	Player string         `json:"player,omitempty"`
	Args   map[string]any `json:"args,omitempty"`
}

type envelope struct {
	Schema      string  `json:"schema"`
	Boot        string  `json:"boot"`
	TransportID string  `json:"transport_id"`
	Deadline    int64   `json:"deadline_unix_ms"`
	Request     Request `json:"request"`
}

type Response struct {
	Schema      string          `json:"schema"`
	Boot        string          `json:"boot"`
	TransportID string          `json:"transport_id"`
	ID          string          `json:"id"`
	OK          bool            `json:"ok"`
	Tick        int64           `json:"tick"`
	Result      json.RawMessage `json:"result,omitempty"`
	Error       string          `json:"error,omitempty"`
}

// Call publishes one request without automatic retries. Interrupted actions
// may already have had effects on the development server.
func Call(ctx context.Context, directory string, request Request) (Response, error) {
	var result Response
	if request.Op == "" {
		return result, errors.New("request op is required")
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return result, fmt.Errorf("open GolemAgent data directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	data, err := readBounded(root, "ready.json")
	if err != nil {
		return result, fmt.Errorf("read GolemAgent readiness (install and start the companion plugin): %w", err)
	}
	var ready Ready
	if err := json.Unmarshal(data, &ready); err != nil {
		return result, fmt.Errorf("decode readiness: %w", err)
	}
	if ready.Schema != Schema || ready.Boot == "" {
		return result, errors.New("unsupported GolemAgent readiness schema or missing boot identity")
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return result, errors.New("golem request requires a context deadline")
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	id := uuid.NewString()
	if request.ID == "" {
		request.ID = id
	}
	data, err = json.Marshal(envelope{Schema: Schema, Boot: ready.Boot, TransportID: id, Deadline: deadline.UnixMilli(), Request: request})
	if err != nil {
		return result, err
	}
	if len(data) > MaxMessageBytes {
		return result, errors.New("request exceeds the explicit 1 MiB limit")
	}
	requestPath := filepath.Join("requests", id+".json")
	if err := atomicWrite(root, requestPath, data); err != nil {
		return result, fmt.Errorf("publish request: %w", err)
	}
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		data, err := readBounded(root, filepath.Join("responses", id+".json"))
		if err == nil {
			if err := json.Unmarshal(data, &result); err != nil {
				return result, fmt.Errorf("decode response: %w", err)
			}
			if result.Schema != Schema || result.Boot != ready.Boot || result.TransportID != id || result.ID != request.ID {
				return result, errors.New("golem response identity mismatch")
			}
			return result, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return result, fmt.Errorf("read response: %w", err)
		}
		select {
		case <-ctx.Done():
			cancelErr := atomicWrite(root, filepath.Join("requests", id+".cancel"), []byte("cancelled\n"))
			return result, errors.Join(fmt.Errorf("golem request interrupted, effects may already exist, inspect the retained request and journal: %w", ctx.Err()), cancelErr)
		case <-ticker.C:
		}
	}
}

func readBounded(root *os.Root, path string) ([]byte, error) {
	f, err := root.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxMessageBytes {
		return nil, errors.New("message exceeds the explicit 1 MiB limit")
	}
	data := make([]byte, info.Size())
	_, err = io.ReadFull(f, data)
	return data, err
}

func atomicWrite(root *os.Root, path string, data []byte) error {
	f, err := root.OpenFile(path+".tmp", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	err = errors.Join(writeErr, f.Close())
	if err != nil {
		return err
	}
	return root.Rename(path+".tmp", path)
}
