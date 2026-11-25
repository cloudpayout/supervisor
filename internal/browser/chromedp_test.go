package browser_test

import (
	"context"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/chromedp/chromedp"

	"github.com/cloudpayout/supervisor/internal/browser"
)

// chromeProbeDisplayMu serializes the dummy DISPLAY used by chromeAllocatorArgs.
// Headed AllocatorOptions enables chromedp.VisibleWindow, which returns
// ErrNoDisplay on Linux before building a command when DISPLAY and
// WAYLAND_DISPLAY are both unset (GitHub Actions). Callers use t.Parallel, so
// t.Setenv is not available.
var chromeProbeDisplayMu sync.Mutex

func TestAllocatorOptionsHeadlessWindowSize(t *testing.T) {
	t.Parallel()

	headlessArgs := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, true))
	if !hasFlagValue(headlessArgs, "window-size", "1280,720") {
		t.Fatalf("headless args missing 1280,720 window-size: %v", headlessArgs)
	}
	if !hasFlagValue(headlessArgs, "headless", "new") {
		t.Fatalf("headless args missing headless=new: %v", headlessArgs)
	}
	if hasFlag(headlessArgs, "enable-automation") {
		t.Fatalf("headless args must not enable automation: %v", headlessArgs)
	}
	for _, name := range []string{
		"disable-background-networking",
		"disable-background-timer-throttling",
		"disable-backgrounding-occluded-windows",
		"disable-renderer-backgrounding",
		"disable-remote-fonts",
		"disable-reading-from-canvas",
		"safebrowsing-disable-auto-update",
		"enable-features",
	} {
		if hasFlag(headlessArgs, name) {
			t.Fatalf("headless args must not pass --%s (Spribe config): %v", name, headlessArgs)
		}
	}
	if !hasFlagValue(headlessArgs, "user-agent", browser.DefaultHeadlessUserAgent) {
		t.Fatalf("headless args missing desktop user-agent: %v", headlessArgs)
	}
	if !strings.Contains(browser.DefaultHeadlessUserAgent, "Chrome/") {
		t.Fatal("default headless UA should look like desktop Chrome")
	}

	headedArgs := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), false, true))
	if hasFlagValue(headedArgs, "window-size", "1280,720") {
		t.Fatalf("headed args must not force headless window-size: %v", headedArgs)
	}
	if !hasFlag(headedArgs, "disable-background-networking") {
		t.Fatalf("headed args keep chromedp background defaults: %v", headedArgs)
	}
	if !hasFlag(headedArgs, "enable-features") {
		t.Fatalf("headed args keep chromedp enable-features: %v", headedArgs)
	}
	if hasFlag(headedArgs, "user-agent") {
		t.Fatalf("headed args must not force user-agent: %v", headedArgs)
	}
	if browser.HeadlessViewportWidth != 1280 || browser.HeadlessViewportHeight != 720 {
		t.Fatalf("constants %dx%d", browser.HeadlessViewportWidth, browser.HeadlessViewportHeight)
	}
}

func chromeAllocatorArgs(t *testing.T, opts []chromedp.ExecAllocatorOption) []string {
	t.Helper()

	chromeProbeDisplayMu.Lock()
	defer chromeProbeDisplayMu.Unlock()
	prevDisplay, hadDisplay := os.LookupEnv("DISPLAY")
	if err := os.Setenv("DISPLAY", ":1"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if hadDisplay {
			_ = os.Setenv("DISPLAY", prevDisplay)
			return
		}
		_ = os.Unsetenv("DISPLAY")
	}()

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

func hasFlag(args []string, name string) bool {
	prefix := "--" + name
	for _, a := range args {
		if a == prefix || strings.HasPrefix(a, prefix+"=") {
			return true
		}
	}
	return false
}

func TestLightweightAllocatorTiers(t *testing.T) {
	t.Parallel()
	full := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, true))
	for _, name := range []string{"force-device-scale-factor", "renderer-process-limit", "disk-cache-size"} {
		if !hasFlag(full, name) {
			t.Fatalf("headless lightweight missing %s: %v", name, full)
		}
	}
	// Spec 016: experimental js-flags replaces phase-2 320 on default headless lightweight.
	expJS := "--max-old-space-size=128 --expose-gc --optimize-for-size --flush-bytecode --gc-global"
	if !hasFlagValue(full, "js-flags", expJS) {
		t.Fatalf("missing experimental js-flags: %v", full)
	}
	if !hasFlagValue(full, "disk-cache-size", "1048576") {
		t.Fatalf("phase-2 must set disk-cache-size=1048576: %v", full)
	}
	if n := countFlag(full, "disk-cache-size"); n != 1 {
		t.Fatalf("disk-cache-size count = %d, want 1: %v", n, full)
	}
	if n := countFlag(full, "js-flags"); n != 1 {
		t.Fatalf("js-flags count = %d, want 1: %v", n, full)
	}
	headed := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), false, true))
	if !hasFlagValue(headed, "force-device-scale-factor", "1") {
		t.Fatalf("headed-safe missing scale: %v", headed)
	}
	if hasFlag(headed, "renderer-process-limit") || hasFlag(headed, "js-flags") || hasFlag(headed, "disk-cache-size") {
		t.Fatalf("headed-safe must not include headless bundle: %v", headed)
	}
	if hasFlag(headed, "aggressive-cache-discard") {
		t.Fatalf("headed-safe must not include phase-2 flags: %v", headed)
	}
	off := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, false))
	if hasFlag(off, "renderer-process-limit") || hasFlag(off, "disk-cache-size") || hasFlag(off, "force-device-scale-factor") {
		t.Fatalf("opt-out must omit 013 flags: %v", off)
	}
	if hasFlag(off, "aggressive-cache-discard") || hasFlag(off, "disable-domain-reliability") {
		t.Fatalf("opt-out must omit 014 flags: %v", off)
	}
	if hasFlag(off, "memory-pressure-off") {
		t.Fatalf("opt-out must omit phase-3 memory-pressure-off: %v", off)
	}
	if hasFlag(headed, "memory-pressure-off") {
		t.Fatalf("headed-safe must omit phase-3 memory-pressure-off: %v", headed)
	}
	headedOff := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), false, false))
	if hasFlag(headedOff, "force-device-scale-factor") {
		t.Fatalf("headed opt-out must omit scale: %v", headedOff)
	}
	if !hasFlag(full, "disable-gpu-compositing") || !hasFlag(full, "disable-component-update") {
		t.Fatalf("default headless lightweight missing 017 phase-4 flags: %v", full)
	}
	for _, name := range []string{
		"disable-gpu", "disable-webgl", "disable-3d-apis", "disable-software-rasterizer",
		"single-process", "no-zygote", "no-sandbox",
	} {
		if !hasFlag(full, name) {
			t.Fatalf("missing 016 experiment flag %s: %v", name, full)
		}
	}
	if hasFlag(full, "enable-webgl") || hasFlag(full, "use-angle") {
		t.Fatalf("016 experiment must omit webgl on lightweight: %v", full)
	}
}

func TestAllocatorNoLightweightOmitsFullTier(t *testing.T) {
	t.Parallel()
	off := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, false))
	on := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, true))
	if hasFlag(off, "renderer-process-limit") {
		t.Fatal("opt-out included renderer-process-limit")
	}
	if hasFlag(off, "aggressive-cache-discard") {
		t.Fatal("opt-out included phase-2 aggressive-cache-discard")
	}
	if hasFlag(off, "memory-pressure-off") {
		t.Fatal("opt-out included phase-3 memory-pressure-off")
	}
	if hasFlag(off, "single-process") || hasFlag(off, "disable-gpu-compositing") {
		t.Fatal("opt-out included phase-4 or process flags")
	}
	if !hasFlag(off, "enable-webgl") {
		t.Fatal("headless opt-out must keep enable-webgl")
	}
	if !hasFlag(on, "renderer-process-limit") {
		t.Fatal("default headless lightweight missing renderer-process-limit")
	}
	if !hasFlag(on, "aggressive-cache-discard") {
		t.Fatal("default headless lightweight missing phase-2 aggressive-cache-discard")
	}
	if hasFlag(on, "enable-webgl") {
		t.Fatal("default headless lightweight must omit enable-webgl (016)")
	}
}

func TestHeadlessPhase2AllocatorFlags(t *testing.T) {
	t.Parallel()
	full := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, true))
	for _, name := range []string{
		"aggressive-cache-discard",
		"disable-back-forward-cache",
		"disable-domain-reliability",
	} {
		if !hasFlag(full, name) {
			t.Fatalf("missing phase-2 flag %s: %v", name, full)
		}
	}
	df := flagValue(full, "disable-features")
	if df == "" || !strings.Contains(df, "BackForwardCache") {
		t.Fatalf("disable-features missing BackForwardCache merge: %q in %v", df, full)
	}
	if !strings.Contains(df, "IsolateOrigins") {
		t.Fatalf("disable-features missing casino IsolateOrigins: %q", df)
	}

	headed := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), false, true))
	if hasFlag(headed, "aggressive-cache-discard") {
		t.Fatalf("headed must omit phase-2: %v", headed)
	}
	off := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, false))
	if hasFlag(off, "disable-domain-reliability") {
		t.Fatalf("opt-out must omit phase-2: %v", off)
	}
}

func TestHeadlessPhase3AllocatorFlags(t *testing.T) {
	t.Parallel()
	full := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, true))
	if !hasFlag(full, "memory-pressure-off") {
		t.Fatalf("missing phase-3 memory-pressure-off: %v", full)
	}
	for _, name := range []string{
		"process-per-site",
		"disable-application-cache",
		"aggressive-cache-discard",
	} {
		if !hasFlag(full, name) {
			t.Fatalf("phase-3 profile missing prior flag %s: %v", name, full)
		}
	}
	if !hasFlagValue(full, "v8-cache-options", "none") {
		t.Fatalf("missing v8-cache-options=none: %v", full)
	}
	if hasFlag(full, "enable-webgl") || hasFlag(full, "use-angle") {
		t.Fatalf("default headless lightweight must not enable webgl (016 experiment): %v", full)
	}

	headed := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), false, true))
	if hasFlag(headed, "memory-pressure-off") {
		t.Fatalf("headed must omit phase-3: %v", headed)
	}
	off := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, false))
	if hasFlag(off, "memory-pressure-off") {
		t.Fatalf("opt-out must omit phase-3: %v", off)
	}
}

func TestHeadlessPhase3ForbiddenFlags(t *testing.T) {
	t.Parallel()
	full := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, true))
	// Bug -4 unset policy; 016 GPU/WebGL disables apply on default lightweight.
	for _, name := range []string{
		"disable-background-networking",
		"disable-background-timer-throttling",
		"disable-backgrounding-occluded-windows",
		"disable-renderer-backgrounding",
		"disable-remote-fonts",
		"disable-reading-from-canvas",
		"safebrowsing-disable-auto-update",
		"enable-features",
	} {
		if hasFlag(full, name) {
			t.Fatalf("forbidden bug -4 flag --%s present: %v", name, full)
		}
	}
}

func TestHeadlessPhase2ForbiddenFlags(t *testing.T) {
	t.Parallel()
	full := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, true))
	for _, args := range [][]string{full} {
		for _, name := range []string{
			"disable-background-networking",
			"disable-background-timer-throttling",
			"disable-backgrounding-occluded-windows",
			"disable-renderer-backgrounding",
			"disable-remote-fonts",
			"disable-reading-from-canvas",
			"safebrowsing-disable-auto-update",
			"enable-features",
		} {
			if hasFlag(args, name) {
				t.Fatalf("forbidden bug -4 flag --%s present: %v", name, args)
			}
		}
	}
}

func TestHeadlessPhase4AllocatorFlags(t *testing.T) {
	t.Parallel()
	full := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, true))
	for _, name := range []string{
		"disable-gpu-compositing",
		"disable-gpu-rasterization",
		"disable-accelerated-2d-canvas",
		"disable-accelerated-video-decode",
		"disable-component-update",
	} {
		if !hasFlag(full, name) {
			t.Fatalf("missing phase-4 flag %s: %v", name, full)
		}
	}
	df := flagValue(full, "disable-features")
	for _, token := range []string{"OnDeviceModel", "PaintHolding", "AudioServiceOutOfProcess", "IsolateOrigins"} {
		if !strings.Contains(df, token) {
			t.Fatalf("disable-features missing %s: %q", token, df)
		}
	}
	if !hasFlag(full, "single-process") {
		t.Fatalf("default lightweight must include single-process: %v", full)
	}
	for _, name := range []string{"disable-gpu", "disable-webgl"} {
		if !hasFlag(full, name) {
			t.Fatalf("missing 016 experiment flag %s: %v", name, full)
		}
	}
	expJS := "--max-old-space-size=128 --expose-gc --optimize-for-size --flush-bytecode --gc-global"
	if !hasFlagValue(full, "js-flags", expJS) {
		t.Fatalf("js-flags = %q, want %q", flagValue(full, "js-flags"), expJS)
	}
	if hasFlag(full, "enable-webgl") {
		t.Fatalf("default lightweight must not pass enable-webgl: %v", full)
	}
}

func TestHeadlessPhase4ForbiddenFlags(t *testing.T) {
	t.Parallel()
	full := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, true))
	if hasFlag(full, "disable-background-networking") {
		t.Fatalf("forbidden --disable-background-networking present: %v", full)
	}
	if !hasFlag(full, "disable-breakpad") {
		t.Fatalf("default lightweight missing disable-breakpad: %v", full)
	}
}

func TestNoLightweightOmitsPhase4Tiers(t *testing.T) {
	t.Parallel()
	off := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, false))
	headedOff := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), false, false))
	for _, name := range []string{
		"disable-gpu-compositing", "disable-gpu-rasterization", "disable-component-update",
		"js-flags", "renderer-process-limit", "aggressive-cache-discard",
		"memory-pressure-off", "disk-cache-size",
	} {
		if hasFlag(off, name) || hasFlag(headedOff, name) {
			t.Fatalf("--%s on no-lightweight: headless=%v headed=%v", name, off, headedOff)
		}
	}
}

func TestHeadlessShellExecPath(t *testing.T) {
	t.Parallel()
	// chromeAllocatorArgs forces ExecPath("/do-not-run-chrome") after AllocatorOptions,
	// so the bundled path is not visible on the captured argv. HeadlessShellPath is
	// what AllocatorOptions uses when headless (see executable.go).
	if _, err := browser.HeadlessShellPath(); err != nil {
		t.Skip(err)
	}
}

// Spec 017/016: --no-lightweight omits RAM tiers, phase-4, and 016 experiment.
func TestNoLightweightOmitsExperimentalAndLightweightTiers(t *testing.T) {
	t.Parallel()
	off := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), true, false))
	headedOff := chromeAllocatorArgs(t, browser.AllocatorOptions(t.TempDir(), false, false))
	for _, name := range []string{
		"disable-gpu", "disable-webgl", "disable-3d-apis", "disable-software-rasterizer",
		"single-process", "no-zygote", "no-sandbox",
		"js-flags", "renderer-process-limit", "aggressive-cache-discard",
		"memory-pressure-off", "disk-cache-size",
		"disable-gpu-compositing", "disable-gpu-rasterization",
		"disable-accelerated-2d-canvas", "disable-accelerated-video-decode",
		"disable-component-update",
	} {
		if hasFlag(off, name) || hasFlag(headedOff, name) {
			t.Fatalf("--%s on no-lightweight: headless=%v headed=%v", name, off, headedOff)
		}
	}
	for _, label := range []struct {
		name string
		args []string
	}{
		{"headless", off},
		{"headed", headedOff},
	} {
		features := flagValue(label.args, "disable-features")
		if strings.Contains(features, "OnDeviceModel") || strings.Contains(features, "PaintHolding") {
			t.Fatalf("017 phase-4 disable-features on %s no-lightweight: %s", label.name, features)
		}
	}
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

func countFlag(args []string, name string) int {
	prefix := "--" + name
	n := 0
	for _, a := range args {
		if a == prefix || strings.HasPrefix(a, prefix+"=") {
			n++
		}
	}
	return n
}

func flagValue(args []string, name string) string {
	prefix := "--" + name + "="
	for _, a := range args {
		if strings.HasPrefix(a, prefix) {
			return strings.TrimPrefix(a, prefix)
		}
	}
	return ""
}
