package browser_test

import (
	"context"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"

	"github.com/cloudpayout/supervisor/internal/browser"
)

func TestAllocatorOptionsHeadlessWindowSize(t *testing.T) {
	t.Parallel()

	headlessArgs := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true))
	if !hasFlagValue(headlessArgs, "window-size", "1280,720") {
		t.Fatalf("headless args missing 1280,720 window-size: %v", headlessArgs)
	}

	headedArgs := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), false))
	if hasFlagValue(headedArgs, "window-size", "1280,720") {
		t.Fatalf("headed args must not force headless window-size: %v", headedArgs)
	}
	if browser.HeadlessViewportWidth != 1280 || browser.HeadlessViewportHeight != 720 {
		t.Fatalf("constants %dx%d", browser.HeadlessViewportWidth, browser.HeadlessViewportHeight)
	}
}

func chromeAllocatorArgs(t *testing.T, opts []chromedp.ExecAllocatorOption) []string {
	t.Helper()

	var args []string
	opts = append(opts,
		chromedp.ExecPath("/do-not-run-chrome"),
		chromedp.ModifyCmdFunc(func(cmd *exec.Cmd) { args = slices.Clone(cmd.Args[1:]) }),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	if err := chromedp.Do(ctx); err == nil {
		t.Fatal("expected chrome launch to fail for bogus binary")
	}
	if args == nil {
		t.Fatal("allocator did not build command")
	}
	return slices.DeleteFunc(args, func(a string) bool {
		return strings.HasPrefix(a, "--remote-debugging-io-pipes=")
	})
}

func hasFlagValue(args []string, name, value string) bool {
	prefix := "--" + name + "="
	want := prefix + value
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}