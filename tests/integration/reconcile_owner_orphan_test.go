package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/proc"
	"github.com/cloudpayout/supervisor/internal/registry"
)

func TestReconcileKillsOrphanOwnerPID(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	child := exec.Command("sleep", "30")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if child.Process != nil {
			_ = child.Process.Kill()
			_, _ = child.Process.Wait()
		}
	})
	writeRegistry(t, regPath, registry.Document{
		SchemaVersion: 1,
		Instances: []registry.Instance{{
			ID:         "orphan-owner",
			State:      registry.StateRunning,
			CreatedAt:  time.Now().UTC().Add(-time.Minute),
			ProfileDir: filepath.Join(t.TempDir(), "prof"),
			OwnerPID:   child.Process.Pid,
			BrowserPID: 1 << 30,
		}},
	})
	_ = runStatusOnce(t, bin, env)
	if proc.Alive(child.Process.Pid) {
		t.Fatal("orphan owner pid still alive")
	}
	doc := readDoc(t, regPath)
	if len(doc.Instances) != 0 {
		t.Fatalf("registry not pruned: %+v", doc.Instances)
	}
}

func writeRegistry(t *testing.T, path string, doc registry.Document) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
