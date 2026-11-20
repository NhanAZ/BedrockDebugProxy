# AUDIT-0264 - Make the item input fixture wait for observation

- Commit subject: `Make the item input fixture wait for observation`
- Change type: `test reliability`
- Runtime impact: `no`

## Intent

Fix the Windows compatibility job failure in `TestClientUsesObservedHotbarItemInPlayerInput` without changing the Go network client's item-use behavior.

## Root cause and scope

The test closed the client immediately after `UseItem` returned, before the local server had confirmed receipt of the flushed packet. The Windows runner reproduced the race as a canceled packet read. The test now waits for the server's existing assertion result before closing the connection, so the test checks the packet rather than scheduler timing.

## Validation

- `go test -count=50 -run TestClientUsesObservedHotbarItemInPlayerInput ./internal/golemnet` passed on the local Windows host.
- `tools/format.ps1` and `tools/quality.ps1` passed. The full quality gate includes the complete test suite, build, static analysis, vulnerability scan, tracked-capture checks, and whitespace check.
- No runtime code or protocol behavior changed.

## Limitations

The corrected commit still requires its hosted Windows compatibility job to pass. The prior hosted failure was run `44739914949`, job `136300673429`.
