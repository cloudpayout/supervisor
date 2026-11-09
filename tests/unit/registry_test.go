package unit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/registry"
)

func TestRegistryRoundTripAndFlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "instances.json")
	store := registry.Store{Path: path}
	headless := true
	created := time.Now().UTC().Truncate(time.Second)

	err := store.Update(func(doc *registry.Document) error {
		doc.Instances = append(doc.Instances, registry.Instance{
			ID:         "abc",
			State:      registry.StateRunning,
			CreatedAt:  created,
			ProfileDir: "/tmp/p",
			OwnerPID:   11,
			BrowserPID: 22,
			Headless:   &headless,
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["schema_version"].(float64) != 1 {
		t.Fatalf("schema: %v", parsed["schema_version"])
	}
	if !json.Valid(raw) {
		t.Fatal("invalid json")
	}
	if !contains(string(raw), `"headless": true`) {
		t.Fatalf("missing headless: %s", raw)
	}

	var got registry.Instance
	if err := store.View(func(doc *registry.Document) error {
		if len(doc.Instances) != 1 {
			t.Fatalf("len %d", len(doc.Instances))
		}
		got = doc.Instances[0]
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got.ID != "abc" || got.OwnerPID != 11 || got.Headless == nil || !*got.Headless {
		t.Fatalf("round trip %+v", got)
	}

	const n = 20
	var wg sync.WaitGroup
	var seq atomic.Int64
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			err := store.Update(func(doc *registry.Document) error {
				id := seq.Add(1)
				doc.Instances = append(doc.Instances, registry.Instance{
					ID:         fmt.Sprintf("id-%d", id),
					State:      registry.StateStarting,
					CreatedAt:  time.Now().UTC(),
					ProfileDir: "/tmp/x",
				})
				return nil
			})
			if err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if err := store.View(func(doc *registry.Document) error {
		if len(doc.Instances) != n+1 {
			t.Fatalf("concurrent appends: got %d want %d", len(doc.Instances), n+1)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReconcileStartingTimeoutAndSplitBrain(t *testing.T) {
	now := time.Now().UTC()
	old := now.Add(-61 * time.Second)
	doc := &registry.Document{SchemaVersion: 1, Instances: []registry.Instance{
		{ID: "stale", State: registry.StateStarting, CreatedAt: old, ProfileDir: "/p/stale"},
		{ID: "fresh", State: registry.StateStarting, CreatedAt: now, ProfileDir: "/p/fresh"},
		{ID: "chrome", State: registry.StateRunning, CreatedAt: old, ProfileDir: "/p/c", OwnerPID: 10, BrowserPID: 20},
		{ID: "owner", State: registry.StateRunning, CreatedAt: old, ProfileDir: "/p/o", OwnerPID: 30, BrowserPID: 40},
		{ID: "dead", State: registry.StateRunning, CreatedAt: old, ProfileDir: "/p/d", OwnerPID: 50, BrowserPID: 60},
	}}
	alive := func(pid int) bool {
		return pid == 20 || pid == 30
	}
	planned := registry.Plan(doc, now, alive)
	got := map[string]registry.Reap{}
	for _, r := range planned {
		got[r.ID] = r
	}
	if _, ok := got["fresh"]; ok {
		t.Fatal("fresh starting instance was reaped")
	}
	if _, ok := got["stale"]; !ok {
		t.Fatal("stale starting instance was kept")
	}
	if _, ok := got["chrome"]; !ok {
		t.Fatal("owner-dead chrome-alive instance was kept")
	}
	if _, ok := got["owner"]; !ok {
		t.Fatal("chrome-dead owner-alive instance was kept")
	}
	if r, ok := got["dead"]; !ok || !r.ProcessesGone {
		t.Fatal("both-dead instance was not marked processes gone")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (func() bool { return stringIndex(s, sub) >= 0 })())
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
