# Audit AUDIT-0118 - Preserve chunk-radius decode header context

## Intent

Retain the complete observed packet header context when a downstream
`RequestChunkRadius` body fails strict forwarding validation.

## Scope

- `internal/proxy/proxy.go` now carries the existing `PacketInfo` sender and
  target sub-client values into the `packet.decode_error` event.
- `internal/proxy/proxy_test.go` verifies both non-zero sub-client values and
  the existing typed packet context.
- Research, fix-history, and changelog entries document the correction.
- No packet bytes, forwarding order, authentication, compression, or resource
  pack behavior changes.

## Evidence and provenance

- The raw observer already receives the packet header and stores
  `SenderSubClient` and `TargetSubClient` in `PacketInfo`.
- The prior decode-error helper copied only the packet ID and type, creating a
  narrower derived event than the raw evidence. The capture schema already
  supports the missing fields, so no schema version change is needed.

## Automated validation

- `go test ./...`
- `tools/quality.ps1 -LowMemory`
- `git diff --check`

## Local and live validation

This is a deterministic capture-context correction covered by the owner-
controlled recorder regression. No live or public-server validation was
performed. Runtime validation remains separately pending for the current
stamped binary.

## Known limitations and follow-up

The change only enriches a derived error event. The raw packet remains the
authoritative evidence, and the startup chunk-radius ordering boundary is
unchanged.
