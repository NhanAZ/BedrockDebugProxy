package viewer

import (
	"bufio"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

const (
	defaultBlobPreviewBytes = 64 << 10
	maxPageSize             = 500
	eventsFileName          = "events.jsonl"
)

//go:embed assets/*
var assets embed.FS

type Options struct {
	BlobPreviewBytes int64
}

type Viewer struct {
	root             string
	resolvedRoot     string
	eventsPath       string
	verification     capture.Verification
	rows             []eventIndex
	bySequence       map[uint64]int
	filters          filterValues
	blobPreviewBytes int64
}

type eventIndex struct {
	Row
	offset int64
	length int
	search string
}

type Row struct {
	Sequence    uint64              `json:"sequence"`
	Time        string              `json:"time"`
	ElapsedNano int64               `json:"elapsed_nano,string"`
	Hop         int                 `json:"hop,omitempty"`
	Direction   capture.Direction   `json:"direction,omitempty"`
	Kind        string              `json:"kind"`
	Severity    capture.Severity    `json:"severity,omitempty"`
	Channel     string              `json:"channel,omitempty"`
	Stage       string              `json:"stage,omitempty"`
	Packet      *capture.PacketInfo `json:"packet,omitempty"`
	Blob        *blobSummary        `json:"blob,omitempty"`
	Error       *capture.ErrorInfo  `json:"error,omitempty"`
	Problem     bool                `json:"problem"`
	Summary     string              `json:"summary"`
}

type blobSummary struct {
	Size           int64  `json:"size"`
	MediaType      string `json:"media_type,omitempty"`
	Representation string `json:"representation,omitempty"`
}

type filterValues struct {
	Directions []string `json:"directions"`
	Kinds      []string `json:"kinds"`
	Packets    []string `json:"packets"`
	Channels   []string `json:"channels"`
}

type sessionResponse struct {
	Manifest     capture.Manifest `json:"manifest"`
	Events       uint64           `json:"events"`
	Blobs        uint64           `json:"blobs"`
	Filters      filterValues     `json:"filters"`
	PreviewBytes int64            `json:"blob_preview_bytes"`
}

type eventsResponse struct {
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
	Total  int   `json:"total"`
	Rows   []Row `json:"rows"`
}

type eventResponse struct {
	Event       capture.Event `json:"event"`
	BlobPreview []byte        `json:"blob_preview,omitempty"`
	Truncated   bool          `json:"blob_preview_truncated"`
}

func Open(root string, options Options) (*Viewer, error) {
	verification, err := capture.Verify(root)
	if err != nil {
		return nil, fmt.Errorf("verify capture: %w", err)
	}
	if len(verification.Issues) != 0 {
		return nil, fmt.Errorf("capture verification found %d issue(s): %s", len(verification.Issues), strings.Join(verification.Issues, "; "))
	}
	if verification.Manifest.Status == "open" || verification.Manifest.EndedAt == "" {
		return nil, errors.New("capture is still open")
	}
	if filepath.ToSlash(filepath.Clean(filepath.FromSlash(verification.Manifest.EventsPath))) != eventsFileName {
		return nil, fmt.Errorf("unsupported event stream path %q", verification.Manifest.EventsPath)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve capture directory: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve capture directory links: %w", err)
	}
	eventsPath, err := resolveCaptureFile(root, resolvedRoot, eventsFileName)
	if err != nil {
		return nil, fmt.Errorf("resolve event stream: %w", err)
	}
	previewBytes := options.BlobPreviewBytes
	if previewBytes <= 0 {
		previewBytes = defaultBlobPreviewBytes
	}
	viewer := &Viewer{
		root:             root,
		resolvedRoot:     resolvedRoot,
		eventsPath:       eventsPath,
		verification:     verification,
		bySequence:       make(map[uint64]int, verification.Events),
		blobPreviewBytes: previewBytes,
	}
	if err := viewer.indexEvents(); err != nil {
		return nil, err
	}
	return viewer, nil
}

func NewAccessToken() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("create viewer access token: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}

func (v *Viewer) Handler(token string) http.Handler {
	prefix := "/" + token
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		setSecurityHeaders(writer)
		if request.URL.Path == prefix {
			http.Redirect(writer, request, prefix+"/", http.StatusTemporaryRedirect)
			return
		}
		if token == "" || !strings.HasPrefix(request.URL.Path, prefix+"/") {
			http.NotFound(writer, request)
			return
		}
		cloned := request.Clone(request.Context())
		cloned.URL.Path = strings.TrimPrefix(request.URL.Path, prefix)
		v.serveHTTP(writer, cloned)
	})
}

func (v *Viewer) serveHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		writer.Header().Set("Allow", "GET, HEAD")
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	switch {
	case request.URL.Path == "/":
		v.serveAsset(writer, request, "assets/index.html", "text/html; charset=utf-8")
	case request.URL.Path == "/assets/app.css":
		v.serveAsset(writer, request, "assets/app.css", "text/css; charset=utf-8")
	case request.URL.Path == "/assets/app.js":
		v.serveAsset(writer, request, "assets/app.js", "text/javascript; charset=utf-8")
	case request.URL.Path == "/api/session":
		v.serveSession(writer)
	case request.URL.Path == "/api/events":
		v.serveEvents(writer, request)
	case strings.HasPrefix(request.URL.Path, "/api/events/"):
		v.serveEvent(writer, request)
	default:
		http.NotFound(writer, request)
	}
}

func (v *Viewer) serveAsset(writer http.ResponseWriter, request *http.Request, name, contentType string) {
	data, err := assets.ReadFile(name)
	if err != nil {
		http.Error(writer, "viewer asset unavailable", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Content-Length", strconv.Itoa(len(data)))
	if request.Method == http.MethodHead {
		return
	}
	_, _ = writer.Write(data)
}

func (v *Viewer) serveSession(writer http.ResponseWriter) {
	writeJSON(writer, http.StatusOK, sessionResponse{
		Manifest:     v.verification.Manifest,
		Events:       v.verification.Events,
		Blobs:        v.verification.Blobs,
		Filters:      v.filters,
		PreviewBytes: v.blobPreviewBytes,
	})
}

func (v *Viewer) serveEvents(writer http.ResponseWriter, request *http.Request) {
	offset, err := queryInteger(request, "offset", 0, 0, len(v.rows))
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, err)
		return
	}
	limit, err := queryInteger(request, "limit", 100, 1, maxPageSize)
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, err)
		return
	}
	query := strings.ToLower(strings.TrimSpace(request.URL.Query().Get("q")))
	direction := request.URL.Query().Get("direction")
	kind := request.URL.Query().Get("kind")
	packet := request.URL.Query().Get("packet")
	channel := request.URL.Query().Get("channel")
	problems := request.URL.Query().Get("problems") == "true"

	matched := make([]Row, 0, limit)
	total := 0
	for i := range v.rows {
		row := &v.rows[i]
		if direction != "" && string(row.Direction) != direction || kind != "" && row.Kind != kind ||
			channel != "" && row.Channel != channel || packet != "" && (row.Packet == nil || row.Packet.Name != packet) ||
			problems && !row.Problem || query != "" && !strings.Contains(row.search, query) {
			continue
		}
		if total >= offset && len(matched) < limit {
			matched = append(matched, row.Row)
		}
		total++
	}
	writeJSON(writer, http.StatusOK, eventsResponse{Offset: offset, Limit: limit, Total: total, Rows: matched})
}

func (v *Viewer) serveEvent(writer http.ResponseWriter, request *http.Request) {
	text := strings.TrimPrefix(request.URL.Path, "/api/events/")
	if text == "" || strings.Contains(text, "/") {
		writeAPIError(writer, http.StatusNotFound, errors.New("event not found"))
		return
	}
	sequence, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, errors.New("invalid event sequence"))
		return
	}
	position, ok := v.bySequence[sequence]
	if !ok {
		writeAPIError(writer, http.StatusNotFound, errors.New("event not found"))
		return
	}
	event, err := v.readEvent(v.rows[position])
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, err)
		return
	}
	response := eventResponse{Event: event}
	if event.Blob != nil {
		response.BlobPreview, response.Truncated, err = v.readBlobPreview(*event.Blob)
		if err != nil {
			writeAPIError(writer, http.StatusInternalServerError, err)
			return
		}
	}
	writeJSON(writer, http.StatusOK, response)
}

func (v *Viewer) indexEvents() error {
	file, err := os.Open(v.eventsPath)
	if err != nil {
		return fmt.Errorf("open event stream: %w", err)
	}
	defer func() { _ = file.Close() }()

	directions := map[string]struct{}{}
	kinds := map[string]struct{}{}
	packets := map[string]struct{}{}
	channels := map[string]struct{}{}
	reader := bufio.NewReaderSize(file, 256*1024)
	var offset int64
	for lineNumber := 1; ; lineNumber++ {
		line, readErr := reader.ReadBytes('\n')
		if len(line) != 0 {
			if line[len(line)-1] != '\n' {
				return fmt.Errorf("event line %d is not newline terminated", lineNumber)
			}
			var event capture.Event
			if err := json.Unmarshal(line, &event); err != nil {
				return fmt.Errorf("decode event line %d: %w", lineNumber, err)
			}
			row := makeRow(event)
			row.offset = offset
			row.length = len(line)
			v.bySequence[event.Sequence] = len(v.rows)
			v.rows = append(v.rows, row)
			addValue(directions, string(event.Direction))
			addValue(kinds, event.Kind)
			addValue(channels, event.Channel)
			if event.Packet != nil {
				addValue(packets, event.Packet.Name)
			}
			offset += int64(len(line))
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read event line %d: %w", lineNumber, readErr)
		}
	}
	v.filters = filterValues{
		Directions: sortedValues(directions),
		Kinds:      sortedValues(kinds),
		Packets:    sortedValues(packets),
		Channels:   sortedValues(channels),
	}
	return nil
}

func (v *Viewer) readEvent(index eventIndex) (capture.Event, error) {
	file, err := os.Open(v.eventsPath)
	if err != nil {
		return capture.Event{}, fmt.Errorf("open event stream: %w", err)
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Seek(index.offset, io.SeekStart); err != nil {
		return capture.Event{}, fmt.Errorf("seek to event %d: %w", index.Sequence, err)
	}
	line := make([]byte, index.length)
	if _, err := io.ReadFull(file, line); err != nil {
		return capture.Event{}, fmt.Errorf("read event %d: %w", index.Sequence, err)
	}
	var event capture.Event
	if err := json.Unmarshal(line, &event); err != nil {
		return capture.Event{}, fmt.Errorf("decode event %d: %w", index.Sequence, err)
	}
	return event, nil
}

func (v *Viewer) readBlobPreview(ref capture.BlobRef) ([]byte, bool, error) {
	path, err := resolveCaptureFile(v.root, v.resolvedRoot, ref.Path)
	if err != nil {
		return nil, false, fmt.Errorf("resolve blob for preview: %w", err)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, false, fmt.Errorf("open blob for preview: %w", err)
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, v.blobPreviewBytes))
	if err != nil {
		return nil, false, fmt.Errorf("read blob preview: %w", err)
	}
	return data, ref.Size > int64(len(data)), nil
}

func makeRow(event capture.Event) eventIndex {
	row := Row{
		Sequence:    event.Sequence,
		Time:        event.Time,
		ElapsedNano: event.ElapsedNano,
		Hop:         event.Hop,
		Direction:   event.Direction,
		Kind:        event.Kind,
		Severity:    event.Severity,
		Channel:     event.Channel,
		Stage:       event.Stage,
		Error:       event.Error,
	}
	if event.Packet != nil {
		packet := *event.Packet
		row.Packet = &packet
		status := strings.ToLower(packet.DecodeStatus)
		row.Problem = status != "" && status != "decoded"
	}
	if event.Blob != nil {
		row.Blob = &blobSummary{Size: event.Blob.Size, MediaType: event.Blob.MediaType, Representation: event.Blob.Representation}
	}
	if event.Error != nil || event.Severity == capture.SeverityError {
		row.Problem = true
	}
	row.Summary = eventSummary(event)
	searchParts := []string{strconv.FormatUint(event.Sequence, 10), event.Time, string(event.Direction), event.Kind, string(event.Severity), event.Channel, event.Stage, row.Summary}
	if event.Packet != nil {
		searchParts = append(searchParts, event.Packet.Name, event.Packet.GoType, event.Packet.DecodeStatus, strconv.FormatUint(uint64(event.Packet.ID), 10))
	}
	if event.Error != nil {
		searchParts = append(searchParts, event.Error.Operation, event.Error.Message, event.Error.Type)
	}
	return eventIndex{Row: row, search: strings.ToLower(strings.Join(searchParts, " "))}
}

func eventSummary(event capture.Event) string {
	if event.Packet != nil {
		name := event.Packet.Name
		if name == "" {
			name = fmt.Sprintf("packet %d", event.Packet.ID)
		}
		if event.Packet.DecodeStatus != "" {
			return name + " - " + event.Packet.DecodeStatus
		}
		return name
	}
	if event.Error != nil {
		return event.Error.Message
	}
	if event.Blob != nil && event.Blob.Representation != "" {
		return event.Blob.Representation
	}
	return event.Kind
}

func resolveCaptureFile(root, resolvedRoot, relative string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relative))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe relative path %q", relative)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(root, clean))
	if err != nil {
		return "", err
	}
	within, err := filepath.Rel(resolvedRoot, resolved)
	if err != nil {
		return "", err
	}
	if within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return "", errors.New("path resolves outside the capture directory")
	}
	return resolved, nil
}

func queryInteger(request *http.Request, name string, fallback, minimum, maximum int) (int, error) {
	text := request.URL.Query().Get(name)
	if text == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(text)
	if err != nil || value < minimum || value > maximum {
		return 0, fmt.Errorf("%s must be between %d and %d", name, minimum, maximum)
	}
	return value, nil
}

func addValue(values map[string]struct{}, value string) {
	if value != "" {
		values[value] = struct{}{}
	}
}

func sortedValues(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func setSecurityHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
	writer.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	writer.Header().Set("Referrer-Policy", "no-referrer")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("X-Frame-Options", "DENY")
}

func writeAPIError(writer http.ResponseWriter, status int, err error) {
	writeJSON(writer, status, struct {
		Error string `json:"error"`
	}{Error: err.Error()})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
