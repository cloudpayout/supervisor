package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMemoryFlagsUsage(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)

	before := instanceCount(t, regPath)
	cmd := exec.Command(bin, "start", "--headless", "--threshold", "90")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected usage error, got %s", out)
	}
	if id := strings.TrimSpace(string(out)); looksLikeID(id) {
		t.Fatalf("printed id on usage error: %s", out)
	}
	if instanceCount(t, regPath) != before {
		t.Fatalf("registry grew after --threshold without limit")
	}

	before = instanceCount(t, regPath)
	bad := exec.Command(bin, "start", "--headless", "--memory-limit", "512MB")
	bad.Env = env
	out, err = bad.CombinedOutput()
	if err == nil {
		t.Fatalf("expected reject 512MB, got %s", out)
	}
	if strings.Contains(string(out), "\n") && looksLikeID(lastLine(string(out))) {
		t.Fatalf("printed id: %s", out)
	}
	if instanceCount(t, regPath) != before {
		t.Fatal("registry grew after 512MB")
	}
}

func TestStartInvalidMemoryLimitNoRegistryRow(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	before := instanceCount(t, regPath)
	cmd := exec.Command(bin, "start", "--headless", "--memory-limit", "512MB")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected error: %s", out)
	}
	if instanceCount(t, regPath) != before {
		t.Fatalf("phantom row: before %d after %d out %s", before, instanceCount(t, regPath), out)
	}
}

func instanceCount(t *testing.T, path string) int {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		return 0
	}
	return len(readDoc(t, path).Instances)
}

func looksLikeID(s string) bool {
	if len(s) < 4 || strings.Contains(s, " ") || strings.Contains(s, "usage") {
		return false
	}
	return true
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return lines[len(lines)-1]
}
