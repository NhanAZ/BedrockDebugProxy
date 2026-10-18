package minecraft

import (
	"bytes"
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestDeferredResourcePackCompletionWaitsForDoSpawn(t *testing.T) {
	conn := newLoginTestConn()
	local, remote := net.Pipe()
	defer func() { _ = local.Close() }()
	defer func() { _ = remote.Close() }()
	conn.conn = local
	var output bytes.Buffer
	conn.enc = packet.NewEncoder(&output)
	conn.resourcePackCompletionDeferred = make(chan struct{})
	conn.resourcePackCompletionResume = make(chan struct{})
	stack := &packet.ResourcePackStack{BaseGameVersion: "synthetic"}
	if err := conn.handleResourcePackStack(stack, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case <-conn.resourcePackCompletionDeferred:
	default:
		t.Fatal("validated stack did not publish the dial boundary")
	}
	if err := conn.Flush(); err != nil {
		t.Fatal(err)
	}
	if output.Len() != 0 || conn.gameData.ResourcePackStack.BaseGameVersion != "synthetic" {
		t.Fatal("deferred stack did not retain metadata without sending completion")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := conn.DoSpawnContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled spawn error = %v", err)
	}
	select {
	case <-conn.resourcePackCompletionResume:
		t.Fatal("cancelled spawn resumed world startup")
	default:
	}
	close(conn.spawn)
	var calls sync.WaitGroup
	for range 8 {
		calls.Go(func() {
			if err := conn.DoSpawnContext(context.Background()); err != nil {
				t.Error(err)
			}
		})
	}
	calls.Wait()
	select {
	case <-conn.resourcePackCompletionResume:
	default:
		t.Fatal("DoSpawn did not release the decoder")
	}
}

func TestDeferredResourcePackCompletionRejectsMissingArchive(t *testing.T) {
	conn := newLoginTestConn()
	conn.resourcePackCompletionDeferred = make(chan struct{})
	err := conn.handleResourcePackStack(&packet.ResourcePackStack{TexturePacks: []protocol.StackResourcePack{{
		UUID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Version: "1.0.0",
	}}}, nil)
	if err == nil {
		t.Fatal("missing archive was accepted")
	}
	select {
	case <-conn.resourcePackCompletionDeferred:
		t.Fatal("invalid stack published a successful dial boundary")
	default:
	}
}

func TestDeferredResourcePackCompletionRejectsExchangeOnly(t *testing.T) {
	dialer := Dialer{ResourcePackExchangeOnly: true, DeferResourcePackCompletion: true}
	if _, err := dialer.DialContextNetwork(context.Background(), nil, "unused"); err == nil {
		t.Fatal("conflicting completion modes were accepted")
	}
}
