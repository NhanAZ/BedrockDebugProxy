# AUDIT-0041 - Expand event details without an internal scrollbar

- Commit subject: `style: expand viewer event details`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Let the selected event details panel grow with its decoded metadata and JSON tree instead of placing a vertical scrollbar inside the panel.

## Scope

Updated the embedded viewer layout and its asset regression checks, plus the changelog and fix-history record. The event table, raw evidence preview, capture API, and canonical capture data remain unchanged.

## Evidence and provenance

The prior layout used fractional inspector rows and `.detail-content { overflow: auto; }`, which matched the scrollbar visible in the reported screenshot. The panel now uses content-sized rows and visible overflow while the workspace remains the outer scroll boundary.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, and `./tools/quality.ps1` passed. `git diff --check` passed.
- Live validation: `pass` via a clean local `view --no-open` build, browser inspection of a selected event with a full decoded tree, and a console check with no errors or warnings.
- Capture or report: `none`

## Limitations and follow-up

Very large decoded trees can make the workspace long. This is intentional so details remain reachable without a nested panel scrollbar. The bounded raw hex preview remains independently scrollable when its content exceeds the preview area.
