package unit

import (
	"strings"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/cli"
	"github.com/cloudpayout/supervisor/internal/registry"
)

func TestFormatSessionInfoBlank(t *testing.T) {
	got := cli.FormatSessionInfo(registry.Instance{
		CreatedAt: time.Date(2026, 11, 26, 9, 15, 45, 0, time.UTC),
	})
	want := "casino: null\ngame: null\nstrategy: null\ndate: 2026-11-26 09:15\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatSessionInfoOnewin(t *testing.T) {
	got := cli.FormatSessionInfo(registry.Instance{
		Casino:    "onewin",
		CreatedAt: time.Date(2026, 11, 26, 16, 5, 59, 0, time.FixedZone("X", -7*3600)),
	})
	if !strings.Contains(got, "casino: onewin\n") || !strings.Contains(got, "game: aviator\n") || !strings.Contains(got, "strategy: null\n") {
		t.Fatalf("fields: %q", got)
	}
	// 16:05 UTC-7 is 23:05 UTC.
	if !strings.Contains(got, "date: 2026-11-26 23:05\n") {
		t.Fatalf("date: %q", got)
	}
}
