# Release workflow

This is the required path from a release candidate commit to an official BedrockDebugProxy release. `docs/validation.md` defines the evidence in detail. This file defines the order of operations.

## Maintainer and AI agent handoff

When the maintainer says "I want to release", the AI agent should drive this document rather than returning the whole checklist as homework.

1. The agent audits the commits since the latest release, proposes or confirms the version, checks the working tree, documentation, licenses, provenance, known limitations, CI, and unresolved review findings.
2. The agent runs the automated quality gate and builds the exact stamped candidate.
3. The agent runs the live six-server matrix with the authenticated Go `golem network` client, the approved cached account, and `--decrypt-resource-packs` on every release session. The client completes the offered packs, reaches lobby spawn, retains bidirectional decoded traffic for 30 seconds, then closes. It does not move, select lobby forms or items, or enter minigames. The PHP Golem CLI remains available for disposable PocketMine-MP or Axolotl-PM plugin scenarios. A retail Minecraft client is not required for release. CubeCraft is included in the matrix without a separate repeated confirmation step; this does not bypass technical gates or allow packet mutation, replay, or anti-cheat evasion.
4. The agent inspects each closed capture, verifies its exact revision, authentication, pack exchange, decryption of every encrypted offered pack, spawn, bidirectional decoded traffic, errors, and artifacts, then generates the sanitized report. The maintainer does not need to run report scripts.
5. After all required sessions, the agent runs the release gate, prepares the checksum and release notes, and reports any exact blocker. A candidate that adds only allowed Markdown documentation after the matrix may use the documented docs-only equivalence with the original validated runtime revision; the reports themselves must not be edited.
6. If the maintainer explicitly requested a release, the version is settled, every gate passes, and signing or GitHub access is available, the agent creates and pushes the tag, publishes the release assets, and verifies the result. It must not publish from an incomplete or mismatched validation set.
7. After the published release is verified, the agent reads the private maintainer handoff README at `https://github.com/NhanAZ-Assets/.github-private/blob/main/profile/README.md`, follows its applicable instructions without copying private content into this repository, synchronizes the recovery mirror, and verifies the refs.
8. Only after the release, private handoff, and recovery synchronization succeed, the agent removes local captures and derived decrypted resource-pack artifacts. Sanitized reports and release assets remain available; raw captures and decrypted third-party content do not.

The remaining sections retain the complete manual procedure so another developer can reproduce and audit what the agent performs.

## 1. Choose the candidate

1. Decide the semantic version, such as `0.1.0`, and set it for the commands below.
2. Confirm the intended release commit is on `main`, all required pull requests are merged, and the working tree is clean.
3. Review user-visible changes, protocol support, known limitations, license notices, and documentation. Do not mix a last-minute refactor or protocol update into release preparation.
4. Review [`CHANGELOG.md`](../CHANGELOG.md). Confirm that `Unreleased` covers every user-visible change since the latest release, then move those entries into a dated version section for the candidate. Keep every version in this one root file. Do not create a separate changelog file for each release. Leave a new empty `Unreleased` section for subsequent work. Audit records provide detailed evidence but are not a substitute for this summary.

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

If the candidate can affect the local web viewer, complete the automated and browser checks in [`web-validation.md`](web-validation.md) from this clean revision before starting the six-server matrix. A release candidate with an untested page, missing embedded asset, JavaScript error, or unexplained browser-console error is not ready for publication.

## 3. Validate the six-server baseline

Use the stamped Go network client for every release session. It authenticates directly to the selected Experience, completes the offered packs, reaches lobby spawn, runs a stationary 30-second `stability` scenario, and then executes `quit`. The release gate does not require a retail Minecraft client and makes no claim about retail rendering, UI, input, or gameplay.

The first successful device login creates the selected default account cache, so later server runs normally do not require another device code. If the machine has multiple profiles, select the intended account with `bedrock-debug-proxy use NAME` before starting the stamped binary. The Go network client uses that approved cached Microsoft account and does not start a local proxy listener.

Transfer following is enabled by default in `golem network` so server-directed hops needed to finish joining the selected Experience are captured in the same session. For CubeCraft, do not select any minigame. For Enchanted, leave its automatically sent Hub form unanswered and close after the post-spawn stability window.

| Report label | Upstream target | Required live flow |
| --- | --- | --- |
| `The Hive` | `experience:The Hive` | Connect, complete packs, reach lobby spawn, observe 30 seconds of bidirectional decoded traffic, then close |
| `Galaxite` | `experience:Galaxite` | Connect, complete packs, reach lobby spawn, observe 30 seconds of bidirectional decoded traffic, then close |
| `Lifeboat` | `experience:Lifeboat` | Connect, complete packs, reach lobby spawn, observe 30 seconds of bidirectional decoded traffic, then close |
| `Mineville Zeqa` | `experience:Mineville Zeqa` | Connect, complete packs, reach the joined lobby spawn, observe 30 seconds of bidirectional decoded traffic, then close |
| `Enchanted` | `experience:Enchanted [Skyblock]` | Complete Hub packs, reach Hub spawn, capture the automatic form, observe 30 seconds without selecting a child, then close |
| `CubeCraft` | `experience:CubeCraft` | Complete lobby packs, reach spawn, observe 30 seconds without entering a minigame, then close |

Create a two-line scenario file containing `{"op":"stability","duration":"30s"}` followed by `{"op":"quit"}`. Set the per-request `--timeout` above 30 seconds. Use the selector shown in the matrix and replace only the value after `--server` for each run. The `experience:` selector resolves both Featured and Creator Experiences through Minecraft services. CubeCraft follows the same owner-controlled, read-only path without a separate repeated risk confirmation. Its public policy boundary still applies. Do not add bypass, mutation, replay, or anti-cheat evasion behavior to the release command.

```powershell
.\bin\bedrock-debug-proxy.exe `
    golem network `
    --server "experience:The Hive" `
    --auth device `
    --decrypt-resource-packs `
    --capture .local\validation\the-hive `
    --validation-report validation\local\<REVISION>\the-hive.json `
    --server-label "The Hive" `
    --validation-result pass `
    --timeout 90s `
    --input .local\validation\quit.jsonl
```

Create `.local\validation\quit.jsonl` with these two lines:

```jsonl
{"op":"stability","duration":"30s"}
{"op":"quit"}
```

`golem network` reaches spawn before it reads the scenario input. The stability operation keeps the client stationary while it observes the post-spawn session. It does not select forms or items or enter minigames.

Before or alongside the public matrix, the agent may use the PHP Golem CLI against a disposable local development server to exercise plugin commands, forms, inventories, movement, events, and other server-side behavior. Keep that simulation journal separate from Go network-client captures. The authenticated Go `golem network` sessions are the public-server protocol reports used by the release matrix.

For every server, confirm all of the following before recording `pass`.

1. Device authentication and Experience resolution succeed when applicable.
2. The Golem client completes the join flow, reaches spawn, and retains bidirectional decoded traffic for at least 30 seconds before closing.
3. No movement, lobby item or form selection, or minigame entry occurs.
4. Client-to-server and server-to-client traffic are present.
5. Decoded packets are present.
6. Offered resource packs complete without an unexplained error.
7. No unexplained decode, protocol, authentication, signaling, or connection error appears.
8. Client closure lets the proxy finish its capture and artifacts, and the verifier succeeds.

If a server cannot be tested, create an `incomplete` report. If an anomaly occurs, create a `fail` report and investigate. Neither result passes the release gate. When a release command encounters an error, the agent owns the diagnosis and fix-forward loop: inspect the earliest divergent event, patch the smallest justified cause, run focused tests, rebuild a new stamped binary, and rerun every affected report. Ask the maintainer only when the next action requires unavailable authority, an ambiguous product decision, or a human-only interaction.

For every offered encrypted pack, the release capture must contain a successful derived decrypted archive and no decryption error. Add `resource_pack_decryption=pass` to the report checks whenever the server offers encrypted packs. A missing content key, unsupported encryption variant, or partial decryption is a release failure, not a reason to accept the encrypted archive as sufficient evidence. The original encrypted archive remains local evidence until cleanup.

For Enchanted, the automatically sent Hub form is captured, but no entry is selected during release validation. The report does not require a form inventory or child-route matrix. Explore child routes separately when collecting the server packs.

The CubeCraft report uses the normal read-only path. Do not ask for a separate repeated rejection, kick, or ban confirmation and do not require `risk_acknowledged` in the report. The readiness command still requires an authenticated session, successful pack decryption, spawn, bidirectional decoded traffic, and verified artifacts. The lack of a separate confirmation is not permission or a policy exception.

## 4. Generate one report per server

Use the exact capture directory printed by the matching run. The following example is for The Hive.

```powershell
.\tools\create-validation-report.ps1 `
    -Binary .\bin\bedrock-debug-proxy.exe `
    -Capture .\captures\<THE-HIVE-SESSION> `
    -Server "The Hive" `
    -Result pass `
    -Check @("cli_session=pass", "authenticated_upstream=pass", "experience_resolution=pass", "resource_pack_decryption=pass") `
    -Output ".\validation\local\$revision\the-hive.json"
```

Use these exact `-Server` labels for the other reports: `Galaxite`, `Lifeboat`, `Mineville Zeqa`, `Enchanted`, and `CubeCraft`. Add `resource_pack_decryption=pass` when encrypted packs were offered. Add a focused manual check when a release specifically changes a flow. Never place raw captures, tokens, addresses, resource-pack keys, private discussions, or third-party assets in a report.

Review each JSON file before using it. Then run the matrix gate.

```powershell
.\tools\check-release-readiness.ps1 -Revision $revision
```

The command must print `pass` for all six servers. This gate verifies report structure, client mode, authentication, spawn and capture evidence, encrypted-pack decryption, and revision matching. It does not replace agent review of local captures. Retail reports use `normal_session=pass`; the required path uses `cli_session=pass` and `authenticated_upstream=pass`.

## 5. Create the release

Before tagging, confirm the candidate revision and validation mode. If product code, tools other than the release checker, workflows, configuration, generated files, or another non-approved path changed after testing, start again from step 2. For a docs-only delta, set `$runtimeRevision` to the exact revision whose binary produced the reports and run the equivalence gate. Review the allowlist output explicitly rather than assuming that a documentation commit is safe. Candidate reports live under the ignored `validation/local/` directory so they do not change the tested Git tree.

Create `release-notes.md` from the reviewed version section in `CHANGELOG.md`. The release body must link back to the version section in the tagged `CHANGELOG.md` and to GitHub's compare view so users can inspect every commit included in the release. Use this shape and replace the placeholders with the actual version and GitHub heading anchor:

```markdown
# BedrockDebugProxy v<version>

Please see the [changelog](https://github.com/NhanAZ/BedrockDebugProxy/blob/v<version>/CHANGELOG.md#<version-anchor>) for details.

[View all commits in this release](https://github.com/NhanAZ/BedrockDebugProxy/compare/v<previous-version>...v<version>)
```

Add the concise release summary, validation, scope, and asset notes below those links. When docs-only equivalence is used, identify both the candidate revision and the validated runtime revision. Keep the notes concise and exclude credentials, raw captures, private addresses, resource-pack keys, decrypted assets, and unsupported compatibility claims.

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

Create a checksum and third-party license bundle for the candidate Windows binary, then create the GitHub release from that exact tag. In docs-only mode, the sanitized reports come from `$validationRevision` and the notes must state that live evidence was inherited from `$runtimeRevision`; do not describe the candidate binary as having been run in those sessions. Include concise release notes, the supported Bedrock and protocol version, important limitations, automated check status, the six-server validation result, the candidate binary and checksum, the project license and notices, the generated dependency license bundle, and the sanitized JSON reports as small release assets. Do not attach raw captures or captured third-party content.

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

A release is not complete until the published tag, binary metadata, reports, release notes, private handoff, cleanup, and verified recovery mirror identify the candidate revision and, when docs-only equivalence is used, the separately stated validated runtime revision.

## 7. Read the private maintainer handoff

After the GitHub release and its assets are verified, read the private maintainer instructions at [`NhanAZ-Assets/.github-private/profile/README.md`](https://github.com/NhanAZ-Assets/.github-private/blob/main/profile/README.md) using an authenticated GitHub client. Do not paste its contents into the public repository, release notes, reports, logs, or chat. Follow only the actions that apply to the released version and record a short public-safe completion note without private details.

If the private repository is unavailable, the agent must not guess the next action. Ask the maintainer to restore access or provide the applicable instruction. The release remains published but post-release completion is incomplete until the handoff has been read.

## 8. Clean local release artifacts

After the published release, private handoff, and recovery synchronization are verified, remove local raw captures and derived decrypted resource-pack artifacts. Keep sanitized reports under `validation/local/<REVISION>/`, the release assets, Git history, and public-safe audit records. Do not remove a capture before any required private archive or release asset has been independently verified.

Use explicit paths and inspect the resolved targets before deletion:

```powershell
$projectRoot = (Resolve-Path .).Path
$captureRoot = (Resolve-Path (Join-Path $projectRoot "captures")).Path
if ([IO.Path]::GetFullPath($captureRoot) -ne [IO.Path]::Combine($projectRoot, "captures")) {
    throw "Refusing to clean an unexpected capture path: $captureRoot"
}
Remove-Item -LiteralPath $captureRoot -Recurse -Force

$tempRoot = [IO.Path]::GetTempPath()
Get-ChildItem -LiteralPath $tempRoot -File -Filter "bedrock-debug-proxy-decrypted-pack-*.mcpack" |
    ForEach-Object { Remove-Item -LiteralPath $_.FullName -Force }
```

The second cleanup removes only stale temporary files created by the documented decryption path. If a process is still running, stop it and rerun cleanup. Never delete the entire temporary directory or the ignored `validation/local/` directory.

## 9. Synchronize the source-history backup

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
