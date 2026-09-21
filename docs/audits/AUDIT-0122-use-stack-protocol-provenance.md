# Audit AUDIT-0122 - Use stack protocol provenance

## Intent

Ensure the raw `ResourcePackStack` branch uses the source protocol ID returned by the listener callback, rather than the listener's unrelated upstream `GameData` field.

## Scope

- Changed the listener send condition to use `rawResourcePackStackProtocolID`.
- Strengthened the nested regression test with an extra retained byte so typed re-encoding cannot satisfy the assertion.
- Updated the resource-pack research note, fix history, and changelog.

## Evidence and provenance

The callback returns upstream raw bytes and `GameData.RawProtocolID` from `internal/proxy/proxy.go`. The listener stores that protocol ID in dedicated connection state before resource-pack negotiation completes. `writeRawStartupPacket` validates the retained packet ID before `WriteRawPacket`; protocol mismatch still uses the cloned typed stack.

## Validation

- `go test ./minecraft -run '^TestResourcePackResponseWritesRetainedStackBytes$' -count=1 -v`
- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v`
- Full project quality checks and `git diff --check` are required before commit.
- No live or public server validation was performed. The prior local session had ended.

## Limitations and follow-up

The extra-byte fixture proves selection of the raw branch, not that a real client would accept malformed trailing data. It is an internal discriminator only. Real-client mixed and encrypted resource-pack validation remains pending for the stamped binary.
