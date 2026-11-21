package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestOneWinStartAppliesCookie(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	secret := "start-secret-value"
	page := newStandIn(t, "session-id", secret)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)

	file := writeSession(t, `{
		"casino":"onewin",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)
	inst := waitEntry(t, regPath, id, "ready")
	if inst.Casino != "onewin" {
		t.Fatalf("casino %q", inst.Casino)
	}
	if inst.Headless == nil || !*inst.Headless {
		t.Fatalf("headless flag not recorded: %+v", inst.Headless)
	}
	if !sawShown(page) {
		t.Fatalf("stand-in did not show balance: %+v", page.hitsCopy())
	}
	assertNoLiveCasino(t, inst.LogPath)

	withheld := newStandIn(t, "session-id", "other-value")
	env2, reg2 := isolatedEnv(t)
	env2 = append(env2, "SUPERVISOR_PAGE_URL="+withheld.url)
	plain := writeSession(t, `{"casino":"onewin","cookies":[{"name":"session-id","value":"different","domain":"127.0.0.1","path":"/"}]}`)
	id2 := startCasino(t, bin, env2, plain)
	t.Cleanup(func() { stopID(t, bin, env2, id2) })
	waitRunning(t, reg2, id2)
	waitEntry(t, reg2, id2, "ready")
	if sawShown(withheld) {
		t.Fatalf("withheld page showed balance: %+v", withheld.hitsCopy())
	}

	both := newStandIn(t, "session-id", secret)
	env3, reg3 := isolatedEnv(t)
	env3 = append(env3, "SUPERVISOR_PAGE_URL="+both.url)
	two := writeSession(t, `{
		"casino":"onewin",
		"cookies":[
			{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/"},
			{"name":"other","value":"second","domain":"127.0.0.1","path":"/"}
		]
	}`)
	id3 := startCasino(t, bin, env3, two)
	t.Cleanup(func() { stopID(t, bin, env3, id3) })
	waitRunning(t, reg3, id3)
	waitEntry(t, reg3, id3, "ready")
	hit := lastHit(t, both)
	if !hit.shown || !strings.Contains(hit.cookie, "session-id="+secret) || !strings.Contains(hit.cookie, "other=second") {
		t.Fatalf("two cookies: %+v", hit)
	}

	ordered := newStandIn(t, "session-id", secret)
	envOrder, regOrder := isolatedEnv(t)
	envOrder = append(envOrder, "SUPERVISOR_PAGE_URL="+ordered.url)
	cmdOrder := exec.Command(bin, "start", "--headless", file)
	cmdOrder.Env = envOrder
	outOrder, err := cmdOrder.CombinedOutput()
	if err != nil {
		t.Fatalf("flag-first start: %v %s", err, outOrder)
	}
	idOrder := strings.TrimSpace(string(outOrder))
	t.Cleanup(func() { stopID(t, bin, envOrder, idOrder) })
	waitRunning(t, regOrder, idOrder)
	waitEntry(t, regOrder, idOrder, "ready")
	if !sawShown(ordered) {
		t.Fatalf("flag-first start did not show balance: %+v", ordered.hitsCopy())
	}

	headedEnv, headedReg := isolatedEnv(t)
	headedEnv = append(headedEnv, "SUPERVISOR_PAGE_URL="+page.url)
	cmdHeaded := exec.Command(bin, "start", file)
	cmdHeaded.Env = headedEnv
	outHeaded, err := cmdHeaded.CombinedOutput()
	if err != nil {
		t.Fatalf("headed start: %v %s", err, outHeaded)
	}
	idHeaded := strings.TrimSpace(string(outHeaded))
	t.Cleanup(func() { stopID(t, bin, headedEnv, idHeaded) })
	headed := waitPresent(t, headedReg, idHeaded)
	if headed.Headless != nil && *headed.Headless {
		t.Fatalf("headed start recorded headless true")
	}

	blankEnv, blankReg := isolatedEnv(t)
	cmd := exec.Command(bin, "start", "--headless")
	cmd.Env = blankEnv
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("blank start: %v %s", err, out)
	}
	blankID := strings.TrimSpace(string(out))
	t.Cleanup(func() { stopID(t, bin, blankEnv, blankID) })
	blank := waitRunning(t, blankReg, blankID)
	if blank.Casino != "" {
		t.Fatalf("blank start stored casino %q", blank.Casino)
	}
}

func startCasino(t *testing.T, bin string, env []string, file string) string {
	t.Helper()
	cmd := exec.Command(bin, "start", file, "--headless")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("start: %v %s", err, out)
	}
	id := strings.TrimSpace(string(out))
	if id == "" {
		t.Fatal("missing id")
	}
	return id
}

func sawShown(page *standIn) bool {
	for _, hit := range page.hitsCopy() {
		if hit.shown {
			return true
		}
	}
	return false
}

func lastHit(t *testing.T, page *standIn) standInHit {
	t.Helper()
	hits := page.hitsCopy()
	if len(hits) == 0 {
		t.Fatal("stand-in received no request")
	}
	return hits[len(hits)-1]
}

func assertNoLiveCasino(t *testing.T, logPath string) {
	t.Helper()
	raw, err := os.ReadFile(logPath)
	if err != nil {
		return
	}
	if strings.Contains(string(raw), "1win.com") {
		t.Fatalf("worker log mentions 1win.com: %s", raw)
	}
}
