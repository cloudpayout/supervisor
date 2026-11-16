package integration

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestOneWinSecretStaysOffTheRecord(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	secret := "record-secret-value"
	page := newStandIn(t, "session-id", secret)
	env, regPath := isolatedEnv(t)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{"casino":"onewin","cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/"}]}`)
	startCmd := exec.Command(bin, "start", file, "--headless")
	startCmd.Env = env
	var startOut bytes.Buffer
	startCmd.Stdout = &startOut
	startCmd.Stderr = &startOut
	if err := startCmd.Start(); err != nil {
		t.Fatal(err)
	}
	// /proc/pid/cmdline is empty between fork and exec. Poll briefly, then fail
	// if the positional path never appears (a real argv regression).
	operatorArgv := waitCmdlineContains(t, startCmd.Process.Pid, file, 500*time.Millisecond)
	if bytes.Contains(operatorArgv, []byte("--session-file")) {
		t.Fatalf("operator start argv includes --session-file: %q", operatorArgv)
	}
	assertNoSecret(t, secret, operatorArgv, "operator start argv")
	if err := startCmd.Wait(); err != nil {
		t.Fatalf("start: %v %s", err, startOut.Bytes())
	}
	id := strings.TrimSpace(startOut.String())
	if id == "" {
		t.Fatal("missing id")
	}
	t.Cleanup(func() { stopID(t, bin, env, id) })
	inst := waitRunning(t, regPath, id)

	assertNoSecret(t, secret, cmdline(t, inst.OwnerPID), "worker cmdline")
	assertNoSecret(t, secret, cmdline(t, inst.BrowserPID), "browser cmdline")
	log, _ := os.ReadFile(inst.LogPath)
	assertNoSecret(t, secret, log, "worker log")
	raw, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}
	assertNoSecret(t, secret, raw, "registry")
	if !bytes.Contains(raw, []byte(`"casino":"onewin"`)) && !bytes.Contains(raw, []byte(`"casino": "onewin"`)) {
		t.Fatalf("registry missing casino: %s", raw)
	}
	if bytes.Contains(raw, []byte("session-file")) || bytes.Contains(raw, []byte("1win.com")) || bytes.Contains(raw, []byte(file)) {
		t.Fatalf("registry stored a secret path or page: %s", raw)
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

func cmdline(t *testing.T, pid int) []byte {
	t.Helper()
	raw, err := os.ReadFile("/proc/" + itoa(pid) + "/cmdline")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func waitCmdlineContains(t *testing.T, pid int, needle string, budget time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(budget)
	var last []byte
	for {
		raw, err := os.ReadFile("/proc/" + itoa(pid) + "/cmdline")
		if err == nil {
			last = raw
			if len(raw) > 0 && (bytes.Contains(raw, []byte(needle)) || bytes.Contains(raw, []byte(filepath.Base(needle)))) {
				return raw
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("operator start argv missing the positional path: %q", last)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func assertNoSecret(t *testing.T, secret string, raw []byte, where string) {
	t.Helper()
	if bytes.Contains(raw, []byte(secret)) {
		t.Fatalf("%s contains the cookie value", where)
	}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
