# Audit AUDIT-0084 - Preserve startup world packets

- Commit subject: `Preserve startup world packets`
- Change type: `runtime`
- Runtime impact: `yes, bounded login startup packet handling`

## Intent

Prevent duplicate and empty startup `JigsawStructureData` and `VoxelShapes` packets when the upstream server supplies those packets during login.

## Scope

- Retain decoded packet copies in `GameData`.
- Accept them in the bounded login states after resource-pack negotiation.
- Reuse the retained packets in the listener-side startup sequence.
- Add unit and loopback integration coverage.
- No gameplay packet mutation, dropping, injection, replay, authentication bypass, or server-policy logic was added.

## Evidence and provenance

- Local capture `captures/session-20260921T005313Z` recorded upstream packet IDs 313 and 337 before the generated downstream copies and later deferred copies.
- The affected Hive capture `captures/session-20260920T135412Z` recorded the same duplicate shape. The captures remain local sensitive evidence and are not added to Git.
- Packet definitions and login handling come from the vendored gophertunnel composite already governed by the repository's provenance records.

## Automated validation

- Nested gophertunnel `go test ./...` passed.
- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed.
- `git diff --check` passed before commit.
- Full `tools/quality.ps1` passed, including tests, build, static analysis, vulnerability scan, capture checks, and whitespace checks.

## Live validation

No public server was contacted for this change. The loopback integration test exercises the complete startup path. Existing real-client captures provide observational evidence but are not a fresh exact-revision live report.

## Known limitations and follow-up

Packets that first arrive after the connection is fully spawned remain ordinary forwarded traffic. Authenticated identity, separate encryption keys, resource-pack delivery, and other login packets remain separate parity boundaries.
