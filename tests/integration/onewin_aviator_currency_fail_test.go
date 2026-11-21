package integration

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestAviatorCurrencyFailRemovesInstance(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "currency-secret-value"
	page := newStandIn(t, "session-id", secret)
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{
		"casino":"onewin",
		"currency":"ZZZ",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	if strings.Contains(id, "currency") {
		t.Fatalf("start CLI named the currency failure: %s", id)
	}
	waitGone(t, regPath, id)
	out := runStatusOnce(t, bin, env)
	if strings.Contains(out, id) || strings.Contains(out, "currency") {
		t.Fatalf("status surfaced the entry failure: %s", out)
	}
	deadline := time.Now().Add(5 * time.Second)
	var raw []byte
	for time.Now().Before(deadline) {
		var err error
		raw, err = os.ReadFile(filepathLog(t, env, id))
		if err == nil && strings.Contains(string(raw), " ERROR currency") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(string(raw), " ERROR currency") {
		t.Fatalf("worker log missing internal currency failure: %s", raw)
	}
	if strings.Contains(string(raw), secret) {
		t.Fatal("worker log contains the cookie value")
	}
	one := runLogs(t, bin, env, id)
	if one.code != 0 || !strings.Contains(one.stdout, " ERROR currency") {
		t.Fatalf("logs %s: code %d stdout %q stderr %q", id, one.code, one.stdout, one.stderr)
	}
	if !strings.Contains(one.stderr, "no longer in status") {
		t.Fatalf("stderr %q", one.stderr)
	}
	recent := runLogs(t, bin, env, "--recent")
	if !strings.Contains(recent.stdout, id) || !strings.Contains(recent.stdout, "currency") {
		t.Fatalf("recent: %s", recent.stdout)
	}
}

func filepathLog(t *testing.T, env []string, id string) string {
	t.Helper()
	for _, item := range env {
		if strings.HasPrefix(item, "XDG_STATE_HOME=") {
			return strings.TrimPrefix(item, "XDG_STATE_HOME=") + "/supervisor/logs/" + id + ".log"
		}
	}
	t.Fatal("missing XDG_STATE_HOME")
	return ""
}
