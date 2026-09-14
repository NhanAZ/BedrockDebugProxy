package proxy

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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
	current = current.Add(liveSummaryInterval)
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

func TestLiveReporterUsesThreeSecondSummaryBuckets(t *testing.T) {
	current := time.Date(2026, time.August, 30, 15, 0, 0, 0, time.UTC)
	var output bytes.Buffer
	reporter := newLiveReporterWithClock(&output, func() time.Time { return current })
	reporter.Packet(capture.DirectionClientToServer, &packet.PlayerAuthInput{})
	current = current.Add(2 * time.Second)
	reporter.Packet(capture.DirectionClientToServer, &packet.PlayerAuthInput{})
	if output.Len() != 0 {
		t.Fatalf("summary flushed before three-second interval: %s", output.String())
	}
	current = current.Add(time.Second)
	reporter.Packet(capture.DirectionClientToServer, &packet.PlayerAuthInput{})
	if !strings.Contains(output.String(), "C->S              3 packets | PlayerAuthInput x3") {
		t.Fatalf("summary was not flushed at three-second interval: %s", output.String())
	}
	reporter.Close()
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

func TestLiveReporterDowngradesCaptureCloseAfterShutdown(t *testing.T) {
	var output bytes.Buffer
	reporter := newLiveReporter(&output)
	message := "read batch: capture transport read: capture recorder is closed"
	reporter.SetShuttingDown()
	reporter.LibraryLog("upstream", slog.LevelError, message)

	text := output.String()
	if !strings.Contains(text, "WARN") || !strings.Contains(text, "Library [upstream] - "+message) {
		t.Fatalf("expected expected shutdown warning, got %q", text)
	}
	if strings.Contains(text, "ERROR             Library [upstream] - "+message) {
		t.Fatalf("expected capture-close message not to be reported as an error: %q", text)
	}
}

func TestLiveReporterKeepsCaptureCloseAsErrorBeforeShutdown(t *testing.T) {
	var output bytes.Buffer
	reporter := newLiveReporter(&output)
	message := "read batch: capture transport read: capture recorder is closed"
	reporter.LibraryLog("upstream", slog.LevelError, message)

	if !strings.Contains(output.String(), "ERROR             Library [upstream] - "+message) {
		t.Fatalf("expected pre-shutdown capture-close error, got %q", output.String())
	}
}

func TestLiveReporterUsesPocketMinePalette(t *testing.T) {
	tests := []struct {
		name  string
		label string
		color string
		emit  func(*liveReporter)
	}{
		{"info", "INFO", "\x1b[38;5;231m", func(r *liveReporter) { r.Info("Connected") }},
		{"client", "C->S", "\x1b[38;5;87m", func(r *liveReporter) {
			r.Packet(capture.DirectionClientToServer, &packet.PlayerAuthInput{})
		}},
		{"server", "S->C", "\x1b[38;5;83m", func(r *liveReporter) {
			r.Packet(capture.DirectionServerToClient, &packet.LevelChunk{})
		}},
		{"transfer", "TRANSFER", "\x1b[38;5;207m", func(r *liveReporter) {
			r.Packet(capture.DirectionServerToClient, &packet.Transfer{Address: "next.example.org", Port: 19132})
		}},
		{"warning", "WARN", "\x1b[38;5;227m", func(r *liveReporter) {
			r.LibraryLog("upstream", slog.LevelWarn, "Warning")
		}},
		{"error", "ERROR", "\x1b[38;5;124m", func(r *liveReporter) {
			r.LibraryLog("upstream", slog.LevelError, "Error")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			r := newLiveReporterWithClock(&output, func() time.Time {
				return time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)
			})
			r.color = true
			tt.emit(r)
			r.Close()

			prefix := "\x1b[38;5;87m[12:00:00.000]\x1b[0m " + tt.color + fmt.Sprintf("%-17s", tt.label) + "\x1b[0m "
			if !strings.Contains(output.String(), prefix) {
				t.Fatalf("missing Minecraft color prefix %q in %q", prefix, output.String())
			}
			if !strings.HasSuffix(output.String(), "\n") {
				t.Fatalf("output has no final newline: %q", output.String())
			}
		})
	}
}

func TestLiveReporterLeavesRedirectedOutputUncolored(t *testing.T) {
	var output bytes.Buffer
	r := newLiveReporter(&output)
	r.Info("Connected")
	r.Packet(capture.DirectionClientToServer, &packet.PlayerAuthInput{})
	r.LibraryLog("upstream", slog.LevelError, "Error")
	r.Close()
	if strings.Contains(output.String(), "\x1b") {
		t.Fatalf("redirected output contains terminal escapes: %q", output.String())
	}

	file, err := os.Create(filepath.Join(t.TempDir(), "console.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	if supportsColor(file) {
		t.Fatal("regular file output must not enable colors")
	}
}

func TestLiveReporterRespectsNoColor(t *testing.T) {
	for _, value := range []string{"", "1"} {
		t.Setenv("NO_COLOR", value)
		if supportsColor(os.Stdout) {
			t.Fatal("NO_COLOR must disable colors, including when present but empty")
		}
	}
}
