# AUDIT-0052 - Compact viewer controls and result status

- Commit subject: `style: compact viewer controls and result status`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Reduce unused viewer chrome while keeping result context and table values readable.

## Scope

Shorten the overview header, widen the Time column to 132px, move the live result count into the event panel heading, remove the footer and search magnifier, and use Space Grotesk for the Ctrl F hint when the font is available. Capture data, APIs, filters, and packet behavior are unchanged.

## Evidence and provenance

The reported screenshots showed excess header height, truncated Time values, a footer that should be removed, a result count area in the event heading, and an unwanted magnifier. The Space Grotesk declaration is a local/system font preference with a generic fallback, so the viewer remains offline and does not fetch a new font asset.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: passed on the clean build at `[local viewer URL]`. Browser smoke confirmed the shorter overview header, full millisecond Time values, in-panel result count, removed footer and magnifier, and readable Ctrl F hint. No browser errors or warnings were visible during smoke test.
- Capture or report: `none`

## Limitations and follow-up

Space Grotesk is selected when installed or otherwise available to the browser. Browsers without it use the declared system sans-serif fallback.
