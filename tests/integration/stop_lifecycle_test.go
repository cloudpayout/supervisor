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

func TestStopLifecycleMessages(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)

	cmd := exec.Command(bin, "stop")
	cmd.Env = env
	cmd.Stdin = strings.NewReader("\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("empty stop: %v %s", err, out)
	}
	if !strings.Contains(string(out), "stop aborted") {
		t.Fatalf("abort message: %s", out)
	}

	cmd = exec.Command(bin, "stop")
	cmd.Env = env
	cmd.Stdin = strings.NewReader("missing-id\n")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("not found exit: %v %s", err, out)
	}
	if !strings.Contains(string(out), "instance not found: missing-id") {
		t.Fatalf("not found message: %s", out)
	}

	if err := os.MkdirAll(filepath.Dir(regPath), 0o755); err != nil {
		t.Fatal(err)
	}
	doc := registry.Document{
		SchemaVersion: 1,
		Instances: []registry.Instance{
			{
				ID:         "gone1",
				State:      registry.StateRunning,
				CreatedAt:  time.Now().UTC().Add(-time.Minute),
				ProfileDir: filepath.Join(t.TempDir(), "gone"),
				OwnerPID:   1 << 30,
				BrowserPID: 1<<30 + 1,
			},
			{
				ID:         "stale-start",
				State:      registry.StateStarting,
				CreatedAt:  time.Now().UTC().Add(-2 * time.Minute),
				ProfileDir: filepath.Join(t.TempDir(), "stale"),
			},
		},
	}
	raw, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(regPath, append(raw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command(bin, "stop")
	cmd.Env = env
	cmd.Stdin = strings.NewReader("gone1\n")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("already stopped exit: %v %s", err, out)
	}
	if !strings.Contains(string(out), "instance already stopped: gone1") {
		t.Fatalf("already stopped message: %s", out)
	}
	left := readDoc(t, regPath)
	for _, inst := range left.Instances {
		if inst.ID == "gone1" || inst.ID == "stale-start" {
			t.Fatalf("row %s still present", inst.ID)
		}
	}
}

func readDoc(t *testing.T, path string) registry.Document {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc registry.Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}
