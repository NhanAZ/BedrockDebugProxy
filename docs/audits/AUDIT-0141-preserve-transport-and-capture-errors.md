# Audit AUDIT-0141

## Intent

Preserve underlying transport errors and partial transfer counts when capture submission fails at the observing connection boundary.

## Scope

- Updated `internal/bedrock/network.go` so `Read`, `ReadPacket`, and `Write` join transport and recorder errors without discarding transferred data counts or payloads.
- Added `TestObservedConnPreservesTransportErrorWhenCaptureFails` in `internal/bedrock/network_test.go`.
- Updated `docs/research/capture-observability.md`, `docs/fix-history.md`, and the Unreleased changelog.

## Evidence and provenance

The observing wrapper receives both the transport operation result and the recorder result. The previous branches returned only the recorder error, which hid the lower-layer cause and returned zero for a partial `Read`. The fix changes only error and count propagation. Raw bytes and event metadata remain unchanged. No external source code was copied.

## Automated validation

- Focused network tests, including the combined-error regression, are required before commit.
- Full `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit.

## Live validation

No live server, public service, or external account was contacted. Real transport-failure behavior remains pending.

## Known limitations and follow-up

The change does not add raw UDP or RakNet datagram observation. It only preserves error provenance at the existing `net.Conn` wrapper boundary.
