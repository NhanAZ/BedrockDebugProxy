# Audit 0151 - Account accepted events lost after writer failure

## Intent

Make asynchronous recorder failures explicit in the canonical capture so accepted events are never silently discarded.

## Scope

- Updated `internal/capture/recorder.go` to drain queued jobs after a writer failure, count the failed and drained jobs as dropped, record one completeness limitation, and keep pending queue accounting consistent.
- Prevented enqueue rollback from rewinding a sequence that has already been committed by the writer.
- Added deterministic coverage in `internal/capture/recorder_test.go`.
- Updated the capture fidelity decision, research note, fix history, and changelog.
- No packet bytes, forwarding policy, capture schema, or protocol definitions changed.

## Evidence and provenance

- The recorder uses a bounded ordered queue and accepts a job before the asynchronous writer persists it.
- Before this change, a writer error returned without accounting for jobs already accepted into the queue. The manifest could therefore show only `write_errors` while omitting the loss of those accepted jobs.
- The regression test uses a synthetic local capture and closes its event stream before queuing three jobs. No credentials, captures, or external endpoints are involved.

## Automated validation

- `go test ./internal/capture` passes.
- `go test -race ./internal/capture` passes.
- `git diff --check` passes before commit.
- The full project quality gate with `-LowMemory` passes.

## Live validation status

No live or public-server validation was performed. Disk-failure accounting is covered only by deterministic local tests.

## Known limitations and follow-up

The writer must be closed through `Recorder.Close` to publish the final manifest after a failure. The manifest explicitly marks incomplete captures and reports dropped accepted jobs. The existing exact-revision live validation matrix remains pending independently of this recorder fix.
