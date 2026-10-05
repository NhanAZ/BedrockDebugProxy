# AUDIT-0201 - Preserve empty pack download-request state

- Commit subject: `fix: preserve state for empty resource-pack download requests`
- Change type: `runtime | resource packs`
- Runtime impact: `yes`

## Intent and root cause

A `PackResponseSendPacks` with no requested packs created or replaced the resource-pack queue, then failed with `no resource packs to download`. The response does not select an archive to transfer and must leave the existing exchange state intact.

## Scope and provenance

Backport only the empty-list guard from MIT upstream revision `e86a90de9b1087c17952d1d123b175a89c214d5d` in `minecraft/conn.go`. Add a regression before and during an active download, and update the patch description, notices, fix history, and changelog. This is copied upstream source with its existing license retained. The [maintenance review](../research/protocol-maintenance-2026-10-05.md) records the source review.

The capture schema, raw packet observer, archive fields, derivation events, ContentKey handling, and per-file key handling are unchanged. BedrockPackArchive's paired fixture extends the synthetic capture with producer-encoded empty request bytes, canonical blob paths, event schemas, and manifest counters. It verifies both manifests, the archive digest, contents.json, key count, and index source sequence. No importer production code change is needed because the consumer contract is unchanged.

The matched consumer commit is `1b1b8b4a156453d444d403d27989a7f398e8b5cf` in the existing private `NhanAZ-Tools/BedrockPackArchive` repository. The paired final audit records this producer commit's exact SHA after it exists.

## Validation

- Automated checks: the previous source failed both empty-request cases. The new cases pass and preserve the queue pointer, expected packet IDs, and incomplete login state. The focused replacement-module suite passed. The consumer passed `go test -count=1 ./...`, `go vet ./...`, build, and whitespace checks. The extended fixture passed through the actual producer verifier built from the reviewed working source. `tools/format.ps1`, `tools/quality.ps1 -LowMemory`, and `git diff --check` passed on the reviewed maintenance source. Stamped fixture verification is recorded in the paired final audit after commit identities are available.
- Live validation: `not_observed`.
- Capture or report: none. Pack and capture fixtures are generated from synthetic source data.

## Human validation and limitations

Use an owner-controlled Bedrock server with a cached resource pack, or an authorized Experience offering packs. Exercise the resource-pack exchange and graceful disconnect. An empty download request should not terminate login or replace an active transfer, and subsequent archive and stack evidence should remain verifiable. Return the closed capture and first pack error if exchange stalls. The paired final audit will cross-reference both repository commits. Retail-client and release-matrix validation remain unobserved.
