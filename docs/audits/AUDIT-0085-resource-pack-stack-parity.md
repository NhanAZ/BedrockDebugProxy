# AUDIT-0085 - Preserve upstream resource-pack stack

- Commit subject: `Preserve upstream resource-pack stack`
- Change type: `runtime`
- Runtime impact: `yes, bounded resource-pack login negotiation`

## Intent

Preserve the upstream `ResourcePackStack` across the terminating bridge so a downstream vanilla client receives the same stack metadata and pack order.

## Scope

- Retain a cloned stack on dialled `GameData`.
- Add a listener callback for an exact per-connection stack.
- Have BedrockDebugProxy provide the retained upstream stack before downstream stack negotiation completes.
- Keep the existing generated stack as the fallback for standalone listener callers.
- No gameplay packet mutation, dropping, injection, replay, authentication bypass, or server-policy logic was added.

## Evidence and provenance

- Local Hive capture [captures/session-20260920T135412Z](https://github.com/BedrockDebugProxy-Captures/session-20260920T135412Z/releases/tag/capture-archive) recorded a 1,449-byte upstream stack and a 1,446-byte downstream stack. The serialized prefix showed 24 upstream texture-pack entries and 32 downstream entries because the old listener appended eight built-in exempted packs.
- Packet field order is defined by the vendored gophertunnel `ResourcePackStack` packet. The immutable upstream reference is [`resource_pack_stack.go`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/protocol/packet/resource_pack_stack.go), reviewed 2026-09-21.
- Detailed reasoning is recorded in [`docs/research/resource-pack-stack-parity.md`](../research/resource-pack-stack-parity.md).

## Automated validation

- Nested `go test ./minecraft -run '^TestHandlePreservesResourcePackStackForBridge$' -count=1 -v` passed.
- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed, including a non-default upstream stack and downstream equality checks.
- Full `tools/quality.ps1` passed, including gofmt, PowerShell and viewer syntax, module checks, all tests, build, static analysis, vulnerability scan, tracked-capture checks, and whitespace checks.
- `git diff --check` passed before commit.

## Live validation

No public server was contacted. The loopback integration test exercises the complete listener and dialler resource-pack stack path. A fresh exact-revision real-client report remains required for any release claim.

## Known limitations and follow-up

`ResourcePacksInfo` optional-policy flags remain an explicit downstream policy boundary. URL-delivered packs still need archive-backed local delivery. Authenticated login chains and other startup metadata remain separate parity boundaries.
