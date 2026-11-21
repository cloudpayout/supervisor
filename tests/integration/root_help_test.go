package integration

import (
	"strings"
	"testing"
)

func TestRootHelpNoCompletion(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)

	for _, args := range [][]string{nil, {"help"}, {"--help"}} {
		cmd := execCommand(bin, env, args...)
		if args == nil {
			cmd = execCommand(bin, env)
		}
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		text := string(out)
		for _, want := range []string{"start", "status", "stop", "screenshot", "logs", "help"} {
			if !strings.Contains(text, want) {
				t.Fatalf("%v: help missing %q:\n%s", args, want, text)
			}
		}
		if strings.Contains(text, "\n  completion ") || strings.Contains(text, "\ncompletion ") {
			t.Fatalf("%v: help lists completion:\n%s", args, text)
		}
		if strings.Contains(text, "worker") {
			t.Fatalf("%v: help must not list worker:\n%s", args, text)
		}
	}
}

func TestCompletionUnknownCommand(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)

	cases := []struct {
		args       []string
		wantExit   int
		stderrNeed string
	}{
		{[]string{"completion"}, 2, "unknown command"},
		{[]string{"completion", "bash"}, 2, "unknown command"},
		// FR-004: Cobra help reports unavailable topic (exit 0), not a completion script.
		{[]string{"help", "completion"}, 0, "Unknown help topic"},
	}
	for _, tc := range cases {
		cmd := execCommand(bin, env, tc.args...)
		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if exitCode(err) != tc.wantExit {
			t.Fatalf("%v: want exit %d, got %v (stdout %q stderr %q)", tc.args, tc.wantExit, err, out, stderr.String())
		}
		if len(out) != 0 {
			t.Fatalf("%v: want empty stdout, got %q", tc.args, out)
		}
		if !strings.Contains(stderr.String(), tc.stderrNeed) {
			t.Fatalf("%v: stderr want %q, got %q", tc.args, tc.stderrNeed, stderr.String())
		}
	}
}
