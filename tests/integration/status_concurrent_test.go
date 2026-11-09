package integration

import (
	"strings"
	"testing"
	"time"
)

func TestStatusReflectsConcurrentChanges(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })

	done := make(chan string, 1)
	go func() {
		done <- runStatusOnce(t, bin, env)
	}()
	time.Sleep(500 * time.Millisecond)
	stopID(t, bin, env, id)
	out := <-done
	if strings.Contains(out, id) && strings.Count(out, id) > 4 {
		t.Fatalf("duplicate id rows: %s", out)
	}
}

func startHeadless(t *testing.T, bin string, env []string) string {
	t.Helper()
	requireChrome(t)
	cmd := execCommand(bin, env, "start", "--headless")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("start: %v %s", err, out)
	}
	return strings.TrimSpace(string(out))
}
