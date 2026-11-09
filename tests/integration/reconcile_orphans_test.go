package integration

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/proc"
	"github.com/cloudpayout/supervisor/internal/registry"
)

func TestReconcileKillsOrphanBrowserPID(t *testing.T) {
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
			ID:         "orphan-chrome",
			State:      registry.StateRunning,
			CreatedAt:  time.Now().UTC().Add(-time.Minute),
			ProfileDir: filepath.Join(t.TempDir(), "prof"),
			OwnerPID:   1 << 30,
			BrowserPID: child.Process.Pid,
		}},
	})
	_ = runStatusOnce(t, bin, env)
	if proc.Alive(child.Process.Pid) {
		t.Fatal("orphan browser pid still alive")
	}
	doc := readDoc(t, regPath)
	if len(doc.Instances) != 0 {
		t.Fatalf("registry not pruned: %+v", doc.Instances)
	}
}
