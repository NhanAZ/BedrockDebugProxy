package bedrock

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	packdecrypt "github.com/NhanAZ/BedrockDebugProxy/internal/resourcepack"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

type ResourcePackCaptureOptions struct {
	Decrypt bool
}

type resourcePackData struct {
	Index          int               `json:"index"`
	Total          int               `json:"total"`
	UUID           string            `json:"uuid"`
	Version        string            `json:"version"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	ArchiveBytes   int               `json:"archive_bytes"`
	ChecksumSHA256 string            `json:"checksum_sha256"`
	DownloadURL    string            `json:"download_url,omitempty"`
	Delivery       string            `json:"delivery"`
	Encrypted      bool              `json:"encrypted"`
	ContentKey     string            `json:"content_key,omitempty"`
	HasScripts     bool              `json:"has_scripts"`
	HasBehaviours  bool              `json:"has_behaviours"`
	HasTextures    bool              `json:"has_textures"`
	WorldTemplate  bool              `json:"world_template"`
	Manifest       resource.Manifest `json:"manifest"`
}

type resourcePackDerivedData struct {
	UUID    string `json:"uuid"`
	Version string `json:"version"`
	Path    string `json:"path,omitempty"`
	Report  any    `json:"report,omitempty"`
}

func RecordResourcePacks(ctx context.Context, recorder *capture.Recorder, sessionID, connectionID string, hop int, source, destination net.Addr, packs []*resource.Pack, options ResourcePackCaptureOptions) error {
	for index, pack := range packs {
		if pack == nil {
			return fmt.Errorf("resource pack %d of %d is nil", index+1, len(packs))
		}
		checksum := pack.Checksum()
		delivery := "raknet"
		if pack.DownloadURL() != "" {
			delivery = "http"
		}
		data, err := json.Marshal(resourcePackData{
			Index:          index + 1,
			Total:          len(packs),
			UUID:           pack.UUID().String(),
			Version:        pack.Version(),
			Name:           pack.Name(),
			Description:    pack.Description(),
			ArchiveBytes:   pack.Len(),
			ChecksumSHA256: hex.EncodeToString(checksum[:]),
			DownloadURL:    pack.DownloadURL(),
			Delivery:       delivery,
			Encrypted:      pack.Encrypted(),
			ContentKey:     pack.ContentKey(),
			HasScripts:     pack.HasScripts(),
			HasBehaviours:  pack.HasBehaviours(),
			HasTextures:    pack.HasTextures(),
			WorldTemplate:  pack.HasWorldTemplate(),
			Manifest:       pack.Manifest(),
		})
		if err != nil {
			return fmt.Errorf("encode resource pack %s metadata: %w", pack.UUID(), err)
		}
		event, err := recorder.RecordReader(ctx, capture.Event{
			SessionID:    sessionID,
			ConnectionID: connectionID,
			Hop:          hop,
			Kind:         "resource_pack.archive",
			Severity:     capture.SeverityInfo,
			Direction:    capture.DirectionServerToClient,
			Channel:      "upstream",
			Stage:        "post_resource_pack_download",
			Source:       endpoint(source),
			Destination:  endpoint(destination),
			Data:         data,
			Annotations:  []string{"Archive bytes are preserved before any decryption or extraction"},
		}, io.NewSectionReader(pack, 0, int64(pack.Len())), int64(pack.Len()), "application/zip", "minecraft_resource_pack_archive")
		if err != nil {
			return fmt.Errorf("record resource pack %s: %w", pack.UUID(), err)
		}
		if event.Blob == nil || event.Blob.SHA256 != hex.EncodeToString(checksum[:]) {
			return fmt.Errorf("resource pack %s checksum does not match its recorded archive", pack.UUID())
		}
		if pack.Encrypted() && options.Decrypt {
			if err := recordDecryptedResourcePack(ctx, recorder, sessionID, connectionID, hop, source, destination, pack, event.Sequence); err != nil {
				return err
			}
		}
	}
	return nil
}

func recordDecryptedResourcePack(ctx context.Context, recorder *capture.Recorder, sessionID, connectionID string, hop int, source, destination net.Addr, pack *resource.Pack, parentSequence uint64) error {
	// A seekable temporary file lets the capture recorder stream the derived ZIP
	// with a known size. Normal returns remove it, but abrupt process termination
	// can leave an operator-owned plaintext file in the OS temporary directory.
	temporary, err := os.CreateTemp("", "bedrock-debug-proxy-decrypted-pack-*.mcpack")
	if err != nil {
		return recordResourcePackDecryptFailure(ctx, recorder, sessionID, connectionID, hop, source, destination, pack, parentSequence, fmt.Errorf("create temporary decrypted archive: %w", err))
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()
		_ = os.Remove(temporaryPath)
	}()

	// In production this key is exactly the ContentKey retained by gophertunnel from
	// ResourcePacksInfo on the current upstream connection. Do not substitute an
	// offline lookup or recovery path without explicit maintainer review.
	result, decryptErr := packdecrypt.DecryptArchive(ctx, pack, int64(pack.Len()), pack.ContentKey(), temporary)
	if decryptErr == nil {
		decryptErr = temporary.Sync()
	}
	if decryptErr == nil {
		_, decryptErr = temporary.Seek(0, io.SeekStart)
	}
	if decryptErr != nil {
		if errors.Is(decryptErr, context.Canceled) || errors.Is(decryptErr, context.DeadlineExceeded) {
			return decryptErr
		}
		return recordResourcePackDecryptFailure(ctx, recorder, sessionID, connectionID, hop, source, destination, pack, parentSequence, decryptErr)
	}
	parent := parentSequence
	for _, manifest := range result.Manifests {
		data, err := json.Marshal(resourcePackDerivedData{
			UUID:    pack.UUID().String(),
			Version: pack.Version(),
			Path:    manifest.Path,
			Report:  manifest.ContentsSummary,
		})
		if err != nil {
			return fmt.Errorf("encode decrypted contents metadata for resource pack %s: %w", pack.UUID(), err)
		}
		if _, err := recorder.Record(ctx, capture.Record{
			Event: capture.Event{
				SessionID:      sessionID,
				ConnectionID:   connectionID,
				Hop:            hop,
				Kind:           "resource_pack.contents_manifest",
				Severity:       capture.SeverityInfo,
				Direction:      capture.DirectionServerToClient,
				Channel:        "upstream",
				Stage:          "derived_resource_pack_decryption",
				Source:         endpoint(source),
				Destination:    endpoint(destination),
				ParentSequence: &parent,
				Data:           data,
				Annotations: []string{
					"Derived plaintext from the original archive without modifying it",
					"The contents manifest contains per-file decryption keys and must be treated as sensitive",
				},
			},
			Raw:            manifest.Plaintext,
			MediaType:      "application/json",
			Representation: "minecraft_resource_pack_decrypted_contents_manifest",
		}); err != nil {
			return fmt.Errorf("record decrypted contents manifest %s for resource pack %s: %w", manifest.Path, pack.UUID(), err)
		}
	}

	archiveInfo, err := temporary.Stat()
	if err != nil {
		return fmt.Errorf("inspect decrypted resource pack %s: %w", pack.UUID(), err)
	}
	data, err := json.Marshal(resourcePackDerivedData{
		UUID:    pack.UUID().String(),
		Version: pack.Version(),
		Report:  result.Report,
	})
	if err != nil {
		return fmt.Errorf("encode decrypted archive metadata for resource pack %s: %w", pack.UUID(), err)
	}
	if _, err := recorder.RecordReader(ctx, capture.Event{
		SessionID:      sessionID,
		ConnectionID:   connectionID,
		Hop:            hop,
		Kind:           "resource_pack.decrypted_archive",
		Severity:       capture.SeverityInfo,
		Direction:      capture.DirectionServerToClient,
		Channel:        "upstream",
		Stage:          "derived_resource_pack_decryption",
		Source:         endpoint(source),
		Destination:    endpoint(destination),
		ParentSequence: &parent,
		Data:           data,
		Annotations: []string{
			"Derived AES-256-CFB8 plaintext archive; the original archive remains the authoritative captured evidence",
			"AES-CFB8 does not authenticate plaintext, so successful decryption is not an integrity proof",
		},
	}, temporary, archiveInfo.Size(), "application/zip", "minecraft_resource_pack_decrypted_archive"); err != nil {
		return fmt.Errorf("record decrypted archive for resource pack %s: %w", pack.UUID(), err)
	}
	return nil
}

func recordResourcePackDecryptFailure(ctx context.Context, recorder *capture.Recorder, sessionID, connectionID string, hop int, source, destination net.Addr, pack *resource.Pack, parentSequence uint64, decryptErr error) error {
	limitation := fmt.Sprintf("Encrypted resource pack %s could not be decrypted; its original archive was retained", pack.UUID())
	if err := recorder.AddLimitation(limitation); err != nil {
		return fmt.Errorf("record resource pack decryption limitation: %w", err)
	}
	data, err := json.Marshal(resourcePackDerivedData{UUID: pack.UUID().String(), Version: pack.Version()})
	if err != nil {
		return fmt.Errorf("encode resource pack decryption failure metadata: %w", err)
	}
	parent := parentSequence
	if _, err := recorder.Record(ctx, capture.Record{Event: capture.Event{
		SessionID:      sessionID,
		ConnectionID:   connectionID,
		Hop:            hop,
		Kind:           "resource_pack.decrypt_error",
		Severity:       capture.SeverityWarn,
		Direction:      capture.DirectionServerToClient,
		Channel:        "upstream",
		Stage:          "derived_resource_pack_decryption",
		Source:         endpoint(source),
		Destination:    endpoint(destination),
		ParentSequence: &parent,
		Data:           data,
		Error: &capture.ErrorInfo{
			Operation: "decrypt_resource_pack",
			Message:   decryptErr.Error(),
			Type:      fmt.Sprintf("%T", decryptErr),
		},
		Annotations: []string{
			"Decryption is derived and optional; failure does not discard or modify the original archive",
			"No undocumented encryption variant or compatibility fallback was attempted",
		},
	}}); err != nil {
		return fmt.Errorf("record resource pack decryption failure: %w", err)
	}
	return nil
}
