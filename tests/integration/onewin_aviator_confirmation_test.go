package integration

import (
	"path/filepath"
	"testing"
)

func TestAviatorConfirm(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "confirm-secret-value"
	page := newStandIn(t, "session-id", secret)
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{
		"casino":"onewin",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	waitRunning(t, regPath, id)
	waitEntry(t, regPath, id, "ready")
	assertStatusRunning(t, bin, env, id)
	work := t.TempDir()
	name := screenshotOnce(t, bin, env, work, id)
	if !pixelIs(t, filepath.Join(work, name), 255, 0, 0) {
		t.Fatal("screenshot is not the post-entry stand-in")
	}
	assertNoLiveCasino(t, waitEntry(t, regPath, id, "ready").LogPath)
	stopID(t, bin, env, id)
	assertStatusGone(t, bin, env, id)
}
