# Audit AUDIT-0119 - Classify raw-forward decode errors

## Intent

Distinguish protocol decode failures from transport read failures while
retaining the original raw packet evidence.

## Scope

- `internal/proxy/proxy.go` now records `packet.decode_error` when
  `DecodePacketRead` fails while decoding a clone for same-protocol forwarding.
- The event retains the raw packet ID and sender or target sub-client header
  values. Clone or transport errors keep their existing bridge classifications.
- `internal/proxy/proxy_test.go` covers the packet decode finalizer and manifest
  counter.
- Capture-format, fix-history, changelog, and audit documentation describe the
  classification boundary.

## Evidence and provenance

- `third_party/gophertunnel/minecraft/raw_packet.go` keeps `PacketRead.Header`
  and raw ownership separate from `DecodePacketRead`, so a failed clone decode
  can be classified without losing the packet header or raw observer event.
- The previous forwarding branch routed every returned error to
  `bridge.read_error`, which obscured the failure class in analysis.

## Automated validation

- `go test ./...`
- `tools/quality.ps1 -LowMemory`
- `git diff --check`

## Local and live validation

The deterministic recorder regression validates classification and retention.
No live or public-server validation was performed. A fresh stamped client run
remains required for runtime validation.

## Known limitations and follow-up

Only same-protocol raw-forward clone decode failures use the new classification.
Non-raw `ReadPacket` failures do not expose a retained header at that boundary
and remain `bridge.read_error`. The wire forwarding path is unchanged.
