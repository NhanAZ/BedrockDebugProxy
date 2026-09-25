# AUDIT-0182 - Share resource-pack request window

- Commit subject: `Share resource-pack request window`
- Change type: `bug fix`
- Runtime impact: `yes`

## Intent

Prevent concurrent resource-pack download workers from multiplying the intended connection-wide bound for in-flight chunk requests.

## Scope

Add a connection-wide 100-request semaphore to the vendored gophertunnel dial path. Keep the per-pack reorder buffers and validation, original pack ordering, downstream sequential delivery, and downstream pacing unchanged. Update the resource-pack guide, research note, fix history, changelog, and vendored patch provenance.

## Evidence and provenance

The `0.3.0-rc.3` Galaxite capture recorded 460 chunk requests in about 10.5 seconds but only 97 chunk responses during its five-minute lifetime. A prior Galaxite capture on revision `29fbfd8e0664be20cb3929f5916f961e4e07b66c` recorded 460 requests and 460 responses before reaching `session.spawned` about 2 minutes 45 seconds after dial start. Source inspection confirmed one independent 100-slot window per `ResourcePackDataInfo` downloader. The request burst is confirmed; the server-side cause of the low response count remains inferred. The bounded client request window is also described in [gophertunnel PR #509](https://github.com/Sandertv/gophertunnel/pull/509), reviewed as an implementation reference only. The referenced raw captures are stored in private capture repositories and remain outside this project's source Git history.

## Validation

- Automated checks: focused gophertunnel resource-pack tests with the race detector passed five runs. `tools/quality.ps1 -LowMemory` passed, including all root-module tests, build, static analysis with 0 issues, vulnerability scan with 0 reachable vulnerabilities, tracked-artifact checks, and whitespace validation. The scan reported 4 imported-package and 2 required-module findings that the code does not appear to call.
- Live validation: pending fresh exact-revision The Hive and Galaxite sessions.
- Capture or report: [captures/session-20260923T044137Z](https://github.com/BedrockDebugProxy-Captures/session-20260923T044137Z/releases/tag/capture-archive) and [captures/session-20260922T153826Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T153826Z/releases/tag/capture-archive); raw archives are attached to private GitHub Releases, not committed to this project's source history.

## Limitations and follow-up

The failed Galaxite capture ended with `context canceled`, not the configured dial timeout. It does not establish whether the operator or the client canceled the session. Rebuild from the final commit, complete the required The Hive baseline, then repeat Galaxite and inspect the closed capture for all resource-pack responses, spawn, and clean shutdown.
