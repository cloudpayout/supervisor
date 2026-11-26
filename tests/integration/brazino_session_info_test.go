package integration

import (
	"testing"
)

func TestSessionInfoBrazinoStandIn(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	secret := "brazino-info-secret"
	page := newStandIn(t, "session-id", secret)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{
		"casino":"brazino",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)

	stdout, stderr, code := runSessionInfo(t, bin, env, id)
	if code != 0 || stderr != "" {
		t.Fatalf("exit %d stderr %q stdout %q", code, stderr, stdout)
	}
	assertSessionLines(t, stdout, "brazino", "null")
}