package capture

import "encoding/json"

const SchemaVersion = "bedrockdebugproxy.capture.v1"

type Direction string

const (
	DirectionClientToServer Direction = "client_to_server"
	DirectionServerToClient Direction = "server_to_client"
	DirectionInternal       Direction = "internal"
	DirectionUnknown        Direction = "unknown"
)

type Severity string

const (
	SeverityDebug Severity = "debug"
	SeverityInfo  Severity = "info"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

type Endpoint struct {
	Network string `json:"network,omitempty"`
	Address string `json:"address,omitempty"`
}

type BlobRef struct {
	SHA256         string `json:"sha256"`
	Size           int64  `json:"size"`
	Path           string `json:"path"`
	MediaType      string `json:"media_type,omitempty"`
	Representation string `json:"representation,omitempty"`
}

type PacketInfo struct {
	ID              uint32 `json:"id"`
	Name            string `json:"name,omitempty"`
	GoType          string `json:"go_type,omitempty"`
	SenderSubClient byte   `json:"sender_sub_client,omitempty"`
	TargetSubClient byte   `json:"target_sub_client,omitempty"`
	DecodeStatus    string `json:"decode_status,omitempty"`
	UnreadBytes     int    `json:"unread_bytes,omitempty"`
}

type ErrorInfo struct {
	Operation string `json:"operation,omitempty"`
	Message   string `json:"message"`
	Type      string `json:"type,omitempty"`
	Temporary *bool  `json:"temporary,omitempty"`
	Timeout   *bool  `json:"timeout,omitempty"`
}

type Event struct {
	Schema         string          `json:"schema"`
	Sequence       uint64          `json:"sequence"`
	Time           string          `json:"time"`
	UnixNano       int64           `json:"unix_nano,string"`
	ElapsedNano    int64           `json:"elapsed_nano,string"`
	CaptureID      string          `json:"capture_id"`
	SessionID      string          `json:"session_id,omitempty"`
	ConnectionID   string          `json:"connection_id,omitempty"`
	Hop            int             `json:"hop,omitempty"`
	Kind           string          `json:"kind"`
	Severity       Severity        `json:"severity,omitempty"`
	Direction      Direction       `json:"direction,omitempty"`
	Channel        string          `json:"channel,omitempty"`
	Stage          string          `json:"stage,omitempty"`
	Source         *Endpoint       `json:"source,omitempty"`
	Destination    *Endpoint       `json:"destination,omitempty"`
	ParentSequence *uint64         `json:"parent_sequence,omitempty"`
	Packet         *PacketInfo     `json:"packet,omitempty"`
	Blob           *BlobRef        `json:"blob,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
	Error          *ErrorInfo      `json:"error,omitempty"`
	Annotations    []string        `json:"annotations,omitempty"`
}

type Generator struct {
	Name       string `json:"name"`
	Version    string `json:"version,omitempty"`
	Commit     string `json:"commit,omitempty"`
	GoVersion  string `json:"go_version"`
	GOOS       string `json:"goos"`
	GOARCH     string `json:"goarch"`
	Executable string `json:"executable,omitempty"`
}

type CaptureOptions struct {
	SyncEachEvent bool              `json:"sync_each_event"`
	RawLayers     []string          `json:"raw_layers,omitempty"`
	Values        map[string]string `json:"values,omitempty"`
}

type CaptureCounts struct {
	Events       uint64 `json:"events"`
	Blobs        uint64 `json:"blobs"`
	BlobBytes    uint64 `json:"blob_bytes"`
	WriteErrors  uint64 `json:"write_errors"`
	Dropped      uint64 `json:"dropped"`
	Truncated    uint64 `json:"truncated"`
	DecodeErrors uint64 `json:"decode_errors"`
}

type Completeness struct {
	Complete    bool     `json:"complete"`
	Limitations []string `json:"limitations,omitempty"`
}

type Manifest struct {
	Schema       string         `json:"schema"`
	CaptureID    string         `json:"capture_id"`
	StartedAt    string         `json:"started_at"`
	EndedAt      string         `json:"ended_at,omitempty"`
	Status       string         `json:"status"`
	Failure      string         `json:"failure,omitempty"`
	Generator    Generator      `json:"generator"`
	Options      CaptureOptions `json:"options"`
	EventsPath   string         `json:"events_path"`
	BlobRoot     string         `json:"blob_root"`
	Counts       CaptureCounts  `json:"counts"`
	Completeness Completeness   `json:"completeness"`
}

type Record struct {
	Event          Event
	Raw            []byte
	MediaType      string
	Representation string
}
