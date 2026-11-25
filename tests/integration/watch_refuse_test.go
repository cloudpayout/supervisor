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

func TestWatchUsageAndLookup(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)

	for _, args := range [][]string{
		{"watch"},
		{"watch", "   "},
		{"watch", "some-id", "extra"},
		{"watch", "id", "--fps", "0"},
		{"watch", "id", "--fps", "31"},
		{"watch", "id", "--fps", "abc"},
	} {
		cmd := exec.Command(bin, args...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if exitCode(err) != 2 {
			t.Fatalf("%v exit %d output %s", args, exitCode(err), out)
		}
		if strings.Contains(strings.Join(args, " "), "--fps") && !strings.Contains(string(out), "invalid --fps") {
			t.Fatalf("%v missing invalid --fps: %s", args, out)
		}
		if strings.Contains(string(out), "http://127.0.0.1:") {
			t.Fatalf("%v printed url: %s", args, out)
		}
	}

	cmd := exec.Command(bin, "watch", "missing-id")
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
			{ID: "w1", State: registry.StateStarting, CreatedAt: time.Now().UTC(), ProfileDir: t.TempDir()},
		},
	}
	raw, _ := json.Marshal(doc)
	if err := os.WriteFile(regPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command(bin, "watch", "w1")
	cmd.Env = env
	out, err = cmd.CombinedOutput()
	if exitCode(err) != 1 || !strings.Contains(string(out), "instance not running: w1") {
		t.Fatalf("not running exit %d output %s", exitCode(err), out)
	}
}

func TestWatchSecondSessionRefused(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)

	first := execCommand(bin, env, "watch", id)
	stdout, err := first.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := bufioReadLine(stdout); err != nil {
		t.Fatal(err)
	}

	second := exec.Command(bin, "watch", id)
	second.Env = env
	out, err := second.CombinedOutput()
	if exitCode(err) != 1 || !strings.Contains(string(out), "already watching") {
		t.Fatalf("second watch exit %d %s", exitCode(err), out)
	}

	_ = first.Process.Kill()
	_ = first.Wait()
}

func bufioReadLine(r interface{ Read([]byte) (int, error) }) (string, error) {
	buf := make([]byte, 0, 128)
	b := make([]byte, 1)
	for {
		n, err := r.Read(b)
		if n > 0 {
			buf = append(buf, b[0])
			if b[0] == '\n' {
				return strings.TrimSpace(string(buf)), nil
			}
		}
		if err != nil {
			return "", err
		}
	}
}
