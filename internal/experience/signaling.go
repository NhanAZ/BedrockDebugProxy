// SPDX-FileCopyrightText: 2026 NhanAZ
// SPDX-FileCopyrightText: bedrocktool contributors
// SPDX-License-Identifier: GPL-3.0-only
//
// Portions of the legacy signaling path are substantially adapted from
// bedrock-tool/bedrocktool commit d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0
// under GPL-3.0. The JSON-RPC path and integration changes are project-owner
// code. See THIRD_PARTY_NOTICES.md and docs/research/experience-routing.md.
package experience

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/df-mc/go-nethernet"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/service"
)

type signalingConnection interface {
	nethernet.Signaling
	Close() error
}

func dialSignaling(ctx context.Context, client *http.Client, baseURI string, token *service.Token, transport string) (signalingConnection, error) {
	switch transport {
	case transportNetherNet:
		return dialLegacySignaling(ctx, client, baseURI, token)
	case transportJSONRPC:
		return dialJSONRPCSignaling(ctx, client, baseURI, token)
	default:
		return nil, fmt.Errorf("unsupported signaling transport %q", transport)
	}
}

type legacyMessage struct {
	Type int    `json:"Type"`
	From string `json:"From,omitempty"`
	To   uint64 `json:"To,omitempty"`
	Data string `json:"Message,omitempty"`
}

const (
	legacyMessageError       = 0
	legacyMessagePing        = 1
	legacyMessageSignal      = 2
	legacyMessageCredentials = 3
)

type legacyServiceError struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
}

func (e *legacyServiceError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("signaling service error %d", e.Code)
	}
	return fmt.Sprintf("signaling service error %d: %s", e.Code, e.Message)
}

type legacySignaling struct {
	conn      *websocket.Conn
	networkID string
	log       *slog.Logger
	ctx       context.Context
	cancel    context.CancelCauseFunc

	closed    chan struct{}
	closeOnce sync.Once
	writeMu   sync.Mutex

	credentials         atomic.Pointer[nethernet.Credentials]
	credentialsReceived chan struct{}
	credentialsOnce     sync.Once

	notifierMu    sync.Mutex
	notifyCount   uint32
	notifiers     map[uint32]nethernet.Notifier
	pendingSignal []*nethernet.Signal
}

func dialLegacySignaling(ctx context.Context, client *http.Client, baseURI string, token *service.Token) (*legacySignaling, error) {
	networkID, err := randomNetworkID()
	if err != nil {
		return nil, err
	}
	endpoint, err := serviceURL(baseURI, "/ws/v1.0/signaling/"+networkID)
	if err != nil {
		return nil, err
	}
	headers := make(http.Header)
	headers.Set("Authorization", token.AuthorizationHeader)
	// websocket.Dial owns and closes its handshake response body.
	connection, _, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{HTTPClient: client, HTTPHeader: headers}) //nolint:bodyclose
	if err != nil {
		return nil, err
	}
	signalContext, cancel := context.WithCancelCause(ctx)
	signaling := &legacySignaling{
		conn:                connection,
		networkID:           networkID,
		log:                 slog.Default(),
		ctx:                 signalContext,
		cancel:              cancel,
		closed:              make(chan struct{}),
		credentialsReceived: make(chan struct{}),
		notifiers:           make(map[uint32]nethernet.Notifier),
	}
	go signaling.read()
	go signaling.ping()
	return signaling, nil
}

func (s *legacySignaling) Signal(ctx context.Context, signal *nethernet.Signal) error {
	if signal == nil || signal.NetworkID == "" {
		return errors.New("signal network ID is empty")
	}
	remoteID, err := strconv.ParseUint(signal.NetworkID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse signal network ID: %w", err)
	}
	return s.write(ctx, legacyMessage{Type: legacyMessageSignal, To: remoteID, Data: signal.String()})
}

func (s *legacySignaling) Notify(notifier nethernet.Notifier) func() {
	s.notifierMu.Lock()
	id := s.notifyCount
	s.notifyCount++
	s.notifiers[id] = notifier
	pending := s.pendingSignal
	s.pendingSignal = nil
	s.notifierMu.Unlock()
	for _, signal := range pending {
		_ = notifier.NotifySignal(signal)
	}
	return sync.OnceFunc(func() {
		s.notifierMu.Lock()
		delete(s.notifiers, id)
		s.notifierMu.Unlock()
	})
}

func (s *legacySignaling) Context() context.Context { return s.ctx }

func (s *legacySignaling) Credentials(ctx context.Context) (*nethernet.Credentials, error) {
	if credentials := s.credentials.Load(); credentials != nil {
		return credentials, nil
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.closed:
		return nil, context.Cause(s.ctx)
	case <-s.credentialsReceived:
		return s.credentials.Load(), nil
	}
}

func (s *legacySignaling) NetworkID() string { return s.networkID }
func (s *legacySignaling) PongData([]byte)   {}

func (s *legacySignaling) Close() (err error) {
	s.closeOnce.Do(func() {
		close(s.closed)
		s.notifierMu.Lock()
		clear(s.notifiers)
		s.notifierMu.Unlock()
		s.cancel(net.ErrClosed)
		err = s.conn.Close(websocket.StatusNormalClosure, "")
	})
	return err
}

func (s *legacySignaling) read() {
	defer func() { _ = s.Close() }()
	for {
		var message legacyMessage
		if err := wsjson.Read(s.ctx, s.conn, &message); err != nil {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, net.ErrClosed) {
				s.cancel(err)
			}
			return
		}
		switch message.Type {
		case legacyMessageCredentials:
			if message.From != "Server" {
				s.log.Warn("ignored signaling credentials from unexpected sender", slog.String("sender", message.From))
				continue
			}
			var credentials nethernet.Credentials
			if err := json.Unmarshal([]byte(message.Data), &credentials); err != nil {
				s.log.Error("decode signaling credentials", slog.Any("error", err))
				continue
			}
			s.credentials.Store(&credentials)
			s.credentialsOnce.Do(func() { close(s.credentialsReceived) })
		case legacyMessageSignal:
			signal := &nethernet.Signal{}
			if err := signal.UnmarshalText([]byte(message.Data)); err != nil {
				s.log.Error("decode NetherNet signal", slog.Any("error", err))
				continue
			}
			signal.NetworkID = message.From
			s.notifySignal(signal)
		case legacyMessageError:
			var serviceError legacyServiceError
			if err := json.Unmarshal([]byte(message.Data), &serviceError); err != nil {
				s.cancel(fmt.Errorf("decode signaling error: %w", err))
			} else {
				s.cancel(&serviceError)
			}
			return
		}
	}
}

func (s *legacySignaling) notifySignal(signal *nethernet.Signal) {
	s.notifierMu.Lock()
	if len(s.notifiers) == 0 {
		s.pendingSignal = append(s.pendingSignal, signal)
		s.notifierMu.Unlock()
		return
	}
	notifiers := make([]nethernet.Notifier, 0, len(s.notifiers))
	for _, notifier := range s.notifiers {
		notifiers = append(notifiers, notifier)
	}
	s.notifierMu.Unlock()
	for _, notifier := range notifiers {
		_ = notifier.NotifySignal(signal)
	}
}

func (s *legacySignaling) ping() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.write(s.ctx, legacyMessage{Type: legacyMessagePing}); err != nil {
				s.cancel(err)
				return
			}
		}
	}
}

func (s *legacySignaling) write(ctx context.Context, message legacyMessage) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	select {
	case <-s.closed:
		return net.ErrClosed
	default:
		return wsjson.Write(ctx, s.conn, message)
	}
}

type jsonRPCSignaling struct {
	conn      *websocket.Conn
	networkID string
	log       *slog.Logger
	ctx       context.Context
	cancel    context.CancelCauseFunc

	closed    chan struct{}
	closeOnce sync.Once
	writeMu   sync.Mutex

	credentials         atomic.Pointer[nethernet.Credentials]
	credentialsReceived chan struct{}
	credentialsOnce     sync.Once

	notifierMu    sync.Mutex
	notifyCount   uint32
	notifiers     map[uint32]nethernet.Notifier
	pendingSignal []*nethernet.Signal
}

type jsonRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      string `json:"id,omitempty"`
}

type jsonRPCResponse struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	ID     json.RawMessage `json:"id"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type receivedMessage struct {
	From    string `json:"From"`
	ID      string `json:"Id"`
	Message string `json:"Message"`
}

func dialJSONRPCSignaling(ctx context.Context, client *http.Client, baseURI string, token *service.Token) (*jsonRPCSignaling, error) {
	endpoint, err := serviceURL(baseURI, "/ws/v1.0/messaging/connect")
	if err != nil {
		return nil, err
	}
	headers := make(http.Header)
	headers.Set("Authorization", token.AuthorizationHeader)
	headers.Set("session-id", uuid.NewString())
	headers.Set("request-id", uuid.NewString())
	// websocket.Dial owns and closes its handshake response body.
	connection, _, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{HTTPClient: client, HTTPHeader: headers}) //nolint:bodyclose
	if err != nil {
		return nil, err
	}
	networkID, err := randomNetworkID()
	if err != nil {
		_ = connection.Close(websocket.StatusInternalError, "network ID generation failed")
		return nil, err
	}
	signalContext, cancel := context.WithCancelCause(ctx)
	signaling := &jsonRPCSignaling{
		conn:                connection,
		networkID:           networkID,
		log:                 slog.Default(),
		ctx:                 signalContext,
		cancel:              cancel,
		closed:              make(chan struct{}),
		credentialsReceived: make(chan struct{}),
		notifiers:           make(map[uint32]nethernet.Notifier),
	}
	go signaling.read()
	go signaling.ping()
	if err := signaling.writeRequest(ctx, "Signaling_TurnAuth_v1_0", map[string]any{}); err != nil {
		_ = signaling.Close()
		return nil, err
	}
	return signaling, nil
}

func (s *jsonRPCSignaling) Signal(ctx context.Context, signal *nethernet.Signal) error {
	if signal == nil || signal.NetworkID == "" {
		return errors.New("signal network ID is empty")
	}
	inner, err := json.Marshal(jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "Signaling_WebRtc_v1_0",
		Params:  map[string]any{"netherNetId": s.networkID, "message": signal.String()},
	})
	if err != nil {
		return err
	}
	return s.writeRequest(ctx, "Signaling_SendClientMessage_v1_0", map[string]any{
		"toPlayerId": signal.NetworkID,
		"messageId":  uuid.NewString(),
		"message":    string(inner),
	})
}

func (s *jsonRPCSignaling) Notify(notifier nethernet.Notifier) func() {
	s.notifierMu.Lock()
	id := s.notifyCount
	s.notifyCount++
	s.notifiers[id] = notifier
	pending := s.pendingSignal
	s.pendingSignal = nil
	s.notifierMu.Unlock()
	for _, signal := range pending {
		_ = notifier.NotifySignal(signal)
	}
	return sync.OnceFunc(func() {
		s.notifierMu.Lock()
		delete(s.notifiers, id)
		s.notifierMu.Unlock()
	})
}

func (s *jsonRPCSignaling) Context() context.Context { return s.ctx }

func (s *jsonRPCSignaling) Credentials(ctx context.Context) (*nethernet.Credentials, error) {
	if credentials := s.credentials.Load(); credentials != nil {
		return credentials, nil
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.closed:
		return nil, context.Cause(s.ctx)
	case <-s.credentialsReceived:
		return s.credentials.Load(), nil
	}
}

func (s *jsonRPCSignaling) NetworkID() string { return s.networkID }
func (s *jsonRPCSignaling) PongData([]byte)   {}

func (s *jsonRPCSignaling) Close() (err error) {
	s.closeOnce.Do(func() {
		close(s.closed)
		s.notifierMu.Lock()
		clear(s.notifiers)
		s.notifierMu.Unlock()
		s.cancel(net.ErrClosed)
		err = s.conn.Close(websocket.StatusNormalClosure, "")
	})
	return err
}

func (s *jsonRPCSignaling) read() {
	defer func() { _ = s.Close() }()
	for {
		var response jsonRPCResponse
		if err := wsjson.Read(s.ctx, s.conn, &response); err != nil {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, net.ErrClosed) {
				s.cancel(err)
			}
			return
		}
		if response.Error != nil {
			s.cancel(fmt.Errorf("JSON-RPC signaling error %d: %s", response.Error.Code, response.Error.Message))
			return
		}
		if len(response.Result) != 0 && string(response.Result) != "null" {
			var credentials nethernet.Credentials
			if err := json.Unmarshal(response.Result, &credentials); err == nil && len(credentials.ICEServers) != 0 {
				s.credentials.Store(&credentials)
				s.credentialsOnce.Do(func() { close(s.credentialsReceived) })
			}
		}
		switch response.Method {
		case "System_Pong_v1_0":
			_ = s.writeResult(s.ctx, response.ID)
		case "Signaling_ReceiveMessage_v1_0":
			_ = s.writeResult(s.ctx, response.ID)
			messages, err := decodeReceivedMessages(response.Params)
			if err != nil {
				s.log.Error("decode JSON-RPC signaling message", slog.Any("error", err))
				continue
			}
			for _, message := range messages {
				s.handleReceivedMessage(message)
			}
		}
	}
}

func decodeReceivedMessages(raw json.RawMessage) ([]receivedMessage, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	if raw[0] == '[' {
		var messages []receivedMessage
		if err := json.Unmarshal(raw, &messages); err != nil {
			return nil, err
		}
		return messages, nil
	}
	var message receivedMessage
	if err := json.Unmarshal(raw, &message); err != nil {
		return nil, err
	}
	return []receivedMessage{message}, nil
}

func (s *jsonRPCSignaling) handleReceivedMessage(message receivedMessage) {
	if message.From == "" || message.Message == "" {
		return
	}
	_ = s.sendDeliveryNotification(s.ctx, message.From, message.ID)
	var inner struct {
		Method string `json:"method"`
		Params struct {
			Message string `json:"message"`
		} `json:"params"`
	}
	if err := json.Unmarshal([]byte(message.Message), &inner); err != nil {
		s.log.Error("decode nested JSON-RPC signaling message", slog.Any("error", err))
		return
	}
	if inner.Method != "Signaling_WebRtc_v1_0" || inner.Params.Message == "" {
		return
	}
	signal := &nethernet.Signal{}
	if err := signal.UnmarshalText([]byte(inner.Params.Message)); err != nil {
		s.log.Error("decode NetherNet signal", slog.Any("error", err))
		return
	}
	signal.NetworkID = message.From
	s.notifySignal(signal)
}

func (s *jsonRPCSignaling) sendDeliveryNotification(ctx context.Context, recipient, receivedID string) error {
	inner, err := json.Marshal(jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "Signaling_DeliveryNotification_V1_0",
		Params:  map[string]any{"messageId": receivedID},
	})
	if err != nil {
		return err
	}
	return s.writeRequest(ctx, "Signaling_SendClientMessage_v1_0", map[string]any{
		"toPlayerId": recipient,
		"messageId":  uuid.NewString(),
		"message":    string(inner),
	})
}

func (s *jsonRPCSignaling) notifySignal(signal *nethernet.Signal) {
	s.notifierMu.Lock()
	if len(s.notifiers) == 0 {
		s.pendingSignal = append(s.pendingSignal, signal)
		s.notifierMu.Unlock()
		return
	}
	notifiers := make([]nethernet.Notifier, 0, len(s.notifiers))
	for _, notifier := range s.notifiers {
		notifiers = append(notifiers, notifier)
	}
	s.notifierMu.Unlock()
	for _, notifier := range notifiers {
		_ = notifier.NotifySignal(signal)
	}
}

func (s *jsonRPCSignaling) ping() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.writeRequest(s.ctx, "System_Ping_v1_0", map[string]any{}); err != nil {
				s.cancel(err)
				return
			}
		}
	}
}

func (s *jsonRPCSignaling) writeRequest(ctx context.Context, method string, params any) error {
	return s.write(ctx, jsonRPCRequest{JSONRPC: "2.0", Method: method, Params: params, ID: uuid.NewString()})
}

func (s *jsonRPCSignaling) writeResult(ctx context.Context, id json.RawMessage) error {
	return s.write(ctx, struct {
		JSONRPC string          `json:"jsonrpc"`
		Result  any             `json:"result"`
		ID      json.RawMessage `json:"id"`
	}{JSONRPC: "2.0", Result: nil, ID: id})
}

func (s *jsonRPCSignaling) write(ctx context.Context, value any) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	select {
	case <-s.closed:
		return net.ErrClosed
	default:
		return wsjson.Write(ctx, s.conn, value)
	}
}

func randomNetworkID() (string, error) {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate signaling network ID: %w", err)
	}
	return strconv.FormatUint(binary.LittleEndian.Uint64(value[:]), 10), nil
}
