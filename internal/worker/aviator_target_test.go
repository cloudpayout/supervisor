package worker

import "testing"

func TestAviatorTargetURL(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"", false},
		{"about:blank", false},
		{"https://aviaport.spribegaming.com/aviator?currency=MXN", true},
		{"https://launch.spribelaunch.com/aviator/foo", true},
		{"https://1win.com/resources/game/aviator.html", true},
		{"https://1win.com/es-MX/casino/play/v_spribe:aviator", false},
	}
	for _, tc := range cases {
		if got := aviatorTargetURL(tc.url); got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.url, got, tc.want)
		}
	}
}