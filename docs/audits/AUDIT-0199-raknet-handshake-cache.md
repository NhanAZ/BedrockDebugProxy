# AUDIT-0199 - Synchronize the RakNet handshake cache

- Commit subject: `fix: synchronize the RakNet handshake cache`
- Change type: `runtime | networking`
- Runtime impact: `yes`

## Intent and root cause

Concurrent dials share the cached first connection-request payloads. The unguarded map caused a data race and could terminate the process with a concurrent map read or write. The new synthetic regression reproduced both problems on the previous source.

## Scope and provenance

Backport the mutex from MIT upstream revision `0667bace52851f53ca3f1c916689620db017e7a3` in `internal/message/open_connection_request_1.go`. Add a concurrent marshal regression and update the patch description, notices, fix history, and changelog. This is copied upstream source with its existing license retained. The [maintenance review](../research/protocol-maintenance-2026-10-05.md) records the immutable evidence. Wire bytes and the local GUID, transfer, retransmission, and cancellation behavior are unchanged.

## Validation

- Automated checks: the previous source failed the new test with `WARNING: DATA RACE` and `fatal error: concurrent map read and map write`. After the mutex backport, `go test -race -count=1 ./...` passed from `third_party/go-raknet` with Go 1.26.6 and CGO enabled. `tools/format.ps1`, `tools/quality.ps1 -LowMemory`, and `git diff --check` passed on the reviewed maintenance source. Static analysis reported zero issues and no project-reachable vulnerability was found.
- Live validation: `not_observed`.
- Capture or report: none. The concurrent test performs no network operation.

## Human validation and limitations

Use an owner-controlled RakNet server or authorized Featured Experience. Exercise overlapping connections and a normal disconnect, then verify the closed captures. Handshakes should complete without a process crash and preserve the configured GUID and packet bytes. Return the first handshake failure and closed capture if the session fails. This cache test does not establish retail-client or six-server release compatibility.
