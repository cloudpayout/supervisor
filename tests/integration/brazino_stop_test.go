package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBrazinoStopRemovesProfile(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "brazino-stop-secret"
	page := newStandIn(t, "session-id", secret)
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{"casino":"brazino","cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/"}]}`)
	id := startCasino(t, bin, env, file)
	inst := waitRunning(t, regPath, id)
	assertStatusRunning(t, bin, env, id)
	work := t.TempDir()
	screenshotOnce(t, bin, env, work, id)
	assertStatusRunning(t, bin, env, id)
	stopID(t, bin, env, id)
	assertStatusGone(t, bin, env, id)
	doc := readDoc(t, regPath)
	for _, row := range doc.Instances {
		if row.ID == id {
			t.Fatal("instance still registered")
		}
	}
	if _, err := os.Stat(inst.ProfileDir); !os.IsNotExist(err) {
		t.Fatalf("profile remains: %v", err)
	}
}

func TestBrazinoFailedStartRemovesProfile(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	script := filepath.Join(t.TempDir(), "chrome")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	env = append(env, "CHROME_PATH="+script)
	file := writeSession(t, `{"casino":"brazino","cookies":[{"name":"session-id","value":"unused","domain":"127.0.0.1","path":"/"}]}`)
	cmd := exec.Command(bin, "start", file, "--headless")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if strings.Contains(string(out), "unused") {
		t.Fatalf("error includes the cookie value: %s", out)
	}
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		id := lines[len(lines)-1]
		stopID(t, bin, env, id)
	}
	deadline := time.Now().Add(20 * time.Second)
	for {
		left := 0
		if _, statErr := os.Stat(regPath); statErr == nil {
			left = len(readDoc(t, regPath).Instances)
		}
		entries, _ := os.ReadDir(profilesDir(env))
		if left == 0 && len(entries) == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("failed start left %d rows and %d profiles; start err=%v output=%s", left, len(entries), err, out)
		}
		time.Sleep(200 * time.Millisecond)
	}
}