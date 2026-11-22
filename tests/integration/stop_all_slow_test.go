//go:build slow

package integration

import (
	"strings"
	"testing"
	"time"
)

func TestStopAllFiveInstances(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	var ids []string
	for i := 0; i < 5; i++ {
		ids = append(ids, startHeadless(t, bin, env))
	}
	start := time.Now()
	_, stderr, code := runStopAll(t, bin, env, nil)
	if code != 0 {
		t.Fatalf("exit %d stderr %s", code, stderr)
	}
	if time.Since(start) > 60*time.Second {
		t.Fatalf("stop all took %s", time.Since(start))
	}
	status := runStatusOnce(t, bin, env)
	for _, id := range ids {
		if strings.Contains(status, id) {
			t.Fatalf("status still has %s: %s", id, status)
		}
	}
	if !strings.Contains(status, "No instances are running.") {
		t.Fatalf("status: %s", status)
	}
}
