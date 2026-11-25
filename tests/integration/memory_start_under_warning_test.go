package integration

import (
	"strings"
	"testing"
	"time"
)

func TestStartSucceedsWhilePeerInMemoryWarning(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	a := startWith(t, bin, env, "start", "--headless", "--memory-limit", "1M", "--threshold", "50")
	t.Cleanup(func() { stopID(t, bin, env, a) })
	deadline := time.Now().Add(120 * time.Second)
	for {
		out := runStatusOnce(t, bin, env)
		if strings.Contains(out, a) && strings.Contains(out, "threshold exceeded") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("peer never warned: %s", out)
		}
		time.Sleep(time.Second)
	}
	b := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, b) })
	if b == a || b == "" {
		t.Fatalf("second start id %q", b)
	}
}
