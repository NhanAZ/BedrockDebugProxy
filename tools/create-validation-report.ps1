[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [string]$Binary,
    [Parameter(Mandatory)]
    [string]$Capture,
    [Parameter(Mandatory)]
    [string]$Server,
    [Parameter(Mandatory)]
    [ValidateSet("pass", "fail", "incomplete")]
    [string]$Result,
    [Parameter(Mandatory)]
    [string[]]$Check,
    [Parameter(Mandatory)]
    [string]$Output
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Get-ObjectProperty {
    param(
        [Parameter(Mandatory)]
        [object]$Object,
        [Parameter(Mandatory)]
        [string]$Name,
        [object]$Default = $null
    )

    $property = $Object.PSObject.Properties[$Name]
    if ($null -eq $property) {
        return $Default
    }
    return $property.Value
}

function Get-NamedCount {
    param(
        [object[]]$Items,
        [string]$Name
    )

    foreach ($item in @($Items)) {
        if ([string](Get-ObjectProperty $item "name" "") -eq $Name) {
            return [uint64](Get-ObjectProperty $item "count" 0)
        }
    }
    return [uint64]0
}

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location $projectRoot
try {
    $binaryPath = (Resolve-Path -LiteralPath $Binary).Path
    if ((Get-Item -LiteralPath $binaryPath).PSIsContainer) {
        throw "Binary must identify a file."
    }
    $capturePath = (Resolve-Path -LiteralPath $Capture).Path
    if (-not (Get-Item -LiteralPath $capturePath).PSIsContainer) {
        throw "Capture must identify a capture directory."
    }
    $outputPath = if ([System.IO.Path]::IsPathRooted($Output)) {
        [System.IO.Path]::GetFullPath($Output)
    } else {
        [System.IO.Path]::GetFullPath((Join-Path $projectRoot $Output))
    }
    if (Test-Path -LiteralPath $outputPath) {
        throw "Refusing to replace existing validation report $outputPath."
    }
    if ([string]::IsNullOrWhiteSpace($Server)) {
        throw "Server must be a non-empty public label."
    }

    $analysisLines = @(& $binaryPath analyze $capturePath)
    $analysisExitCode = $LASTEXITCODE
    if ($analysisExitCode -ne 0) {
        throw "Capture analysis failed with exit code $analysisExitCode."
    }
    $analysisText = [string]::Join([Environment]::NewLine, $analysisLines)
    try {
        $analysis = $analysisText | ConvertFrom-Json
    } catch {
        throw "Capture analysis did not return valid JSON. $($_.Exception.Message)"
    }

    $build = Get-ObjectProperty $analysis "build"
    $commit = [string](Get-ObjectProperty $build "commit" "")
    if ($commit -notmatch '^[0-9a-f]{40}$') {
        throw "Capture does not identify an exact 40-character Git revision. Use tools/build.ps1 for the tested binary."
    }

    $manualChecksByName = @{}
    foreach ($entry in $Check) {
        if ($entry -notmatch '^(?<name>[a-z][a-z0-9_]*?)=(?<state>pass|fail|not_observed)$') {
            throw "Invalid check '$entry'. Use lowercase_name=pass, lowercase_name=fail, or lowercase_name=not_observed."
        }
        $name = $Matches.name
        if ($manualChecksByName.ContainsKey($name)) {
            throw "Manual check '$name' was supplied more than once."
        }
        $manualChecksByName[$name] = $Matches.state
    }
    $manualChecks = @(
        foreach ($name in @($manualChecksByName.Keys | Sort-Object)) {
            [ordered]@{ name = $name; result = $manualChecksByName[$name] }
        }
    )

    $directions = @(Get-ObjectProperty $analysis "directions" @())
    $clientToServerEvents = Get-NamedCount $directions "client_to_server"
    $serverToClientEvents = Get-NamedCount $directions "server_to_client"
    $session = Get-ObjectProperty $analysis "session"
    $manifestCounts = Get-ObjectProperty $analysis "manifest_counts"
    $errors = @(Get-ObjectProperty $analysis "errors" @())
    $safeErrorsByKind = @{}
    $blockingErrorEvents = [uint64]0
    $terminationErrorEvents = [uint64]0
    foreach ($item in $errors) {
        $kind = [string](Get-ObjectProperty $item "kind" "unknown")
        $operation = [string](Get-ObjectProperty $item "operation" "")
        $count = [uint64](Get-ObjectProperty $item "count" 0)
        if (-not $safeErrorsByKind.ContainsKey($kind)) {
            $safeErrorsByKind[$kind] = [uint64]0
        }
        $safeErrorsByKind[$kind] = [uint64]$safeErrorsByKind[$kind] + $count
        if ($kind -eq "bridge.read_error" -or ($kind -eq "transport.payload" -and $operation -eq "read")) {
            $terminationErrorEvents += $count
        } else {
            $blockingErrorEvents += $count
        }
    }
    $safeErrors = @(
        foreach ($kind in @($safeErrorsByKind.Keys | Sort-Object)) {
            [ordered]@{ kind = $kind; count = [uint64]$safeErrorsByKind[$kind] }
        }
    )

    $verificationIssues = @(Get-ObjectProperty $analysis "verification_issues" @())
    $resourcePacks = @(Get-ObjectProperty $analysis "resource_packs" @())
    $packDecryptions = @(Get-ObjectProperty $analysis "resource_pack_decryptions" @())
    $decryptedPackCount = @($packDecryptions | Where-Object { [string](Get-ObjectProperty $_ "status" "") -eq "decrypted" }).Count
    $packDecryptErrorCount = @($packDecryptions | Where-Object { [string](Get-ObjectProperty $_ "status" "") -eq "error" }).Count
    $automaticRequirements = [ordered]@{
        closed_capture = [string](Get-ObjectProperty $analysis "status" "") -eq "closed"
        integrity_verified = $verificationIssues.Count -eq 0
        no_capture_write_errors = [uint64](Get-ObjectProperty $manifestCounts "write_errors" 0) -eq 0
        no_dropped_events = [uint64](Get-ObjectProperty $manifestCounts "dropped" 0) -eq 0
        no_truncated_events = [uint64](Get-ObjectProperty $manifestCounts "truncated" 0) -eq 0
        upstream_connected = [bool](Get-ObjectProperty $session "upstream_connected" $false)
        session_negotiated = [bool](Get-ObjectProperty $session "negotiated" $false)
        session_spawned = [bool](Get-ObjectProperty $session "spawned" $false)
        session_closed = [bool](Get-ObjectProperty $session "closed" $false)
        client_to_server_traffic = $clientToServerEvents -gt 0
        server_to_client_traffic = $serverToClientEvents -gt 0
        decoded_packets = [uint64](Get-ObjectProperty $session "decoded_packet_events" 0) -gt 0
        no_decode_errors = [uint64](Get-ObjectProperty $session "decode_error_events" 0) -eq 0
        no_structured_view_errors = [uint64](Get-ObjectProperty $session "structured_view_errors" 0) -eq 0
        no_blocking_error_events = $blockingErrorEvents -eq 0
        no_resource_pack_decrypt_errors = $packDecryptErrorCount -eq 0
    }
    $captureManifest = Get-Content -LiteralPath (Join-Path $capturePath "manifest.json") -Raw | ConvertFrom-Json
    $captureOptions = Get-ObjectProperty $captureManifest "options" @{}
    $captureValues = Get-ObjectProperty $captureOptions "values" @{}
    $artifactFormat = [string](Get-ObjectProperty $captureValues "automatic_artifacts" "")
    if ($artifactFormat -ne "") {
        $artifactComplete = $false
        $artifactStatusPath = Join-Path $capturePath "artifacts/status.json"
        if (Test-Path -LiteralPath $artifactStatusPath -PathType Leaf) {
            try {
                $artifactStatus = Get-Content -LiteralPath $artifactStatusPath -Raw | ConvertFrom-Json
                $artifactBuild = Get-ObjectProperty $artifactStatus "generator" @{}
                $artifactComplete = $artifactFormat -eq "bedrockdebugproxy.artifacts.v1" -and
                    [string](Get-ObjectProperty $artifactStatus "schema" "") -eq $artifactFormat -and
                    [string](Get-ObjectProperty $artifactStatus "state" "") -eq "complete" -and
                    [string](Get-ObjectProperty $artifactStatus "capture_id" "") -eq [string](Get-ObjectProperty $analysis "capture_id" "") -and
                    [string](Get-ObjectProperty $artifactBuild "commit" "") -eq $commit -and
                    [uint64](Get-ObjectProperty $artifactStatus "errors" 1) -eq 0 -and
                    [uint64](Get-ObjectProperty $artifactStatus "last_source_sequence" 0) -eq [uint64](Get-ObjectProperty $analysis "observed_events" 0)
            } catch {
                Write-Warning "Automatic artifact status is unreadable. This capture cannot receive a passing validation report."
            }
        }
        $automaticRequirements["artifact_folders_complete"] = $artifactComplete
    }
    $automaticChecksPassed = -not ($automaticRequirements.Values -contains $false)
    $manualChecksPassed = $manualChecks.Count -gt 0 -and $manualChecksByName.ContainsKey("normal_session") -and
        $manualChecksByName["normal_session"] -eq "pass" -and -not ($manualChecksByName.Values -contains "fail") -and
        -not ($manualChecksByName.Values -contains "not_observed")
    if ($Result -eq "pass" -and (-not $automaticChecksPassed -or -not $manualChecksPassed)) {
        throw "A passing report requires all automatic evidence and manual checks, including normal_session=pass, to pass."
    }

    $protocol = Get-ObjectProperty $analysis "protocol"
    $report = [ordered]@{
        schema = "bedrockdebugproxy.validation.v1"
        generated_at = [DateTimeOffset]::UtcNow.ToString("o")
        tested_revision = $commit
        binary_version = [string](Get-ObjectProperty $build "version" "")
        target_server = $Server.Trim()
        capture = [ordered]@{
            id = [string](Get-ObjectProperty $analysis "capture_id" "")
            started_at = [string](Get-ObjectProperty $analysis "started_at" "")
            ended_at = [string](Get-ObjectProperty $analysis "ended_at" "")
            status = [string](Get-ObjectProperty $analysis "status" "")
            complete = [bool](Get-ObjectProperty $analysis "complete" $false)
            declared_limitation_count = @(Get-ObjectProperty $analysis "limitations" @()).Count
            observed_events = [uint64](Get-ObjectProperty $analysis "observed_events" 0)
        }
        protocol = [ordered]@{
            configured_id = [string](Get-ObjectProperty $protocol "configured_id" "")
            configured_version = [string](Get-ObjectProperty $protocol "configured_version" "")
            downstream_id = [int64](Get-ObjectProperty $protocol "downstream_id" 0)
            downstream_version = [string](Get-ObjectProperty $protocol "downstream_version" "")
            upstream_id = [int64](Get-ObjectProperty $protocol "upstream_id" 0)
            upstream_version = [string](Get-ObjectProperty $protocol "upstream_version" "")
        }
        evidence = [ordered]@{
            automatic_requirements = $automaticRequirements
            automatic_checks_passed = $automaticChecksPassed
            connection_metadata_views = [uint64](Get-ObjectProperty $session "connection_views" 0)
            game_data_views = [uint64](Get-ObjectProperty $session "game_data_views" 0)
            client_to_server_events = $clientToServerEvents
            server_to_client_events = $serverToClientEvents
            raw_packet_events = [uint64](Get-ObjectProperty $session "raw_packet_events" 0)
            transport_payload_events = [uint64](Get-ObjectProperty $session "transport_payload_events" 0)
            decoded_packet_events = [uint64](Get-ObjectProperty $session "decoded_packet_events" 0)
            unknown_packet_events = [uint64](Get-ObjectProperty $session "unknown_packet_events" 0)
            termination_error_events = $terminationErrorEvents
            blocking_error_events = $blockingErrorEvents
            error_kinds = $safeErrors
            resource_pack_count = $resourcePacks.Count
            decrypted_resource_pack_count = $decryptedPackCount
            resource_pack_decrypt_error_count = $packDecryptErrorCount
        }
        manual_checks = $manualChecks
        manual_checks_passed = $manualChecksPassed
        result = $Result
        limitations = @(
            "This report contains derived counts and manual checks, not raw packet evidence"
            "Capture limitation text is omitted because it may contain sensitive identifiers; review the local capture before accepting the result"
        )
    }

    $outputDirectory = Split-Path -Parent $outputPath
    [System.IO.Directory]::CreateDirectory($outputDirectory) | Out-Null
    $json = $report | ConvertTo-Json -Depth 12
    [System.IO.File]::WriteAllText($outputPath, $json + "`n", [System.Text.UTF8Encoding]::new($false))
    Write-Host "Wrote validation report for revision $commit to $outputPath"
} finally {
    Pop-Location
}
