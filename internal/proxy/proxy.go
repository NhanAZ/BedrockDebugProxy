package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	sessionID                 = "session-1"
	transferGracePeriod       = 5 * time.Second
	transferRouteProbeTimeout = 750 * time.Millisecond
	// A resource-pack exchange can legitimately take several minutes on a
	// featured experience. Galaxite can advertise many packs and pace chunk
	// responses across them, so a one-minute deadline aborts an active exchange.
	// Keep the wait bounded while allowing the observed multi-pack flow to
	// complete.
	upstreamDialTimeout = 5 * time.Minute
	// Enchanted's featured-experience edge can drop fragmented 1492-byte
	// handshake probes on the transfer target. Keep transfer probes within a
	// single common UDP datagram while leaving initial connections unchanged.
	transferRakNetMaxMTU = 1200
	maxFollowedHops      = 64
)

type upstreamSlot struct {
	once          sync.Once
	conn          *minecraft.Conn
	resourcePacks []*resource.Pack
	err           error
}

type hopState struct {
	hop          int
	target       string
	downstreamID string
	upstreamID   string
	// transferLocalAddr is the UDP source address used by the preceding hop.
	// Featured-experience transfer frontends may select a backend by the
	// source flow, so the next RakNet dial must be able to reuse it.
	transferLocalAddr   *net.UDPAddr
	transferClientGUID  int64
	transferRouteProbed bool
	upstreamNetwork     bedrock.Network
	upstreamLogger      *slog.Logger
	slot                upstreamSlot
}

type transferError struct {
	target     string
	localAddr  *net.UDPAddr
	clientGUID int64
}

func (e *transferError) Error() string {
	return "server transfer to " + e.target
}

type transferRewriteError struct {
	err error
}

func (e *transferRewriteError) Error() string { return "rewrite Transfer: " + e.err.Error() }
func (e *transferRewriteError) Unwrap() error { return e.err }

type upstreamDialTimeoutError struct {
	timeout time.Duration
	err     error
}

func (e *upstreamDialTimeoutError) Error() string {
	return fmt.Sprintf("upstream login and resource-pack exchange timed out after %s: %v", e.timeout, e.err)
}

func (e *upstreamDialTimeoutError) Unwrap() error { return e.err }

func udpAddress(address net.Addr) *net.UDPAddr {
	if address == nil {
		return nil
	}
	if udp, ok := address.(*net.UDPAddr); ok {
		copy := *udp
		copy.IP = append(net.IP(nil), udp.IP...)
		return &copy
	}
	udp, err := net.ResolveUDPAddr("udp", address.String())
	if err != nil {
		return nil
	}
	return udp
}

type Runner struct {
	config Config
}

// connectionContext tracks the transport lifecycle for the physical client
// connection currently entering the listener. The listener's resource-pack
// callback does not receive a *minecraft.Conn, so the transport hook supplies
// the context that is canceled when that client disconnects.
type connectionContext struct {
	mu         sync.RWMutex
	ctx        context.Context
	generation uint64
}

func (c *connectionContext) Set(ctx context.Context) uint64 {
	if ctx == nil {
		return 0
	}
	c.mu.Lock()
	c.generation++
	c.ctx = ctx
	generation := c.generation
	c.mu.Unlock()
	return generation
}

func (c *connectionContext) Get() (context.Context, uint64) {
	c.mu.RLock()
	ctx := c.ctx
	generation := c.generation
	c.mu.RUnlock()
	return ctx, generation
}

func (c *connectionContext) Clear() {
	c.mu.Lock()
	c.generation++
	c.ctx = nil
	c.mu.Unlock()
}

// linkedConnectionContext cancels work when either the proxy run or the
// physical downstream connection ends.
func linkedConnectionContext(parent, connection context.Context) (context.Context, func()) {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	if connection == nil {
		return ctx, cancel
	}
	stop := context.AfterFunc(connection, cancel)
	return ctx, func() {
		stop()
		cancel()
	}
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
	live := newLiveReporter(r.config.Output)
	defer func() {
		live.SetShuttingDown()
		live.Close()
	}()
	failures := bedrock.NewFailureSink(cancelRun)
	observer := bedrock.NewObserver(r.config.Recorder, failures, sessionID, 1, live.RawPacket)
	live.SetFollowTransfers(r.config.FollowTransfers)
	downstreamContext := new(connectionContext)
	downstreamDisconnect := make(chan uint64, 1)
	stateMu := &sync.RWMutex{}
	current := &hopState{hop: 1, target: r.config.UpstreamAddress, downstreamID: "downstream-1", upstreamID: "upstream-1"}
	var pending *hopState
	getAcceptState := func() *hopState {
		stateMu.RLock()
		defer stateMu.RUnlock()
		if pending != nil {
			return pending
		}
		return current
	}
	setPendingState := func(state *hopState) {
		stateMu.Lock()
		pending = state
		stateMu.Unlock()
	}
	promotePendingState := func() {
		stateMu.Lock()
		if pending != nil {
			current = pending
			pending = nil
		}
		stateMu.Unlock()
	}
	downstreamLogger := slog.New(bedrock.NewCaptureLogHandler(r.config.Recorder, failures, sessionID, current.downstreamID, "downstream", live.LibraryLog))
	downstreamNetwork := bedrock.Network{
		Recorder:         r.config.Recorder,
		Observer:         observer,
		Failures:         failures,
		Logger:           downstreamLogger,
		SessionID:        sessionID,
		ConnectionID:     current.downstreamID,
		ConnectionIDFunc: func() string { return getAcceptState().downstreamID },
		Channel:          "downstream",
		Hop:              1,
		HopFunc:          func() int { return getAcceptState().hop },
		ReadDirection:    capture.DirectionClientToServer,
		WriteDirection:   capture.DirectionServerToClient,
		ConnectionContextFunc: func(ctx context.Context) {
			generation := downstreamContext.Set(ctx)
			context.AfterFunc(ctx, func() {
				select {
				case downstreamDisconnect <- generation:
				default:
				}
			})
		},
	}

	if err := r.record("proxy.start", capture.SeverityInfo, map[string]any{
		"listen_address":               r.config.ListenAddress,
		"upstream_address":             r.config.UpstreamAddress,
		"allow_unauthenticated_client": r.config.AllowUnauthenticatedClient,
		"follow_transfers":             r.config.FollowTransfers,
		"protocol_id":                  minecraft.DefaultProtocol.ID(),
		"game_version":                 minecraft.DefaultProtocol.Ver(),
	}); err != nil {
		return err
	}

	listenerConfig := minecraft.ListenConfig{
		ErrorLog:               downstreamLogger,
		AuthenticationDisabled: r.config.AllowUnauthenticatedClient,
		MaximumPlayers:         1,
		TexturePacksRequired:   true,
		AllowUnknownPackets:    true,
		AllowInvalidPackets:    true,
		StatusProvider:         minecraft.NewStatusProvider("BedrockDebugProxy", r.config.UpstreamAddress),
		MaxDecompressedLen:     r.config.MaxDecompressedBytes,
		PacketFunc:             observer.PacketFuncDynamic("downstream", func() string { return getAcceptState().downstreamID }),
		FetchResourcePacks: func(_ login.IdentityData, clientData login.ClientData, _ []*resource.Pack) []*resource.Pack {
			state := getAcceptState()
			state.slot.once.Do(func() {
				clientContext, _ := downstreamContext.Get()
				connectCtx, releaseContext := linkedConnectionContext(runCtx, clientContext)
				defer releaseContext()
				state.upstreamLogger = slog.New(bedrock.NewCaptureLogHandler(r.config.Recorder, failures, sessionID, state.upstreamID, "upstream", live.LibraryLog))
				transport := r.config.UpstreamNetwork
				if transport == nil && (state.transferLocalAddr != nil || state.transferClientGUID != 0) {
					raknetTransport := minecraft.NewRakNet(state.upstreamLogger)
					raknetTransport.LocalAddr = state.transferLocalAddr
					raknetTransport.ClientGUID = state.transferClientGUID
					raknetTransport.MaxMTU = transferRakNetMaxMTU
					transport = raknetTransport
				}
				state.upstreamNetwork = bedrock.Network{
					Transport: transport, Recorder: r.config.Recorder, Observer: observer, Failures: failures,
					Logger: state.upstreamLogger, SessionID: sessionID, ConnectionID: state.upstreamID,
					Channel: "upstream", Hop: state.hop, ReadDirection: capture.DirectionServerToClient,
					WriteDirection: capture.DirectionClientToServer,
				}
				state.slot.conn, state.slot.resourcePacks, state.slot.err = r.connectUpstream(connectCtx, state.target, state.upstreamNetwork, observer, state.upstreamID, state.hop, state.upstreamLogger, live, clientData, state.transferLocalAddr != nil, state.transferRouteProbed)
			})
			if state.slot.err != nil || state.slot.conn == nil {
				return nil
			}
			if len(state.slot.resourcePacks) != 0 {
				return state.slot.resourcePacks
			}
			return state.slot.conn.ResourcePacks()
		},
	}
	if r.config.FollowTransfers {
		// A reconnect can overlap the listener's asynchronous player-count
		// decrement by a few milliseconds after a Transfer. Allow the next
		// login to reach Accept while the previous Conn is being released.
		listenerConfig.MaximumPlayers = 0
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

	for {
		// A transport context may have completed while the previous accepted
		// connection was being released. Discard that notification before
		// waiting for the next physical client connection.
		draining := true
		for draining {
			select {
			case generation := <-downstreamDisconnect:
				clientContext, currentGeneration := downstreamContext.Get()
				if generation == currentGeneration && clientContext != nil {
					select {
					case <-clientContext.Done():
						downstreamContext.Clear()
						return nil
					default:
					}
				}
			default:
				draining = false
			}
		}
		clientContext, _ := downstreamContext.Get()
		if clientContext != nil {
			select {
			case <-clientContext.Done():
				downstreamContext.Clear()
				return nil
			default:
			}
		}
		state := getAcceptState()
		observer.SetHop(state.hop)
		type acceptResult struct {
			conn net.Conn
			err  error
		}
		acceptedResult := make(chan acceptResult, 1)
		go func() {
			accepted, err := listener.Accept()
			acceptedResult <- acceptResult{conn: accepted, err: err}
		}()
		var accepted net.Conn
		var err error
		waitingForAccept := true
		for waitingForAccept {
			select {
			case result := <-acceptedResult:
				accepted, err = result.conn, result.err
				waitingForAccept = false
			case generation := <-downstreamDisconnect:
				currentContext, currentGeneration := downstreamContext.Get()
				if generation != currentGeneration || currentContext == nil {
					// A previous physical connection closed after a new one had
					// already taken its place. Keep waiting for the current client.
					continue
				}
				_ = listener.Close()
				result := <-acceptedResult
				if result.conn != nil {
					_ = result.conn.Close()
				}
				downstreamContext.Clear()
				return nil
			case <-runCtx.Done():
				_ = listener.Close()
				result := <-acceptedResult
				if result.conn != nil {
					_ = result.conn.Close()
				}
				return nil
			}
		}
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
		downstreamContext.Clear()
		authentication := "Xbox-authenticated"
		if !clientConn.Authenticated() {
			authentication = "self-signed client accepted by explicit configuration"
		}
		live.Info("Client connected - %s, Bedrock %s protocol %d (hop %d)", authentication, clientConn.Proto().Ver(), clientConn.Proto().ID(), state.hop)
		if state.slot.err != nil {
			reason := "BedrockDebugProxy could not connect to the destination server"
			var timeoutErr *upstreamDialTimeoutError
			if errors.As(state.slot.err, &timeoutErr) {
				reason = fmt.Sprintf("The destination server did not respond within %s", timeoutErr.timeout)
			}
			_ = listener.Disconnect(clientConn, reason)
			_ = clientConn.Close()
			return state.slot.err
		}
		if state.slot.conn == nil {
			_ = listener.Disconnect(clientConn, "BedrockDebugProxy did not establish an upstream connection")
			_ = clientConn.Close()
			return errors.New("upstream connection was not established during resource-pack negotiation")
		}
		bridgeErr := r.runHop(runCtx, actualAddress, state, clientConn, state.slot.conn, observer, failures, live, func(target string) {
			setPendingState(&hopState{hop: state.hop + 1, target: target, downstreamID: fmt.Sprintf("downstream-%d", state.hop+1), upstreamID: fmt.Sprintf("upstream-%d", state.hop+1)})
		})
		var followed *transferError
		if r.config.FollowTransfers && errors.As(bridgeErr, &followed) {
			if state.hop >= maxFollowedHops {
				return fmt.Errorf("transfer hop limit %d reached at %s", maxFollowedHops, followed.target)
			}
			stateMu.Lock()
			if pending != nil {
				pending.transferLocalAddr = followed.localAddr
				pending.transferClientGUID = followed.clientGUID
			}
			stateMu.Unlock()
			if followed.localAddr != nil && r.config.UpstreamNetwork == nil {
				if err := r.probeTransferRoute(runCtx, followed.target, followed.localAddr, followed.clientGUID, state.hop+1, state.upstreamLogger, live); err != nil {
					live.Info("Transfer route probe failed - %s", err)
				}
				stateMu.Lock()
				if pending != nil {
					pending.transferRouteProbed = true
				}
				stateMu.Unlock()
			}
			live.Info("Following transfer to %s; waiting for the client to reconnect", followed.target)
			select {
			case <-clientConn.Context().Done():
			case <-time.After(transferGracePeriod):
				_ = clientConn.Close()
			case <-runCtx.Done():
				_ = clientConn.Close()
				return nil
			}
			promotePendingState()
			// Release the listener's single-player slot before accepting the
			// reconnect. The client normally closes this connection after
			// receiving Transfer, but the server-side Conn may observe that close
			// slightly later.
			_ = clientConn.Close()
			deadline := time.NewTimer(transferGracePeriod)
			ticker := time.NewTicker(25 * time.Millisecond)
			waitForSlot := true
			for waitForSlot && listener.PlayerCount() != 0 {
				select {
				case <-ticker.C:
				case <-deadline.C:
					waitForSlot = false
				case <-runCtx.Done():
					ticker.Stop()
					if !deadline.Stop() {
						<-deadline.C
					}
					return nil
				}
			}
			ticker.Stop()
			if !deadline.Stop() {
				select {
				case <-deadline.C:
				default:
				}
			}
			continue
		}
		_ = clientConn.Close()
		if bridgeErr != nil {
			return bridgeErr
		}
		return nil
	}
}

func (r *Runner) runHop(runCtx context.Context, localAddress string, state *hopState, clientConn, serverConn *minecraft.Conn, observer *bedrock.Observer, failures *bedrock.FailureSink, live *liveReporter, onTransfer func(string)) (runErr error) {
	defer func() { _ = serverConn.Close() }()
	keepClientOpen := false
	defer func() {
		if !keepClientOpen {
			_ = clientConn.Close()
		}
	}()
	var shuttingDown atomic.Bool
	stopConnections := context.AfterFunc(runCtx, func() {
		shuttingDown.Store(true)
		_ = clientConn.Close()
		_ = serverConn.Close()
	})
	defer stopConnections()
	if err := failures.Err(); err != nil {
		return fmt.Errorf("capture hook failed during login: %w", err)
	}
	if err := r.recordAtHop("session.negotiated", capture.SeverityInfo, map[string]any{
		"downstream_protocol_id": clientConn.Proto().ID(), "downstream_game_version": clientConn.Proto().Ver(),
		"upstream_protocol_id": serverConn.Proto().ID(), "upstream_game_version": serverConn.Proto().Ver(),
		"resource_pack_count": len(serverConn.ResourcePacks()), "target": state.target,
	}, state.hop); err != nil {
		return err
	}
	live.Info("Session negotiated - downstream protocol %d, upstream protocol %d, %d resource packs (hop %d)", clientConn.Proto().ID(), serverConn.Proto().ID(), len(serverConn.ResourcePacks()), state.hop)
	if err := r.recordStructuredViewAtHop("session.connection_metadata", newConnectionMetadata("downstream", clientConn), capture.DirectionInternal, state.downstreamID, "downstream", "decoded_login_state", state.hop,
		[]string{"Exact Login packet bytes remain in packet.raw; binary fields in this structured view use size, SHA-256, and preview metadata"}); err != nil {
		return err
	}
	if err := r.recordStructuredViewAtHop("session.connection_metadata", newConnectionMetadata("upstream", serverConn), capture.DirectionInternal, state.upstreamID, "upstream", "decoded_login_state", state.hop,
		[]string{"Exact Login packet bytes remain in packet.raw; binary fields in this structured view use size, SHA-256, and preview metadata"}); err != nil {
		return err
	}
	gameData := serverConn.GameData()
	if err := r.recordStructuredViewAtHop("session.game_data", gameData, capture.DirectionServerToClient, state.upstreamID, "upstream", "decoded_upstream_game_data", state.hop,
		[]string{"This is a bounded GameData view exposed by gophertunnel; collections may be truncated and the exact StartGame packet remains in packet.raw"}); err != nil {
		return err
	}
	if err := startGame(clientConn, serverConn, gameData); err != nil {
		_ = r.recordAtHop("session.spawn_error", capture.SeverityError, map[string]any{"error": err.Error(), "type": fmt.Sprintf("%T", err)}, state.hop)
		return err
	}
	live.SetSpawned()
	if err := r.recordAtHop("session.spawned", capture.SeverityInfo, map[string]any{
		"downstream_latency": clientConn.Latency().String(), "upstream_latency": serverConn.Latency().String(),
		"downstream_client_cache_enabled": clientConn.ClientCacheEnabled(), "upstream_client_cache_enabled": serverConn.ClientCacheEnabled(),
		"downstream_chunk_radius": clientConn.ChunkRadius(), "upstream_chunk_radius": serverConn.ChunkRadius(),
	}, state.hop); err != nil {
		return err
	}
	live.Info("Session spawned - downstream latency %s, upstream latency %s (hop %d)", clientConn.Latency(), serverConn.Latency(), state.hop)
	transferAddress := strings.TrimSpace(clientConn.ClientData().ServerAddress)
	if transferAddress == "" {
		transferAddress = localAddress
	}
	first := make(chan error, 2)
	go func() {
		first <- r.forward(clientConn, serverConn, failures, live, &shuttingDown, transferAddress, capture.DirectionClientToServer, "downstream", state.downstreamID, state.hop, onTransfer)
	}()
	go func() {
		first <- r.forward(serverConn, clientConn, failures, live, &shuttingDown, transferAddress, capture.DirectionServerToClient, "upstream", state.upstreamID, state.hop, onTransfer)
	}()
	bridgeErr := <-first
	var transfer *transferError
	if r.config.FollowTransfers && errors.As(bridgeErr, &transfer) {
		keepClientOpen = true
		shuttingDown.Store(true)
		_ = serverConn.Close()
		secondErr := <-first
		if err := failures.Err(); err != nil {
			return fmt.Errorf("capture hook failed: %w", err)
		}
		if err := r.recordAtHop("session.close", capture.SeverityInfo, map[string]any{
			"first_loop_error": errorString(bridgeErr), "second_loop_error": errorString(secondErr), "transfer_target": transfer.target,
		}, state.hop); err != nil {
			return err
		}
		live.Info("Session hop %d ended after transfer - target %s", state.hop, transfer.target)
		return transfer
	}
	shuttingDown.Store(true)
	_ = clientConn.Close()
	_ = serverConn.Close()
	secondErr := <-first
	if err := failures.Err(); err != nil {
		return fmt.Errorf("capture hook failed: %w", err)
	}
	if err := r.recordAtHop("session.close", severityForError(bridgeErr), map[string]any{
		"first_loop_error": errorString(bridgeErr), "second_loop_error": errorString(secondErr),
	}, state.hop); err != nil {
		return err
	}
	live.Info("Session ended - first loop: %s; second loop: %s (hop %d)", errorString(bridgeErr), errorString(secondErr), state.hop)
	return nil
}

func (r *Runner) probeTransferRoute(ctx context.Context, target string, localAddr *net.UDPAddr, clientGUID int64, hop int, logger *slog.Logger, live *liveReporter) error {
	probeCtx, cancel := context.WithTimeout(ctx, transferRouteProbeTimeout)
	defer cancel()
	transport := minecraft.NewRakNet(logger)
	transport.LocalAddr = localAddr
	transport.ClientGUID = clientGUID
	_, err := transport.PingContext(probeCtx, target)
	fields := map[string]any{
		"address":               target,
		"local_address":         localAddr.String(),
		"timeout":               transferRouteProbeTimeout.String(),
		"client_guid_preserved": clientGUID != 0,
	}
	if err != nil {
		fields["error"] = err.Error()
		fields["type"] = fmt.Sprintf("%T", err)
		_ = r.recordAtHop("transfer.route_probe", capture.SeverityWarn, fields, hop)
		return err
	}
	_ = r.recordAtHop("transfer.route_probe", capture.SeverityInfo, fields, hop)
	live.Info("Transfer route probe completed - %s (local %s)", target, localAddr)
	return nil
}

func (r *Runner) connectUpstream(ctx context.Context, target string, network bedrock.Network, observer *bedrock.Observer, connectionID string, hop int, logger *slog.Logger, live *liveReporter, clientData login.ClientData, transferRoute, routeProbed bool) (*minecraft.Conn, []*resource.Pack, error) {
	if err := r.recordAtHop("upstream.dial_start", capture.SeverityInfo, map[string]any{"address": target}, hop); err != nil {
		return nil, nil, err
	}
	live.Info("Connecting upstream - %s (hop %d) - waiting for upstream login and resource-pack exchange (buffered by design)", target, hop)
	dialCtx, cancelDial := context.WithTimeout(ctx, upstreamDialTimeout)
	defer cancelDial()
	urlResourcePacks := newURLResourcePackCache(dialCtx)
	recordPacket := observer.PacketFunc("upstream", connectionID)
	dialer := minecraft.Dialer{
		TokenSource: r.config.TokenSource,
		ClientData:  clientData,
		ErrorLog:    logger,
		// A pre-probe can happen several seconds before Minecraft reconnects.
		// Always perform a fresh bounded ping on the actual hop-2 dial so the
		// featured-experience route is active when RakNet starts its handshake.
		SkipPing: false,
		PingTimeout: func() time.Duration {
			if transferRoute {
				return transferRouteProbeTimeout
			}
			return 0
		}(),
		PacketFunc: func(header packet.Header, payload []byte, src, dst net.Addr) {
			recordPacket(header, payload, src, dst)
			urlResourcePacks.Observe(header, payload)
		},
		ResourcePackCache:          urlResourcePacks,
		DisconnectOnUnknownPackets: false,
		DisconnectOnInvalidPackets: false,
		DownloadResourcePack: func(_ uuid.UUID, _ string, _, _ int) bool {
			return true
		},
	}
	waitStarted := time.Now()
	conn, err := dialer.DialContextNetwork(dialCtx, network, target)
	waitDuration := time.Since(waitStarted).Round(time.Millisecond)
	if waitDuration < time.Millisecond {
		waitDuration = time.Millisecond
	}
	if err != nil {
		dialErr := err
		timedOut := errors.Is(dialCtx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded)
		message := "Upstream login and resource-pack exchange failed after %s"
		if timedOut {
			message = "Upstream login and resource-pack exchange timed out after %s"
			err = &upstreamDialTimeoutError{timeout: upstreamDialTimeout, err: err}
		}
		live.Info(message, waitDuration)
		fields := map[string]any{
			"error":         err.Error(),
			"type":          fmt.Sprintf("%T", dialErr),
			"timeout":       timedOut,
			"wait_duration": waitDuration.String(),
		}
		if timedOut {
			fields["timeout_duration"] = upstreamDialTimeout.String()
		}
		_ = r.recordAtHop("upstream.dial_error", capture.SeverityError, fields, hop)
		return nil, nil, fmt.Errorf("connect to upstream server: %w", err)
	}
	live.Info("Upstream login and resource-pack exchange completed in %s - downstream pack delivery can begin (sequential buffered mode)", waitDuration)
	// Dialer returns only after the upstream login and resource-pack exchange has
	// completed. Waiting on a second cache barrier here can deadlock the listener
	// login path because the URL offer is observed on the same decoder goroutine.
	// The connection's pack list is therefore authoritative at this point.
	resourcePacks := conn.ResourcePacks()
	if prefetched := urlResourcePacks.Packs(); len(prefetched) != 0 {
		resourcePacks = prefetched
	}
	if err := bedrock.RecordResourcePacks(ctx, r.config.Recorder, sessionID, connectionID, hop, conn.RemoteAddr(), conn.LocalAddr(), resourcePacks, bedrock.ResourcePackCaptureOptions{
		Decrypt: r.config.DecryptResourcePacks,
	}); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	if count := len(resourcePacks); count != 0 {
		derivation := "disabled; encrypted originals remain available"
		if r.config.DecryptResourcePacks {
			derivation = "enabled for supported encrypted archives"
		}
		live.Info("Resource packs retained - %d archives in %s; plaintext derivation %s", count, r.config.Recorder.Root(), derivation)
	}
	deliveryPacks, err := resourcePacksForDownstream(resourcePacks)
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("prepare resource packs for downstream delivery: %w", err)
	}
	if err := r.recordAtHop("upstream.connected", capture.SeverityInfo, map[string]any{
		"local_address":       conn.LocalAddr().String(),
		"remote_address":      conn.RemoteAddr().String(),
		"protocol_id":         conn.Proto().ID(),
		"game_version":        conn.Proto().Ver(),
		"resource_pack_count": len(resourcePacks),
	}, hop); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	live.Info("Upstream connected - %s, Bedrock %s protocol %d", conn.RemoteAddr(), conn.Proto().Ver(), conn.Proto().ID())
	return conn, deliveryPacks, nil
}

// resourcePacksForDownstream prevents a server-advertised URL from being
// forwarded to the real Bedrock client. The client must receive the pack over
// the proxy's RakNet connection so that the listener can observe and preserve
// the exact upstream URL while serving the same archive as chunk data.
func resourcePacksForDownstream(packs []*resource.Pack) ([]*resource.Pack, error) {
	if len(packs) == 0 {
		return nil, nil
	}
	delivery := make([]*resource.Pack, 0, len(packs))
	for _, pack := range packs {
		if pack == nil {
			return nil, errors.New("resource pack is nil")
		}
		if pack.DownloadURL() == "" {
			delivery = append(delivery, pack)
			continue
		}
		clone, err := resource.Read(io.NewSectionReader(pack, 0, int64(pack.Len())))
		if err != nil {
			return nil, fmt.Errorf("clone URL resource pack %s: %w", pack.UUID(), err)
		}
		if pack.Encrypted() {
			clone = clone.WithContentKey(pack.ContentKey())
		}
		delivery = append(delivery, clone)
	}
	return delivery, nil
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

func (r *Runner) forward(source, destination *minecraft.Conn, failures *bedrock.FailureSink, live *liveReporter, shuttingDown *atomic.Bool, localAddress string, direction capture.Direction, sourceChannel, connectionID string, hop int, onTransfer func(string)) error {
	for {
		readStarted := time.Now()
		if err := failures.Err(); err != nil {
			return err
		}
		decoded, err := source.ReadPacket()
		readDuration := time.Since(readStarted)
		if err != nil {
			return r.finishForwardAtHop("bridge.read_error", direction, sourceChannel, connectionID, hop, "read packet", err, shuttingDown)
		}
		live.Packet(direction, decoded)
		recordStarted := time.Now()
		decodedEvent, err := r.recordDecoded(decoded, direction, connectionID, hop)
		recordDuration := time.Since(recordStarted)
		if err != nil {
			return err
		}
		outgoing := decoded
		if r.config.FollowTransfers && direction == capture.DirectionServerToClient {
			if transfer, ok := decoded.(*packet.Transfer); ok {
				rewritten, target, rewriteErr := transferForProxyAddress(transfer, localAddress)
				if rewriteErr != nil {
					captureErr := r.finishForwardAtHop("bridge.transfer_rewrite_error", direction, sourceChannel, connectionID, hop, "rewrite Transfer", rewriteErr, shuttingDown)
					return errors.Join(&transferRewriteError{err: rewriteErr}, captureErr)
				}
				outgoing = rewritten
				if err := r.recordTransferRewrite(transfer, rewritten, target, connectionID, hop); err != nil {
					return err
				}
			}
		}
		writeStarted := time.Now()
		if err := destination.WritePacket(outgoing); err != nil {
			return r.finishForwardAtHop("bridge.write_error", direction, sourceChannel, connectionID, hop, "write packet", err, shuttingDown)
		}
		writeDuration := time.Since(writeStarted)
		flushMode := "automatic"
		var flushDuration time.Duration
		if outgoing != decoded {
			flushMode = "explicit"
			flushStarted := time.Now()
			if err := destination.Flush(); err != nil {
				return r.finishForwardAtHop("bridge.flush_error", direction, sourceChannel, connectionID, hop, "flush transfer", err, shuttingDown)
			}
			flushDuration = time.Since(flushStarted)
		}
		if shouldRecordForwardTiming(decoded) && decodedEvent.Sequence != 0 {
			if err := r.recordForwardTiming(decodedEvent, direction, connectionID, hop, readDuration, recordDuration, writeDuration, flushDuration, flushMode); err != nil {
				return err
			}
		}
		if outgoing != decoded {
			if transfer, ok := decoded.(*packet.Transfer); ok {
				_, target, _ := transferForProxyAddress(transfer, localAddress)
				if onTransfer != nil {
					onTransfer(target)
				}
				clientGUID, _ := source.RakNetClientGUID()
				return &transferError{target: target, localAddr: udpAddress(source.LocalAddr()), clientGUID: clientGUID}
			}
		}
	}
}

func (r *Runner) finishForward(kind string, direction capture.Direction, sourceChannel, connectionID, operation string, operationErr error, shuttingDown *atomic.Bool) error {
	return r.finishForwardAtHop(kind, direction, sourceChannel, connectionID, 1, operation, operationErr, shuttingDown)
}

func (r *Runner) finishForwardAtHop(kind string, direction capture.Direction, sourceChannel, connectionID string, hop int, operation string, operationErr error, shuttingDown *atomic.Bool) error {
	if !expectedShutdownError(operationErr, shuttingDown) {
		captureErr := r.recordNetworkErrorAtHop(kind, direction, sourceChannel, connectionID, hop, operation, operationErr)
		return errors.Join(operationErr, captureErr)
	}
	return operationErr
}

func expectedShutdownError(err error, shuttingDown *atomic.Bool) bool {
	if err == nil || shuttingDown == nil || !shuttingDown.Load() {
		return false
	}
	return errors.Is(err, context.Canceled) || errors.Is(err, net.ErrClosed)
}

func (r *Runner) recordDecoded(decoded packet.Packet, direction capture.Direction, connectionID string, hop int) (capture.Event, error) {
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
			Hop:          hop,
			Kind:         "packet.decode_error",
			Severity:     capture.SeverityError,
			Direction:    direction,
			Channel:      "bridge",
			Stage:        "decoded_latest_protocol",
			Packet:       &capture.PacketInfo{ID: decoded.ID(), Name: name, GoType: goType, DecodeStatus: "view_error"},
			Error:        &capture.ErrorInfo{Operation: "encode decoded view", Message: encodeErr.Error(), Type: fmt.Sprintf("%T", encodeErr)},
		}})
		return capture.Event{}, recordErr
	}
	annotations := make([]string, 0, 1)
	if _, transfer := decoded.(*packet.Transfer); transfer {
		if r.config.FollowTransfers {
			annotations = append(annotations, "Transfer target is rewritten to the local listener and followed because --follow-transfers is enabled")
		} else {
			annotations = append(annotations, "Transfer is recorded but automatic hop following is disabled")
		}
	}
	event, err := r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		SessionID:    sessionID,
		ConnectionID: connectionID,
		Hop:          hop,
		Kind:         kind,
		Severity:     capture.SeverityInfo,
		Direction:    direction,
		Channel:      "bridge",
		Stage:        "decoded_latest_protocol",
		Packet:       &capture.PacketInfo{ID: decoded.ID(), Name: name, GoType: goType, DecodeStatus: status},
		Data:         data,
		Annotations:  annotations,
	}})
	return event, err
}

// shouldRecordForwardTiming selects packets whose ordering and delivery path
// are useful when diagnosing world collision, spawn, and transfer behavior.
// Ordinary movement input is already retained as decoded evidence and is not
// repeated here to avoid doubling the event volume of a busy session.
func shouldRecordForwardTiming(decoded packet.Packet) bool {
	name, _ := packetNames(decoded)
	switch name {
	case "LevelChunk", "UpdateBlock", "UpdateSubChunkBlocks", "BlockActorData",
		"NetworkChunkPublisherUpdate", "ChunkRadiusUpdated", "MovePlayer", "SetActorMotion", "Transfer":
		return true
	default:
		return false
	}
}

type forwardTimingData struct {
	ReadPacketDurationNano    int64  `json:"read_packet_duration_nano,string"`
	RecordDecodedDurationNano int64  `json:"record_decoded_duration_nano,string"`
	WritePacketDurationNano   int64  `json:"write_packet_duration_nano,string"`
	FlushMode                 string `json:"flush_mode"`
	FlushDurationNano         int64  `json:"flush_duration_nano,string"`
	Measurement               string `json:"measurement"`
}

func (r *Runner) recordForwardTiming(decodedEvent capture.Event, direction capture.Direction, connectionID string, hop int, readDuration, recordDuration, writeDuration, flushDuration time.Duration, flushMode string) error {
	name := ""
	if decodedEvent.Packet != nil {
		name = decodedEvent.Packet.Name
	}
	data, err := json.Marshal(forwardTimingData{
		ReadPacketDurationNano:    readDuration.Nanoseconds(),
		RecordDecodedDurationNano: recordDuration.Nanoseconds(),
		WritePacketDurationNano:   writeDuration.Nanoseconds(),
		FlushMode:                 flushMode,
		FlushDurationNano:         flushDuration.Nanoseconds(),
		Measurement:               "bridge operation durations; automatic flush completion is represented by subsequent transport.payload events",
	})
	if err != nil {
		return fmt.Errorf("encode forward timing for %s: %w", name, err)
	}
	_, err = r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		SessionID:    sessionID,
		ConnectionID: connectionID,
		Hop:          hop,
		Kind:         "bridge.forward_timing",
		Severity:     capture.SeverityDebug,
		Direction:    direction,
		Channel:      "bridge",
		Stage:        "post_forward_write",
		ParentSequence: func() *uint64 {
			sequence := decodedEvent.Sequence
			return &sequence
		}(),
		Packet: decodedEvent.Packet,
		Data:   data,
	}})
	return err
}

func (r *Runner) recordTransferRewrite(original, rewritten *packet.Transfer, target, connectionID string, hop int) error {
	data, err := r.structuredView(map[string]any{
		"original":        original,
		"rewritten":       rewritten,
		"upstream_target": target,
		"listener_target": net.JoinHostPort(rewritten.Address, fmt.Sprintf("%d", rewritten.Port)),
	})
	if err != nil {
		return fmt.Errorf("encode Transfer rewrite view: %w", err)
	}
	_, err = r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		SessionID: sessionID, ConnectionID: connectionID, Hop: hop,
		Kind: "packet.transfer_rewrite", Severity: capture.SeverityInfo,
		Direction: capture.DirectionServerToClient, Channel: "bridge", Stage: "pre_forward_mutation",
		Packet:      &capture.PacketInfo{ID: rewritten.ID(), Name: "Transfer", GoType: reflect.TypeOf(rewritten).String(), DecodeStatus: "decoded"},
		Data:        data,
		Annotations: []string{"The original Transfer remains in packet.raw; this opt-in rewrite routes the client back to the local listener"},
	}})
	return err
}

func (r *Runner) recordStructuredView(kind string, value any, direction capture.Direction, connectionID, channel, stage string, annotations []string) error {
	return r.recordStructuredViewAtHop(kind, value, direction, connectionID, channel, stage, 1, annotations)
}

func (r *Runner) recordStructuredViewAtHop(kind string, value any, direction capture.Direction, connectionID, channel, stage string, hop int, annotations []string) error {
	data, encodeErr := r.structuredView(value)
	if encodeErr != nil {
		_, recordErr := r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
			SessionID: sessionID, ConnectionID: connectionID, Hop: hop,
			Kind: "capture.view_error", Severity: capture.SeverityError, Direction: direction,
			Channel: channel, Stage: stage,
			Error: &capture.ErrorInfo{Operation: "encode " + kind, Message: encodeErr.Error(), Type: fmt.Sprintf("%T", encodeErr)},
		}})
		return recordErr
	}
	_, err := r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		SessionID: sessionID, ConnectionID: connectionID, Hop: hop,
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

func (r *Runner) recordNetworkErrorAtHop(kind string, direction capture.Direction, channel, connectionID string, hop int, operation string, operationErr error) error {
	_, err := r.config.Recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		SessionID:    sessionID,
		ConnectionID: connectionID,
		Hop:          hop,
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
	return r.recordAtHop(kind, severity, value, 1)
}

func (r *Runner) recordAtHop(kind string, severity capture.Severity, value any, hop int) error {
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
		Hop:       hop,
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
