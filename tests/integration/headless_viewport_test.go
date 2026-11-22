package integration

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudpayout/supervisor/internal/browser"
)

func TestHeadlessViewportScreenshot(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)

	work := t.TempDir()
	name := runScreenshot(t, bin, env, work, id)
	assertPNGSize(t, filepath.Join(work, name), browser.HeadlessViewportWidth, browser.HeadlessViewportHeight)
}

func TestHeadlessViewportTwoInstances(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	a := startHeadless(t, bin, env)
	b := startHeadless(t, bin, env)
	t.Cleanup(func() {
		stopID(t, bin, env, a)
		stopID(t, bin, env, b)
	})
	waitRunning(t, regPath, a)
	waitRunning(t, regPath, b)

	work := t.TempDir()
	for _, id := range []string{a, b} {
		name := runScreenshot(t, bin, env, work, id)
		assertPNGSize(t, filepath.Join(work, name), browser.HeadlessViewportWidth, browser.HeadlessViewportHeight)
	}
}

func TestHeadedStartRecordsHeaded(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startArgs(t, bin, env, "start")
	t.Cleanup(func() { stopID(t, bin, env, id) })
	row := waitPresent(t, regPath, id)
	if row.Headless != nil && *row.Headless {
		t.Fatal("headed start recorded headless")
	}
}

func assertPNGSize(t *testing.T, path string, wantW, wantH int) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	if b.Dx() != wantW || b.Dy() != wantH {
		t.Fatalf("%s size %dx%d, want %dx%d", path, b.Dx(), b.Dy(), wantW, wantH)
	}
}