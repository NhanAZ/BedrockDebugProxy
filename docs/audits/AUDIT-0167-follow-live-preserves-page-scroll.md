# Audit AUDIT-0167 - Keep page scroll stable during live follow

## Intent

Keep the outer viewer page and Decoded view - Event details at the operator's current scroll position while Follow live selects newer packets.

## Scope

- Replace page-level row scrolling with table-wrapper-only scrolling for Follow live.
- Use the same bounded scroll behavior for keyboard event navigation.
- Update viewer regression coverage, fix-history, and changelog text.

No capture format, packet forwarding, filtering semantics, or HTTP API behavior changed.

## Evidence and provenance

The live refresh already selects the final row returned by the `tail=true` event query. The page jump came from calling `scrollIntoView` on that row after the table was rendered. The replacement compares the row and table-wrapper rectangles and adjusts only the wrapper's `scrollTop`.

## Automated validation

- Viewer asset coverage requires the `scrollRowWithinTable` hook.
- `node --check internal/viewer/assets/app.js` passed.
- `go test ./internal/viewer` passed.
- `git diff --check` passed.
- `powershell -NoProfile -ExecutionPolicy Bypass -File tools/quality.ps1 -LowMemory` passed, including viewer JavaScript syntax validation, all tests, build, static analysis, and the vulnerability scan.

## Live validation

Pass via a local live capture viewer smoke test using the source committed in this change. With Decoded view - Event details in view, the page remained at `scrollY=80.8` while live events advanced from sequence 14 to 19. The newest row was selected and the details panel remained visible. The browser console had no errors or warnings.

No public-server or real-client validation is required for this viewer-only change.

## Known limitations and follow-up

The table wrapper may scroll internally to keep the selected row visible. The page-level scroll position is intentionally not changed by live-follow selection.
