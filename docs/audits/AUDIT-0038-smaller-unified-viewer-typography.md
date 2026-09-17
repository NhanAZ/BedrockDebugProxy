# AUDIT-0038 - Smaller unified viewer typography

- Commit subject: `style: unify viewer typography`
- Change type: `runtime`
- Runtime impact: `no - viewer presentation only; proxy forwarding and capture behavior are unchanged`

## Intent

Make the local capture viewer easier to scan by reducing the oversized text, removing inconsistent visible font faces, and restoring natural casing instead of forcing labels into all caps.

## Scope

- Use one shared 12px size variable for visible viewer text.
- Use regular Monocraft for headings, controls, metadata, packet rows, JSON, raw evidence, and footer text, with local Minecraft Ten lookup before the embedded fallback.
- Remove the unused separate heading face and all uppercase transforms.
- Change static all-caps eyebrow labels to natural casing.
- Update the viewer regression contract, changelog, fix-history index, and current font documentation.

## Evidence and provenance

- The pre-change stylesheet repeated 14px declarations and selected a separate heading face through browser-local Mojangles lookup.
- The embedded Monocraft asset remains the previously audited SIL Open Font License 1.1 fallback. No font file was added or copied from the operator's local files.
- The protected Minecraft-Bold asset route remains available for compatibility but is not selected by the viewer stylesheet.

## Validation

- Automated checks: pass - `tools/format.ps1`, `go test ./internal/viewer`, `tools/quality.ps1`, and `git diff --check` all passed. The quality gate also completed the full test suite, build, static analysis, vulnerability scan, and third-party license inventory.
- Live validation: pass - exact binary revision `04869ec31093c5efc9699eb859efd2328da94da7` served closed capture `session-20260917T034859Z` in Chrome. Computed styles for body, headings, labels, table headers, controls, and footer were 12px, weight 400, Monocraft, and no text transform. The themed direction control opened and updated the native value to `server_to_client`.
- Capture or report: `none`

## Limitations and follow-up

The shared CSS size does not override browser zoom, operating-system scaling, or glyph fallback for characters absent from the selected font. A browser smoke check is still required to confirm readability at the operator's viewport size.
