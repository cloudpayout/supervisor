package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOneWinUsage(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	secret := "usage-secret-value"
	file := writeSession(t, `{"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1"}]}`)

	for _, args := range [][]string{
		{"start", "--casino", "Onewin", "--session-file", file},
		{"start", "--casino", "1win.com", "--session-file", file},
		{"start", "--casino", "onewin"},
		{"start", "--session-file", file},
	} {
		cmd := exec.Command(bin, args...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if exitCode(err) != 2 {
			t.Fatalf("%v exit %d output %s", args, exitCode(err), out)
		}
		if strings.Contains(string(out), secret) {
			t.Fatalf("%v leaked the cookie value: %s", args, out)
		}
	}

	missing := exec.Command(bin, "start", "--casino", "onewin", "--session-file", filepath.Join(t.TempDir(), "missing.json"))
	missing.Env = env
	out, err := missing.CombinedOutput()
	if err == nil || strings.TrimSpace(string(out)) == "" && exitCode(err) == 0 {
		t.Fatalf("missing file exit %d output %s", exitCode(err), out)
	}
	if exitCode(err) == 0 || strings.Contains(string(out), secret) {
		t.Fatalf("missing file exit %d output %s", exitCode(err), out)
	}

	empty := writeSession(t, `{"cookies":[]}`)
	bad := exec.Command(bin, "start", "--casino", "onewin", "--session-file", empty)
	bad.Env = env
	out, err = bad.CombinedOutput()
	if exitCode(err) == 0 {
		t.Fatalf("empty cookies started: %s", out)
	}
	if _, err := os.Stat(regPath); err == nil {
		doc := readDoc(t, regPath)
		if len(doc.Instances) != 0 {
			t.Fatalf("registry has %+v", doc.Instances)
		}
	}
	entries, _ := os.ReadDir(profilesDir(env))
	if len(entries) != 0 {
		t.Fatalf("profiles left behind: %d", len(entries))
	}
}
