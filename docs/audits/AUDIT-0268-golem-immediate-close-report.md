# AUDIT-0268 - Accept expected Golem close race after spawn

- Commit subject: `Accept expected Golem close race after spawn`
- Change type: `tooling`
- Runtime impact: `yes`

## Intent

Let the authenticated Go Golem produce a passing release report after it completes packs, reaches spawn, and closes immediately, while keeping real session errors blocking.

## Scope

The release check now accepts the observed `net.ErrClosed` from a server teleport acknowledgement that races with the intentional socket close after `quit`. It still rejects any other Golem client error. Packet forwarding, capture events and fields, capture schema, resource-pack handling, and BedrockPackArchive's importer contract are unchanged.

## Evidence and provenance

An owner-authorized authenticated Hive session completed its offered encrypted packs, reached spawn, and closed immediately. The capture-derived report checks passed. The remaining Golem client error was a teleport acknowledgement write to the intentionally closed socket. No raw capture or server payload is included in this audit.

## Validation

- Automated checks: focused `golemValidationChecks` regression and `tools/quality.ps1` pass in the working tree. The quality gate includes tests, build, static analysis with 0 issues, vulnerability scan, tracked-capture checks, and whitespace checks.
- Live validation: the first Hive session observed the close race. The complete six-server matrix must be rerun against the final stamped revision.
- Capture or report: sanitized Hive diagnostic report retained locally. The final release reports will be generated from the final stamped revision.

## Limitations and follow-up

This narrowly recognizes one expected shutdown error. Other Golem errors continue to block the CLI session check. No BedrockPackArchive code or fixture update is needed because producer capture data and importer behavior do not change. Release remains pending until the final exact-revision six-server matrix passes.
