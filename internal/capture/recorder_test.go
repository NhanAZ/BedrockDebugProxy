package capture

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecorderWritesOrderedEventsAndDeduplicatedBlobs(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	base := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.FixedZone("test", 7*60*60))
	clock := sequenceClock(base, time.Millisecond)
	recorder, err := New(root, Options{
		CaptureID:     "test-capture",
		GeneratorName: "capture-test",
		RawLayers:     []string{"packet_payload"},
		Now:           clock,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	raw := []byte{0x01, 0x02, 0x03}
	first, err := recorder.Record(context.Background(), Record{
		Event: Event{
			SessionID: "session-1",
			Kind:      "packet.raw",
			Direction: DirectionClientToServer,
		},
		Raw:            raw,
		MediaType:      "application/octet-stream",
		Representation: "packet_payload",
	})
	if err != nil {
		t.Fatalf("Record(first) error = %v", err)
	}
	second, err := recorder.Record(context.Background(), Record{
		Event: Event{
			SessionID: "session-1",
			Kind:      "packet.raw",
			Direction: DirectionServerToClient,
		},
		Raw:            append([]byte(nil), raw...),
		MediaType:      "application/octet-stream",
		Representation: "packet_payload",
	})
	if err != nil {
		t.Fatalf("Record(second) error = %v", err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if first.Sequence != 1 || second.Sequence != 2 {
		t.Fatalf("sequences = %d, %d, want 1, 2", first.Sequence, second.Sequence)
	}
	if first.ElapsedNano != int64(time.Millisecond) || second.ElapsedNano != int64(2*time.Millisecond) {
		t.Fatalf("elapsed = %d, %d", first.ElapsedNano, second.ElapsedNano)
	}
	if first.Blob == nil || second.Blob == nil || first.Blob.SHA256 != second.Blob.SHA256 {
		t.Fatalf("blob references were not deduplicated: %#v %#v", first.Blob, second.Blob)
	}
	manifest, err := ReadManifest(root)
	if err != nil {
		t.Fatalf("ReadManifest() error = %v", err)
	}
	if manifest.Status != "closed" || manifest.Counts.Events != 2 || manifest.Counts.Blobs != 1 || manifest.Counts.BlobBytes != 3 {
		t.Fatalf("unexpected manifest: %#v", manifest)
	}
	if !manifest.Completeness.Complete {
		t.Fatalf("capture should be complete: %#v", manifest.Completeness)
	}
	verification, err := Verify(root)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(verification.Issues) != 0 {
		t.Fatalf("Verify() issues = %v", verification.Issues)
	}
}

func TestRecorderMarksLimitationsAndFailures(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := New(root, Options{Limitations: []string{"raw UDP is unavailable"}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := recorder.AddLoss(2, 1, "test loss"); err != nil {
		t.Fatalf("AddLoss() error = %v", err)
	}
	if err := recorder.Close("failed", errors.New("test failure")); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	manifest, err := ReadManifest(root)
	if err != nil {
		t.Fatalf("ReadManifest() error = %v", err)
	}
	if manifest.Completeness.Complete || manifest.Counts.Dropped != 2 || manifest.Counts.Truncated != 1 {
		t.Fatalf("unexpected completeness: %#v %#v", manifest.Completeness, manifest.Counts)
	}
	if manifest.Failure != "test failure" || manifest.Status != "failed" {
		t.Fatalf("unexpected failure: %#v", manifest)
	}
}

func TestRecorderRejectsNonEmptyRootAndWritesEmptyBlob(t *testing.T) {
	nonEmpty := t.TempDir()
	if err := os.WriteFile(filepath.Join(nonEmpty, "existing"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := New(nonEmpty, Options{}); err == nil {
		t.Fatal("New() succeeded for non-empty root")
	}

	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := New(root, Options{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	event, err := recorder.Record(context.Background(), Record{Event: Event{Kind: "empty"}, Raw: []byte{}})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if event.Blob == nil || event.Blob.Size != 0 {
		t.Fatalf("empty blob = %#v", event.Blob)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
}

func TestRecordHonorsCancelledContextAndClosedState(t *testing.T) {
	recorder, err := New(filepath.Join(t.TempDir(), "capture"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := recorder.Record(ctx, Record{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Record() error = %v, want context.Canceled", err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), Record{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("Record() error = %v, want ErrClosed", err)
	}
}

func TestEventDataRemainsValidJSON(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := New(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	data := json.RawMessage(`{"packet":"value"}`)
	if _, err := recorder.Record(context.Background(), Record{Event: Event{Kind: "decoded", Data: data}}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if err := ScanEvents(root, func(event Event) error {
		if string(event.Data) != string(data) {
			t.Fatalf("data = %s, want %s", event.Data, data)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidEventDataDoesNotConsumeSequence(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := New(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), Record{Event: Event{Kind: "invalid", Data: json.RawMessage(`{"broken"`)}}); err == nil {
		t.Fatal("Record() accepted invalid JSON")
	}
	event, err := recorder.Record(context.Background(), Record{Event: Event{Kind: "valid"}})
	if err != nil {
		t.Fatal(err)
	}
	if event.Sequence != 1 {
		t.Fatalf("sequence = %d, want 1", event.Sequence)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	verification, err := Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification.Issues) != 0 {
		t.Fatalf("Verify() issues = %v", verification.Issues)
	}
}

func sequenceClock(start time.Time, step time.Duration) func() time.Time {
	current := start.Add(-step)
	return func() time.Time {
		current = current.Add(step)
		return current
	}
}
