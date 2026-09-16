# AUDIT-0019 - Add controlled Dependabot checks

- Commit subject: `Add controlled Dependabot checks`
- Change type: `process`
- Runtime impact: `no`

## Intent

Add scheduled dependency discovery without allowing automated updates to bypass protocol review, validation, or the local patches maintained in the `third_party/` trees.

## Scope

Added `.github/dependabot.yml` for the root Go module and GitHub Actions. Added contributor guidance for reviewing Dependabot pull requests and a changelog entry under `Unreleased`. The local `third_party/gophertunnel` and `third_party/go-raknet` modules remain outside Dependabot's update directories. No product code, dependency version, workflow job, or runtime behavior changes in this commit.

## Evidence and provenance

The repository had no existing `.github/dependabot.yml`; the GitHub repository setting `dependabot_security_updates` was enabled, but scheduled version updates were not configured. The root `go.mod` uses local replacements for the two patched third-party trees, so including those directories in an automatic update schedule could overwrite or bypass project-specific patches.

## Validation

- Automated checks: `tools/quality.ps1` passed, including formatting, module integrity, tests, build, static analysis, vulnerability scan, tracked-artifact, and whitespace checks.
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

Dependabot will propose updates for the root module and Actions only. A dependency PR still needs the normal quality, CI, provenance, and runtime validation gates when its diff can affect observable behavior. No runtime impact.
