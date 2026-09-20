# AUDIT-0048 - Keep viewer footer below the inspector

- Commit subject: `fix: keep viewer footer below inspector`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Keep the viewer footer below the complete workspace instead of allowing it to overlap decoded details or Blob bytes.

## Scope

Make the page and workspace grid tracks content-sized. The change affects only viewer layout. Capture data, event ordering, protocol behavior, and the existing page-level scrollbar remain unchanged.

## Evidence and provenance

The reported screenshot showed the footer between the decoded details panel and the Blob bytes panel. The inspector is a content-sized grid with two panels, so its full height must contribute to the parent workspace track before the footer is placed.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: pass via a clean local `view --no-open` build at `[local viewer URL]`. The browser smoke test shows the compact header and, after selecting an event and scrolling through the expanded details, keeps the footer below Blob bytes rather than overlaying the inspector. No browser errors or warnings were visible during the smoke test.
- Capture or report: `none`

## Limitations and follow-up

Long decoded events intentionally increase page scroll height so the viewer does not hide evidence behind an internal details scrollbar.
