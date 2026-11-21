package integration

import (
	"strings"
	"testing"
	"time"
)

func TestAviatorPreferencesReachReady(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)

	t.Run("defaults", func(t *testing.T) {
		page := newStandIn(t, "session-id", "pref-secret")
		env, regPath := isolatedEnv(t)
		env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
		file := writeSession(t, `{"casino":"onewin","cookies":[{"name":"session-id","value":"pref-secret","domain":"127.0.0.1","path":"/","secure":false}]}`)
		id := startCasino(t, bin, env, file)
		waitRunning(t, regPath, id)
		waitEntry(t, regPath, id, "ready")
		waitStandInEvent(t, page, "currency", "USD")
		waitStandInEvent(t, page, "play", "")
		for _, path := range page.pathsCopy() {
			if strings.Contains(path, "es-MX") {
				t.Fatalf("default entry used a locale path: %s", path)
			}
		}
		stopID(t, bin, env, id)
	})

	t.Run("demo", func(t *testing.T) {
		page := newStandIn(t, "session-id", "pref-secret")
		env, regPath := isolatedEnv(t)
		env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
		file := writeSession(t, `{"casino":"onewin","demo":true,"cookies":[{"name":"session-id","value":"pref-secret","domain":"127.0.0.1","path":"/","secure":false}]}`)
		id := startCasino(t, bin, env, file)
		waitRunning(t, regPath, id)
		waitEntry(t, regPath, id, "ready")
		waitStandInEvent(t, page, "demo", "")
		for _, ev := range page.eventsCopy() {
			if ev.kind == "play" {
				t.Fatal("demo entry also activated real play")
			}
		}
		stopID(t, bin, env, id)
	})

	t.Run("locale", func(t *testing.T) {
		page := newStandIn(t, "session-id", "pref-secret")
		env, regPath := isolatedEnv(t)
		env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
		file := writeSession(t, `{"casino":"onewin","locale":"es-MX","currency":"MXN","cookies":[{"name":"session-id","value":"pref-secret","domain":"127.0.0.1","path":"/","secure":false}]}`)
		id := startCasino(t, bin, env, file)
		waitRunning(t, regPath, id)
		waitEntry(t, regPath, id, "ready")
		waitStandInEvent(t, page, "currency", "MXN")
		found := false
		for _, path := range page.pathsCopy() {
			if strings.Contains(path, "es-MX") {
				found = true
			}
		}
		if !found {
			t.Fatalf("locale path missing es-MX: %v", page.pathsCopy())
		}
		stopID(t, bin, env, id)
	})
}

func waitStandInEvent(t *testing.T, page *standIn, kind, value string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, ev := range page.eventsCopy() {
			if ev.kind == kind && (value == "" || ev.value == value) {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("stand-in missing %s %q: %+v", kind, value, page.eventsCopy())
}
