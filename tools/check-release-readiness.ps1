[CmdletBinding()]
param(
    [string]$Revision = "",
    [string]$Reports = "validation",
    [string]$ValidatedRevision = ""
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$requiredServers = @(
    "The Hive"
    "Galaxite"
    "Lifeboat"
    "Mineville Zeqa"
    "Enchanted"
    "CubeCraft"
)

Push-Location $projectRoot
try {
    if ([string]::IsNullOrWhiteSpace($Revision)) {
        $Revision = (& git rev-parse HEAD).Trim()
        if ($LASTEXITCODE -ne 0) {
            throw "Could not resolve the current Git revision."
        }
    }
    if ($Revision -notmatch '^[0-9a-f]{40}$') {
        throw "Revision must be an exact lowercase 40-character Git commit."
    }

    $reportRevision = $Revision
    $docsOnlyMode = -not [string]::IsNullOrWhiteSpace($ValidatedRevision)
    if ($docsOnlyMode) {
        if ($ValidatedRevision -notmatch '^[0-9a-f]{40}$') {
            throw "ValidatedRevision must be an exact lowercase 40-character Git commit."
        }
        & git cat-file -e "$ValidatedRevision^{commit}"
        if ($LASTEXITCODE -ne 0) {
            throw "ValidatedRevision does not identify an existing Git commit."
        }
        & git cat-file -e "$Revision^{commit}"
        if ($LASTEXITCODE -ne 0) {
            throw "Revision does not identify an existing Git commit."
        }
        & git merge-base --is-ancestor $ValidatedRevision $Revision
        if ($LASTEXITCODE -ne 0) {
            throw "ValidatedRevision must be an ancestor of Revision."
        }

        $changedPaths = @(& git diff --name-only --diff-filter=ACDMRTUXB "$ValidatedRevision..$Revision")
        if ($LASTEXITCODE -ne 0) {
            throw "Could not inspect the candidate revision diff."
        }
        $nonDocumentationPaths = @($changedPaths | Where-Object {
                $_ -and $_ -notmatch '^(AGENTS|CHANGELOG|CONTRIBUTING)\.md$' -and $_ -notmatch '^docs/.+\.md$' -and $_ -ne 'tools/check-release-readiness.ps1'
            })
        if ($nonDocumentationPaths.Count -ne 0) {
            throw "ValidatedRevision can only be reused when every candidate change is approved documentation or the release checker. Disallowed paths: $([string]::Join(', ', $nonDocumentationPaths))"
        }
        $reportRevision = $ValidatedRevision
        Write-Host "Docs-only validation equivalence: candidate $Revision reuses reports from runtime revision $ValidatedRevision."
    }

    $reportsPath = if ([System.IO.Path]::IsPathRooted($Reports)) {
        [System.IO.Path]::GetFullPath($Reports)
    } else {
        [System.IO.Path]::GetFullPath((Join-Path $projectRoot $Reports))
    }
    if (-not (Test-Path -LiteralPath $reportsPath -PathType Container)) {
        throw "Validation report directory does not exist at $reportsPath."
    }

    $reportsByServer = @{}
    foreach ($file in @(Get-ChildItem -LiteralPath $reportsPath -Recurse -File -Filter "*.json")) {
        try {
            $report = Get-Content -Raw -LiteralPath $file.FullName | ConvertFrom-Json
        } catch {
            throw "Validation report $($file.FullName) is not valid JSON. $($_.Exception.Message)"
        }
        if ([string]$report.schema -ne "bedrockdebugproxy.validation.v1") {
            continue
        }
        if ([string]$report.tested_revision -ne $reportRevision) {
            continue
        }
        $server = [string]$report.target_server
        if (-not $reportsByServer.ContainsKey($server)) {
            $reportsByServer[$server] = @()
        }
        $reportsByServer[$server] += [pscustomobject]@{
            Path = $file.FullName
            Result = [string]$report.result
            Automatic = [bool]$report.evidence.automatic_checks_passed
            Manual = [bool]$report.manual_checks_passed
            ManualChecks = @($report.manual_checks)
            ResourcePackCount = [int]$report.evidence.resource_pack_count
            EncryptedResourcePackCount = if ($null -ne $report.evidence.PSObject.Properties["encrypted_resource_pack_count"]) { [int]$report.evidence.encrypted_resource_pack_count } else { -1 }
            DecryptedResourcePackCount = [int]$report.evidence.decrypted_resource_pack_count
            ResourcePackDecryptErrorCount = [int]$report.evidence.resource_pack_decrypt_error_count
        }
    }

    $failures = @()
    $rows = @()
    foreach ($server in $requiredServers) {
        $candidates = @()
        if ($reportsByServer.ContainsKey($server)) {
            $candidates = @($reportsByServer[$server])
        }
        $passing = @($candidates | Where-Object {
                $checks = @{}
                foreach ($check in $_.ManualChecks) {
                    $checks[[string]$check.name] = [string]$check.result
                }
                $retailSession = $checks["normal_session"] -eq "pass"
                $cliSession = $checks["cli_session"] -eq "pass" -and $checks["authenticated_upstream"] -eq "pass" -and
                    $checks["post_spawn_stability"] -eq "pass"
                $sessionGate = ($retailSession -xor $cliSession) -and -not ($checks.Values -contains "fail") -and
                    -not ($checks.Values -contains "not_observed")
                $conditionalTransferStates = @("pass", "not_applicable")
                $enchantedFlowGate = $server -ne "Enchanted" -or
                    ($checks["enchanted_form_review"] -eq "pass" -and
                        $checks["enchanted_route_matrix"] -eq "pass" -and
                        $conditionalTransferStates -contains $checks["transfer_following"] -and
                        $conditionalTransferStates -contains $checks["resource_pack_transfer"])
                $decryptionGate = $_.EncryptedResourcePackCount -ge 0 -and
                    ($_.EncryptedResourcePackCount -eq 0 -or
                        ($_.DecryptedResourcePackCount -eq $_.EncryptedResourcePackCount -and
                            $_.ResourcePackDecryptErrorCount -eq 0 -and
                            $checks["resource_pack_decryption"] -eq "pass"))
                $_.Result -eq "pass" -and $_.Automatic -and $_.Manual -and $sessionGate -and $enchantedFlowGate -and $decryptionGate
            })
        if ($passing.Count -eq 0) {
            $state = if ($candidates.Count -eq 0) { "missing" } else { "not passing" }
            $failures += "$server - $state"
            $rows += [pscustomobject]@{ Server = $server; State = $state; Report = "" }
            continue
        }
        $rows += [pscustomobject]@{ Server = $server; State = "pass"; Report = $passing[-1].Path }
    }

    $rows | Format-Table -AutoSize
    if ($failures.Count -ne 0) {
        throw "Release validation is incomplete for runtime revision $reportRevision (candidate $Revision). $([string]::Join(', ', $failures))"
    }
    if ($docsOnlyMode) {
        Write-Host "Release validation is complete for candidate $Revision using runtime revision $reportRevision."
    } else {
        Write-Host "Release validation is complete for revision $Revision."
    }
} finally {
    Pop-Location
}
