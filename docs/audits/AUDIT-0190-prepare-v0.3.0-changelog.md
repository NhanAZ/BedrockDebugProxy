# Audit AUDIT-0190 - Prepare v0.3.0 changelog

- Commit subject: `Prepare v0.3.0 changelog`
- Change type: `release documentation`
- Runtime impact: `no`

## Intent

Move the reviewed user-visible changes accumulated since v0.2.1 into the v0.3.0 changelog section and leave an empty Unreleased section for subsequent work.

## Scope

- Date the candidate section 2026-09-25 and identify Bedrock 1.26.51 with protocol 2193.
- Link the candidate release and comparison range to v0.2.1.
- Preserve the reviewed Added, Changed, and Fixed summaries without modifying runtime behavior.
- No validation reports, captures, source files, or tested revisions changed.

## Evidence and provenance

The version 0.3.0 is the settled release candidate. The supported game and protocol metadata match the candidate binary and the project protocol baseline. The six-server reports are tied to runtime revision `d7b0f35f77cffec08e8ffa91d1b43021f973c415`; candidate `89d5c2b4a399de426545075c329e3257d3c3bf67` passes the documented Markdown-only equivalence gate against that runtime revision.

## Validation

- Documentation review confirms the existing Unreleased summaries are preserved under v0.3.0 and the new Unreleased section is empty.
- `git diff --check` passed.
- No tests were run because this commit changes documentation only. The release quality gate and candidate build remain required after this changelog update.

## Live validation status

Not applicable. The change does not alter the runtime and does not edit the six server reports.

## Limitations and follow-up

The GitHub release and tag links are prospective until publication. The release remains incomplete until the final candidate quality gate, stamped build, exact-commit CI, publication, asset verification, and recovery-mirror synchronization pass.
