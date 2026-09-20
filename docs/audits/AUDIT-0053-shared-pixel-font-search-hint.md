# AUDIT-0053 - Shared pixel font for search hint

- Commit subject: `style: use shared pixel font for search hint`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Keep the Ctrl F hint consistent with the single pixel-font presentation used throughout the viewer.

## Scope

Replace the search hint's Space Grotesk override with the existing `var(--font-content)` stack. Search behavior, keyboard handling, capture data, and all other layout behavior are unchanged.

## Evidence and provenance

The reported screenshot showed Ctrl F rendered in a different non-pixel style. The project owner clarified that the entire web viewer should use one pixel font. The existing stack resolves locally installed Minecraft Ten first and otherwise uses the embedded Monocraft fallback.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: passed on the clean build at `http://127.0.0.1:50196/70c7428d1c5aa63debb219b677aa8d94/`. Browser smoke confirmed Ctrl F matches the pixel font used by the search field and other viewer text. No browser errors or warnings were visible during smoke test.
- Capture or report: `none`

## Limitations and follow-up

The exact glyph shape depends on whether the operator has Minecraft Ten installed locally. The embedded Monocraft font remains the deterministic fallback.
