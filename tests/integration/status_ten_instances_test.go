//go:build slow

package integration

import (
	"strings"
	"testing"
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
}
