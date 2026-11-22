package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/registry"
)

func TestStopAllTwoInstances(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	id1 := startHeadless(t, bin, env)
	id2 := startHeadless(t, bin, env)

	stdout, stderr, code := runStopAll(t, bin, env, nil)
	if code != 0 {
		t.Fatalf("exit %d stdout %s stderr %s", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "stopped instance "+id1) || !strings.Contains(stderr, "stopped instance "+id2) {
		t.Fatalf("per-id lines: %s", stderr)
	}
	if !strings.Contains(stderr, "stop all: stopped 2, skipped 0 (already stopped or not found)") {
		t.Fatalf("summary: %s", stderr)
	}
	status := runStatusOnce(t, bin, env)
	if !strings.Contains(status, "No instances are running.") {
		t.Fatalf("status: %s", status)
	}
}

func TestStopAllEmptyRegistry(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	start := time.Now()
	_, stderr, code := runStopAll(t, bin, env, nil)
	if time.Since(start) > 3*time.Second {
		t.Fatalf("took %s", time.Since(start))
	}
	if code != 0 {
		t.Fatalf("exit %d stderr %s", code, stderr)
	}
	if !strings.Contains(stderr, "No instances are running.") {
		t.Fatalf("stderr: %s", stderr)
	}
	if _, err := os.Stat(regPath); err == nil {
		doc := readDoc(t, regPath)
		if len(doc.Instances) != 0 {
			t.Fatalf("instances: %+v", doc.Instances)
		}
	}
}

func TestStopAllEmptyRegistryRepeated(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	writeEmptyRegistry(t, regPath)
	before, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		_, stderr, code := runStopAll(t, bin, env, nil)
		if code != 0 {
			t.Fatalf("run %d exit %d %s", i, code, stderr)
		}
		if !strings.Contains(stderr, "No instances are running.") {
			t.Fatalf("run %d stderr %s", i, stderr)
		}
	}
	after, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(t, before, after) {
		t.Fatalf("registry changed\nbefore %s\nafter %s", before, after)
	}
}

func TestStopAllAllStaleRows(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
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
		},
	}
	raw, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(regPath, append(raw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	_, stderr, code := runStopAll(t, bin, env, nil)
	if code != 0 {
		t.Fatalf("exit %d stderr %s", code, stderr)
	}
	// Initial reconciliation removes dead rows before the snapshot, so bulk
	// stop sees an empty registry (FR-003) rather than per-ID already-stopped lines.
	if !strings.Contains(stderr, "No instances are running.") {
		t.Fatalf("stderr: %s", stderr)
	}
	left := readDoc(t, regPath)
	if len(left.Instances) != 0 {
		t.Fatalf("rows left: %+v", left.Instances)
	}
	status := runStatusOnce(t, bin, env)
	if !strings.Contains(status, "No instances are running.") {
		t.Fatalf("status: %s", status)
	}
}

func TestStopHelpDocumentsAll(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	cmd := execCommand(bin, env, "stop", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help: %v %s", err, out)
	}
	text := string(out)
	if !strings.Contains(text, "all") || !strings.Contains(text, "instance-id") {
		t.Fatalf("help missing forms:\n%s", text)
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "prompt") && strings.Contains(lower, "interactive prompt") {
		t.Fatalf("help mentions interactive prompt:\n%s", text)
	}
	if strings.Contains(lower, "stdin prompt") || strings.Contains(text, "asks for") {
		t.Fatalf("help mentions prompting:\n%s", text)
	}
}

func TestStopAllUsageErrors(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)

	cmd := execCommand(bin, env, "stop")
	out, err := cmd.CombinedOutput()
	if err == nil || cmd.ProcessState.ExitCode() != 2 {
		t.Fatalf("missing arg exit %v %s", err, out)
	}

	cmd = execCommand(bin, env, "stop", "all", "extra")
	out, err = cmd.CombinedOutput()
	if err == nil || cmd.ProcessState.ExitCode() != 2 {
		t.Fatalf("extra arg exit %v %s", err, out)
	}

	cmd = execCommand(bin, env, "stop", "All")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("All exit: %v %s", err, out)
	}
	if !strings.Contains(string(out), "instance not found: All") {
		t.Fatalf("All message: %s", out)
	}
}

func TestStopAllSIGINT(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	t.Cleanup(func() { stopAllInstancesForTeardown(t, bin, env, regPath) })
	id1 := startHeadless(t, bin, env)
	id2 := startHeadless(t, bin, env)
	env = append(env, "SUPERVISOR_STOP_ALL_PAUSE=3s")

	cmd := execCommand(bin, env, "stop", "all")
	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	cmd.Stdin = stdin
	var buf lockedBuffer
	cmd.Stderr = &buf
	cmd.Stdout = &buf
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), "stopped instance") {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !strings.Contains(buf.String(), "stopped instance") {
		_ = cmd.Process.Kill()
		t.Fatalf("no stop line before signal: %s", buf.String())
	}
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	waitErr := cmd.Wait()
	if waitErr == nil || cmd.ProcessState.ExitCode() == 0 {
		t.Fatalf("expected non-zero exit, output %s", buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "stop all interrupted: completed") {
		t.Fatalf("interrupt summary: %s", out)
	}
	left := readDoc(t, regPath)
	remaining := map[string]bool{}
	for _, inst := range left.Instances {
		remaining[inst.ID] = true
	}
	stoppedCount := 0
	if !remaining[id1] {
		stoppedCount++
	}
	if !remaining[id2] {
		stoppedCount++
	}
	if stoppedCount < 1 {
		t.Fatalf("expected at least one instance removed, left %+v output %s", left.Instances, out)
	}
}

// stopAllInstancesForTeardown stops leftover instances after interrupt tests so
// Chrome releases profile dirs before t.TempDir() cleanup (Go 1.20+).
func stopAllInstancesForTeardown(t *testing.T, bin string, env []string, regPath string) {
	t.Helper()
	if _, err := os.Stat(regPath); err == nil {
		doc := readDoc(t, regPath)
		for _, inst := range doc.Instances {
			stopID(t, bin, env, inst.ID)
		}
	}
	_, _, _ = runStopAll(t, bin, env, nil)
	prof := profilesDir(env)
	if prof == "" {
		return
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		entries, err := os.ReadDir(prof)
		if err != nil || len(entries) == 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	if entries, _ := os.ReadDir(prof); len(entries) != 0 {
		t.Errorf("profiles left after teardown: %d under %s", len(entries), prof)
	}
}

func runStopAll(t *testing.T, bin string, env []string, extra []string) (stdout, stderr string, code int) {
	t.Helper()
	args := append([]string{"stop", "all"}, extra...)
	cmd := execCommand(bin, env, args...)
	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	cmd.Stdin = stdin
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	code = 0
	if err != nil {
		if cmd.ProcessState != nil {
			code = cmd.ProcessState.ExitCode()
		} else {
			code = 1
		}
	}
	return outBuf.String(), errBuf.String(), code
}

func writeEmptyRegistry(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	doc := registry.Document{SchemaVersion: 1, Instances: []registry.Instance{}}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func jsonEqual(t *testing.T, a, b []byte) bool {
	t.Helper()
	var da, db any
	if err := json.Unmarshal(a, &da); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &db); err != nil {
		t.Fatal(err)
	}
	ab, _ := json.Marshal(da)
	bb, _ := json.Marshal(db)
	return bytes.Equal(ab, bb)
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

var _ io.Writer = (*lockedBuffer)(nil)
