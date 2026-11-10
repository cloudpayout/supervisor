package integration

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestStartRegistersAndReachesRunning(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	cmd := exec.Command(bin, "start", "--headless")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("start: %v %s", err, out)
	}
	id := strings.TrimSpace(string(out))
	if id == "" {
		t.Fatal("missing id")
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		doc := readDoc(t, regPath)
		for _, inst := range doc.Instances {
			if inst.ID == id && inst.State == "running" && inst.OwnerPID > 0 && inst.BrowserPID > 0 {
				t.Cleanup(func() { stopID(t, bin, env, id) })
				return
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatalf("instance %s did not reach running", id)
}

func stopID(t *testing.T, bin string, env []string, id string) {
	t.Helper()
	cmd := exec.Command(bin, "stop", id)
	cmd.Env = env
	_, _ = cmd.CombinedOutput()
}
