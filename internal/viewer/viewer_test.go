package viewer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

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
	if response.StatusCode != http.StatusOK || !strings.Contains(string(index), "Capture Viewer") {
		t.Fatalf("index status = %d, body = %s", response.StatusCode, index)
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
