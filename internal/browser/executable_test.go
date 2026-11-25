package browser_test

import (
	"strings"
	"testing"

	"github.com/cloudpayout/supervisor/internal/browser"
)

func TestHeadlessShellPath(t *testing.T) {
	t.Parallel()
	p, err := browser.HeadlessShellPath()
	if err != nil {
		t.Skip(err)
	}
	if !strings.HasSuffix(p, "chrome-headless-shell") {
		t.Fatalf("path = %s", p)
	}
}
