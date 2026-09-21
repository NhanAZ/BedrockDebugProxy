# Audit AUDIT-0116 - Record chunk-radius forward errors

## Intent

Keep failures in the observed downstream `RequestChunkRadius` forwarding path
visible in the capture instead of silently discarding them.

## Scope

- `internal/proxy/proxy.go` now records `bridge.write_error` when the paired
  upstream connection is unavailable or rejects the typed request.
- `internal/proxy/proxy_test.go` covers successful typed forwarding and error
  propagation through the writer boundary.
- `docs/research/chunk-radius-parity.md`, `docs/fix-history.md`, and
  `CHANGELOG.md` document the observability correction.
- No packet bytes, ordering, authentication, compression, or resource-pack
  behavior changed.

## Evidence and provenance

- Owner-controlled capture `captures/session-20260921T122228Z` records the
  downstream `RequestChunkRadius` payload `301c` and the prior upstream
  `2010` fallback.
- Source review found that the later typed write used `_ = upstream.WritePacket`
  and had no error event for a closed or unavailable upstream connection.
- The existing gophertunnel packet definition remains the authority for the
  typed `ChunkRadius` and `MaxChunkRadius` fields.

## Automated validation

- `go test ./internal/proxy -run 'TestForwardRequestChunkRadius|TestDecodeRequestChunkRadius' -count=1`
- `git diff --check`
- Full `tools/quality.ps1 -LowMemory` is required before commit.

## Local validation

The focused tests verify that the forwarded packet retains both typed values
and that a writer error is returned to the capture boundary. The existing local
loopback test continues to verify 16/16 followed by 24/28 forwarding.

## Live validation status

No public-server validation was performed. A fresh stamped local client run
against the current commit remains pending.

## Known limitations and follow-up

The upstream connection still sends its normal 16/16 initial request before the
downstream client emits packet 69. This change only makes failures of the later
forward explicit; it does not redesign the startup/resource-pack barrier.
