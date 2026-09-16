# AUDIT-0018 - Permit the controlled release checker in docs-only deltas

- Commit subject: `Permit the release checker in docs-only validation`
- Change type: `validation tooling | documentation | process`
- Runtime impact: `no`

## Intent

Allow the docs-only release equivalence introduced by AUDIT-0017 to cover the release-checker update that
implements that equivalence, without treating the checker as product runtime or permitting unrelated tooling.

## Scope

Permit only `tools/check-release-readiness.ps1` in addition to the approved root and `docs/` Markdown paths.
Clarify the exception and its trust boundary in `docs/decisions/0007-docs-only-release-equivalence.md`,
`docs/validation.md`, `docs/releasing.md`, and `AGENTS.md`. No product runtime, CLI, packet, transport, capture,
authentication, resource-pack, or schema behavior changes.

## Root cause or reason

The first implementation correctly failed closed on all non-Markdown paths, but that made the candidate containing
the checker implementation unable to use its own policy. The allowlist is widened only for this exact checker and
the change remains subject to PowerShell syntax, quality, CI, ancestry, and fail-closed path tests.

## Evidence and provenance

The committed delta from runtime revision `2a94412e11139184cc2b949dca75b18d0f4048e5` to this candidate contains
the approved Markdown files and `tools/check-release-readiness.ps1`. The five reports remain unchanged under the
ignored `validation/local/2a94412e11139184cc2b949dca75b18d0f4048e5/` directory.

## Validation

- Automated checks: `git diff --check`; `tools/quality.ps1`; exact-mode checker pass; docs-only checker pass; and
  a negative checker run that rejects a candidate delta containing product code
- Live validation: inherited from the five reports for runtime revision `2a94412e11139184cc2b949dca75b18d0f4048e5`
  because the product runtime is unchanged
- Capture or report: all five reports retain their original `tested_revision`; no report facts were edited

## Limitations and follow-up

Only the release checker is accepted as a non-Markdown path, and only when all other changes are approved
documentation. Any product code, other tool, workflow, configuration, generated file, or protocol-affecting
change requires a fresh exact-revision matrix. Release notes must identify both revisions and the normal release,
asset, CI, signing, and private-backup checks remain required.
