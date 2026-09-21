# Audit AUDIT-0126 - Test client-cache blob forwarding

- Commit subject: `Test client-cache blob forwarding`
- Change type: `test | documentation`
- Runtime impact: `no`

## Intent

Extend the owner-controlled loopback evidence beyond login capability
negotiation so the cache-enabled chunk and miss-response exchange is covered by
the raw forwarding regression.

## Scope

- Send a cache-enabled `LevelChunk` with a deterministic blob hash and payload
  from the local upstream fixture.
- Send `ClientCacheBlobStatus` from the client side and return a matching
  `ClientCacheMissResponse` from the fixture.
- Assert typed values and raw packet hash presence on both bridge channels.
- Update the client-cache research note, ADR limitation, and fix-history index.

## Evidence and provenance

The packet structures and raw forwarding boundary are the vendored gophertunnel
`LevelChunk`, `ClientCacheBlobStatus`, and `ClientCacheMissResponse` types plus
the existing same-protocol loopback integration. The fixture uses synthetic
bytes and does not contain account or server data.

## Validation

- Focused test: `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed.
- Full quality gate: `tools/quality.ps1 -LowMemory` passed before commit.
- Live validation: not observed. No public server or Minecraft client was
  contacted.

## Limitations and follow-up

The fixture proves protocol forwarding and raw evidence preservation. It does
not prove that a real vanilla client stores, hits, or reuses cache blobs. A
controlled local real-client capture remains the appropriate follow-up.
