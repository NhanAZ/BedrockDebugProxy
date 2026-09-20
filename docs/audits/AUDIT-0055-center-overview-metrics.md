# AUDIT-0055 - Center compact overview metrics

- Commit subject: `style: center compact overview metrics`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Align the compact capture overview metrics vertically with the BedrockDebugProxy name.

## Scope

Change the metric flex rows from baseline alignment to center alignment. The header height, metric order, values, typography, and capture data remain unchanged.

## Evidence and provenance

The reported screenshot showed the Events, Blobs, Duration, and Generator text sitting higher than the brand name inside the same header row. The existing metric rows used `align-items: baseline`, which was unnecessary for the single-line label and value pairs.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `git diff --check`, and `./tools/quality.ps1` passed before commit.
- Live validation: passed on the clean build at `http://127.0.0.1:52342/fd78759633a08d005c30c74498ce9a1a/`. Browser smoke confirmed the Events, Blobs, Duration, and Generator metrics are vertically centered with the BedrockDebugProxy name. No browser errors or warnings were visible during smoke test.
- Capture or report: `none`

## Limitations and follow-up

The change is limited to the header metric alignment and does not alter responsive breakpoints or page-level scrolling.
