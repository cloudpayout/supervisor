package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/registry"
)

func TestScreenshotDefaultWritesPNG(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	before := waitRunning(t, regPath, id)

	work := t.TempDir()
	first := runScreenshot(t, bin, env, work, id)
	assertPNG(t, filepath.Join(work, first))
	after := waitRunning(t, regPath, id)
	if after.OwnerPID != before.OwnerPID || after.BrowserPID != before.BrowserPID {
		t.Fatalf("pids changed: before %+v after %+v", before, after)
	}

	second := runScreenshot(t, bin, env, work, id)
	if second == first {
		t.Fatalf("second capture reused %s", first)
	}
	firstBytes, err := os.ReadFile(filepath.Join(work, first))
	if err != nil {
		t.Fatal(err)
	}
	again, err := os.ReadFile(filepath.Join(work, first))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, again) {
		t.Fatal("first png changed")
	}
	assertPNG(t, filepath.Join(work, second))
	still := waitRunning(t, regPath, id)
	if still.OwnerPID != before.OwnerPID || still.BrowserPID != before.BrowserPID {
		t.Fatalf("pids changed after second capture: %+v", still)
	}
}

func TestScreenshotTwoHeadlessInstances(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	a := startHeadless(t, bin, env)
	b := startHeadless(t, bin, env)
	t.Cleanup(func() {
		stopID(t, bin, env, a)
		stopID(t, bin, env, b)
	})
	beforeA := waitRunning(t, regPath, a)
	beforeB := waitRunning(t, regPath, b)
	work := t.TempDir()
	fileA := runScreenshot(t, bin, env, work, a)
	fileB := runScreenshot(t, bin, env, work, b)
	assertPNG(t, filepath.Join(work, fileA))
	assertPNG(t, filepath.Join(work, fileB))
	afterA := waitRunning(t, regPath, a)
	afterB := waitRunning(t, regPath, b)
	if afterA.OwnerPID != beforeA.OwnerPID || afterA.BrowserPID != beforeA.BrowserPID {
		t.Fatalf("instance a pids changed")
	}
	if afterB.OwnerPID != beforeB.OwnerPID || afterB.BrowserPID != beforeB.BrowserPID {
		t.Fatalf("instance b pids changed")
	}
}

func waitRunning(t *testing.T, regPath, id string) registry.Instance {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		doc := readDoc(t, regPath)
		for _, inst := range doc.Instances {
			if inst.ID == id && inst.State == registry.StateRunning && inst.OwnerPID > 0 && inst.BrowserPID > 0 {
				return inst
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("instance %s did not reach running", id)
	return registry.Instance{}
}

func waitPresent(t *testing.T, regPath, id string) registry.Instance {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(regPath); err == nil {
			doc := readDoc(t, regPath)
			for _, inst := range doc.Instances {
				if inst.ID == id {
					return inst
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("instance %s not in registry", id)
	return registry.Instance{}
}

func runScreenshot(t *testing.T, bin string, env []string, dir, id string) string {
	t.Helper()
	namePattern := regexp.MustCompile(`^screenshot-` + regexp.QuoteMeta(id) + `-\d{8}-\d{6}(-\d+)?\.png$`)
	deadline := time.Now().Add(20 * time.Second)
	var last string
	for time.Now().Before(deadline) {
		cmd := exec.Command(bin, "screenshot", id)
		cmd.Env = env
		cmd.Dir = dir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		last = stderr.String()
		if err == nil {
			name := strings.TrimSpace(stdout.String())
			if strings.Contains(name, "\n") {
				t.Fatalf("stdout has more than one line: %q", name)
			}
			if !namePattern.MatchString(name) {
				t.Fatalf("stdout %q", name)
			}
			return name
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("screenshot %s: %s", id, last)
	return ""
}

func assertPNG(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 8 || !bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("%s is not a PNG (%d bytes)", path, len(data))
	}
}
