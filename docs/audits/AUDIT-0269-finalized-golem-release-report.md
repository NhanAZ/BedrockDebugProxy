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

- Automated checks: focused `golemValidationChecks` regression and `tools/quality.ps1` pass in the working tree. The quality gate includes tests, build, static analysis with 0 issues, vulnerability scan, tracked-capture checks, and whitespace checks.
- Live validation: earlier Hive and Galaxite attempts reached spawn and completed pack decryption. The complete six-server matrix must be rerun against the final stamped revision.
- Capture or report: local diagnostic reports remain in ignored validation storage. Final sanitized release reports will be generated from the final stamped revision.

## Limitations and follow-up

Only the expected `net.ErrClosed` teleport acknowledgement error after immediate close is allowed by the session check. The finalized capture must still pass every automatic requirement. Release remains pending until CI and all six exact-revision reports pass.
