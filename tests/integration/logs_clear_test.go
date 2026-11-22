package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func supervisorDir(regPath string) string {
	return filepath.Dir(regPath)
}

func seedLog(t *testing.T, regPath, name, body string) string {
	t.Helper()
	dir := filepath.Join(supervisorDir(regPath), "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func seedInstances(t *testing.T, regPath, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(regPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(regPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runLogsClear(t *testing.T, bin string, env []string, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := execCommand(bin, env, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	out, err := cmd.Output()
	return string(out), errBuf.String(), exitCode(err)
}

func TestLogsClearYesOrphanAndRunning(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	seedInstances(t, regPath, `{"schema_version":1,"instances":[{"id":"active01","state":"running","created_at":"2026-01-01T00:00:00Z","profile_dir":"/tmp/p"}]}`+"\n")
	orphan := seedLog(t, regPath, "orphan01.log", "orphan\n")
	active := seedLog(t, regPath, "active01.log", "active\n")
	before, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runLogsClear(t, bin, env, "", "logs", "clear", "--yes")
	if code != 0 {
		t.Fatalf("exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Fatalf("orphan still present: %v", err)
	}
	if _, err := os.Stat(active); err != nil {
		t.Fatalf("active log: %v", err)
	}
	after, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("registry changed:\n%s", after)
	}
}

func TestLogsClearProtectsStarting(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	seedInstances(t, regPath, `{"schema_version":1,"instances":[{"id":"start01","state":"starting","created_at":"2026-01-01T00:00:00Z","profile_dir":"/tmp/p"}]}`+"\n")
	path := seedLog(t, regPath, "start01.log", "starting\n")
	stdout, stderr, code := runLogsClear(t, bin, env, "", "logs", "clear", "--yes")
	if code != 0 {
		t.Fatalf("exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("starting log removed: %v", err)
	}
}

func TestLogsClearYesOnlyProtected(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	seedInstances(t, regPath, `{"schema_version":1,"instances":[{"id":"run01","state":"running","created_at":"2026-01-01T00:00:00Z","profile_dir":"/tmp/p"}]}`+"\n")
	path := seedLog(t, regPath, "run01.log", "run\n")
	stdout, stderr, code := runLogsClear(t, bin, env, "", "logs", "clear", "--yes")
	if code != 0 {
		t.Fatalf("exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	if !strings.Contains(strings.ToLower(stdout), "eligible") {
		t.Fatalf("stdout: %q", stdout)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("protected log removed: %v", err)
	}
}

func TestLogsClearNonTTYRequiresYes(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	path := seedLog(t, regPath, "orphan01.log", "orphan\n")
	stdout, stderr, code := runLogsClear(t, bin, env, "yes\n", "logs", "clear")
	if code != 2 {
		t.Fatalf("exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "--yes") && !strings.Contains(stderr, "not a terminal") {
		t.Fatalf("stderr: %q", stderr)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("log deleted: %v", err)
	}
}

func TestLogsClearHelp(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	cmd := execCommand(bin, env, "logs", "clear", "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(out))
	for _, want := range []string{"starting", "running", "stopping", "--yes", "registry"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
}

func TestLogsHelpListsClearSubcommand(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	cmd := execCommand(bin, env, "logs", "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "clear") {
		t.Fatalf("logs help missing clear:\n%s", out)
	}
}

func TestLogsClearRejectsRecentFlag(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	path := seedLog(t, regPath, "orphan01.log", "orphan\n")
	stdout, stderr, code := runLogsClear(t, bin, env, "", "logs", "clear", "--recent")
	if code != 2 {
		t.Fatalf("exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("log deleted: %v", err)
	}
}
