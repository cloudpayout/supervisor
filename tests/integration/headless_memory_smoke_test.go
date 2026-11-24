package integration

import (
	"strings"
	"testing"
	"time"
)

func TestHeadlessMemorySmoke(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	deadline := time.Now().Add(30 * time.Second)
	for {
		out := runStatusOnce(t, bin, env)
		if strings.Contains(out, id) && strings.Contains(out, " MiB") && strings.Contains(out, "running") {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("no MiB memory: %s", out)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
