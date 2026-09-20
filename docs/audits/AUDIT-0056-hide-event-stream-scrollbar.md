# AUDIT-0056 - Hide event stream scrollbar chrome

- Commit subject: `style: hide event stream scrollbar chrome`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Remove the nested scrollbar chrome highlighted in the event stream while keeping the rows scrollable.

## Scope

Hide the `.table-wrap` scrollbar track and thumb with the standard Firefox property, the legacy Windows property, and the Chromium pseudo-element. The wrapper remains `overflow: auto`, so wheel and keyboard scrolling continue to work and capture data is unchanged.

## Evidence and provenance

The reported screenshot highlighted the event stream's inner right-edge scrollbar. The viewer still needs a bounded table viewport for its 100-row page, so removing overflow would clip rows. Hiding only the scrollbar chrome preserves the existing interaction and page-level layout.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `git diff --check`, and `./tools/quality.ps1` passed before commit.
- Live validation: passed on the clean build at `[local viewer URL]`. Browser smoke confirmed the nested scrollbar is hidden while the event stream remains present and the page-level scrollbar remains available. No browser errors or warnings were visible during smoke test.
- Capture or report: `none`

## Limitations and follow-up

The scroll position remains available through wheel, keyboard, touchpad, and assistive scrolling. No visible drag thumb is provided inside the event stream.
