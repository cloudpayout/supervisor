package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/registry"
)

func TestScreenshotUsageAndLookup(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)

	for _, args := range [][]string{
		{"screenshot"},
		{"screenshot", "   "},
		{"screenshot", "some-id", "extra"},
	} {
		cmd := exec.Command(bin, args...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if exitCode(err) != 2 {
			t.Fatalf("%v exit %d output %s", args, exitCode(err), out)
		}
	}

	cmd := exec.Command(bin, "screenshot", "missing-id")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if exitCode(err) != 1 || !strings.Contains(string(out), "instance not found: missing-id") {
		t.Fatalf("not found exit %d output %s", exitCode(err), out)
	}

	if err := os.MkdirAll(filepath.Dir(regPath), 0o755); err != nil {
		t.Fatal(err)
	}
	doc := registry.Document{
		SchemaVersion: 1,
		Instances: []registry.Instance{
			{ID: "start1", State: registry.StateStarting, CreatedAt: time.Now().UTC(), ProfileDir: filepath.Join(t.TempDir(), "s")},
			{ID: "stop1", State: registry.StateStopped, CreatedAt: time.Now().UTC(), ProfileDir: filepath.Join(t.TempDir(), "t")},
		},
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(regPath, append(raw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"start1", "stop1"} {
		cmd := exec.Command(bin, "screenshot", id)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if exitCode(err) != 1 || !strings.Contains(string(out), "instance not running: "+id) {
			t.Fatalf("%s exit %d output %s", id, exitCode(err), out)
		}
	}
	left := readDoc(t, regPath)
	if len(left.Instances) != 2 {
		t.Fatalf("registry changed: %+v", left.Instances)
	}
}

func TestScreenshotHelp(t *testing.T) {
	bin := supervisorBin(t)
	cmd := exec.Command(bin, "screenshot", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help: %v %s", err, out)
	}
	text := string(out)
	for _, want := range []string{"instance", "screenshot-", "--output", "-o", "relative", "absolute"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(strings.ToLower(text), "prompt") && strings.Contains(strings.ToLower(text), "type an id") {
		t.Fatalf("help describes a prompt:\n%s", text)
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	return -1
}
