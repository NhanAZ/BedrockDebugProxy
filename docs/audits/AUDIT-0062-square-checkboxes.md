# AUDIT-0062 - Square filter checkboxes

- Commit subject: `style: square viewer checkboxes`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Make the Problems only checkbox match the viewer's square pixel controls instead of the rounded native browser checkbox.

## Scope

Replace native checkbox rendering for the Problems only and live-follow controls with twelve-pixel square inputs, preserve the existing labels and checked behavior, and use the red or cyan palette for the checked fill.

## Evidence and provenance

The supplied screenshot showed a rounded native checkbox inside the Problems only control while the surrounding viewer controls use square corners. The change is CSS-only and does not alter filter query handling or live-follow state.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `git diff --check`, and `./tools/quality.ps1` passed before commit. The viewer asset test checks the custom checkbox geometry and checked-state declarations.
- Live validation: passed on the clean build at `[local viewer URL]`. Browser smoke inspected the Problems only control, confirmed square corners and a visible red checked fill, and confirmed the filter query updates from 36,867 to 20,238 matching events.
- Capture or report: `none`

## Limitations and follow-up

The exact pixel appearance depends on the browser's support for `appearance: none`, which is covered by the Chromium viewer smoke path.
