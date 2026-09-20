# AUDIT-0044 - Collapse Blob bytes by default

- Commit subject: `style: collapse viewer blob bytes`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Keep the raw Blob bytes preview compact by default so decoded event details remain the primary view.

## Scope

Updated the embedded viewer markup, styling, and toggle behavior, plus the viewer asset tests, changelog, and fix-history record. Capture storage, blob bytes, event selection, and forwarding behavior are unchanged.

## Evidence and provenance

The reported viewer showed a large hexdump consuming the page after selecting a packet. The hexdump is still generated from the same bounded blob preview, but the panel now starts collapsed. The header retains the byte count and representation, and an accessible button controls the preview with `aria-expanded` state.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: pass via a clean local `view --no-open` build at `[local viewer URL]`. The initial page showed Blob bytes collapsed, selecting an event preserved the collapsed state, and activating the toggle revealed the hexdump. No browser errors or warnings were visible during the smoke test.
- Capture or report: `none`

## Limitations and follow-up

The expanded state is intentionally page-local and resets when the viewer reloads. No capture data or raw bytes are changed.
