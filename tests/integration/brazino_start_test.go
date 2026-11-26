package integration

import (
	"os"
	"strings"
	"testing"
)

func TestBrazinoStartAppliesCookie(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	secret := "brazino-start-secret"
	page := newStandIn(t, "session-id", secret)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)

	file := writeSession(t, `{
		"casino":"brazino",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)
	inst := waitRunning(t, regPath, id)
	if inst.Casino != "brazino" {
		t.Fatalf("casino %q", inst.Casino)
	}
	if !sawShown(page) {
		t.Fatalf("stand-in did not show balance: %+v", page.hitsCopy())
	}
	assertNoLiveBrazino(t, inst.LogPath)

	withheld := newStandIn(t, "session-id", "other-value")
	env2, reg2 := isolatedEnv(t)
	env2 = append(env2, "SUPERVISOR_PAGE_URL="+withheld.url)
	plain := writeSession(t, `{"casino":"brazino","cookies":[{"name":"session-id","value":"different","domain":"127.0.0.1","path":"/"}]}`)
	id2 := startCasino(t, bin, env2, plain)
	t.Cleanup(func() { stopID(t, bin, env2, id2) })
	waitRunning(t, reg2, id2)
	if sawShown(withheld) {
		t.Fatalf("withheld page showed balance: %+v", withheld.hitsCopy())
	}
}

func assertNoLiveBrazino(t *testing.T, logPath string) {
	t.Helper()
	raw, err := os.ReadFile(logPath)
	if err != nil {
		return
	}
	if strings.Contains(string(raw), "brazino777.mx") {
		t.Fatalf("worker log mentions live brazino host: %s", raw)
	}
}