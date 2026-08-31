package artifacts

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/NhanAZ/BedrockDebugProxy/internal/packetview"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (w *Live) packetSkins(event capture.Event) (err error) {
	// Reuse the pinned codec, with its allocation limits and panic boundary. The
	// packet hook stores the body without a header. No protocol layout is guessed.
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("decode skin packet: %v", recovered)
		}
	}()
	data, err := w.readSkinBlob(event.Blob)
	if err != nil {
		return err
	}
	var pk packet.Packet
	switch event.Packet.ID {
	case packet.IDPlayerSkin:
		pk = &packet.PlayerSkin{}
	case packet.IDPlayerList:
		pk = &packet.PlayerList{}
	default:
		return errors.New("unsupported skin packet")
	}
	buffer := bytes.NewBuffer(data)
	pk.Marshal(protocol.NewReader(buffer, 0, true))
	if buffer.Len() != 0 {
		return fmt.Errorf("skin packet has %d unread bytes", buffer.Len())
	}
	switch typed := pk.(type) {
	case *packet.PlayerSkin:
		return w.skin(event, typed.UUID.String(), "", typed.Skin)
	case *packet.PlayerList:
		for _, entry := range typed.Entries {
			if entry.ActionType == protocol.PlayerListActionAdd {
				if err := w.skin(event, entry.UUID.String(), entry.Username, entry.Skin); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (w *Live) skin(event capture.Event, owner, name string, skin protocol.Skin) error {
	paths := make(map[string]string)
	for _, item := range []struct {
		name          string
		data          []byte
		width, height uint32
	}{
		{"skin", skin.SkinData, skin.SkinImageWidth, skin.SkinImageHeight},
		{"cape", skin.CapeData, skin.CapeImageWidth, skin.CapeImageHeight},
	} {
		path, err := w.imageFile(item.data, int64(item.width), int64(item.height))
		if err != nil {
			return fmt.Errorf("%s image: %w", item.name, err)
		}
		if path != "" {
			paths[item.name] = path
		}
	}
	for i, animation := range skin.Animations {
		path, err := w.imageFile(animation.ImageData, int64(animation.ImageWidth), int64(animation.ImageHeight))
		if err != nil {
			return fmt.Errorf("animation %d: %w", i, err)
		}
		if path != "" {
			paths[fmt.Sprintf("animation_%d", i)] = path
		}
	}
	for _, item := range []struct {
		name string
		data []byte
	}{
		{"geometry", skin.SkinGeometry}, {"resource_patch", skin.SkinResourcePatch},
		{"animation_data", skin.AnimationData}, {"geometry_engine_version", skin.GeometryDataEngineVersion},
	} {
		path, err := w.skinDataFile(item.data)
		if err != nil {
			return err
		}
		if path != "" {
			paths[item.name] = path
		}
	}
	metadata, _, err := packetview.Encode(skin, packetview.Options{})
	if err != nil {
		return err
	}
	return w.skinIndex(event, owner, name, paths, metadata)
}

func (w *Live) loginSkin(event capture.Event) error {
	// Read only the skin fields. Other snapshot values may use packetview's
	// large-integer wrappers and are not a wire-format login.ClientData object.
	type loginSkinData struct {
		SkinID              string                `json:"SkinId"`
		CapeID              string                `json:"CapeId"`
		SkinImageWidth      int                   `json:"SkinImageWidth"`
		SkinImageHeight     int                   `json:"SkinImageHeight"`
		CapeImageWidth      int                   `json:"CapeImageWidth"`
		CapeImageHeight     int                   `json:"CapeImageHeight"`
		SkinData            string                `json:"SkinData"`
		CapeData            string                `json:"CapeData"`
		SkinResourcePatch   string                `json:"SkinResourcePatch"`
		SkinGeometry        string                `json:"SkinGeometryData"`
		SkinGeometryVersion string                `json:"SkinGeometryDataEngineVersion"`
		AnimatedImageData   []login.SkinAnimation `json:"AnimatedImageData"`
	}
	var view struct {
		Fields struct {
			Identity struct {
				Identity string `json:"identity"`
			} `json:"identity"`
			ClientData loginSkinData `json:"client_data"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(event.Data, &view); err != nil {
		return err
	}
	data := view.Fields.ClientData
	paths := make(map[string]string)
	for _, item := range []struct {
		name, encoded string
		width, height int
	}{
		{"skin", data.SkinData, data.SkinImageWidth, data.SkinImageHeight},
		{"cape", data.CapeData, data.CapeImageWidth, data.CapeImageHeight},
	} {
		pixels, err := decodeBase64(item.encoded)
		if err != nil {
			return fmt.Errorf("login %s: %w", item.name, err)
		}
		path, err := w.imageFile(pixels, int64(item.width), int64(item.height))
		if err != nil {
			return fmt.Errorf("login %s: %w", item.name, err)
		}
		if path != "" {
			paths[item.name] = path
		}
	}
	for i, animation := range data.AnimatedImageData {
		pixels, err := decodeBase64(animation.Image)
		if err != nil {
			return err
		}
		path, err := w.imageFile(pixels, int64(animation.ImageWidth), int64(animation.ImageHeight))
		if err != nil {
			return err
		}
		if path != "" {
			paths[fmt.Sprintf("animation_%d", i)] = path
		}
	}
	for _, item := range []struct{ name, encoded string }{
		{"geometry", data.SkinGeometry}, {"resource_patch", data.SkinResourcePatch},
		{"geometry_engine_version", data.SkinGeometryVersion},
	} {
		decoded, err := decodeBase64(item.encoded)
		if err != nil {
			return err
		}
		path, err := w.skinDataFile(decoded)
		if err != nil {
			return err
		}
		if path != "" {
			paths[item.name] = path
		}
	}
	// Do not copy the entire connection snapshot into the convenience index.
	// Other login properties and their exact representation remain at source_sequence.
	metadata, err := json.Marshal(map[string]any{"skin_id": data.SkinID, "cape_id": data.CapeID, "source": "login_client_data"})
	if err != nil {
		return err
	}
	return w.skinIndex(event, view.Fields.Identity.Identity, "", paths, metadata)
}

func decodeBase64(value string) ([]byte, error) {
	if base64.StdEncoding.DecodedLen(len(value)) > maxSkinPacketBytes {
		return nil, errors.New("base64 skin field exceeds the 16 MiB view limit")
	}
	return base64.StdEncoding.DecodeString(value)
}

func (w *Live) skinIndex(event capture.Event, owner, name string, paths map[string]string, metadata json.RawMessage) error {
	return w.appendJSON("skins/index.jsonl", map[string]any{
		"source_sequence": event.Sequence, "source_blob": event.Blob,
		"connection_id": event.ConnectionID, "channel": event.Channel, "direction": event.Direction,
		"time": event.Time, "player_uuid": owner, "player_name": name, "files": paths, "metadata": metadata,
	})
}

func (w *Live) imageFile(pixels []byte, width, height int64) (string, error) {
	if width == 0 && height == 0 && len(pixels) == 0 {
		return "", nil
	}
	if width <= 0 || height <= 0 || width > maxImagePixels || height > maxImagePixels || width*height > maxImagePixels {
		return "", errors.New("invalid dimensions or image exceeds the 1,048,576 pixel view limit")
	}
	if width*height*4 != int64(len(pixels)) {
		return "", errors.New("RGBA byte count does not match image dimensions")
	}
	hash := sha256.Sum256(pixels)
	path := fmt.Sprintf("skins/images/%dx%d-%x.png", width, height, hash)
	if err := w.publish(path, func(dst io.Writer) error {
		// Bedrock transmits RGBA channels, not premultiplied-alpha color values.
		// NRGBA preserves even RGB channels under fully transparent pixels.
		img := &image.NRGBA{Pix: pixels, Stride: int(width) * 4, Rect: image.Rect(0, 0, int(width), int(height))}
		encoder := png.Encoder{CompressionLevel: png.BestSpeed}
		return encoder.Encode(dst, img)
	}); err != nil {
		return "", err
	}
	return "artifacts/" + path, nil
}

func (w *Live) skinDataFile(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}
	hash := sha256.Sum256(data)
	ext := ".bin"
	if json.Valid(data) {
		ext = ".json"
	}
	path := "skins/data/" + hex.EncodeToString(hash[:]) + ext
	if err := w.publish(path, func(dst io.Writer) error { _, err := dst.Write(data); return err }); err != nil {
		return "", err
	}
	return "artifacts/" + path, nil
}
