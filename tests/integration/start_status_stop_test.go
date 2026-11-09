package integration

import (
	"strings"
	"testing"
)

func TestStartStatusStop(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	out := runStatusOnce(t, bin, env)
	if !strings.Contains(out, id) {
		t.Fatalf("status missing %s: %s", id, out)
	}
	stopID(t, bin, env, id)
	doc := readDoc(t, regPath)
	for _, inst := range doc.Instances {
		if inst.ID == id {
			t.Fatal("instance still registered after stop")
		}
	}
}
