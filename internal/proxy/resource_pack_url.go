package proxy

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

// maxURLResourcePackBytes is a local safety bound for an archive fetched from a
// server-advertised URL. The advertised compressed size is also checked after
// parsing, so a response cannot be accepted when it does not match the packet.
const maxURLResourcePackBytes = 512 << 20

const urlResourcePackOfferWait = 5 * time.Second

const urlResourcePackDownloadWait = 60 * time.Second

// urlResourcePackCache adapts the URL form of ResourcePacksInfo to
// gophertunnel's ResourcePackCache interface. gophertunnel invokes PacketFunc
// before decoding a packet, which lets this cache observe the advertised URL
// and populate a cache entry before handleResourcePacksInfo asks for it.
type urlResourcePackCache struct {
	ctx context.Context

	mu        sync.Mutex
	packs     map[minecraft.ResourcePackCacheKey]*resource.Pack
	packOrder []minecraft.ResourcePackCacheKey
	errors    map[minecraft.ResourcePackCacheKey]error
	observed  map[minecraft.ResourcePackCacheKey]struct{}
	offered   chan struct{}
	ready     chan struct{}
	offerOnce sync.Once
	readyOnce sync.Once
	urlCount  int
}

func newURLResourcePackCache(ctx context.Context) *urlResourcePackCache {
	if ctx == nil {
		ctx = context.Background()
	}
	return &urlResourcePackCache{
		ctx:       ctx,
		packs:     make(map[minecraft.ResourcePackCacheKey]*resource.Pack),
		packOrder: make([]minecraft.ResourcePackCacheKey, 0),
		errors:    make(map[minecraft.ResourcePackCacheKey]error),
		observed:  make(map[minecraft.ResourcePackCacheKey]struct{}),
		offered:   make(chan struct{}),
		ready:     make(chan struct{}),
	}
}

// Observe looks for ResourcePacksInfo packets and downloads only packs that
// carry a URL. The original packet is still passed to the normal observer and
// remains the canonical evidence for the advertised URL.
func (cache *urlResourcePackCache) Observe(header packet.Header, payload []byte) {
	if header.PacketID != packet.IDResourcePacksInfo {
		return
	}
	info, err := decodeResourcePacksInfo(payload)
	if err != nil {
		// The normal gophertunnel decoder will retain the malformed packet as an
		// observable failure. There is no cache key to attach this error to.
		return
	}
	urlCount := 0
	for _, advertised := range info.TexturePacks {
		if advertised.DownloadURL != "" {
			urlCount++
		}
	}
	cache.offerOnce.Do(func() {
		cache.mu.Lock()
		cache.urlCount = urlCount
		cache.mu.Unlock()
		close(cache.offered)
	})
	defer cache.readyOnce.Do(func() { close(cache.ready) })
	for _, advertised := range info.TexturePacks {
		if advertised.DownloadURL == "" {
			continue
		}
		key := minecraft.ResourcePackCacheKey{
			UUID:    advertised.UUID,
			Version: advertised.Version,
			Size:    advertised.Size,
		}
		cache.mu.Lock()
		_, alreadyObserved := cache.observed[key]
		cache.observed[key] = struct{}{}
		cache.mu.Unlock()
		if alreadyObserved {
			continue
		}

		pack, err := cache.download(advertised)
		cache.mu.Lock()
		if err != nil {
			cache.errors[key] = err
		} else {
			cache.packs[key] = pack
			cache.packOrder = append(cache.packOrder, key)
		}
		cache.mu.Unlock()
	}
}

func (cache *urlResourcePackCache) download(advertised protocol.TexturePackInfo) (*resource.Pack, error) {
	parsed, err := url.Parse(advertised.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("parse resource pack URL %q: %w", advertised.DownloadURL, err)
	}
	scheme := strings.ToLower(parsed.Scheme)
	if (scheme != "http" && scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("resource pack URL must use http or https with a host: %q", advertised.DownloadURL)
	}
	if advertised.Size > maxURLResourcePackBytes {
		return nil, fmt.Errorf("resource pack URL size %d exceeds local limit %d", advertised.Size, maxURLResourcePackBytes)
	}
	select {
	case <-cache.ctx.Done():
		return nil, cache.ctx.Err()
	default:
	}

	pack, err := resource.ReadURL(advertised.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("download resource pack from %q: %w", advertised.DownloadURL, err)
	}
	key := minecraft.ResourcePackCacheKey{UUID: advertised.UUID, Version: advertised.Version, Size: advertised.Size}
	if !key.Matches(pack) {
		return nil, fmt.Errorf("resource pack from %q does not match advertised UUID, version, or size", advertised.DownloadURL)
	}
	if advertised.ContentKey != "" {
		pack = pack.WithContentKey(advertised.ContentKey)
	}
	return pack, nil
}

// Load implements minecraft.ResourcePackCache. URL failures are returned to
// gophertunnel, which records the warning and falls back to its normal chunk
// request path. This keeps failures visible without hiding the raw packet.
func (cache *urlResourcePackCache) Load(ctx context.Context, key minecraft.ResourcePackCacheKey) (*resource.Pack, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if err, ok := cache.errors[key]; ok {
		return nil, err
	}
	return cache.packs[key], nil
}

// Store implements minecraft.ResourcePackCache for non-URL packs. It allows
// the same per-connection cache to retain normal RakNet downloads if a server
// mixes URL and chunk-delivered packs.
func (cache *urlResourcePackCache) Store(ctx context.Context, key minecraft.ResourcePackCacheKey, pack *resource.Pack) error {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if pack == nil {
		return fmt.Errorf("cannot cache a nil resource pack")
	}
	if !key.Matches(pack) {
		return fmt.Errorf("resource pack does not match cache key")
	}
	cache.mu.Lock()
	_, alreadyStored := cache.packs[key]
	cache.packs[key] = pack
	if !alreadyStored {
		cache.packOrder = append(cache.packOrder, key)
	}
	cache.mu.Unlock()
	return nil
}

// WaitForOffer waits for the upstream ResourcePacksInfo packet and, when it
// advertises URL packs, for those downloads to finish. The first timeout keeps
// the proxy tolerant of implementations that omit that optional phase. The
// longer download timeout prevents a slow pack from being exposed downstream
// before the cache is ready.
func (cache *urlResourcePackCache) WaitForOffer(ctx context.Context) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(urlResourcePackOfferWait)
	defer timer.Stop()
	select {
	case <-cache.offered:
	case <-timer.C:
		return false
	case <-ctx.Done():
		return false
	}
	cache.mu.Lock()
	urlCount := cache.urlCount
	cache.mu.Unlock()
	if urlCount == 0 {
		return true
	}
	downloadTimer := time.NewTimer(urlResourcePackDownloadWait)
	defer downloadTimer.Stop()
	select {
	case <-cache.ready:
		return true
	case <-downloadTimer.C:
		return false
	case <-ctx.Done():
		return false
	}
}

// Packs returns the successfully prefetched URL packs in ResourcePacksInfo
// order. It is used while gophertunnel is still completing its own packet
// handler and therefore conn.ResourcePacks may not yet contain the entries.
func (cache *urlResourcePackCache) Packs() []*resource.Pack {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	packs := make([]*resource.Pack, 0, len(cache.packOrder))
	for _, key := range cache.packOrder {
		if pack := cache.packs[key]; pack != nil {
			packs = append(packs, pack)
		}
	}
	return packs
}

func decodeResourcePacksInfo(payload []byte) (info *packet.ResourcePacksInfo, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("decode ResourcePacksInfo: %v", recovered)
			info = nil
		}
	}()
	info = &packet.ResourcePacksInfo{}
	reader := protocol.NewReader(bytes.NewReader(payload), 0, true)
	info.Marshal(reader)
	return info, nil
}

var _ minecraft.ResourcePackCache = (*urlResourcePackCache)(nil)
