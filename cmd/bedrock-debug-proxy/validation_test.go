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
	for _, name := range []string{"legacy", "complete", "missing", "open", "errors", "wrong_revision", "wrong_capture", "wrong_sequence", "invalid_json"} {
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
			write("analysis.json", map[string]any{
				"capture_id": "synthetic", "status": "closed", "observed_events": 2,
				"build": map[string]any{"commit": commit}, "manifest_counts": map[string]any{}, "protocol": map[string]any{},
				"session":    map[string]any{"upstream_connected": true, "negotiated": true, "spawned": true, "closed": true, "decoded_packet_events": 1},
				"directions": []map[string]any{{"name": "client_to_server", "count": 1}, {"name": "server_to_client", "count": 1}},
			})
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
			// #nosec G204 -- Fixed project script and synthetic t.TempDir paths, passed as separate arguments.
			output, err := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-File", script, "-Binary", binary, "-Capture", root, "-Server", "Synthetic test", "-Result", "pass", "-Check", "normal_session=pass", "-Output", report).CombinedOutput()
			wantPass := name == "legacy" || name == "complete"
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
