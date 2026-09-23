package minecraft

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestResourcePackChunkCountBounds(t *testing.T) {
	tests := []struct {
		name      string
		size      uint64
		chunkSize uint32
		want      uint32
		ok        bool
	}{
		{name: "empty", size: 0, chunkSize: 1024, want: 0, ok: true},
		{name: "partial final chunk", size: 1025, chunkSize: 1024, want: 2, ok: true},
		{name: "zero chunk size", size: 1, chunkSize: 0, ok: false},
		{name: "too many chunks", size: maxResourcePackChunkCount + 1, chunkSize: 1, ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resourcePackChunkCount(tt.size, tt.chunkSize)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("resourcePackChunkCount(%d, %d) = (%d, %t), want (%d, %t)", tt.size, tt.chunkSize, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestResourcePackDownloadWindowReplenishesAfterResponse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn := newResourcePackPipelineTestConn(ctx)
	conn.packQueue = &resourcePackQueue{
		packAmount:       1,
		downloadingPacks: map[string]*downloadingPack{"pack": {size: 101, buf: bytes.NewBuffer(nil)}},
		awaitingPacks:    make(map[string]*downloadingPack),
	}
	if err := conn.handleResourcePackDataInfo(&packet.ResourcePackDataInfo{UUID: "pack_1", Size: 101, DataChunkSize: 1}); err != nil {
		t.Fatalf("handle ResourcePackDataInfo: %v", err)
	}

	pack := conn.packQueue.awaitingPacks["pack"]
	waitForResourcePackPackets(t, conn, int(DefaultResourcePackMaxInFlightChunks))
	pack.mu.Lock()
	requested := len(pack.requested)
	pack.mu.Unlock()
	if requested != int(DefaultResourcePackMaxInFlightChunks) {
		t.Fatalf("initial requested chunk count = %d, want %d", requested, DefaultResourcePackMaxInFlightChunks)
	}

	if err := conn.handleResourcePackChunkData(&packet.ResourcePackChunkData{UUID: "pack_1", ChunkIndex: 0, Data: []byte{0}}); err != nil {
		t.Fatalf("handle first ResourcePackChunkData: %v", err)
	}
	waitForResourcePackPackets(t, conn, int(DefaultResourcePackMaxInFlightChunks)+1)
	cancel()
}

func TestResourcePackDataInfoTransitionsConcurrentPacksSafely(t *testing.T) {
	const packCount = 128
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn := newResourcePackPipelineTestConn(ctx)
	conn.packQueue = &resourcePackQueue{
		downloadingPacks: make(map[string]*downloadingPack, packCount),
		awaitingPacks:    make(map[string]*downloadingPack, packCount),
	}
	for index := range packCount {
		id := fmt.Sprintf("pack-%d", index)
		conn.packQueue.downloadingPacks[id] = &downloadingPack{
			size: 1,
			buf:  bytes.NewBuffer(nil),
		}
	}

	start := make(chan struct{})
	var handlers sync.WaitGroup
	for index := range packCount {
		index := index
		handlers.Add(1)
		go func() {
			defer handlers.Done()
			<-start
			id := fmt.Sprintf("pack-%d", index)
			if err := conn.handleResourcePackDataInfo(&packet.ResourcePackDataInfo{
				UUID:          id + "_1",
				Size:          1,
				DataChunkSize: 1,
			}); err != nil {
				t.Errorf("handle ResourcePackDataInfo for %s: %v", id, err)
			}
		}()
	}
	close(start)
	handlers.Wait()
	waitForResourcePackPackets(t, conn, packCount)

	conn.packMu.Lock()
	awaiting := len(conn.packQueue.awaitingPacks)
	downloading := len(conn.packQueue.downloadingPacks)
	conn.packMu.Unlock()
	if awaiting != packCount || downloading != 0 {
		t.Fatalf("resource-pack queue has %d awaiting and %d downloading packs, want %d and 0", awaiting, downloading, packCount)
	}
}

func TestHandleResourcePackChunkDataRequiresRequestedChunkAndCopiesBytes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fragment := make([]byte, 2)
	conn := newResourcePackPipelineTestConn(ctx)
	conn.packQueue = &resourcePackQueue{
		awaitingPacks: map[string]*downloadingPack{
			"pack": {
				size:       4,
				chunkSize:  2,
				chunkCount: 2,
				requested:  map[uint32]struct{}{1: {}},
				newFrag:    make(chan resourcePackChunk, 1),
			},
		},
	}
	if err := conn.handleResourcePackChunkData(&packet.ResourcePackChunkData{UUID: "pack_1", ChunkIndex: 1, Data: fragment}); err != nil {
		t.Fatalf("handle requested chunk: %v", err)
	}
	fragment[0] = 9
	got := <-conn.packQueue.awaitingPacks["pack"].newFrag
	if got.index != 1 || !bytes.Equal(got.data, []byte{0, 0}) {
		t.Fatalf("queued fragment = %#v, want copied chunk 1", got)
	}
	if err := conn.handleResourcePackChunkData(&packet.ResourcePackChunkData{UUID: "pack_1", ChunkIndex: 1, Data: []byte{0, 0}}); err == nil {
		t.Fatal("duplicate unrequested chunk was accepted")
	}
}

func newResourcePackPipelineTestConn(ctx context.Context) *Conn {
	return &Conn{
		ctx:   ctx,
		log:   slog.Default(),
		proto: DefaultProtocol,
		enc:   packet.NewEncoder(io.Discard),
		hdr:   &packet.Header{},
	}
}

func waitForResourcePackPackets(t *testing.T, conn *Conn, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn.sendMu.Lock()
		got := len(conn.bufferedSend)
		conn.sendMu.Unlock()
		if got >= want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	conn.sendMu.Lock()
	got := len(conn.bufferedSend)
	conn.sendMu.Unlock()
	t.Fatalf("buffered resource-pack requests = %d, want at least %d", got, want)
}
