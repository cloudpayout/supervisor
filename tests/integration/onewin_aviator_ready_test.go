package integration

import (
	"strings"
	"testing"
)

func TestAviatorReadyOnStandIn(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "ready-secret-value"
	page := newStandIn(t, "session-id", secret)
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{
		"casino":"onewin",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)
	got := waitEntry(t, regPath, id, "ready")
	if got.Casino != "onewin" {
		t.Fatalf("casino %q", got.Casino)
	}
	out := runStatusOnce(t, bin, env)
	if !strings.Contains(out, id) || !strings.Contains(out, "ready") {
		t.Fatalf("status: %s", out)
	}
	assertNoLiveCasino(t, got.LogPath)
}
