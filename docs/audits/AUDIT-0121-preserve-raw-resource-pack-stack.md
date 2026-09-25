# Audit AUDIT-0121 - Preserve raw ResourcePackStack bytes

## Intent

Preserve the exact upstream `ResourcePackStack` packet across a terminating bridge when the downstream protocol is the same, while retaining a typed fallback for protocol changes or invalid retained bytes.

## Scope

- Added raw stack bytes and source protocol tracking to gophertunnel `GameData` and listener connection state.
- Added a listener callback that returns the retained bytes and protocol ID.
- Updated BedrockDebugProxy to supply the upstream retained stack to its downstream listener.
- Added unit and loopback integration assertions for semantic and wire-level preservation.
- Updated the resource-pack research note, fix history, and changelog.

## Evidence and provenance

- Local research note: [`docs/research/resource-pack-stack-parity.md`](../research/resource-pack-stack-parity.md).
- Existing Hive capture evidence: [captures/session-20260920T135412Z](https://github.com/BedrockDebugProxy-Captures/session-20260920T135412Z/releases/tag/capture-archive), sequences 979 and 1,183. Its archive is stored privately and remains sensitive.
- Packet serializer: vendored gophertunnel `minecraft/protocol/packet/resource_pack_stack.go` at the repository-pinned source revision.
- The change does not retrieve, decrypt, or redistribute resource-pack content. The downstream offer still uses archive-backed URLs owned by the local bridge.

## Validation

- `go test ./minecraft -run 'Test(HandlePreservesResourcePackStackForBridge|ResourcePackResponseWritesRetainedStackBytes|HandlePreservesRawStartupPacketsForBridge)' -count=1`
- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v`
- Full project quality and build checks are run before commit.
- No live or public server validation was performed in this change because the owner had ended the prior local session. Real-client validation remains pending.

## Limitations and follow-up

The raw path is limited to matching protocol IDs and a validated `ResourcePackStack` packet ID. A protocol mismatch selects the typed stack. If packet validation fails, the typed stack is sent and a warning is logged. `ResourcePacksInfo` remains a typed rebuild boundary because URLs must resolve to locally served archives. A future owner-controlled real-client resource-pack session should verify mixed and encrypted offers against the stamped binary.
