package proxy

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
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

func TestURLResourcePackCacheRejectsIncompleteHTTPBody(t *testing.T) {
	archive := testURLResourcePackArchive(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Error("test server does not support hijacking")
			return
		}
		conn, rw, err := hijacker.Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = conn.Close() }()
		// Deliver a complete ZIP but omit the final HTTP chunk. The archive
		// parser must not turn the failed HTTP body into a successful pack.
		_, _ = fmt.Fprintf(rw, "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n%x\r\n", len(archive))
		_, _ = rw.Write(archive)
		_, _ = rw.WriteString("\r\n")
		if err := rw.Flush(); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
		UUID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Version:     "1.0.0",
		Size:        uint64(len(archive)),
		DownloadURL: server.URL,
	}}}
	cache := newURLResourcePackCache(context.Background())
	cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, encodeResourcePacksInfo(t, info))
	pack, err := cache.Load(context.Background(), minecraft.ResourcePackCacheKey{
		UUID: info.TexturePacks[0].UUID, Version: "1.0.0", Size: uint64(len(archive)),
	})
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("cache error = %v, want the HTTP body read failure", err)
	}
	if pack != nil || len(cache.Packs()) != 0 {
		t.Fatal("accepted an archive from an incomplete HTTP body")
	}
}

func TestURLResourcePackCacheBoundsResponseBodyToAdvertisedSize(t *testing.T) {
	archive := testURLResourcePackArchive(t)
	extra := bytes.Repeat([]byte{'x'}, 1024*1024)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "test server does not support flushing", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(archive); err != nil {
			return
		}
		flusher.Flush()
		_, _ = w.Write(extra)
	}))
	defer server.Close()

	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
		UUID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Version:     "1.0.0",
		Size:        uint64(len(archive)),
		DownloadURL: server.URL,
	}}}
	cache := newURLResourcePackCache(context.Background())
	cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, encodeResourcePacksInfo(t, info))
	key := minecraft.ResourcePackCacheKey{UUID: info.TexturePacks[0].UUID, Version: "1.0.0", Size: uint64(len(archive))}
	pack, err := cache.Load(context.Background(), key)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized response error = %v, want an exceeds-limit error", err)
	}
	if pack != nil {
		t.Fatal("oversized response returned a resource pack")
	}
}

func TestURLResourcePackCacheCancelsInFlightDownload(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()

	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
		UUID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Version:     "1.0.0",
		Size:        1,
		DownloadURL: server.URL,
	}}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache := newURLResourcePackCache(ctx)
	done := make(chan struct{})
	go func() {
		cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, encodeResourcePacksInfo(t, info))
		close(done)
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("resource-pack request did not start")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("resource-pack download did not stop after cancellation")
	}
	key := minecraft.ResourcePackCacheKey{UUID: info.TexturePacks[0].UUID, Version: "1.0.0", Size: 1}
	pack, err := cache.Load(context.Background(), key)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("canceled response error = %v, want context canceled", err)
	}
	if pack != nil {
		t.Fatal("canceled response returned a resource pack")
	}
}

func TestURLResourcePackCacheDownloadsPacksConcurrentlyInOfferOrder(t *testing.T) {
	firstUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	secondUUID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	archives := map[string][]byte{
		"/first.zip":  testURLResourcePackArchiveWith(t, firstUUID, "1.0.0"),
		"/second.zip": testURLResourcePackArchiveWith(t, secondUUID, "2.0.0"),
	}
	started := make(chan string, len(archives))
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		archive, ok := archives[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		started <- r.URL.Path
		<-release
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{
		{UUID: firstUUID, Version: "1.0.0", Size: uint64(len(archives["/first.zip"])), DownloadURL: server.URL + "/first.zip"},
		{UUID: secondUUID, Version: "2.0.0", Size: uint64(len(archives["/second.zip"])), DownloadURL: server.URL + "/second.zip"},
	}}
	cache := newURLResourcePackCache(context.Background())
	done := make(chan struct{})
	go func() {
		cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, encodeResourcePacksInfo(t, info))
		close(done)
	}()

	for range archives {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("resource-pack downloads did not start concurrently")
		}
	}
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("resource-pack downloads did not complete")
	}

	packs := cache.Packs()
	if len(packs) != 2 {
		t.Fatalf("prefetched packs = %d, want 2", len(packs))
	}
	if packs[0].UUID() != firstUUID || packs[1].UUID() != secondUUID {
		t.Fatalf("prefetched pack order = %s, %s, want %s, %s", packs[0].UUID(), packs[1].UUID(), firstUUID, secondUUID)
	}
}

func TestURLResourcePackCacheMergePreservesMixedOfferOrder(t *testing.T) {
	firstUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	middleUUID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	lastUUID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	archives := map[string][]byte{
		"/first.zip":  testURLResourcePackArchiveWith(t, firstUUID, "1.0.0"),
		"/middle.zip": testURLResourcePackArchiveWith(t, middleUUID, "2.0.0"),
		"/last.zip":   testURLResourcePackArchiveWith(t, lastUUID, "3.0.0"),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		archive, ok := archives[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{
		{UUID: firstUUID, Version: "1.0.0", Size: uint64(len(archives["/first.zip"])), DownloadURL: server.URL + "/first.zip"},
		{UUID: middleUUID, Version: "2.0.0", Size: uint64(len(archives["/middle.zip"]))},
		{UUID: lastUUID, Version: "3.0.0", Size: uint64(len(archives["/last.zip"])), DownloadURL: server.URL + "/last.zip"},
	}}
	cache := newURLResourcePackCache(context.Background())
	cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, encodeResourcePacksInfo(t, info))
	if !cache.WaitForOffer(context.Background()) {
		t.Fatal("resource-pack offer was not observed")
	}
	middlePack, err := resource.Read(bytes.NewReader(archives["/middle.zip"]))
	if err != nil {
		t.Fatal(err)
	}
	merged := cache.Merge([]*resource.Pack{middlePack})
	if len(merged) != 3 {
		t.Fatalf("merged packs = %d, want 3", len(merged))
	}
	for index, want := range []uuid.UUID{firstUUID, middleUUID, lastUUID} {
		if merged[index].UUID() != want {
			t.Fatalf("merged pack %d = %s, want %s", index, merged[index].UUID(), want)
		}
	}
}

func TestURLResourcePackCacheMergePreservesChunkOnlyOfferOrder(t *testing.T) {
	firstUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	secondUUID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	firstArchive := testURLResourcePackArchiveWith(t, firstUUID, "1.0.0")
	secondArchive := testURLResourcePackArchiveWith(t, secondUUID, "2.0.0")
	first, err := resource.Read(bytes.NewReader(firstArchive))
	if err != nil {
		t.Fatal(err)
	}
	second, err := resource.Read(bytes.NewReader(secondArchive))
	if err != nil {
		t.Fatal(err)
	}
	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{
		{UUID: firstUUID, Version: "1.0.0", Size: uint64(len(firstArchive))},
		{UUID: secondUUID, Version: "2.0.0", Size: uint64(len(secondArchive))},
	}}
	cache := newURLResourcePackCache(context.Background())
	cache.Observe(packet.Header{PacketID: packet.IDResourcePacksInfo}, encodeResourcePacksInfo(t, info))
	for _, pack := range []*resource.Pack{second, first} {
		if err := cache.Store(context.Background(), resourcePackKey(pack), pack); err != nil {
			t.Fatal(err)
		}
	}
	merged := cache.Merge([]*resource.Pack{second, first})
	if len(merged) != 2 || merged[0] != first || merged[1] != second {
		t.Fatalf("merged packs = %#v, want two original packs in offer order", merged)
	}
}

func TestDecodeResourcePacksInfoAcceptsFinalEmptyDownloadURL(t *testing.T) {
	info := &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
		UUID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Version: "1.0.0", Size: 123,
	}}}
	payload := encodeResourcePacksInfo(t, info)
	original := bytes.Clone(payload)
	decoded, err := decodeResourcePacksInfo(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.TexturePacks) != 1 || decoded.TexturePacks[0] != info.TexturePacks[0] {
		t.Fatalf("decoded offer = %#v, want original chunk-delivered offer", decoded)
	}
	if !bytes.Equal(payload, original) {
		t.Fatal("decoder changed the observed packet bytes")
	}
}

func TestURLResourcePackCacheStoreClearsPreviousPrefetchError(t *testing.T) {
	packUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	archive := testURLResourcePackArchive(t)
	pack, err := resource.Read(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	key := minecraft.ResourcePackCacheKey{UUID: packUUID, Version: "1.0.0", Size: uint64(len(archive))}
	if !key.Matches(pack) {
		t.Fatalf("test pack does not match key: %#v", key)
	}

	cache := newURLResourcePackCache(context.Background())
	cache.mu.Lock()
	cache.errors[key] = errors.New("synthetic URL prefetch failure")
	cache.mu.Unlock()

	if err := cache.Store(context.Background(), key, pack); err != nil {
		t.Fatalf("store fallback pack: %v", err)
	}
	got, err := cache.Load(context.Background(), key)
	if err != nil {
		t.Fatalf("load stored fallback pack: %v", err)
	}
	if got == nil || !key.Matches(got) {
		t.Fatalf("loaded fallback pack = %#v, want a matching pack", got)
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

func TestDecodeResourcePacksInfoRejectsTrailingBytes(t *testing.T) {
	payload := append(encodeResourcePacksInfo(t, &packet.ResourcePacksInfo{}), 0x00)
	if info, err := decodeResourcePacksInfo(payload); err == nil || info != nil {
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
	return testURLResourcePackArchiveWith(t, uuid.MustParse("11111111-1111-1111-1111-111111111111"), "1.0.0")
}

func testURLResourcePackArchiveWith(t *testing.T, packUUID uuid.UUID, version string) []byte {
	t.Helper()
	versionParts := make([]int, 3)
	for index, part := range strings.Split(version, ".") {
		if index == len(versionParts) {
			break
		}
		versionParts[index], _ = strconv.Atoi(part)
	}
	manifest := map[string]any{
		"format_version": 2,
		"header": map[string]any{
			"description":        "URL test pack",
			"name":               "URL Test Pack",
			"uuid":               packUUID.String(),
			"version":            versionParts,
			"min_engine_version": []int{1, 20, 0},
		},
		"modules": []any{map[string]any{
			"description": "URL test module",
			"type":        "resources",
			"uuid":        "33333333-3333-3333-3333-333333333333",
			"version":     versionParts,
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
