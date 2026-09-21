# Audit AUDIT-0089 - Assert startup raw capture parity

- Commit subject: `Assert startup raw capture parity`
- Change type: `test`
- Runtime impact: `no`

## Intent

Strengthen the owner-controlled loopback integration test so it compares raw packet hashes between the upstream and downstream capture channels for the complete startup boundary, not only decoded `GameData` fields.

## Scope

- Add `ResourcePacksInfo`, `ResourcePackStack`, `StartGame`, `ItemRegistry`, `JigsawStructureData`, and `VoxelShapes` to the existing raw-channel assertions.
- Keep the test fixture local, synthetic, and unauthenticated. No production code or packet behavior changes.
- Reuse the canonical capture blobs and hash comparison already exercised by the integration test.

## Evidence and provenance

- The test fixture uses the owner-controlled RakNet loopback listener and the existing capture verifier. It supplies deterministic startup packets and archive-free resource-pack metadata.
- The assertion passed with the raw startup retention implementation at commit `a86b2d7`.

## Validation

- `go test ./internal/proxy -run TestRunnerForwardsBidirectionalPacketsAndCapturesSession -count=1 -v` passed.
- Live vanilla-client validation: `not_observed`.

## Limitations and follow-up

The fixture does not prove parity for URL-advertised or encrypted resource packs, protocol conversion, authenticated login chains, or post-spawn traffic. Those remain separate validation boundaries.
