package analysis

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
)

type Summary struct {
	Schema             string                `json:"schema"`
	CaptureID          string                `json:"capture_id"`
	Build              Build                 `json:"build"`
	Protocol           Protocol              `json:"protocol"`
	Session            Session               `json:"session"`
	Status             string                `json:"status"`
	StartedAt          string                `json:"started_at"`
	EndedAt            string                `json:"ended_at,omitempty"`
	Complete           bool                  `json:"complete"`
	Limitations        []string              `json:"limitations,omitempty"`
	ManifestCounts     capture.CaptureCounts `json:"manifest_counts"`
	ObservedEvents     uint64                `json:"observed_events"`
	FirstSequence      uint64                `json:"first_sequence,omitempty"`
	LastSequence       uint64                `json:"last_sequence,omitempty"`
	Kinds              []NamedCount          `json:"kinds,omitempty"`
	Directions         []NamedCount          `json:"directions,omitempty"`
	Channels           []NamedCount          `json:"channels,omitempty"`
	Packets            []PacketCount         `json:"packets,omitempty"`
	Artifacts          []ArtifactCount       `json:"artifacts,omitempty"`
	ResourcePacks      []ResourcePack        `json:"resource_packs,omitempty"`
	PackDecryptions    []PackDecryption      `json:"resource_pack_decryptions,omitempty"`
	Errors             []ErrorCount          `json:"errors,omitempty"`
	VerificationIssues []string              `json:"verification_issues,omitempty"`
}

type Build struct {
	Version   string `json:"version,omitempty"`
	Commit    string `json:"commit,omitempty"`
	GoVersion string `json:"go_version"`
	GOOS      string `json:"goos"`
	GOARCH    string `json:"goarch"`
}

type Protocol struct {
	ConfiguredID       string `json:"configured_id,omitempty"`
	ConfiguredVersion  string `json:"configured_version,omitempty"`
	DownstreamID       int32  `json:"downstream_id,omitempty"`
	DownstreamVersion  string `json:"downstream_version,omitempty"`
	UpstreamID         int32  `json:"upstream_id,omitempty"`
	UpstreamVersion    string `json:"upstream_version,omitempty"`
	PackDecryptEnabled bool   `json:"resource_pack_decryption_enabled"`
}

type Session struct {
	UpstreamConnected    bool   `json:"upstream_connected"`
	Negotiated           bool   `json:"negotiated"`
	Spawned              bool   `json:"spawned"`
	Closed               bool   `json:"closed"`
	ConnectionViews      uint64 `json:"connection_views"`
	GameDataViews        uint64 `json:"game_data_views"`
	DecodedPacketEvents  uint64 `json:"decoded_packet_events"`
	UnknownPacketEvents  uint64 `json:"unknown_packet_events"`
	TransportPayloads    uint64 `json:"transport_payload_events"`
	RawPacketEvents      uint64 `json:"raw_packet_events"`
	DecodeErrorEvents    uint64 `json:"decode_error_events"`
	ErrorEvents          uint64 `json:"error_events"`
	StructuredViewErrors uint64 `json:"structured_view_errors"`
}

type NamedCount struct {
	Name  string `json:"name"`
	Count uint64 `json:"count"`
}

type PacketCount struct {
	Direction    capture.Direction `json:"direction"`
	ID           uint32            `json:"id"`
	Name         string            `json:"name,omitempty"`
	DecodeStatus string            `json:"decode_status,omitempty"`
	Count        uint64            `json:"count"`
}

type ArtifactCount struct {
	Representation string `json:"representation"`
	References     uint64 `json:"references"`
	UniqueBlobs    uint64 `json:"unique_blobs"`
	UniqueBytes    uint64 `json:"unique_bytes"`
}

type ResourcePack struct {
	Sequence       uint64 `json:"sequence"`
	UUID           string `json:"uuid"`
	Version        string `json:"version"`
	Name           string `json:"name"`
	ArchiveBytes   int64  `json:"archive_bytes"`
	ChecksumSHA256 string `json:"checksum_sha256"`
	Delivery       string `json:"delivery"`
	Encrypted      bool   `json:"encrypted"`
	HasContentKey  bool   `json:"has_content_key"`
	BlobSHA256     string `json:"blob_sha256,omitempty"`
}

type PackDecryption struct {
	Sequence       uint64 `json:"sequence"`
	ParentSequence uint64 `json:"parent_sequence,omitempty"`
	UUID           string `json:"uuid"`
	Version        string `json:"version"`
	Status         string `json:"status"`
	Algorithm      string `json:"algorithm,omitempty"`
	Authenticated  bool   `json:"authenticated"`
	DecryptedFiles int    `json:"decrypted_files,omitempty"`
	BlobSHA256     string `json:"blob_sha256,omitempty"`
	Error          string `json:"error,omitempty"`
}

type ErrorCount struct {
	Kind          string `json:"kind"`
	Operation     string `json:"operation,omitempty"`
	Type          string `json:"type,omitempty"`
	Message       string `json:"message,omitempty"`
	Count         uint64 `json:"count"`
	FirstSequence uint64 `json:"first_sequence"`
	LastSequence  uint64 `json:"last_sequence"`
}

type packetKey struct {
	direction    capture.Direction
	id           uint32
	name         string
	decodeStatus string
}

type errorKey struct {
	kind      string
	operation string
	typeName  string
	message   string
}

type artifactAggregate struct {
	references uint64
	blobs      map[string]int64
}

func Analyze(root string) (Summary, error) {
	verification, err := capture.Verify(root)
	if err != nil {
		return Summary{}, err
	}
	manifest := verification.Manifest
	summary := Summary{
		Schema:    manifest.Schema,
		CaptureID: manifest.CaptureID,
		Build: Build{
			Version: manifest.Generator.Version, Commit: manifest.Generator.Commit,
			GoVersion: manifest.Generator.GoVersion, GOOS: manifest.Generator.GOOS, GOARCH: manifest.Generator.GOARCH,
		},
		Protocol: Protocol{
			ConfiguredID: manifest.Options.Values["protocol_id"], ConfiguredVersion: manifest.Options.Values["game_version"],
			PackDecryptEnabled: manifest.Options.Values["decrypt_resource_packs"] == "true",
		},
		Status:             manifest.Status,
		StartedAt:          manifest.StartedAt,
		EndedAt:            manifest.EndedAt,
		Complete:           manifest.Completeness.Complete,
		Limitations:        append([]string(nil), manifest.Completeness.Limitations...),
		ManifestCounts:     manifest.Counts,
		VerificationIssues: append([]string(nil), verification.Issues...),
	}
	kinds := make(map[string]uint64)
	directions := make(map[string]uint64)
	channels := make(map[string]uint64)
	packets := make(map[packetKey]uint64)
	errorsByKey := make(map[errorKey]*ErrorCount)
	artifacts := make(map[string]*artifactAggregate)
	err = capture.ScanEvents(root, func(event capture.Event) error {
		summary.ObservedEvents++
		if summary.FirstSequence == 0 {
			summary.FirstSequence = event.Sequence
		}
		summary.LastSequence = event.Sequence
		kinds[event.Kind]++
		directions[string(event.Direction)]++
		channels[event.Channel]++
		switch event.Kind {
		case "upstream.connected":
			summary.Session.UpstreamConnected = true
		case "session.negotiated":
			summary.Session.Negotiated = true
			var negotiated struct {
				DownstreamID      int32  `json:"downstream_protocol_id"`
				DownstreamVersion string `json:"downstream_game_version"`
				UpstreamID        int32  `json:"upstream_protocol_id"`
				UpstreamVersion   string `json:"upstream_game_version"`
			}
			if err := json.Unmarshal(event.Data, &negotiated); err != nil {
				return fmt.Errorf("decode session negotiation event %d: %w", event.Sequence, err)
			}
			summary.Protocol.DownstreamID = negotiated.DownstreamID
			summary.Protocol.DownstreamVersion = negotiated.DownstreamVersion
			summary.Protocol.UpstreamID = negotiated.UpstreamID
			summary.Protocol.UpstreamVersion = negotiated.UpstreamVersion
		case "session.spawned":
			summary.Session.Spawned = true
		case "session.close":
			summary.Session.Closed = true
		case "session.connection_metadata":
			summary.Session.ConnectionViews++
		case "session.game_data":
			summary.Session.GameDataViews++
		case "packet.decoded":
			summary.Session.DecodedPacketEvents++
		case "packet.unknown":
			summary.Session.UnknownPacketEvents++
		case "transport.payload":
			summary.Session.TransportPayloads++
		case "packet.raw":
			summary.Session.RawPacketEvents++
		case "packet.decode_error":
			summary.Session.DecodeErrorEvents++
		case "capture.view_error":
			summary.Session.StructuredViewErrors++
		}
		if event.Packet != nil {
			packets[packetKey{
				direction: event.Direction, id: event.Packet.ID,
				name: event.Packet.Name, decodeStatus: event.Packet.DecodeStatus,
			}]++
		}
		if event.Blob != nil {
			representation := event.Blob.Representation
			aggregate := artifacts[representation]
			if aggregate == nil {
				aggregate = &artifactAggregate{blobs: make(map[string]int64)}
				artifacts[representation] = aggregate
			}
			aggregate.references++
			aggregate.blobs[event.Blob.SHA256] = event.Blob.Size
		}
		if event.Error != nil || event.Severity == capture.SeverityError {
			summary.Session.ErrorEvents++
			key := errorKey{kind: event.Kind}
			if event.Error != nil {
				key.operation = event.Error.Operation
				key.typeName = event.Error.Type
				key.message = event.Error.Message
			}
			aggregate := errorsByKey[key]
			if aggregate == nil {
				aggregate = &ErrorCount{
					Kind: key.kind, Operation: key.operation, Type: key.typeName, Message: key.message,
					FirstSequence: event.Sequence,
				}
				errorsByKey[key] = aggregate
			}
			aggregate.Count++
			aggregate.LastSequence = event.Sequence
		}
		if event.Kind == "resource_pack.archive" {
			var pack struct {
				UUID           string `json:"uuid"`
				Version        string `json:"version"`
				Name           string `json:"name"`
				ArchiveBytes   int64  `json:"archive_bytes"`
				ChecksumSHA256 string `json:"checksum_sha256"`
				Delivery       string `json:"delivery"`
				Encrypted      bool   `json:"encrypted"`
				ContentKey     string `json:"content_key"`
			}
			if err := json.Unmarshal(event.Data, &pack); err != nil {
				return fmt.Errorf("decode resource pack event %d: %w", event.Sequence, err)
			}
			entry := ResourcePack{
				Sequence: event.Sequence, UUID: pack.UUID, Version: pack.Version, Name: pack.Name,
				ArchiveBytes: pack.ArchiveBytes, ChecksumSHA256: pack.ChecksumSHA256,
				Delivery: pack.Delivery, Encrypted: pack.Encrypted, HasContentKey: pack.ContentKey != "",
			}
			if event.Blob != nil {
				entry.BlobSHA256 = event.Blob.SHA256
			}
			summary.ResourcePacks = append(summary.ResourcePacks, entry)
		}
		if event.Kind == "resource_pack.decrypted_archive" {
			var derived struct {
				UUID    string `json:"uuid"`
				Version string `json:"version"`
				Report  struct {
					Algorithm      string `json:"algorithm"`
					Authenticated  bool   `json:"authenticated"`
					DecryptedFiles int    `json:"decrypted_files"`
				} `json:"report"`
			}
			if err := json.Unmarshal(event.Data, &derived); err != nil {
				return fmt.Errorf("decode resource pack decryption event %d: %w", event.Sequence, err)
			}
			entry := PackDecryption{
				Sequence: event.Sequence, UUID: derived.UUID, Version: derived.Version, Status: "decrypted",
				Algorithm: derived.Report.Algorithm, Authenticated: derived.Report.Authenticated,
				DecryptedFiles: derived.Report.DecryptedFiles,
			}
			if event.ParentSequence != nil {
				entry.ParentSequence = *event.ParentSequence
			}
			if event.Blob != nil {
				entry.BlobSHA256 = event.Blob.SHA256
			}
			summary.PackDecryptions = append(summary.PackDecryptions, entry)
		}
		if event.Kind == "resource_pack.decrypt_error" {
			var derived struct {
				UUID    string `json:"uuid"`
				Version string `json:"version"`
			}
			if err := json.Unmarshal(event.Data, &derived); err != nil {
				return fmt.Errorf("decode resource pack decryption error event %d: %w", event.Sequence, err)
			}
			entry := PackDecryption{Sequence: event.Sequence, UUID: derived.UUID, Version: derived.Version, Status: "error"}
			if event.ParentSequence != nil {
				entry.ParentSequence = *event.ParentSequence
			}
			if event.Error != nil {
				entry.Error = event.Error.Message
			}
			summary.PackDecryptions = append(summary.PackDecryptions, entry)
		}
		return nil
	})
	if err != nil {
		return Summary{}, err
	}
	summary.Kinds = sortedNamedCounts(kinds)
	summary.Directions = sortedNamedCounts(directions)
	summary.Channels = sortedNamedCounts(channels)
	for key, count := range packets {
		summary.Packets = append(summary.Packets, PacketCount{
			Direction: key.direction, ID: key.id, Name: key.name, DecodeStatus: key.decodeStatus, Count: count,
		})
	}
	sort.Slice(summary.Packets, func(i, j int) bool {
		left, right := summary.Packets[i], summary.Packets[j]
		if left.Direction != right.Direction {
			return left.Direction < right.Direction
		}
		if left.ID != right.ID {
			return left.ID < right.ID
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		return left.DecodeStatus < right.DecodeStatus
	})
	for representation, aggregate := range artifacts {
		entry := ArtifactCount{Representation: representation, References: aggregate.references, UniqueBlobs: uint64(len(aggregate.blobs))}
		for _, size := range aggregate.blobs {
			if size > 0 {
				entry.UniqueBytes += uint64(size)
			}
		}
		summary.Artifacts = append(summary.Artifacts, entry)
	}
	sort.Slice(summary.Artifacts, func(i, j int) bool { return summary.Artifacts[i].Representation < summary.Artifacts[j].Representation })
	for _, aggregate := range errorsByKey {
		summary.Errors = append(summary.Errors, *aggregate)
	}
	sort.Slice(summary.Errors, func(i, j int) bool {
		if summary.Errors[i].FirstSequence != summary.Errors[j].FirstSequence {
			return summary.Errors[i].FirstSequence < summary.Errors[j].FirstSequence
		}
		return summary.Errors[i].Kind < summary.Errors[j].Kind
	})
	return summary, nil
}

func Explain(summary Summary) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# BedrockDebugProxy capture explanation\n\n")
	fmt.Fprintf(&output, "Capture `%s` uses schema `%s` and has status `%s`. It contains %d observed events from sequence %d through %d.\n\n", sanitizeInline(summary.CaptureID), sanitizeInline(summary.Schema), sanitizeInline(summary.Status), summary.ObservedEvents, summary.FirstSequence, summary.LastSequence)
	if summary.Build.Commit != "" {
		fmt.Fprintf(&output, "The capture was produced by revision `%s` version `%s` on `%s/%s`.\n\n", sanitizeInline(summary.Build.Commit), sanitizeInline(summary.Build.Version), sanitizeInline(summary.Build.GOOS), sanitizeInline(summary.Build.GOARCH))
	}
	fmt.Fprintf(&output, "## Integrity and completeness\n\n")
	if len(summary.VerificationIssues) == 0 {
		fmt.Fprintf(&output, "The capture verifier found no event ordering, manifest count, blob path, size, or SHA-256 integrity issues.\n\n")
	} else {
		fmt.Fprintf(&output, "The capture verifier found %d issue(s). Treat conclusions as provisional.\n\n", len(summary.VerificationIssues))
		for _, issue := range summary.VerificationIssues {
			fmt.Fprintf(&output, "- `%s`\n", sanitizeInline(issue))
		}
		output.WriteString("\n")
	}
	if summary.Complete {
		output.WriteString("The manifest marks the capture complete.\n\n")
	} else {
		output.WriteString("The manifest marks the capture incomplete or limited.\n\n")
	}
	if len(summary.Limitations) != 0 {
		for _, limitation := range summary.Limitations {
			fmt.Fprintf(&output, "- `%s`\n", sanitizeInline(limitation))
		}
		output.WriteString("\n")
	}
	fmt.Fprintf(&output, "## Traffic\n\n")
	fmt.Fprintf(&output, "Session evidence reports upstream connected `%t`, negotiated `%t`, spawned `%t`, and closed `%t`. It retained %d raw packet event(s), %d transport payload event(s), and %d decoded packet event(s).\n\n", summary.Session.UpstreamConnected, summary.Session.Negotiated, summary.Session.Spawned, summary.Session.Closed, summary.Session.RawPacketEvents, summary.Session.TransportPayloads, summary.Session.DecodedPacketEvents)
	for _, direction := range summary.Directions {
		fmt.Fprintf(&output, "- `%s` has %d event(s)\n", sanitizeInline(direction.Name), direction.Count)
	}
	output.WriteString("\n")
	if len(summary.Packets) == 0 {
		output.WriteString("No packet metadata events were found.\n\n")
	} else {
		fmt.Fprintf(&output, "The capture contains %d distinct packet ID, direction, name, and decode-status combinations.\n\n", len(summary.Packets))
	}
	fmt.Fprintf(&output, "## Resource packs\n\n")
	if len(summary.ResourcePacks) == 0 {
		output.WriteString("No reconstructed resource-pack archive events were found.\n\n")
	} else {
		for _, pack := range summary.ResourcePacks {
			fmt.Fprintf(&output, "- `%s` version `%s` UUID `%s` has %d archive bytes via `%s`. Encrypted is %t and content key retained is %t.\n", sanitizeInline(pack.Name), sanitizeInline(pack.Version), sanitizeInline(pack.UUID), pack.ArchiveBytes, sanitizeInline(pack.Delivery), pack.Encrypted, pack.HasContentKey)
		}
		for _, decryption := range summary.PackDecryptions {
			if decryption.Status == "decrypted" {
				fmt.Fprintf(&output, "- Derived archive for UUID `%s` decrypted %d file(s) with `%s`. Authenticated plaintext is %t.\n", sanitizeInline(decryption.UUID), decryption.DecryptedFiles, sanitizeInline(decryption.Algorithm), decryption.Authenticated)
			} else {
				fmt.Fprintf(&output, "- Derived decryption for UUID `%s` failed and the original archive remains available", sanitizeInline(decryption.UUID))
				if decryption.Error != "" {
					fmt.Fprintf(&output, " with error `%s`", sanitizeInline(decryption.Error))
				}
				output.WriteString(".\n")
			}
		}
		output.WriteString("\n")
	}
	fmt.Fprintf(&output, "## Errors\n\n")
	if len(summary.Errors) == 0 {
		output.WriteString("No error-severity or structured error events were found.\n\n")
	} else {
		for _, entry := range summary.Errors {
			fmt.Fprintf(&output, "- `%s` occurred %d time(s) from sequence %d through %d", sanitizeInline(entry.Kind), entry.Count, entry.FirstSequence, entry.LastSequence)
			if entry.Message != "" {
				fmt.Fprintf(&output, " with message `%s`", sanitizeInline(entry.Message))
			}
			output.WriteString(".\n")
		}
		output.WriteString("\n")
	}
	output.WriteString("Capture integrity does not by itself prove compatibility with a real Minecraft client or every Bedrock server implementation. Compare packet sequences and failures with the named server flow used during collection.\n")
	return output.String()
}

func sortedNamedCounts(values map[string]uint64) []NamedCount {
	result := make([]NamedCount, 0, len(values))
	for name, count := range values {
		result = append(result, NamedCount{Name: name, Count: count})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func sanitizeInline(value string) string {
	value = strings.ReplaceAll(value, "`", "'")
	value = strings.ReplaceAll(value, "\r", " ")
	return strings.ReplaceAll(value, "\n", " ")
}
