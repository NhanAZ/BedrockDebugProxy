# Audit AUDIT-0144

## Intent

Propagate capture-writer failures from the listener's `RequestChunkRadius` callback boundary.

## Scope

- Updated `internal/proxy/proxy.go` so decode-error and upstream-forwarding error events from the listener callback reach the run's `FailureSink`.
- Added `TestSetCaptureFailurePropagatesCallbackRecorderError` in `internal/proxy/proxy_test.go`.
- Updated `docs/research/chunk-radius-parity.md`, `docs/research/capture-observability.md`, `docs/fix-history.md`, and the Unreleased changelog.

## Evidence and provenance

The gophertunnel packet callback cannot return an error. The previous callback discarded errors returned by `recordPacketDecodeErrorAtHop` and `recordNetworkErrorAtHop`, so a closed or failed recorder could leave the run active after losing canonical evidence. `setCaptureFailure` reuses the existing first-error cancellation path and does not alter packet values or forwarding decisions. No external source code was copied.

## Automated validation

- Focused proxy tests, including callback failure propagation, are required before commit.
- Full `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit.

## Live validation

No live server, public service, or external account was contacted. Live capture-writer failure behavior remains pending.

## Known limitations and follow-up

The change covers the two `RequestChunkRadius` listener callback recorder boundaries. It does not change the packet decode policy or the typed forwarding policy.
