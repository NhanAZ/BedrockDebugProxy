package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	liveSummaryInterval = time.Second
	liveSummaryPerSide  = 6
)

type livePacketKey struct {
	direction capture.Direction
	name      string
}

type livePacketCount struct {
	name  string
	count int
}

type liveReporter struct {
	mu        sync.Mutex
	output    io.Writer
	now       func() time.Time
	lastFlush time.Time
	counts    map[livePacketKey]int
	hints     map[string]struct{}
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
		output:    output,
		now:       now,
		lastFlush: now(),
		counts:    make(map[livePacketKey]int),
		hints:     make(map[string]struct{}),
	}
}

func (r *liveReporter) Info(format string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	r.flushLocked(now)
	r.writeLocked(now, fmt.Sprintf(format, args...))
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
		r.writeLocked(now, fmt.Sprintf(
			"Transfer observed %s - target %s:%d; automatic hop following is not implemented",
			liveDirection(direction), transfer.Address, transfer.Port,
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
		r.writeLocked(now, "Downstream login rejected - the client used self-signed LAN authentication. On a trusted LAN, restart with --allow-unauthenticated-client; otherwise add the proxy in the Servers tab for Xbox-authenticated login")
		return
	}
	r.writeLocked(now, fmt.Sprintf("Library %s [%s] - %s", strings.ToLower(level.String()), channel, message))
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
	sections := make([]string, 0, 2)
	for _, direction := range []capture.Direction{capture.DirectionClientToServer, capture.DirectionServerToClient} {
		counts := make([]livePacketCount, 0)
		for key, count := range r.counts {
			if key.direction == direction {
				counts = append(counts, livePacketCount{name: key.name, count: count})
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
		sections = append(sections, liveDirection(direction)+" "+strings.Join(parts, ", "))
	}
	for key := range r.counts {
		delete(r.counts, key)
	}
	r.lastFlush = now
	if len(sections) != 0 {
		r.writeLocked(now, "Traffic - "+strings.Join(sections, " | "))
	}
}

func (r *liveReporter) writeLocked(now time.Time, message string) {
	_, _ = fmt.Fprintf(r.output, "[%s] %s\n", now.Format("15:04:05.000"), message)
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
