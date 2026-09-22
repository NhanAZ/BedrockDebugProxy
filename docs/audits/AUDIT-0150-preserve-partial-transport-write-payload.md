# Audit 0150 - Preserve partial transport write payload boundaries

## Intent

Ensure raw transport payload evidence never presents bytes that a partial write reported as unsent.

## Scope

- Updated `internal/bedrock/network.go` so `observedConn.Write` records only the transferred prefix of a payload.
- Added deterministic coverage in `internal/bedrock/network_test.go` for a partial write with an underlying transport error.
- Updated `docs/research/capture-observability.md`, `docs/fix-history.md`, and `CHANGELOG.md`.
- No wire behavior, forwarding policy, capture schema, or protocol definition changed.

## Evidence and provenance

- The Go `net.Conn` contract permits `Write` to return fewer bytes than requested together with an error.
- Before this change, write metadata preserved both counts but the raw blob used the full attempted buffer, which could make unsent bytes appear to be wire evidence.
- The regression test uses synthetic bytes and a local fake transport. No credentials, captures, or external endpoints are involved.

## Automated validation

- Focused `internal/bedrock` regression tests pass.
- The full project quality gate with `-LowMemory` and `git diff --check` are required before commit.

## Live validation status

No live or public-server validation was performed. Partial transport failure behavior is covered only by deterministic local tests.

## Known limitations and follow-up

The wrapper records application payloads above the reliable transport. It does not expose raw UDP datagrams, retransmissions, or lower-layer fragmentation. The existing live validation matrix remains pending independently of this capture-boundary fix.
