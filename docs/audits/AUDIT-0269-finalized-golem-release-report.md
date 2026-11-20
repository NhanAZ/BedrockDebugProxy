# AUDIT-0269 - Use finalized capture state for Golem release reports

- Commit subject: `Use finalized capture state for Golem release reports`
- Change type: `tooling`
- Runtime impact: `yes`

## Intent

Prevent an immediate-spawn release report from being marked incomplete by capture-observer state that changes after the canonical capture has closed.

## Scope

The Go Golem session check uses the completed client snapshot, scenario result, and recognized intentional-close error. The validation report generator remains responsible for checking the finalized capture, including integrity, packet decoding, pack decryption, and artifacts. No packet flow, capture event, capture schema, resource-pack behavior, or BedrockPackArchive contract changes.

## Evidence and provenance

An authenticated Galaxite Golem session spawned, closed, and decrypted all 20 encrypted packs. Its generated report marked `cli_session=not_observed` while every automatic capture requirement passed. Source inspection showed that the CLI re-read mutable capture-observer state after `finishCapture` had finalized the capture. The previous Hive diagnostic showed the same report pattern. These observations classify a local report gate issue, not a server-side failure.

## Validation

- Automated checks: `tools/quality.ps1 -Race` passed locally. Hosted run [44799947359](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/44799947359) passed both the race-enabled quality gate and Windows compatibility job. The preceding hosted failure [44798722087](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/44798722087) exposed a test-only snapshot timing race, corrected in commit `84ab313`.
- Live validation: the stamped `0.6.0` binary from revision `84ab31396cc2b80aa094c47d924489a69581ed43` passed The Hive, Galaxite, Lifeboat, Mineville Zeqa, Enchanted, and CubeCraft. Each report records authenticated upstream, spawn, immediate close, bidirectional decoded traffic, and passing capture evidence. The 61 encrypted packs offered by The Hive, Galaxite, Lifeboat, and CubeCraft were all decrypted. Enchanted and Mineville offered no encrypted packs during these sessions.
- Capture or report: each of the six closed captures passed the CLI verifier, and `tools/check-release-readiness.ps1 -Revision 84ab31396cc2b80aa094c47d924489a69581ed43` passed all six sanitized reports. The reports remain in ignored local validation storage.

## Limitations and follow-up

Only the expected `net.ErrClosed` teleport acknowledgement error after immediate close is allowed by the session check. The finalized capture must still pass every automatic requirement. The reports prove the observed Go Golem protocol sessions only. They do not claim retail rendering or gameplay. The official tag and release publication remain to be completed.
