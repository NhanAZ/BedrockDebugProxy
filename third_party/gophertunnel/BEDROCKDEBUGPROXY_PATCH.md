# BedrockDebugProxy patch

This directory is based on Sandertv gophertunnel `v1.61.0` at commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee` and retains its MIT license.

The only project-specific source change is in `minecraft/conn.go`. It treats `PlayStatusPlayerSpawn` as the completed game-data signal for featured experiences that omit `ChunkRadiusUpdated`, including the Enchanted sequence recorded by BedrockDebugProxy. This allows the existing dialer handshake to complete without changing packet bytes, order, or forwarding behavior.

The focused regression test is `minecraft/conn_featured_experience_test.go`. See `docs/research/bedrock-ecosystem.md` and `THIRD_PARTY_NOTICES.md` in the parent project for evidence and provenance.
