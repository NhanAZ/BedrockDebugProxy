# AUDIT-0043 - Use one page-level viewer scrollbar

- Commit subject: `style: move viewer scrolling to page level`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Make the viewer's vertical scrollbar span the complete page instead of starting below the header and filter controls.

## Scope

Changed the embedded viewer page and workspace sizing rules, the viewer CSS regression checks, the changelog, and the fix-history index. Capture loading, event filtering, decoded details, raw evidence, and forwarding behavior are unchanged.

## Evidence and provenance

The reported screenshot showed the scrollbar beginning at the workspace boundary. The body was fixed to the viewport with `overflow: hidden`, and `.workspace` owned the vertical scroll context. The body now grows with the page, the workspace leaves overflow visible, and the event stream uses a bounded viewport-height row so the page-level scrollbar remains useful.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: pass via a clean local `view --no-open` build at `[local viewer URL]`. The outer scrollbar was visible from the page top, the workspace scrollbar remained limited to the event table, a paginated event was selected, and details plus raw evidence remained reachable below the event stream. The browser console reported no errors or warnings.
- Capture or report: `none`

## Limitations and follow-up

The event table remains paginated and its table area can scroll within the bounded event panel. This is separate from the page-level scrollbar and prevents a 100-row page from making the initial page unusably tall.
