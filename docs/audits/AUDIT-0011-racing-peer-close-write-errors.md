# AUDIT-0011 - Suppress racing peer-close write errors

## Intent

Prevent a normal connection close from producing a blocking `bridge.write_error` when the two forwarding loops observe shutdown concurrently.

## Scope

The change is limited to terminal forwarding error classification in `internal/proxy/proxy.go`, regression coverage in `internal/proxy/shutdown_test.go`, and the validation and fix-history documentation for this behavior.

## Affected interfaces

- `Runner.finishForwardAtHop` now atomically claims the first terminal shutdown signal.
- `bridge.write_error` and `bridge.flush_error` caused by `context.Canceled` or `net.ErrClosed` are treated as expected peer-close termination, including when they win the race.
- Read-side termination evidence and `session.close` diagnostics remain available.

## Evidence and provenance

The Mineville Zeqa capture `captures/session-20260916T025636Z` at revision `9a7859c351b55b66a89cfa7e3448b809f5a48584` recorded a downstream context-cancelled read followed by a concurrent upstream `bridge.write_error`. This is a race variant of FIX-0004 in [`docs/fix-history.md`](../fix-history.md). The transport error classification uses Go's `context.Canceled` and `net.ErrClosed` values exposed by gophertunnel connection shutdown.

## Automated validation

- `internal/proxy` shutdown tests cover coordinated read-first ordering, first write-side close ordering, and preservation of the existing coordinated-shutdown helper contract.
- `tools/format.ps1`, `tools/quality.ps1`, and `git diff --check` must pass before commit.

## Live validation status

Pending. Re-run Mineville Zeqa through the stamped binary and inspect the closed capture and sanitized report. The expected result is no blocking `bridge.write_error` or `bridge.flush_error` for normal peer close, while `session.close` still records the loop errors.

## Known limitations

This change does not alter packet order, transport behavior, or resource-pack handling. It does not suppress non-close write failures or protocol/decode errors. The existing capture scope boundary and live five-server release gate remain unchanged.

## Follow-up work

If a fresh capture still reports a blocking error, classify its exact error chain and packet phase before changing the shutdown policy again.
