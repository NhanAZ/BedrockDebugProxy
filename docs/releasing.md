# Release workflow

This is the required path from a release candidate commit to an official BedrockDebugProxy release. `docs/validation.md` defines the evidence in detail. This file defines the order of operations.

## Maintainer and AI agent handoff

When the maintainer says "I want to release", the AI agent should drive this document rather than returning the whole checklist as homework.

1. The agent audits the commits since the latest release, proposes or confirms the version, checks the working tree, documentation, licenses, provenance, known limitations, CI, and unresolved review findings.
2. The agent runs the automated quality gate and builds the exact stamped candidate.
3. The agent asks the maintainer to perform the live Minecraft sessions that cannot be automated. A short instruction should identify the server, target command, actions to perform, approximate duration, and when to press `Ctrl+C`.
4. The maintainer reports that the session is complete and mentions any visible problem. The agent locates and inspects the closed capture, verifies the revision and evidence, and generates the sanitized report. The maintainer does not need to run report scripts unless they are performing the release without an agent.
5. After all required sessions, the agent runs the release gate, prepares the checksum and release notes, and reports any exact blocker.
6. If the maintainer explicitly requested a release, the version is settled, every gate passes, and signing or GitHub access is available, the agent creates and pushes the tag, publishes the release assets, and verifies the result. It must not publish from an incomplete or mismatched validation set.

The remaining sections retain the complete manual procedure so another developer can reproduce and audit what the agent performs.

## 1. Choose the candidate

1. Decide the semantic version, such as `0.1.0`, and set it for the commands below.
2. Confirm the intended release commit is on `main`, all required pull requests are merged, and the working tree is clean.
3. Review user-visible changes, protocol support, known limitations, license notices, and documentation. Do not mix a last-minute refactor or protocol update into release preparation.

```powershell
$version = "0.1.0"
git switch main
git pull --ff-only
git status --short
git log -1 --oneline
```

Do not continue if `git status --short` prints anything.

## 2. Run the automated gate

```powershell
.\tools\quality.ps1
$revision = (git rev-parse HEAD).Trim()
.\tools\build.ps1 -Version $version
.\bin\bedrock-debug-proxy.exe version
```

The printed binary commit must equal `$revision`. Keep this exact binary for all six live tests. Do not rebuild from a different commit between servers.

## 3. Validate the six-server baseline

Use a real Minecraft Bedrock client. For each server, start the same stamped binary, complete a normal session, exercise the named flow, stop with `Ctrl+C`, and wait for the capture verification message.

| Report label | Upstream target | Required manual flow |
| --- | --- | --- |
| `The Hive` | `experience:The Hive` | Connect, spawn, move, interact, and observe normal traffic |
| `CubeCraft` | `experience:CubeCraft` | Connect, spawn, move, interact, and observe normal traffic |
| `Galaxite` | `experience:Galaxite` | Connect, spawn, move, interact, and observe normal traffic |
| `Lifeboat` | `experience:Lifeboat` | Connect, spawn, move, interact, and observe normal traffic |
| `Mineville Zeqa` | `experience:Mineville` | Enter the Zeqa flow and observe it directly |
| `Enchanted` | The maintainer-approved `HOST:PORT` | Connect, spawn, move, interact, and observe normal traffic |

Use this command for an Experience target. Replace only the value after `--upstream` for each run.

```powershell
.\bin\bedrock-debug-proxy.exe `
    run `
    --listen 0.0.0.0:19132 `
    --upstream "experience:The Hive" `
    --auth device
```

For every server, confirm all of the following before recording `pass`.

1. Device authentication and Experience resolution succeed when applicable.
2. The Minecraft client connects to the proxy and reaches spawn.
3. The session remains stable during normal movement and interaction.
4. Client-to-server and server-to-client traffic are present.
5. Decoded packets are present.
6. Offered resource packs complete without an unexplained error.
7. No unexplained decode, protocol, authentication, signaling, or connection error appears.
8. `Ctrl+C` closes the proxy and the capture verifier succeeds.

If a server cannot be tested, create an `incomplete` report. If an anomaly occurs, create a `fail` report and investigate. Neither result passes the release gate.

## 4. Generate one report per server

Use the exact capture directory printed by the matching run. The following example is for The Hive.

```powershell
.\tools\create-validation-report.ps1 `
    -Binary .\bin\bedrock-debug-proxy.exe `
    -Capture .\captures\<THE-HIVE-SESSION> `
    -Server "The Hive" `
    -Result pass `
    -Check @("normal_session=pass", "experience_resolution=pass") `
    -Output ".\validation\local\$revision\the-hive.json"
```

Use these exact `-Server` labels for the other reports: `CubeCraft`, `Galaxite`, `Lifeboat`, `Mineville Zeqa`, and `Enchanted`. Add a focused manual check when a release specifically changes a flow. Never place raw captures, tokens, addresses, resource-pack keys, or third-party assets in a report.

Review each JSON file before using it. Then run the matrix gate.

```powershell
.\tools\check-release-readiness.ps1 -Revision $revision
```

The command must print `pass` for all six servers. This gate verifies report structure and revision matching. It does not replace human review of the local captures and anomalies.

## 5. Create the release

Before tagging, confirm the code revision still equals `$revision` and no code changed after testing. If code changed, start again from step 2. Documentation-only handling after validation must still be reviewed explicitly rather than assumed safe. Candidate reports live under the ignored `validation/local/` directory so they do not change the tested Git tree.

```powershell
if ((git rev-parse HEAD).Trim() -ne $revision) { throw "HEAD changed after validation." }
git status --short
git tag -s "v$version" $revision -m "BedrockDebugProxy v$version"
git push origin "v$version"
```

Use a signed annotated tag when signing is configured. If it is not configured, stop and make an explicit maintainer decision before using an unsigned annotated tag.

Create a checksum for the exact tested Windows binary, then create the GitHub release from that exact tag. Include concise release notes, the supported Bedrock and protocol version, important limitations, automated check status, the six-server validation result, the tested binary and checksum, and the sanitized JSON reports as small release assets. Do not attach raw captures or captured third-party content.

```powershell
$binary = (Resolve-Path .\bin\bedrock-debug-proxy.exe).Path
$checksumFile = "$binary.sha256"
$hashLine = "{0}  {1}" -f `
    ((Get-FileHash -Algorithm SHA256 -LiteralPath $binary).Hash.ToLowerInvariant()), `
    (Split-Path -Leaf $binary)
[System.IO.File]::WriteAllText($checksumFile, $hashLine, [System.Text.Encoding]::ASCII)
$reportAssets = @(Get-ChildItem ".\validation\local\$revision" -File -Filter "*.json" | ForEach-Object FullName)
$releaseAssets = @($binary, $checksumFile) + $reportAssets
gh release create "v$version" `
    @releaseAssets `
    --repo NhanAZ/BedrockDebugProxy `
    --title "BedrockDebugProxy v$version" `
    --notes-file .\release-notes.md
```

The release notes file must be reviewed and must not contain credentials, server addresses that are meant to remain private, captured payloads, or claims that live validation proves universal compatibility.

## 6. Verify the published release

1. Confirm the GitHub release points to the intended tag and commit.
2. Download the published binary or build from the tag and run `version`.
3. Confirm checksums and attached validation reports are present.
4. Confirm the README quick start matches the released CLI.
5. Record any external outage, known limitation, or pending compatibility investigation in the release notes instead of silently omitting it.

A release is complete only after the published tag, binary metadata, reports, and release notes all identify the same tested revision.
