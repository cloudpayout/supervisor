package browser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/chromedp/chromedp"
)

// Headless viewport dimensions (spec 012-headless-mode-viewport).
const (
	HeadlessViewportWidth  = 1280
	HeadlessViewportHeight = 720
)

// Preflight fails when Chrome cannot be resolved. It does not touch the registry (FR-011).
func Preflight() error {
	if _, err := Executable(); err != nil {
		return fmt.Errorf("chrome is not available: %w", err)
	}
	return nil
}

// Executable returns the Chrome binary path.
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

// AllocatorOptions builds a dedicated ExecAllocator for one instance.
func AllocatorOptions(profileDir string, headless bool) []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(profileDir),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)
	if headless {
		opts = append(opts, chromedp.WindowSize(HeadlessViewportWidth, HeadlessViewportHeight))
	} else {
		opts = append(opts, chromedp.VisibleWindow)
	}
	// Live casino embeds Spribe in cross-site iframes; without this, the top
	// document cannot read frame DOM and entry ready probes never succeed.
	opts = append(opts,
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"),
		chromedp.Flag("disable-site-isolation-trials", true),
	)
	if p, err := Executable(); err == nil {
		opts = append(opts, chromedp.ExecPath(p))
	}
	return opts
}

// EnsureDirs creates the profile directory.
func EnsureDirs(profileDir string) error {
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return fmt.Errorf("create profile dir: %w", err)
	}
	return nil
}

// RemoveProfile removes a profile directory created for a failed start.
func RemoveProfile(profileDir string) {
	if profileDir == "" {
		return
	}
	_ = os.RemoveAll(profileDir)
}

// LogFile opens the worker log path, creating parent directories.
func LogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}
