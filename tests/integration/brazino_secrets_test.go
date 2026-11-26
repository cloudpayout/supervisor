package integration

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrazinoSecretStaysOffTheRecord(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "brazino-record-secret"
	page := newStandIn(t, "session-id", secret)
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{"casino":"brazino","cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/"}]}`)
	startCmd := exec.Command(bin, "start", file, "--headless")
	startCmd.Env = env
	startOut, err := startCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("start: %v %s", err, startOut)
	}
	assertNoSecret(t, secret, startOut, "operator start output")
	id := strings.TrimSpace(string(startOut))
	t.Cleanup(func() { stopID(t, bin, env, id) })
	inst := waitRunning(t, regPath, id)
	waitRunning(t, regPath, id)

	workerArgv := cmdline(t, inst.OwnerPID)
	if !bytes.Contains(workerArgv, []byte("--session-file")) || !bytes.Contains(workerArgv, []byte(file)) {
		t.Fatalf("worker argv missing the session path: %q", workerArgv)
	}
	assertNoSecret(t, secret, workerArgv, "worker cmdline")
	log, _ := os.ReadFile(inst.LogPath)
	assertNoSecret(t, secret, log, "worker log")
	raw, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}
	assertNoSecret(t, secret, raw, "registry")
	if !bytes.Contains(raw, []byte(`"casino":"brazino"`)) && !bytes.Contains(raw, []byte(`"casino": "brazino"`)) {
		t.Fatalf("registry missing casino: %s", raw)
	}
	if bytes.Contains(raw, []byte("brazino777")) || bytes.Contains(raw, []byte(file)) {
		t.Fatalf("registry stored page or path: %s", raw)
	}

	work := t.TempDir()
	cmd := exec.Command(bin, "screenshot", id)
	cmd.Env = env
	cmd.Dir = work
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	assertNoSecret(t, secret, cmdline(t, cmd.Process.Pid), "screenshot cmdline")
	out, err := io.ReadAll(stdout)
	if waitErr := cmd.Wait(); err != nil || waitErr != nil {
		t.Fatalf("screenshot: %v %v", err, waitErr)
	}
	name := strings.TrimSpace(string(out))
	if strings.Contains(name, secret) {
		t.Fatalf("filename contains the cookie value: %s", name)
	}
	if _, err := os.Stat(filepath.Join(work, name)); err != nil {
		t.Fatal(err)
	}
}