package integration

import (
	"os/exec"
	"strings"
	"testing"
)

func TestWatchHelp(t *testing.T) {
	bin := supervisorBin(t)
	cmd := exec.Command(bin, "watch", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help: %v %s", err, out)
	}
	text := string(out)
	for _, want := range []string{
		"instance-id",
		"--fps",
		"1 frame per second",
		"30",
		"URL",
		"Interrupt",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}
