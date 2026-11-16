package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/cli"
)

func TestResolveScreenshotOutput(t *testing.T) {
	dir := t.TempDir()
	shots := filepath.Join(dir, "shots")
	if err := os.Mkdir(shots, 0o755); err != nil {
		t.Fatal(err)
	}

	dest, printed, err := cli.ResolveScreenshotOutput(dir, " shots/page.png ")
	if err != nil {
		t.Fatal(err)
	}
	if printed != "shots/page.png" {
		t.Fatalf("printed %q", printed)
	}
	if dest != filepath.Join(dir, "shots", "page.png") {
		t.Fatalf("dest %q", dest)
	}

	abs := filepath.Join(dir, "abs.png")
	dest, printed, err = cli.ResolveScreenshotOutput(dir, abs)
	if err != nil {
		t.Fatal(err)
	}
	if printed != abs || dest != abs {
		t.Fatalf("printed %q dest %q", printed, dest)
	}

	if _, _, err := cli.ResolveScreenshotOutput(dir, "   "); err == nil || !strings.Contains(err.Error(), "output path required") {
		t.Fatalf("blank flag: %v", err)
	}

	existing := filepath.Join(shots, "taken.png")
	if err := os.WriteFile(existing, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cli.ResolveScreenshotOutput(dir, existing); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("existing file: %v", err)
	}
	if _, _, err := cli.ResolveScreenshotOutput(dir, shots); err == nil || !strings.Contains(err.Error(), "directory") {
		t.Fatalf("directory: %v", err)
	}
	if _, _, err := cli.ResolveScreenshotOutput(dir, "missing/page.png"); err == nil || !strings.Contains(err.Error(), "parent missing") {
		t.Fatalf("missing parent: %v", err)
	}
}

func TestDefaultScreenshotName(t *testing.T) {
	when := time.Date(2026, 11, 15, 21, 4, 5, 0, time.Local)
	name := cli.DefaultScreenshotName("abc", when)
	if name != "screenshot-abc-20261115-210405.png" {
		t.Fatalf("name %q", name)
	}
}
