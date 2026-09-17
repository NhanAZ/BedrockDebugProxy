# AUDIT-0034 - Prefer locally installed Mojang fonts

- Commit subject: `style: prefer local Mojang fonts`
- Change type: `documentation | runtime`
- Runtime impact: `no - viewer CSS selection only; proxy forwarding and capture behavior are unchanged`

## Intent

Allow an operator who has separately installed the official Mojangles or Minecraft Ten font to see that font in the local capture viewer without copying proprietary files into the project or serving them from the viewer.

## Scope

- Add browser `local()` sources for the exact `Mojangles` and `Minecraft Ten` family names before the existing embedded OFL fallbacks.
- Document the local-only behavior, license boundary, and fallback behavior.
- Add an HTTP regression assertion for both local font sources.

## Evidence and provenance

- The supplied `Mojangles.ttf` name table identifies family `Mojangles` and states that the font is a commercial product whose license holder should be contacted for permission.
- The supplied `minecraft-ten.ttf` name table identifies family `Minecraft Ten` and carries a Mojang copyright notice with no redistribution license.
- The official [Minecraft Usage Guidelines](https://www.minecraft.net/en-us/usage-guidelines) identify fonts as part of the Minecraft brand, reserve rights, and prohibit redistribution of game files.
- The project continues to embed only the two previously audited SIL Open Font License 1.1 fan-made fallback fonts. The proprietary files are not copied, hashed into the build, or exposed through an asset route.

## Validation

- Automated checks: pass - `tools/format.ps1`, `go test ./internal/viewer`, `tools/quality.ps1`, and `git diff --check` all passed. The quality gate also completed the full test suite, build, static analysis, vulnerability scan, and third-party license inventory.
- Live validation: `not_observed - presentation-only behavior; local font availability is machine-specific`
- Capture or report: `none`

## Limitations and follow-up

The browser must have the font installed and permitted for local lookup. If it is unavailable, the embedded OFL fallback remains in use. This change does not grant a license to redistribute or publish official Mojang font files.
