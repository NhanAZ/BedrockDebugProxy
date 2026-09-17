# AUDIT-0037 - Require same-change web validation

- Commit subject: `ci: require same-change web validation`
- Change type: `process`
- Runtime impact: `no - validation and contributor guidance only`

## Intent

Prevent web-facing changes from being merged with an untested or broken capture viewer. A successful Go build does not prove that embedded HTML, CSS, JavaScript, HTTP asset routes, or browser interactions work.

## Scope

- Add the same-change web validation rule to `AGENTS.md`, `CONTRIBUTING.md`, `docs/validation.md`, the pull request template, and the release workflow.
- Add [`docs/web-validation.md`](../web-validation.md) with the automated gate and browser smoke checklist.
- Extend `tools/quality.ps1` to fail when the embedded viewer JavaScript does not pass `node --check`.
- Extend viewer HTTP tests to verify linked page assets, filter controls, and the JavaScript asset.
- Add a concise Unreleased changelog entry and this audit record.
- Intentionally unchanged: proxy forwarding, capture storage, protocol behavior, viewer production assets, and CLI semantics.

## Evidence and provenance

- The viewer embeds its HTML, CSS, JavaScript, and fonts in the Go binary, so HTTP asset tests exercise the same serving boundary used by operators.
- Existing viewer tests already cover closed and live capture indexing, loopback serving, and response boundaries. This change makes the required asset linkage and JavaScript syntax checks explicit.
- No external source, dependency, credential, capture, or generated asset is introduced.

## Validation

- Automated checks: pass - `tools/format.ps1`, `tools/quality.ps1`, and `git diff --check` passed. The quality gate ran the viewer JavaScript syntax check, full Go tests including viewer HTTP tests, build, static analysis, vulnerability scan, third-party license inventory, and tracked-artifact checks.
- Live validation: `not_observed - process and test-only change`
- Capture or report: `none`

## Limitations and follow-up

The automated checks do not replace a browser smoke check for visual or interaction changes. Contributors must open the clean committed build and record the browser result in the audit for those changes.
