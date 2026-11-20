# Golem use of observed hotbar items

## Scope

This note records the protocol basis for the bounded `golem network` `use_item`
scenario operation. The operation can use one nonempty hotbar item already
received on the current connection. It does not infer inventory contents,
replay packet bytes, move the player, or interact with an entity.

## Protocol evidence

The stable Mojang `bedrock-protocol-docs` release `v1.26.62` identifies network
protocol `2225`. Its [`PlayerAuthInputPacketPayload.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/v1.26.62/json/PlayerAuthInputPacketPayload.json)
defines an optional item-use transaction field at ordinal 11. The matching
[`PlayerAuthInputData.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/v1.26.62/json/PlayerAuthInputData.json)
defines `PerformItemInteraction`, which gates that field. The
[`ItemUseInventoryTransaction.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/v1.26.62/json/ItemUseInventoryTransaction.json)
defines the action, trigger, hotbar slot, item, position, and prediction fields.
These schemas are the official protocol source, not copied project code.

The selected implementation uses the repository's pinned Gophertunnel packet
types in `third_party/gophertunnel/minecraft/protocol`. It first sends
`MobEquipment` with the exact nonempty item instance observed in an
`InventoryContent` or `InventorySlot` packet. It then sends one
`PlayerAuthInput` packet with `PerformItemInteraction` and a
`UseItemActionClickAir` transaction for that same slot and item. This follows
the current client-side protocol representation and keeps item selection tied
to observed server state.

## Compatibility decision

The repository's current stable protocol baseline is `2225` for Bedrock
`1.26.62`. The official schema contains the fields required for this action,
and the local codec exposes those fields. No protocol baseline or capture
schema change is needed. A focused loopback test verifies that Golem uses the
observed item and emits the transaction under the matching input flag.

No upstream implementation code was copied or adapted for this operation.
