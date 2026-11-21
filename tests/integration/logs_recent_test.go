package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogsRecentOrphans(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	state := xdgState(t, env)
	dir := filepath.Join(state, "supervisor", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeLog := func(id, body string, age time.Duration) {
		t.Helper()
		path := filepath.Join(dir, id+".log")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		when := time.Now().Add(-age)
		if err := os.Chtimes(path, when, when); err != nil {
			t.Fatal(err)
		}
	}
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	writeLog("old", ts+" INFO only\n", 2*time.Hour)
	writeLog("new", ts+" ERROR disk full\n", time.Minute)

	out := runLogs(t, bin, env, "--recent")
	if !strings.Contains(out.stdout, "new") || !strings.Contains(out.stdout, "disk full") {
		t.Fatalf("recent: %s", out.stdout)
	}
	if !strings.Contains(out.stdout, "no error recorded") {
		t.Fatalf("placeholder missing: %s", out.stdout)
	}
	idxNew := strings.Index(out.stdout, "new")
	idxOld := strings.Index(out.stdout, "old")
	if idxNew < 0 || idxOld < 0 || idxNew > idxOld {
		t.Fatalf("order: %s", out.stdout)
	}

	one := runLogs(t, bin, env, "new")
	if !strings.Contains(one.stderr, "no longer in status") {
		t.Fatalf("stderr %q", one.stderr)
	}
	if !strings.Contains(one.stdout, "ERROR disk full") {
		t.Fatalf("stdout %q", one.stdout)
	}
}

func TestLogsRecentCap(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	dir := filepath.Join(xdgState(t, env), "supervisor", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	base := time.Now().Add(-time.Hour)
	for i := 0; i < 55; i++ {
		id := fmt.Sprintf("id%02d", i)
		path := filepath.Join(dir, id+".log")
		if err := os.WriteFile(path, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		when := base.Add(time.Duration(i) * time.Second)
		if err := os.Chtimes(path, when, when); err != nil {
			t.Fatal(err)
		}
	}
	cmd := execCommand(bin, env, "logs", "--recent")
	var stdout strings.Builder
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	text := stdout.String()
	if !strings.Contains(text, "5 additional orphaned logs were not shown") {
		t.Fatalf("cap notice: %s", text)
	}
	if strings.Count(text, "\n") > 53 {
		t.Fatalf("too many lines: %d\n%s", strings.Count(text, "\n"), text)
	}
}

func xdgState(t *testing.T, env []string) string {
	t.Helper()
	for _, item := range env {
		if strings.HasPrefix(item, "XDG_STATE_HOME=") {
			return strings.TrimPrefix(item, "XDG_STATE_HOME=")
		}
	}
	t.Fatal("missing XDG_STATE_HOME")
	return ""
}
