package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/artifacts"
)

func TestValidationReportRequiresCompleteArtifactFolders(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("PowerShell is unavailable")
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "tools", "create-validation-report.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	commit := strings.Repeat("a", 40)
	for _, name := range []string{"legacy", "complete", "missing", "open", "errors", "wrong_revision", "wrong_capture", "wrong_sequence", "invalid_json", "cli_complete", "cli_missing_auth", "cli_missing_stability", "cli_packs_only", "cli_blocking_error"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			write := func(name string, value any) {
				t.Helper()
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, name), data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			// Synthetic analysis isolates the report gate. These temporary reports
			// are unit-test fixtures, never evidence of a real client session.
			session := map[string]any{"upstream_connected": true, "negotiated": true, "spawned": true, "closed": true, "decoded_packet_events": 1}
			analysis := map[string]any{
				"capture_id": "synthetic", "status": "closed", "observed_events": 2,
				"build": map[string]any{"commit": commit}, "manifest_counts": map[string]any{}, "protocol": map[string]any{},
				"session":    session,
				"directions": []map[string]any{{"name": "client_to_server", "count": 1}, {"name": "server_to_client", "count": 1}},
			}
			if name == "cli_packs_only" {
				session["spawned"] = false
			}
			if name == "cli_blocking_error" {
				analysis["errors"] = []map[string]any{{"kind": "packet.decode_error", "count": 1}}
			}
			write("analysis.json", analysis)
			values := map[string]string{}
			if name != "legacy" {
				values["automatic_artifacts"] = artifacts.Schema
			}
			write("manifest.json", map[string]any{"options": map[string]any{"values": values}})
			if name != "missing" && name != "legacy" {
				if err := os.Mkdir(filepath.Join(root, "artifacts"), 0o700); err != nil {
					t.Fatal(err)
				}
				status := map[string]any{
					"schema": artifacts.Schema, "capture_id": "synthetic", "state": "complete",
					"generator": map[string]any{"commit": commit}, "errors": 0, "last_source_sequence": 2,
				}
				switch name {
				case "open":
					status["state"] = "open"
				case "errors":
					status["errors"] = 1
				case "wrong_revision":
					status["generator"] = map[string]any{"commit": strings.Repeat("b", 40)}
				case "wrong_capture":
					status["capture_id"] = "another-capture"
				case "wrong_sequence":
					status["last_source_sequence"] = 1
				}
				write("artifacts/status.json", status)
				if name == "invalid_json" {
					if err := os.WriteFile(filepath.Join(root, "artifacts/status.json"), []byte("{"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			binary := filepath.Join(root, "synthetic-analysis.ps1")
			if err := os.WriteFile(binary, []byte("Get-Content -LiteralPath (Join-Path $PSScriptRoot 'analysis.json') -Raw\nexit 0\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()
			report := filepath.Join(root, "report.json")
			checks := []string{"normal_session=pass"}
			if strings.HasPrefix(name, "cli_") {
				checks = []string{"cli_session=pass", "authenticated_upstream=pass", "post_spawn_stability=pass"}
				if name == "cli_missing_auth" {
					checks = []string{"cli_session=pass", "post_spawn_stability=pass"}
				}
				if name == "cli_missing_stability" {
					checks = []string{"cli_session=pass", "authenticated_upstream=pass"}
				}
			}
			write("checks.json", checks)
			wrapper := filepath.Join(root, "invoke.ps1")
			invocation := "param([string]$ReportScript)\n& $ReportScript -Binary (Join-Path $PSScriptRoot 'synthetic-analysis.ps1') -Capture $PSScriptRoot -Server 'Synthetic test' -Result pass -Check @(Get-Content -Raw (Join-Path $PSScriptRoot 'checks.json') | ConvertFrom-Json) -Output (Join-Path $PSScriptRoot 'report.json')\n"
			if err := os.WriteFile(wrapper, []byte(invocation), 0o600); err != nil {
				t.Fatal(err)
			}
			// #nosec G204 -- Fixed project script and synthetic t.TempDir paths, passed as separate arguments.
			output, err := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-File", wrapper, "-ReportScript", script).CombinedOutput()
			wantPass := name == "legacy" || name == "complete" || name == "cli_complete"
			if (err == nil) != wantPass {
				t.Fatalf("want pass=%t, error=%v\n%s", wantPass, err, output)
			}
			if !wantPass && !strings.Contains(string(output), "A passing report requires all automatic evidence") {
				t.Fatalf("report failed for an unexpected reason: %s", output)
			}
			if !wantPass {
				if _, err := os.Stat(report); !os.IsNotExist(err) {
					t.Fatal("failed validation unexpectedly produced a passing report")
				}
			}
		})
	}
}

func TestReleaseReadinessRequiresFullSessionEvidence(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("PowerShell is unavailable")
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "tools", "check-release-readiness.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	commit := strings.Repeat("a", 40)
	servers := []string{"The Hive", "Galaxite", "Lifeboat", "Mineville Zeqa", "Enchanted", "CubeCraft"}
	for _, name := range []string{"cli", "retail", "missing_auth", "missing_stability", "missing_transfer", "missing_risk", "failed_automatic", "failed_result", "wrong_revision", "packs_only", "ambiguous_session"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			for _, server := range servers {
				checks := map[string]string{"cli_session": "pass", "authenticated_upstream": "pass", "post_spawn_stability": "pass"}
				if name == "retail" {
					checks = map[string]string{"normal_session": "pass"}
				}
				if server == "Enchanted" {
					checks["transfer_following"] = "pass"
					checks["resource_pack_transfer"] = "pass"
				}
				if server == "CubeCraft" {
					checks["risk_acknowledged"] = "pass"
				}
				revision, result, automatic := commit, "pass", true
				if server == "The Hive" {
					switch name {
					case "missing_auth":
						delete(checks, "authenticated_upstream")
					case "missing_stability":
						delete(checks, "post_spawn_stability")
					case "failed_automatic":
						automatic = false
					case "failed_result":
						result = "fail"
					case "wrong_revision":
						revision = strings.Repeat("b", 40)
					case "packs_only":
						checks = map[string]string{"normal_session": "not_observed", "packs_exchange": "pass"}
					case "ambiguous_session":
						checks["normal_session"] = "pass"
					}
				}
				if server == "Enchanted" && name == "missing_transfer" {
					delete(checks, "transfer_following")
				}
				if server == "CubeCraft" && name == "missing_risk" {
					delete(checks, "risk_acknowledged")
				}
				var review []map[string]string
				for key, value := range checks {
					review = append(review, map[string]string{"name": key, "result": value})
				}
				// These synthetic reports deliberately retain true summary flags so
				// missing review evidence cannot pass by trusting those flags alone.
				report := map[string]any{"schema": "bedrockdebugproxy.validation.v1", "tested_revision": revision, "target_server": server, "result": result, "evidence": map[string]any{"automatic_checks_passed": automatic}, "manual_checks_passed": true, "manual_checks": review}
				data, err := json.Marshal(report)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, server+".json"), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()
			// #nosec G204 -- Fixed project script and synthetic t.TempDir paths, passed as separate arguments.
			output, err := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-File", script, "-Revision", commit, "-Reports", root).CombinedOutput()
			wantPass := name == "cli" || name == "retail"
			if (err == nil) != wantPass {
				t.Fatalf("want pass=%t, error=%v\n%s", wantPass, err, output)
			}
			if !wantPass && !strings.Contains(string(output), "Release validation is incomplete") {
				t.Fatalf("unexpected failure: %s", output)
			}
		})
	}
}
