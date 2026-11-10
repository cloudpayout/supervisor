package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/registry"
)

func TestStatusEmptyAndStartingUnavailable(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)

	out := runStatusOnce(t, bin, env)
	if !strings.Contains(out, "No instances are running.") {
		t.Fatalf("empty status: %s", out)
	}

	if err := os.MkdirAll(filepath.Dir(regPath), 0o755); err != nil {
		t.Fatal(err)
	}
	doc := registry.Document{
		SchemaVersion: 1,
		Instances: []registry.Instance{{
			ID:         "starting1",
			State:      registry.StateStarting,
			CreatedAt:  time.Now().UTC(),
			ProfileDir: filepath.Join(t.TempDir(), "profile"),
		}},
	}
	raw, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(regPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	out = runStatusOnce(t, bin, env)
	if !strings.Contains(out, "starting1") || !strings.Contains(out, "unavailable") || !strings.Contains(out, "starting") {
		t.Fatalf("starting row: %s", out)
	}
}

func TestStatusFollowExitsOnSIGINT(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)

	cmd := exec.Command(bin, "status", "-f")
	cmd.Env = env
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	waitDone := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-time.After(3 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("status -f did not exit on SIGINT")
	}
	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("status -f exit %d output %s", cmd.ProcessState.ExitCode(), buf.String())
	}
}

func runStatusOnce(t *testing.T, bin string, env []string) string {
	t.Helper()
	cmd := exec.Command(bin, "status")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("status: %v %s", err, out)
	}
	return string(out)
}
