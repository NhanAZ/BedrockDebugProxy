package artifacts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

// Blob reads are confined to the capture root and checked against recorded bytes.
func (w *Live) openBlob(ref *capture.BlobRef) (*os.File, error) {
	if ref == nil || len(ref.SHA256) != sha256.Size*2 || ref.Size < 0 {
		return nil, errors.New("invalid source blob")
	}
	if _, err := hex.DecodeString(ref.SHA256); err != nil {
		return nil, err
	}
	expected := "blobs/sha256/" + ref.SHA256[:2] + "/" + ref.SHA256 + ".bin"
	if ref.Path != expected || strings.ToLower(ref.SHA256) != ref.SHA256 {
		return nil, errors.New("non-canonical source blob path")
	}
	f, err := w.source.Open(expected)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("stat source blob: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() != ref.Size {
		_ = f.Close()
		return nil, errors.New("source blob size or file type mismatch")
	}
	return f, nil
}

func (w *Live) readSkinBlob(ref *capture.BlobRef) ([]byte, error) {
	if ref == nil || ref.Size > maxSkinPacketBytes {
		return nil, errors.New("skin packet exceeds the 16 MiB view limit")
	}
	f, err := w.openBlob(ref)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, maxSkinPacketBytes+1))
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	if int64(len(data)) != ref.Size || hex.EncodeToString(hash[:]) != ref.SHA256 {
		return nil, errors.New("skin source blob checksum mismatch")
	}
	return data, nil
}

// Each file is a separate copy. Hard links would let editing an artifact damage
// the original capture. No pack entry paths are extracted to the filesystem.
func (w *Live) publish(path string, write func(io.Writer) error) error {
	if slash := strings.LastIndexByte(path, '/'); slash >= 0 {
		if err := w.output.MkdirAll(path[:slash], 0o700); err != nil {
			return err
		}
	}
	if info, err := w.output.Lstat(path); err == nil {
		if !info.Mode().IsRegular() {
			return errors.New("artifact destination is not a regular file")
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	temporary := path + ".tmp"
	f, err := w.output.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	err = write(f)
	err = errors.Join(err, f.Close())
	if err == nil {
		err = w.output.Rename(temporary, path)
	}
	if err != nil {
		_ = w.output.Remove(temporary)
	}
	return err
}

func (w *Live) archive(event capture.Event) error {
	var metadata struct {
		Name      string `json:"name"`
		UUID      string `json:"uuid"`
		Version   string `json:"version"`
		Encrypted bool   `json:"encrypted"`
	}
	if err := json.Unmarshal(event.Data, &metadata); err != nil {
		return err
	}
	f, err := w.openBlob(event.Blob)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	variant := "original"
	if event.Kind == "resource_pack.decrypted_archive" {
		variant = "decrypted"
	}
	name := metadata.Name
	if name == "" {
		name = metadata.UUID
	}
	path := "packs/" + variant + "/" + safeName(name) + "-" + event.Blob.SHA256 + ".zip"
	if err := w.publish(path, func(dst io.Writer) error {
		hash := sha256.New()
		n, err := io.Copy(io.MultiWriter(dst, hash), io.LimitReader(f, event.Blob.Size+1))
		if err != nil {
			return err
		}
		if n != event.Blob.Size || hex.EncodeToString(hash.Sum(nil)) != event.Blob.SHA256 {
			return errors.New("pack source blob checksum mismatch")
		}
		return nil
	}); err != nil {
		return err
	}
	return w.appendJSON("packs/index.jsonl", map[string]any{
		"source_sequence": event.Sequence, "parent_sequence": event.ParentSequence,
		"connection_id": event.ConnectionID, "direction": event.Direction, "time": event.Time,
		"uuid": metadata.UUID, "name": metadata.Name, "version": metadata.Version,
		"encrypted": metadata.Encrypted, "variant": variant, "source_blob": event.Blob,
		"path": "artifacts/" + path,
	})
}

func safeName(name string) string {
	var result strings.Builder
	for _, c := range name {
		if result.Len() >= 40 {
			break
		}
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
			result.WriteRune(c)
		} else {
			result.WriteByte('_')
		}
	}
	if result.Len() == 0 {
		return "pack"
	}
	return result.String()
}
