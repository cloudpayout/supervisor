package integration

import (
	"strings"
	"testing"
	"time"
)

func TestAviatorStatusShowsEnteringBeforeReady(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "entering-secret-value"
	page := newStandIn(t, "session-id", secret)
	page.slowEntry = true
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{
		"casino":"onewin",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	t.Cleanup(func() { stopID(t, bin, env, id) })

	deadline := time.Now().Add(30 * time.Second)
	sawEntering := false
	for time.Now().Before(deadline) && !sawEntering {
		out := runStatusOnce(t, bin, env)
		if strings.Contains(out, id) && strings.Contains(out, "entering") {
			sawEntering = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !sawEntering {
		t.Fatal("status never showed entering")
	}
	got := waitEntry(t, regPath, id, "ready")
	out := runStatusOnce(t, bin, env)
	if !strings.Contains(out, id) || !strings.Contains(out, "ready") {
		t.Fatalf("status after ready: %s", out)
	}
	if got.Entry != "ready" {
		t.Fatalf("entry %q", got.Entry)
	}
}
