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
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	manifestName = "manifest.json"
	eventsName   = "events.jsonl"
	blobRoot     = "blobs/sha256"

	defaultWriterQueueCapacity = 65536
	defaultWriterQueueBytes    = 64 << 20
)

var ErrClosed = errors.New("capture recorder is closed")

type Options struct {
	CaptureID     string
	GeneratorName string
	Version       string
	Commit        string
	SyncEachEvent bool
	// WriterQueueCapacity bounds the lossless asynchronous event writer queue.
	// A zero value selects the default capacity.
	WriterQueueCapacity int
	// WriterQueueBytes bounds the lossless asynchronous writer memory budget.
	// A zero value selects the default budget. A single larger record is allowed
	// when the queue is otherwise empty.
	WriterQueueBytes int64
	RawLayers        []string
	Values           map[string]string
	Limitations      []string
	Now              func() time.Time
}

type Recorder struct {
	mu               sync.Mutex
	root             string
	events           *os.File
	start            time.Time
	now              func() time.Time
	sequence         uint64
	assignedSequence uint64
	seenBlob         map[string]struct{}
	blobDirs         map[string]struct{}
	failed           error
	closed           bool
	manifest         Manifest

	enqueueMu         sync.Mutex
	queue             chan recordJob
	writerDone        chan struct{}
	writerStopped     chan struct{}
	writerStoppedOnce sync.Once
	idle              chan struct{}
	pending           int
	pendingBytes      int64
	queuePeakRecords  int
	queuePeakBytes    int64
	queueByteLimit    int64
	spaceChanged      chan struct{}
}

type recordJob struct {
	event          Event
	raw            []byte
	hasRaw         bool
	mediaType      string
	representation string
	bytes          int64
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
	queueCapacity := options.WriterQueueCapacity
	if queueCapacity <= 0 {
		queueCapacity = defaultWriterQueueCapacity
	}
	queueBytes := options.WriterQueueBytes
	if queueBytes <= 0 {
		queueBytes = defaultWriterQueueBytes
	}
	values := cloneMap(options.Values)
	if values == nil {
		values = make(map[string]string)
	}
	values["capture_writer"] = "ordered_async"
	values["capture_writer_queue_capacity"] = strconv.Itoa(queueCapacity)
	values["capture_writer_queue_bytes"] = strconv.FormatInt(queueBytes, 10)
	values["capture_writer_backpressure"] = "block"
	values["capture_writer_loss_policy"] = "never_drop"

	r := &Recorder{
		root:     root,
		events:   events,
		start:    start,
		now:      now,
		seenBlob: make(map[string]struct{}),
		blobDirs: make(map[string]struct{}),
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
				Values:        values,
			},
			EventsPath: eventsName,
			BlobRoot:   blobRoot,
			Completeness: Completeness{
				Complete:    len(options.Limitations) == 0,
				Limitations: append([]string(nil), options.Limitations...),
			},
		},
		queue:          make(chan recordJob, queueCapacity),
		writerDone:     make(chan struct{}),
		writerStopped:  make(chan struct{}),
		queueByteLimit: queueBytes,
		spaceChanged:   make(chan struct{}),
	}
	r.idle = make(chan struct{})
	close(r.idle)
	if err := r.writeManifestLocked(); err != nil {
		_ = events.Close()
		return nil, err
	}
	go r.runWriter()
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
	if record.Raw == nil {
		return r.enqueue(ctx, record.Event, nil, false, record.MediaType, record.Representation)
	}
	return r.enqueue(ctx, record.Event, record.Raw, true, record.MediaType, record.Representation)
}

// RecordReader records an event while streaming its raw blob from raw. expectedSize may be -1 when the size is unknown.
func (r *Recorder) RecordReader(ctx context.Context, event Event, raw io.Reader, expectedSize int64, mediaType, representation string) (Event, error) {
	if raw == nil {
		return Event{}, errors.New("raw reader is nil")
	}
	if expectedSize < -1 {
		return Event{}, errors.New("expected raw size must be -1 or greater")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.enqueueMu.Lock()
	defer r.enqueueMu.Unlock()
	if err := r.waitForIdle(ctx); err != nil {
		return Event{}, err
	}
	return r.recordSync(ctx, event, raw, expectedSize, mediaType, representation)
}

func (r *Recorder) enqueue(ctx context.Context, event Event, raw []byte, hasRaw bool, mediaType, representation string) (Event, error) {
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
	if _, err := json.Marshal(event); err != nil {
		return Event{}, fmt.Errorf("encode event: %w", err)
	}
	event = cloneEvent(event)
	if hasRaw {
		// The caller may reuse its packet buffer as soon as Record returns.
		raw = append([]byte(nil), raw...)
	}

	r.enqueueMu.Lock()
	defer r.enqueueMu.Unlock()
	if err := r.waitForQueueSpace(ctx, int64(len(raw))); err != nil {
		return Event{}, err
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return Event{}, ErrClosed
	}
	if r.failed != nil {
		err := fmt.Errorf("capture recorder failed: %w", r.failed)
		r.mu.Unlock()
		return Event{}, err
	}
	event, err := r.prepareEventLocked(event)
	if err != nil {
		r.mu.Unlock()
		return Event{}, err
	}
	if hasRaw {
		digest := sha256.Sum256(raw)
		ref := r.blobRef(hex.EncodeToString(digest[:]), int64(len(raw)), mediaType, representation)
		event.Blob = &ref
	}
	job := recordJob{
		event:          event,
		raw:            raw,
		hasRaw:         hasRaw,
		mediaType:      mediaType,
		representation: representation,
		bytes:          int64(len(raw)),
	}
	r.assignedSequence = event.Sequence
	if r.pending == 0 {
		r.idle = make(chan struct{})
	}
	r.pending++
	r.pendingBytes += job.bytes
	if r.pending > r.queuePeakRecords {
		r.queuePeakRecords = r.pending
	}
	if r.pendingBytes > r.queuePeakBytes {
		r.queuePeakBytes = r.pendingBytes
	}
	r.mu.Unlock()

	select {
	case r.queue <- job:
		return event, nil
	case <-ctx.Done():
		r.rollbackEnqueue(event.Sequence, job.bytes)
		return Event{}, ctx.Err()
	case <-r.writerStopped:
		r.rollbackEnqueue(event.Sequence, job.bytes)
		r.mu.Lock()
		err = r.failed
		r.mu.Unlock()
		if err == nil {
			err = errors.New("capture writer stopped")
		}
		return Event{}, fmt.Errorf("capture recorder failed: %w", err)
	}
}

func (r *Recorder) recordSync(ctx context.Context, event Event, raw io.Reader, expectedSize int64, mediaType, representation string) (Event, error) {
	if event.Blob != nil {
		return Event{}, errors.New("event blob references are recorder-owned")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := json.Marshal(event); err != nil {
		return Event{}, fmt.Errorf("encode event: %w", err)
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

	event, err := r.prepareEventLocked(event)
	if err != nil {
		return Event{}, err
	}
	r.assignedSequence = event.Sequence

	var blobBytes uint64
	if raw != nil {
		ref, err := r.storeBlobLocked(ctx, raw, expectedSize, mediaType, representation, event.Sequence)
		if err != nil {
			r.markFailedLocked(err)
			return Event{}, err
		}
		if ref.Size < 0 {
			return Event{}, errors.New("stored blob size is negative")
		}
		blobBytes = uint64(ref.Size)
		event.Blob = &ref
	}
	return r.appendEventLocked(event, blobBytes)
}

func (r *Recorder) prepareEventLocked(event Event) (Event, error) {
	if event.Data != nil {
		var value any
		if err := json.Unmarshal(event.Data, &value); err != nil {
			return Event{}, fmt.Errorf("event data is invalid JSON: %w", err)
		}
	}
	now := r.now()
	event.Schema = SchemaVersion
	event.Sequence = r.assignedSequence + 1
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
	return event, nil
}

func (r *Recorder) appendEventLocked(event Event, blobBytes uint64) (Event, error) {
	if event.Sequence != r.sequence+1 {
		return Event{}, fmt.Errorf("event sequence %d follows committed sequence %d", event.Sequence, r.sequence)
	}

	line, err := json.Marshal(event)
	if err != nil {
		return Event{}, fmt.Errorf("encode event %d: %w", event.Sequence, err)
	}
	line = append(line, '\n')
	if err := writeAll(r.events, line); err != nil {
		r.markFailedLocked(fmt.Errorf("write event %d: %w", event.Sequence, err))
		return Event{}, r.failed
	}
	if r.manifest.Options.SyncEachEvent {
		if err := r.events.Sync(); err != nil {
			r.markFailedLocked(fmt.Errorf("sync event %d: %w", event.Sequence, err))
			return Event{}, r.failed
		}
	}
	r.sequence = event.Sequence
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

func (r *Recorder) runWriter() {
	defer close(r.writerDone)
	for job := range r.queue {
		r.mu.Lock()
		if r.failed == nil && !r.closed {
			_, err := r.writeJobLocked(job)
			if err != nil {
				r.markFailedLocked(err)
			}
		} else if r.failed == nil {
			r.markFailedLocked(ErrClosed)
		}
		if r.pending > 0 {
			r.pending--
		}
		r.pendingBytes -= job.bytes
		if r.pending == 0 {
			close(r.idle)
		}
		changed := r.spaceChanged
		r.spaceChanged = make(chan struct{})
		failed := r.failed != nil
		r.mu.Unlock()
		close(changed)
		if failed {
			r.writerStoppedOnce.Do(func() { close(r.writerStopped) })
			return
		}
	}
}

func (r *Recorder) writeJobLocked(job recordJob) (Event, error) {
	var blobBytes uint64
	event := job.event
	if job.hasRaw {
		ref, err := r.storeBlobBytesLocked(context.Background(), job.raw, job.mediaType, job.representation, event.Sequence)
		if err != nil {
			return Event{}, err
		}
		if ref.Size < 0 {
			return Event{}, errors.New("stored blob size is negative")
		}
		blobBytes = uint64(ref.Size)
		event.Blob = &ref
	}
	return r.appendEventLocked(event, blobBytes)
}

func (r *Recorder) waitForIdle(ctx context.Context) error {
	for {
		r.mu.Lock()
		if r.closed {
			r.mu.Unlock()
			return ErrClosed
		}
		if r.failed != nil {
			err := fmt.Errorf("capture recorder failed: %w", r.failed)
			r.mu.Unlock()
			return err
		}
		if r.pending == 0 {
			r.mu.Unlock()
			return nil
		}
		idle := r.idle
		stopped := r.writerStopped
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-idle:
		case <-stopped:
		}
	}
}

func (r *Recorder) waitForQueueSpace(ctx context.Context, bytes int64) error {
	for {
		r.mu.Lock()
		if r.closed {
			r.mu.Unlock()
			return ErrClosed
		}
		if r.failed != nil {
			err := fmt.Errorf("capture recorder failed: %w", r.failed)
			r.mu.Unlock()
			return err
		}
		withinRecords := r.pending < cap(r.queue)
		withinBytes := r.pendingBytes+bytes <= r.queueByteLimit || r.pending == 0
		if withinRecords && withinBytes {
			r.mu.Unlock()
			return nil
		}
		changed := r.spaceChanged
		stopped := r.writerStopped
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		case <-stopped:
		}
	}
}

func (r *Recorder) rollbackEnqueue(sequence uint64, bytes int64) {
	r.mu.Lock()
	if r.assignedSequence == sequence {
		r.assignedSequence--
	}
	if r.pending > 0 {
		r.pending--
	}
	r.pendingBytes -= bytes
	if r.pending == 0 {
		close(r.idle)
	}
	changed := r.spaceChanged
	r.spaceChanged = make(chan struct{})
	r.mu.Unlock()
	close(changed)
}

func (r *Recorder) markFailedLocked(err error) {
	if err == nil {
		return
	}
	if r.failed == nil {
		r.failed = err
		r.manifest.Counts.WriteErrors++
	}
	r.manifest.Completeness.Complete = false
}

func (r *Recorder) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failed == nil {
		return nil
	}
	return fmt.Errorf("capture recorder failed: %w", r.failed)
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
	r.enqueueMu.Lock()
	defer r.enqueueMu.Unlock()

	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	queue := r.queue
	writerDone := r.writerDone
	close(queue)
	r.mu.Unlock()
	<-writerDone

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
	if r.failed != nil && status == "closed" {
		status = "failed"
		r.manifest.Status = status
	}
	if r.manifest.Options.Values == nil {
		r.manifest.Options.Values = make(map[string]string)
	}
	r.manifest.Options.Values["capture_writer_queue_peak_records"] = strconv.Itoa(r.queuePeakRecords)
	r.manifest.Options.Values["capture_writer_queue_peak_bytes"] = strconv.FormatInt(r.queuePeakBytes, 10)
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

func (r *Recorder) storeBlobBytesLocked(ctx context.Context, payload []byte, mediaType, representation string, sequence uint64) (BlobRef, error) {
	digest := sha256.Sum256(payload)
	hexDigest := hex.EncodeToString(digest[:])
	ref := r.blobRef(hexDigest, int64(len(payload)), mediaType, representation)
	if _, seen := r.seenBlob[hexDigest]; seen {
		return ref, nil
	}
	return r.storeBlobLocked(ctx, bytes.NewReader(payload), int64(len(payload)), mediaType, representation, sequence)
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
	if r.manifest.Options.SyncEachEvent {
		if err := f.Sync(); err != nil {
			cleanup()
			return BlobRef{}, fmt.Errorf("sync blob stream: %w", err)
		}
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(temporary)
		return BlobRef{}, fmt.Errorf("close blob stream: %w", err)
	}
	hexDigest := hex.EncodeToString(hash.Sum(nil))
	ref := r.blobRef(hexDigest, written, mediaType, representation)
	destination := filepath.Join(r.root, filepath.FromSlash(ref.Path))
	if info, err := os.Stat(destination); err == nil {
		_ = os.Remove(temporary)
		if info.Size() != written {
			return BlobRef{}, fmt.Errorf("existing blob %s has size %d, expected %d", hexDigest, info.Size(), written)
		}
		return ref, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(temporary)
		return BlobRef{}, fmt.Errorf("inspect blob %s: %w", hexDigest, err)
	}
	if err := r.ensureBlobDirectoryLocked(hexDigest); err != nil {
		_ = os.Remove(temporary)
		return BlobRef{}, err
	}
	if err := os.Rename(temporary, destination); err != nil {
		_ = os.Remove(temporary)
		return BlobRef{}, fmt.Errorf("publish blob %s: %w", hexDigest, err)
	}
	return ref, nil
}

func (r *Recorder) blobRef(hexDigest string, size int64, mediaType, representation string) BlobRef {
	relative := filepath.ToSlash(filepath.Join(blobRoot, hexDigest[:2], hexDigest+".bin"))
	return BlobRef{SHA256: hexDigest, Size: size, Path: relative, MediaType: mediaType, Representation: representation}
}

func (r *Recorder) ensureBlobDirectoryLocked(hexDigest string) error {
	prefix := hexDigest[:2]
	if _, exists := r.blobDirs[prefix]; exists {
		return nil
	}
	directory := filepath.Join(r.root, filepath.FromSlash(blobRoot), prefix)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create blob directory: %w", err)
	}
	r.blobDirs[prefix] = struct{}{}
	return nil
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

func cloneEvent(source Event) Event {
	destination := source
	destination.Data = append(json.RawMessage(nil), source.Data...)
	destination.Annotations = append([]string(nil), source.Annotations...)
	if source.Source != nil {
		value := *source.Source
		destination.Source = &value
	}
	if source.Destination != nil {
		value := *source.Destination
		destination.Destination = &value
	}
	if source.ParentSequence != nil {
		value := *source.ParentSequence
		destination.ParentSequence = &value
	}
	if source.Packet != nil {
		value := *source.Packet
		destination.Packet = &value
	}
	if source.Blob != nil {
		value := *source.Blob
		destination.Blob = &value
	}
	if source.Error != nil {
		value := *source.Error
		if source.Error.Temporary != nil {
			temporary := *source.Error.Temporary
			value.Temporary = &temporary
		}
		if source.Error.Timeout != nil {
			timeout := *source.Error.Timeout
			value.Timeout = &timeout
		}
		destination.Error = &value
	}
	return destination
}

func cloneManifest(source Manifest) Manifest {
	source.Options.RawLayers = append([]string(nil), source.Options.RawLayers...)
	source.Options.Values = cloneMap(source.Options.Values)
	source.Completeness.Limitations = append([]string(nil), source.Completeness.Limitations...)
	return source
}
