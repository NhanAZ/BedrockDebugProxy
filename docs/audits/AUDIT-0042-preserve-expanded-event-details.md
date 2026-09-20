# AUDIT-0042 - Preserve the expanded event-details track

- Commit subject: `style: keep viewer details content-sized`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Prevent the Event details panel from collapsing to a thin strip when the viewer is displayed in a constrained browser viewport.

## Scope

Adjusted the embedded viewer grid sizing and its CSS regression checks. Capture parsing, event selection, raw evidence storage, and the forwarding path are unchanged.

## Evidence and provenance

The previous expansion patch removed the internal details scrollbar but left the inspector and detail panel with zero minimum heights. In the page grid, that allowed the content-sized row to resolve to almost zero even though the detail content was visible in the DOM. The revised layout uses max-content tracks and minimums so the outer workspace remains the scroll boundary while the details panel keeps its full content height.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: pass via a clean local `view --no-open` build at `[local viewer URL]`. A constrained browser viewport was scrolled through the details section, an event was selected, the metadata and decoded tree remained visible, and the browser console reported no errors or warnings.
- Capture or report: `none`

## Limitations and follow-up

Large decoded trees can make the workspace scroll vertically. That is intentional. The raw hex preview remains independently bounded and may retain its own scrollbar.
