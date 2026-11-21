package golemnet

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

func TestClientLoadsServerAdvertisedURLResourcePack(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	packUUID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	archive := golemURLResourcePackArchive(t, packUUID, "1.0.0")
	var downloads atomic.Int32
	var offerPacketsObserved atomic.Int32
	var checkingClientLoad atomic.Bool
	packServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if checkingClientLoad.Load() && offerPacketsObserved.Load() != 1 {
			t.Errorf("ResourcePacksInfo observer count before URL fetch = %d, want 1", offerPacketsObserved.Load())
		}
		downloads.Add(1)
		_, _ = w.Write(archive)
	}))
	defer packServer.Close()
	pack, err := resource.ReadURL(packServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	downloads.Store(0)
	checkingClientLoad.Store(true)
	responses := make(chan uint32, 4)
	listener, err := (minecraft.ListenConfig{
		AuthenticationDisabled: true,
		ResourcePacks:          []*resource.Pack{pack},
		PacketFunc: func(header packet.Header, payload []byte, _, _ net.Addr) {
			if header.PacketID != packet.IDResourcePackClientResponse {
				return
			}
			response := &packet.ResourcePackClientResponse{}
			response.Marshal(protocol.NewReader(bytes.NewReader(payload), 0, true))
			select {
			case responses <- response.Response:
			default:
			}
		},
		FetchResourcePacksInfo: func(login.IdentityData, login.ClientData) *packet.ResourcePacksInfo {
			return &packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{{
				UUID: packUUID, Version: "1.0.0", Size: uint64(len(archive)), DownloadURL: packServer.URL,
			}}}
		},
		FetchResourcePackStack: func(login.IdentityData, login.ClientData) *packet.ResourcePackStack {
			return &packet.ResourcePackStack{TexturePacks: []protocol.StackResourcePack{{UUID: packUUID.String(), Version: "1.0.0"}}}
		},
	}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	serverErrors := make(chan error, 1)
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
		serverErrors <- conn.StartGame(minecraft.GameData{WorldName: "URL pack Golem test", EntityUniqueID: 1, EntityRuntimeID: 1})
	}()

	client, err := New(Config{
		Address:  listener.Addr().String(),
		Identity: login.IdentityData{Identity: uuid.NewString(), DisplayName: "URLPackTest"},
		PacketObserver: func(_ *minecraft.Conn, _ int, header packet.Header, _ []byte, _, _ net.Addr) {
			if header.PacketID == packet.IDResourcePacksInfo {
				offerPacketsObserved.Add(1)
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	report := client.Snapshot()
	if !report.Spawned || report.ResourcePacks != 1 {
		t.Fatalf("Golem report = %+v, want one URL pack and spawn", report)
	}
	if got := downloads.Load(); got != 1 {
		t.Fatalf("HTTP pack downloads = %d, want 1", got)
	}
	select {
	case response := <-responses:
		if response != packet.PackResponseAllPacksDownloaded {
			t.Fatalf("first pack response = %v, want AllPacksDownloaded without chunk requests", response)
		}
	case <-ctx.Done():
		t.Fatal("server did not receive the completed pack response")
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-serverErrors; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func golemURLResourcePackArchive(t *testing.T, packUUID uuid.UUID, version string) []byte {
	t.Helper()
	versionParts := make([]int, 3)
	for index, part := range strings.Split(version, ".") {
		if index == len(versionParts) {
			break
		}
		versionParts[index], _ = strconv.Atoi(part)
	}
	manifest, err := json.Marshal(map[string]any{
		"format_version": 2,
		"header": map[string]any{
			"description":        "Synthetic Golem URL pack",
			"name":               "Synthetic Golem URL pack",
			"uuid":               packUUID.String(),
			"version":            versionParts,
			"min_engine_version": []int{1, 20, 0},
		},
		"modules": []any{map[string]any{
			"description": "Synthetic Golem URL pack module",
			"type":        "resources",
			"uuid":        "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
			"version":     versionParts,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, err := writer.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(manifest); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return archive.Bytes()
}

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
		form := map[string]any{"type": "form", "title": "Hub", "text": "Choose", "elements": []map[string]string{{"type": "label", "text": "Choose"}, {"type": "button", "text": "Mines"}, {"type": "button", "text": "BedWars"}}}
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
	if form.ID != 7 || form.Title != "Hub" || len(form.Buttons) != 2 || form.Buttons[0].Index != 1 || form.Buttons[1].Index != 2 || form.Buttons[1].Text != "BedWars" {
		t.Fatalf("form = %+v", form)
	}
	if err := client.SelectForm(ctx, 2); err != nil {
		t.Fatal(err)
	}
	if err := client.Chat("/help"); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		select {
		case response := <-responses:
			if response != "2" {
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

func TestClientUsesObservedHotbarItemInPlayerInput(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	listener, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	serverErrors := make(chan error, 1)
	item := protocol.ItemInstance{StackNetworkID: 1, Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 400}, Count: 1}}
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
		if err := conn.StartGame(minecraft.GameData{WorldName: "Golem item test", EntityUniqueID: 1, EntityRuntimeID: 1}); err != nil {
			serverErrors <- err
			return
		}
		for {
			value, readErr := conn.ReadPacket()
			if readErr != nil {
				serverErrors <- readErr
				return
			}
			if _, ok := value.(*packet.PlayerAuthInput); ok {
				break
			}
		}
		content := make([]protocol.ItemInstance, 9)
		content[2] = item
		if err := conn.WritePacket(&packet.InventoryContent{WindowID: 0, Content: content}); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.WritePacket(&packet.PlayerHotBar{SelectedHotBarSlot: 0, WindowID: 0, SelectHotBarSlot: true}); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.Flush(); err != nil {
			serverErrors <- err
			return
		}
		equipmentSeen, interactionSeen := false, false
		for !interactionSeen {
			value, readErr := conn.ReadPacket()
			if readErr != nil {
				serverErrors <- readErr
				return
			}
			switch value := value.(type) {
			case *packet.MobEquipment:
				if value.EntityRuntimeID != 1 || value.InventorySlot != 2 || value.HotBarSlot != 2 || value.NewItem.Stack.ItemType.NetworkID != 400 {
					serverErrors <- fmt.Errorf("unexpected equipment selection: %+v", value)
					return
				}
				equipmentSeen = true
			case *packet.PlayerAuthInput:
				transaction, ok := value.ItemInteractionData.Value()
				if !ok {
					continue
				}
				if !value.InputData.Load(packet.InputFlagPerformItemInteraction) || transaction.ActionType != protocol.UseItemActionClickAir || transaction.TriggerType != protocol.TriggerTypePlayerInput || transaction.HotBarSlot != 2 || transaction.HeldItem.Stack.ItemType.NetworkID != 400 {
					serverErrors <- fmt.Errorf("unexpected item interaction: flags=%v transaction=%+v", value.InputData, transaction)
					return
				}
				interactionSeen = true
			}
		}
		if !equipmentSeen || !interactionSeen {
			serverErrors <- fmt.Errorf("equipment=%v interaction=%v", equipmentSeen, interactionSeen)
			return
		}
		serverErrors <- nil
	}()

	client, err := New(Config{
		Address:  listener.Addr().String(),
		Identity: login.IdentityData{Identity: "11111111-1111-1111-1111-111111111111", DisplayName: "GolemItemTest"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	for {
		client.stateMu.Lock()
		_, observed := client.inventory[2]
		client.stateMu.Unlock()
		if observed {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := client.UseItem(ctx, 2); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-serverErrors:
		if err != nil {
			t.Fatalf("local server did not observe the item interaction: %v", err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestClientAcknowledgesServerTeleport(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	listener, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	serverErrors := make(chan error, 1)
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
		if err := conn.StartGame(minecraft.GameData{WorldName: "Golem teleport test", EntityUniqueID: 1, EntityRuntimeID: 1}); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.WritePacket(&packet.MovePlayer{EntityRuntimeID: 2, Position: mgl32.Vec3{90, 91, 92}, Mode: packet.MoveModeTeleport}); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.Flush(); err != nil {
			serverErrors <- err
			return
		}
		teleport := &packet.MovePlayer{EntityRuntimeID: 1, Position: mgl32.Vec3{12.5, 64, -7.25}, Yaw: 90, HeadYaw: 90, Pitch: -15, OnGround: true, Mode: packet.MoveModeTeleport}
		if err := conn.WritePacket(teleport); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.Flush(); err != nil {
			serverErrors <- err
			return
		}
		previousTick := uint64(0)
		for _, move := range []*packet.MovePlayer{
			teleport,
			{EntityRuntimeID: 1, Position: mgl32.Vec3{-40, 72, 18}, Yaw: -30, HeadYaw: -30, Pitch: 5, Mode: packet.MoveModeReset},
		} {
			if move != teleport {
				if err := conn.WritePacket(move); err != nil {
					serverErrors <- err
					return
				}
				if err := conn.Flush(); err != nil {
					serverErrors <- err
					return
				}
			}
			for {
				value, readErr := conn.ReadPacket()
				if readErr != nil {
					serverErrors <- readErr
					return
				}
				input, ok := value.(*packet.PlayerAuthInput)
				if !ok || !input.InputData.Load(packet.InputFlagHandledTeleport) {
					continue
				}
				if input.Position != move.Position || input.Yaw != move.Yaw || input.HeadYaw != move.HeadYaw || input.Pitch != move.Pitch {
					serverErrors <- fmt.Errorf("teleport acknowledgement position/rotation = %v/%v/%v/%v, want %v/%v/%v/%v", input.Position, input.Yaw, input.HeadYaw, input.Pitch, move.Position, move.Yaw, move.HeadYaw, move.Pitch)
					return
				}
				if input.InputData.Load(packet.InputFlagVerticalCollision) != move.OnGround || !input.InputData.Load(packet.InputFlagPersistSneak) || input.Tick <= previousTick {
					serverErrors <- fmt.Errorf("teleport acknowledgement flags=%v tick=%d, want ground=%v, persist sneak, and tick after %d", input.InputData, input.Tick, move.OnGround, previousTick)
					return
				}
				previousTick = input.Tick
				break
			}
		}
		serverErrors <- nil
	}()

	client, err := New(Config{
		Address:  listener.Addr().String(),
		Identity: login.IdentityData{Identity: "11111111-1111-1111-1111-111111111111", DisplayName: "GolemTPTest"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-serverErrors:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestParseFormButtonArraysUseZeroBasedIndexes(t *testing.T) {
	form := parseForm(9, []byte(`{"type":"form","buttons":[{"text":"One"},{"text":"Two"}]}`), 1, time.Time{})
	if form.DecodeStatus != "decoded" || len(form.Buttons) != 2 || form.Buttons[0].Index != 0 || form.Buttons[1].Index != 1 {
		t.Fatalf("parsed button form = %+v", form)
	}
}

func TestParseFormFlattensRawTextComponents(t *testing.T) {
	form := parseForm(12, []byte(`{"type":"form","title":{"rawtext":[{"text":"§6§lWEEKEND EVENT!!"},{"text":"|ACRE|"}]},"content":{"rawtext":[{"text":"Sale details"}]},"elements":[{"type":"button","text":{"rawtext":[{"text":"Let's go!"}]}}]}`), 1, time.Time{})
	if form.DecodeStatus != "decoded" {
		t.Fatalf("form decode status = %q, error %q", form.DecodeStatus, form.DecodeError)
	}
	if form.Title != "§6§lWEEKEND EVENT!!|ACRE|" || form.Text != "Sale details" {
		t.Fatalf("form title/text = %q / %q", form.Title, form.Text)
	}
	if len(form.Buttons) != 1 || form.Buttons[0].Index != 0 || form.Buttons[0].Text != "Let's go!" {
		t.Fatalf("form buttons = %+v", form.Buttons)
	}
	if !json.Valid(form.Raw) {
		t.Fatal("raw form was not preserved")
	}
}

func TestClientDoesNotQueueGenericPacketTelemetry(t *testing.T) {
	client, err := New(Config{Address: "127.0.0.1:19132"})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < maxHistory+maxEventBuffer; i++ {
		if !client.record(Event{Kind: "packet", Packet: "LevelChunk", Hop: 1}) {
			t.Fatalf("record returned false at packet %d", i)
		}
	}
	if got := len(client.events); got != 0 {
		t.Fatalf("generic telemetry queue length = %d, want 0", got)
	}
	if got := client.Snapshot().EventCount; got != maxHistory {
		t.Fatalf("bounded history length = %d, want %d", got, maxHistory)
	}
}

func TestClientNotifiesNetworkObservers(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	listener, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	serverErrors := make(chan error, 1)
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
		if err := conn.StartGame(minecraft.GameData{WorldName: "Observer test", EntityUniqueID: 1, EntityRuntimeID: 1}); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.WritePacket(&packet.Text{TextType: packet.TextTypeRaw, Message: "observer-ready"}); err != nil {
			serverErrors <- err
			return
		}
		if err := conn.Flush(); err != nil {
			serverErrors <- err
			return
		}
		for {
			if _, err := conn.ReadPacket(); err != nil {
				serverErrors <- nil
				return
			}
		}
	}()

	var connections atomic.Int32
	var packets atomic.Int32
	client, err := New(Config{
		Address: listener.Addr().String(),
		ConnectionObserver: func(conn *minecraft.Conn, hop int) {
			if conn == nil || hop != 1 {
				t.Errorf("connection observer received conn=%v hop=%d", conn, hop)
			}
			connections.Add(1)
		},
		PacketObserver: func(conn *minecraft.Conn, hop int, header packet.Header, payload []byte, source, destination net.Addr) {
			if conn == nil || hop != 1 || source == nil || destination == nil {
				t.Errorf("packet observer received nil_or_invalid metadata hop=%d id=%d payload=%d source=%v destination=%v", hop, header.PacketID, len(payload), source, destination)
			}
			packets.Add(1)
		},
	})
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
		if event.Kind == "text" && event.Message == "observer-ready" {
			break
		}
	}
	if connections.Load() != 1 || packets.Load() == 0 {
		t.Fatalf("observer counts connections=%d packets=%d", connections.Load(), packets.Load())
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
	var report Report
	metricTicker := time.NewTicker(time.Millisecond)
	defer metricTicker.Stop()
	for {
		report = client.Snapshot()
		if report.ClientInputPackets >= len(received) {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatalf("input report did not catch up with received packets: got %d, received %d: %v", report.ClientInputPackets, len(received), ctx.Err())
		case <-metricTicker.C:
		}
	}
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

func TestClientTransferDialHonorsConnectTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Second)
	defer cancel()
	first, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AllowUnknownPackets: true, AllowInvalidPackets: true}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = first.Close() }()
	serverErrors := make(chan error, 1)
	go serveGolemTransferHop(ctx, first, "127.0.0.1:19133", true, serverErrors)
	network := &blockingTransferDialNetwork{Network: minecraft.NewRakNet(nil)}
	client, err := New(Config{
		Address:         first.Addr().String(),
		Network:         network,
		FollowTransfers: true,
		ConnectTimeout:  time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	transferObserved := false
	for {
		_, eventErr := client.NextEvent(ctx)
		if eventErr == nil {
			transferObserved = true
			continue
		}
		if !transferObserved {
			t.Fatalf("transfer event was not observed before dial error: %v", eventErr)
		}
		if !errors.Is(eventErr, context.DeadlineExceeded) {
			t.Fatalf("transfer dial error = %v, want deadline exceeded", eventErr)
		}
		break
	}
	if elapsed := time.Since(started); elapsed > 3*time.Second {
		t.Fatalf("transfer dial took %s, want the configured 1s deadline", elapsed)
	}
	report := client.Snapshot()
	if len(report.Transfers) != 1 || report.Transfers[0].Followed || report.Transfers[0].FollowError == "" {
		t.Fatalf("transfer report = %+v", report.Transfers)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-serverErrors; err != nil {
		t.Fatal(err)
	}
}

type blockingTransferDialNetwork struct {
	minecraft.Network
	dials atomic.Int32
}

func (n *blockingTransferDialNetwork) DialContext(ctx context.Context, address string) (net.Conn, error) {
	n.dials.Add(1)
	return n.Network.DialContext(ctx, address)
}

func (n *blockingTransferDialNetwork) PingContext(ctx context.Context, address string) ([]byte, error) {
	if n.dials.Load() > 0 {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return n.Network.PingContext(ctx, address)
}

func TestServiceNetherNetTransferIDClassification(t *testing.T) {
	const target = "22222222-2222-4222-8222-222222222222"
	for _, test := range []struct {
		name      string
		transport string
		address   string
		port      uint16
		want      bool
	}{
		{name: "NetherNet", transport: "nethernet", address: target, want: true},
		{name: "JSON RPC NetherNet", transport: "nethernet-jsonrpc", address: target, want: true},
		{name: "RakNet", transport: "raknet", address: target},
		{name: "nonzero port", transport: "nethernet", address: target, port: 19132},
		{name: "invalid ID", transport: "nethernet", address: "not-an-id"},
	} {
		t.Run(test.name, func(t *testing.T) {
			gotID, got := serviceNetherNetTransferID(test.transport, test.address, test.port)
			if got != test.want || (got && gotID != target) {
				t.Fatalf("serviceNetherNetTransferID() = %q, %t, want %q, %t", gotID, got, target, test.want)
			}
		})
	}
}

func TestServiceNetherNetTransferTargetReusesNetworkAndStripsZeroPort(t *testing.T) {
	const address = "22222222-2222-4222-8222-222222222222"
	network := minecraft.NewRakNet(nil)
	selectedNetwork, target, selected, err := serviceNetherNetTransferTarget("nethernet", network, &Transfer{Address: address, Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	if selectedNetwork != network || target != address || !selected {
		t.Fatalf("selected route = network %T, target %q, selected %t", selectedNetwork, target, selected)
	}
}

func TestServiceNetherNetTransferTargetRequiresExistingNetwork(t *testing.T) {
	_, _, selected, err := serviceNetherNetTransferTarget("nethernet", nil, &Transfer{
		Address: "22222222-2222-4222-8222-222222222222",
		Port:    0,
	})
	if !selected || err == nil || err.Error() != "service NetherNet transfer has no signaling network" {
		t.Fatalf("service route selected=%t error=%v", selected, err)
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
