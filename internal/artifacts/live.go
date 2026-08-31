// Package artifacts creates disposable, human-readable views of a live capture.
// It never modifies the canonical event stream or blobs.
package artifacts

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft"
)

const (
	Schema             = "bedrockdebugproxy.artifacts.v1"
	maxEventBytes      = 64 << 20
	maxSkinPacketBytes = 16 << 20
	maxImagePixels     = 1 << 20
)

type Status struct {
	Schema       string            `json:"schema"`
	CaptureID    string            `json:"capture_id"`
	Generator    capture.Generator `json:"generator"`
	ProtocolID   int32             `json:"protocol_id"`
	State        string            `json:"state"`
	LastSequence uint64            `json:"last_source_sequence"`
	Errors       uint64            `json:"errors"`
	Failure      string            `json:"failure,omitempty"`
	Limits       map[string]int    `json:"view_limits"`
}

type outputStream struct {
	file   *os.File
	buffer *bufio.Writer
}

type Live struct {
	source  *os.Root
	output  *os.Root
	events  *os.File
	streams map[string]outputStream
	names   map[uint32]string
	status  Status
	notify  func()
	finish  chan struct{}
	done    chan struct{}
	once    sync.Once
	err     error
}

// Start follows only a newly opened capture. Finish must be called after its
// recorder closes so the worker can drain a finite, fully published event stream.
func Start(path string, protocolID int32, notify func()) (*Live, error) {
	manifest, err := capture.ReadManifest(path)
	if err != nil {
		return nil, err
	}
	if manifest.Schema != capture.SchemaVersion || manifest.Status != "open" {
		return nil, errors.New("automatic artifacts require an open current-schema capture")
	}
	if protocolID != minecraft.DefaultProtocol.ID() {
		return nil, fmt.Errorf("unsupported artifact protocol %d", protocolID)
	}
	source, err := os.OpenRoot(path)
	if err != nil {
		return nil, err
	}
	if err := source.Mkdir("artifacts", 0o700); err != nil {
		_ = source.Close()
		return nil, err
	}
	output, err := source.OpenRoot("artifacts")
	if err != nil {
		_ = source.Close()
		return nil, err
	}
	events, err := source.Open("events.jsonl")
	if err != nil {
		_ = output.Close()
		_ = source.Close()
		return nil, err
	}
	w := &Live{
		source: source, output: output, events: events, notify: notify,
		streams: make(map[string]outputStream), names: make(map[uint32]string),
		finish: make(chan struct{}), done: make(chan struct{}),
		status: Status{Schema: Schema, CaptureID: manifest.CaptureID, Generator: manifest.Generator, ProtocolID: protocolID, State: "open",
			Limits: map[string]int{"event_bytes": maxEventBytes, "skin_packet_bytes": maxSkinPacketBytes, "image_pixels": maxImagePixels}},
	}
	for _, listener := range []bool{true, false} {
		for id, factory := range minecraft.DefaultProtocol.Packets(listener) {
			w.names[id] = reflect.TypeOf(factory()).Elem().Name()
		}
	}
	if err := w.writeStatus(); err != nil {
		_ = events.Close()
		_ = output.Close()
		_ = source.Close()
		return nil, err
	}
	go w.run()
	return w, nil
}

func (w *Live) Finish() (Status, error) {
	w.once.Do(func() { close(w.finish) })
	<-w.done
	return w.status, w.err
}

func (w *Live) run() {
	defer close(w.done)
	w.err = w.follow()
	w.err = errors.Join(w.err, w.events.Close())
	for _, stream := range w.streams {
		w.err = errors.Join(w.err, stream.buffer.Flush(), stream.file.Close())
	}
	w.status.State = "complete"
	if w.err != nil {
		w.status.State = "failed"
		w.status.Failure = w.err.Error()
		w.warn()
	} else if w.status.Errors != 0 {
		w.status.State = "complete_with_errors"
		w.err = fmt.Errorf("%d automatic artifact derivation(s) failed; see artifacts/errors.jsonl", w.status.Errors)
	}
	if err := w.writeStatus(); err != nil {
		w.err = errors.Join(w.err, err)
		w.status.State = "failed"
		w.status.Failure = w.err.Error()
		w.warn()
	}
	w.err = errors.Join(w.err, w.output.Close(), w.source.Close())
}

func (w *Live) follow() error {
	reader := bufio.NewReaderSize(w.events, 256<<10)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	var pending []byte
	lastFlush := time.Now()
	finished := false
	for {
		part, readErr := reader.ReadSlice('\n')
		if len(pending)+len(part) > maxEventBytes {
			return errors.New("artifact event line exceeds the 64 MiB view limit; original capture is unchanged")
		}
		if len(pending) != 0 || errors.Is(readErr, bufio.ErrBufferFull) || errors.Is(readErr, io.EOF) {
			pending = append(pending, part...)
			part = pending
		}
		if readErr == nil {
			if err := w.consume(part); err != nil {
				return err
			}
			pending = nil
		}
		if time.Since(lastFlush) >= time.Second || errors.Is(readErr, io.EOF) {
			for _, stream := range w.streams {
				if err := stream.buffer.Flush(); err != nil {
					return err
				}
			}
			lastFlush = time.Now()
		}
		switch {
		case readErr == nil, errors.Is(readErr, bufio.ErrBufferFull):
			continue
		case errors.Is(readErr, io.EOF):
			if finished {
				if len(pending) != 0 {
					return errors.New("capture ended with an incomplete event line")
				}
				data, err := w.source.ReadFile("manifest.json")
				if err != nil {
					return err
				}
				var manifest capture.Manifest
				if err := json.Unmarshal(data, &manifest); err != nil {
					return err
				}
				if manifest.Status == "open" || manifest.CaptureID != w.status.CaptureID || manifest.Counts.Events != w.status.LastSequence {
					return errors.New("artifact view does not match the closed capture event count or identity")
				}
				return nil
			}
			select {
			case <-w.finish:
				finished = true
			case <-ticker.C:
			}
		default:
			return readErr
		}
	}
}

func (w *Live) consume(line []byte) error {
	var event capture.Event
	if err := json.Unmarshal(line, &event); err != nil {
		return fmt.Errorf("read artifact source event: %w", err)
	}
	if event.Schema != capture.SchemaVersion || event.CaptureID != w.status.CaptureID || event.Sequence != w.status.LastSequence+1 {
		return errors.New("artifact source event schema, identity, or sequence mismatch")
	}
	w.status.LastSequence = event.Sequence
	name := ""
	if event.Packet != nil {
		name = event.Packet.Name
		if name == "" {
			name = w.names[event.Packet.ID]
		}
	}
	if event.Kind == "packet.raw" || event.Kind == "packet.decoded" || event.Kind == "packet.unknown" {
		if group := packetGroup(name); group != "" {
			if err := w.appendLine("observations/"+group+".jsonl", line); err != nil {
				return err
			}
		}
	}
	if event.Kind == "session.game_data" {
		if err := w.appendLine("observations/world.jsonl", line); err != nil {
			return err
		}
	}
	var err error
	switch event.Kind {
	case "resource_pack.archive", "resource_pack.decrypted_archive":
		err = w.archive(event)
	case "session.connection_metadata":
		err = w.loginSkin(event)
	case "packet.raw":
		if (name == "PlayerSkin" || name == "PlayerList") &&
			((event.Channel == "upstream" && event.Direction == capture.DirectionServerToClient) ||
				(event.Channel == "downstream" && event.Direction == capture.DirectionClientToServer)) {
			err = w.packetSkins(event)
		}
	}
	if err != nil {
		w.status.Errors++
		w.warn()
		return w.appendJSON("errors.jsonl", map[string]any{"source_sequence": event.Sequence, "kind": event.Kind, "error": err.Error()})
	}
	return nil
}

func (w *Live) warn() {
	if w.notify != nil {
		w.notify()
		w.notify = nil
	}
}

func (w *Live) appendJSON(path string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return w.appendLine(path, append(data, '\n'))
}

func (w *Live) appendLine(path string, line []byte) error {
	stream, ok := w.streams[path]
	if !ok {
		if slash := strings.LastIndexByte(path, '/'); slash >= 0 {
			if err := w.output.MkdirAll(path[:slash], 0o700); err != nil {
				return err
			}
		}
		file, err := w.output.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		stream = outputStream{file: file, buffer: bufio.NewWriterSize(file, 64<<10)}
		w.streams[path] = stream
	}
	_, err := stream.buffer.Write(line)
	return err
}

func (w *Live) writeStatus() error {
	data, err := json.MarshalIndent(w.status, "", "  ")
	if err != nil {
		return err
	}
	if err := w.output.WriteFile("status.json.tmp", append(data, '\n'), 0o600); err != nil {
		return err
	}
	return w.output.Rename("status.json.tmp", "status.json")
}

// Categories are navigation, not packet validation or a reconstruction of state.
func packetGroup(name string) string {
	switch {
	case name == "PlayerSkin" || name == "PlayerList":
		return "skins"
	case strings.Contains(name, "Inventory") || strings.Contains(name, "Container") || strings.HasPrefix(name, "ItemStack") || name == "MobEquipment" || name == "MobArmourEquipment" || name == "CreativeContent" || name == "CraftingData" || name == "ItemRegistry":
		return "inventory"
	case strings.HasPrefix(name, "UpdateBlock") || name == "UpdateSubChunkBlocks" || name == "BlockActorData" || name == "BlockEvent":
		return "blocks"
	case strings.Contains(name, "Chunk") || strings.HasPrefix(name, "ClientCache") || strings.Contains(name, "Biome") || name == "StartGame" || name == "ChangeDimension" || name == "SetTime" || name == "LevelEvent":
		return "world"
	case strings.Contains(name, "Actor") || strings.Contains(name, "Player") || name == "Animate" || name == "AnimateEntity" || name == "MobEffect" || name == "UpdateAttributes" || name == "Interact" || name == "Emote":
		return "entities"
	default:
		return ""
	}
}
