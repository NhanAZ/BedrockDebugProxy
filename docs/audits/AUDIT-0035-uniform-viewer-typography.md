# AUDIT-0035 - Uniform normal-weight viewer typography

- Commit subject: `style: normalize viewer typography`
- Change type: `documentation | runtime`
- Runtime impact: `no - viewer presentation only; proxy forwarding and capture behavior are unchanged`

## Intent

Make the local capture viewer use one stable medium text size and no bold styling across its visible interface and evidence views.

## Scope

- Set visible viewer text to a single 14px size.
- Remove bold weights from headings, labels, status text, table headers, packet rows, metadata, JSON, hex evidence, and footer text.
- Use regular Monocraft as the embedded heading fallback while preserving optional browser-local Mojangles and Minecraft Ten lookup.
- Update the user-facing changelog, provenance notes, analysis documentation, and fix-history index.
- Add a stylesheet regression assertion for the size and weight contract.

## Evidence and provenance

- The pre-change stylesheet used 9px, 10px, 11px, 12px, 16px, 17px, and 18px declarations plus 650, 700, and 750 weights.
- The current embedded Monocraft asset remains the previously audited SIL Open Font License 1.1 fallback. No new font file is introduced.
- The supplied proprietary Mojangles and Minecraft Ten files remain local-only and are not copied into the repository, executable, or license bundle.

## Validation

- Automated checks: pass - `tools/format.ps1`, `go test ./internal/viewer`, `tools/quality.ps1`, and `git diff --check` all passed. The quality gate also completed the full test suite, build, static analysis, vulnerability scan, and third-party license inventory.
- Live validation: `not_observed - presentation-only behavior; browser rendering is platform-specific`
- Capture or report: `none`

## Limitations and follow-up

The CSS size contract does not control browser zoom or operating-system text scaling. The older Minecraft-Bold asset remains available from its existing protected route but is not selected by the current stylesheet.
