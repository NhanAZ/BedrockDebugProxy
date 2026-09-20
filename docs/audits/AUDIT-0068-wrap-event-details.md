# AUDIT-0068 - Wrap long event details

- Commit subject: `fix: wrap long event details`
- Change type: `viewer`
- Runtime impact: `no`

## Intent

Expose complete long error summaries in the event stream instead of hiding their tail behind an ellipsis.

## Scope

Give Details cells a dedicated class and allow normal and unbroken text wrapping. Other table columns keep their compact single-line behavior. No capture, protocol, transport, or server behavior changes.

## Evidence and provenance

The reported Hive capture row `#93459` displayed a long `bridge.read_error` summary with the server address and ban reason truncated in the Details column. The existing shared cell rule applied `white-space: nowrap`, hidden overflow, and ellipsis to every table cell.

## Validation

- Automated checks: `git diff --check`, the project quality gate, viewer HTTP tests, JavaScript syntax validation, and a rebuilt binary.
- Live validation: browser smoke on a clean committed viewer using the local Hive capture `session-20260920T135412Z`; no new Hive connection is required or authorized.
- Capture or report: the local Hive capture remains unchanged and sensitive.

## Limitations and follow-up

Long Details messages increase the height of their event row. The other columns remain single-line to preserve the compact event-stream layout.
