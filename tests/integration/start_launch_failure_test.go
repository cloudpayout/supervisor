package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartLaunchFailureLeavesRegistryUntouched(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	env = append(env, "CHROME_PATH=/nonexistent/chrome")
	cmd := exec.Command(bin, "start")
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

func TestStartHeadlessFailsWithoutBundledShell(t *testing.T) {
	bin := supervisorBin(t)
	root := moduleRoot(t)
	shell := filepath.Join(root, "bin", "chrome-headless-shell")
	hidden := shell + ".missing-test"
	if _, err := os.Lstat(shell); err == nil {
		if err := os.Rename(shell, hidden); err != nil {
			t.Fatalf("hide shell: %v", err)
		}
		t.Cleanup(func() {
			if err := os.Rename(hidden, shell); err != nil {
				t.Errorf("restore shell: %v", err)
			}
		})
	}
	env, regPath := isolatedEnv(t)
	cmd := exec.Command(bin, "start", "--headless")
	cmd.Env = env
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure, output %s", out)
	}
	text := string(out)
	lower := strings.ToLower(text)
	if !strings.Contains(lower, "chrome-headless-shell") || !strings.Contains(lower, "fetch-chrome-headless-shell") {
		t.Fatalf("expected missing shell error citing the fetch script, got %s", out)
	}
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, " ") || strings.Contains(line, "/") {
			continue
		}
		if len(line) >= 8 && len(line) <= 16 {
			t.Fatalf("unexpected instance id line: %q in %s", line, out)
		}
	}
	if _, statErr := os.Stat(regPath); statErr == nil {
		t.Fatalf("registry was written: %s", regPath)
	}
}
