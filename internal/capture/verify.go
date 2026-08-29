package capture

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Verification struct {
	Manifest Manifest
	Events   uint64
	Blobs    uint64
	Issues   []string
}

func ReadManifest(root string) (Manifest, error) {
	data, err := os.ReadFile(filepath.Join(root, manifestName))
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return manifest, nil
}

func ScanEvents(root string, visit func(Event) error) error {
	f, err := os.Open(filepath.Join(root, eventsName))
	if err != nil {
		return fmt.Errorf("open event stream: %w", err)
	}
	defer func() { _ = f.Close() }()

	reader := bufio.NewReaderSize(f, 256*1024)
	lineNumber := 0
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) != 0 {
			lineNumber++
			if line[len(line)-1] != '\n' {
				return fmt.Errorf("event line %d is not newline terminated", lineNumber)
			}
			var event Event
			if err := json.Unmarshal(line, &event); err != nil {
				return fmt.Errorf("decode event line %d: %w", lineNumber, err)
			}
			if err := visit(event); err != nil {
				return fmt.Errorf("visit event line %d: %w", lineNumber, err)
			}
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read event line %d: %w", lineNumber+1, readErr)
		}
	}
}

func Verify(root string) (Verification, error) {
	manifest, err := ReadManifest(root)
	if err != nil {
		return Verification{}, err
	}
	result := Verification{Manifest: manifest}
	seenBlobs := map[string]struct{}{}
	var previousSequence uint64
	err = ScanEvents(root, func(event Event) error {
		result.Events++
		if event.Schema != SchemaVersion {
			result.Issues = append(result.Issues, fmt.Sprintf("event %d has schema %q", event.Sequence, event.Schema))
		}
		if event.CaptureID != manifest.CaptureID {
			result.Issues = append(result.Issues, fmt.Sprintf("event %d has capture ID %q", event.Sequence, event.CaptureID))
		}
		if event.Sequence != previousSequence+1 {
			result.Issues = append(result.Issues, fmt.Sprintf("event sequence jumps from %d to %d", previousSequence, event.Sequence))
		}
		previousSequence = event.Sequence
		if event.Blob == nil {
			return nil
		}
		if _, ok := seenBlobs[event.Blob.SHA256]; ok {
			return nil
		}
		seenBlobs[event.Blob.SHA256] = struct{}{}
		result.Blobs++
		if err := verifyBlob(root, *event.Blob); err != nil {
			result.Issues = append(result.Issues, fmt.Sprintf("event %d blob is invalid - %v", event.Sequence, err))
		}
		return nil
	})
	if err != nil {
		return Verification{}, err
	}
	if manifest.Schema != SchemaVersion {
		result.Issues = append(result.Issues, fmt.Sprintf("manifest has schema %q", manifest.Schema))
	}
	if manifest.Counts.Events != result.Events {
		result.Issues = append(result.Issues, fmt.Sprintf("manifest counts %d events but stream has %d", manifest.Counts.Events, result.Events))
	}
	if manifest.Counts.Blobs != result.Blobs {
		result.Issues = append(result.Issues, fmt.Sprintf("manifest counts %d blobs but events reference %d unique blobs", manifest.Counts.Blobs, result.Blobs))
	}
	return result, nil
}

func verifyBlob(root string, ref BlobRef) error {
	if len(ref.SHA256) != sha256.Size*2 {
		return fmt.Errorf("invalid SHA-256 length %d", len(ref.SHA256))
	}
	if _, err := hex.DecodeString(ref.SHA256); err != nil {
		return fmt.Errorf("invalid SHA-256 text: %w", err)
	}
	cleanRelative := filepath.Clean(filepath.FromSlash(ref.Path))
	if filepath.IsAbs(cleanRelative) || cleanRelative == ".." || strings.HasPrefix(cleanRelative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("unsafe relative path %q", ref.Path)
	}
	expectedRelative := filepath.Clean(filepath.FromSlash(filepath.ToSlash(filepath.Join(blobRoot, ref.SHA256[:2], ref.SHA256+".bin"))))
	if cleanRelative != expectedRelative {
		return fmt.Errorf("non-canonical path %q", ref.Path)
	}
	f, err := os.Open(filepath.Join(root, cleanRelative))
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()
	hash := sha256.New()
	n, err := io.Copy(hash, f)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if n != ref.Size {
		return fmt.Errorf("size is %d, expected %d", n, ref.Size)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != ref.SHA256 {
		return fmt.Errorf("SHA-256 is %s, expected %s", actual, ref.SHA256)
	}
	return nil
}
