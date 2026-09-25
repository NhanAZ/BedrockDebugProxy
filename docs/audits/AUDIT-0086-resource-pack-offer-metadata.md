# Audit AUDIT-0086 - Preserve upstream resource-pack offer metadata

- Commit subject: `Preserve upstream resource-pack offer metadata`
- Change type: `runtime`
- Runtime impact: `yes, bounded resource-pack login negotiation`

## Intent

Preserve the decoded upstream `ResourcePacksInfo` metadata across the terminating bridge so a downstream vanilla client sees the upstream offer policy while the proxy continues to deliver archive-backed packs through its own connection.

## Scope

- Retain a cloned `ResourcePacksInfo` on dialled `GameData`.
- Add a listener callback for per-connection offer metadata.
- Have BedrockDebugProxy mirror the upstream metadata and rebuild only the downstream `TexturePacks` entries it can serve.
- Keep the configured listener policy as the fallback when no upstream offer is available.
- Do not forward stale URL entries, mutate gameplay packets, drop packets, inject packets, replay packets, bypass authentication, or implement server-policy logic.

## Evidence and provenance

- Owner-controlled PocketMine capture [captures/session-20260921T003441Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T003441Z/releases/tag/capture-archive) recorded an upstream `ResourcePacksInfo` payload beginning with `0x00` and a pre-fix downstream payload beginning with `0x01`. Its raw archive is stored privately and is excluded from this project's source history.
- The packet field order is defined by the vendored gophertunnel `minecraft/protocol/packet/resource_packs_info.go` implementation.
- The boundary and remaining URL-delivery constraint are documented in [`docs/research/resource-pack-negotiation.md`](../research/resource-pack-negotiation.md) and [`docs/resource-packs.md`](../resource-packs.md).

## Automated validation

- Nested `go test ./minecraft -run 'TestHandlePreservesResourcePackStackForBridge|TestHandleDrainsResourcePacksInfoDeferredBeforePlayStatus' -count=1` passed.
- `go test ./internal/proxy -run TestRunnerForwardsBidirectionalPacketsAndCapturesSession -count=1` passed, including non-default optional-offer metadata checks.
- `gofmt` and `git diff --check` passed before the full quality gate.
- Full `tools/quality.ps1` remains required before commit.

## Live validation

No public server was contacted. The loopback integration test exercises the listener and dialler resource-pack metadata path. A fresh exact-revision real-client report on an owner-controlled server remains required for any release claim.

## Known limitations and follow-up

URL offers still require archive-backed local delivery, so the downstream packet does not preserve upstream download URLs. Pack ordering and stack metadata are handled by the separate `ResourcePackStack` parity path. Authenticated login and other startup metadata remain separate validation boundaries.
