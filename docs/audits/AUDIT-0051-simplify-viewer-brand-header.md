# AUDIT-0051 - Simplify viewer brand header

- Commit subject: `style: simplify viewer brand header`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Reduce header clutter by keeping only the BedrockDebugProxy name in the visible brand area.

## Scope

Remove the decorative header logo and the secondary Capture Viewer label. Favicon and web manifest assets remain available for browser metadata. Capture overview metrics, filters, event data, and viewer APIs are unchanged.

## Evidence and provenance

The reported screenshot identified the logo and Capture Viewer line as unwanted header content. The replacement uses a single text brand element and does not alter capture data or serving boundaries.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: passed on the clean build at `[local viewer URL]`. Browser smoke confirmed that only BedrockDebugProxy remains in the brand area, the decorative logo and Capture Viewer text are absent, and the overview metrics remain aligned. No browser errors or warnings were visible during smoke test.
- Capture or report: `none`

## Limitations and follow-up

The favicon package remains embedded because it provides browser tab and install metadata even though the header no longer displays its logo.
