# Audit 0095 - Correct NetherNet research status

## Intent

Correct the ecosystem research note so it matches the current capability-preserving observation wrapper and does not claim that live NetherNet behavior has been validated.

## Scope

- Updated `docs/research/bedrock-ecosystem.md`.
- Updated `docs/fix-history.md` with FIX-0069.
- No runtime code, protocol behavior, capture schema, dependency, or CLI change.

## Evidence and provenance

- `internal/bedrock/network.go` delegates `BatchHeader`, `DisableEncryption`, and `ReadPacket` through `observedConn`.
- `internal/bedrock/network_test.go` covers the wrapper capability contract and packet-read capture path.
- `internal/experience/target.go` selects the authenticated NetherNet transport for resolved NetherNet targets.
- The vendored NetherNet implementation declares `packet.TransportCapabilities` on its connection type.

## Automated validation

- `go test ./internal/bedrock`
- `git diff --check`

## Live validation

Not performed. A live controlled NetherNet capture remains pending.

## Known limitations and follow-up

This change does not add NetherNet live testing or raw WebRTC/UDP capture. Do not infer live transport behavior from the wrapper tests alone.
