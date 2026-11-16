package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestStartRejectsHeaded(t *testing.T) {
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	file := writeSession(t, `{"casino":"onewin","cookies":[{"name":"a","value":"b"}]}`)
	cmd := exec.Command(bin, "start", "--headed", file)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if exitCode(err) != 2 {
		t.Fatalf("exit %v output %s", err, out)
	}
	if _, err := os.Stat(regPath); err == nil {
		t.Fatal("registry should not be created")
	}
}

func TestStartRecordsHeadless(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)

	blankEnv, blankReg := isolatedEnv(t)
	blankID := startArgs(t, bin, blankEnv, "start")
	t.Cleanup(func() { stopID(t, bin, blankEnv, blankID) })
	blank := waitPresent(t, blankReg, blankID)
	if blank.Headless != nil && *blank.Headless {
		t.Fatal("blank start recorded headless")
	}

	headlessEnv, headlessReg := isolatedEnv(t)
	headlessID := startArgs(t, bin, headlessEnv, "start", "--headless")
	t.Cleanup(func() { stopID(t, bin, headlessEnv, headlessID) })
	row := waitRunning(t, headlessReg, headlessID)
	if row.Headless == nil || !*row.Headless {
		t.Fatal("headless start did not record headless")
	}

	file := writeSession(t, `{"casino":"onewin","cookies":[{"name":"a","value":"b","domain":"127.0.0.1","path":"/"}]}`)
	page := newStandIn(t, "a", "b")
	fileEnv, fileReg := isolatedEnv(t)
	fileEnv = append(fileEnv, "SUPERVISOR_PAGE_URL="+page.url)
	fileID := startArgs(t, bin, fileEnv, "start", file)
	t.Cleanup(func() { stopID(t, bin, fileEnv, fileID) })
	fileRow := waitPresent(t, fileReg, fileID)
	if fileRow.Headless != nil && *fileRow.Headless {
		t.Fatal("session start recorded headless")
	}

	fileHeadlessEnv, fileHeadlessReg := isolatedEnv(t)
	fileHeadlessEnv = append(fileHeadlessEnv, "SUPERVISOR_PAGE_URL="+page.url)
	fileHeadlessID := startArgs(t, bin, fileHeadlessEnv, "start", file, "--headless")
	t.Cleanup(func() { stopID(t, bin, fileHeadlessEnv, fileHeadlessID) })
	fileHeadless := waitRunning(t, fileHeadlessReg, fileHeadlessID)
	if fileHeadless.Headless == nil || !*fileHeadless.Headless {
		t.Fatal("session headless start did not record headless")
	}
}

func startArgs(t *testing.T, bin string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v %s", args, err, out)
	}
	id := strings.TrimSpace(string(out))
	if id == "" {
		t.Fatal("missing id")
	}
	return id
}
