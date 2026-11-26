package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrazinoScreenshotShowsBalance(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "brazino-shot-secret"
	page := newStandIn(t, "session-id", secret)
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{"casino":"brazino","cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/"}]}`)
	id := startCasino(t, bin, env, file)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	before := waitRunning(t, regPath, id)
	waitRunning(t, regPath, id)
	assertStatusRunning(t, bin, env, id)

	work := t.TempDir()
	name := screenshotOnce(t, bin, env, work, id)
	if strings.Contains(name, secret) {
		t.Fatalf("filename contains the cookie value: %s", name)
	}
	if !pixelIs(t, filepath.Join(work, name), 255, 0, 0) {
		t.Fatal("shown page pixel is not the balance color")
	}
	after := waitRunning(t, regPath, id)
	if after.OwnerPID != before.OwnerPID || after.BrowserPID != before.BrowserPID {
		t.Fatal("screenshot stopped the instance")
	}
	if _, err := os.Stat(before.ProfileDir); err != nil {
		t.Fatal("profile was removed by screenshot")
	}
	assertStatusRunning(t, bin, env, id)
}