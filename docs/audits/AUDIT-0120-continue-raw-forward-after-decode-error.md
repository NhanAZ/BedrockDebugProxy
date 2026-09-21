# Audit AUDIT-0120 - Continue raw forwarding after decode errors

## Intent

Keep the same-protocol bridge observational and byte-preserving when a
derived decode fails but the original packet read is still valid.

## Scope

- `internal/proxy/proxy.go` now records `packet.decode_error`, creates a copied
  `packet.Unknown` derived view, and continues raw forwarding after a clone
  decode failure.
- Recorder, clone, and transport failures keep their existing terminal paths.
- `internal/proxy/proxy_test.go` verifies that the unknown fallback copies the
  payload and handles a nil read, and the loopback integration test forwards a
  malformed raw packet byte-for-byte before continuing with later packets.
- The raw-forwarding decision, capture-format, fix-history, changelog, and
  audit documentation describe the fallback boundary.

## Evidence and provenance

- `third_party/gophertunnel/minecraft/packet.go` reports invalid trailing data,
  while `raw_packet.go` returns the clone error and retains the independently
  owned original header and payload for forwarding.
- The previous bridge branch terminated on the derived decode error despite
  retaining an independently owned raw packet.

## Automated validation

- Focused proxy tests
- `go test ./...`
- `tools/quality.ps1 -LowMemory`
- `git diff --check`

## Local and live validation

The deterministic fallback and loopback integration tests validate payload
ownership, decode-error retention, byte-preserving forwarding, and continued
traffic. No live or public-server validation was performed. A fresh stamped
client run remains required for runtime validation.

## Known limitations and follow-up

This fallback applies only when same-protocol raw forwarding has a valid
`PacketRead` and its owned clone reaches `DecodePacketRead`. Clone failures,
transport failures, and capture recorder failures still stop forwarding. The
destination may independently reject malformed bytes; that result remains
observable through subsequent capture events.
