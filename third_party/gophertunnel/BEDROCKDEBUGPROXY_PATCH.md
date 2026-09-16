# BedrockDebugProxy patch

This directory retains Sandertv gophertunnel's MIT license and is a reviewed composite of:

- stable `v1.61.0` at commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`
- the Bedrock 1.26.50 feature branch through commit `481f3bd138766304a73d7a0412a47a87acec15ed`
- the required `ClientboundUpdateSoundData` correction from master through commit `b8bd7357c24fcba3e32f940d1afb1d3b4e43b96d`
- the final upstream 1.26.50 merge `v1.62.0` at commit `7a556a07335b663744b50d38062636ad8283f314`, including the reviewed `HeightMap` correction from `709768d`

The feature branch diverged before the final `v1.61.0` fixes. Its reviewed protocol changes and the final `v1.62.0` corrections were merged onto the stable base instead of replacing the tree. This preserves the stable optional `FilteredCustomName` item-stack encoding while updating the selected protocol to Bedrock 1.26.50, protocol 2193. The packed `PlayerInventoryAction` `Hand` correction from `d564b7d` was already present in the local source and remains covered by the parent project's regression test, so no duplicate patch was added.

The project-specific source changes are in `minecraft/conn.go`, `minecraft/dial.go`, `minecraft/packet.go`, `minecraft/raw_packet.go`, and `minecraft/raknet.go`. The connection file treats `PlayStatusPlayerSpawn` as the completed game-data signal for featured experiences that omit `ChunkRadiusUpdated`, including the Enchanted sequence recorded by BedrockDebugProxy. The dial and RakNet files expose the optional client GUID, source-address, and MTU settings used by the proxy when following a `Transfer`. The raw packet adapter exposes an owned read and write boundary so a same-protocol bridge can forward the exact decoded packet payload without re-marshalling it. These hooks do not mutate forwarded Bedrock packet bytes.

The focused local compatibility regression test is `minecraft/conn_featured_experience_test.go`. The parent project's `internal/bedrock/protocol_update_test.go` verifies the selected protocol metadata, new packet IDs and pools, and exact new-packet bodies. See `docs/research/protocol-1.26.50.md`, `docs/research/bedrock-ecosystem.md`, and `THIRD_PARTY_NOTICES.md` in the parent project for evidence and provenance.
