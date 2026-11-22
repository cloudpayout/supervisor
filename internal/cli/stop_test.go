package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/registry"
)

func TestRunStopAllAggregatesStopInstanceErrors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))

	layout, err := registry.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.StateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pid := os.Getpid()
	now := time.Now().UTC()
	doc := registry.Document{
		SchemaVersion: registry.SchemaVersion,
		Instances: []registry.Instance{
			{ID: "first", State: registry.StateRunning, CreatedAt: now, OwnerPID: pid, BrowserPID: pid, ProfileDir: t.TempDir()},
			{ID: "second", State: registry.StateRunning, CreatedAt: now, OwnerPID: pid, BrowserPID: pid, ProfileDir: t.TempDir()},
		},
	}
	store := registry.Store{Path: layout.RegistryPath()}
	if err := store.Update(func(d *registry.Document) error {
		*d = doc
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	orig := stopOne
	t.Cleanup(func() { stopOne = orig })
	var got []string
	stopOne = func(id string) (stopOutcome, error) {
		got = append(got, id)
		if id == "second" {
			return 0, fmt.Errorf("registry update failed")
		}
		return outcomeStopped, nil
	}

	err = runStopAll()
	if err == nil {
		t.Fatal("expected aggregated error")
	}
	if len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("stop order: %v", got)
	}
}
