# Audit 0186 - Synchronize network settings metadata

- Commit subject: `Synchronize negotiated network settings metadata`
- Change type: `race safety fix`
- Runtime impact: `yes`

## Intent

Remove the login-time data race between processing `NetworkSettings` and reading negotiated connection metadata for capture events.

## Scope

- Protect the stored `NetworkSettings` value in the in-tree gophertunnel fork with a read-write mutex.
- Return a synchronized value copy and lock both listener-side and dial-side publication.
- Add a concurrent reader and negotiation regression test.
- Update the fork provenance note, fix history, changelog, and prior transfer-fix audit with the exact CI result.

## Evidence and provenance

GitHub Actions run `35947403349` for commit `d8724a45feed4889e92616fdb845ec9f2d2f50fa` passed Windows compatibility but the Linux race-enabled job detected a race between `Conn.NetworkSettings()` and `Conn.handleNetworkSettings()` during `TestRunnerForwardsBidirectionalPacketsAndCapturesSession` and `TestRunPacksOnlyCapturesUpstreamWithoutListener`. The same getter is used by the proxy's `transportMetadata` callback. The relevant methods are project-maintained changes in the in-tree MIT-licensed gophertunnel fork.

## Validation

- Automated checks: `tools/quality.ps1 -Race` passed the complete race-detected Go suite, build, viewer checks, module and license checks, static analysis (0 issues), and whitespace validation. The vulnerability scan found 0 reachable vulnerabilities, plus 4 in imported packages and 2 in required modules that the code does not appear to call.
- Focused race test: `go test -race ./minecraft -run '^TestNetworkSettingsCanBeReadDuringNegotiation$' -count=10` passed in the vendored gophertunnel module. The full vendored gophertunnel race suite passed, and the two CI-affected proxy tests passed three race-enabled repetitions locally.
- CI: pending for the commit containing this record.
- Live validation: not run. A stamped binary must complete the required The Hive baseline and release server matrix.

## Limitations and follow-up

The change synchronizes the retained metadata only. It does not change compression negotiation, packet order, or wire contents. Exact-revision CI and live validation remain required before release approval.
