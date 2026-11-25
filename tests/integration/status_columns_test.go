package integration

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestStatusColumnsNoteAndUnlimitedMemory(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })

	var out string
	deadline := time.Now().Add(30 * time.Second)
	for {
		out = runStatusOnce(t, bin, env)
		if strings.Contains(out, "NOTE") && strings.Contains(out, id) && strings.Contains(out, "running") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("status never running: %s", out)
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !strings.Contains(out, "NOTE") {
		t.Fatalf("header missing NOTE: %s", out)
	}
	line := rowFor(t, out, id)
	if strings.Contains(line, "/") {
		t.Fatalf("unlimited row has limit suffix: %s", line)
	}
	if !regexp.MustCompile(`\d+\.\d+ MiB\s*$`).MatchString(strings.TrimSpace(line)) {
		t.Fatalf("unlimited MEMORY: %s", line)
	}
	if strings.Contains(line, "threshold exceeded") {
		t.Fatalf("NOTE set: %s", line)
	}
}

func rowFor(t *testing.T, out, id string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), id) {
			return line
		}
	}
	t.Fatalf("missing %s in %s", id, out)
	return ""
}
