[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$goFiles = @(Get-ChildItem -LiteralPath $projectRoot -Recurse -File -Filter "*.go" |
    Where-Object { $_.FullName -notmatch "[\\/]\.git[\\/]" -and $_.FullName -notmatch "[\\/]vendor[\\/]" } |
    ForEach-Object { $_.FullName })

if ($goFiles.Count -eq 0) {
    throw "No Go source files were found."
}

& gofmt -w @goFiles
if ($LASTEXITCODE -ne 0) {
    throw "gofmt failed with exit code $LASTEXITCODE."
}

Write-Host "Formatted $($goFiles.Count) Go source files."
