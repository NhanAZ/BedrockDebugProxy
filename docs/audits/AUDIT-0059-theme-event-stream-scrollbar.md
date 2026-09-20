# AUDIT-0059 - Theme event stream scrollbar

- Commit subject: `style: theme event stream scrollbar`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Give the bounded event stream a visible scrollbar that matches the Packet and Kind filter menus.

## Scope

Restore the `.table-wrap` scrollbar and apply the same thin six-pixel Chromium styling plus Firefox `scrollbar-width` and `scrollbar-color` properties used by the custom filter menus. Table paging, row ordering, wheel scrolling, keyboard scrolling, and capture data are unchanged.

## Evidence and provenance

The reported screenshot highlighted the empty right edge of the event stream where the internal scrollbar had previously been hidden. The table remains a bounded scrollable viewport, so a themed scrollbar restores a direct drag affordance without changing layout or evidence handling.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `git diff --check`, and `./tools/quality.ps1` passed before commit.
- Live validation: passed on the clean build at `http://127.0.0.1:53612/e98980b4c3b322c785a63fe850e10b34/`. Browser smoke confirmed the event stream scrollbar is thin, dark, and consistent with the Packet and Kind menus. No browser errors or warnings were visible during the smoke test.
- Capture or report: `none`

## Limitations and follow-up

Scrollbar pseudo-elements are honored by Chromium-based browsers, while Firefox uses its standard `scrollbar-width` and `scrollbar-color` properties.
