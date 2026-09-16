# Bedrock 1.26.50 protocol update

Reviewed on 2026-09-16.

## Decision

BedrockDebugProxy identifies the current protocol as Bedrock `1.26.50`, protocol `2193`.

The in-tree gophertunnel source is a reviewed composite rather than a direct checkout of one branch:

- stable base `v1.61.0` at [`283a5a97dfe65da94bcc0b401807f6aefa9e72ee`](https://github.com/Sandertv/gophertunnel/tree/283a5a97dfe65da94bcc0b401807f6aefa9e72ee)
- upstream `feature/26.50` through [`481f3bd138766304a73d7a0412a47a87acec15ed`](https://github.com/Sandertv/gophertunnel/tree/481f3bd138766304a73d7a0412a47a87acec15ed)
- upstream master correction through [`b8bd7357c24fcba3e32f940d1afb1d3b4e43b96d`](https://github.com/Sandertv/gophertunnel/tree/b8bd7357c24fcba3e32f940d1afb1d3b4e43b96d)

The 1.26.50 feature branch diverged before the final `v1.61.0` fixes. Replacing the stable tree wholesale would have regressed the optional `FilteredCustomName` encoding in item stacks. The selected approach merges the reviewed 1.26.50 changes onto the stable base and preserves the project-specific compatibility hooks documented in [`third_party/gophertunnel/BEDROCKDEBUGPROXY_PATCH.md`](../../third_party/gophertunnel/BEDROCKDEBUGPROXY_PATCH.md).

## Evidence table

| Claim | Primary evidence | Independent implementation evidence | Decision |
| --- | --- | --- | --- |
| Bedrock 1.26.50 uses protocol 2193 | Mojang [`bedrock-protocol-docs` release `v1.26.50`](https://github.com/Mojang/bedrock-protocol-docs/tree/v1.26.50) at `c0bd91f7d896cec780f1185cc548b5e46a46f5d5`. Released schemas declare `x-protocol-version` 2193. | gophertunnel commit [`18485d3847098cd1bd8ba0d0644e816d3200337a`](https://github.com/Sandertv/gophertunnel/commit/18485d3847098cd1bd8ba0d0644e816d3200337a) changes the release protocol to 2193. | Set `CurrentProtocol` to 2193 and `CurrentVersion` to `1.26.50`. |
| Packet IDs 351 and 352 represent `SetPlayerFurnaceOptions` and `RecordStarted` | Mojang [`MinecraftPacketIds.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/v1.26.50/json/MinecraftPacketIds.json) and the matching packet schemas. | gophertunnel `feature/26.50` packet definitions and pools. | Register both packets in the direction-specific pools and add exact-body encoding tests. |
| Existing packet and structure layouts changed in 1.26.50 | Mojang 1.26.50 released schemas for the affected packets and types. | gophertunnel commits from [`e95f6c6026d55625ec40089abac3c0131cb49355`](https://github.com/Sandertv/gophertunnel/commit/e95f6c6026d55625ec40089abac3c0131cb49355) through `481f3bd138766304a73d7a0412a47a87acec15ed`. | Import the reviewed definitions, IDs, readers, writers, and value types without changing capture schema. |
| `ClientboundUpdateSoundData` has seven required fields in this release | Mojang 1.26.50 `ClientboundUpdateSoundDataPacket` schema. | gophertunnel master commit `b8bd7357c24fcba3e32f940d1afb1d3b4e43b96d`. | Include the required-field correction that is newer than the feature branch head. |
| NetherNet uses the vanilla identity domain and a URL-shaped server address | Mojang's [Minecraft Bedrock Edition 26.50 changelog](https://feedback.minecraft.net/hc/en-us/articles/48826825649933-Minecraft-Bedrock-Edition-26-50-Changelog-Wilderness-Bound) states that NetherNet is the default networking protocol for dedicated servers. | gophertunnel commit `481f3bd138766304a73d7a0412a47a87acec15ed` changes the identity domain and server-address construction. | Import the transport update while preserving BedrockDebugProxy transfer settings. |

## Resolved source disagreement

Cloudburst Protocol branch `3.0` at `9df9864c2bf79197a30ddff72dc0008dca93a976` listed Minecraft 1.26.50 as protocol 2192 during this review. That conflicts with Mojang's released 1.26.50 schemas and the later gophertunnel release correction, both of which identify protocol 2193. The project selects protocol 2193 because the released Mojang schema is the primary definition for this claim.

PrismarineJS minecraft-data and axolotl-pm BedrockProtocol still advertised 1.26.45 as their newest supported version when reviewed. They are useful comparisons for shared structures but do not provide contradictory 1.26.50 evidence.

## Imported scope

The update includes current protocol metadata, new furnace and recording packet definitions, packet IDs and pools, changed packet and value serializers, corrected subchunk sizing, reader and writer support, changed disconnect and input values, and the corresponding NetherNet transport update. It does not change the BedrockDebugProxy capture schema or remove unknown-packet and raw-payload representation.

The root module retains `github.com/sandertv/gophertunnel v1.61.0` as the declared dependency because the `replace` directive selects the documented in-tree composite. Labeling the feature branch as a later module release would be misleading because it diverged before `v1.61.0`.

## Provenance and license

Sandertv gophertunnel code is MIT licensed. Its license remains in the in-tree source, and the exact imported revisions are recorded above and in [`THIRD_PARTY_NOTICES.md`](../../THIRD_PARTY_NOTICES.md).

Mojang `bedrock-protocol-docs` was used as a protocol definition and validation source. Its repository states that the material is subject to Mojang's terms. No Mojang source file was copied into BedrockDebugProxy.

Cloudburst, PrismarineJS, Endstone, axolotl-pm, Altay, and BetterAltay were comparison sources only for this change. No code from those repositories was copied or adapted.

## Validation status

- The pre-update project quality gate passed at commit `be9ac1c`.
- Focused tests verify protocol 2193, game version 1.26.50, packet IDs, direction-specific packet pools, and exact new-packet bodies.
- The full project quality gate and nested gophertunnel test suite pass on the prepared source tree.
- A stamped exact-revision The Hive report is required before this runtime change is complete.
