[CmdletBinding()]
param(
    [string]$Revision = "",
    [string]$Reports = "validation"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$requiredServers = @(
    "The Hive"
    "CubeCraft"
    "Galaxite"
    "Lifeboat"
    "Mineville Zeqa"
    "Enchanted"
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
        if ([string]$report.tested_revision -ne $Revision) {
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
                $_.Result -eq "pass" -and $_.Automatic -and $_.Manual
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
        throw "Release validation is incomplete for revision $Revision. $([string]::Join(', ', $failures))"
    }
    Write-Host "Release validation is complete for revision $Revision."
} finally {
    Pop-Location
}
