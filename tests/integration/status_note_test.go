package integration

import (
	"strings"
	"testing"
	"time"
)

func TestNoteThresholdExceeded(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	id := startWith(t, bin, env, "start", "--headless", "--memory-limit", "1M", "--threshold", "50")
	t.Cleanup(func() { stopID(t, bin, env, id) })
	deadline := time.Now().Add(120 * time.Second)
	for {
		out := runStatusOnce(t, bin, env)
		if strings.Contains(out, id) && strings.Contains(out, "threshold exceeded") {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("NOTE not set: %s", out)
		}
		time.Sleep(time.Second)
	}
}
