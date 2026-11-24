package integration

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestHeadlessIgnoresChromePath(t *testing.T) {
	requireHeadlessShell(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	env = append(env, "CHROME_PATH=/usr/bin/google-chrome", "SUPERVISOR_PAGE_URL=about:blank")
	cmd := exec.Command(bin, "start", "--headless")
	cmd.Env = env
	cmd.Dir = moduleRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("start: %v %s", err, out)
	}
	id := strings.TrimSpace(string(out))
	t.Cleanup(func() {
		stop := exec.Command(bin, "stop", id)
		stop.Env = env
		stop.Dir = moduleRoot(t)
		_ = stop.Run()
	})
	deadline := time.Now().Add(45 * time.Second)
	var cmdline string
	for time.Now().Before(deadline) {
		st := exec.Command(bin, "status")
		st.Env = env
		raw, _ := st.CombinedOutput()
		if strings.Contains(string(raw), "running") {
			break
		}
		time.Sleep(time.Second)
	}
	// Process list should mention the bundled shell, not only CHROME_PATH selection.
	ps, _ := exec.Command("ps", "-eo", "args").Output()
	cmdline = string(ps)
	if !strings.Contains(cmdline, "chrome-headless-shell") {
		t.Fatalf("expected chrome-headless-shell in process list, status started id %s", id)
	}
}
