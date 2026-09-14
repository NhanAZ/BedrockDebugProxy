# AUDIT-0002 - Add fix history and login-order regression coverage

- Commit subject: `Add fix history and login-order regression coverage`
- Change type: `documentation | tests | process`
- Runtime impact: `no`

## Intent

Give agents and maintainers one place to find earlier fixes before changing the proxy, and reduce the risk that a
featured-server ordering fix is treated as a universal Bedrock sequence.

## Scope

Add [`docs/fix-history.md`](../fix-history.md) as a concise index of the major compatibility fixes, add ADR 0006
for the bounded deferred-login policy, and require contributors to consult the index before a new fix. Extend the
in-tree gophertunnel regression tests to prove that a later-phase packet stays queued while currently expected
resource-pack packets drain.

No production packet handling, capture schema, CLI behavior, authentication behavior, or forwarding bytes change
in this commit.

## Evidence and provenance

The history index links existing commits, architecture decisions, research notes, local captures, and validation
reports. The ordering contract is based on the failed Enchanted hop 2 capture and the documented upstream
gophertunnel login-order evidence. No source code was copied from another project.

## Validation

- `gofmt -w C:\Users\NhanAZ\Documents\GitHub\BedrockDebugProxy\third_party\gophertunnel\minecraft\conn_featured_experience_test.go`
- `go test ./minecraft` from `third_party/gophertunnel`
- `git diff --check`
- Live validation: `not_observed`; this commit has no runtime impact. The latest queue implementation still needs a stamped The Hive run and the affected Enchanted flow before release.
- Capture or report: existing paths are linked in the fix history and ADR; no capture is added to Git.

## Limitations and follow-up

The index summarizes the major fixes rather than creating synthetic audit files for historical commits. Historical
commits remain unchanged because shared Git history must not be rewritten. Build the exact post-commit binary and
run The Hive first, then Enchanted transfer/resource-pack validation and the remaining six-server matrix. If a
server fails, classify the capture as a regression or a new root cause before changing the generic login state.

## Compatibility and security impact

The tests enforce the narrow expected-ID boundary and preserve deferred packet evidence. They do not bypass
authentication, anti-cheat checks, encryption, resource-pack policy, or server rules.
