package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	// Live output is an operator aid only. Keep the canonical packet-by-packet
	// evidence in the capture while using a slower console cadence so that a
	// resource-pack exchange or a busy world does not flood the terminal.
	liveSummaryInterval = 3 * time.Second
	liveSummaryPerSide  = 4
	ansiReset           = "\x1b[0m"
	// PocketMine-MP's MainLogger and Terminal use semantic Minecraft colors
	// mapped to an xterm 256-color palette. Keep the same roles here while
	// retaining packet-direction colors for operator navigation.
	ansiCyan    = "\x1b[38;5;87m"  // #5FFFFF
	ansiGreen   = "\x1b[38;5;83m"  // #5FFF5F
	ansiYellow  = "\x1b[38;5;227m" // #FFFF5F
	ansiMagenta = "\x1b[38;5;207m" // #FF5FFF
	ansiDarkRed = "\x1b[38;5;124m" // #AF0000
	ansiWhite   = "\x1b[38;5;231m" // #FFFFFF
)

type livePacketKey struct {
	channel   string
	direction capture.Direction
	name      string
}

type livePacketCount struct {
	name  string
	count int
}

type liveReporter struct {
	mu              sync.Mutex
	output          io.Writer
	now             func() time.Time
	lastFlush       time.Time
	counts          map[livePacketKey]int
	hints           map[string]struct{}
	currentHop      int
	spawnedHops     map[int]struct{}
	shuttingDown    atomic.Bool
	color           bool
	clientRaw       map[uint32]string
	serverRaw       map[uint32]string
	followTransfers bool
}

func newLiveReporter(output io.Writer) *liveReporter {
	return newLiveReporterWithClock(output, time.Now)
}

func newLiveReporterWithClock(output io.Writer, now func() time.Time) *liveReporter {
	if output == nil {
		output = io.Discard
	}
	if now == nil {
		now = time.Now
	}
	return &liveReporter{
		output:      output,
		now:         now,
		lastFlush:   now(),
		counts:      make(map[livePacketKey]int),
		hints:       make(map[string]struct{}),
		currentHop:  1,
		spawnedHops: make(map[int]struct{}),
		color:       supportsColor(output),
		clientRaw:   rawPacketNames(true),
		serverRaw:   rawPacketNames(false),
	}
}

func (r *liveReporter) Info(format string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	r.flushLocked(now)
	r.writeLocked(now, "INFO", ansiWhite, fmt.Sprintf(format, args...))
}

func (r *liveReporter) RawPacket(channel string, direction capture.Direction, header packet.Header) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, spawned := r.spawnedHops[r.currentHop]; spawned {
		return
	}
	now := r.now()
	r.counts[livePacketKey{channel: channel, direction: direction, name: r.rawPacketName(direction, header.PacketID)}]++
	if now.Sub(r.lastFlush) >= liveSummaryInterval {
		r.flushLocked(now)
	}
}

func (r *liveReporter) SetSpawned() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.flushLocked(r.now())
	r.spawnedHops[r.currentHop] = struct{}{}
}

// SetHop flushes the current summary bucket before a transfer hop becomes
// active. Raw login and resource-pack packets are shown until that hop has
// spawned, even when an earlier hop already reached gameplay.
func (r *liveReporter) SetHop(hop int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.currentHop == hop {
		return
	}
	r.flushLocked(r.now())
	r.currentHop = hop
}

// SetShuttingDown marks the end of forwarding so late library messages caused
// by closing the transport can be presented as expected shutdown diagnostics.
// The capture log handler still records the original library level and text.
func (r *liveReporter) SetShuttingDown() {
	r.shuttingDown.Store(true)
}

func (r *liveReporter) SetFollowTransfers(enabled bool) {
	r.mu.Lock()
	r.followTransfers = enabled
	r.mu.Unlock()
}

func (r *liveReporter) Packet(direction capture.Direction, decoded packet.Packet) {
	if decoded == nil {
		return
	}
	name, _ := packetNames(decoded)
	now := r.now()

	r.mu.Lock()
	defer r.mu.Unlock()
	r.counts[livePacketKey{direction: direction, name: name}]++
	if transfer, ok := decoded.(*packet.Transfer); ok {
		r.flushLocked(now)
		message := "automatic hop following is disabled"
		if r.followTransfers {
			message = "hop following is enabled"
		}
		r.writeLocked(now, "TRANSFER", ansiMagenta, fmt.Sprintf(
			"Transfer observed %s - target %s:%d; %s",
			liveDirection(direction), transfer.Address, transfer.Port,
			message,
		))
		return
	}
	if now.Sub(r.lastFlush) >= liveSummaryInterval {
		r.flushLocked(now)
	}
}

func (r *liveReporter) LibraryLog(channel string, level slog.Level, message string) {
	if level < slog.LevelWarn {
		return
	}
	now := r.now()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.flushLocked(now)

	if channel == "downstream" && strings.Contains(message, `unexpected signature algorithm "ES384"; expected ["RS256"]`) {
		const hint = "self-signed-downstream-login"
		if _, exists := r.hints[hint]; exists {
			return
		}
		r.hints[hint] = struct{}{}
		r.writeLocked(now, "WARN", ansiYellow, "Downstream login rejected - the client used self-signed LAN authentication. On a trusted LAN, restart with --allow-unauthenticated-client; otherwise add the proxy in the Servers tab for Xbox-authenticated login")
		return
	}
	if r.shuttingDown.Load() && strings.Contains(strings.ToLower(message), "capture recorder is closed") {
		r.writeLocked(now, "WARN", ansiYellow, fmt.Sprintf("Library [%s] - %s", channel, message))
		return
	}
	color := libraryLogColor(level)
	r.writeLocked(now, strings.ToUpper(level.String()), color, fmt.Sprintf("Library [%s] - %s", channel, message))
}

func libraryLogColor(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return ansiDarkRed
	case level >= slog.LevelWarn:
		return ansiYellow
	default:
		return ansiWhite
	}
}

func (r *liveReporter) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.flushLocked(r.now())
}

func (r *liveReporter) flushLocked(now time.Time) {
	if len(r.counts) == 0 {
		r.lastFlush = now
		return
	}
	type liveGroup struct {
		channel   string
		direction capture.Direction
	}
	groups := make([]liveGroup, 0, 4)
	seen := make(map[liveGroup]struct{})
	for key := range r.counts {
		group := liveGroup{channel: key.channel, direction: key.direction}
		if _, ok := seen[group]; !ok {
			seen[group] = struct{}{}
			groups = append(groups, group)
		}
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].channel != groups[j].channel {
			return groups[i].channel < groups[j].channel
		}
		return groups[i].direction < groups[j].direction
	})
	for _, group := range groups {
		counts := make([]livePacketCount, 0)
		total := 0
		for key, count := range r.counts {
			if key.channel == group.channel && key.direction == group.direction {
				counts = append(counts, livePacketCount{name: key.name, count: count})
				total += count
			}
		}
		if len(counts) == 0 {
			continue
		}
		sort.Slice(counts, func(i, j int) bool {
			if counts[i].count != counts[j].count {
				return counts[i].count > counts[j].count
			}
			return counts[i].name < counts[j].name
		})
		visible := counts
		if len(visible) > liveSummaryPerSide {
			visible = visible[:liveSummaryPerSide]
		}
		parts := make([]string, 0, len(visible)+1)
		for _, count := range visible {
			parts = append(parts, fmt.Sprintf("%s x%d", count.name, count.count))
		}
		if len(counts) > len(visible) {
			otherPackets := 0
			for _, count := range counts[len(visible):] {
				otherPackets += count.count
			}
			parts = append(parts, fmt.Sprintf("other x%d (%d types)", otherPackets, len(counts)-len(visible)))
		}
		label := liveDirection(group.direction)
		if group.channel != "" {
			label = strings.ToUpper(group.channel) + " " + label
		}
		color := ansiCyan
		if group.direction == capture.DirectionServerToClient {
			color = ansiGreen
		}
		r.writeLocked(now, label, color, fmt.Sprintf("%d packets | %s", total, strings.Join(parts, " | ")))
	}
	for key := range r.counts {
		delete(r.counts, key)
	}
	r.lastFlush = now
}

func (r *liveReporter) writeLocked(now time.Time, label, color, message string) {
	timestamp := "[" + now.Format("15:04:05.000") + "]"
	label = fmt.Sprintf("%-17s", label)
	if r.color {
		timestamp = ansiCyan + timestamp + ansiReset
		label = color + label + ansiReset
	}
	_, _ = fmt.Fprintf(r.output, "%s %s %s\n", timestamp, label, message)
}

func supportsColor(output io.Writer) bool {
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return false
	}
	file, ok := output.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func rawPacketNames(listener bool) map[uint32]string {
	names := make(map[uint32]string)
	for id, factory := range minecraft.DefaultProtocol.Packets(listener) {
		name, _ := packetNames(factory())
		names[id] = name
	}
	return names
}

func (r *liveReporter) rawPacketName(direction capture.Direction, id uint32) string {
	names := r.serverRaw
	if direction == capture.DirectionClientToServer {
		names = r.clientRaw
	}
	if name, ok := names[id]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(0x%x)", id)
}

func liveDirection(direction capture.Direction) string {
	switch direction {
	case capture.DirectionClientToServer:
		return "C->S"
	case capture.DirectionServerToClient:
		return "S->C"
	default:
		return "?"
	}
}
