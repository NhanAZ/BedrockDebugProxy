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
)

func TestMojangProtocolWatcherReleaseMetadata(t *testing.T) {
	pwsh, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("PowerShell is unavailable")
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "tools", "check-mojang-protocol.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, tag, value, protocol string
		issue                      bool
	}{
		{"plain", "v1.26.51", "2193", "2193", false},
		{"italic release body", "v1.26.51", "*2193*", "2193", false},
		{"bold", "v1.26.51", "**2193**", "2193", false},
		{"inline code", "v1.26.51", "`2193`", "2193", false},
		{"mismatched emphasis", "v1.26.51", "*2193**", "unknown", false},
		{"new protocol only", "v1.26.51", "*2194*", "2194", true},
		{"new game version", "v1.26.52", "*2193*", "2193", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			fixturePath := filepath.Join(root, "release.json")
			issuePath := filepath.Join(root, "issue.json")
			fixture, err := json.Marshal(map[string]string{
				"tag_name": test.tag, "body": "- Network protocol version: " + test.value + "\r\n",
				"html_url": "https://example.invalid/release", "published_at": "2026-10-05T00:00:00Z",
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(fixturePath, fixture, 0o600); err != nil {
				t.Fatal(err)
			}
			quote := func(path string) string { return "'" + strings.ReplaceAll(path, "'", "''") + "'" }
			command := "$env:GITHUB_TOKEN = 'synthetic-watcher-test-token'\n" +
				"$fixture = Get-Content -Raw -LiteralPath " + quote(fixturePath) + " | ConvertFrom-Json\n" +
				"$issuePath = " + quote(issuePath) + "\n" + `
function Invoke-RestMethod {
    param([string]$Uri, [string]$Method, [object]$Body, [object]$Headers, [string]$ContentType)
    if ($Uri.EndsWith('/releases/latest')) { return $fixture }
    if ($Uri.Contains('/git/ref/tags/')) {
        return @{object = @{type = 'commit'; sha = '1111111111111111111111111111111111111111'}}
    }
    if ($Uri.Contains('/issues')) {
        if ($Method -eq 'Post') {
            $Body | Set-Content -LiteralPath $issuePath
            return @{number = 1; html_url = 'https://example.invalid/issue'}
        }
        return @()
    }
    throw "Unexpected request $Uri"
}
` + "& " + quote(script) + " -Repository fixture/project -SourceRepository fixture/mojang -CreateIssue\n"
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			// #nosec G204 -- The installed PowerShell runs quoted synthetic fixtures and the repository's watcher with every HTTP request mocked.
			output, err := exec.CommandContext(ctx, pwsh, "-NoProfile", "-Command", command).CombinedOutput()
			if err != nil {
				t.Fatalf("watcher: %v\n%s", err, output)
			}
			if !strings.Contains(string(output), "protocol "+test.protocol+")") {
				t.Fatalf("wrong detected protocol: %s", output)
			}
			issueData, err := os.ReadFile(issuePath)
			if !test.issue {
				if !os.IsNotExist(err) {
					t.Fatalf("unchanged or unknown protocol created an issue: %s, %v", issueData, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var issue struct {
				Title string `json:"title"`
				Body  string `json:"body"`
			}
			if err := json.Unmarshal(issueData, &issue); err != nil {
				t.Fatal(err)
			}
			if issue.Title != "[Protocol update] Mojang "+test.tag+" (protocol "+test.protocol+")" {
				t.Fatalf("issue title = %q", issue.Title)
			}
			for _, expected := range []string{
				"| Game version | `1.26.51` | `" + strings.TrimPrefix(test.tag, "v") + "` |",
				"| Network protocol | `2193` | `" + test.protocol + "` |",
				"| Release tag | n/a | `" + test.tag + "` |",
				"| Release commit | n/a | `1111111111111111111111111111111111111111` |",
				"| Published | n/a | `2026-10-05T00:00:00Z` |",
			} {
				if !strings.Contains(issue.Body, expected) {
					t.Fatalf("issue is missing %q: %s", expected, issue.Body)
				}
			}
		})
	}
}
