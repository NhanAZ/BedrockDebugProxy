package proxy

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestTransferForProxyAddressPreservesSemantics(t *testing.T) {
	original := &packet.Transfer{Address: "[2001:db8::20]", Port: 19133, ReloadWorld: true}
	rewritten, target, err := transferForProxyAddress(original, "192.168.1.10:19132")
	if err != nil {
		t.Fatalf("transferForProxyAddress() error = %v", err)
	}
	if target != "[2001:db8::20]:19133" {
		t.Fatalf("target = %q", target)
	}
	if rewritten.Address != "192.168.1.10" || rewritten.Port != 19132 || !rewritten.ReloadWorld {
		t.Fatalf("rewritten = %#v", rewritten)
	}
	if !reflect.DeepEqual(rewritten.GatheringJoinInfo, original.GatheringJoinInfo) {
		t.Fatalf("GatheringJoinInfo was not preserved")
	}
}

func TestTransferForProxyAddressRejectsWildcardListener(t *testing.T) {
	_, _, err := transferForProxyAddress(&packet.Transfer{Address: "next.example", Port: 19132}, "0.0.0.0:19132")
	if err == nil {
		t.Fatal("transferForProxyAddress() accepted wildcard listener")
	}
}

func TestRunnerFollowsTransferAndAcceptsNextHop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	firstListener, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = firstListener.Close() }()
	secondListener, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = secondListener.Close() }()
	serverErrors := make(chan error, 2)
	go serveTransferHop(ctx, firstListener, secondListener.Addr().String(), serverErrors, true)
	go serveTransferHop(ctx, secondListener, "", serverErrors, false)

	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{SyncEachEvent: true})
	if err != nil {
		t.Fatal(err)
	}
	output := newListeningAddressWriter()
	runner, err := New(Config{
		ListenAddress: "127.0.0.1:0", UpstreamAddress: firstListener.Addr().String(),
		AllowUnauthenticatedClient: true, FollowTransfers: true, Recorder: recorder, Output: output,
	})
	if err != nil {
		t.Fatal(err)
	}
	runnerErrors := make(chan error, 1)
	go func() { runnerErrors <- runner.Run(ctx) }()
	var proxyAddress string
	select {
	case proxyAddress = <-output.address:
	case <-ctx.Done():
		t.Fatalf("wait for proxy listener: %v", ctx.Err())
	}
	dial := func() *minecraft.Conn {
		client, dialErr := (minecraft.Dialer{IdentityData: login.IdentityData{Identity: "11111111-1111-1111-1111-111111111111", DisplayName: "TransferInt"}, ClientData: integrationClientData()}).DialContext(ctx, "raknet", proxyAddress)
		if dialErr != nil {
			t.Logf("proxy output: %s", output.buffer.String())
			t.Fatalf("dial proxy: %v", dialErr)
		}
		return client
	}
	client := dial()
	if err := client.DoSpawnContext(ctx); err != nil {
		_ = client.Close()
		t.Fatal(err)
	}
	var transfer *packet.Transfer
	for {
		value, readErr := client.ReadPacket()
		if readErr != nil {
			t.Logf("proxy output after transfer read: %s", output.buffer.String())
			_ = client.Close()
			t.Fatalf("read transfer: %v", readErr)
		}
		if typed, ok := value.(*packet.Transfer); ok {
			transfer = typed
			break
		}
	}
	if transfer.Address != "127.0.0.1" || transfer.Port == 0 {
		_ = client.Close()
		t.Fatalf("client received transfer = %#v", transfer)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	client = dial()
	if err := client.DoSpawnContext(ctx); err != nil {
		_ = client.Close()
		t.Fatal(err)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case runErr := <-runnerErrors:
		if runErr != nil {
			t.Fatalf("Run() error = %v", runErr)
		}
	case <-ctx.Done():
		t.Fatalf("wait for proxy shutdown: %v", ctx.Err())
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	verification, err := capture.Verify(root)
	if err != nil || len(verification.Issues) != 0 {
		t.Fatalf("capture verification = %+v, error = %v", verification, err)
	}
	var transferSeen, rewriteSeen, routeProbeSeen, hopTwo bool
	if err := capture.ScanEvents(root, func(event capture.Event) error {
		if event.Kind == "packet.decoded" && event.Packet != nil && event.Packet.Name == "Transfer" && event.Hop == 1 {
			transferSeen = true
		}
		if event.Kind == "packet.transfer_rewrite" {
			rewriteSeen = true
		}
		if event.Kind == "transfer.route_probe" && event.Hop == 2 {
			routeProbeSeen = true
		}
		if event.Hop == 2 && event.Kind == "session.spawned" {
			hopTwo = true
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !transferSeen || !rewriteSeen || !routeProbeSeen || !hopTwo {
		t.Fatalf("transfer evidence = transfer %v, rewrite %v, route probe %v, hop2 %v", transferSeen, rewriteSeen, routeProbeSeen, hopTwo)
	}
	for i := 0; i < 2; i++ {
		select {
		case serverErr := <-serverErrors:
			if serverErr != nil {
				t.Fatal(serverErr)
			}
		case <-ctx.Done():
			t.Fatalf("wait for upstream servers: %v", ctx.Err())
		}
	}
}

func serveTransferHop(ctx context.Context, listener *minecraft.Listener, target string, errorsChannel chan<- error, sendTransfer bool) {
	accepted, err := listener.Accept()
	if err != nil {
		errorsChannel <- fmt.Errorf("accept transfer hop: %w", err)
		return
	}
	conn, ok := accepted.(*minecraft.Conn)
	if !ok {
		_ = accepted.Close()
		errorsChannel <- fmt.Errorf("unexpected transfer hop connection %T", accepted)
		return
	}
	defer func() { _ = conn.Close() }()
	if err := conn.StartGame(minecraft.GameData{WorldName: "Transfer integration", EntityUniqueID: 1, EntityRuntimeID: 1}); err != nil {
		errorsChannel <- fmt.Errorf("start transfer hop game: %w", err)
		return
	}
	if !sendTransfer {
		for {
			if _, err := conn.ReadPacket(); err != nil {
				errorsChannel <- nil
				return
			}
		}
	}
	host, portText, splitErr := net.SplitHostPort(target)
	if splitErr != nil {
		errorsChannel <- fmt.Errorf("split transfer target: %w", splitErr)
		return
	}
	port, parseErr := strconv.ParseUint(portText, 10, 16)
	if parseErr != nil {
		errorsChannel <- fmt.Errorf("parse transfer target port: %w", parseErr)
		return
	}
	if err := conn.WritePacket(&packet.Transfer{Address: host, Port: uint16(port)}); err != nil {
		errorsChannel <- fmt.Errorf("write transfer: %w", err)
		return
	}
	if err := conn.Flush(); err != nil {
		errorsChannel <- fmt.Errorf("flush transfer: %w", err)
	}
	errorsChannel <- nil
}
