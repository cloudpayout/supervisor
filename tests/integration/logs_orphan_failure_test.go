package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogsAfterStop(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	stopID(t, bin, env, id)
	waitProfileGone(t, env, id)

	one := runLogs(t, bin, env, id)
	if !strings.Contains(one.stderr, "no longer in status") {
		t.Fatalf("stderr %q stdout %q", one.stderr, one.stdout)
	}
	recent := runLogs(t, bin, env, "--recent")
	if !strings.Contains(recent.stdout, id) {
		t.Fatalf("recent missing %s: %s", id, recent.stdout)
	}
}

func waitProfileGone(t *testing.T, env []string, id string) {
	t.Helper()
	var data string
	for _, item := range env {
		if strings.HasPrefix(item, "XDG_DATA_HOME=") {
			data = strings.TrimPrefix(item, "XDG_DATA_HOME=")
		}
	}
	profile := filepath.Join(data, "supervisor", "profiles", id)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		_ = os.RemoveAll(profile)
		_, err := os.Stat(profile)
		if os.IsNotExist(err) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
