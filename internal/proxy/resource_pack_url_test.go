package proxy

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestURLResourcePackCacheLoadsAdvertisedPack(t *testing.T) {
	const contentKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"
	archive := testURLResourcePackArchive(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	packUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
		UUID:        packUUID,
		Version:     "1.0.0",
		Size:        uint64(len(archive)),
		ContentKey:  contentKey,
		DownloadURL: server.URL,
	}}}
	payload := encodeResourcePacksInfo(t, info)
	cache := newURLResourcePackCache(context.Background())
	cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, payload)
	if !cache.WaitForOffer(context.Background()) {
		t.Fatal("resource-pack offer was not observed")
	}

	key := minecraft.ResourcePackCacheKey{UUID: packUUID, Version: "1.0.0", Size: uint64(len(archive))}
	pack, err := cache.Load(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	if pack == nil {
		t.Fatal("cache returned no resource pack")
	}
	if pack.DownloadURL() != server.URL {
		t.Fatalf("download URL = %q, want %q", pack.DownloadURL(), server.URL)
	}
	if pack.ContentKey() != contentKey {
		t.Fatalf("content key = %q, want %q", pack.ContentKey(), contentKey)
	}
	if got := cache.Packs(); len(got) != 1 || got[0] != pack {
		t.Fatalf("prefetched packs = %#v, want the cached pack", got)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("HTTP requests = %d, want 1", got)
	}

	// Retransmitted ResourcePacksInfo packets must not trigger a second download.
	cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, payload)
	if got := requests.Load(); got != 1 {
		t.Fatalf("HTTP requests after duplicate offer = %d, want 1", got)
	}
}

func TestURLResourcePackCacheReportsAdvertisedPackMismatch(t *testing.T) {
	archive := testURLResourcePackArchive(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
		UUID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Version:     "1.0.0",
		Size:        uint64(len(archive)),
		DownloadURL: server.URL,
	}}}
	cache := newURLResourcePackCache(context.Background())
	cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, encodeResourcePacksInfo(t, info))

	key := minecraft.ResourcePackCacheKey{UUID: info.TexturePacks[0].UUID, Version: "1.0.0", Size: uint64(len(archive))}
	pack, err := cache.Load(context.Background(), key)
	if err == nil {
		t.Fatal("cache mismatch returned nil error")
	}
	if pack != nil {
		t.Fatal("cache mismatch returned a pack")
	}
}

func TestResourcePacksForDownstreamClearsURL(t *testing.T) {
	archive := testURLResourcePackArchive(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer server.Close()
	packUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
		UUID:        packUUID,
		Version:     "1.0.0",
		Size:        uint64(len(archive)),
		DownloadURL: server.URL + "/pack.zip",
		ContentKey:  "synthetic-content-key",
	}}}

	cache := newURLResourcePackCache(context.Background())
	cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, encodeResourcePacksInfo(t, info))
	packs, err := resourcePacksForDownstream(cache.Packs())
	if err != nil {
		t.Fatalf("resourcePacksForDownstream() error = %v", err)
	}
	if len(packs) != 1 {
		t.Fatalf("downstream packs = %d, want 1", len(packs))
	}
	if got := packs[0].DownloadURL(); got != "" {
		t.Fatalf("downstream DownloadURL = %q, want empty", got)
	}
	if got := packs[0].ContentKey(); got != info.TexturePacks[0].ContentKey {
		t.Fatalf("downstream ContentKey = %q, want %q", got, info.TexturePacks[0].ContentKey)
	}
	key := minecraft.ResourcePackCacheKey{UUID: packUUID, Version: "1.0.0", Size: uint64(len(archive))}
	if !key.Matches(packs[0]) {
		t.Fatalf("downstream pack no longer matches advertised key")
	}
}

func TestDecodeResourcePacksInfoMalformedPayloadDoesNotPanic(t *testing.T) {
	if info, err := decodeResourcePacksInfo([]byte{0xff}); err == nil || info != nil {
		t.Fatalf("decode result = %#v, err = %v", info, err)
	}
}

func encodeResourcePacksInfo(t *testing.T, info *packet.ResourcePacksInfo) []byte {
	t.Helper()
	var payload bytes.Buffer
	info.Marshal(protocol.NewWriter(&payload, 0))
	return payload.Bytes()
}

func testURLResourcePackArchive(t *testing.T) []byte {
	t.Helper()
	manifest := map[string]any{
		"format_version": 2,
		"header": map[string]any{
			"description":        "URL test pack",
			"name":               "URL Test Pack",
			"uuid":               "11111111-1111-1111-1111-111111111111",
			"version":            []int{1, 0, 0},
			"min_engine_version": []int{1, 20, 0},
		},
		"modules": []any{map[string]any{
			"description": "URL test module",
			"type":        "resources",
			"uuid":        "33333333-3333-3333-3333-333333333333",
			"version":     []int{1, 0, 0},
		}},
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	manifestFile, err := writer.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manifestFile.Write(manifestJSON); err != nil {
		t.Fatal(err)
	}
	readme, err := writer.Create("README.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(readme, fmt.Sprintf("generated %d", len(manifestJSON))); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return archive.Bytes()
}
