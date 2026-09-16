# Release workflow

This is the required path from a release candidate commit to an official BedrockDebugProxy release. `docs/validation.md` defines the evidence in detail. This file defines the order of operations.

## Maintainer and AI agent handoff

When the maintainer says "I want to release", the AI agent should drive this document rather than returning the whole checklist as homework.

1. The agent audits the commits since the latest release, proposes or confirms the version, checks the working tree, documentation, licenses, provenance, known limitations, CI, and unresolved review findings.
2. The agent runs the automated quality gate and builds the exact stamped candidate.
3. The agent asks the maintainer to perform the live Minecraft sessions that cannot be automated. A short instruction should identify the server, target command, actions to perform, approximate duration, and when to press `Ctrl+C`.
4. The maintainer reports that the session is complete and mentions any visible problem. The agent locates and inspects the closed capture, verifies the revision and evidence, and generates the sanitized report. The maintainer does not need to run report scripts unless they are performing the release without an agent.
5. After all required sessions, the agent runs the release gate, prepares the checksum and release notes, and reports any exact blocker. A candidate that adds only allowed Markdown documentation after the matrix may use the documented docs-only equivalence with the original validated runtime revision; the reports themselves must not be edited.
6. If the maintainer explicitly requested a release, the version is settled, every gate passes, and signing or GitHub access is available, the agent creates and pushes the tag, publishes the release assets, and verifies the result. It must not publish from an incomplete or mismatched validation set.
7. After the published release is verified, the agent synchronizes the Git-tracked source history, branches, and tags with the private recovery mirror described in section 7. The release is not reported as fully complete until this synchronization and ref check succeed.

The remaining sections retain the complete manual procedure so another developer can reproduce and audit what the agent performs.

## 1. Choose the candidate

1. Decide the semantic version, such as `0.1.0`, and set it for the commands below.
2. Confirm the intended release commit is on `main`, all required pull requests are merged, and the working tree is clean.
3. Review user-visible changes, protocol support, known limitations, license notices, and documentation. Do not mix a last-minute refactor or protocol update into release preparation.
4. Review [`CHANGELOG.md`](../CHANGELOG.md). Confirm that `Unreleased` covers every user-visible change since the latest release, then move those entries into a dated version section for the candidate. Leave a new empty `Unreleased` section for subsequent work. Audit records provide detailed evidence but are not a substitute for this summary.

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

The printed binary commit must equal `$revision`. Keep this exact binary for all five live tests. Do not rebuild from a different commit between servers.

## 3. Validate the five-server baseline

Use a real Minecraft Bedrock client. For each server, start the same stamped binary, complete a normal session, exercise the named flow, stop with `Ctrl+C`, and wait for the capture verification message.

The first successful device login creates a per-user token cache, so later server runs normally do not require another device code. Wait for `Listening on` before connecting Minecraft. Use the default Xbox-authenticated downstream path for the required release reports, normally through the Servers tab. Do not add `--allow-unauthenticated-client` to the release command unless the release is separately validating the trusted-LAN opt-in flow.

Include `--follow-transfers` in the release command when the server's flow can send a `Transfer` packet. This keeps subsequent hops in the same capture. Use a concrete reachable address in `--listen` for that mode; a wildcard listener address cannot be sent back to the client.

| Report label | Upstream target | Required manual flow |
| --- | --- | --- |
| `The Hive` | `experience:The Hive` | Connect, spawn, move, interact, and observe normal traffic |
| `Galaxite` | `experience:Galaxite` | Connect, spawn, move, interact, and observe normal traffic |
| `Lifeboat` | `experience:Lifeboat` | Connect, spawn, move, interact, and observe normal traffic |
| `Mineville Zeqa` | `experience:Mineville Zeqa` | Enter the Zeqa flow and observe it directly |
| `Enchanted` | `experience:Enchanted` | Receive the server-sent form, select any minigame, observe `Transfer`, follow the next hop, complete its resource-pack exchange when offered, and reach minigame spawn |

Use this command for an Experience target. Replace `192.168.1.10` with the concrete LAN address used by the Minecraft client, and replace only the value after `--upstream` for each run. Do not add CubeCraft to this matrix because its published policy prohibits Bedrock proxy use. See [`server-targets.md`](server-targets.md) for the policy disclaimer and endpoint references.

```powershell
.\bin\bedrock-debug-proxy.exe `
    run `
    --listen 192.168.1.10:19132 `
    --upstream "experience:The Hive" `
    --auth device `
    --follow-transfers
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

The Enchanted report must include `transfer_following=pass` and `resource_pack_transfer=pass`. The server sends the form automatically when the player joins. A hub-only session, or a session where no minigame is selected from that form, does not satisfy the Enchanted release gate.

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

Use these exact `-Server` labels for the other reports: `Galaxite`, `Lifeboat`, `Mineville Zeqa`, and `Enchanted`. Add a focused manual check when a release specifically changes a flow. Never place raw captures, tokens, addresses, resource-pack keys, or third-party assets in a report.

Review each JSON file before using it. Then run the matrix gate.

```powershell
.\tools\check-release-readiness.ps1 -Revision $revision
```

The command must print `pass` for all five servers. This gate verifies report structure and revision matching. It does not replace human review of the local captures and anomalies.

## 5. Create the release

Before tagging, confirm the candidate revision and validation mode. If product code, tools other than the release checker, workflows, configuration, generated files, or another non-approved path changed after testing, start again from step 2. For a docs-only delta, set `$runtimeRevision` to the exact revision whose binary produced the reports and run the equivalence gate. Review the allowlist output explicitly rather than assuming that a documentation commit is safe. Candidate reports live under the ignored `validation/local/` directory so they do not change the tested Git tree.

Create `release-notes.md` from the reviewed version section in `CHANGELOG.md`. When docs-only equivalence is used, identify both the candidate revision and the validated runtime revision. Keep the notes concise and exclude credentials, raw captures, private addresses, resource-pack keys, decrypted assets, and unsupported compatibility claims.

```powershell
$revision = (git rev-parse HEAD).Trim()
$runtimeRevision = "<validated-runtime-40-character-commit>"
if ($revision -eq $runtimeRevision) {
    .\tools\check-release-readiness.ps1 -Revision $revision
} else {
    .\tools\check-release-readiness.ps1 -Revision $revision -ValidatedRevision $runtimeRevision
}
git status --short
git tag -s "v$version" $revision -m "BedrockDebugProxy v$version"
git push origin "v$version"
```

Use a signed annotated tag when signing is configured. If it is not configured, stop and make an explicit maintainer decision before using an unsigned annotated tag.

Create a checksum and third-party license bundle for the candidate Windows binary, then create the GitHub release from that exact tag. In docs-only mode, the sanitized reports come from `$validationRevision` and the notes must state that live evidence was inherited from `$runtimeRevision`; do not describe the candidate binary as having been run in those sessions. Include concise release notes, the supported Bedrock and protocol version, important limitations, automated check status, the five-server validation result, the candidate binary and checksum, the project license and notices, the generated dependency license bundle, and the sanitized JSON reports as small release assets. Do not attach raw captures or captured third-party content.

```powershell
$binary = (Resolve-Path .\bin\bedrock-debug-proxy.exe).Path
$checksumFile = "$binary.sha256"
$hashLine = "{0}  {1}" -f `
    ((Get-FileHash -Algorithm SHA256 -LiteralPath $binary).Hash.ToLowerInvariant()), `
    (Split-Path -Leaf $binary)
[System.IO.File]::WriteAllText($checksumFile, $hashLine, [System.Text.Encoding]::ASCII)
$validationRevision = if ($revision -eq $runtimeRevision) { $revision } else { $runtimeRevision }
$licenseBundle = ".\validation\local\$validationRevision\THIRD_PARTY_LICENSES.txt"
.\tools\collect-third-party-licenses.ps1 -Output $licenseBundle
$reportAssets = @(Get-ChildItem ".\validation\local\$validationRevision" -File -Filter "*.json" | ForEach-Object FullName)
$releaseAssets = @(
    $binary
    $checksumFile
    (Resolve-Path .\LICENSE).Path
    (Resolve-Path .\THIRD_PARTY_NOTICES.md).Path
    (Resolve-Path $licenseBundle).Path
) + $reportAssets
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

A release is complete only after the published tag, binary metadata, reports, release notes, and verified recovery mirror identify the candidate revision and, when docs-only equivalence is used, the separately stated validated runtime revision.

## 7. Synchronize the source-history backup

The private repository [`NhanAZ/BedrockDebugProxy-Backup`](https://github.com/NhanAZ/BedrockDebugProxy-Backup) is a recovery mirror for Git-tracked source and history. After the release has been verified, push all local branches and tags to this remote and verify that its default branch points to the candidate revision.

Do not mirror ignored captures, authentication state, resource-pack keys, private keys, build output, or other local artifacts. GitHub Release assets are not copied by Git pushes and remain attached to the release itself.

```powershell
$backupUrl = "https://github.com/NhanAZ/BedrockDebugProxy-Backup.git"
$configuredBackupUrl = git remote get-url backup 2>$null
if ($LASTEXITCODE -ne 0) {
    git remote add backup $backupUrl
} elseif (($configuredBackupUrl | Select-Object -First 1).TrimEnd('/') -ne $backupUrl.TrimEnd('/')) {
    throw "The backup remote does not point to $backupUrl"
}

git push backup --all
git push backup --tags

$backupRevision = (git ls-remote backup "refs/heads/main" | ForEach-Object { ($_ -split "\s+")[0] } | Select-Object -First 1)
if ($backupRevision -ne $revision) {
    throw "Backup main is $backupRevision, expected tested revision $revision"
}

$backupTag = git ls-remote --tags backup "refs/tags/v$version"
if (-not $backupTag) {
    throw "Backup tag v$version was not found"
}
```

If the backup remote is unavailable or the ref check fails, retain the published release but report the backup step as incomplete. Do not claim that the recovery mirror is current until the push and verification succeed. Keep the backup repository private and restrict its collaborators and tokens.
