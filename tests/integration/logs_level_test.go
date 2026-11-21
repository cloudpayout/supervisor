package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogsLevelFilter(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	state := ""
	for _, item := range env {
		if strings.HasPrefix(item, "XDG_STATE_HOME=") {
			state = strings.TrimPrefix(item, "XDG_STATE_HOME=")
		}
	}
	dir := filepath.Join(state, "supervisor", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	id := "abc123"
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	body := ts + " INFO hello\n" + ts + " NOTICE phase\n" + ts + " ERROR boom\n"
	path := filepath.Join(dir, id+".log")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	errOut := runLogs(t, bin, env, id, "--level", "error")
	if errOut.stderr != "" && !strings.Contains(errOut.stderr, "no longer in status") {
		t.Fatalf("stderr %q", errOut.stderr)
	}
	if strings.TrimSpace(errOut.stdout) != ts+" ERROR boom" {
		t.Fatalf("error filter: %q", errOut.stdout)
	}
	noteOut := runLogs(t, bin, env, id, "--level", "notice")
	if strings.Contains(noteOut.stdout, " INFO ") {
		t.Fatalf("notice filter kept info: %q", noteOut.stdout)
	}
	if !strings.Contains(noteOut.stdout, " NOTICE phase") || !strings.Contains(noteOut.stdout, " ERROR boom") {
		t.Fatalf("notice filter: %q", noteOut.stdout)
	}
}

func TestLogsLevelErrorMatchesNothing(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	dir := filepath.Join(xdgState(t, env), "supervisor", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	id := "quiet1"
	body := ts + " INFO hello\n" + ts + " NOTICE phase\n"
	if err := os.WriteFile(filepath.Join(dir, id+".log"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out := runLogs(t, bin, env, id, "--level", "error")
	if strings.TrimSpace(out.stdout) != "" {
		t.Fatalf("expected empty stdout, got %q", out.stdout)
	}
}

type cmdResult struct {
	stdout string
	stderr string
	code   int
}

func runLogs(t *testing.T, bin string, env []string, args ...string) cmdResult {
	t.Helper()
	cmd := execCommand(bin, env, append([]string{"logs"}, args...)...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := exitCode(err)
	if code != 0 {
		t.Fatalf("logs %v: exit %d stderr %s", args, code, stderr.String())
	}
	return cmdResult{stdout: stdout.String(), stderr: stderr.String(), code: code}
}
