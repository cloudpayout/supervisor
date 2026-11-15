package unit

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func defaultChecksScript(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", ".github", "scripts", "default-checks-result.sh")
}

func runDefaultChecksResult(t *testing.T, body, status string) (int, string) {
	t.Helper()
	dir := t.TempDir()
	results := filepath.Join(dir, "test-results.json")
	if err := os.WriteFile(results, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", defaultChecksScript(t), results, status)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal(err)
	}
	return exitErr.ExitCode(), string(out)
}

func TestPackageSkipDoesNotFailDefaultChecks(t *testing.T) {
	body := `{"Time":"2026-11-14T07:50:00.030563787Z","Action":"skip","Package":"github.com/cloudpayout/supervisor/cmd/supervisor","Elapsed":0}
{"Time":"2026-11-14T07:50:00.030615418Z","Action":"skip","Package":"github.com/cloudpayout/supervisor/internal/browser","Elapsed":0}
{"Time":"2026-11-14T07:50:01.352907856Z","Action":"pass","Package":"github.com/cloudpayout/supervisor/tests/unit","Elapsed":0.132}
`
	code, out := runDefaultChecksResult(t, body, "0")
	if code != 0 {
		t.Fatalf("package skip exited %d, output %q", code, out)
	}
	if strings.Contains(out, "A check was skipped") {
		t.Fatalf("package skip reported as a skipped check: %q", out)
	}
}

func TestTestSkipFailsDefaultChecks(t *testing.T) {
	cases := []string{
		`{"Time":"2026-11-14T07:50:00Z","Action":"skip","Package":"github.com/cloudpayout/supervisor/tests/integration","Test":"TestParallelStartRegistry","Elapsed":0}` + "\n",
		`{"Test":"TestOther","Action":"skip","Package":"example.com/pkg"}` + "\n",
	}
	for _, body := range cases {
		code, out := runDefaultChecksResult(t, body, "0")
		if code != 1 {
			t.Fatalf("test skip exited %d, output %q, body %s", code, out, body)
		}
		if !strings.Contains(out, "A check was skipped") {
			t.Fatalf("test skip missing message, output %q, body %s", out, body)
		}
	}
}

func TestNonZeroGoTestStatusFailsDefaultChecks(t *testing.T) {
	body := `{"Time":"2026-11-14T07:50:00Z","Action":"fail","Package":"github.com/cloudpayout/supervisor/tests/unit","Test":"TestAggregateSelfAndMissingPID","Elapsed":0.1}
`
	code, out := runDefaultChecksResult(t, body, "1")
	if code != 1 {
		t.Fatalf("non-zero go test status exited %d, output %q", code, out)
	}
	if strings.Contains(out, "A check was skipped") {
		t.Fatalf("failed test reported as a skipped check: %q", out)
	}
}
