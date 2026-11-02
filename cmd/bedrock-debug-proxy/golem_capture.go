package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"

	"github.com/NhanAZ/BedrockDebugProxy/internal/bedrock"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/NhanAZ/BedrockDebugProxy/internal/packetview"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// golemCapture adapts the direct network client to the canonical capture
// contract used by analysis, release reports, and BedrockPackArchive. It is
// deliberately kept in the CLI package so the network client remains useful
// without importing capture or resource-pack storage policy.
type golemCapture struct {
	recorder  *capture.Recorder
	sessionID string
	decrypt   bool

	mu             sync.Mutex
	firstError     error
	encryptedPacks int
}

func newGolemCapture(recorder *capture.Recorder, sessionID string, decrypt bool) *golemCapture {
	return &golemCapture{recorder: recorder, sessionID: sessionID, decrypt: decrypt}
}

func (g *golemCapture) record(ctx context.Context, event capture.Event, raw []byte, mediaType, representation string) error {
	if g == nil || g.recorder == nil {
		return nil
	}
	_, err := g.recorder.Record(ctx, capture.Record{
		Event:          event,
		Raw:            raw,
		MediaType:      mediaType,
		Representation: representation,
	})
	if err != nil {
		g.setError(err)
	}
	return err
}

func (g *golemCapture) recordJSON(ctx context.Context, kind string, severity capture.Severity, direction capture.Direction, channel, stage string, hop int, value any) error {
	var data []byte
	var err error
	if value != nil {
		data, err = json.Marshal(value)
		if err != nil {
			g.setError(err)
			return fmt.Errorf("encode %s event: %w", kind, err)
		}
	}
	return g.record(ctx, capture.Event{
		SessionID:    g.sessionID,
		ConnectionID: golemConnectionID(hop),
		Hop:          hop,
		Kind:         kind,
		Severity:     severity,
		Direction:    direction,
		Channel:      channel,
		Stage:        stage,
		Data:         data,
	}, nil, "", "")
}

func (g *golemCapture) onStart() error {
	return g.recordJSON(context.Background(), "golem.start", capture.SeverityInfo, capture.DirectionInternal, "golem", "lifecycle", 1, map[string]any{
		"client":       "go-network-client",
		"capture_mode": "direct_upstream",
	})
}

func (g *golemCapture) onConnection(conn *minecraft.Conn, hop int) {
	if conn == nil {
		g.setError(errors.New("golem connection observer received a nil connection"))
		return
	}
	if err := g.record(context.Background(), capture.Event{
		SessionID:    g.sessionID,
		ConnectionID: golemConnectionID(hop),
		Hop:          hop,
		Kind:         "upstream.connected",
		Severity:     capture.SeverityInfo,
		Direction:    capture.DirectionInternal,
		Channel:      "upstream",
		Stage:        "transport_connected",
		Source:       golemEndpoint(conn.LocalAddr()),
		Destination:  golemEndpoint(conn.RemoteAddr()),
		Annotations:  []string{"Direct Golem connection; no retail client is involved"},
	}, nil, "", ""); err != nil {
		return
	}
	if err := g.recordStructured(context.Background(), "session.connection_metadata", connectionMetadata("upstream", conn), capture.DirectionInternal, hop, "decoded_login_state", []string{
		"Exact Login packet bytes remain in packet.raw; binary fields in this structured view use size, SHA-256, and preview metadata",
	}); err != nil {
		return
	}
}

func (g *golemCapture) onPacket(conn *minecraft.Conn, hop int, header packet.Header, payload []byte, source, destination net.Addr) {
	if conn == nil {
		g.setError(errors.New("golem packet observer received a nil connection"))
		return
	}
	direction := golemPacketDirection(conn, source, destination)
	info := &capture.PacketInfo{
		ID:              header.PacketID,
		SenderSubClient: header.SenderSubClient,
		TargetSubClient: header.TargetSubClient,
		DecodeStatus:    "not_yet_decoded",
	}
	if factory, ok := conn.Proto().Packets(direction == capture.DirectionClientToServer)[header.PacketID]; ok {
		name, goType := golemPacketNames(factory())
		info.Name = name
		info.GoType = goType
	}
	if err := g.record(context.Background(), capture.Event{
		SessionID:    g.sessionID,
		ConnectionID: golemConnectionID(hop),
		Hop:          hop,
		Kind:         "packet.raw",
		Severity:     capture.SeverityInfo,
		Direction:    direction,
		Channel:      "upstream",
		Stage:        "post_decrypt_decompress_frame",
		Source:       golemEndpoint(source),
		Destination:  golemEndpoint(destination),
		Packet:       info,
	}, payload, "application/octet-stream", "bedrock_packet_payload"); err != nil {
		return
	}

	decoded, decodeErr := golemDecodePacket(conn, header, payload, direction)
	if decodeErr != nil {
		info.DecodeStatus = "decode_error"
		if err := g.record(context.Background(), capture.Event{
			SessionID:    g.sessionID,
			ConnectionID: golemConnectionID(hop),
			Hop:          hop,
			Kind:         "packet.decode_error",
			Severity:     capture.SeverityError,
			Direction:    direction,
			Channel:      "upstream",
			Stage:        "decoded_latest_protocol",
			Packet:       info,
			Error: &capture.ErrorInfo{
				Operation: "decode packet for canonical capture",
				Message:   decodeErr.Error(),
				Type:      fmt.Sprintf("%T", decodeErr),
			},
		}, nil, "", ""); err != nil {
			return
		}
		g.setError(decodeErr)
		return
	}
	for _, value := range decoded {
		if err := g.recordDecoded(context.Background(), value, direction, hop); err != nil {
			g.setError(err)
			return
		}
	}
}

func (g *golemCapture) onResourcePacks(ctx context.Context, conn *minecraft.Conn, hop int) error {
	if conn == nil {
		return errors.New("golem resource-pack observer received a nil connection")
	}
	for _, pack := range conn.ResourcePacks() {
		if pack != nil && pack.Encrypted() {
			g.mu.Lock()
			g.encryptedPacks++
			g.mu.Unlock()
		}
	}
	return bedrock.RecordResourcePacks(ctx, g.recorder, g.sessionID, golemConnectionID(hop), hop, conn.RemoteAddr(), conn.LocalAddr(), conn.ResourcePacks(), bedrock.ResourcePackCaptureOptions{Decrypt: g.decrypt})
}

func (g *golemCapture) onSpawn(ctx context.Context, conn *minecraft.Conn, hop int) error {
	if conn == nil {
		return errors.New("golem spawn observer received a nil connection")
	}
	clientData := conn.ClientData()
	settings, settingsOK := conn.NetworkSettings()
	negotiated := map[string]any{
		"downstream_protocol_id":              conn.Proto().ID(),
		"downstream_game_version":             conn.Proto().Ver(),
		"upstream_protocol_id":                conn.Proto().ID(),
		"upstream_game_version":               conn.Proto().Ver(),
		"downstream_client_data_game_version": clientData.GameVersion,
		"upstream_client_data_game_version":   clientData.GameVersion,
		"client_data_game_version_match":      true,
		"downstream_client_data_device_os":    clientData.DeviceOS,
		"upstream_client_data_device_os":      clientData.DeviceOS,
		"client_data_device_os_match":         true,
		"network_settings_available":          settingsOK,
		"upstream_network_settings":           golemNetworkSettingsMetadata(settings, settingsOK),
		"downstream_network_settings":         golemNetworkSettingsMetadata(settings, settingsOK),
		"network_settings_match":              settingsOK,
		"resource_pack_count":                 len(conn.ResourcePacks()),
		"client_mode":                         "direct_golem_network_client",
	}
	if err := g.recordJSON(ctx, "session.negotiated", capture.SeverityInfo, capture.DirectionInternal, "upstream", "negotiated", hop, negotiated); err != nil {
		return err
	}
	if err := g.recordStructured(ctx, "session.game_data", conn.GameData(), capture.DirectionServerToClient, hop, "decoded_upstream_game_data", []string{
		"This is a bounded GameData view exposed by gophertunnel; exact startup packets remain in packet.raw",
	}); err != nil {
		return err
	}
	return g.recordJSON(ctx, "session.spawned", capture.SeverityInfo, capture.DirectionInternal, "upstream", "lifecycle", hop, map[string]any{
		"world_spawned":       true,
		"resource_pack_count": len(conn.ResourcePacks()),
	})
}

func (g *golemCapture) onClose(err error, snapshot any, hop int) {
	value := map[string]any{"world_spawned": true}
	if snapshot != nil {
		value["client_snapshot"] = snapshot
	}
	if err != nil {
		value["close_error"] = err.Error()
	}
	if recordErr := g.recordJSON(context.Background(), "session.close", severityForGolemError(err), capture.DirectionInternal, "upstream", "lifecycle", hop, value); recordErr != nil {
		g.setError(recordErr)
	}
}

func (g *golemCapture) recordStructured(ctx context.Context, kind string, value any, direction capture.Direction, hop int, stage string, annotations []string) error {
	fields, notices, err := packetview.Encode(value, packetview.Options{MaxCollectionItems: 512, BinaryPreviewBytes: 64})
	if err != nil {
		return g.record(ctx, capture.Event{
			SessionID:    g.sessionID,
			ConnectionID: golemConnectionID(hop),
			Hop:          hop,
			Kind:         "capture.view_error",
			Severity:     capture.SeverityError,
			Direction:    direction,
			Channel:      "upstream",
			Stage:        stage,
			Error:        &capture.ErrorInfo{Operation: "encode " + kind, Message: err.Error(), Type: fmt.Sprintf("%T", err)},
		}, nil, "", "")
	}
	data, err := json.Marshal(map[string]any{"fields": json.RawMessage(fields), "notices": notices})
	if err != nil {
		return fmt.Errorf("encode structured view envelope: %w", err)
	}
	return g.record(ctx, capture.Event{
		SessionID:    g.sessionID,
		ConnectionID: golemConnectionID(hop),
		Hop:          hop,
		Kind:         kind,
		Severity:     capture.SeverityInfo,
		Direction:    direction,
		Channel:      "upstream",
		Stage:        stage,
		Data:         data,
		Annotations:  annotations,
	}, nil, "", "")
}

func (g *golemCapture) recordDecoded(ctx context.Context, decoded packet.Packet, direction capture.Direction, hop int) error {
	name, goType := golemPacketNames(decoded)
	status := "decoded"
	kind := "packet.decoded"
	if _, ok := decoded.(*packet.Unknown); ok {
		status = "unknown"
		kind = "packet.unknown"
	}
	fields, notices, err := packetview.Encode(decoded, packetview.Options{MaxCollectionItems: 512, BinaryPreviewBytes: 64})
	if err != nil {
		return g.record(ctx, capture.Event{
			SessionID:    g.sessionID,
			ConnectionID: golemConnectionID(hop),
			Hop:          hop,
			Kind:         "packet.decode_error",
			Severity:     capture.SeverityError,
			Direction:    direction,
			Channel:      "upstream",
			Stage:        "decoded_latest_protocol",
			Packet:       &capture.PacketInfo{ID: decoded.ID(), Name: name, GoType: goType, DecodeStatus: "view_error"},
			Error:        &capture.ErrorInfo{Operation: "encode decoded packet view", Message: err.Error(), Type: fmt.Sprintf("%T", err)},
		}, nil, "", "")
	}
	data, err := json.Marshal(map[string]any{"fields": json.RawMessage(fields), "notices": notices})
	if err != nil {
		return fmt.Errorf("encode packet view envelope: %w", err)
	}
	return g.record(ctx, capture.Event{
		SessionID:    g.sessionID,
		ConnectionID: golemConnectionID(hop),
		Hop:          hop,
		Kind:         kind,
		Severity:     capture.SeverityInfo,
		Direction:    direction,
		Channel:      "upstream",
		Stage:        "decoded_latest_protocol",
		Packet:       &capture.PacketInfo{ID: decoded.ID(), Name: name, GoType: goType, DecodeStatus: status},
		Data:         data,
	}, nil, "", "")
}

func (g *golemCapture) setError(err error) {
	if err == nil {
		return
	}
	g.mu.Lock()
	if g.firstError == nil {
		g.firstError = err
	}
	g.mu.Unlock()
}

func (g *golemCapture) error() error {
	if g == nil {
		return nil
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.firstError
}

func (g *golemCapture) encryptedPackCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.encryptedPacks
}

func (g *golemCapture) decryptionErrorCount() int {
	if g == nil || g.recorder == nil {
		return 0
	}
	count := 0
	_ = capture.ScanEvents(g.recorder.Root(), func(event capture.Event) error {
		if event.Kind == "resource_pack.decrypt_error" {
			count++
		}
		return nil
	})
	return count
}

func golemDecodePacket(conn *minecraft.Conn, header packet.Header, payload []byte, direction capture.Direction) (decodedPackets []packet.Packet, err error) {
	if conn == nil {
		return nil, errors.New("connection is nil")
	}
	pool := conn.Proto().Packets(direction == capture.DirectionClientToServer)
	factory, ok := pool[header.PacketID]
	if !ok {
		return []packet.Packet{&packet.Unknown{PacketID: header.PacketID, Payload: append([]byte(nil), payload...)}}, nil
	}
	decoded := factory()
	reader := bytes.NewBuffer(payload)
	defer func() {
		if recovered := recover(); recovered != nil {
			decodedPackets = nil
			err = fmt.Errorf("decode packet %T: %v", decoded, recovered)
		}
	}()
	decoded.Marshal(conn.Proto().NewReader(reader, 0, true))
	if reader.Len() != 0 {
		return nil, fmt.Errorf("decode packet %T: %d unread bytes", decoded, reader.Len())
	}
	converted := conn.Proto().ConvertToLatest(decoded, conn)
	if len(converted) == 0 {
		return nil, fmt.Errorf("decode packet %T produced no latest-protocol packet", decoded)
	}
	return converted, nil
}

func connectionMetadata(role string, conn *minecraft.Conn) map[string]any {
	identity := conn.IdentityData()
	return map[string]any{
		"role":          role,
		"authenticated": conn.Authenticated(),
		"protocol_id":   conn.Proto().ID(),
		"game_version":  conn.Proto().Ver(),
		"identity": map[string]any{
			"xuid":             identity.XUID,
			"identity":         identity.Identity,
			"display_name":     identity.DisplayName,
			"title_id":         identity.TitleID,
			"playfab_title_id": identity.PlayFabTitleID,
			"playfab_id":       identity.PlayFabID,
		},
		"client_data": conn.ClientData(),
	}
}

func golemNetworkSettingsMetadata(settings packet.NetworkSettings, available bool) map[string]any {
	if !available {
		return nil
	}
	return map[string]any{
		"compression_threshold":     settings.CompressionThreshold,
		"compression_algorithm":     settings.CompressionAlgorithm,
		"client_throttle":           settings.ClientThrottle,
		"client_throttle_threshold": settings.ClientThrottleThreshold,
		"client_throttle_scalar":    settings.ClientThrottleScalar,
	}
}

func golemConnectionID(hop int) string { return fmt.Sprintf("upstream-%d", hop) }

func golemEndpoint(address net.Addr) *capture.Endpoint {
	if address == nil {
		return nil
	}
	return &capture.Endpoint{Network: address.Network(), Address: address.String()}
}

func golemPacketDirection(conn *minecraft.Conn, source, destination net.Addr) capture.Direction {
	if conn == nil || source == nil || destination == nil || conn.LocalAddr() == nil || conn.RemoteAddr() == nil {
		return capture.DirectionUnknown
	}
	local := conn.LocalAddr().String()
	remote := conn.RemoteAddr().String()
	sourceValue := source.String()
	destinationValue := destination.String()
	if sourceValue == local && destinationValue == remote {
		return capture.DirectionClientToServer
	}
	if sourceValue == remote && destinationValue == local {
		return capture.DirectionServerToClient
	}
	return capture.DirectionUnknown
}

func golemPacketNames(value packet.Packet) (string, string) {
	if value == nil {
		return "", "<nil>"
	}
	typeInfo := reflect.TypeOf(value)
	goType := typeInfo.String()
	if typeInfo.Kind() == reflect.Pointer {
		typeInfo = typeInfo.Elem()
	}
	return typeInfo.Name(), goType
}

func severityForGolemError(err error) capture.Severity {
	if err == nil {
		return capture.SeverityInfo
	}
	return capture.SeverityWarn
}
