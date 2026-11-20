# AUDIT-0261 - Improve Go Golem experience acquisition

- Commit subject: `Improve Go Golem experience acquisition`
- Change type: `runtime, protocol, capture compatibility`
- Runtime impact: `yes`

## Intent

The Creator Experience acquisition workflow depends on reading server forms and lobby state, following observed routes, and retaining an accurate canonical capture. Investigation exposed missing Golem actions and a strict decoder limit that marked large packet arrays as errors. This change adds behavior supported by the current form and transfer paths and keeps the capture schema unchanged.

## Scope

- Go Golem now uses a currently observed hotbar stack, acknowledges server `MovePlayer` teleport and reset instructions for the active player, and follows service NetherNet Transfer IDs through the existing signaling network.
- The form view accepts string or structured `rawtext` title, content, and button labels. It accepts `content` when `text` is absent and retains the raw JSON.
- Strict protocol decoding uses a finite 65,536-item bound for the `AvailableCommands` collections and `UpdateSubChunkBlocks` block-change lists. The generic 1,024-item bound remains in place for all other collections, and declared counts must fit the remaining packet payload.
- Capture schema, resource-pack record fields, and forwarding behavior are unchanged.

## Evidence and provenance

- The packet-bound decision is reviewed in [`docs/research/large-command-and-subchunk-slices.md`](../research/large-command-and-subchunk-slices.md) against Mojang's pinned protocol 2225 schemas.
- Go Golem teleport and item-use protocol details are in [`golem-server-teleport-ack.md`](../research/golem-server-teleport-ack.md) and [`golem-observed-item-use.md`](../research/golem-observed-item-use.md).
- Service NetherNet route handling is in [`nethernet-transfer-id-routing.md`](../research/nethernet-transfer-id-routing.md). Nonzero-port and ordinary RakNet Transfer behavior keep their prior route selection.
- Paired BedrockPackArchive commit: [`7cd45b5`](https://github.com/NhanAZ-Tools/BedrockPackArchive/commit/7cd45b5).
  Its fixtures cover both proxy-run and Go Golem capture identities while
  preserving the same pack archive and index contract.

## Validation

- Automated checks: `pass` locally with `tools/quality.ps1`, including tests, build, static analysis, vulnerability scan, artifact checks, and `git diff --check`. Focused packet and form tests also pass.
- BedrockPackArchive: full Go tests, `go vet`, build, and `git diff --check` pass. Its importer saved the five lobby offers from the earlier CubeCraft capture. The compatibility test against this exact stamped producer revision remains pending.
- Live validation: `pending` on a stamped binary for the exact runtime revision.
- Capture or report: `pending`; do not use the previous failed Rob a Mob capture as a snapshot source.

## Limitations and follow-up

This commit does not establish complete pack or child-route coverage for the 14 Creator Experiences. Run a clean stamped Golem session, preserve and verify the capture, inspect the server's current forms and entities, follow observed Transfers, and import its pack events with BedrockPackArchive. Keep CubeCraft at lobby spawn and close immediately. Do not advance a current snapshot until the exact producer revision and the complete offer data are verified.
