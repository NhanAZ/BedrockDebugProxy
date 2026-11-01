# AUDIT-0249 - Prepare the v0.5.0 release changelog

- Commit subject: `Prepare the v0.5.0 release`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Move the reviewed user-facing changes from `Unreleased` into the dated v0.5.0 section and leave a clean section for subsequent development.

## Scope

- Update the root `CHANGELOG.md` with the v0.5.0 date, supported Bedrock version, release link, and comparison link.
- Preserve the reviewed protocol, transfer, Enchanted validation, release-stewardship, and compatibility summaries.
- Leave `Unreleased` empty for future changes.

No product code, CLI behavior, capture format, protocol implementation, dependency, or validation logic changes in this commit.

## Evidence and provenance

The release section was reviewed against the commits since v0.4.2, their per-commit audit records, the current release procedure, and the exact candidate validation reports. The live reports were produced by the preceding runtime candidate `b2a6651209f2460cc1b44132a9f027fca7855ebd`; this commit is a documentation-only release-preparation delta.

## Validation

- The six-server release-readiness gate passed for the validated runtime revision before this documentation-only delta.
- The candidate must use the documented docs-only equivalence gate with the validated runtime revision before tagging.
- The normal quality and stamped build gates remain required for the release candidate.

## Limitations and follow-up

This record does not claim retail Minecraft compatibility or universal compatibility with every server route. The release notes must identify the inherited runtime evidence when the candidate revision differs from the validated runtime revision.

## Security and privacy

No credentials, captures, server-provided addresses, resource-pack keys, decrypted packs, private maintainer instructions, or third-party payloads are stored in this record.
