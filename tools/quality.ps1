[CmdletBinding()]
param(
    [switch]$Race
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$golangCILintVersion = "v2.13.2"
$govulncheckVersion = "v1.1.4"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

function Invoke-QualityCommand {
    param(
        [Parameter(Mandatory)]
        [string]$Label,
        [Parameter(Mandatory)]
        [string]$Executable,
        [Parameter(Mandatory)]
        [string[]]$Arguments
    )

    Write-Host "==> $Label"
    & $Executable @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$Label failed with exit code $LASTEXITCODE."
    }
}

Push-Location $projectRoot
try {
    $goFiles = @(Get-ChildItem -LiteralPath $projectRoot -Recurse -File -Filter "*.go" |
        Where-Object { $_.FullName -notmatch "[\\/]\.git[\\/]" -and $_.FullName -notmatch "[\\/]vendor[\\/]" } |
        ForEach-Object { $_.FullName })
    if ($goFiles.Count -eq 0) {
        throw "No Go source files were found."
    }

    Write-Host "==> gofmt"
    $unformatted = [System.Collections.Generic.List[string]]::new()
    foreach ($file in $goFiles) {
        $formattedOutput = @(& gofmt -l $file)
        if ($LASTEXITCODE -ne 0) {
            throw "gofmt failed for $file with exit code $LASTEXITCODE."
        }
        foreach ($entry in $formattedOutput) {
            [void]$unformatted.Add([string]$entry)
        }
    }
    if ($unformatted.Count -ne 0) {
        $unformatted | ForEach-Object { Write-Host $_ }
        throw "Go source is not formatted. Run tools/format.ps1."
    }

    Write-Host "==> PowerShell syntax"
    $powerShellFiles = @(Get-ChildItem -LiteralPath $projectRoot -Recurse -File -Filter "*.ps1" |
        Where-Object { $_.FullName -notmatch "[\\/]\.git[\\/]" } |
        ForEach-Object { $_.FullName })
    $parseFailures = @()
    foreach ($file in $powerShellFiles) {
        $tokens = $null
        $parseErrors = $null
        [System.Management.Automation.Language.Parser]::ParseFile($file, [ref]$tokens, [ref]$parseErrors) | Out-Null
        foreach ($parseError in @($parseErrors)) {
            $parseFailures += "$file`:$($parseError.Extent.StartLineNumber) $($parseError.Message)"
        }
    }
    if ($parseFailures.Count -ne 0) {
        $parseFailures | ForEach-Object { Write-Host $_ }
        throw "PowerShell source contains syntax errors."
    }

    Invoke-QualityCommand "module tidiness" "go" @("mod", "tidy", "-diff")
    Invoke-QualityCommand "module integrity" "go" @("mod", "verify")
    Invoke-QualityCommand "third-party license inventory" "pwsh" @("-NoProfile", "-File", (Join-Path $PSScriptRoot "collect-third-party-licenses.ps1"))

    $testArguments = @("test", "-count=1")
    if ($Race) {
        $testArguments += "-race"
    }
    $testArguments += "./..."
    Invoke-QualityCommand "tests" "go" $testArguments

    Invoke-QualityCommand "build" "go" @("build", "./...")
    Invoke-QualityCommand "golangci-lint config" "go" @("run", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$golangCILintVersion", "config", "verify")
    Invoke-QualityCommand "static analysis" "go" @("run", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$golangCILintVersion", "run", "./...")
    Invoke-QualityCommand "vulnerability scan" "go" @("run", "golang.org/x/vuln/cmd/govulncheck@$govulncheckVersion", "./...")

    Write-Host "==> tracked capture artifacts"
    $trackedFiles = @(& git ls-files)
    if ($LASTEXITCODE -ne 0) {
        throw "git ls-files failed with exit code $LASTEXITCODE."
    }
    $blockedArtifacts = @($trackedFiles | Where-Object {
            $_ -match '(^|/)(captures|exports|resource-packs|decrypted-resource-packs)/' -or
            $_ -match '\.(bdpcap|pcap|pcapng|mcpack|mcaddon|mctemplate|mcworld)$'
        })
    if ($blockedArtifacts.Count -ne 0) {
        $blockedArtifacts | ForEach-Object { Write-Host $_ }
        throw "Tracked capture or resource-pack artifacts are forbidden. Use synthetic source-generated fixtures."
    }

    Invoke-QualityCommand "whitespace errors" "git" @("diff", "--check", "HEAD")
} finally {
    Pop-Location
}

Write-Host "Quality checks passed."
