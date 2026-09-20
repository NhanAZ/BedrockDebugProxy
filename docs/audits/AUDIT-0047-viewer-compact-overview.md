# AUDIT-0047 - Compact capture overview in viewer header

- Commit subject: `style: compact viewer capture overview`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Reduce the vertical space used by the capture overview while keeping all four metrics visible.

## Scope

Moved the overview into the header row and rendered Events, Blobs, Duration, and Generator as compact horizontal label/value pairs. Capture data, viewer filtering, and protocol behavior are unchanged.

## Evidence and provenance

The reported viewer showed a separate overview strip consuming a full row below the header. The compact header keeps the existing metric IDs and labels and only changes layout and placement.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: pass via a clean local `view --no-open` build at `[local viewer URL]`. The browser smoke test shows the brand, all four metrics, status, and capture ID sharing one header row, with the event stream and filters below it. The long Generator value ellipsizes within its metric cell instead of creating a second row. No browser errors or warnings were visible during the smoke test.
- Capture or report: `none`

## Limitations and follow-up

The compact header assumes the existing minimum page width. Long generator and capture IDs may ellipsize in narrow windows.
