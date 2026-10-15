# AUDIT-0045 - Add viewer favicon package

- Commit subject: `style: add viewer favicon package`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Give the local capture viewer a recognizable browser icon and standard install metadata without changing capture or protocol behavior.

## Scope

Embedded the supplied favicon package, served the icon variants and web manifest from the token-scoped viewer handler, and linked them from the viewer document head. The manifest uses the viewer's dark theme and resolves its scope to the current token root.

## Evidence and provenance

The assets were supplied in `the supplied icon package`. Only the generated icon files and a project-local manifest were copied into the embedded viewer. The package's README states that the source logo was not uploaded, so no external source code or third-party license text was imported.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: pass via a clean local `view --no-open` build at `[local viewer URL]`. The viewer loaded from the committed asset set, its document head contained the favicon and manifest links, and all seven icon variants plus `site.webmanifest` returned HTTP 200 from token-scoped paths. No browser errors or warnings were visible during the smoke test.
- Capture or report: `none`

## Limitations and follow-up

The manifest is scoped to the local token URL and is not intended as a public service worker or cross-capture identity. The favicon assets are static and are not stored in captures.
