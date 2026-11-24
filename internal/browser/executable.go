package browser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Executable returns the full Chrome binary for headed starts (CHROME_PATH, then PATH).
func Executable() (string, error) {
	if p := os.Getenv("CHROME_PATH"); p != "" {
		st, err := os.Stat(p)
		if err != nil {
			return "", fmt.Errorf("CHROME_PATH %q: %w", p, err)
		}
		if st.IsDir() {
			return "", fmt.Errorf("CHROME_PATH %q is a directory", p)
		}
		return p, nil
	}
	names := []string{
		"google-chrome",
		"google-chrome-stable",
		"chromium",
		"chromium-browser",
		"chrome",
	}
	for _, name := range names {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no Chrome or Chromium binary on PATH")
}

// HeadlessShellPath returns the project chrome-headless-shell (spec 017).
// CHROME_PATH is not consulted.
func HeadlessShellPath() (string, error) {
	root, err := moduleRoot()
	if err != nil {
		return "", err
	}
	p := filepath.Join(root, "bin", "chrome-headless-shell")
	st, err := os.Stat(p)
	if err != nil {
		return "", fmt.Errorf("chrome-headless-shell missing at %s (run scripts/fetch-chrome-headless-shell.sh): %w", p, err)
	}
	if st.IsDir() {
		return "", fmt.Errorf("chrome-headless-shell path %s is a directory", p)
	}
	if st.Mode()&0o111 == 0 {
		return "", fmt.Errorf("chrome-headless-shell at %s is not executable", p)
	}
	return p, nil
}

// Preflight fails when headed Chrome cannot be resolved.
func Preflight() error {
	if _, err := Executable(); err != nil {
		return fmt.Errorf("chrome is not available: %w", err)
	}
	return nil
}

// PreflightHeadless fails when the bundled headless shell is missing.
func PreflightHeadless() error {
	if _, err := HeadlessShellPath(); err != nil {
		return fmt.Errorf("chrome-headless-shell is not available: %w", err)
	}
	return nil
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if exe, err := os.Executable(); err == nil {
		dir = filepath.Dir(exe)
		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", fmt.Errorf("could not locate module root for chrome-headless-shell")
}
