# AUDIT-0021 - Adopt a single-file changelog release format

- Commit subject: `Adopt a single-file changelog release format`
- Change type: `documentation | release process`
- Runtime impact: `no`

## Intent

Make release history easier to read without splitting it into one changelog file per version. Keep the root changelog
as the canonical curated summary and expose the complete release history through GitHub's compare view.

## Scope

Updated `CHANGELOG.md` with a single-file policy, release metadata, and release and compare links for the published
versions. Updated `docs/releasing.md` with the required release-body link format and the rule to keep all versions in
the root changelog. No runtime code, CLI behavior, capture format, protocol implementation, validation logic, or
release assets changed.

## Evidence and provenance

The format is based on the existing project changelog and release workflow, and on the published [PocketMine-MP 5.0.0
changelog](https://github.com/pmmp/PocketMine-MP/blob/5.0.0/changelogs/5.0.md), [PocketMine-MP 5.0.0 release](https://github.com/pmmp/PocketMine-MP/releases/tag/5.0.0), and GitHub's [comparing releases
documentation](https://docs.github.com/en/repositories/releasing-projects-on-github/comparing-releases). The project keeps its own concise
summary and audit policy rather than copying PocketMine-MP's release structure or source text.

## Validation

- Automated checks: `git diff --check`; Markdown headings, links, and release instructions reviewed locally.
- Live validation: `not_observed`; documentation-only change with no runtime impact.
- Capture or report: `none`.

## Limitations and follow-up

The compare link shows the complete Git history between tags, while the changelog remains a curated summary and audit
records remain the evidence-backed per-commit record. Future releases must add their section to the root
`CHANGELOG.md` and include both links in the release body. Existing v0.2.0 release notes can be updated separately to
use the same links.
