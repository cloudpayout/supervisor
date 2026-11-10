package integration

import (
	"os/exec"
	"testing"
	"time"
)

func TestStatusOneShotExitsImmediately(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)

	done := make(chan struct{})
	var exitCode int
	var runErr error
	go func() {
		cmd := exec.Command(bin, "status")
		cmd.Env = env
		runErr = cmd.Run()
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("status did not finish within 5s")
	}
	if runErr != nil {
		t.Fatalf("status error: %v exit %d", runErr, exitCode)
	}
	if exitCode != 0 {
		t.Fatalf("status exit %d", exitCode)
	}
}

func TestStatusRejectsExtraArgs(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)

	cmd := exec.Command(bin, "status", "extra-arg")
	cmd.Env = env
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected usage error for extra positional arg")
	}
	if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != 2 {
		t.Fatalf("exit %v", cmd.ProcessState)
	}
}
