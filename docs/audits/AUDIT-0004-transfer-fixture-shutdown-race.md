# AUDIT-0004 - Keep transfer integration fixture alive until proxy shutdown

- Commit subject: `Stabilize transfer integration fixture shutdown`
- Change type: `test | documentation`
- Runtime impact: `no`

## Intent

Make the transfer-following regression test deterministic on slower runners. The fixture must keep its upstream
connection open after sending `Transfer` so the proxy can flush and forward the rewritten packet before the fixture
closes the transport.

## Evidence

GitHub Actions run 27 for commit `91c2dba` passed `Quality gate` but failed `Windows compatibility` in
`TestRunnerFollowsTransferAndAcceptsNextHop`. The Windows log showed `Transfer` observed on the upstream login path,
no downstream transfer before the test context expired, and a temporary-directory cleanup error because the capture
file was still open. The local test passed repeatedly, which is consistent with a transport-close race in the fixture.

## Scope

Change only `internal/proxy/transfer_test.go`. After `Flush`, the first-hop fixture now reads until the proxy closes
the hop, and reports the server result then. No production forwarding, protocol, capture, or CLI behavior changes.

## Validation

- Focused transfer regression test: `go test -count=5 -run '^TestRunnerFollowsTransferAndAcceptsNextHop$' -v ./internal/proxy` passed five consecutive runs.
- Project quality gate: `tools/format.ps1` and `tools/quality.ps1` passed, including formatting, all tests, build,
  static analysis, vulnerability scan, license inventory, and `git diff --check`.
- Post-push GitHub Actions validation: pending until the pushed commit completes.
- Live server validation: not applicable to a test-fixture-only change.

## Follow-up

If the focused test still fails on Windows, inspect the new capture and transfer rewrite event before changing runtime
code. Do not hide a forwarding defect by extending the test timeout alone.
