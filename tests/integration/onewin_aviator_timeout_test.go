package integration

import "testing"

func TestAviatorTimeoutRemovesInstance(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "timeout-secret-value"
	page := newStandIn(t, "session-id", secret)
	page.blockReady = true
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{
		"casino":"onewin",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	waitGone(t, regPath, id)
	assertStatusGone(t, bin, env, id)
}
