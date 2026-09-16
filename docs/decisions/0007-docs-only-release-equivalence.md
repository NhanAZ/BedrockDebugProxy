# ADR 0007 - Reuse live reports across a docs-only release delta

## Status

Accepted on 2026-09-16.

## Context

The release candidate may gain changelog, audit, or workflow documentation after the five-server live matrix
has been completed. Requiring the same real-client sessions again for a candidate whose committed diff contains
only Markdown adds cost without testing a changed runtime. Editing a report's `tested_revision` would be
incorrect because that field records the binary and source revision actually exercised.

## Decision

Keep each report's original `tested_revision` and allow the release gate to accept a later candidate only when
the maintainer supplies the validated runtime revision explicitly and that revision is an ancestor of the
candidate. The gate inspects the committed diff and permits only Markdown documentation in the root release
files and `docs/`. Any code, tool, workflow, configuration, generated, or other non-Markdown change requires a
fresh exact-revision matrix.

The release notes must identify both revisions and state that live evidence was inherited from the validated
runtime revision. The candidate binary still carries its own candidate revision; the equivalence does not claim
that the older binary was rebuilt from the candidate commit.

## Consequences

- Existing live evidence remains truthful and reusable for a strictly documentation-only release delta.
- The validation command has an explicit `-ValidatedRevision` input and fails closed when the ancestry or path
  allowlist check does not pass.
- Documentation-only changes must still be reviewed, pass automated quality and CI checks, and be recorded in an
  audit entry.
- Runtime, protocol, transport, capture, authentication, resource-pack, workflow, or generated-file changes
  cannot use this exception.

## Evidence and validation

- The candidate delta from `2a94412e11139184cc2b949dca75b18d0f4048e5` to the current release preparation commit
  contains only Markdown files.
- The five existing reports remain under the ignored `validation/local/2a94412e11139184cc2b949dca75b18d0f4048e5/`
  directory and retain their original tested revision.
- The checker has focused tests through its quality-gate PowerShell syntax and command-path checks. The next
  release run must invoke the checker with both candidate and validated runtime revisions.
