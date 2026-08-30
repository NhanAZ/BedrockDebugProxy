package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"reflect"
	"sync"

	"github.com/NhanAZ/BedrockDebugProxy/internal/bedrock"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/NhanAZ/BedrockDebugProxy/internal/packetview"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

const (
	sessionID              = "session-1"
	downstreamConnectionID = "downstream-1"
	upstreamConnectionID   = "upstream-1"
)

type upstreamSlot struct {
	once sync.Once
	conn *minecraft.Conn
	err  error
}

type Runner struct {
	config Config
}

type connectionMetadata struct {
	Role          string           `json:"role"`
	Authenticated bool             `json:"authenticated"`
	ProtocolID    int32            `json:"protocol_id"`
	GameVersion   string           `json:"game_version"`
	Identity      identityMetadata `json:"identity"`
	ClientData    login.ClientData `json:"client_data"`
}

type identityMetadata struct {
	XUID           string `json:"xuid"`
	Identity       string `json:"identity"`
	DisplayName    string `json:"display_name"`
	TitleID        string `json:"title_id,omitempty"`
	PlayFabTitleID string `json:"playfab_title_id,omitempty"`
	PlayFabID      string `json:"playfab_id,omitempty"`
}

func New(config Config) (*Runner, error) {
	if err := config.normalize(); err != nil {
		return nil, err
	}
	return &Runner{config: config}, nil
}

func (r *Runner) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()
	failures := bedrock.NewFailureSink(cancelRun)
	observer := bedrock.NewObserver(r.config.Recorder, failures, sessionID, 1)
	downstreamLogger := slog.New(bedrock.NewCaptureLogHandler(r.config.Recorder, failures, sessionID, downstreamConnectionID, "downstream"))
	upstreamLogger := slog.New(bedrock.NewCaptureLogHandler(r.config.Recorder, failures, sessionID, upstreamConnectionID, "upstream"))
	downstreamNetwork := bedrock.Network{
		Recorder:       r.config.Recorder,
		Observer:       observer,
		Failures:       failures,
		Logger:         downstreamLogger,
		SessionID:      sessionID,
		ConnectionID:   downstreamConnectionID,
		Channel:        "downstream",
		Hop:            1,
		ReadDirection:  capture.DirectionClientToServer,
		WriteDirection: capture.DirectionServerToClient,
	}
	upstreamNetwork := bedrock.Network{
		Recorder:       r.config.Recorder,
		Observer:       observer,
		Failures:       failures,
		Logger:         upstreamLogger,
		SessionID:      sessionID,
		ConnectionID:   upstreamConnectionID,
		Channel:        "upstream",
		Hop:            1,
		ReadDirection:  capture.DirectionServerToClient,
		WriteDirection: capture.DirectionClientToServer,
	}

	if err := r.record("proxy.start", capture.SeverityInfo, map[string]any{
		"listen_address":   r.config.ListenAddress,
		"upstream_address": r.config.UpstreamAddress,
		"protocol_id":      minecraft.DefaultProtocol.ID(),
		"game_version":     minecraft.DefaultProtocol.Ver(),
	}); err != nil {
		return err
	}

	var slot upstreamSlot
	listenerConfig := minecraft.ListenConfig{
		ErrorLog:               downstreamLogger,
		AuthenticationDisabled: r.config.AllowUnauthenticatedClient,
		MaximumPlayers:         1,
		AllowUnknownPackets:    true,
		AllowInvalidPackets:    true,
		StatusProvider:         minecraft.NewStatusProvider("BedrockDebugProxy", r.config.UpstreamAddress),
		MaxDecompressedLen:     r.config.MaxDecompressedBytes,
		PacketFunc:             observer.PacketFunc("downstream", downstreamConnectionID),
		FetchResourcePacks: func(_ login.IdentityData, clientData login.ClientData, _ []*resource.Pack) []*resource.Pack {
			slot.once.Do(func() {
				slot.conn, slot.err = r.connectUpstream(runCtx, upstreamNetwork, observer, upstreamLogger, clientData)
			})
			if slot.err != nil || slot.conn == nil {
				return nil
			}
			return slot.conn.ResourcePacks()
		},
	}
	listener, err := listenerConfig.ListenNetwork(downstreamNetwork, r.config.ListenAddress)
	if err != nil {
		return fmt.Errorf("listen for Bedrock client: %w", err)
	}
	defer func() { _ = listener.Close() }()
	stopAccept := context.AfterFunc(runCtx, func() { _ = listener.Close() })
	defer stopAccept()

	actualAddress := listener.Addr().String()
	_, _ = fmt.Fprintf(r.config.Output, "Listening on %s and forwarding to %s.\n", actualAddress, r.config.UpstreamAddress)
	if err := r.record("proxy.listening", capture.SeverityInfo, map[string]any{"listen_address": actualAddress}); err != nil {
		return err
	}

	accepted, err := listener.Accept()
	if err != nil {
		if failure := failures.Err(); failure != nil {
			return fmt.Errorf("capture hook failed during login: %w", failure)
		}
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("accept Bedrock client: %w", err)
	}
	clientConn, ok := accepted.(*minecraft.Conn)
	if !ok {
		_ = accepted.Close()
		return fmt.Errorf("accepted unexpected connection type %T", accepted)
	}
	defer func() { _ = clientConn.Close() }()
	if slot.err != nil {
		_ = listener.Disconnect(clientConn, "BedrockDebugProxy could not connect to the destination server")
		return slot.err
	}
	if slot.conn == nil {
		_ = listener.Disconnect(clientConn, "BedrockDebugProxy did not establish an upstream connection")
		return errors.New("upstream connection was not established during resource-pack negotiation")
	}
	serverConn := slot.conn
	defer func() { _ = serverConn.Close() }()
	stopConnections := context.AfterFunc(runCtx, func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	})
	defer stopConnections()
	if err := failures.Err(); err != nil {
		return fmt.Errorf("capture hook failed during login: %w", err)
	}
	if err := r.record("session.negotiated", capture.SeverityInfo, map[string]any{
		"downstream_protocol_id":  clientConn.Proto().ID(),
		"downstream_game_version": clientConn.Proto().Ver(),
		"upstream_protocol_id":    serverConn.Proto().ID(),
		"upstream_game_version":   serverConn.Proto().Ver(),
		"resource_pack_count":     len(serverConn.ResourcePacks()),
	}); err != nil {
		return err
	}
	if err := r.recordStructuredView(
		"session.connection_metadata", newConnectionMetadata("downstream", clientConn),
		capture.DirectionInternal, downstreamConnectionID, "downstream", "decoded_login_state",
		[]string{"Exact Login packet bytes remain in packet.raw; binary fields in this structured view use size, SHA-256, and preview metadata"},
	); err != nil {
		return err
	}
	if err := r.recordStructuredView(
		"session.connection_metadata", newConnectionMetadata("upstream", serverConn),
		capture.DirectionInternal, upstreamConnectionID, "upstream", "decoded_login_state",
		[]string{"Exact Login packet bytes remain in packet.raw; binary fields in this structured view use size, SHA-256, and preview metadata"},
	); err != nil {
		return err
	}
	gameData := serverConn.GameData()
	if err := r.recordStructuredView(
		"session.game_data", gameData, capture.DirectionServerToClient, upstreamConnectionID, "upstream", "decoded_upstream_game_data",
		[]string{"This is a bounded GameData view exposed by gophertunnel; collections may be truncated and the exact StartGame packet remains in packet.raw"},
	); err != nil {
		return err
	}
	if err := startGame(clientConn, serverConn, gameData); err != nil {
		_ = r.record("session.spawn_error", capture.SeverityError, map[string]any{"error": err.Error(), "type": fmt.Sprintf("%T", err)})
		return err
	}
	if err := r.record("session.spawned", capture.SeverityInfo, map[string]any{
		"downstream_latency":              clientConn.Latency().String(),
		"upstream_latency":                serverConn.Latency().String(),
		"downstream_client_cache_enabled": clientConn.ClientCacheEnabled(),
		"upstream_client_cache_enabled":   serverConn.ClientCacheEnabled(),
		"downstream_chunk_radius":         clientConn.ChunkRadius(),
		"upstream_chunk_radius":           serverConn.ChunkRadius(),
	}); err != nil {
		return err
	}

	first := make(chan error, 2)
	go func() {
		first <- r.forward(clientConn, serverConn, failures, capture.DirectionClientToServer, "downstream", downstreamConnectionID)
	}()
	go func() {
		first <- r.forward(serverConn, clientConn, failures, capture.DirectionServerToClient, "upstream", upstreamConnectionID)
	}()
	bridgeErr := <-first
	_ = clientConn.Close()
	_ = serverConn.Close()
	secondErr := <-first
	if err := failures.Err(); err != nil {
		return fmt.Errorf("capture hook failed: %w", err)
	}
	if err := r.record("session.close", severityForError(bridgeErr), map[string]any{
		"first_loop_error":  errorString(bridgeErr),
		"second_loop_error": errorString(secondErr),
	}); err != nil {
		return err
	}
	return nil
}

func (r *Runner) connectUpstream(ctx context.Context, network bedrock.Network, observer *bedrock.Observer, logger *slog.Logger, clientData login.ClientData) (*minecraft.Conn, error) {
	if err := r.record("upstream.dial_start", capture.SeverityInfo, map[string]any{"address": r.config.UpstreamAddress}); err != nil {
		return nil, err
	}
	dialer := minecraft.Dialer{
		TokenSource:                r.config.TokenSource,
		ClientData:                 clientData,
		ErrorLog:                   logger,
		PacketFunc:                 observer.PacketFunc("upstream", upstreamConnectionID),
		DisconnectOnUnknownPackets: false,
		DisconnectOnInvalidPackets: false,
		DownloadResourcePack: func(_ uuid.UUID, _ string, _, _ int) bool {
			return true
		},
	}
	conn, err := dialer.DialContextNetwork(ctx, network, r.config.UpstreamAddress)
	if err != nil {
		_ = r.record("upstream.dial_error", capture.SeverityError, map[string]any{"error": err.Error(), "type": fmt.Sprintf("%T", err)})
		return nil, fmt.Errorf("connect to upstream server: %w", err)
	}
	if err := bedrock.RecordResourcePacks(ctx, r.config.Recorder, sessionID, upstreamConnectionID, 1, conn.RemoteAddr(), conn.LocalAddr(), conn.ResourcePacks(), bedrock.ResourcePackCaptureOptions{
		Decrypt: r.config.DecryptResourcePacks,
	}); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := r.record("upstream.connected", capture.SeverityInfo, map[string]any{
		"local_address":       conn.LocalAddr().String(),
		"remote_address":      conn.RemoteAddr().String(),
		"protocol_id":         conn.Proto().ID(),
		"game_version":        conn.Proto().Ver(),
		"resource_pack_count": len(conn.ResourcePacks()),
	}); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func startGame(client, server *minecraft.Conn, gameData minecraft.GameData) error {
	errorsChannel := make(chan error, 2)
	go func() { errorsChannel <- client.StartGame(gameData) }()
	go func() { errorsChannel <- server.DoSpawn() }()
	first := <-errorsChannel
	if first != nil {
		_ = client.Close()
		_ = server.Close()
	}
	second := <-errorsChannel
	if first != nil || second != nil {
		_ = client.Close()
		_ = server.Close()
		return errors.Join(first, second)
	}
	return nil
}

func newConnectionMetadata(role string, conn *minecraft.Conn) connectionMetadata {
	identity := conn.IdentityData()
	return connectionMetadata{
		Role: role, Authenticated: conn.Authenticated(), ProtocolID: conn.Proto().ID(), GameVersion: conn.Proto().Ver(),
		Identity: identityMetadata{
			XUID: identity.XUID, Identity: identity.Identity, DisplayName: identity.DisplayName,
			TitleID: identity.TitleID, PlayFabTitleID: identity.PlayFabTitleID, PlayFabID: identity.PlayFabID,
		},
		ClientData: conn.ClientData(),
	}
}

func (r *Runner) forward(source, destination *minecraft.Conn, failures *bedrock.FailureSink, direction capture.Direction, sourceChannel, connectionID string) error {
	for {
		if err := failures.Err(); err != nil {
			return err
		}
		decoded, err := source.ReadPacket()
		if err != nil {
			captureErr := r.recordNetworkError("bridge.read_error", direction, sourceChannel, connectionID, "read packet", err)
			return errors.Join(err, captureErr)
		}
		if err := r.recordDecoded(decoded, direction, connectionID); err != nil {
			return err
		}
		if err := destination.WritePacket(decoded); err != nil {
			captureErr := r.recordNetworkError("bridge.write_error", direction, sourceChannel, connectionID, "write packet", err)
			return errors.Join(err, captureErr)
		}
	}
}

func (r *Runner) recordDecoded(decoded packet.Packet, direction capture.Direction, connectionID string) error {
	name, goType := packetNames(decoded)
	status := "decoded"
	kind := "packet.decoded"
	if _, unknown := decoded.(*packet.Unknown); unknown {
		status = "unknown"
		kind = "packet.unknown"
	}
	data, encodeErr := r.structuredView(decoded)
	if encodeErr != nil {
		_, recordErr := r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
			SessionID:    sessionID,
			ConnectionID: connectionID,
			Hop:          1,
			Kind:         "packet.decode_error",
			Severity:     capture.SeverityError,
			Direction:    direction,
			Channel:      "bridge",
			Stage:        "decoded_latest_protocol",
			Packet:       &capture.PacketInfo{ID: decoded.ID(), Name: name, GoType: goType, DecodeStatus: "view_error"},
			Error:        &capture.ErrorInfo{Operation: "encode decoded view", Message: encodeErr.Error(), Type: fmt.Sprintf("%T", encodeErr)},
		}})
		return recordErr
	}
	annotations := make([]string, 0, 1)
	if _, transfer := decoded.(*packet.Transfer); transfer {
		annotations = append(annotations, "Transfer is recorded but automatic hop following is not implemented")
	}
	_, err := r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		SessionID:    sessionID,
		ConnectionID: connectionID,
		Hop:          1,
		Kind:         kind,
		Severity:     capture.SeverityInfo,
		Direction:    direction,
		Channel:      "bridge",
		Stage:        "decoded_latest_protocol",
		Packet:       &capture.PacketInfo{ID: decoded.ID(), Name: name, GoType: goType, DecodeStatus: status},
		Data:         data,
		Annotations:  annotations,
	}})
	return err
}

func (r *Runner) recordStructuredView(kind string, value any, direction capture.Direction, connectionID, channel, stage string, annotations []string) error {
	data, encodeErr := r.structuredView(value)
	if encodeErr != nil {
		_, recordErr := r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
			SessionID: sessionID, ConnectionID: connectionID, Hop: 1,
			Kind: "capture.view_error", Severity: capture.SeverityError, Direction: direction,
			Channel: channel, Stage: stage,
			Error: &capture.ErrorInfo{Operation: "encode " + kind, Message: encodeErr.Error(), Type: fmt.Sprintf("%T", encodeErr)},
		}})
		return recordErr
	}
	_, err := r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		SessionID: sessionID, ConnectionID: connectionID, Hop: 1,
		Kind: kind, Severity: capture.SeverityInfo, Direction: direction,
		Channel: channel, Stage: stage, Data: data, Annotations: annotations,
	}})
	return err
}

func (r *Runner) structuredView(value any) ([]byte, error) {
	fields, notices, err := packetview.Encode(value, packetview.Options{
		MaxCollectionItems: r.config.MaxDecodedCollectionItems,
		BinaryPreviewBytes: r.config.DecodedBinaryPreviewBytes,
	})
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(map[string]any{"fields": json.RawMessage(fields), "notices": notices})
	if err != nil {
		return nil, fmt.Errorf("encode structured view envelope: %w", err)
	}
	return data, nil
}

func (r *Runner) recordNetworkError(kind string, direction capture.Direction, channel, connectionID, operation string, operationErr error) error {
	_, err := r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		SessionID:    sessionID,
		ConnectionID: connectionID,
		Hop:          1,
		Kind:         kind,
		Severity:     capture.SeverityError,
		Direction:    direction,
		Channel:      channel,
		Stage:        "bridge",
		Error:        networkError(operation, operationErr),
	}})
	return err
}

func (r *Runner) record(kind string, severity capture.Severity, value any) error {
	var data []byte
	var err error
	if value != nil {
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("encode %s event: %w", kind, err)
		}
	}
	_, err = r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		SessionID: sessionID,
		Hop:       1,
		Kind:      kind,
		Severity:  severity,
		Direction: capture.DirectionInternal,
		Channel:   "proxy",
		Stage:     "lifecycle",
		Data:      data,
	}})
	return err
}

func packetNames(decoded packet.Packet) (name, goType string) {
	typeInfo := reflect.TypeOf(decoded)
	if typeInfo == nil {
		return "", "<nil>"
	}
	goType = typeInfo.String()
	if typeInfo.Kind() == reflect.Pointer {
		typeInfo = typeInfo.Elem()
	}
	return typeInfo.Name(), goType
}

func networkError(operation string, err error) *capture.ErrorInfo {
	if err == nil {
		return nil
	}
	info := &capture.ErrorInfo{Operation: operation, Message: err.Error(), Type: fmt.Sprintf("%T", err)}
	var netErr net.Error
	if errors.As(err, &netErr) {
		timeout := netErr.Timeout()
		info.Timeout = &timeout
	}
	return info
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func severityForError(err error) capture.Severity {
	if err == nil {
		return capture.SeverityInfo
	}
	return capture.SeverityWarn
}
