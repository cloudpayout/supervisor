package integration

import (
	"strings"
	"testing"
)

func TestSessionInfoHelp(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	stdout, stderr, code := runSupervisor(t, bin, env, "session", "info", "--help")
	if code != 0 {
		t.Fatalf("exit %d stderr %s stdout %s", code, stderr, stdout)
	}
	for _, want := range []string{"instance", "casino", "game", "strategy", "date", "null", "casino: onewin"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("help missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "{") && strings.Contains(stdout, "\"casino\"") {
		t.Fatalf("help looks like JSON: %s", stdout)
	}
}
