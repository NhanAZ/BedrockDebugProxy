package main

import (
	"bytes"
	"context"
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
