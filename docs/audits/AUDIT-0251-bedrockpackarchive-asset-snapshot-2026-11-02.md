# AUDIT-0251 - BedrockPackArchive asset snapshot for v0.5.0

- Commit subject: `Record the v0.5.0 private asset snapshot`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Record the post-release archive handoff required by the private NhanAZ-Assets
intake contract. The handoff uses a fresh BedrockPackArchive output produced
with the validated BedrockDebugProxy v0.5.0 revision and keeps the asset
repositories private.

## Producer and consumer revisions

- BedrockDebugProxy release commit: `2029eee910f39b635f5c88803d7cc9ed669a19b6`.
- BedrockDebugProxy validated runtime revision: `b2a6651209f2460cc1b44132a9f027fca7855ebd`.
- BedrockPackArchive route-policy commit: `d17af54c725008a817c069d3450748e9f1daeffe`.
- BedrockPackArchive refreshed-button fallback commit: `aa654737e06bafc3d23b3495944484a5e5fe5510`.

The BedrockPackArchive changes exclude the retired Enchanted Steal a Mob route
and allow a unique current-label fallback when a refreshed Enchanted form
changes a button's stable identity. Ambiguous label matches remain rejected.

## Inventory and routing

- The output contained 12 Experience directories with 13 logical destinations;
  Treasure Hunt was routed to the Enchanted repository as required.
- The validated current set contains 216 deduplicated index rows across the
  11 updated repositories.
- Enchanted covers Hub, Skyblock Factory, OP SMP, Box PvP, BedWars SMP, Mines,
  Fishing, Dragons, and Treasure Hunt. Repeated Hub visits were deduplicated;
  the retired Steal a Mob route was excluded.
- Mob Maze was normalized to its established single-pack repository root.
- Mineville mode-specific PvP packs retain the established Shared, DUELS, and
  BEDWARS routing; shared Hub packs remain at the repository root.
- Dimension Clash was re-tried with the release binary. Its child transfer
  still supplied an unroutable target, so its existing repository main was not
  advanced and no incomplete child pack was imported.

## Private repository refs

Each updated repository received the immutable snapshot branch
`snapshots/2026-11-02-02`, the matching lightweight tag
`backup-2026-11-02-02`, and a fast-forwarded `main` at the same commit. The
repositories remain private and their default branch remains `main`.

## Validation

- Every imported `archive-index.jsonl` row parsed and resolved to its intended
  destination path.
- All listed `contents_json` paths existed; all present manifests parsed.
- No encrypted output files, raw captures, packet traces, credentials, or
  debug logs were imported.
- No imported file exceeded 100 MiB.
- The private repository working trees were clean after push, and remote
  branch/tag refs matched the local commit for every updated destination.
- BedrockPackArchive `go test -count=1 ./...`, `go vet ./...`, and the release
  binary build passed after the route-policy and fallback changes.

## Limitations

This audit records private asset storage and the producer/consumer handoff. It
does not claim that the unresolved Dimension Clash child transfer is healthy,
and it does not store raw packet captures, server addresses, account state, or
resource-pack keys.

## Security and privacy

All destination repositories are private. The private handoff instructions
were used as an operational contract and their contents are not reproduced in
this public repository.
