package viewer

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

func TestViewerServesFilteredEventsAndBlobPreview(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "viewer-test"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := recorder.Record(context.Background(), capture.Record{
		Event: capture.Event{
			Kind:      "packet.decoded",
			Direction: capture.DirectionServerToClient,
			Channel:   "bridge",
			Packet:    &capture.PacketInfo{ID: 58, Name: "LevelChunk", DecodeStatus: "decoded"},
			Data:      json.RawMessage(`{"dimension":0}`),
		},
		Raw:            []byte{0, 1, 2, 3, 4, 5},
		MediaType:      "application/octet-stream",
		Representation: "packet_payload",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{
		Kind: "packet.decode_error", Severity: capture.SeverityError, Direction: capture.DirectionClientToServer,
		Packet: &capture.PacketInfo{ID: 9, Name: "Text", DecodeStatus: "decode_error"},
		Error:  &capture.ErrorInfo{Operation: "decode", Message: "synthetic failure"},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}

	view, err := Open(root, Options{BlobPreviewBytes: 4})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(view.Handler("test-token"))
	defer server.Close()

	response := get(t, server.URL+"/test-token/")
	index, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK || !strings.Contains(string(index), "BedrockDebugProxy") {
		t.Fatalf("index status = %d, body = %s", response.StatusCode, index)
	}
	for _, marker := range []string{`href="assets/app.css"`, `src="assets/app.js"`, `href="assets/favicon-32x32.png"`, `href="assets/favicon.ico"`, `href="assets/apple-touch-icon.png"`, `href="assets/site.webmanifest"`, `<meta name="theme-color" content="#000000">`, `aria-label="Capture overview"`, `class="brand-name"`, `id="viewer-status"`, `id="event-count"`, `id="blob-count"`, `id="duration"`, `id="generator"`, `id="direction"`, `id="packet"`, `id="kind"`, `id="channel"`, `id="page-input"`, `id="page-label"`, `id="bytes-panel"`, `id="bytes-toggle"`, `aria-expanded="false"`, `id="hex-view" class="hex-view" hidden`, `class="panel-title"`, `class="panel-title-separator"`, "BedrockDebugProxy", "Ordered stream", "Decoded view", "Raw evidence"} {
		if !strings.Contains(string(index), marker) {
			t.Fatalf("index is missing viewer marker %q", marker)
		}
	}
	headerEnd := strings.Index(string(index), "</header>")
	overviewStart := strings.Index(string(index), `class="overview"`)
	if overviewStart < 0 || headerEnd < 0 || overviewStart > headerEnd {
		t.Fatal("capture overview is not rendered inside the compact header")
	}
	for _, forbidden := range []string{"BEDROCKDEBUGPROXY", "ORDERED STREAM", "DECODED VIEW", "RAW EVIDENCE", "Capture Viewer", "capture-status", "capture-id", "capture-identity", "status-pill", "brand-mark", "<footer", "footer-meta", "⌕"} {
		if strings.Contains(string(index), forbidden) {
			t.Fatalf("index still contains forced uppercase label %q", forbidden)
		}
	}

	var session sessionResponse
	decodeURL(t, server.URL+"/test-token/api/session", &session)
	if session.Manifest.CaptureID != "viewer-test" || session.Events != 2 || session.Blobs != 1 {
		t.Fatalf("session = %+v", session)
	}

	var events eventsResponse
	decodeURL(t, server.URL+"/test-token/api/events?packet=Text&problems=true", &events)
	if events.Total != 1 || len(events.Rows) != 1 || events.Rows[0].Packet == nil || events.Rows[0].Packet.Name != "Text" {
		t.Fatalf("events = %+v", events)
	}

	var detail eventResponse
	decodeURL(t, server.URL+"/test-token/api/events/"+sequenceText(first.Sequence), &detail)
	if detail.Event.Sequence != first.Sequence || string(detail.Event.Data) != `{"dimension":0}` {
		t.Fatalf("detail event = %+v", detail.Event)
	}
	if string(detail.BlobPreview) != string([]byte{0, 1, 2, 3}) || !detail.Truncated {
		t.Fatalf("blob preview = %v, truncated = %v", detail.BlobPreview, detail.Truncated)
	}
}

func TestViewerRequiresTokenAndClosedVerifiedCapture(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "viewer-boundary"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root, Options{}); err == nil {
		t.Fatal("Open() succeeded for an open capture")
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	view, err := Open(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(view.Handler("secret"))
	defer server.Close()
	response := get(t, server.URL+"/api/session")
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestViewerServesEmbeddedFontAssets(t *testing.T) {
	server := httptest.NewServer((&Viewer{}).Handler("font-token"))
	defer server.Close()

	response := get(t, server.URL+"/font-token/assets/app.css")
	css, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	for _, source := range []string{"url(\"glyph_e031.svg\")", "url(\"glyph_v.svg\")"} {
		if !strings.Contains(string(css), source) {
			t.Fatalf("CSS does not use the supplied glyph asset %s", source)
		}
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("CSS status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	jsResponse := get(t, server.URL+"/font-token/assets/app.js")
	js, err := io.ReadAll(jsResponse.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = jsResponse.Body.Close()
	if jsResponse.StatusCode != http.StatusOK || !strings.Contains(string(js), "setupCustomSelects") {
		t.Fatalf("JavaScript asset status = %d or custom viewer setup is missing", jsResponse.StatusCode)
	}
	if !strings.Contains(string(css), `local("Minecraft Ten")`) {
		t.Fatal("CSS does not prefer the locally installed Minecraft Ten font")
	}
	if strings.Contains(string(css), `local("Mojangles")`) || strings.Contains(string(css), "font-heading") {
		t.Fatal("CSS still selects a separate heading font")
	}
	if !strings.Contains(string(css), "font-size: var(--font-size)") || !strings.Contains(string(css), "--font-size: 12px") {
		t.Fatal("CSS does not expose the shared 12px font-size contract")
	}
	for _, layout := range []string{"min-width: 0", "grid-template-columns: minmax(0, 1fr)", "grid-template-columns: auto minmax(0, 1fr)", "grid-template-rows: max-content max-content max-content", "grid-template-rows: minmax(360px, 65vh) max-content", "grid-template-rows: max-content max-content", "min-height: max-content", "overflow: auto", "overflow: visible", "th:nth-child(2) { width: 132px; }", ".event-status", ".search-field kbd { color: var(--subtle); font-family: var(--font-content);", "@media (max-width: 1100px)", "flex-wrap: wrap", "@media (max-width: 640px)", "grid-template-columns: repeat(2, minmax(0, 1fr))", "grid-template-rows: minmax(320px, 62vh) max-content", "table { min-width: 620px; }"} {
		if !strings.Contains(string(css), layout) {
			t.Fatalf("CSS does not contain the single-column workspace layout %q", layout)
		}
	}
	if strings.Contains(string(css), "minmax(500px, 1.08fr) minmax(420px, 0.92fr)") || strings.Contains(string(css), "minmax(430px, 1fr) minmax(360px, 0.8fr)") {
		t.Fatal("CSS still contains a two-column workspace layout")
	}
	if strings.Contains(string(css), "grid-template-rows: minmax(280px, 1.2fr) minmax(210px, 0.8fr)") || strings.Contains(string(css), ".detail-content { min-height: 0; padding: 14px; overflow: auto; }") || strings.Contains(string(css), ".inspector { min-width: 0; min-height: 0;") {
		t.Fatal("CSS still constrains the event details panel with an internal scrollbar")
	}
	if strings.Contains(string(css), "grid-template-rows: auto auto auto minmax(0, 1fr) auto") || strings.Contains(string(css), "  overflow: hidden;\n}\n\nbutton, input, select") {
		t.Fatal("CSS still keeps the page inside a workspace-only scroll container")
	}
	for _, source := range []string{".select-trigger", ".select-menu", ".select-option", "width: max-content", "max-width: min(480px, calc(100vw - 16px))", "scrollbar-width: thin", "scrollbar-color: var(--line) var(--surface-raised)", "html, body {\n  scrollbar-width: thin;\n  scrollbar-color: var(--line) var(--bg);\n}", "html::-webkit-scrollbar, body::-webkit-scrollbar { width: 4px; height: 4px; }", "html::-webkit-scrollbar-button, body::-webkit-scrollbar-button", "html::-webkit-scrollbar-thumb, body::-webkit-scrollbar-thumb", ".select-menu::-webkit-scrollbar { width: 4px; }", ".select-menu::-webkit-scrollbar-button", ".select-menu::-webkit-scrollbar-thumb", ".table-wrap::-webkit-scrollbar { width: 4px; height: 4px; }", ".table-wrap::-webkit-scrollbar-button", ".table-wrap::-webkit-scrollbar-thumb", ".bytes-panel.collapsed", ".panel-heading-toggle", ".panel-title {", ".panel-title-separator", ".brand-name"} {
		if !strings.Contains(string(css), source) {
			t.Fatalf("CSS does not contain themed dropdown selector %s", source)
		}
	}
	if strings.Contains(string(css), "scrollbar-width: none") || strings.Contains(string(css), ".table-wrap::-webkit-scrollbar { display: none; }") {
		t.Fatal("CSS still hides the event stream scrollbar")
	}
	for _, source := range []string{"--gold: #ffaa00", "tbody tr.problem.decode-pending", ".decode-pending-text"} {
		if !strings.Contains(string(css), source) {
			t.Fatalf("CSS does not contain pending-decode styling %s", source)
		}
	}
	allowedColors := map[string]bool{
		"#000000": true, "#00002a": true, "#002a00": true, "#002a2a": true,
		"#2a0000": true, "#2a002a": true, "#2a2a00": true, "#2a2a2a": true,
		"#151515": true, "#15153f": true, "#153f15": true, "#153f3f": true,
		"#3f1515": true, "#3f153f": true, "#3f3f15": true, "#3f3f3f": true,
		"#373501": true, "#383534": true, "#333232": true, "#110e0e": true,
		"#250501": true, "#2d1a13": true, "#372c0b": true, "#04280d": true,
		"#0b2e2a": true, "#08121e": true, "#261731": true, "#3b1d05": true,
		"#232d40": true, "#0000aa": true, "#00aa00": true, "#00aaaa": true,
		"#aa0000": true, "#aa00aa": true, "#ffaa00": true, "#aaaaaa": true,
		"#555555": true, "#5555ff": true, "#55ff55": true, "#55ffff": true,
		"#ff5555": true, "#ff55ff": true, "#ffff55": true, "#ffffff": true,
		"#ddd605": true, "#e3d4d1": true, "#cecaca": true, "#443a3b": true,
		"#971607": true, "#b4684d": true, "#deb12d": true, "#47a036": true,
		"#2cbaa8": true, "#21497b": true, "#9a5cc6": true, "#eb7114": true,
		"#8bb3ff": true,
	}
	for _, color := range regexp.MustCompile(`#[0-9a-fA-F]{6}`).FindAllString(string(css), -1) {
		if !allowedColors[strings.ToLower(color)] {
			t.Fatalf("CSS uses a color outside the Minecraft palette: %s", color)
		}
	}
	for _, forbidden := range []string{"linear-gradient", "radial-gradient", "rgb(", "rgba(", "hsl(", "hsla("} {
		if strings.Contains(string(css), forbidden) {
			t.Fatalf("CSS uses forbidden color function or gradient %q", forbidden)
		}
	}
	jsResponse = get(t, server.URL+"/font-token/assets/app.js")
	js, err = io.ReadAll(jsResponse.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = jsResponse.Body.Close()
	if !strings.Contains(string(js), "decode-pending") || !strings.Contains(string(js), "not_yet_decoded") || !strings.Contains(string(js), "setBytesExpanded") || !strings.Contains(string(js), "elements[\"hex-view\"].hidden") {
		t.Fatal("JavaScript does not classify pending-decode rows")
	}
	for _, source := range []string{"function goToPage", "elements[\"page-input\"].max", "event.key === \"Enter\""} {
		if !strings.Contains(string(js), source) {
			t.Fatalf("JavaScript does not support direct page entry %q", source)
		}
	}
	for _, forbidden := range []string{"font-weight: 650", "font-weight: 700", "font-weight: 750", "font-size: 9px", "font-size: 10px", "font-size: 11px", "font-size: 13px", "font-size: 14px", "font-size: 15px", "font-size: 16px", "font-size: 17px", "font-size: 18px", "font: 11px/1.55", "text-transform: uppercase"} {
		if strings.Contains(string(css), forbidden) {
			t.Fatalf("CSS still contains non-uniform typography declaration %q", forbidden)
		}
	}

	tests := []struct {
		path        string
		contentType string
		prefix      string
		sha256      string
	}{
		{path: "glyph_e031.svg", contentType: "image/svg+xml", prefix: "<?xml"},
		{path: "glyph_v.svg", contentType: "image/svg+xml", prefix: "<?xml"},
		{path: "favicon-16x16.png", contentType: "image/png", prefix: "\x89PNG\r\n\x1a\n", sha256: "221dc39a54736e4858934b808210f8cdec08b2eeb86f73b919705aab48166fcc"},
		{path: "favicon-32x32.png", contentType: "image/png", prefix: "\x89PNG\r\n\x1a\n", sha256: "a665788fd84c8a293a25719101354b411d951bebf7cc958200f245b279ac55ee"},
		{path: "favicon.ico", contentType: "image/x-icon", prefix: "\x00\x00\x01\x00", sha256: "18fa56a2443be147555797ce573e5922cd371183cf81acfe5f40e760335d0763"},
		{path: "original-logo.png", contentType: "image/png", prefix: "\x89PNG\r\n\x1a\n", sha256: "31546288c624301e9cfbcf25a4ff922a1d358247bc78f964482780b318b4e057"},
		{path: "apple-touch-icon.png", contentType: "image/png", prefix: "\x89PNG\r\n\x1a\n", sha256: "f5012c266c3f17118b9bc2aea0a231b667d800c683f422ce30680e8dcac47d56"},
		{path: "android-chrome-192x192.png", contentType: "image/png", prefix: "\x89PNG\r\n\x1a\n", sha256: "60e1934ee8c6ddf03d55d6cf73304ba7b2d61a39c0f768df4a834b9e6c1757db"},
		{path: "android-chrome-512x512.png", contentType: "image/png", prefix: "\x89PNG\r\n\x1a\n", sha256: "21b1cbd71b4a8f46320e80a3f121010e49acdcf74afdfe1001a75ea7498253a0"},
		{path: "android-chrome-512x512-maskable.png", contentType: "image/png", prefix: "\x89PNG\r\n\x1a\n", sha256: "21b1cbd71b4a8f46320e80a3f121010e49acdcf74afdfe1001a75ea7498253a0"},
		{path: "site.webmanifest", contentType: "application/manifest+json", prefix: "{", sha256: ""},
		{path: "Monocraft.ttf", contentType: "font/ttf", prefix: "\x00\x01\x00\x00", sha256: "262e31822e0a75be567b02320bfa092143d871e8a273e529d1c5ab0c04cd215f"},
		{path: "Minecraft-Bold.otf", contentType: "font/otf", prefix: "OTTO", sha256: "2a7216d047fd2991aa8a2fcbc002e0782a903d68de3cc248092ffa88aa941b4f"},
		{path: "OFL-1.1.txt", contentType: "text/plain; charset=utf-8", prefix: "Monocraft\n"},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			assetPath := "/font-token/assets/fonts/" + test.path
			if !strings.HasSuffix(test.path, ".ttf") && !strings.HasSuffix(test.path, ".otf") && test.path != "OFL-1.1.txt" {
				assetPath = "/font-token/assets/" + test.path
			}
			response := get(t, server.URL+assetPath)
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			_ = response.Body.Close()
			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
			}
			if got := response.Header.Get("Content-Type"); got != test.contentType {
				t.Fatalf("Content-Type = %q, want %q", got, test.contentType)
			}
			if !strings.HasPrefix(string(body), test.prefix) {
				t.Fatalf("asset %s does not have the expected prefix", test.path)
			}
			if test.sha256 != "" {
				digest := sha256.Sum256(body)
				if got := fmt.Sprintf("%x", digest); got != test.sha256 {
					t.Fatalf("SHA-256 = %s, want %s", got, test.sha256)
				}
			}
			if got := response.Header.Get("Content-Security-Policy"); !strings.Contains(got, "font-src 'self'") {
				t.Fatalf("Content-Security-Policy = %q, want a self-only font source", got)
			}
		})
	}
	manifestResponse := get(t, server.URL+"/font-token/assets/site.webmanifest")
	manifest, err := io.ReadAll(manifestResponse.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = manifestResponse.Body.Close()
	if !strings.Contains(string(manifest), `"name": "BedrockDebugProxy"`) || !strings.Contains(string(manifest), `"start_url": "../"`) || !strings.Contains(string(manifest), `"scope": "../"`) || !strings.Contains(string(manifest), `"theme_color": "#000000"`) || !strings.Contains(string(manifest), `"background_color": "#000000"`) {
		t.Fatalf("manifest does not target the token-scoped dark viewer root: %s", manifest)
	}
}

func TestLiveViewerIndexesAppendedEventsAndVerifiesOnClose(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "live-viewer"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recorder.Close("failed", nil) })
	view, err := OpenLive(root, Options{BlobPreviewBytes: 8})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(view.Handler("live-token"))
	defer server.Close()

	var initial sessionResponse
	decodeURL(t, server.URL+"/live-token/api/session", &initial)
	if initial.Mode != "live" || initial.Verification != "pending" || initial.Events != 0 {
		t.Fatalf("initial session = %+v", initial)
	}

	for i := 1; i <= 3; i++ {
		_, err := recorder.Record(context.Background(), capture.Record{
			Event: capture.Event{
				Kind:      "packet.decoded",
				Direction: capture.DirectionServerToClient,
				Packet:    &capture.PacketInfo{ID: uint32(i), Name: "Packet" + strconv.Itoa(i), DecodeStatus: "decoded"},
			},
			Raw:            []byte{byte(i), byte(i + 1)},
			MediaType:      "application/octet-stream",
			Representation: "packet_payload",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	var tail eventsResponse
	deadline := time.Now().Add(5 * time.Second)
	for {
		decodeURL(t, server.URL+"/live-token/api/events?tail=true&limit=2", &tail)
		if tail.Total == 3 || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if tail.Total != 3 || tail.Offset != 1 || len(tail.Rows) != 2 || tail.Rows[0].Sequence != 2 || tail.Rows[1].Sequence != 3 {
		t.Fatalf("tail events = %+v", tail)
	}

	var detail eventResponse
	decodeURL(t, server.URL+"/live-token/api/events/3", &detail)
	if detail.Event.Sequence != 3 || string(detail.BlobPreview) != string([]byte{3, 4}) || detail.Truncated {
		t.Fatalf("detail = %+v", detail)
	}

	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	var completed sessionResponse
	decodeURL(t, server.URL+"/live-token/api/session", &completed)
	if completed.Mode != "closed" || completed.Verification != "verified" || completed.Events != 3 || completed.Blobs != 3 {
		t.Fatalf("completed session = %+v", completed)
	}
}

func TestLocalServerUsesLoopbackAndStops(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "local-server"})
	if err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
	view, err := Open(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	server, err := StartLocal(view)
	if err != nil {
		t.Fatal(err)
	}
	response := get(t, server.URL+"api/session")
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK || !strings.HasPrefix(server.URL, "http://127.0.0.1:") {
		t.Fatalf("server URL = %q, status = %d", server.URL, response.StatusCode)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Close(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-server.Done():
	default:
		t.Fatal("server did not stop")
	}
}

func TestLiveViewerIgnoresIncompleteTrailingEvent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "capture")
	recorder, err := capture.New(root, capture.Options{CaptureID: "partial-live-event"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recorder.Close("failed", nil) })
	if _, err := recorder.Record(context.Background(), capture.Record{Event: capture.Event{Kind: "complete"}}); err != nil {
		t.Fatal(err)
	}
	view, err := OpenLive(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(view.Handler("partial-token"))
	defer server.Close()

	eventsPath := filepath.Join(root, eventsFileName)
	var originalSize int64
	deadline := time.Now().Add(5 * time.Second)
	for {
		info, statErr := os.Stat(eventsPath)
		if statErr != nil {
			t.Fatal(statErr)
		}
		originalSize = info.Size()
		var session sessionResponse
		decodeURL(t, server.URL+"/partial-token/api/session", &session)
		if session.Events == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("event was not written before deadline: %+v", session)
		}
		time.Sleep(10 * time.Millisecond)
	}
	file, err := os.OpenFile(eventsPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(`{"schema":"incomplete"`); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	var session sessionResponse
	decodeURL(t, server.URL+"/partial-token/api/session", &session)
	if session.Events != 1 || session.Verification != "pending" {
		t.Fatalf("session with partial event = %+v", session)
	}
	if err := os.Truncate(eventsPath, originalSize); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Close("closed", nil); err != nil {
		t.Fatal(err)
	}
}

func TestNewAccessToken(t *testing.T) {
	first, err := NewAccessToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewAccessToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 32 || first == second {
		t.Fatalf("tokens = %q and %q", first, second)
	}
}

func TestOpenBrowserRejectsNonLoopbackURL(t *testing.T) {
	if err := OpenBrowser("https://example.com/capture"); err == nil {
		t.Fatal("OpenBrowser() accepted a non-loopback URL")
	}
}

func TestViewerPaletteTextContrast(t *testing.T) {
	pairs := []struct {
		name, foreground, background string
	}{
		{"muted on black", "#aaaaaa", "#000000"},
		{"muted on surface", "#aaaaaa", "#151515"},
		{"muted on raised", "#aaaaaa", "#2a2a2a"},
		{"white on black", "#ffffff", "#000000"},
		{"white on surface", "#ffffff", "#151515"},
		{"white on raised", "#ffffff", "#2a2a2a"},
		{"aqua on selection", "#55ffff", "#15153f"},
		{"green on surface", "#55ff55", "#151515"},
		{"gold on surface", "#ffaa00", "#151515"},
		{"red on surface", "#ff5555", "#151515"},
		{"magenta on surface", "#ff55ff", "#151515"},
		{"yellow on surface", "#ffff55", "#151515"},
	}
	for _, pair := range pairs {
		if ratio := contrastRatio(pair.foreground, pair.background); ratio < 4.5 {
			t.Errorf("%s contrast ratio = %.2f, want at least 4.5", pair.name, ratio)
		}
	}
}

func contrastRatio(foreground, background string) float64 {
	parse := func(color string) [3]float64 {
		value, err := strconv.ParseUint(strings.TrimPrefix(color, "#"), 16, 32)
		if err != nil {
			panic(err)
		}
		return [3]float64{float64(value >> 16 & 0xff), float64(value >> 8 & 0xff), float64(value & 0xff)}
	}
	luminance := func(color string) float64 {
		channels := parse(color)
		linear := func(channel float64) float64 {
			channel /= 255
			if channel <= 0.03928 {
				return channel / 12.92
			}
			return math.Pow((channel+0.055)/1.055, 2.4)
		}
		return 0.2126*linear(channels[0]) + 0.7152*linear(channels[1]) + 0.0722*linear(channels[2])
	}
	fg, bg := luminance(foreground), luminance(background)
	if fg < bg {
		fg, bg = bg, fg
	}
	return (fg + 0.05) / (bg + 0.05)
}

func get(t *testing.T, url string) *http.Response {
	t.Helper()
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func decodeURL(t *testing.T, url string, target any) {
	t.Helper()
	response := get(t, url)
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, body = %s", response.StatusCode, body)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}

func sequenceText(sequence uint64) string {
	return strconv.FormatUint(sequence, 10)
}
