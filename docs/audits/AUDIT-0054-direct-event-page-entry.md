# AUDIT-0054 - Direct event page entry

- Commit subject: `feat: add direct event page entry`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Make distant pages in large capture viewers reachable without repeatedly clicking the next-page control.

## Scope

Replace the current-page text in the event pager with a number input. Enter or a committed value loads the corresponding event offset, while empty, invalid, and out-of-range values restore or clamp to the available page range. Capture data, API responses, event ordering, and page size remain unchanged.

## Evidence and provenance

The reported viewer contained 369 pages and exposed only previous and next controls. The existing API already accepts an offset and limit, so direct page entry can reuse the established pagination path without changing capture storage or server behavior.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `git diff --check`, and `./tools/quality.ps1` passed before commit.
- Live validation: passed on the clean build at `http://127.0.0.1:54887/fe0964c553f9b374abd32a058e6241da/`. Entering `369` updated the pager to `369 / 369` and the event range to `36,801-36,867`. No browser errors or warnings were visible during smoke test.
- Capture or report: `none`

## Limitations and follow-up

The page input uses the current fixed page size of 100 events. It changes the viewer offset only and does not add a new capture or API pagination mode.
