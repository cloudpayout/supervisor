package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

var (
	binOnce sync.Once
	binPath string
	binErr  error
)

func supervisorBin(t *testing.T) string {
	t.Helper()
	binOnce.Do(func() {
		dir, err := os.MkdirTemp("", "supervisor-bin")
		if err != nil {
			binErr = err
			return
		}
		binPath = filepath.Join(dir, "supervisor")
		cmd := exec.Command("go", "build", "-o", binPath, "./cmd/supervisor")
		cmd.Dir = moduleRoot(t)
		out, err := cmd.CombinedOutput()
		if err != nil {
			binErr = err
			binPath = string(out)
		}
	})
	if binErr != nil {
		t.Fatalf("build supervisor: %v %s", binErr, binPath)
	}
	return binPath
}

func isolatedEnv(t *testing.T) (env []string, stateDir string) {
	t.Helper()
	root := t.TempDir()
	stateDir = filepath.Join(root, "state")
	dataDir := filepath.Join(root, "data")
	env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
		"XDG_STATE_HOME=" + stateDir,
		"XDG_DATA_HOME=" + dataDir,
	}
	return env, filepath.Join(stateDir, "supervisor", "instances.json")
}

func chromeAvailable() bool {
	if p := os.Getenv("CHROME_PATH"); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return true
		}
		return false
	}
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "chrome"} {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
}

func requireChrome(t *testing.T) {
	t.Helper()
	if !chromeAvailable() {
		t.Skip("chrome not installed")
	}
}
