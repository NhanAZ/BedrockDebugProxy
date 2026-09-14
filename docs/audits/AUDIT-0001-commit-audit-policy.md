# AUDIT-0001 - Require per-commit audit records

- Commit subject: `Require per-commit audit records`
- Change type: `documentation | process`
- Runtime impact: `no`

## Intent

Make every authored commit easy for an AI agent or maintainer to understand without reconstructing intent from the diff alone.

## Scope

Add the commit audit policy to `AGENTS.md` and `CONTRIBUTING.md`. Add the audit record guide and reusable template under `docs/audits/`.

This policy applies from this commit forward. Existing commits are not rewritten because Git history must remain recoverable and shared history must not be force-rewritten.

## Evidence and provenance

The policy follows the existing project requirements to inspect history before changes, record protocol findings in research notes, record material decisions in ADRs, and keep exact-revision validation evidence separate from raw captures. Git metadata remains authoritative for commit identity.

## Validation

- Automated checks: `git diff --check`
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

Older commits do not contain these records. Their intent must be reconstructed from Git messages, existing ADRs, research notes, captures, and validation reports. Future authored commits must include exactly one audit record staged with the change.
