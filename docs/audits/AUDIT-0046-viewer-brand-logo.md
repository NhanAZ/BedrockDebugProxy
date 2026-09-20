# AUDIT-0046 - Use favicon package logo in viewer brand mark

- Commit subject: `style: use favicon package logo in viewer brand mark`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Keep the viewer header logo consistent with the favicon and touch icon artwork supplied for the viewer.

## Scope

Embedded `original-logo.png`, served it from the token-scoped viewer handler, and replaced the CSS gradient decoration in the header with the image at the same 38px size. Capture data, page layout, and protocol behavior are unchanged.

## Evidence and provenance

The asset was supplied in `C:\Users\NhanAZ\Downloads\favicon-package\original-logo.png` alongside the deployed favicon variants.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: pass via a clean local `view --no-open` build at `http://127.0.0.1:49242/096eb76741caf798d74f337118b8695c/`. The header now renders the supplied pixel logo at the existing mark size, matching the favicon artwork. No browser errors or warnings were visible during the smoke test.
- Capture or report: `none`

## Limitations and follow-up

The logo is a static viewer asset. It is not written to captures or used in protocol processing.
