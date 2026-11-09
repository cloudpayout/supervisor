package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestStartLaunchFailureLeavesRegistryUntouched(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	env = append(env, "CHROME_PATH=/nonexistent/chrome")
	cmd := exec.Command(bin, "start", "--headless")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure, output %s", out)
	}
	if strings.Contains(string(out), "\n") && len(strings.TrimSpace(string(out))) < 8 {
		t.Fatalf("unexpected short stdout id: %s", out)
	}
	if _, statErr := os.Stat(regPath); statErr == nil {
		t.Fatalf("registry was written: %s", regPath)
	}
	lower := strings.ToLower(string(out))
	if !strings.Contains(lower, "chrome") {
		t.Fatalf("expected chrome error, got %s", out)
	}
}
