package capture

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	manifestName = "manifest.json"
	eventsName   = "events.jsonl"
	blobRoot     = "blobs/sha256"
)

var ErrClosed = errors.New("capture recorder is closed")

type Options struct {
	CaptureID     string
	GeneratorName string
	Version       string
	Commit        string
	SyncEachEvent bool
	RawLayers     []string
	Values        map[string]string
	Limitations   []string
	Now           func() time.Time
}

type Recorder struct {
	mu       sync.Mutex
	root     string
	events   *os.File
	start    time.Time
	now      func() time.Time
	sequence uint64
	seenBlob map[string]struct{}
	failed   error
	closed   bool
	manifest Manifest
}

func New(root string, options Options) (*Recorder, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("capture root is empty")
	}
	if err := requireEmptyOrMissingDirectory(root); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(blobRoot)), 0o700); err != nil {
		return nil, fmt.Errorf("create capture directories: %w", err)
	}

	eventsPath := filepath.Join(root, eventsName)
	events, err := os.OpenFile(eventsPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create event stream: %w", err)
	}

	now := options.Now
	if now == nil {
		now = time.Now
	}
	start := now()
	captureID := options.CaptureID
	if captureID == "" {
		captureID, err = randomID()
		if err != nil {
			_ = events.Close()
			return nil, fmt.Errorf("create capture ID: %w", err)
		}
	}
	executable, _ := os.Executable()
	if options.GeneratorName == "" {
		options.GeneratorName = "bedrock-debug-proxy"
	}

	r := &Recorder{
		root:     root,
		events:   events,
		start:    start,
		now:      now,
		seenBlob: make(map[string]struct{}),
		manifest: Manifest{
			Schema:    SchemaVersion,
			CaptureID: captureID,
			StartedAt: start.UTC().Format(time.RFC3339Nano),
			Status:    "open",
			Generator: Generator{
				Name:       options.GeneratorName,
				Version:    options.Version,
				Commit:     options.Commit,
				GoVersion:  runtime.Version(),
				GOOS:       runtime.GOOS,
				GOARCH:     runtime.GOARCH,
				Executable: executable,
			},
			Options: CaptureOptions{
				SyncEachEvent: options.SyncEachEvent,
				RawLayers:     append([]string(nil), options.RawLayers...),
				Values:        cloneMap(options.Values),
			},
			EventsPath: eventsName,
			BlobRoot:   blobRoot,
			Completeness: Completeness{
				Complete:    len(options.Limitations) == 0,
				Limitations: append([]string(nil), options.Limitations...),
			},
		},
	}
	if err := r.writeManifestLocked(); err != nil {
		_ = events.Close()
		return nil, err
	}
	return r, nil
}

func (r *Recorder) Root() string {
	return r.root
}

func (r *Recorder) CaptureID() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.manifest.CaptureID
}

func (r *Recorder) Manifest() Manifest {
	r.mu.Lock()
	defer r.mu.Unlock()
	return cloneManifest(r.manifest)
}

func (r *Recorder) Record(ctx context.Context, record Record) (Event, error) {
	var raw io.Reader
	expectedSize := int64(-1)
	if record.Raw != nil {
		raw = bytes.NewReader(record.Raw)
		expectedSize = int64(len(record.Raw))
	}
	return r.record(ctx, record.Event, raw, expectedSize, record.MediaType, record.Representation)
}

// RecordReader records an event while streaming its raw blob from raw. expectedSize may be -1 when the size is unknown.
func (r *Recorder) RecordReader(ctx context.Context, event Event, raw io.Reader, expectedSize int64, mediaType, representation string) (Event, error) {
	if raw == nil {
		return Event{}, errors.New("raw reader is nil")
	}
	if expectedSize < -1 {
		return Event{}, errors.New("expected raw size must be -1 or greater")
	}
	return r.record(ctx, event, raw, expectedSize, mediaType, representation)
}

func (r *Recorder) record(ctx context.Context, event Event, raw io.Reader, expectedSize int64, mediaType, representation string) (Event, error) {
	if event.Blob != nil {
		return Event{}, errors.New("event blob references are recorder-owned")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return Event{}, ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return Event{}, ErrClosed
	}
	if r.failed != nil {
		return Event{}, fmt.Errorf("capture recorder failed: %w", r.failed)
	}
	select {
	case <-ctx.Done():
		return Event{}, ctx.Err()
	default:
	}

	now := r.now()
	sequence := r.sequence + 1
	event.Schema = SchemaVersion
	event.Sequence = sequence
	event.Time = now.UTC().Format(time.RFC3339Nano)
	event.UnixNano = now.UnixNano()
	event.ElapsedNano = now.Sub(r.start).Nanoseconds()
	event.CaptureID = r.manifest.CaptureID
	if event.Kind == "" {
		event.Kind = "annotation"
	}
	if event.Severity == "" {
		event.Severity = SeverityInfo
	}
	if event.Direction == "" {
		event.Direction = DirectionUnknown
	}

	var blobBytes uint64
	if raw != nil {
		ref, err := r.storeBlobLocked(ctx, raw, expectedSize, mediaType, representation, sequence)
		if err != nil {
			r.manifest.Counts.WriteErrors++
			r.manifest.Completeness.Complete = false
			r.failed = err
			return Event{}, err
		}
		if ref.Size < 0 {
			return Event{}, errors.New("stored blob size is negative")
		}
		blobBytes = uint64(ref.Size)
		event.Blob = &ref
	}

	line, err := json.Marshal(event)
	if err != nil {
		r.manifest.Counts.WriteErrors++
		r.manifest.Completeness.Complete = false
		return Event{}, fmt.Errorf("encode event %d: %w", event.Sequence, err)
	}
	line = append(line, '\n')
	if err := writeAll(r.events, line); err != nil {
		r.manifest.Counts.WriteErrors++
		r.manifest.Completeness.Complete = false
		r.failed = fmt.Errorf("write event %d: %w", event.Sequence, err)
		return Event{}, r.failed
	}
	if r.manifest.Options.SyncEachEvent {
		if err := r.events.Sync(); err != nil {
			r.manifest.Counts.WriteErrors++
			r.manifest.Completeness.Complete = false
			r.failed = fmt.Errorf("sync event %d: %w", event.Sequence, err)
			return Event{}, r.failed
		}
	}
	r.sequence = sequence
	r.manifest.Counts.Events++
	if event.Blob != nil {
		if _, seen := r.seenBlob[event.Blob.SHA256]; !seen {
			r.seenBlob[event.Blob.SHA256] = struct{}{}
			r.manifest.Counts.Blobs++
			r.manifest.Counts.BlobBytes += blobBytes
		}
	}
	if event.Kind == "packet.decode_error" {
		r.manifest.Counts.DecodeErrors++
	}
	return event, nil
}

func (r *Recorder) AddLimitation(message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return errors.New("limitation is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrClosed
	}
	for _, existing := range r.manifest.Completeness.Limitations {
		if existing == message {
			return nil
		}
	}
	r.manifest.Completeness.Complete = false
	r.manifest.Completeness.Limitations = append(r.manifest.Completeness.Limitations, message)
	return nil
}

func (r *Recorder) AddLoss(dropped, truncated uint64, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrClosed
	}
	r.manifest.Counts.Dropped += dropped
	r.manifest.Counts.Truncated += truncated
	if dropped != 0 || truncated != 0 {
		r.manifest.Completeness.Complete = false
		if reason != "" {
			r.manifest.Completeness.Limitations = append(r.manifest.Completeness.Limitations, reason)
		}
	}
	return nil
}

func (r *Recorder) Close(status string, cause error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	if status == "" {
		status = "closed"
	}
	if r.failed != nil && cause == nil {
		cause = r.failed
		status = "failed"
	}
	r.manifest.Status = status
	if cause != nil {
		r.manifest.Failure = cause.Error()
		r.manifest.Completeness.Complete = false
	}
	r.manifest.EndedAt = r.now().UTC().Format(time.RFC3339Nano)

	var result error
	if err := r.events.Sync(); err != nil {
		result = errors.Join(result, fmt.Errorf("sync event stream: %w", err))
	}
	if err := r.events.Close(); err != nil {
		result = errors.Join(result, fmt.Errorf("close event stream: %w", err))
	}
	if err := r.writeManifestLocked(); err != nil {
		result = errors.Join(result, err)
	}
	return result
}

func (r *Recorder) storeBlobLocked(ctx context.Context, source io.Reader, expectedSize int64, mediaType, representation string, sequence uint64) (BlobRef, error) {
	temporary := filepath.Join(r.root, filepath.FromSlash(blobRoot), fmt.Sprintf(".tmp-%d", sequence))
	f, err := os.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return BlobRef{}, fmt.Errorf("create blob temporary file: %w", err)
	}
	cleanup := func() {
		_ = f.Close()
		_ = os.Remove(temporary)
	}
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(f, hash), &contextReader{ctx: ctx, source: source})
	if err != nil {
		cleanup()
		return BlobRef{}, fmt.Errorf("write blob stream: %w", err)
	}
	if expectedSize >= 0 && written != expectedSize {
		cleanup()
		return BlobRef{}, fmt.Errorf("raw stream size is %d, expected %d", written, expectedSize)
	}
	if err := f.Sync(); err != nil {
		cleanup()
		return BlobRef{}, fmt.Errorf("sync blob stream: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(temporary)
		return BlobRef{}, fmt.Errorf("close blob stream: %w", err)
	}
	hexDigest := hex.EncodeToString(hash.Sum(nil))
	relative := filepath.ToSlash(filepath.Join(blobRoot, hexDigest[:2], hexDigest+".bin"))
	destination := filepath.Join(r.root, filepath.FromSlash(relative))
	if info, err := os.Stat(destination); err == nil {
		_ = os.Remove(temporary)
		if info.Size() != written {
			return BlobRef{}, fmt.Errorf("existing blob %s has size %d, expected %d", hexDigest, info.Size(), written)
		}
		return BlobRef{SHA256: hexDigest, Size: written, Path: relative, MediaType: mediaType, Representation: representation}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(temporary)
		return BlobRef{}, fmt.Errorf("inspect blob %s: %w", hexDigest, err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		_ = os.Remove(temporary)
		return BlobRef{}, fmt.Errorf("create blob directory: %w", err)
	}
	if err := os.Rename(temporary, destination); err != nil {
		_ = os.Remove(temporary)
		return BlobRef{}, fmt.Errorf("publish blob %s: %w", hexDigest, err)
	}
	return BlobRef{SHA256: hexDigest, Size: written, Path: relative, MediaType: mediaType, Representation: representation}, nil
}

type contextReader struct {
	ctx    context.Context
	source io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.source.Read(buffer)
	}
}

func (r *Recorder) writeManifestLocked() error {
	data, err := json.MarshalIndent(r.manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	data = append(data, '\n')
	temporary := filepath.Join(r.root, manifestName+".tmp")
	f, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create manifest temporary file: %w", err)
	}
	cleanup := func() {
		_ = f.Close()
		_ = os.Remove(temporary)
	}
	if err := writeAll(f, data); err != nil {
		cleanup()
		return fmt.Errorf("write manifest: %w", err)
	}
	if err := f.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync manifest: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("close manifest: %w", err)
	}
	destination := filepath.Join(r.root, manifestName)
	if err := os.Rename(temporary, destination); err != nil {
		_ = os.Remove(destination)
		if retryErr := os.Rename(temporary, destination); retryErr != nil {
			_ = os.Remove(temporary)
			return fmt.Errorf("publish manifest: %w", retryErr)
		}
	}
	return nil
}

func requireEmptyOrMissingDirectory(path string) error {
	entries, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect capture root: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("capture root %q is not empty", path)
	}
	return nil
}

func randomID() (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(rand.Reader, value[:]); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) != 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}

func cloneMap(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	destination := make(map[string]string, len(source))
	for key, value := range source {
		destination[key] = value
	}
	return destination
}

func cloneManifest(source Manifest) Manifest {
	source.Options.RawLayers = append([]string(nil), source.Options.RawLayers...)
	source.Options.Values = cloneMap(source.Options.Values)
	source.Completeness.Limitations = append([]string(nil), source.Completeness.Limitations...)
	return source
}
