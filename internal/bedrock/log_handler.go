package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

type CaptureLogHandler struct {
	recorder     *capture.Recorder
	failures     *FailureSink
	sessionID    string
	connectionID string
	channel      string
	attrs        []boundAttribute
	groups       []string
}

type boundAttribute struct {
	groups []string
	value  slog.Attr
}

func NewCaptureLogHandler(recorder *capture.Recorder, failures *FailureSink, sessionID, connectionID, channel string) *CaptureLogHandler {
	return &CaptureLogHandler{
		recorder: recorder, failures: failures, sessionID: sessionID,
		connectionID: connectionID, channel: channel,
	}
}

func (h *CaptureLogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *CaptureLogHandler) Handle(ctx context.Context, record slog.Record) error {
	attributes := make(map[string]any)
	for _, attribute := range h.attrs {
		putSlogAttribute(attributes, attribute.groups, attribute.value)
	}
	record.Attrs(func(attribute slog.Attr) bool {
		putSlogAttribute(attributes, h.groups, attribute)
		return true
	})
	data, err := json.Marshal(map[string]any{
		"message":     record.Message,
		"level":       record.Level.String(),
		"logger_time": record.Time.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		"attributes":  attributes,
	})
	if err != nil {
		h.failures.Set(err)
		return err
	}
	kind := "library.log"
	lower := strings.ToLower(record.Message)
	if strings.Contains(lower, "decode packet") || strings.Contains(lower, "read packet header") ||
		strings.Contains(lower, "decode batch") || strings.Contains(lower, "decompress batch") ||
		strings.Contains(lower, "verify batch") {
		kind = "packet.decode_error"
	}
	_, err = h.recorder.Record(ctx, capture.Record{Event: capture.Event{
		SessionID:    h.sessionID,
		ConnectionID: h.connectionID,
		Kind:         kind,
		Severity:     slogSeverity(record.Level),
		Direction:    capture.DirectionUnknown,
		Channel:      h.channel,
		Stage:        "library",
		Data:         data,
	}})
	if err != nil {
		h.failures.Set(err)
	}
	return err
}

func (h *CaptureLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append([]boundAttribute(nil), h.attrs...)
	for _, attribute := range attrs {
		clone.attrs = append(clone.attrs, boundAttribute{
			groups: append([]string(nil), h.groups...),
			value:  attribute,
		})
	}
	clone.groups = append([]string(nil), h.groups...)
	return &clone
}

func (h *CaptureLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	clone := *h
	clone.attrs = append([]boundAttribute(nil), h.attrs...)
	clone.groups = append(append([]string(nil), h.groups...), name)
	return &clone
}

func putSlogAttribute(destination map[string]any, groups []string, attribute slog.Attr) {
	attribute.Value = attribute.Value.Resolve()
	if attribute.Equal(slog.Attr{}) {
		return
	}
	current := destination
	for _, group := range groups {
		next, ok := current[group].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[group] = next
		}
		current = next
	}
	if attribute.Value.Kind() == slog.KindGroup {
		nestedGroups := append([]string(nil), groups...)
		if attribute.Key != "" {
			nestedGroups = append(nestedGroups, attribute.Key)
		}
		for _, child := range attribute.Value.Group() {
			putSlogAttribute(destination, nestedGroups, child)
		}
		return
	}
	current[attribute.Key] = slogValue(attribute.Value)
}

func slogValue(value slog.Value) any {
	switch value.Kind() {
	case slog.KindBool:
		return value.Bool()
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindFloat64:
		return value.Float64()
	case slog.KindInt64:
		return strconv.FormatInt(value.Int64(), 10)
	case slog.KindString:
		return value.String()
	case slog.KindTime:
		return value.Time().UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
	case slog.KindUint64:
		return strconv.FormatUint(value.Uint64(), 10)
	case slog.KindAny:
		if err, ok := value.Any().(error); ok {
			return map[string]string{"message": err.Error(), "type": fmt.Sprintf("%T", err)}
		}
		return fmt.Sprint(value.Any())
	default:
		return value.String()
	}
}

func slogSeverity(level slog.Level) capture.Severity {
	switch {
	case level >= slog.LevelError:
		return capture.SeverityError
	case level >= slog.LevelWarn:
		return capture.SeverityWarn
	case level <= slog.LevelDebug:
		return capture.SeverityDebug
	default:
		return capture.SeverityInfo
	}
}
