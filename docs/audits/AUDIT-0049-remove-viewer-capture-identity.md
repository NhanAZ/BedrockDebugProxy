# AUDIT-0049 - Remove viewer capture identity header metadata

- Commit subject: `style: remove viewer capture identity`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Remove the verification pill and capture identifier from the viewer header so the compact overview remains focused on packet-review metrics.

## Scope

Removed the status and capture identity markup, styles, and JavaScript updates. The capture manifest, session API, verification state, and stored evidence are unchanged.

## Evidence and provenance

The reported screenshot identified the right-side `verified` pill and UUID as unwanted header content. The requested layout keeps the brand and compact Events, Blobs, Duration, and Generator overview only.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: pass via a clean local `view --no-open` build at `[local viewer URL]`. The browser smoke test shows only the brand and compact Events, Blobs, Duration, and Generator overview in the header. The verification pill and capture identifier are absent, while filters, details, Blob bytes, and the footer remain available. No browser errors or warnings were visible during the smoke test.
- Capture or report: `none`

## Limitations and follow-up

Capture identity and verification remain available through the local session API and canonical capture metadata rather than the visual header.
