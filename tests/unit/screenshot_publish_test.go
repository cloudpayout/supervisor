package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudpayout/supervisor/internal/cli"
)

func TestPublishPNGLeavesExistingFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "page.png")
	if err := os.WriteFile(dest, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := cli.PublishPNG(dest, []byte{0x89, 'P', 'N', 'G', 1, 2, 3})
	if err == nil {
		t.Fatal("expected link failure")
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep" {
		t.Fatalf("dest = %q", got)
	}
	partials, err := filepath.Glob(filepath.Join(dir, ".screenshot-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(partials) != 0 {
		t.Fatalf("partial files left: %v", partials)
	}
}

func TestPublishPNGRemovesTempOnSuccess(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "page.png")
	png := []byte{0x89, 'P', 'N', 'G', 1, 2, 3}
	if err := cli.PublishPNG(dest, png); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(png) {
		t.Fatalf("dest = %v", got)
	}
	partials, err := filepath.Glob(filepath.Join(dir, ".screenshot-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(partials) != 0 {
		t.Fatalf("partial files left: %v", partials)
	}
}
