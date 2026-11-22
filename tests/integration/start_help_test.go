package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestStartHelp(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	file := writeSession(t, `{"casino":"onewin","cookies":[{"name":"a","value":"b"}]}`)

	for _, args := range [][]string{
		{"start", "--help"},
		{"start", file, "--help"},
		{"start", "--headless", "--help"},
		{"start", file, "--headless", "--help"},
	} {
		cmd := exec.Command(bin, args...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v %s", args, err, out)
		}
		text := string(out)
		if !strings.Contains(text, "path") || !strings.Contains(text, "blank") || !strings.Contains(text, "--headless") {
			t.Fatalf("help missing expected text: %s", text)
		}
		if strings.Contains(text, "--headless") && (!strings.Contains(text, "1280") || !strings.Contains(text, "720")) {
			t.Fatalf("help missing headless viewport dimensions: %s", text)
		}
		for _, banned := range []string{"--session-file", "--headed", "--casino"} {
			if strings.Contains(text, banned) {
				t.Fatalf("help lists %s: %s", banned, text)
			}
		}
	}
	if _, err := os.Stat(regPath); err == nil {
		t.Fatal("help created a registry")
	}

	bad := exec.Command(bin, "start", "-h")
	bad.Env = env
	out, err := bad.CombinedOutput()
	if exitCode(err) != 2 {
		t.Fatalf("-h exit %v output %s", err, out)
	}
	if _, err := os.Stat(regPath); err == nil {
		doc := readDoc(t, regPath)
		if len(doc.Instances) != 0 {
			t.Fatalf("-h left instances: %+v", doc.Instances)
		}
	}
}
