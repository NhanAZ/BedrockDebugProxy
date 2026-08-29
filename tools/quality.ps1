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
    $unformatted = @(& gofmt -l @goFiles)
    if ($LASTEXITCODE -ne 0) {
        throw "gofmt failed with exit code $LASTEXITCODE."
    }
    if ($unformatted.Count -ne 0) {
        $unformatted | ForEach-Object { Write-Host $_ }
        throw "Go source is not formatted. Run tools/format.ps1."
    }

    Invoke-QualityCommand "module tidiness" "go" @("mod", "tidy", "-diff")
    Invoke-QualityCommand "module integrity" "go" @("mod", "verify")

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
    Invoke-QualityCommand "whitespace errors" "git" @("diff", "--check", "HEAD")
} finally {
    Pop-Location
}

Write-Host "Quality checks passed."
