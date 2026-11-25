package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScreenshotWhileWatching(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)

	watch := execCommand(bin, env, "watch", id)
	wstdout, err := watch.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := watch.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := bufioReadLine(wstdout); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = watch.Process.Kill()
		_ = watch.Wait()
	}()

	work := t.TempDir()
	cmd := exec.Command(bin, "screenshot", id)
	cmd.Env = env
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	if exitCode(err) != 0 {
		t.Fatalf("screenshot: %v %s", err, out)
	}
	name := strings.TrimSpace(string(out))
	if !strings.HasSuffix(name, ".png") {
		t.Fatalf("name %q", name)
	}
	if _, err := os.Stat(filepath.Join(work, name)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if watch.ProcessState != nil && watch.ProcessState.Exited() {
		t.Fatal("watch exited during screenshot")
	}
}