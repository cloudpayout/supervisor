package unit

import (
	"testing"

	"github.com/cloudpayout/supervisor/internal/memlimit"
)

func TestParseLimit(t *testing.T) {
	a, err := memlimit.ParseLimit("512M")
	if err != nil {
		t.Fatal(err)
	}
	b, err := memlimit.ParseLimit("512MiB")
	if err != nil {
		t.Fatal(err)
	}
	if a != b || a != 512*1024*1024 {
		t.Fatalf("512M=%d 512MiB=%d", a, b)
	}
	g, err := memlimit.ParseLimit("1G")
	if err != nil {
		t.Fatal(err)
	}
	if g != 1024*1024*1024 {
		t.Fatalf("1G=%d", g)
	}
	for _, bad := range []string{"512MB", "512", "0M", "512GB", ""} {
		if _, err := memlimit.ParseLimit(bad); err == nil {
			t.Fatalf("expected reject %q", bad)
		}
	}
}

func TestFormatMiB(t *testing.T) {
	got := memlimit.FormatMiBOneDecimal(512 * 1024 * 1024)
	if got != "512.0 MiB" {
		t.Fatalf("got %q", got)
	}
}
