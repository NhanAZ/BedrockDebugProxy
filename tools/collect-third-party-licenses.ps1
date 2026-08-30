[CmdletBinding()]
param(
    [string]$Package = "./cmd/bedrock-debug-proxy",
    [string]$Output = ""
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location $projectRoot
try {
    $dependencyTemplate = '{{with .Module}}{{if .Path}}{{.Path}}{{end}}{{end}}'
    $compiledModules = @(& go list -deps -f $dependencyTemplate $Package | Where-Object {
            -not [string]::IsNullOrWhiteSpace($_) -and $_ -ne "github.com/NhanAZ/BedrockDebugProxy"
        } | Sort-Object -Unique)
    if ($LASTEXITCODE -ne 0) {
        throw "List compiled dependencies failed with exit code $LASTEXITCODE."
    }
    if ($compiledModules.Count -eq 0) {
        throw "No compiled third-party modules were found for $Package."
    }

    $moduleTemplate = '{{.Path}}|{{.Version}}|{{.Dir}}'
    $moduleDetails = @{}
    foreach ($line in @(& go list -m -f $moduleTemplate all)) {
        $parts = $line -split '\|', 3
        if ($parts.Count -ne 3) {
            throw "Could not parse module metadata line '$line'."
        }
        $moduleDetails[$parts[0]] = [pscustomobject]@{
            Version = $parts[1]
            Directory = $parts[2]
        }
    }
    if ($LASTEXITCODE -ne 0) {
        throw "List module metadata failed with exit code $LASTEXITCODE."
    }

    $inventory = @()
    foreach ($modulePath in $compiledModules) {
        if (-not $moduleDetails.ContainsKey($modulePath)) {
            throw "Compiled module $modulePath is missing from the module graph."
        }
        $details = $moduleDetails[$modulePath]
        $licenseFiles = @(Get-ChildItem -LiteralPath $details.Directory -File | Where-Object {
                $_.Name -match '^(LICENSE|LICENCE|COPYING|NOTICE|PATENTS|AUTHORS|CONTRIBUTORS|CREDITS|COPYRIGHT)(\.|$)'
            } | Sort-Object Name)
        if ($licenseFiles.Count -eq 0) {
            throw "Compiled module $modulePath $($details.Version) has no root license or notice file."
        }
        $inventory += [pscustomobject]@{
            Path = $modulePath
            Version = $details.Version
            Files = $licenseFiles
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($Output)) {
        $outputPath = if ([System.IO.Path]::IsPathRooted($Output)) {
            [System.IO.Path]::GetFullPath($Output)
        } else {
            [System.IO.Path]::GetFullPath((Join-Path $projectRoot $Output))
        }
        if (Test-Path -LiteralPath $outputPath) {
            throw "Refusing to replace existing third-party license bundle $outputPath."
        }

        $sections = [System.Collections.Generic.List[string]]::new()
        $sections.Add("BedrockDebugProxy third-party license bundle")
        $sections.Add("Compiled package $Package")
        $sections.Add("This generated file reproduces license, notice, patent, copyright, and attribution files from the exact Go modules compiled into the release package.")
        foreach ($item in $inventory) {
            foreach ($file in $item.Files) {
                $content = [System.IO.File]::ReadAllText($file.FullName)
                $content = $content.Replace("`r`n", "`n").Replace("`r", "`n").TrimEnd([char[]]"`n")
                $sections.Add("")
                $sections.Add("================================================================================")
                $sections.Add("Module $($item.Path) $($item.Version)")
                $sections.Add("File $($file.Name)")
                $sections.Add("================================================================================")
                $sections.Add($content)
            }
        }

        $outputDirectory = Split-Path -Parent $outputPath
        [System.IO.Directory]::CreateDirectory($outputDirectory) | Out-Null
        $bundle = [string]::Join("`n", $sections) + "`n"
        [System.IO.File]::WriteAllText($outputPath, $bundle, [System.Text.UTF8Encoding]::new($false))
        Write-Host "Wrote licenses for $($inventory.Count) compiled third-party modules to $outputPath"
    } else {
        Write-Host "Verified license files for $($inventory.Count) compiled third-party modules."
    }
} finally {
    Pop-Location
}
