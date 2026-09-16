[CmdletBinding()]
param(
    [string]$Repository = "",
    [string]$SourceRepository = "Mojang/bedrock-protocol-docs",
    [switch]$CreateIssue
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$baselinePath = Join-Path $projectRoot "docs/protocol-updates.md"
$apiHeaders = @{
    Accept = "application/vnd.github+json"
    "X-GitHub-Api-Version" = "2022-11-28"
    "User-Agent" = "BedrockDebugProxy-protocol-release-check"
}
if (-not [string]::IsNullOrWhiteSpace($env:GITHUB_TOKEN)) {
    $apiHeaders.Authorization = "Bearer $($env:GITHUB_TOKEN)"
}

function Invoke-GitHubJson {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Uri,
        [ValidateSet("Get", "Post")]
        [string]$Method = "Get",
        [object]$Body = $null
    )

    $request = @{
        Uri = $Uri
        Method = $Method
        Headers = $apiHeaders
    }
    if ($null -ne $Body) {
        $request.ContentType = "application/json"
        $request.Body = $Body | ConvertTo-Json -Depth 10
    }
    try {
        return Invoke-RestMethod @request
    } catch {
        throw "GitHub API request failed for $Uri. $($_.Exception.Message)"
    }
}

function Resolve-ReleaseCommit {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Tag
    )

    $encodedTag = [Uri]::EscapeDataString($Tag)
    try {
        $ref = Invoke-GitHubJson "https://api.github.com/repos/$SourceRepository/git/ref/tags/$encodedTag"
        $objectType = [string]$ref.object.type
        $objectSha = [string]$ref.object.sha
        if ($objectType -eq "commit") {
            return $objectSha
        }
        if ($objectType -eq "tag") {
            $tagObject = Invoke-GitHubJson "https://api.github.com/repos/$SourceRepository/git/tags/$objectSha"
            if ([string]$tagObject.object.type -eq "commit") {
                return [string]$tagObject.object.sha
            }
        }
        Write-Warning "The release tag $Tag did not resolve to a commit object. Resolve it before relying on protocol details."
        return "unknown"
    } catch {
        Write-Warning "Could not resolve immutable commit for release tag $Tag. $($_.Exception.Message)"
        return "unknown"
    }
}

if (-not (Test-Path -LiteralPath $baselinePath -PathType Leaf)) {
    throw "Protocol baseline file does not exist at $baselinePath."
}

$baselineText = Get-Content -Raw -LiteralPath $baselinePath
$versionMatch = [regex]::Match($baselineText, '(?m)^\| Bedrock game version \| `([^`]+)` \|')
$protocolMatch = [regex]::Match($baselineText, '(?m)^\| Bedrock protocol \| `([^`]+)` \|')
if (-not $versionMatch.Success -or -not $protocolMatch.Success) {
    throw "Could not read the Bedrock game version and protocol baseline from $baselinePath."
}
$baselineVersion = $versionMatch.Groups[1].Value
$baselineProtocol = $protocolMatch.Groups[1].Value

$release = Invoke-GitHubJson "https://api.github.com/repos/$SourceRepository/releases/latest"
$latestTag = [string]$release.tag_name
if ([string]::IsNullOrWhiteSpace($latestTag)) {
    throw "The source repository latest release did not contain a tag name."
}
$latestVersion = $latestTag
if ($latestVersion.StartsWith("v", [System.StringComparison]::OrdinalIgnoreCase)) {
    $latestVersion = $latestVersion.Substring(1)
}
$releaseBody = [string]$release.body
if ([string]::IsNullOrWhiteSpace($releaseBody)) {
    $latestProtocol = "unknown"
} else {
    $releaseProtocolMatch = [regex]::Match($releaseBody, '(?im)^\s*(?:-\s*)?Network protocol version:\s*([0-9]+)\s*$')
    $latestProtocol = if ($releaseProtocolMatch.Success) { $releaseProtocolMatch.Groups[1].Value } else { "unknown" }
}
$releaseUrl = [string]$release.html_url
if ([string]::IsNullOrWhiteSpace($releaseUrl)) {
    $releaseUrl = "https://github.com/$SourceRepository/releases/tag/$latestTag"
}
$sourceCommit = Resolve-ReleaseCommit $latestTag

$versionChanged = $latestVersion -ne $baselineVersion
$protocolChanged = $latestProtocol -ne "unknown" -and $latestProtocol -ne $baselineProtocol
Write-Host "Mojang release: $latestTag (version $latestVersion, protocol $latestProtocol)"
Write-Host "Project baseline: version $baselineVersion, protocol $baselineProtocol"
if (-not $versionChanged -and -not $protocolChanged) {
    Write-Host "No protocol release newer than the project baseline was detected."
    exit 0
}

$issueTitle = "[Protocol update] Mojang $latestTag (protocol $latestProtocol)"
$issueBody = @"
## Official Bedrock protocol release detected

The scheduled protocol watcher found a Mojang Bedrock protocol release that differs from the project baseline.

| Field | Current baseline | Detected release |
| --- | --- | --- |
| Game version | `$baselineVersion` | `$latestVersion` |
| Network protocol | `$baselineProtocol` | `$latestProtocol` |
| Release tag | n/a | `$latestTag` |
| Release commit | n/a | `$sourceCommit` |
| Published | n/a | `$([string]$release.published_at)` |

Source: [$latestTag]($releaseUrl)

Process guide: [protocol update workflow](https://github.com/$Repository/blob/main/docs/protocol-updates.md)

## Required next step

This issue is a review trigger only. It is not an instruction to copy schemas or update code automatically. Follow the protocol update workflow:

1. Verify the release schemas, tag commit, protocol number, and license terms.
2. Record an evidence table under `docs/research/` and cross-check independent implementations or captures.
3. Update packet definitions and focused tests only after the wire claims are confirmed.
4. Run the quality gate, build the exact revision, and complete the required live validation before release.

No product code or dependency was changed by this workflow.
"@

if (-not $CreateIssue) {
    Write-Host "Protocol update detected, but issue creation was not requested."
    Write-Host "Would open: $issueTitle"
    exit 0
}

if ([string]::IsNullOrWhiteSpace($Repository)) {
    $Repository = [string]$env:GITHUB_REPOSITORY
}
if ([string]::IsNullOrWhiteSpace($Repository)) {
    throw "Repository is required when creating an issue."
}
if ([string]::IsNullOrWhiteSpace($env:GITHUB_TOKEN)) {
    throw "GITHUB_TOKEN is required when creating an issue."
}

$openIssues = @()
for ($page = 1; $page -le 10; $page++) {
    $batch = @(Invoke-GitHubJson "https://api.github.com/repos/$Repository/issues?state=open&per_page=100&page=$page")
    if ($batch.Count -eq 0) {
        break
    }
    $openIssues += $batch
    if ($batch.Count -lt 100) {
        break
    }
}
$existing = @($openIssues | Where-Object {
        $isPullRequest = $null -ne $_.PSObject.Properties["pull_request"]
        -not $isPullRequest -and [string]$_.title -eq $issueTitle
    })
if ($existing.Count -ne 0) {
    Write-Host "An issue for $latestTag already exists: #$(($existing[0]).number) $($existing[0].html_url)"
    exit 0
}

$created = Invoke-GitHubJson "https://api.github.com/repos/$Repository/issues" -Method Post -Body @{
    title = $issueTitle
    body = $issueBody
}
Write-Host "Opened protocol update issue #$(($created).number): $($created.html_url)"
