package integration

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestAviatorGameConfigErrorRemovesInstance(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "game-config-secret"
	page := newStandIn(t, "session-id", secret)
	page.spribeConfigError = true
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{
		"casino":"onewin",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	waitGone(t, regPath, id)
	out := runStatusOnce(t, bin, env)
	if strings.Contains(out, id) {
		t.Fatalf("status still lists instance: %s", out)
	}
	deadline := time.Now().Add(5 * time.Second)
	var raw []byte
	for time.Now().Before(deadline) {
		var err error
		raw, err = os.ReadFile(filepathLog(t, env, id))
		if err == nil && strings.Contains(string(raw), " ERROR game_config") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(string(raw), " ERROR game_config") {
		t.Fatalf("worker log missing game_config failure: %s", raw)
	}
	if strings.Contains(string(raw), secret) {
		t.Fatal("worker log contains the cookie value")
	}
}