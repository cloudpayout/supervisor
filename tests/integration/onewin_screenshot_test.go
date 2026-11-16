package integration

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestOneWinScreenshotShowsBalance(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "shot-secret-value"
	page := newStandIn(t, "session-id", secret)
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{"casino":"onewin","cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/"}]}`)
	id := startCasino(t, bin, env, file)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	before := waitRunning(t, regPath, id)
	assertStatusRunning(t, bin, env, id)

	work := t.TempDir()
	name := screenshotOnce(t, bin, env, work, id)
	if strings.Contains(name, secret) {
		t.Fatalf("filename contains the cookie value: %s", name)
	}
	if !pixelIs(t, filepath.Join(work, name), 255, 0, 0) {
		t.Fatal("shown page pixel is not the balance color")
	}
	if !sawShown(page) {
		t.Fatal("shown page did not include balance")
	}
	after := waitRunning(t, regPath, id)
	if after.OwnerPID != before.OwnerPID || after.BrowserPID != before.BrowserPID {
		t.Fatal("screenshot stopped the instance")
	}
	if _, err := os.Stat(before.ProfileDir); err != nil {
		t.Fatal("profile was removed by screenshot")
	}
	assertStatusRunning(t, bin, env, id)
	stopID(t, bin, env, id)
	assertStatusGone(t, bin, env, id)

	withheld := newStandIn(t, "session-id", "not-this")
	env2, reg2 := isolatedEnv(t)
	env2 = append(env2, "SUPERVISOR_PAGE_URL="+withheld.url)
	plain := writeSession(t, `{"casino":"onewin","cookies":[{"name":"session-id","value":"different","domain":"127.0.0.1","path":"/"}]}`)
	id2 := startCasino(t, bin, env2, plain)
	t.Cleanup(func() { stopID(t, bin, env2, id2) })
	waitRunning(t, reg2, id2)
	assertStatusRunning(t, bin, env2, id2)
	work2 := t.TempDir()
	name2 := screenshotOnce(t, bin, env2, work2, id2)
	if !pixelIs(t, filepath.Join(work2, name2), 0, 0, 255) {
		t.Fatal("withheld page pixel is not the withheld color")
	}
	if sawShown(withheld) {
		t.Fatal("withheld page included balance")
	}
	assertStatusRunning(t, bin, env2, id2)
	stopID(t, bin, env2, id2)
	assertStatusGone(t, bin, env2, id2)
}

func screenshotOnce(t *testing.T, bin string, env []string, dir, id string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "screenshot", id)
	cmd.Env = env
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("screenshot: %v", err)
	}
	name := strings.TrimSpace(string(out))
	pattern := regexp.MustCompile(`^screenshot-` + regexp.QuoteMeta(id) + `-\d{8}-\d{6}(-\d+)?\.png$`)
	if !pattern.MatchString(name) {
		t.Fatalf("stdout %q", name)
	}
	return name
}

func pixelIs(t *testing.T, path string, r, g, b uint32) bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("%s is not a PNG", path)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	bounds := img.Bounds()
	c := img.At(bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+bounds.Dy()/2)
	rr, gg, bb, _ := c.RGBA()
	return rr>>8 == r && gg>>8 == g && bb>>8 == b
}
