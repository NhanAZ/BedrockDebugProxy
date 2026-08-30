package proxy

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestLiveReporterSummarisesPacketsAndHighlightsTransfer(t *testing.T) {
	current := time.Date(2026, time.August, 30, 15, 0, 0, 0, time.UTC)
	var output bytes.Buffer
	reporter := newLiveReporterWithClock(&output, func() time.Time { return current })
	reporter.Packet(capture.DirectionClientToServer, &packet.PlayerAuthInput{})
	reporter.Packet(capture.DirectionClientToServer, &packet.PlayerAuthInput{})
	current = current.Add(time.Second)
	reporter.Packet(capture.DirectionServerToClient, &packet.LevelChunk{})
	reporter.Packet(capture.DirectionServerToClient, &packet.Transfer{Address: "next.example.org", Port: 19133})
	reporter.Close()

	text := output.String()
	for _, expected := range []string{
		"C->S              2 packets | PlayerAuthInput x2",
		"S->C              1 packets | LevelChunk x1",
		"Transfer observed S->C - target next.example.org:19133",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("output does not contain %q: %s", expected, text)
		}
	}
}

func TestLiveReporterSummarisesPreSpawnRawPacketsByChannel(t *testing.T) {
	current := time.Date(2026, time.August, 30, 15, 0, 0, 0, time.UTC)
	var output bytes.Buffer
	reporter := newLiveReporterWithClock(&output, func() time.Time { return current })
	reporter.RawPacket("upstream", capture.DirectionServerToClient, packet.Header{PacketID: packet.IDResourcePacksInfo})
	reporter.RawPacket("downstream", capture.DirectionClientToServer, packet.Header{PacketID: packet.IDResourcePackClientResponse})
	reporter.SetSpawned()
	reporter.RawPacket("upstream", capture.DirectionServerToClient, packet.Header{PacketID: packet.IDLevelChunk})
	reporter.Close()

	text := output.String()
	for _, expected := range []string{
		"DOWNSTREAM C->S   1 packets | ResourcePackClientResponse x1",
		"UPSTREAM S->C     1 packets | ResourcePacksInfo x1",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("output does not contain %q: %s", expected, text)
		}
	}
	if strings.Contains(text, "LevelChunk") {
		t.Fatalf("post-spawn raw packet was reported: %s", text)
	}
}

func TestLiveReporterExplainsSelfSignedLANLoginOnce(t *testing.T) {
	var output bytes.Buffer
	reporter := newLiveReporter(&output)
	message := `handle *packet.Login: verify ID token: unexpected signature algorithm "ES384"; expected ["RS256"]`
	reporter.LibraryLog("downstream", slog.LevelError, message)
	reporter.LibraryLog("downstream", slog.LevelError, message)

	text := output.String()
	if strings.Count(text, "--allow-unauthenticated-client") != 1 {
		t.Fatalf("output = %s", text)
	}
}
