package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProfileRemovedAfterStop(t *testing.T) {
	requireHeadlessShell(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	data := ""
	for _, item := range env {
		if strings.HasPrefix(item, "XDG_DATA_HOME=") {
			data = strings.TrimPrefix(item, "XDG_DATA_HOME=")
		}
	}
	cmd := exec.Command(bin, "start", "--headless")
	cmd.Env = append(env, "SUPERVISOR_PAGE_URL=about:blank")
	cmd.Dir = moduleRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("start: %v %s", err, out)
	}
	id := strings.TrimSpace(string(out))
	profile := filepath.Join(data, "supervisor", "profiles", id)
	waitInstanceRunning(t, bin, env, id)
	if _, err := os.Stat(profile); err != nil {
		t.Fatalf("profile missing while running: %v", err)
	}
	stop := exec.Command(bin, "stop", id)
	stop.Env = env
	stop.Dir = moduleRoot(t)
	if out, err := stop.CombinedOutput(); err != nil {
		t.Fatalf("stop: %v %s", err, out)
	}
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(profile); os.IsNotExist(err) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("profile still present: %s", profile)
}

func TestProfileUniquePerInstance(t *testing.T) {
	requireHeadlessShell(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	var ids []string
	for i := 0; i < 2; i++ {
		cmd := exec.Command(bin, "start", "--headless")
		cmd.Env = append(env, "SUPERVISOR_PAGE_URL=about:blank")
		cmd.Dir = moduleRoot(t)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("start: %v %s", err, out)
		}
		id := strings.TrimSpace(string(out))
		ids = append(ids, id)
		waitInstanceRunning(t, bin, env, id)
		stop := exec.Command(bin, "stop", id)
		stop.Env = env
		stop.Dir = moduleRoot(t)
		if out, err := stop.CombinedOutput(); err != nil {
			t.Fatalf("stop: %v %s", err, out)
		}
	}
	if ids[0] == ids[1] {
		t.Fatal("instance ids reused")
	}
}

func TestProfileTenCycles(t *testing.T) {
	requireHeadlessShell(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	data := ""
	for _, item := range env {
		if strings.HasPrefix(item, "XDG_DATA_HOME=") {
			data = strings.TrimPrefix(item, "XDG_DATA_HOME=")
		}
	}
	profiles := filepath.Join(data, "supervisor", "profiles")
	for i := 0; i < 10; i++ {
		cmd := exec.Command(bin, "start", "--headless")
		cmd.Env = append(env, "SUPERVISOR_PAGE_URL=about:blank")
		cmd.Dir = moduleRoot(t)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("cycle %d start: %v %s", i, err, out)
		}
		id := strings.TrimSpace(string(out))
		waitInstanceRunning(t, bin, env, id)
		stop := exec.Command(bin, "stop", id)
		stop.Env = env
		stop.Dir = moduleRoot(t)
		if out, err := stop.CombinedOutput(); err != nil {
			t.Fatalf("cycle %d stop: %v %s", i, err, out)
		}
	}
	entries, _ := os.ReadDir(profiles)
	if len(entries) != 0 {
		t.Fatalf("leftover profiles: %d", len(entries))
	}
}

func waitInstanceRunning(t *testing.T, bin string, env []string, id string) {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		st := exec.Command(bin, "status")
		st.Env = env
		raw, _ := st.CombinedOutput()
		if strings.Contains(string(raw), id) && strings.Contains(string(raw), "running") {
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("instance %s did not reach running", id)
}
