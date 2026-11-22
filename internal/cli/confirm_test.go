package cli

import "testing"

func TestConfirmYes(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"y", true},
		{"Y", true},
		{"yes", true},
		{"YES", true},
		{"  Yes  ", true},
		{"", false},
		{"n", false},
		{"no", false},
		{"yeah", false},
		{" y es ", false},
	}
	for _, tc := range cases {
		if got := confirmYes(tc.line); got != tc.want {
			t.Errorf("confirmYes(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}
