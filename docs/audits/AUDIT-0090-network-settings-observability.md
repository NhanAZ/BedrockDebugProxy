# Audit AUDIT-0090 - Index per-connection NetworkSettings

- Commit subject: `Index per-connection NetworkSettings`
- Change type: `capture observability`
- Runtime impact: `no wire behavior change`

## Intent

Expose the compression and client-throttle settings negotiated on each side of a session so local parity analysis can identify a real `NetworkSettings` difference without decoding raw blobs manually.

## Scope

- Retain a copy of the decoded `NetworkSettings` packet on each gophertunnel connection.
- Add bounded downstream and upstream settings objects and an explicit match flag to `session.negotiated`.
- Keep raw packet capture authoritative and leave compression, batching, forwarding, and listener handshake behavior unchanged.
- Document the confirmed local threshold mismatch and the unresolved design boundary.

## Evidence and provenance

- Owner-controlled capture `captures/session-20260921T003441Z` recorded upstream threshold `1` and downstream threshold `256`, with algorithm `0` on both sides.
- The packet fields and login handling are defined by the vendored gophertunnel `NetworkSettings` packet and connection handlers.
- No public server, external account, authentication token, or external service was used.

## Validation

- `go test ./...` in `third_party/gophertunnel` passed.
- `go test ./...` in the root module passed.
- `go test ./internal/proxy -run TestRunnerForwardsBidirectionalPacketsAndCapturesSession -count=1 -v` passed.
- `tools/quality.ps1` passed, including formatting, tests, build, static analysis, vulnerability scan, capture checks, and whitespace validation.
- Live vanilla-client validation: `not_observed`.

## Limitations and follow-up

The change does not normalize downstream compression settings or claim that matching them improves compatibility. A safe per-connection policy would require evidence and a design that obtains upstream settings before the downstream response is sent. The existing raw packet and transport evidence remain necessary for exact batching analysis.
