# Audit AUDIT-0193 - Require private capture archive citations

- Commit subject: `docs: require private capture archive citations`
- Change type: `documentation policy`
- Runtime impact: `no`

## Intent

Make a verified private archive link the required citation for capture evidence in project records, and prevent local filesystem paths from being used as evidence citations.

## Scope

- Add the per-capture private repository, release asset, SHA-256 verification, citation, and local-retention workflow to `AGENTS.md`.
- Align the validation guide with the archive-first citation rule.
- Mark the older Hive resource-pack observation unverified because its source session ID and archive link are unavailable.
- Do not change runtime code or capture contents.

## Evidence and provenance

This change implements the maintainer's instruction to initialize a private archive repository in `BedrockDebugProxy-Captures` before citing a capture, and to cite its release URL instead of a local path. The existing 131 archive links still resolve to 35 same-named private repositories. No new capture was uploaded or deleted.

## Validation

- Automated checks: the documentation scan found 131 valid archive links across 35 targets and no local session-path citations. `git diff --check` passed.
- Full project quality and test gates were not run because this change only edits Markdown documentation.
- Live validation: not applicable. This change has no runtime impact.

## Limitations and follow-up

The older Hive observation in `docs/research/resource-pack-negotiation.md` remains unverified until its source capture can be identified and archived. The new workflow requires the maintainer to resolve that missing provenance before relying on the observation again.
