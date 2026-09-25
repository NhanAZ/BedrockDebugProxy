# Audit AUDIT-0191 - Report RakNet acceptance wait on cancellation

- Commit subject: `Report RakNet acceptance wait on cancellation`
- Change type: `runtime diagnostics`
- Runtime impact: `yes`

## Intent

Preserve the RakNet handshake stage when a dial context is canceled while the dialer waits for `CONNECTION_REQUEST_ACCEPTED`, so the existing capture error can identify that wait.

## Scope

- Wrap the acceptance-wait cancellation in `third_party/go-raknet/dial.go` with the stage description.
- Update the changelog, fix history, Enchanted research note, and third-party patch provenance.
- Do not change RakNet packets, retries, timeout values, connection behavior, or capture schema.

## Evidence and provenance

Source review confirms the affected branch runs after `CONNECTION_REQUEST` is sent and waits for the acceptance signal. The failed Enchanted capture [captures/session-20260925T013400Z](https://github.com/BedrockDebugProxy-Captures/session-20260925T013400Z/releases/tag/capture-archive) recorded a canceled hop 3 dial without transport-open evidence, but does not prove that this branch caused that failure. The change is diagnostic and does not claim a root-cause fix.

## Validation

- Automated checks: `tools/format.ps1` and `tools/quality.ps1 -LowMemory` passed, including tests, build, static analysis, vulnerability scan, and whitespace checks. The vulnerability scan found no issues reachable from this code and noted additional imported-package and required-module advisories.
- Live validation: `pass` for the successful Enchanted transfer sessions below; the cancellation diagnostic itself was not exercised.
- Capture or report: [captures/session-20260925T022556Z](https://github.com/BedrockDebugProxy-Captures/session-20260925T022556Z/releases/tag/capture-archive), [captures/session-20260925T023226Z](https://github.com/BedrockDebugProxy-Captures/session-20260925T023226Z/releases/tag/capture-archive)

## Live validation status

Both captures closed completely. The first completed six hops and the second four hops. Every recorded upstream hop connected and spawned, transfer route probes completed, pack artifacts completed, and no upstream dial or decode errors were present. One final-hop `context canceled` read event in each capture coincided with session shutdown.

## Limitations and follow-up

The earlier transfer stall did not reproduce, so its cause remains unknown. If it recurs, inspect whether the new stage appears before investigating GUID, source-port, MTU, or edge behavior.
