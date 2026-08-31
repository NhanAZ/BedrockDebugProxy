package artifacts

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capturearchive"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func fixture(t *testing.T, notify func()) (string, *capture.Recorder, *Live) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "capture")
	r, err := capture.New(root, capture.Options{CaptureID: "artifact-test"})
	if err != nil {
		t.Fatal(err)
	}
	w, err := Start(root, protocol.CurrentProtocol, notify)
	if err != nil {
		_ = r.Close("failed", err)
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close("closed", nil); _, _ = w.Finish() })
	return root, r, w
}

func record(t *testing.T, r *capture.Recorder, event capture.Event, raw []byte) capture.Event {
	t.Helper()
	result, err := r.Record(context.Background(), capture.Record{Event: event, Raw: raw})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func rawPacket(t *testing.T, r *capture.Recorder, pk packet.Packet) capture.Event {
	t.Helper()
	var body bytes.Buffer
	pk.Marshal(protocol.NewWriter(&body, 0))
	return record(t, r, capture.Event{Kind: "packet.raw", Channel: "upstream", Direction: capture.DirectionServerToClient, Packet: &capture.PacketInfo{ID: pk.ID()}}, body.Bytes())
}

func readJSONLines(t *testing.T, root, path string) []map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]json.RawMessage
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte{'\n'}) {
		var row map[string]json.RawMessage
		if err := json.Unmarshal(line, &row); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, row)
	}
	return rows
}

func TestLiveWritesImagesAndOrderedViewsBeforeShutdown(t *testing.T) {
	root, r, w := fixture(t, nil)
	skin := protocol.Skin{SkinID: "synthetic", SkinImageWidth: 2, SkinImageHeight: 1, SkinData: []byte{1, 2, 3, 0, 4, 5, 6, 127}, CapeImageWidth: 1, CapeImageHeight: 1, CapeData: []byte{7, 8, 9, 255}, SkinGeometry: []byte(`{"geometry":"synthetic"}`)}
	rawPacket(t, r, &packet.PlayerList{Entries: []protocol.PlayerListEntry{
		{ActionType: protocol.PlayerListActionAdd, UUID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Skin: skin},
		{ActionType: protocol.PlayerListActionAdd, UUID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Skin: skin},
		{ActionType: protocol.PlayerListActionRemove, UUID: uuid.MustParse("33333333-3333-3333-3333-333333333333")},
	}})
	rawPacket(t, r, &packet.PlayerSkin{Skin: skin})
	last := record(t, r, capture.Event{Kind: "packet.decoded", Packet: &capture.PacketInfo{ID: packet.IDLevelChunk, Name: "LevelChunk"}, Data: json.RawMessage(`{"fields":{"X":1}}`)}, nil)
	// A long event exercises split reads across the 256 KiB input buffer.
	longData, err := json.Marshal(map[string]string{"text": strings.Repeat("x", 300<<10)})
	if err != nil {
		t.Fatal(err)
	}
	record(t, r, capture.Event{Kind: "packet.decoded", Packet: &capture.PacketInfo{Name: "UpdateBlock"}, Data: longData}, nil)
	deadline := time.Now().Add(5 * time.Second)
	for {
		data, readErr := os.ReadFile(filepath.Join(root, "artifacts", "observations", "world.jsonl"))
		if readErr == nil && bytes.Contains(data, []byte("LevelChunk")) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("artifact views were not published during the open session")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := r.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	status, err := w.Finish()
	if err != nil || status.State != "complete" || status.LastSequence != 4 {
		t.Fatalf("status=%+v error=%v", status, err)
	}
	if _, err := w.Finish(); err != nil {
		t.Fatal(err)
	}
	rows := readJSONLines(t, root, "artifacts/skins/index.jsonl")
	if len(rows) != 3 {
		t.Fatalf("skin observations = %d, want 3 (remove entries have no image)", len(rows))
	}
	var paths map[string]string
	if err := json.Unmarshal(rows[0]["files"], &paths); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(filepath.Join(root, paths["skin"]))
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(file)
	_ = file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got := color.NRGBAModel.Convert(img.At(0, 0)); got != (color.NRGBA{R: 1, G: 2, B: 3, A: 0}) {
		t.Fatalf("transparent pixel changed: %v", got)
	}
	if got := color.NRGBAModel.Convert(img.At(1, 0)); got != (color.NRGBA{R: 4, G: 5, B: 6, A: 127}) {
		t.Fatalf("partial-alpha pixel changed: %v", got)
	}
	images, err := os.ReadDir(filepath.Join(root, "artifacts/skins/images"))
	if err != nil || len(images) != 2 {
		t.Fatalf("duplicate images: %d %v", len(images), err)
	}
	world := readJSONLines(t, root, "artifacts/observations/world.jsonl")
	encoded, err := json.Marshal(last)
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &expected); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(world[0]["sequence"], expected["sequence"]) || !bytes.Equal(world[0]["data"], expected["data"]) {
		t.Fatal("view did not retain source sequence and fields")
	}
	verification, err := capture.Verify(root)
	if err != nil || len(verification.Issues) != 0 {
		t.Fatalf("canonical capture changed: %v %v", verification.Issues, err)
	}
	archive, err := capturearchive.Export(root, filepath.Join(t.TempDir(), "capture.bdpcap"))
	if err != nil {
		t.Fatal(err)
	}
	// Two source packet blobs plus manifest.json and events.jsonl only.
	if verification.Blobs != 2 || archive.Entries != 4 {
		t.Fatal("derived folders were unexpectedly added to canonical export")
	}
}

func TestLoginSkinAndPackCopiesKeepProvenanceWithoutKeys(t *testing.T) {
	root, r, w := fixture(t, nil)
	loginData, err := json.Marshal(map[string]any{"fields": map[string]any{"client_data": map[string]any{
		"SkinID": "login-skin", "SkinImageWidth": 1, "SkinImageHeight": 1,
		"SkinData":         base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4}),
		"SkinGeometryData": base64.StdEncoding.EncodeToString([]byte(`{"test":true}`)),
		"DeviceId":         "must-not-be-copied-to-index",
		"ClientRandomId":   map[string]any{"$integer": "9223372036854775807", "$type": "int64"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	record(t, r, capture.Event{Kind: "session.connection_metadata", Data: loginData}, nil)
	pack := record(t, r, capture.Event{Kind: "resource_pack.archive", Data: json.RawMessage(`{"name":"../CON:<unsafe>","uuid":"synthetic-pack","encrypted":true,"content_key":"secret-test-key"}`)}, []byte("original synthetic archive"))
	parent := pack.Sequence
	record(t, r, capture.Event{Kind: "resource_pack.decrypted_archive", ParentSequence: &parent, Data: json.RawMessage(`{"uuid":"synthetic-pack"}`)}, []byte("derived synthetic archive"))
	if err := r.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	status, err := w.Finish()
	if err != nil || status.Errors != 0 {
		t.Fatalf("status=%+v error=%v", status, err)
	}
	rows := readJSONLines(t, root, "artifacts/packs/index.jsonl")
	if len(rows) != 2 {
		t.Fatalf("packs = %d", len(rows))
	}
	for i, row := range rows {
		if bytes.Contains(row["content_key"], []byte("secret")) {
			t.Fatal("copied pack key")
		}
		var path string
		if err := json.Unmarshal(row["path"], &path); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		if (i == 0 && string(data) != "original synthetic archive") || (i == 1 && string(data) != "derived synthetic archive") {
			t.Fatalf("incorrect pack copy %q", data)
		}
		if i == 0 {
			if strings.Contains(path, "..") || strings.ContainsAny(path, "<:>") {
				t.Fatalf("unsafe name %s", path)
			}
			if err := os.WriteFile(filepath.Join(root, path), []byte("edited copy"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	index, err := os.ReadFile(filepath.Join(root, "artifacts/skins/index.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(index, []byte("must-not-be-copied-to-index")) {
		t.Fatal("copied unrelated login identifier")
	}
	verification, err := capture.Verify(root)
	if err != nil || len(verification.Issues) != 0 {
		t.Fatal("editing a pack copy changed the capture", err)
	}
}

func TestMalformedSkinDoesNotStopFollowingOrAlterRawCapture(t *testing.T) {
	var warnings atomic.Int32
	root, r, w := fixture(t, func() { warnings.Add(1) })
	for range 2 {
		record(t, r, capture.Event{Kind: "packet.raw", Channel: "upstream", Direction: capture.DirectionServerToClient, Packet: &capture.PacketInfo{ID: packet.IDPlayerList}}, []byte{255, 255, 255, 255, 15})
	}
	record(t, r, capture.Event{Kind: "packet.decoded", Packet: &capture.PacketInfo{Name: "InventorySlot"}}, nil)
	if err := r.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	status, err := w.Finish()
	if err == nil || status.State != "complete_with_errors" || status.Errors != 2 || status.LastSequence != 3 || warnings.Load() != 1 {
		t.Fatalf("status=%+v error=%v warnings=%d", status, err, warnings.Load())
	}
	if len(readJSONLines(t, root, "artifacts/errors.jsonl")) != 2 {
		t.Fatal("missing error evidence")
	}
	if len(readJSONLines(t, root, "artifacts/observations/inventory.jsonl")) != 1 {
		t.Fatal("later event was not processed")
	}
	verification, err := capture.Verify(root)
	if err != nil || len(verification.Issues) != 0 {
		t.Fatal("malformed source evidence changed", err)
	}
}

func TestImageBoundsAndPacketGrouping(t *testing.T) {
	_, r, w := fixture(t, nil)
	if err := r.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Finish(); err != nil {
		t.Fatal(err)
	}
	for _, size := range [][2]int64{{-1, 1}, {1 << 32, 1 << 32}, {1, 0}, {1025, 1024}, {1, 1}} {
		if _, err := w.imageFile(nil, size[0], size[1]); err == nil {
			t.Fatalf("accepted invalid image size %v", size)
		}
	}
	for name, want := range map[string]string{"PlayerList": "skins", "PlayerSkin": "skins", "AddPlayer": "entities", "AddActor": "entities", "LevelChunk": "world", "SubChunkRequest": "world", "ClientCacheMissResponse": "world", "UpdateBlock": "blocks", "UpdateSubChunkBlocks": "blocks", "BlockActorData": "blocks", "InventoryContent": "inventory", "InventorySlot": "inventory", "MobEquipment": "inventory", "ItemStackResponse": "inventory", "Unknown": ""} {
		if got := packetGroup(name); got != want {
			t.Errorf("%s: got %s want %s", name, got, want)
		}
	}
}

func TestStartRefusesExistingArtifactsAndClosedCaptures(t *testing.T) {
	root, r, w := fixture(t, nil)
	if _, err := Start(root, protocol.CurrentProtocol, nil); err == nil {
		t.Fatal("overwrote existing artifacts")
	}
	if err := r.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Finish(); err != nil {
		t.Fatal(err)
	}
	if _, err := Start(root, protocol.CurrentProtocol, nil); err == nil {
		t.Fatal("accepted a historical closed capture")
	}
}

func TestArtifactWriteFailureIsVisibleAndDoesNotDamageCapture(t *testing.T) {
	var warnings atomic.Int32
	root, r, w := fixture(t, func() { warnings.Add(1) })
	if err := os.MkdirAll(filepath.Join(root, "artifacts/observations/world.jsonl"), 0o700); err != nil {
		t.Fatal(err)
	}
	record(t, r, capture.Event{Kind: "packet.decoded", Packet: &capture.PacketInfo{Name: "LevelChunk"}}, nil)
	if err := r.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	status, err := w.Finish()
	if err == nil || status.State != "failed" || warnings.Load() != 1 {
		t.Fatalf("status=%+v error=%v", status, err)
	}
	data, err := os.ReadFile(filepath.Join(root, "artifacts/status.json"))
	if err != nil || !bytes.Contains(data, []byte(`"state": "failed"`)) {
		t.Fatalf("missing persistent failure status: %s %v", data, err)
	}
	verification, err := capture.Verify(root)
	if err != nil || len(verification.Issues) != 0 {
		t.Fatal("view failure damaged original capture", err)
	}
}

func TestBlobBoundariesAndChecksums(t *testing.T) {
	root, r, w := fixture(t, nil)
	hash := sha256.Sum256([]byte("expected"))
	digest := hex.EncodeToString(hash[:])
	ref := capture.BlobRef{SHA256: digest, Size: 8, Path: "blobs/sha256/" + digest[:2] + "/" + digest + ".bin"}
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, ref.Path)), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ref.Path), []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := w.readSkinBlob(&ref); err == nil {
		t.Fatal("accepted a checksum mismatch")
	}
	ref.Path = "../outside.bin"
	if _, err := w.openBlob(&ref); err == nil {
		t.Fatal("accepted a source path escape")
	}
	ref.Size = maxSkinPacketBytes + 1
	if _, err := w.readSkinBlob(&ref); err == nil {
		t.Fatal("accepted an oversized skin blob")
	}
	if err := r.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Finish(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkRepeatedSkinImage(b *testing.B) {
	root, err := os.OpenRoot(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	w := &Live{output: root}
	pixels := bytes.Repeat([]byte{1, 2, 3, 255}, 128*128)
	if _, err := w.imageFile(pixels, 128, 128); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(pixels)))
	b.ResetTimer()
	for b.Loop() {
		if _, err := w.imageFile(pixels, 128, 128); err != nil {
			b.Fatal(err)
		}
	}
}

func FuzzSafeName(f *testing.F) {
	for _, name := range []string{"pack", "../CON:<name>", "", "packs/../../outside", "skin\x00name", "texture\u202e.png"} {
		f.Add(name)
	}
	f.Fuzz(func(t *testing.T, name string) {
		result := safeName(name)
		if len(result) == 0 || len(result) > 40 {
			t.Fatalf("invalid name length %q", result)
		}
		for _, c := range result {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
				t.Fatalf("unsafe character in %q", result)
			}
		}
	})
}
