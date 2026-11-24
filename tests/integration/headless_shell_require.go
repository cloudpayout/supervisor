package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func headlessShellPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(moduleRoot(t), "bin", "chrome-headless-shell")
}

func requireHeadlessShell(t *testing.T) {
	t.Helper()
	st, err := os.Stat(headlessShellPath(t))
	if err != nil || st.IsDir() {
		t.Skip("chrome-headless-shell missing; run scripts/fetch-chrome-headless-shell.sh")
	}
}
