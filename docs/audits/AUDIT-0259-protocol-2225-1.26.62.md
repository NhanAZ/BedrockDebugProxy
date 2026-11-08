# AUDIT-0259 - Bedrock 1.26.62 metadata baseline

## Scope

Update the BedrockDebugProxy selected stable metadata from Bedrock `1.26.61` to `1.26.62` while preserving protocol `2225`, the reviewed 1.26.60 packet codec, capture schema `bedrockdebugproxy.capture.v1`, and all project-specific gophertunnel patches.

## Evidence

- Mojang released [`v1.26.62`](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.62) on 2026-11-05 with network protocol `2225`.
- The immutable release commit is [`f8d1496b5efc95105bd9ae1f2f2f076a43f71d82`](https://github.com/Mojang/bedrock-protocol-docs/commit/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82).
- The selected local codec was compared with the final stable schema changes. The local source already contains `AddVolumeEntity`, the stable audio playback enum, and the stable command packet layouts; preview-only Chat/music and raw-message additions are not present.
- Current upstream gophertunnel still advertises protocol `2225` with game version `1.26.60`, so no upstream dependency bump was imported.
- The detailed evidence table is in [`docs/research/protocol-1.26.62.md`](../research/protocol-1.26.62.md).

## Changes

- Set `minecraft/protocol.CurrentVersion` to `1.26.62`.
- Updated protocol metadata tests, capture fixtures, target metadata fixtures, the README, changelog, third-party notices, and the protocol baseline.
- Added the 1.26.62 research note and this audit record.
- Kept `CurrentProtocol = 2225` and made no packet-code changes.

## Verification

- `.\tools\quality.ps1 -LowMemory` passed: formatting, PowerShell and JavaScript syntax, module tidiness/integrity, license inventory, `go test -count=1 ./...`, build, golangci-lint (`0 issues`), vulnerability scan (`0 vulnerabilities` in reachable code), tracked-artifact checks, and whitespace checks.
- `go run .\cmd\bedrock-debug-proxy version` reported `bedrock-debug-proxy dev - Bedrock 1.26.62 protocol 2225`.
- `.\tools\check-mojang-protocol.ps1` reported `Project baseline: version 1.26.62, protocol 2225` and `No protocol release newer than the project baseline was detected.`
- `git diff --check` passed.
- The final diff contains no changes under `third_party/gophertunnel/minecraft/protocol/packet`; no packet source or preview packet was imported.

## Open completion item

The private BedrockPackArchive checkout was not present in the workspace. Its paired producer-baseline update and compatibility check must be completed separately before treating the protocol workflow as fully complete. Live validation across the required server categories is also pending.
