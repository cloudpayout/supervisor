package unit

import (
	"strings"
	"testing"

	"github.com/cloudpayout/supervisor/internal/cli"
)

func TestWatchFPSValidation(t *testing.T) {
	for _, tc := range []struct {
		fps int
		ok  bool
	}{
		{1, true},
		{30, true},
		{0, false},
		{31, false},
	} {
		err := cli.ValidateWatchFPS(tc.fps)
		if tc.ok && err != nil {
			t.Fatalf("fps %d: %v", tc.fps, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("fps %d: want error", tc.fps)
		}
	}
}

func TestParseWatchFPSFlag(t *testing.T) {
	n, err := cli.ParseWatchFPSFlag("1")
	if err != nil || n != 1 {
		t.Fatalf("1: %d %v", n, err)
	}
	n, err = cli.ParseWatchFPSFlag("30")
	if err != nil || n != 30 {
		t.Fatalf("30: %d %v", n, err)
	}
	for _, raw := range []string{"abc", "0", "31", "1.5", ""} {
		if _, err := cli.ParseWatchFPSFlag(raw); err == nil || !strings.Contains(err.Error(), "invalid --fps") {
			t.Fatalf("%q: %v", raw, err)
		}
	}
}
