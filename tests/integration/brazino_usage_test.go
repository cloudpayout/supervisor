package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrazinoUsage(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	secret := "brazino-usage-secret"
	file := writeSession(t, `{"casino":"brazino","cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1"}]}`)

	for _, args := range [][]string{
		{"start", "--session-file", file},
		{"start", file, "extra"},
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

	missingPath := filepath.Join(t.TempDir(), "missing.json")
	missing := exec.Command(bin, "start", missingPath)
	missing.Env = env
	out, err := missing.CombinedOutput()
	if err == nil || exitCode(err) == 0 {
		t.Fatalf("missing file exit %d output %s", exitCode(err), out)
	}

	badCasino := writeSession(t, `{"casino":"Brazino","cookies":[{"name":"session-id","value":"`+secret+`"}]}`)
	unknown := writeSession(t, `{"casino":"brazino777","cookies":[{"name":"session-id","value":"`+secret+`"}]}`)
	for _, path := range []string{badCasino, unknown} {
		cmd := exec.Command(bin, "start", path)
		cmd.Env = env
		out, err = cmd.CombinedOutput()
		if err == nil || exitCode(err) == 0 {
			t.Fatalf("%s started: %s", path, out)
		}
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