// Package golemnet provides an agent-controlled Bedrock network client.
//
// It is intentionally separate from the server-side Golem mailbox package.
// The client connects through gophertunnel, records bounded structured events,
// and exposes only scenario actions needed by the validation workflow.
package golemnet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/experience"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/oauth2"
)

const (
	Schema          = "bedrockdebugproxy.golem.network.v1"
	maxEventBuffer  = 1024
	maxHistory      = 4096
	maxFormBytes    = 1 << 20
	maxChatBytes    = 64 << 10
	maxFormButtons  = 1024
	maxTransferHops = 64

	defaultClientInputInterval = 50 * time.Millisecond
	minClientInputInterval     = 20 * time.Millisecond
	maxClientInputInterval     = time.Second
)

var (
	ErrClosed        = errors.New("golem network client is closed")
	ErrEventBuffer   = errors.New("golem network event buffer is full")
	ErrNoCurrentForm = errors.New("no form is currently available")
)

// Config controls one agent-controlled network client session.
type Config struct {
	Address        string
	Transport      string
	Network        minecraft.Network
	TokenSource    oauth2.TokenSource
	Identity       login.IdentityData
	Client         login.ClientData
	HTTPClient     *http.Client
	Logger         *slog.Logger
	ConnectTimeout time.Duration

	// FollowTransfers reconnects when the server sends a Transfer packet.
	// The client preserves the RakNet source address and client GUID when the
	// selected transport supports those properties.
	FollowTransfers bool
	MaxHops         int

	// ClientInputInterval controls the stationary PlayerAuthInput heartbeat
	// after spawn. The default is one packet per Bedrock tick.
	ClientInputInterval time.Duration

	// ConnectionObserver is called after a connection has completed its login
	// setup. It is intended for evidence collection and must not mutate the
	// connection protocol state.
	ConnectionObserver func(*minecraft.Conn, int)
	// PacketObserver receives the post-transport packet payload for every packet
	// read from or written to a connection, including login and resource-pack
	// negotiation packets. Implementations must copy data they retain.
	PacketObserver func(*minecraft.Conn, int, packet.Header, []byte, net.Addr, net.Addr)
	// ResourcePackObserver runs after the connection's resource-pack exchange
	// has completed and before world spawn is published.
	ResourcePackObserver func(context.Context, *minecraft.Conn, int) error
	// SpawnObserver runs after DoSpawnContext has completed for a hop.
	SpawnObserver func(context.Context, *minecraft.Conn, int) error
}

// FormButton is a bounded projection of a server-sent menu form button.
type FormButton struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// Form is a server-sent form with its raw JSON retained for scenario-specific
// inspection. Raw form data is local validation evidence and must not be
// published without review.
type Form struct {
	ID           uint32          `json:"id"`
	Type         string          `json:"type,omitempty"`
	Title        string          `json:"title,omitempty"`
	Text         string          `json:"text,omitempty"`
	Buttons      []FormButton    `json:"buttons,omitempty"`
	Raw          json.RawMessage `json:"raw"`
	DecodeStatus string          `json:"decode_status"`
	DecodeError  string          `json:"decode_error,omitempty"`
	Hop          int             `json:"hop"`
	ReceivedAt   time.Time       `json:"received_at"`
}

// Transfer describes a server-requested next hop.
type Transfer struct {
	Address     string `json:"address"`
	Port        uint16 `json:"port"`
	Target      string `json:"target"`
	ReloadWorld bool   `json:"reload_world"`
	Hop         int    `json:"hop"`
	Followed    bool   `json:"followed"`
	FollowError string `json:"follow_error,omitempty"`
}

// Event is the bounded machine-readable stream observed by a scenario.
type Event struct {
	Sequence uint64    `json:"sequence"`
	Kind     string    `json:"kind"`
	Packet   string    `json:"packet,omitempty"`
	Message  string    `json:"message,omitempty"`
	Form     *Form     `json:"form,omitempty"`
	Transfer *Transfer `json:"transfer,omitempty"`
	Hop      int       `json:"hop"`
	At       time.Time `json:"at"`
}

// Report is a bounded session report. It is not a release validation report.
// Public-server release tooling still requires a verified capture and the
// existing validation report schema.
type Report struct {
	Schema             string         `json:"schema"`
	StartedAt          time.Time      `json:"started_at"`
	EndedAt            time.Time      `json:"ended_at,omitempty"`
	Hops               int            `json:"hops"`
	Spawned            bool           `json:"spawned"`
	Closed             bool           `json:"closed"`
	Authenticated      bool           `json:"authenticated"`
	ResourcePacks      int            `json:"resource_pack_count"`
	Forms              []Form         `json:"forms,omitempty"`
	Transfers          []Transfer     `json:"transfers,omitempty"`
	PacketCounts       map[string]int `json:"packet_counts,omitempty"`
	Messages           []string       `json:"messages,omitempty"`
	Errors             []string       `json:"errors,omitempty"`
	EventCount         int            `json:"event_count"`
	FollowTransfers    bool           `json:"follow_transfers"`
	ClientInput        bool           `json:"client_input_enabled"`
	ClientInputRate    string         `json:"client_input_interval,omitempty"`
	ClientInputPackets int            `json:"client_input_packet_count"`
	LastClientInputAt  time.Time      `json:"last_client_input_at,omitempty"`
}

// Client is a single network bot session.
type Client struct {
	config Config

	ctx    context.Context
	cancel context.CancelFunc

	connMu sync.RWMutex
	conn   *minecraft.Conn
	hop    int

	inputMu     sync.Mutex
	inputPaused bool

	events   chan Event
	done     chan struct{}
	doneOnce sync.Once

	stateMu            sync.Mutex
	history            []Event
	sequence           uint64
	currentForm        *Form
	started            time.Time
	ended              time.Time
	spawned            bool
	closed             bool
	readerStarted      bool
	terminalErr        error
	closeOnce          sync.Once
	packetCount        map[string]int
	resourcePacks      int
	inputTick          uint64
	clientInputPackets int
	lastClientInputAt  time.Time
	forms              []Form
	transfers          []Transfer
	messages           []string
	errors             []string
	hops               int
}

// New validates a network bot configuration without opening a connection.
func New(config Config) (*Client, error) {
	if strings.TrimSpace(config.Address) == "" {
		return nil, errors.New("golem network server address is empty")
	}
	if strings.ContainsAny(config.Address, "\r\n") {
		return nil, errors.New("golem network server address contains a line break")
	}
	if config.Network == nil {
		if err := validateAddress(config.Address); err != nil {
			return nil, err
		}
	}
	if config.Transport == "" {
		config.Transport = "raknet"
	}
	if config.MaxHops == 0 {
		config.MaxHops = maxTransferHops
	}
	if config.MaxHops < 1 || config.MaxHops > maxTransferHops {
		return nil, fmt.Errorf("golem network maximum hops must be between 1 and %d", maxTransferHops)
	}
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 15 * time.Minute
	}
	if config.ConnectTimeout <= 0 || config.ConnectTimeout > 30*time.Minute {
		return nil, errors.New("golem network connect timeout must be positive and at most 30m")
	}
	if config.ClientInputInterval == 0 {
		config.ClientInputInterval = defaultClientInputInterval
	}
	if config.ClientInputInterval < minClientInputInterval || config.ClientInputInterval > maxClientInputInterval {
		return nil, fmt.Errorf("golem network client input interval must be between %s and %s", minClientInputInterval, maxClientInputInterval)
	}
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}
	if config.Logger == nil {
		config.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Client{config: config, events: make(chan Event, maxEventBuffer), done: make(chan struct{}), packetCount: make(map[string]int)}, nil
}

// Start dials, completes resource-pack negotiation, spawns into the world,
// and starts the event reader.
func (c *Client) Start(parent context.Context) error {
	if parent == nil {
		parent = context.Background()
	}
	c.ctx, c.cancel = context.WithCancel(parent)
	connectCtx, cancelConnect := context.WithTimeout(c.ctx, c.config.ConnectTimeout)
	defer cancelConnect()
	c.started = time.Now().UTC()
	conn, err := c.dial(connectCtx, c.config.Address, c.config.Network, c.config.Identity, c.config.Client, nil, 0, 1)
	if err != nil {
		c.fail(err)
		return fmt.Errorf("dial server: %w", err)
	}
	if err := conn.DoSpawnContext(connectCtx); err != nil {
		_ = conn.CloseImmediately()
		c.fail(err)
		return fmt.Errorf("spawn in server: %w", err)
	}
	if c.config.ResourcePackObserver != nil {
		if err := c.config.ResourcePackObserver(connectCtx, conn, 1); err != nil {
			_ = conn.CloseImmediately()
			c.fail(err)
			return fmt.Errorf("observe resource packs: %w", err)
		}
	}
	c.setConnection(conn, 1)
	c.stateMu.Lock()
	c.spawned = true
	c.hops = 1
	c.resourcePacks = len(conn.ResourcePacks())
	c.readerStarted = true
	c.stateMu.Unlock()
	c.record(Event{Kind: "session.spawned", Hop: 1, At: time.Now().UTC()})
	if c.config.SpawnObserver != nil {
		if err := c.config.SpawnObserver(connectCtx, conn, 1); err != nil {
			c.fail(err)
			return fmt.Errorf("observe spawn: %w", err)
		}
	}
	go c.readLoop()
	go c.inputLoop()
	return nil
}

func (c *Client) dial(ctx context.Context, address string, network minecraft.Network, identity login.IdentityData, client login.ClientData, localAddr *net.UDPAddr, clientGUID int64, hop int) (*minecraft.Conn, error) {
	var callbackConn *minecraft.Conn
	dialer := minecraft.Dialer{
		ErrorLog:                   c.config.Logger,
		HTTPClient:                 c.config.HTTPClient,
		TokenSource:                c.config.TokenSource,
		IdentityData:               identity,
		ClientData:                 client,
		RakNetLocalAddr:            localAddr,
		RakNetClientGUID:           clientGUID,
		PingTimeout:                5 * time.Second,
		DisconnectOnUnknownPackets: false,
		DisconnectOnInvalidPackets: false,
		DownloadResourcePack:       func(uuid.UUID, string, int, int) bool { return true },
		ConnectionFunc: func(conn *minecraft.Conn) {
			// The dialer invokes this before login starts. Keep the callback
			// non-blocking because it runs on the login path.
			callbackConn = conn
		},
		PacketFunc: func(header packet.Header, payload []byte, src, dst net.Addr) {
			if callbackConn != nil && c.config.PacketObserver != nil {
				c.config.PacketObserver(callbackConn, hop, header, payload, src, dst)
			}
		},
	}
	var conn *minecraft.Conn
	var err error
	if network == nil {
		conn, err = dialer.DialContext(ctx, "raknet", address)
	} else {
		conn, err = dialer.DialContextNetwork(ctx, network, address)
	}
	if err != nil {
		return nil, err
	}
	if callbackConn == nil {
		callbackConn = conn
	}
	if c.config.ConnectionObserver != nil {
		c.config.ConnectionObserver(callbackConn, hop)
	}
	return conn, nil
}

func (c *Client) setConnection(conn *minecraft.Conn, hop int) {
	c.connMu.Lock()
	c.conn = conn
	c.hop = hop
	c.connMu.Unlock()
}

func (c *Client) connection() (*minecraft.Conn, int) {
	c.connMu.RLock()
	conn, hop := c.conn, c.hop
	c.connMu.RUnlock()
	return conn, hop
}

func (c *Client) readLoop() {
	defer c.closeDone()
	for {
		conn, hop := c.connection()
		if conn == nil {
			c.fail(ErrClosed)
			return
		}
		value, err := conn.ReadPacket()
		if err != nil {
			if c.ctx != nil && c.ctx.Err() != nil {
				c.finish(nil)
			} else {
				c.fail(err)
			}
			return
		}
		event := c.eventForPacket(value, hop)
		if event.Kind == "" {
			continue
		}
		if !c.record(event) {
			return
		}
		if event.Transfer != nil && c.config.FollowTransfers {
			if err := c.followTransfer(event.Transfer, conn, hop); err != nil {
				c.fail(err)
				return
			}
		}
	}
}

func (c *Client) eventForPacket(value packet.Packet, hop int) Event {
	event := Event{Kind: "packet", Packet: packetName(value), Hop: hop, At: time.Now().UTC()}
	switch value := value.(type) {
	case *packet.ModalFormRequest:
		form := parseForm(value.FormID, value.FormData, hop, event.At)
		event.Kind = "form"
		event.Form = &form
	case *packet.Transfer:
		address := strings.TrimSpace(strings.Trim(value.Address, "[]"))
		target := net.JoinHostPort(address, strconv.Itoa(int(value.Port)))
		event.Kind = "transfer"
		event.Transfer = &Transfer{Address: address, Port: value.Port, Target: target, ReloadWorld: value.ReloadWorld, Hop: hop}
	case *packet.Text:
		event.Kind = "text"
		event.Message = value.Message
	case *packet.Disconnect:
		event.Kind = "disconnect"
		event.Message = value.Message
	}
	return event
}

func (c *Client) followTransfer(transfer *Transfer, previous *minecraft.Conn, hop int) error {
	c.inputMu.Lock()
	defer c.inputMu.Unlock()
	c.inputPaused = true
	if transfer == nil {
		return errors.New("transfer event has no target")
	}
	if hop >= c.config.MaxHops {
		transfer.FollowError = fmt.Sprintf("transfer hop limit %d reached", c.config.MaxHops)
		c.updateTransfer(transfer)
		return fmt.Errorf("transfer hop limit %d reached", c.config.MaxHops)
	}
	if err := validateAddress(transfer.Target); err != nil {
		transfer.FollowError = err.Error()
		c.updateTransfer(transfer)
		return fmt.Errorf("validate transfer target: %w", err)
	}
	identity := previous.IdentityData()
	client := previous.ClientData()
	localAddr, _ := udpAddress(previous.LocalAddr())
	clientGUID, _ := previous.RakNetClientGUID()
	_ = previous.CloseImmediately()

	network := c.config.Network
	address := transfer.Target
	if strings.Contains(c.config.Transport, "nethernet") {
		if netherNetwork, httpsAddress, err := transferHTTPSNetherNet(c.ctx, c.config.HTTPClient, transfer.Target); err == nil {
			network = netherNetwork
			address = httpsAddress
		}
	}
	if network == nil || address == transfer.Target {
		raknet := minecraft.NewRakNet(c.config.Logger)
		raknet.LocalAddr = localAddr
		raknet.ClientGUID = clientGUID
		network = raknet
	}
	nextHop := hop + 1
	conn, err := c.dial(c.ctx, address, network, identity, client, localAddr, clientGUID, nextHop)
	if err != nil {
		transfer.FollowError = err.Error()
		c.updateTransfer(transfer)
		return fmt.Errorf("dial transfer target %s: %w", transfer.Target, err)
	}
	if err := conn.DoSpawnContext(c.ctx); err != nil {
		_ = conn.CloseImmediately()
		transfer.FollowError = err.Error()
		c.updateTransfer(transfer)
		return fmt.Errorf("spawn transfer target %s: %w", transfer.Target, err)
	}
	if c.config.ResourcePackObserver != nil {
		if err := c.config.ResourcePackObserver(c.ctx, conn, nextHop); err != nil {
			_ = conn.CloseImmediately()
			transfer.FollowError = err.Error()
			c.updateTransfer(transfer)
			return fmt.Errorf("observe resource packs for transfer target %s: %w", transfer.Target, err)
		}
	}
	transfer.Followed = true
	c.setConnection(conn, nextHop)
	c.stateMu.Lock()
	c.spawned = true
	c.hops = nextHop
	c.resourcePacks += len(conn.ResourcePacks())
	for index := len(c.transfers) - 1; index >= 0; index-- {
		if c.transfers[index].Target == transfer.Target && c.transfers[index].Hop == transfer.Hop && !c.transfers[index].Followed {
			c.transfers[index] = *transfer
			break
		}
	}
	c.stateMu.Unlock()
	c.inputPaused = false
	c.record(Event{Kind: "session.spawned", Hop: nextHop, At: time.Now().UTC()})
	if c.config.SpawnObserver != nil {
		if err := c.config.SpawnObserver(c.ctx, conn, nextHop); err != nil {
			return fmt.Errorf("observe spawn for transfer target %s: %w", transfer.Target, err)
		}
	}
	return nil
}

func (c *Client) inputLoop() {
	ticker := time.NewTicker(c.config.ClientInputInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := c.sendClientInput(); err != nil {
				if c.ctx != nil && c.ctx.Err() != nil {
					return
				}
				c.fail(err)
				return
			}
		case <-c.ctx.Done():
			return
		case <-c.done:
			return
		}
	}
}

func (c *Client) sendClientInput() error {
	c.inputMu.Lock()
	defer c.inputMu.Unlock()
	if c.inputPaused {
		return nil
	}
	conn, _ := c.connection()
	if conn == nil {
		return nil
	}
	data := conn.GameData()
	clientData := conn.ClientData()
	inputMode := clientData.CurrentInputMode
	if inputMode <= packet.InputModeUndefined || inputMode >= packet.InputModeCount {
		inputMode = packet.InputModeTouch
	}
	interactionModel := packet.InteractionModelTouch
	if inputMode == packet.InputModeMouse || inputMode == packet.InputModeGamePad {
		interactionModel = packet.InteractionModelCrosshair
	}
	tick := c.nextInputTick()
	input := &packet.PlayerAuthInput{
		Pitch:             data.Pitch,
		Yaw:               data.Yaw,
		Position:          data.PlayerPosition,
		HeadYaw:           data.Yaw,
		InputData:         protocol.NewInputFlags(packet.InputFlagCount),
		InputMode:         uint32(inputMode),
		PlayMode:          packet.PlayModeNormal,
		InteractionModel:  int32(interactionModel),
		InteractPitch:     data.Pitch,
		InteractYaw:       data.Yaw,
		Tick:              tick,
		CameraOrientation: cameraOrientation(data.Pitch, data.Yaw),
	}
	if err := conn.WritePacket(input); err != nil {
		return fmt.Errorf("write client input: %w", err)
	}
	if err := conn.Flush(); err != nil {
		return fmt.Errorf("flush client input: %w", err)
	}
	c.stateMu.Lock()
	c.clientInputPackets++
	c.lastClientInputAt = time.Now().UTC()
	c.stateMu.Unlock()
	return nil
}

func (c *Client) nextInputTick() uint64 {
	c.stateMu.Lock()
	c.inputTick++
	tick := c.inputTick
	c.stateMu.Unlock()
	return tick
}

func cameraOrientation(pitch, yaw float32) (orientation mgl32.Vec3) {
	pitchRadians := float64(pitch) * math.Pi / 180
	yawRadians := float64(yaw) * math.Pi / 180
	cosPitch := float32(math.Cos(pitchRadians))
	return mgl32.Vec3{
		float32(-math.Sin(yawRadians)) * cosPitch,
		float32(-math.Sin(pitchRadians)),
		float32(math.Cos(yawRadians)) * cosPitch,
	}
}

func (c *Client) updateTransfer(transfer *Transfer) {
	if transfer == nil {
		return
	}
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	for index := len(c.transfers) - 1; index >= 0; index-- {
		if c.transfers[index].Target == transfer.Target && c.transfers[index].Hop == transfer.Hop {
			c.transfers[index] = *transfer
			return
		}
	}
}

func transferHTTPSNetherNet(ctx context.Context, client *http.Client, target string) (minecraft.Network, string, error) {
	host, port, err := net.SplitHostPort(target)
	if err != nil || strings.TrimSpace(host) == "" {
		return nil, "", fmt.Errorf("split transfer target: %w", err)
	}
	address := "https://" + net.JoinHostPort(host, port)
	network, err := experience.NewHTTPSNetherNetNetwork(ctx, client, address)
	if err != nil {
		return nil, "", err
	}
	return network, address, nil
}

func udpAddress(address net.Addr) (*net.UDPAddr, bool) {
	if address == nil {
		return nil, false
	}
	if value, ok := address.(*net.UDPAddr); ok {
		copy := *value
		copy.IP = append(net.IP(nil), value.IP...)
		return &copy, true
	}
	value, err := net.ResolveUDPAddr("udp", address.String())
	return value, err == nil
}

func (c *Client) record(event Event) bool {
	c.stateMu.Lock()
	c.sequence++
	event.Sequence = c.sequence
	cloned := cloneEvent(event)
	if len(c.history) >= maxHistory {
		c.history = append(c.history[1:], cloned)
	} else {
		c.history = append(c.history, cloned)
	}
	if event.Packet != "" {
		c.packetCount[event.Packet]++
	}
	if event.Form != nil {
		c.forms = append(c.forms, cloneForm(*event.Form))
		c.currentForm = cloneFormPtr(event.Form)
	}
	if event.Transfer != nil {
		c.transfers = append(c.transfers, *event.Transfer)
	}
	if event.Message != "" && event.Kind == "text" {
		if len(c.messages) < maxHistory {
			c.messages = append(c.messages, event.Message)
		}
	}
	c.stateMu.Unlock()
	// Generic packet telemetry is retained in history and packet counts, but it
	// is not a scenario notification. Scenarios only need actionable events such
	// as forms, text, transfers, disconnects, and lifecycle changes. Keeping the
	// telemetry out of this bounded queue prevents a normal login burst from
	// starving the event a scenario is waiting for.
	if event.Kind == "packet" {
		return true
	}
	select {
	case c.events <- cloneEvent(event):
		return true
	default:
		c.fail(ErrEventBuffer)
		return false
	}
}

func (c *Client) fail(err error) {
	if err == nil {
		c.finish(nil)
		return
	}
	c.stateMu.Lock()
	if c.terminalErr == nil {
		c.terminalErr = err
		if len(c.errors) < maxHistory {
			c.errors = append(c.errors, err.Error())
		}
	}
	c.stateMu.Unlock()
	c.finish(err)
}

func (c *Client) finish(_ error) {
	c.closeOnce.Do(func() {
		if c.cancel != nil {
			c.cancel()
		}
		conn, _ := c.connection()
		if conn != nil {
			_ = conn.CloseImmediately()
		}
		c.stateMu.Lock()
		readerStarted := c.readerStarted
		c.closed = true
		c.ended = time.Now().UTC()
		c.stateMu.Unlock()
		if !readerStarted {
			c.closeDone()
		}
	})
}

func (c *Client) closeDone() {
	c.doneOnce.Do(func() { close(c.done) })
}

// Close stops the session and releases the active network connection.
func (c *Client) Close() error {
	c.finish(nil)
	select {
	case <-c.done:
	case <-time.After(5 * time.Second):
		return errors.New("close network client timed out")
	}
	return nil
}

// NextEvent waits for the next recorded event.
func (c *Client) NextEvent(ctx context.Context) (Event, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case event := <-c.events:
		return event, nil
	case <-ctx.Done():
		return Event{}, ctx.Err()
	case <-c.done:
		c.stateMu.Lock()
		err := c.terminalErr
		c.stateMu.Unlock()
		if err == nil {
			err = ErrClosed
		}
		return Event{}, err
	}
}

// WaitForForm waits until a form is observed or the context expires.
func (c *Client) WaitForForm(ctx context.Context) (Form, error) {
	for {
		event, err := c.NextEvent(ctx)
		if err != nil {
			return Form{}, err
		}
		if event.Form != nil {
			return cloneForm(*event.Form), nil
		}
	}
}

// CurrentForm returns the most recently observed form.
func (c *Client) CurrentForm() (Form, bool) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	if c.currentForm == nil {
		return Form{}, false
	}
	return cloneForm(*c.currentForm), true
}

// SubmitForm sends a raw JSON form response for the current form.
func (c *Client) SubmitForm(ctx context.Context, response json.RawMessage) error {
	if len(response) == 0 {
		return errors.New("form response is required")
	}
	if len(response) > maxFormBytes || !json.Valid(response) {
		return errors.New("form response must be valid JSON within the size limit")
	}
	form, ok := c.CurrentForm()
	if !ok {
		return ErrNoCurrentForm
	}
	conn, _ := c.connection()
	if conn == nil {
		return ErrClosed
	}
	if err := conn.WritePacket(&packet.ModalFormResponse{FormID: form.ID, ResponseData: protocol.Option(append([]byte(nil), response...))}); err != nil {
		return fmt.Errorf("write form response: %w", err)
	}
	if err := conn.Flush(); err != nil {
		return fmt.Errorf("flush form response: %w", err)
	}
	_ = ctx
	return nil
}

// SelectForm submits a menu button index from the current form.
func (c *Client) SelectForm(ctx context.Context, index int) error {
	form, ok := c.CurrentForm()
	if !ok {
		return ErrNoCurrentForm
	}
	if index < 0 || index >= len(form.Buttons) {
		return fmt.Errorf("form button index %d is outside 0..%d", index, len(form.Buttons)-1)
	}
	return c.SubmitForm(ctx, json.RawMessage(strconv.Itoa(index)))
}

// Chat sends a client chat message.
func (c *Client) Chat(message string) error {
	if strings.TrimSpace(message) == "" {
		return errors.New("chat message is empty")
	}
	if len(message) > maxChatBytes {
		return fmt.Errorf("chat message exceeds %d bytes", maxChatBytes)
	}
	conn, _ := c.connection()
	if conn == nil {
		return ErrClosed
	}
	if err := conn.WritePacket(&packet.Text{TextType: packet.TextTypeChat, Message: message}); err != nil {
		return fmt.Errorf("write chat: %w", err)
	}
	return conn.Flush()
}

// Wait waits while the reader continues collecting server events.
func (c *Client) Wait(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return errors.New("wait duration must be positive")
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		c.stateMu.Lock()
		err := c.terminalErr
		c.stateMu.Unlock()
		if err == nil {
			err = ErrClosed
		}
		return err
	}
}

// Snapshot returns the currently retained report state.
func (c *Client) Snapshot() Report {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	ended := c.ended
	if ended.IsZero() {
		ended = time.Now().UTC()
	}
	return Report{
		Schema:             Schema,
		StartedAt:          c.started,
		EndedAt:            ended,
		Hops:               c.hops,
		Spawned:            c.spawned,
		Closed:             c.closed,
		Authenticated:      c.currentAuthenticated(),
		ResourcePacks:      c.resourcePacks,
		Forms:              cloneForms(c.forms),
		Transfers:          append([]Transfer(nil), c.transfers...),
		PacketCounts:       copyCounts(c.packetCount),
		Messages:           append([]string(nil), c.messages...),
		Errors:             append([]string(nil), c.errors...),
		EventCount:         len(c.history),
		FollowTransfers:    c.config.FollowTransfers,
		ClientInput:        c.config.ClientInputInterval > 0,
		ClientInputRate:    c.config.ClientInputInterval.String(),
		ClientInputPackets: c.clientInputPackets,
		LastClientInputAt:  c.lastClientInputAt,
	}
}

func (c *Client) currentAuthenticated() bool {
	conn, _ := c.connection()
	return conn != nil && conn.Authenticated()
}

func validateAddress(address string) error {
	host, port, err := net.SplitHostPort(strings.TrimSpace(address))
	if err != nil {
		return fmt.Errorf("invalid server address %q: %w", address, err)
	}
	if strings.TrimSpace(host) == "" {
		return errors.New("server address host is empty")
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return fmt.Errorf("server address port %q is invalid", port)
	}
	return nil
}

func packetName(value packet.Packet) string {
	if value == nil {
		return "<nil>"
	}
	name := fmt.Sprintf("%T", value)
	if index := strings.LastIndexByte(name, '.'); index >= 0 {
		name = name[index+1:]
	}
	return strings.TrimPrefix(name, "*")
}

func parseForm(id uint32, raw []byte, hop int, at time.Time) Form {
	form := Form{ID: id, Raw: append(json.RawMessage(nil), raw...), DecodeStatus: "invalid", Hop: hop, ReceivedAt: at}
	if len(raw) > maxFormBytes {
		form.DecodeError = "form data exceeds size limit"
		return form
	}
	var data struct {
		Type    string `json:"type"`
		Title   string `json:"title"`
		Text    string `json:"text"`
		Buttons []struct {
			Text string `json:"text"`
		} `json:"buttons"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		form.DecodeError = err.Error()
		return form
	}
	if len(data.Buttons) > maxFormButtons {
		form.DecodeError = fmt.Sprintf("form contains more than %d buttons", maxFormButtons)
		return form
	}
	form.Type, form.Title, form.Text = data.Type, data.Title, data.Text
	form.Buttons = make([]FormButton, len(data.Buttons))
	for index, button := range data.Buttons {
		form.Buttons[index] = FormButton{Index: index, Text: button.Text}
	}
	form.DecodeStatus = "decoded"
	return form
}

func cloneForm(value Form) Form {
	value.Raw = append(json.RawMessage(nil), value.Raw...)
	value.Buttons = append([]FormButton(nil), value.Buttons...)
	return value
}

func cloneFormPtr(value *Form) *Form {
	if value == nil {
		return nil
	}
	copy := cloneForm(*value)
	return &copy
}

func cloneForms(values []Form) []Form {
	result := make([]Form, len(values))
	for index, value := range values {
		result[index] = cloneForm(value)
	}
	return result
}

func cloneEvent(value Event) Event {
	value.Form = cloneFormPtr(value.Form)
	if value.Transfer != nil {
		transfer := *value.Transfer
		value.Transfer = &transfer
	}
	return value
}

func copyCounts(values map[string]int) map[string]int {
	result := make(map[string]int, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
