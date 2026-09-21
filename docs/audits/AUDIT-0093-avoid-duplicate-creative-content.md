# Audit AUDIT-0093 - Avoid duplicate generated CreativeContent

- Commit subject: `Avoid duplicate generated CreativeContent`
- Change type: `runtime-fidelity`
- Runtime impact: `narrow startup boundary change for terminating bridges`

## Intent

Avoid adding gophertunnel's empty listener compatibility `CreativeContent` packet when the upstream server packet arrives after spawn and is already preserved by the forwarding path.

## Scope

- Add an opt-in `GameData` listener option and use it only for the terminating proxy's downstream connection.
- Preserve the default generated packet for ordinary listener callers.
- Extend nested and root loopback tests to distinguish the generated packet from the upstream packet and compare raw hashes.
- Update the startup parity research note, fix-history index, changelog, and vendored patch provenance.

## Evidence and provenance

- Owner-controlled capture `captures/session-20260921T005313Z` recorded upstream packet ID 145 at sequence 65 with a 49,283-byte payload, downstream generated packet ID 145 at sequence 455 with a 2-byte payload, and the original 49,283-byte payload at sequence 1077.
- Vendored gophertunnel `minecraft/conn.go` showed the unconditional listener response in `handleRequestChunkRadius`.
- The change does not copy external source and does not contact a public server.

## Automated validation

- Nested gophertunnel suppression test passed.
- Root proxy loopback integration test passed and verified decoded CreativeContent plus matching upstream and downstream raw hashes.
- Full `tools/quality.ps1`, `go test ./...`, `git diff --check`, and the nested module test suite must pass before commit.

## Live validation status

`not_run`. Direct vanilla-client validation remains pending in an owner-controlled local environment.

## Known limitations and follow-up

The option suppresses only the proxy-generated empty compatibility packet. It does not suppress, mutate, or reorder an upstream CreativeContent packet. Packets arriving after the supported login boundary remain ordinary forwarded traffic.
