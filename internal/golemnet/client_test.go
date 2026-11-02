package golemnet

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestClientCapturesFormAndSendsActions(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	listener, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	serverErrors := make(chan error, 1)
	responses := make(chan string, 2)
	chats := make(chan string, 2)
	go func() {
		accepted, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverErrors <- acceptErr
			return
		}
		conn, ok := accepted.(*minecraft.Conn)
		if !ok {
			serverErrors <- fmt.Errorf("unexpected server connection %T", accepted)
			_ = accepted.Close()
			return
		}
		defer func() { _ = conn.Close() }()
		if err := conn.StartGame(minecraft.GameData{WorldName: "Golem test", EntityUniqueID: 1, EntityRuntimeID: 1}); err != nil {
			serverErrors <- err
			return
		}
		form := map[string]any{"type": "form", "title": "Hub", "text": "Choose", "buttons": []map[string]string{{"text": "Mines"}, {"text": "BedWars"}}}
		formData, _ := json.Marshal(form)
		if err := conn.WritePacket(&packet.ModalFormRequest{FormID: 7, FormData: formData}); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.Flush(); err != nil {
			serverErrors <- err
			return
		}
		for {
			value, readErr := conn.ReadPacket()
			if readErr != nil {
				serverErrors <- nil
				return
			}
			switch value := value.(type) {
			case *packet.ModalFormResponse:
				response, ok := value.ResponseData.Value()
				if ok {
					responses <- string(response)
				}
			case *packet.Text:
				chats <- value.Message
				if err := conn.WritePacket(&packet.Text{TextType: packet.TextTypeRaw, Message: "received"}); err != nil {
					serverErrors <- err
					return
				}
				if err := conn.Flush(); err != nil {
					serverErrors <- err
					return
				}
			}
		}
	}()

	client, err := New(Config{
		Address:  listener.Addr().String(),
		Identity: login.IdentityData{Identity: "11111111-1111-1111-1111-111111111111", DisplayName: "GolemTest"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	form, err := client.WaitForForm(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if form.ID != 7 || form.Title != "Hub" || len(form.Buttons) != 2 || form.Buttons[1].Text != "BedWars" {
		t.Fatalf("form = %+v", form)
	}
	if err := client.SelectForm(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if err := client.Chat("/help"); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		select {
		case response := <-responses:
			if response != "1" {
				t.Fatalf("form response = %q", response)
			}
		case message := <-chats:
			if message != "/help" {
				t.Fatalf("chat = %q", message)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	for {
		event, eventErr := client.NextEvent(ctx)
		if eventErr != nil {
			t.Fatal(eventErr)
		}
		if event.Kind == "text" && event.Message == "received" {
			break
		}
	}
	report := client.Snapshot()
	if len(report.Forms) != 1 || report.PacketCounts["ModalFormRequest"] != 1 || report.PacketCounts["Text"] < 1 {
		t.Fatalf("report = %+v", report)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-serverErrors; err != nil {
		t.Fatal(err)
	}
}

func TestClientSendsStationaryPlayerAuthInput(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	listener, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	serverErrors := make(chan error, 1)
	inputPackets := make(chan *packet.PlayerAuthInput, 3)
	go func() {
		accepted, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverErrors <- acceptErr
			return
		}
		conn, ok := accepted.(*minecraft.Conn)
		if !ok {
			serverErrors <- fmt.Errorf("unexpected server connection %T", accepted)
			_ = accepted.Close()
			return
		}
		defer func() { _ = conn.Close() }()
		if err := conn.StartGame(minecraft.GameData{
			WorldName:       "Input test",
			EntityUniqueID:  1,
			EntityRuntimeID: 1,
			PlayerPosition:  mgl32.Vec3{12, 64, -8},
			Pitch:           10,
			Yaw:             90,
		}); err != nil {
			serverErrors <- err
			return
		}
		for {
			value, readErr := conn.ReadPacket()
			if readErr != nil {
				serverErrors <- nil
				return
			}
			if input, ok := value.(*packet.PlayerAuthInput); ok {
				select {
				case inputPackets <- input:
				default:
				}
			}
		}
	}()

	client, err := New(Config{Address: listener.Addr().String()})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	var received []*packet.PlayerAuthInput
	for len(received) < 3 {
		select {
		case input := <-inputPackets:
			received = append(received, input)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	if received[0].Position != (mgl32.Vec3{12, 64, -8}) || received[0].Yaw != 90 || received[0].Pitch != 10 {
		t.Fatalf("input position and rotation = %+v", received[0])
	}
	if received[0].InputMode != packet.InputModeTouch || received[0].PlayMode != packet.PlayModeNormal {
		t.Fatalf("input modes = %+v", received[0])
	}
	if received[0].Tick == 0 || received[1].Tick <= received[0].Tick || received[2].Tick <= received[1].Tick {
		t.Fatalf("input ticks = %d, %d, %d", received[0].Tick, received[1].Tick, received[2].Tick)
	}
	report := client.Snapshot()
	if report.ClientInputPackets < 3 || !report.ClientInput || report.ClientInputRate != defaultClientInputInterval.String() {
		t.Fatalf("input report = %+v", report)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-serverErrors; err != nil {
		t.Fatal(err)
	}
}

func TestClientFollowsTransferAndPreservesScenarioEvents(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	first, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = first.Close() }()
	second, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = second.Close() }()
	serverErrors := make(chan error, 2)
	go serveGolemTransferHop(ctx, first, second.Addr().String(), true, serverErrors)
	go serveGolemTransferHop(ctx, second, "", false, serverErrors)

	client, err := New(Config{Address: first.Addr().String(), FollowTransfers: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	for {
		event, eventErr := client.NextEvent(ctx)
		if eventErr != nil {
			t.Fatal(eventErr)
		}
		if event.Kind == "text" && event.Message == "child-ready" {
			break
		}
	}
	report := client.Snapshot()
	if report.Hops != 2 || len(report.Transfers) != 1 || !report.Transfers[0].Followed || report.PacketCounts["Transfer"] != 1 {
		t.Fatalf("transfer report = %+v", report)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := <-serverErrors; err != nil {
			t.Fatal(err)
		}
	}
}

func serveGolemTransferHop(ctx context.Context, listener *minecraft.Listener, target string, transfer bool, errorsChannel chan<- error) {
	accepted, err := listener.Accept()
	if err != nil {
		errorsChannel <- err
		return
	}
	conn, ok := accepted.(*minecraft.Conn)
	if !ok {
		errorsChannel <- fmt.Errorf("unexpected server connection %T", accepted)
		_ = accepted.Close()
		return
	}
	defer func() { _ = conn.Close() }()
	if err := conn.StartGame(minecraft.GameData{WorldName: "Transfer test", EntityUniqueID: 1, EntityRuntimeID: 1}); err != nil {
		errorsChannel <- err
		return
	}
	if transfer {
		host, portText, splitErr := net.SplitHostPort(target)
		if splitErr != nil {
			errorsChannel <- splitErr
			return
		}
		var port uint16
		if _, scanErr := fmt.Sscanf(portText, "%d", &port); scanErr != nil || port == 0 {
			errorsChannel <- fmt.Errorf("invalid transfer port %q", portText)
			return
		}
		if err := conn.WritePacket(&packet.Transfer{Address: host, Port: port}); err != nil {
			errorsChannel <- err
			return
		}
		if err := conn.Flush(); err != nil {
			errorsChannel <- err
			return
		}
	}
	if !transfer {
		if err := conn.WritePacket(&packet.Text{TextType: packet.TextTypeRaw, Message: "child-ready"}); err != nil {
			errorsChannel <- err
			return
		}
		if err := conn.Flush(); err != nil {
			errorsChannel <- err
			return
		}
	}
	for {
		if _, err := conn.ReadPacket(); err != nil {
			if ctx.Err() != nil {
				errorsChannel <- ctx.Err()
			} else {
				errorsChannel <- nil
			}
			return
		}
	}
}
