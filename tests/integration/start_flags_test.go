package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestStartRejectsBothDisplayFlags(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	cmd := exec.Command(bin, "start", "--headed", "--headless")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected usage error, output %s", out)
	}
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 2 {
		t.Fatalf("exit %v output %s", err, out)
	}
	lower := strings.ToLower(string(out))
	if !strings.Contains(lower, "mutually exclusive") && !strings.Contains(lower, "headed") {
		t.Fatalf("usage output: %s", out)
	}
	if _, err := os.Stat(regPath); err == nil {
		t.Fatal("registry should not be created")
	}
}
