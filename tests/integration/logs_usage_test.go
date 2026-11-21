package integration

import (
	"strings"
	"testing"
)

func TestLogsUsageRequiresID(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	for _, args := range [][]string{{"logs"}, {"logs", "--level", "error"}} {
		cmd := execCommand(bin, env, args...)
		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if exitCode(err) != 2 {
			t.Fatalf("%v: exit %v stderr %q", args, err, stderr.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("%v: stdout %q", args, stdout.String())
		}
		if !strings.Contains(stderr.String(), "instance ID required") {
			t.Fatalf("%v: stderr %q", args, stderr.String())
		}
	}
}

func TestLogsUnknownID(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	cmd := execCommand(bin, env, "logs", "missing-instance")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if exitCode(err) == 0 {
		t.Fatal("expected non-zero exit")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "log not found") {
		t.Fatalf("stderr %q", stderr.String())
	}
}
