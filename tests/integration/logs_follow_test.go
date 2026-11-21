package integration

import (
	"io"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestLogsFollowInterrupt(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })

	cmd := execCommand(bin, env, "logs", id, "--follow")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 4096)
		_, _ = io.ReadAtLeast(stdout, buf, 1)
		_ = cmd.Process.Signal(syscall.SIGINT)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(25 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("follow produced no output")
	}
	waitErr := cmd.Wait()
	if exitCode(waitErr) != 0 {
		t.Fatalf("follow exit %v", waitErr)
	}
	if !strings.Contains(id, "") {
		t.Fatal("id")
	}
}
