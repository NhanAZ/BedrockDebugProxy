# Audit AUDIT-0117 - Preserve chunk-radius forwarding decode errors

## Intent

Keep strict validation failures in the observed downstream `RequestChunkRadius`
forwarding path visible without changing the packet sent on the wire.

## Scope

- `internal/proxy/proxy.go` records `packet.decode_error` when the observed
  chunk-radius body is truncated or contains trailing bytes.
- The event includes the packet ID, typed packet name, Go type, direction,
  connection, hop, operation, and underlying error. The existing raw packet
  observer remains the authoritative byte evidence.
- `internal/proxy/proxy_test.go` verifies the event context and manifest decode
  error counter.
- Research, fix-history, and changelog entries document the correction.

## Evidence and provenance

- The local capture [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive) established the
  downstream packet-69 body as raw evidence while the strict forwarding helper
  already rejected unread bytes.
- The capture schema already defines `packet.decode_error`, `PacketInfo`, and a
  decode-error manifest counter. No schema version change is needed.

## Automated validation

- `go test ./...`
- `tools/quality.ps1 -LowMemory`
- `git diff --check`

## Local and live validation

The recorder regression is owner-controlled and deterministic. No live or
public-server validation was performed. A fresh stamped local session remains
the next runtime check, and any public-server result requires separate explicit
authorization and reporting.

## Known limitations and follow-up

The initial upstream `RequestChunkRadius{16,16}` remains before the downstream
client's later request because the startup/resource-pack barrier is unchanged.
The new event reports a forwarding-validation failure; it does not attempt to
repair, suppress, or replay malformed traffic.
