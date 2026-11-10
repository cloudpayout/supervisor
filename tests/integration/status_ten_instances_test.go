//go:build slow

package integration

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestStatusTenInstances(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	ids := make([]string, 0, 10)
	t.Cleanup(func() {
		for _, id := range ids {
			stopID(t, bin, env, id)
		}
	})
	for i := 0; i < 10; i++ {
		ids = append(ids, startHeadless(t, bin, env))
	}
	out := runStatusOnce(t, bin, env)
	for _, id := range ids {
		if !strings.Contains(out, id) {
			t.Fatalf("missing %s in %s", id, out)
		}
	}
	if !strings.Contains(out, "CPU%") || !strings.Contains(out, "MEMORY") {
		t.Fatalf("missing columns: %s", out)
	}

	cmd := exec.Command(bin, "status", "-f")
	cmd.Env = env
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), "CPU%") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(buf.String(), "CPU%") {
		t.Fatalf("no initial follow output: %q", buf.String())
	}
	before := strings.Count(buf.String(), "CPU%")
	refreshDeadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(refreshDeadline) {
		if strings.Count(buf.String(), "CPU%") >= before+1 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	after := strings.Count(buf.String(), "CPU%")
	if after < before+1 {
		t.Fatalf("expected a second status refresh with 10 instances; headers before=%d after=%d", before, after)
	}
	followOut := buf.String()
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
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("status -f did not exit after SIGINT")
	}
	followOut = buf.String()
	for _, id := range ids {
		if !strings.Contains(followOut, id) {
			t.Fatalf("missing %s in follow output", id)
		}
	}
}
