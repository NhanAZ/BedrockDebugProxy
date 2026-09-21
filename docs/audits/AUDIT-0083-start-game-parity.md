# Audit AUDIT-0083 - Preserve complete StartGame world metadata

- Commit subject: `Preserve complete StartGame world metadata`
- Change type: `runtime`
- Runtime impact: `yes, login-boundary world metadata only`

## Intent

Preserve the complete decoded upstream `StartGame` packet when a terminating bridge sends world-start data to the downstream client. The prior compact `GameData` projection omitted fields and caused the listener serializer to supply proxy defaults.

## Scope

- Added `GameData.StartGamePacket` to retain the decoded packet from a dialled connection.
- Updated listener `StartGame` serialization to reuse the retained packet when available.
- Added unit and loopback integration coverage for non-default world and telemetry fields.
- Updated the changelog, fix history, and research note.
- No packet mutation, dropping, injection, replay, authentication bypass, or server-policy logic was added.

## Evidence and provenance

- The upstream gophertunnel `GameData` wrapper and `StartGame` definition were reviewed at immutable commit `7a556a07335b663744b50d38062636ad8283f314`.
- The project uses its vendored composite under `third_party/gophertunnel`; no external source was copied into the runtime change.
- The changed fields are decoded from the server's current connection and reused only for that downstream session.

## Automated validation

- Nested `go test ./...` in `third_party/gophertunnel` passed.
- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed.
- Root `go test ./...` passed.
- `git diff --check` passed before commit.
- Full `tools/quality.ps1` remains required before committing.

## Live validation

No public server was contacted. Owner-controlled loopback validation is covered by the integration test. A real-client validation of the exact current binary remains separate from this commit.

## Known limitations and follow-up

Separate login packets and authenticated identity remain distinct parity boundaries. Future changes must preserve raw evidence and require a focused capture-backed reproduction rather than assuming a universal Bedrock login sequence.
