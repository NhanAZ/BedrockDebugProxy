# AUDIT-0017 - Allow evidence-preserving docs-only release deltas

- Commit subject: `Allow docs-only release validation reuse`
- Change type: `validation tooling | documentation | process`
- Runtime impact: `no`

## Intent

Avoid repeating five real-client sessions when a release candidate adds only reviewed Markdown documentation
after an already validated runtime revision, while keeping the original report facts immutable.

## Scope

Update `tools/check-release-readiness.ps1` with an explicit `-ValidatedRevision` mode, document its ancestry and
Markdown-path allowlist in `AGENTS.md`, `docs/validation.md`, `docs/releasing.md`, and add ADR 0007. No product
runtime, CLI, packet, transport, capture, authentication, resource-pack, or schema behavior changes.

## Root cause or reason

The existing release gate required `tested_revision` to equal the candidate revision even when only documentation
commits followed the runtime commit. Changing that field would falsify which binary produced the evidence. The
new gate verifies that the validated runtime revision is an ancestor and that every intervening committed path is
Markdown documentation in the approved locations.

## Evidence and provenance

The candidate delta from `2a94412e11139184cc2b949dca75b18d0f4048e5` to the release-preparation commits was
reviewed with `git diff --name-status` and contains only Markdown files. The five existing reports remain
unchanged under the ignored `validation/local/2a94412e11139184cc2b949dca75b18d0f4048e5/` directory.

## Validation

- Automated checks: `git diff --check`; `tools/quality.ps1`; focused checker dry run with the ancestry and path
  allowlist on the current candidate
- Live validation: inherited only for a docs-only candidate; no runtime behavior changed
- Capture or report: the existing five reports retain `tested_revision=2a94412e11139184cc2b949dca75b18d0f4048e5`

## Limitations and follow-up

The exception is intentionally narrow and fails closed for code, tools, workflows, configuration, generated files,
or any non-Markdown path. It does not claim that the candidate binary itself was run by the older live sessions.
Release notes must identify both revisions, and the normal quality, CI, release, asset, and backup checks remain
required.
