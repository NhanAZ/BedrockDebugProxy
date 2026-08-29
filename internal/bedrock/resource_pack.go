package bedrock

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

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

func RecordResourcePacks(ctx context.Context, recorder *capture.Recorder, sessionID, connectionID string, hop int, source, destination net.Addr, packs []*resource.Pack) error {
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
	}
	return nil
}
