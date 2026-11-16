package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScreenshotOutputPath(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)

	work := t.TempDir()
	shots := filepath.Join(work, "shots")
	if err := os.Mkdir(shots, 0o755); err != nil {
		t.Fatal(err)
	}
	rel := runScreenshotOutput(t, bin, env, work, id, "shots/page.png")
	if rel != "shots/page.png" {
		t.Fatalf("printed %q", rel)
	}
	assertPNG(t, filepath.Join(work, "shots", "page.png"))
	defaults, err := filepath.Glob(filepath.Join(work, "screenshot-*.png"))
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 0 {
		t.Fatalf("default files created: %v", defaults)
	}

	abs := filepath.Join(work, "absolute.png")
	printed := runScreenshotOutput(t, bin, env, work, id, abs)
	if printed != abs {
		t.Fatalf("printed %q", printed)
	}
	assertPNG(t, abs)

	before, err := os.ReadFile(abs)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(bin, "screenshot", id, "-o", abs)
	cmd.Env = env
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	if exitCode(err) != 1 || !strings.Contains(string(out), "screenshot failed:") {
		t.Fatalf("overwrite exit %d output %s", exitCode(err), out)
	}
	after, err := os.ReadFile(abs)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("existing png changed")
	}
}

func runScreenshotOutput(t *testing.T, bin string, env []string, dir, id, output string) string {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last string
	for time.Now().Before(deadline) {
		cmd := exec.Command(bin, "screenshot", id, "-o", output)
		cmd.Env = env
		cmd.Dir = dir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		last = stderr.String()
		if err == nil {
			return strings.TrimSpace(stdout.String())
		}
		if strings.Contains(last, "already exists") {
			t.Fatalf("output rejected: %s", last)
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("screenshot -o %s: %s", output, last)
	return ""
}
