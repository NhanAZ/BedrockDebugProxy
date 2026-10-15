package main

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGolemHelpAndInvalidInput(t *testing.T) {
	var out, diagnostic bytes.Buffer
	if code := run([]string{"help", "golem"}, &out, &diagnostic); code != 0 {
		t.Fatal(code)
	}
	for _, text := range []string{"--data", "--request", "--input", "--timeout", "stdin", "simulation", "capabilities", "move", "stop", "personality", "autonomy"} {
		if !strings.Contains(out.String(), text) {
			t.Errorf("help lacks %q", text)
		}
	}
	for _, input := range []string{"", "null", `{}`, `{"op":"reset","typo":true}`, `{"op":"reset"} {"op":"spawn"}`} {
		out.Reset()
		diagnostic.Reset()
		if code := runGolemInput(context.Background(), t.TempDir(), strings.NewReader(input), &out, &diagnostic, time.Second); code != 2 {
			t.Errorf("input %q exit=%d", input, code)
		}
		if out.Len() != 0 {
			t.Errorf("invalid input wrote stdout %s", out.String())
		}
	}
	for _, args := range [][]string{{"golem"}, {"golem", "--data", "x", "--timeout", "0s"}, {"golem", "--data", "x", "--request", "{}", "--input", "-"}, {"golem", "--data", "x", "--request", ""}} {
		if code := run(args, &out, &diagnostic); code != 2 {
			t.Errorf("args=%v exit=%d", args, code)
		}
	}
}

func TestGolemCompanionSyntaxAndPackage(t *testing.T) {
	php, err := exec.LookPath("php")
	if err != nil {
		t.Skip("PHP is optional for the network proxy; required for the Golem companion")
	}
	root := filepath.Join("..", "..", "dev", "golem-agent")
	for _, path := range []string{"src/Agent.php", "src/Actions.php", "src/View.php", "src/Personality.php", "src/Behaviour.php", "build.php", "tests/smoke.php", "tests/personality.php", "tests/behaviour-smoke.php"} {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		// #nosec G204 -- Fixed repository PHP sources are linted with an installed runtime.
		output, err := exec.CommandContext(ctx, php, "-l", filepath.Join(root, filepath.FromSlash(path))).CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("%s: %v\n%s", path, err, output)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// #nosec G204 -- Execute the fixed deterministic personality regression fixture.
	if output, err := exec.CommandContext(ctx, php, filepath.Join(root, "tests/personality.php")).CombinedOutput(); err != nil {
		t.Fatalf("personality regression: %v\n%s", err, output)
	}
	// #nosec G204 -- Fixed packaging script and an isolated test output, without a shell.
	output, err := exec.CommandContext(ctx, php, "-d", "phar.readonly=0", filepath.Join(root, "build.php"), filepath.Join(t.TempDir(), "GolemAgent.phar")).CombinedOutput()
	if err != nil {
		t.Fatalf("build companion: %v\n%s", err, output)
	}
}
