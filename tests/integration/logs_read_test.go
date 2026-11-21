package integration

import (
	"strings"
	"testing"
	"time"
)

func TestLogsReadStructured(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, _ := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })

	deadline := time.Now().Add(20 * time.Second)
	var out cmdResult
	for time.Now().Before(deadline) {
		cmd := execCommand(bin, env, "logs", id)
		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		out = cmdResult{stdout: stdout.String(), stderr: stderr.String(), code: exitCode(err)}
		if out.code == 0 && strings.Contains(out.stdout, " NOTICE ") {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if out.code != 0 || !strings.Contains(out.stdout, " NOTICE ") {
		t.Fatalf("logs: code %d stdout %q stderr %q", out.code, out.stdout, out.stderr)
	}
	help := execCommand(bin, env, "help")
	helpOut, err := help.Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(helpOut), "logs") {
		t.Fatalf("help missing logs:\n%s", helpOut)
	}
}
