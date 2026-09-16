# AUDIT-0015 - Add a user-facing changelog workflow

- Commit subject: `Add a user-facing changelog workflow`
- Change type: `documentation | process`
- Runtime impact: `no`

## Intent

Add a concise release history so maintainers and users can understand the changes accumulated between releases without reading every commit or technical audit.

## Scope

Add the root `CHANGELOG.md` with an `Unreleased` section covering the eight commits after `v0.1.0`. Document when agents and contributors update that section, distinguish it from per-commit audit records, and make release preparation move reviewed entries into a dated version section. No runtime code, CLI behavior, capture format, protocol implementation, validation logic, or release assets are changed.

## Evidence and provenance

The entries summarize `git log v0.1.0..HEAD`, the existing commit audit records, `docs/fix-history.md`, and `docs/releasing.md`. The project already required one audit record per authored commit, but it had no `CHANGELOG.md` and no rule for maintaining a user-facing summary.

## Validation

- Automated checks: `git diff --check`; documentation structure and links reviewed locally
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

The initial `Unreleased` section is a concise reconstruction of the commits after `v0.1.0`, not a replacement for their detailed evidence. Future user-visible changes should update `Unreleased` in the same commit. Internal-only commits may omit a changelog entry. No runtime impact.
