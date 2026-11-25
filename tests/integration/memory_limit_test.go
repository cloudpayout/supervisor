package integration

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestMemoryLimitParsingOnStart(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	cmd := exec.Command(bin, "start", "--headless", "--memory-limit", "512MiB")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("start: %v %s", err, out)
	}
	id := strings.TrimSpace(string(out))
	t.Cleanup(func() { stopID(t, bin, env, id) })
	doc := readDoc(t, regPath)
	var found bool
	for _, inst := range doc.Instances {
		if inst.ID == id {
			found = true
			if inst.MemoryLimitBytes != 512*1024*1024 {
				t.Fatalf("limit %d", inst.MemoryLimitBytes)
			}
		}
	}
	if !found {
		t.Fatal("row missing")
	}
	deadline := time.Now().Add(45 * time.Second)
	re := regexp.MustCompile(`\d+\.\d+ MiB/512\.0 MiB`)
	for {
		status := runStatusOnce(t, bin, env)
		if strings.Contains(status, id) && re.MatchString(status) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("MEMORY format: %s", status)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func TestMemoryLimitEquivalentSuffixes(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	a := startWith(t, bin, env, "start", "--headless", "--memory-limit", "512M")
	b := startWith(t, bin, env, "start", "--headless", "--memory-limit", "512MiB")
	t.Cleanup(func() {
		stopID(t, bin, env, a)
		stopID(t, bin, env, b)
	})
	deadline := time.Now().Add(60 * time.Second)
	for {
		status := runStatusOnce(t, bin, env)
		if strings.Count(status, "512.0 MiB") >= 2 && strings.Contains(status, a) && strings.Contains(status, b) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("denominators: %s", status)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func startWith(t *testing.T, bin string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}
