package integration

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/registry"
)

var sessionDateLine = regexp.MustCompile(`^date: \d{4}-\d{2}-\d{2} \d{2}:\d{2}$`)

func TestSessionInfoBlankStart(t *testing.T) {
	requireChrome(t)
	requireHeadlessShell(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })

	stdout, stderr, code := runSessionInfo(t, bin, env, id)
	if code != 0 {
		t.Fatalf("exit %d stderr %s stdout %s", code, stderr, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr %q", stderr)
	}
	assertSessionLines(t, stdout, "null", "null")
}

func TestSessionInfoOnewinStandIn(t *testing.T) {
	requireChrome(t)
	requireHeadlessShell(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	secret := "info-secret"
	page := newStandIn(t, "session-id", secret)
	env = append(env, "SUPERVISOR_PAGE_URL="+page.url)
	file := writeSession(t, `{
		"casino":"onewin",
		"cookies":[{"name":"session-id","value":"`+secret+`","domain":"127.0.0.1","path":"/","secure":false}]
	}`)
	id := startCasino(t, bin, env, file)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)

	stdout, stderr, code := runSessionInfo(t, bin, env, id)
	if code != 0 || stderr != "" {
		t.Fatalf("exit %d stderr %q stdout %q", code, stderr, stdout)
	}
	assertSessionLines(t, stdout, "onewin", "aviator")
}

func TestSessionInfoSeededRows(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	// starting + fresh CreatedAt survives registry.Apply without live PIDs (StartingTimeout).
	when := time.Now().UTC()
	doc := registry.Document{
		SchemaVersion: 1,
		Instances: []registry.Instance{
			{ID: "blank1", State: registry.StateStarting, CreatedAt: when, ProfileDir: t.TempDir()},
			{ID: "ow1", State: registry.StateStarting, CreatedAt: when, ProfileDir: t.TempDir(), Casino: "onewin"},
		},
	}
	writeRegistry(t, regPath, doc)
	wantDate := "date: " + when.Format("2006-01-02 15:04")

	stdout, stderr, code := runSessionInfo(t, bin, env, "blank1")
	if code != 0 || stderr != "" {
		t.Fatalf("blank exit %d stderr %q stdout %q", code, stderr, stdout)
	}
	if stdout != "casino: null\ngame: null\nstrategy: null\n"+wantDate+"\n" {
		t.Fatalf("blank stdout %q", stdout)
	}

	stdout, stderr, code = runSessionInfo(t, bin, env, "ow1")
	if code != 0 || stderr != "" {
		t.Fatalf("onewin exit %d stderr %q stdout %q", code, stderr, stdout)
	}
	if stdout != "casino: onewin\ngame: aviator\nstrategy: null\n"+wantDate+"\n" {
		t.Fatalf("onewin stdout %q", stdout)
	}
}

func TestSessionInfoRefuse(t *testing.T) {
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)

	for _, args := range [][]string{
		{"session", "info"},
		{"session", "info", "id", "extra"},
		{"session", "info", "   "},
	} {
		stdout, stderr, code := runSupervisor(t, bin, env, args...)
		if code != 2 {
			t.Fatalf("%v exit %d stdout %q stderr %q", args, code, stdout, stderr)
		}
		if stdout != "" {
			t.Fatalf("%v stdout %q", args, stdout)
		}
		if stderr == "" {
			t.Fatalf("%v empty stderr", args)
		}
	}

	stdout, stderr, code := runSupervisor(t, bin, env, "session", "info", "missing-id")
	if code != 1 {
		t.Fatalf("not found exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	if stdout != "" {
		t.Fatalf("not found stdout %q", stdout)
	}
	if !strings.Contains(stderr, "instance not found: missing-id") {
		t.Fatalf("stderr %q", stderr)
	}
}

func runSessionInfo(t *testing.T, bin string, env []string, id string) (stdout, stderr string, code int) {
	t.Helper()
	return runSupervisor(t, bin, env, "session", "info", id)
}

func runSupervisor(t *testing.T, bin string, env []string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	return out.String(), errb.String(), exitCode(err)
}

func assertSessionLines(t *testing.T, stdout, casino, game string) {
	t.Helper()
	if strings.Contains(stdout, "{") {
		t.Fatalf("json in stdout: %s", stdout)
	}
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("lines %d: %q", len(lines), stdout)
	}
	if lines[0] != "casino: "+casino || lines[1] != "game: "+game || lines[2] != "strategy: null" {
		t.Fatalf("fields: %q", stdout)
	}
	if !sessionDateLine.MatchString(lines[3]) {
		t.Fatalf("date line: %q", lines[3])
	}
}
