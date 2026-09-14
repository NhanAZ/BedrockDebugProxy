# BedrockDebugProxy patch

This directory is based on Sandertv gophertunnel `v1.61.0` at commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee` and retains its MIT license.

The project-specific source changes are in `minecraft/conn.go`, `minecraft/dial.go`, `minecraft/packet.go`, `minecraft/raw_packet.go`, and `minecraft/raknet.go`. The connection file treats `PlayStatusPlayerSpawn` as the completed game-data signal for featured experiences that omit `ChunkRadiusUpdated`, including the Enchanted sequence recorded by BedrockDebugProxy. The dial and RakNet files expose the optional client GUID, source-address, and MTU settings used by the proxy when following a `Transfer`. The raw packet adapter exposes an owned read and write boundary so a same-protocol bridge can forward the exact decoded packet payload without re-marshalling it. These hooks do not mutate forwarded Bedrock packet bytes.

The focused regression test is `minecraft/conn_featured_experience_test.go`. See `docs/research/bedrock-ecosystem.md` and `THIRD_PARTY_NOTICES.md` in the parent project for evidence and provenance.
