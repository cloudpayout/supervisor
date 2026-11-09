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

func runStatusOnce(t *testing.T, bin string, env []string) string {
	t.Helper()
	cmd := exec.Command(bin, "status")
	cmd.Env = env
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if buf.Len() > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	_ = cmd.Process.Signal(syscall.SIGINT)
	waitDone := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-time.After(3 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("status did not exit on SIGINT")
	}
	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 && cmd.ProcessState.ExitCode() != -1 {
		t.Fatalf("status exit %d output %s", cmd.ProcessState.ExitCode(), buf.String())
	}
	return buf.String()
}
