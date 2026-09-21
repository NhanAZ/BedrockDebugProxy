# Audit AUDIT-0142

## Intent

Preserve a transport close error when connection-open capture fails during observation setup.

## Scope

- Updated `internal/bedrock/network.go` so `Network.wrap` joins the open-event capture error with the cleanup close error.
- Added `TestNetworkWrapPreservesCloseErrorWhenOpenCaptureFails` in `internal/bedrock/network_test.go`.
- Updated `docs/research/capture-observability.md`, `docs/fix-history.md`, and the Unreleased changelog.

## Evidence and provenance

`Network.wrap` owns the connection-open recorder boundary and must close the transport when recording fails. The previous cleanup discarded the close result, hiding a lower-layer failure. The change preserves both errors without changing the cleanup order or wire behavior. No external source code was copied.

## Automated validation

- Focused network tests, including the combined open-capture and close-error regression, are required before commit.
- Full `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit.

## Live validation

No live server, public service, or external account was contacted. Live transport-failure behavior remains pending.

## Known limitations and follow-up

The change does not add raw UDP or RakNet datagram observation. It only preserves error provenance during observation setup.
