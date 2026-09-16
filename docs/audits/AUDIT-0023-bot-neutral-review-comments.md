# AUDIT-0023 - Use bot-neutral dependency review comments

- Commit subject: `Use bot-neutral dependency review comments`
- Change type: `documentation | review process`
- Runtime impact: `no`

## Intent

Keep Dependabot review comments distinct from human contributor communication. Routine bot review comments should state
the decision and evidence without greetings, thanks, or other language that treats automation as a person.

## Scope

Updated `AGENTS.md`, `CONTRIBUTING.md`, and `CHANGELOG.md` to prohibit anthropomorphic social language in bot-authored
dependency review comments. Edited the owner-authored comments on PR #2 and PR #3 to remove the opening thanks while
retaining their reviewed validation and dependency-provenance findings. No runtime code, CLI behavior, capture format,
protocol implementation, validation logic, release assets, or credentials changed.

## Evidence and provenance

The earlier revisions of [PR #2](https://github.com/NhanAZ/BedrockDebugProxy/pull/2) and [PR #3](https://github.com/NhanAZ/BedrockDebugProxy/pull/3) began with `Thanks for the update` even though both pull requests were authored by Dependabot. The revised comments use direct maintainer statements and preserve the exact review decisions.

## Validation

- Automated checks: `git diff --check`; documentation and public comment text reviewed locally and through the GitHub API.
- Live validation: `not_observed`; documentation and communication-only change with no runtime impact.
- Capture or report: `none`.

## Limitations and follow-up

This guidance applies to future bot-authored dependency issues and pull requests. Human contributors may still receive a
brief acknowledgement when it is appropriate. Posting from the maintainer account remains subject to explicit
maintainer authorization.
