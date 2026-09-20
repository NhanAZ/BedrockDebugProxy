# AUDIT-0061 - JSON tree glyph marker

- Commit subject: `style: replace native JSON tree marker`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Replace the browser-native triangle shown by JSON `<details>` nodes with the supplied pixel glyph used by the rest of the capture viewer.

## Scope

Hide the native summary marker in Chromium and Firefox, render the embedded `glyph_e031.svg` asset for JSON tree nodes, and rotate the same glyph for collapsed nodes. Capture decoding, JSON structure, and evidence retention are unchanged.

## Evidence and provenance

The screenshot showed a native `▼` marker beside expanded JSON event fields. The viewer already bundles the maintainer-supplied `glyph_e031.svg` asset for dropdown controls. Reusing that same-origin asset keeps the marker artwork consistent without adding a new dependency.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `git diff --check`, and `./tools/quality.ps1` passed before commit. The viewer asset test checks both native marker removal declarations and the shared glyph reference.
- Live validation: passed on the clean build at `http://127.0.0.1:50765/14ed701f93ff83c771027a1110fda25e/`. Browser smoke selected an event with decoded JSON, expanded its tree, and confirmed the native marker is replaced by the bundled pixel glyph.
- Capture or report: `none`

## Limitations and follow-up

The SVG artwork is bundled as an embedded project asset. Its source file is the maintainer-supplied `C:\Users\NhanAZ\Downloads\glyph_e031.svg`.
