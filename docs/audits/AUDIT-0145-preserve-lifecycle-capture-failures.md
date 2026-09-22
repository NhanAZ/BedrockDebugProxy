# Audit AUDIT-0145

## Intent

Preserve capture failures from lifecycle error events instead of allowing incomplete sessions to continue.

## Scope

- Updated `internal/proxy/proxy.go` so spawn, upstream-dial, and transfer route-probe diagnostics join recorder failures and report them through the run failure sink.
- Added `TestTransferRouteProbePropagatesCaptureFailure` in `internal/proxy/proxy_test.go`.
- Updated `docs/research/capture-observability.md`, `docs/fix-history.md`, and the Unreleased changelog.

## Evidence and provenance

The affected paths already have explicit lifecycle events and a shared `FailureSink`, but previously discarded recorder errors after the underlying operation failed or a route probe completed. The patch preserves both causes and keeps route-probe network failures non-fatal to the normal dial decision. No packet values, ordering, or routing policy changed. No external source code was copied.

## Automated validation

- Focused proxy tests, including route-probe capture failure propagation, are required before commit.
- Full `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit.

## Live validation

No live server, public service, or external account was contacted. Live recorder-failure behavior remains pending.

## Known limitations and follow-up

The change covers the named lifecycle diagnostics. It does not make operator output or best-effort transport shutdown writes part of the canonical capture path.
