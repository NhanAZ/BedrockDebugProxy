# Audit 0096 - Correct transport-state limitation metadata

## Intent

Make new capture manifests describe the current transport metadata observation boundary instead of reporting a limitation that the implementation has already removed.

## Scope

- Centralize the CLI limitation list in `captureLimitations`.
- Replace the stale transport-state limitation with the early-login boundary documented by the capture format.
- Add a regression test and a concise changelog entry.
- No packet, wire, schema, or transport behavior changes.

## Evidence and provenance

- `internal/bedrock/network_test.go` verifies transport payload metadata and raw blob retention.
- `internal/proxy/proxy_test.go` verifies negotiated compression metadata on both channels during the login sequence.
- `docs/capture-format.md` defines the valid `not_negotiated`, `enabled`, and `not_enabled` states and the early-payload boundary.

## Automated validation

- `go test ./cmd/bedrock-debug-proxy`
- `git diff --check`
- Full `tools/quality.ps1` is required before commit.

## Live validation

Not performed. This metadata wording change has no wire impact and does not require a live server to establish its scope.

## Known limitations and follow-up

Raw UDP datagrams, RakNet acknowledgements, fragmentation, and retransmission remain outside the canonical observation boundary.
