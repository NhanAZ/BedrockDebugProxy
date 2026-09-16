# AUDIT-0009 - Align packed PlayerAuthInput item interaction

- Commit subject: `Fix packed PlayerAuthInput hand alignment`
- Change type: `protocol | runtime | documentation`
- Runtime impact: `yes`

## Intent

Correct the Bedrock 1.26.50 packed item-use transaction path so `PlayerAuthInput` preserves the `Hand` byte and does not shift every field that follows it.

## Scope

The change adds the missing `Hand` read and write operation to the in-tree gophertunnel `PlayerInventoryAction` implementation, adds a deterministic round-trip regression test, and records the capture-backed protocol finding in the research and fix-history documents. It does not change capture storage, packet forwarding policy, resource-pack handling, transfer behavior, or authentication.

## Evidence and provenance

- Mojang's released [`ItemUseInventoryTransaction` schema](https://github.com/Mojang/bedrock-protocol-docs/blob/v1.26.50/json/ItemUseInventoryTransaction.json) requires the hand field between the hotbar slot and item descriptor.
- Mojang's [`PackedItemUseLegacyInventoryTransaction` schema](https://github.com/Mojang/bedrock-protocol-docs/blob/v1.26.50/json/PackedItemUseLegacyInventoryTransaction.json) embeds that item-use structure in `PlayerAuthInput`.
- The selected gophertunnel 1.26.50 feature revision added `Hand` to `UseItemTransactionData.Marshal` but omitted it from the packed `PlayerInventoryAction` reader and writer. The source revision is [`e95f6c6026d55625ec40089abac3c0131cb49355`](https://github.com/Sandertv/gophertunnel/commit/e95f6c6026d55625ec40089abac3c0131cb49355).
- Local ignored capture `session-20260916T002518Z` contained a The Hive `PlayerAuthInput` body whose decoded fields shifted and left four bytes before this fix. The raw capture remains outside Git.
- The detailed analysis is in [`docs/research/protocol-1.26.50.md`](../research/protocol-1.26.50.md).

## Validation

- Automated checks: `pass - tools/format.ps1; tools/quality.ps1; nested gophertunnel go test ./...; focused protocol tests including TestPlayerInventoryActionRoundTripIncludesHand; git diff --check`
- Live validation: `pending - rebuild a stamped binary and repeat the The Hive flow`
- Capture or report: `session-20260916T002518Z` is diagnostic evidence only; a new exact-revision report is required.

## Limitations and follow-up

This is a focused serializer alignment fix. It does not infer that all Bedrock servers use one packet order, and it does not replace the required live compatibility checks. After committing, build a revision-matched binary and verify The Hive before proceeding to the remaining release matrix.
