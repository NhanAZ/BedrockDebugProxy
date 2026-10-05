# AUDIT-0202 - Read formatted Mojang protocol release metadata

- Commit subject: `fix: preserve Mojang protocol watcher metadata`
- Change type: `tooling | documentation`
- Runtime impact: `no`

## Intent and root cause

The official release body wraps the protocol value in Markdown emphasis, while the watcher accepted only plain digits. It therefore reported `protocol unknown` and could miss a protocol-only change. Its issue template also escaped the variable names instead of interpolating their values, and PowerShell's parsed publication timestamp lost its UTC form when cast to a string.

## Scope

Accept plain digits or balanced italic, bold, and inline-code markers. Preserve unknown status for malformed emphasis. Interpolate the actual game version, protocol, tag, immutable commit, and UTC publication time in the existing issue table. Document the watcher formats and add mocked end-to-end script tests within the existing CLI test suite. No new product command, watcher flag, dependency, or workflow is added.

## Evidence and validation

The [maintenance review](../research/protocol-maintenance-2026-10-05.md) records the actual Mojang release and watcher reproduction. PowerShell's `Invoke-RestMethod` returns the release publication time as a DateTime on the test host, which confirms the timestamp path.

- Automated checks: the previous watcher failed italic, bold, inline-code, and protocol-only detection cases. The updated mocked suite passed all seven cases, including unchanged-release silence, a protocol-only issue, and an issue for a new game version with protocol 2193. All HTTP requests and issue posts are intercepted by the test. The read-only real API check now reports `protocol 2193` and no new stable release. `tools/format.ps1`, `tools/quality.ps1 -LowMemory`, and `git diff --check` passed on the reviewed maintenance source. Static analysis reported zero issues.
- Live validation: no Minecraft server or client was contacted.
- Capture or report: none. No real issue was created or modified.

## Limitations

This change improves the existing notification workflow. It does not upgrade the packet model, adopt preview schemas, or establish real-world compatibility. Release and GitHub issue decisions remain under the existing maintainer workflow.
