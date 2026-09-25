# Audit AUDIT-0115 - Forward downstream chunk-radius requests

## Intent

Preserve the observed downstream `RequestChunkRadius` values on the paired upstream connection without delaying upstream login or changing gameplay packet forwarding.

## Scope

The proxy listener decodes packet ID 69 and writes its typed `ChunkRadius` and `MaxChunkRadius` values to the active upstream connection. The vendored gophertunnel dialer has a small opt-in provider used by deterministic client-side loopback fixtures. No capture schema, authentication, compression, or resource-pack behavior changed.

## Evidence and provenance

- Owner-controlled closed capture: [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive).
- Downstream packet ID 69 body: `301c`, decoded as 24/28.
- Upstream packet ID 69 body: `2010`, decoded as 16/16.
- Upstream source baseline: Sandertv gophertunnel commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`.
- Detailed classification: [`docs/research/chunk-radius-parity.md`](../research/chunk-radius-parity.md).

## Affected files and interfaces

- `internal/proxy/proxy.go` adds strict packet-69 decoding and forwarding at the listener boundary.
- `internal/proxy/proxy_test.go` verifies a local bridge observes 16/16 followed by 24/28 upstream.
- `third_party/gophertunnel/minecraft/conn.go` and `dial.go` add the opt-in fixture provider while retaining the default.
- `third_party/gophertunnel/minecraft/conn_featured_experience_test.go` verifies provider encoding.
- `CHANGELOG.md`, `docs/fix-history.md`, and the linked research note document the behavior and limitation.

## Automated validation

- `go test ./internal/proxy -run TestRunnerForwardsBidirectionalPacketsAndCapturesSession -count=1`
- `go test ./minecraft -run TestHandleItemRegistryUsesRequestChunkRadiusProvider -count=1` from `third_party/gophertunnel`.
- `git diff --check`
- Full `tools/quality.ps1 -LowMemory` is required before commit.

## Live validation status

No public-server validation was performed for this change. The local loopback test is the available deterministic validation. The Hive baseline remains a separate required runtime check under the project validation policy.

## Known limitations and follow-up

The upstream connection still sends its normal 16/16 initial request before the downstream client emits packet 69. The observed downstream request is then forwarded as a later request, and the server remains authoritative for the approved radius. A future startup redesign would need a controlled client/resource-pack barrier test before changing this ordering.
