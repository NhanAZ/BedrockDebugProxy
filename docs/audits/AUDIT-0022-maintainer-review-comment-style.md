# AUDIT-0022 - Clarify maintainer review comment style

- Commit subject: `Clarify maintainer review comment style`
- Change type: `documentation | review process`
- Runtime impact: `no`

## Intent

Make public review comments sound like concise, respectful maintainer communication and avoid mechanical label-style
fragments or unnecessary semicolons. Give bot-authored dependency pull requests a shorter default comment while
preserving enough evidence for audit and debugging.

## Scope

Updated `AGENTS.md`, `CONTRIBUTING.md`, and `CHANGELOG.md` with guidance for authorized maintainer-account comments,
Dependabot review comments, natural sentence structure, and evidence boundaries. The existing review comments on PR #2
and PR #3 were also edited on GitHub to apply the guidance. No runtime code, CLI behavior, capture format, protocol
implementation, validation logic, release assets, or credentials changed.

## Evidence and provenance

The guidance responds to the public comments on [PR #2](https://github.com/NhanAZ/BedrockDebugProxy/pull/2) and [PR #3](https://github.com/NhanAZ/BedrockDebugProxy/pull/3). The original text used `Review status:`,
`Closing this PR; please`, and `Closed after review: this tag`. The revised text keeps the dependency provenance finding
and validation outcome while removing those awkward constructions. The review decisions remain based on the immutable
PR heads and the current dependency graph.

## Validation

- Automated checks: `git diff --check`; documentation and public comment text reviewed locally and through the GitHub API.
- Live validation: `not_observed`; documentation and communication-only change with no runtime impact.
- Capture or report: `none`.

## Limitations and follow-up

This policy applies to future public issue and pull request comments when the maintainer has explicitly authorized the
agent to post from the maintainer account. It does not authorize external communication by itself. Future review
comments should remain concise unless additional evidence is needed.
