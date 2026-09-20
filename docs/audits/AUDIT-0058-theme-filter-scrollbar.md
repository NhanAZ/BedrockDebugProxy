# AUDIT-0058 - Theme filter menu scrollbar

- Commit subject: `style: theme filter menu scrollbar`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Make the Packet and Kind filter menu scrollbar match the viewer's compact dark controls.

## Scope

Add cross-browser scrollbar styling to `.select-menu` with a thin track, transparent background, and the existing `--line-bright` and `--muted` colors. Menu sizing, option text, filtering, and capture data are unchanged.

## Evidence and provenance

The reported screenshots showed a thick native scrollbar inside the expanded Packet and Kind menus. The custom dropdown already owns the menu surface, so styling its scrollbar is a local presentation change.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `git diff --check`, and `./tools/quality.ps1` passed before commit.
- Live validation: passed on the clean build at `[local viewer URL]`. Browser smoke confirmed the Packet dropdown uses a thin dark thumb with a transparent track while retaining the full option labels. No browser errors or warnings were visible during the smoke test.
- Capture or report: `none`

## Limitations and follow-up

Scrollbar pseudo-elements are honored by Chromium-based browsers, while Firefox uses its standard `scrollbar-width` and `scrollbar-color` properties.
