# AUDIT-0027 - Reorder README context sections

- Commit subject: `docs: move README context sections near the end`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Improve the README reading flow by keeping product usage and technical documentation near the beginning and moving contextual disclaimers to the closing sections.

## Scope

Move the `Server policy disclaimer` and `AI-assisted development` sections from the opening of `README.md` to the end of the document before the license section. The text and links remain unchanged. No code, CLI behavior, capture schema, release workflow, or policy behavior changes.

## Evidence and provenance

- Maintainer requested the section-order change on 2026-09-17.
- The change is limited to the existing README structure and preserves the project's current disclosures.

## Validation

- Automated checks: `git diff --check` and `.\tools\quality.ps1`.
- Live validation: `not_observed` because this is documentation-only.
- Capture or report: `none`.

## Limitations and follow-up

This is a documentation-only ordering change with no runtime impact.
