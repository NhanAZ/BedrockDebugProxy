# AUDIT-0181 - Synchronize resource-pack queue state

- Commit subject: `Synchronize resource-pack queue state`
- Change type: `bug fix`
- Runtime impact: `yes`

## Intent

Prevent the proxy from fatally crashing when an upstream resource pack finishes downloading at the same time that another pack begins.

## Scope

Guard the shared gophertunnel resource-pack map lookup and transition with the existing connection mutex. Add a concurrent pack setup regression test and document the incident. Packet contents, ordering, download window, and downstream pacing are unchanged.

## Evidence and provenance

The maintainer-provided Hive console trace on 2026-09-23 records `fatal error: concurrent map writes` in `Conn.handleResourcePackDataInfo`. Source inspection confirmed that the download goroutine already removed completed entries from `awaitingPacks` under `Conn.packMu`, while the next pack-info handler inserted into the same map without that lock. The relevant bounded pipeline change is local commit `3a77b024ef44f37479f925ed1cba0aacddffc448`. No third-party source was copied and no capture, credential, or account data is included.

## Validation

- Focused automated checks: `go test ./minecraft -run 'TestResourcePack(DataInfoTransitionsConcurrentPacksSafely|DownloadWindowReplenishesAfterResponse)$' -count=10` and `go test -race ./minecraft -run TestResourcePackDataInfoTransitionsConcurrentPacksSafely -count=5` passed in the nested gophertunnel module.
- Project quality gate: `tools/quality.ps1 -LowMemory` passed. This included formatting, PowerShell and viewer JavaScript syntax, module tidiness and integrity, third-party license inventory, all Go tests, build, static analysis with 0 issues, vulnerability scan with 0 reachable vulnerabilities, tracked-artifact checks, and whitespace validation. The scan noted 4 imported-package and 2 required-module findings that the code does not appear to call.
- Live validation: pending a new build and a fresh The Hive session. The previous candidate crashed and is not valid evidence.
- Capture or report: the user's console output was inspected in chat and was not added to Git.

## Limitations and follow-up

The focused test stresses concurrent pack state setup but does not establish compatibility with a live server. Build from the fixed revision and complete a fresh The Hive flow before considering runtime validation complete or resuming release approval.
