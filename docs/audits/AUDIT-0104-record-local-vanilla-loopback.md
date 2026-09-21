# Audit 0104 - Record local vanilla loopback evidence

- Intent: record the post-fix real-client loopback evidence for startup packet parity instead of leaving the result only in an operator terminal or ignored capture.
- Scope: research documentation only. Added the controlled capture result to `docs/research/start-game-parity.md`.
- Affected interfaces: none. No Go code, capture schema, protocol behavior, binary output, or runtime configuration was changed.
- Evidence and provenance: exact-revision capture `captures/session-20260921T122228Z`, analysis output, and raw packet hash comparison for IDs 313, 337, 11, 162, and 145. The local PocketMine Xbox-auth setting was restored to `on` after the run.
- Automated validation: `tools/quality.ps1 -LowMemory` passed before this documentation update. `git diff --check` is required before commit.
- Local validation: the real Windows Bedrock client completed the owner-controlled local login, resource-pack negotiation, spawn, play, and quit flow. The capture verified 10,645 events, 4,222 blobs, zero drops, zero write errors, zero truncation, and zero decode errors.
- Live validation: no public server was contacted. The Hive and other release-gate reports remain pending and are not implied by this local evidence.
- Known limitations: sequence positions across the two login chains are not directly comparable. Raw UDP and RakNet recovery frames remain outside the observation boundary, and mixed URL resource-pack offers still need a controlled real-client test.
- Follow-up: use this capture as the local baseline when evaluating future startup or resource-pack changes, and keep public validation separate from owner-controlled loopback evidence.
