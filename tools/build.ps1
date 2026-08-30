[CmdletBinding()]
param(
    [string]$Version = "dev",
    [string]$Output = ""
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location $projectRoot
try {
    if ($Version -notmatch '^[0-9A-Za-z._+-]+$') {
        throw "Version may contain only letters, numbers, dots, underscores, plus signs, and hyphens."
    }
    $changes = @(& git status --porcelain --untracked-files=normal)
    if ($LASTEXITCODE -ne 0) {
        throw "git status failed with exit code $LASTEXITCODE."
    }
    if ($changes.Count -ne 0) {
        throw "Refusing to stamp a validation build from a dirty working tree. Commit the tested revision first."
    }
    $commit = (& git rev-parse HEAD).Trim()
    if ($LASTEXITCODE -ne 0 -or $commit -notmatch '^[0-9a-f]{40}$') {
        throw "Could not resolve the tested Git revision."
    }
    if ([string]::IsNullOrWhiteSpace($Output)) {
        $goExecutableSuffix = (& go env GOEXE).Trim()
        if ($LASTEXITCODE -ne 0) {
            throw "go env GOEXE failed with exit code $LASTEXITCODE."
        }
        $fileName = "bedrock-debug-proxy$goExecutableSuffix"
        $Output = Join-Path $projectRoot (Join-Path "bin" $fileName)
    }
    $outputPath = [System.IO.Path]::GetFullPath($Output, $projectRoot)
    $outputDirectory = Split-Path -Parent $outputPath
    [System.IO.Directory]::CreateDirectory($outputDirectory) | Out-Null
    $ldflags = "-X github.com/NhanAZ/BedrockDebugProxy/internal/buildinfo.Version=$Version -X github.com/NhanAZ/BedrockDebugProxy/internal/buildinfo.Commit=$commit"
    & go build -trimpath "-ldflags=$ldflags" -o $outputPath ./cmd/bedrock-debug-proxy
    if ($LASTEXITCODE -ne 0) {
        throw "Validation build failed with exit code $LASTEXITCODE."
    }
    Write-Host "Built revision $commit at $outputPath"
} finally {
    Pop-Location
}
