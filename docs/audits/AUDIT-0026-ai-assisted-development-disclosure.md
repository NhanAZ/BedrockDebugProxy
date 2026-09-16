# AUDIT-0026 - Add AI-assisted development disclosure

- Commit subject: `docs: disclose AI-assisted development`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Make the project's AI-assisted maintenance transparent and set respectful expectations for readers and contributors.

## Scope

Add a README section describing AI-agent assistance, maintainer responsibility, review and validation expectations, and respectful participation. No code, CLI, capture schema, release workflow, license, or server policy behavior changes.

## Evidence and provenance

- Maintainer request received on 2026-09-17.
- The repository's `AGENTS.md` requires evidence, provenance, review, and proportional validation for accepted changes.
- The existing README already documents the project's scope, security boundaries, and server-policy disclaimer.

## Validation

- Automated checks: `git diff --check` and the project quality gate `.\tools\quality.ps1`.
- Live validation: `not_observed` because this is documentation-only.
- Capture or report: `none`.

## Limitations and follow-up

This disclosure has no runtime impact and does not claim that every line of the project is AI-generated. It does not replace technical review, provenance checks, or the project's contribution rules.
