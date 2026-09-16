# AUDIT-0020 - Add Mojang protocol release watcher

- Commit subject: `Add Mojang protocol release watcher`
- Change type: `process`
- Runtime impact: `no`

## Intent

Detect new official Minecraft Bedrock protocol schema releases and create a review issue so a later maintainer request can run the documented protocol-update process without silently changing packet code.

## Scope

Added `.github/workflows/check-mojang-protocol.yml` and `tools/check-mojang-protocol.ps1`. The workflow checks the latest non-prerelease release in `Mojang/bedrock-protocol-docs`, compares its release tag and release-note protocol number with the baseline in `docs/protocol-updates.md`, resolves the release tag commit when possible, and creates one deduplicated issue when a difference is found. Updated `AGENTS.md`, `docs/protocol-updates.md`, and `CHANGELOG.md` to define the notification and triage boundary. No product code, packet definition, dependency version, or capture behavior changes.

## Evidence and provenance

The official [Mojang bedrock-protocol-docs repository](https://github.com/Mojang/bedrock-protocol-docs) states that it publishes packet, type, and enum schemas and that generated documentation is assembled from GitHub Releases. The current latest release is `v1.26.50`, whose release notes state game version `1.26.50` and network protocol `2193`, matching this project's baseline. The workflow therefore tracks Releases rather than the mutable `main` branch.

## Validation

- Automated checks: `tools/check-mojang-protocol.ps1` matched the current `v1.26.50` / protocol `2193` baseline without opening an issue; `tools/quality.ps1` passed, including formatting, module integrity, tests, build, static analysis, vulnerability scan, tracked-artifact, and whitespace checks.
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

The watcher reports release differences but never updates protocol code or dependencies. Release-note protocol metadata may be unavailable, in which case the issue records `unknown` and still requires manual schema verification. The workflow requires repository Actions settings that permit the `GITHUB_TOKEN` to create issues. No runtime impact.
