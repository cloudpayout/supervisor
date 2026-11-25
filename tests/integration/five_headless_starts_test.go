package integration

import (
	"strings"
	"testing"
	"time"
)

func TestFiveConcurrentHeadlessBlankStarts(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	var ids []string
	for i := 0; i < 5; i++ {
		ids = append(ids, startHeadless(t, bin, env))
	}
	t.Cleanup(func() {
		for _, id := range ids {
			stopID(t, bin, env, id)
		}
	})
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id] = true
	}
	deadline := time.Now().Add(90 * time.Second)
	for {
		out := runStatusOnce(t, bin, env)
		all := true
		for _, id := range ids {
			if !strings.Contains(out, id) || !strings.Contains(out, "running") {
				all = false
			}
		}
		if all && strings.Count(out, "running") >= 5 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("not all running: %s", out)
		}
		time.Sleep(time.Second)
	}
}
