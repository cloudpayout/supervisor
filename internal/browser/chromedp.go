package browser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chromedp/chromedp"
)

// Headless viewport dimensions (spec 012-headless-mode-viewport).
const (
	HeadlessViewportWidth  = 1280
	HeadlessViewportHeight = 720

	// DefaultHeadlessUserAgent is a desktop Linux Chrome UA (no HeadlessChrome token).
	DefaultHeadlessUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

const headlessUserAgentEnv = "SUPERVISOR_HEADLESS_USER_AGENT"

// chromedp defaults disable background networking/throttling (Puppeteer-style).
// Spribe config fetches in headless iframes need normal background behavior.
var headlessUnsetBackgroundSuppression = [...]chromedp.ExecAllocatorOption{
	chromedp.Flag("disable-background-networking", false),
	chromedp.Flag("disable-background-timer-throttling", false),
	chromedp.Flag("disable-backgrounding-occluded-windows", false),
	chromedp.Flag("disable-renderer-backgrounding", false),
}

// Headless Chrome may enable these implicitly; Spribe canvas/config may need them off.
var headlessUnsetCanvasAndFonts = [...]chromedp.ExecAllocatorOption{
	chromedp.Flag("disable-remote-fonts", false),
	chromedp.Flag("disable-reading-from-canvas", false),
}

var headlessUnsetChromedpDefaults = [...]chromedp.ExecAllocatorOption{
	chromedp.Flag("safebrowsing-disable-auto-update", false),
	chromedp.Flag("enable-features", false),
}

// HeadlessUserAgent returns the UA for headless casino runs (env override for experiments).
func HeadlessUserAgent() string {
	if ua := os.Getenv(headlessUserAgentEnv); ua != "" {
		return ua
	}
	return DefaultHeadlessUserAgent
}

// headlessLightweightOptions is the 013 full headless bundle (contracts/browser-allocator.md).
func headlessLightweightOptions() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.Flag("force-device-scale-factor", "1"),
		chromedp.Flag("renderer-process-limit", "1"),
		chromedp.Flag("process-per-site", true),
		chromedp.Flag("media-cache-size", "1"),
		chromedp.Flag("disable-application-cache", true),
		chromedp.Flag("v8-cache-options", "none"),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-permissions-api", true),
		chromedp.Flag("disable-speech-api", true),
		chromedp.Flag("disable-webrtc", true),
		chromedp.Flag("disable-logging", true),
		chromedp.Flag("disable-crash-reporter", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-in-process-stack-traces", true),
		chromedp.Flag("disable-component-extensions-with-background-pages", true),
	}
}

// headedSafeLightweightOptions is the 013 headed-safe tier (device scale only).
func headedSafeLightweightOptions() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.Flag("force-device-scale-factor", "1"),
	}
}

// headlessPhase2LightweightOptions is spec 014 RAM slice (contracts/browser-allocator-phase2.md).
// When experimentalLightweight is true, phase-2 js-flags is omitted (016 wins with one js-flags).
func headlessPhase2LightweightOptions(experimentalLightweight bool) []chromedp.ExecAllocatorOption {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("aggressive-cache-discard", true),
		chromedp.Flag("disable-back-forward-cache", true),
		chromedp.Flag("disable-domain-reliability", true),
		chromedp.Flag("disk-cache-size", "1048576"),
	}
	if !experimentalLightweight {
		opts = append(opts, chromedp.Flag("js-flags", "--max-old-space-size=320"))
	}
	return opts
}

// headlessPhase3LightweightOptions is spec 015 RAM slice (contracts/browser-allocator-phase3.md).
func headlessPhase3LightweightOptions() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.Flag("memory-pressure-off", true),
	}
}

// headlessExperimentalLightweightOptions is spec 016 (contracts/browser-allocator-experimental.md).
// Applied on default headless lightweight after 013/014/015/017 tiers.
func headlessExperimentalLightweightOptions() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-webgl", true),
		chromedp.Flag("disable-3d-apis", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("js-flags", "--max-old-space-size=128 --expose-gc --optimize-for-size --flush-bytecode --gc-global"),
	}
}

// headlessPhase4LightweightOptions is spec 017 (contracts/browser-allocator-phase4.md).
// Adopt set limits GPU compositing/raster and component updates (FR-014).
func headlessPhase4LightweightOptions() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.Flag("disable-gpu-compositing", true),
		chromedp.Flag("disable-gpu-rasterization", true),
		chromedp.Flag("disable-accelerated-2d-canvas", true),
		chromedp.Flag("disable-accelerated-video-decode", true),
		chromedp.Flag("disable-component-update", true),
	}
}

func casinoDisableFeatures(headless, lightweight bool) string {
	base := "IsolateOrigins,site-per-process"
	if headless && lightweight {
		// AudioServiceOutOfProcess is 014; OnDeviceModel and PaintHolding are 017 FR-014.
		return base + ",BackForwardCache,MediaRouter,AudioServiceOutOfProcess,OnDeviceModel,PaintHolding"
	}
	return base
}

// AllocatorOptions builds a dedicated ExecAllocator for one instance.
// lightweight selects the 013 tier: full bundle when headless, headed-safe when visible.
// lightweight false skips all 013 extras (--no-lightweight).
func AllocatorOptions(profileDir string, headless bool, lightweight bool) []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(profileDir),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)
	if headless {
		opts = append(opts,
			chromedp.WindowSize(HeadlessViewportWidth, HeadlessViewportHeight),
			chromedp.Flag("headless", "new"),
			chromedp.UserAgent(HeadlessUserAgent()),
			chromedp.Flag("enable-automation", false),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
		)
		// WebGL + SwiftShader on headless --no-lightweight only; 016 experiment disables GPU/WebGL on lightweight.
		if !lightweight {
			opts = append(opts,
				chromedp.Flag("enable-webgl", true),
				chromedp.Flag("use-angle", "swiftshader"),
			)
		}
		opts = append(opts, headlessUnsetBackgroundSuppression[:]...)
		opts = append(opts, headlessUnsetCanvasAndFonts[:]...)
		opts = append(opts, headlessUnsetChromedpDefaults[:]...)
	} else {
		opts = append(opts, chromedp.VisibleWindow)
	}
	// Live casino embeds Spribe in cross-site iframes; without this, the top
	// document cannot read frame DOM and entry ready probes never succeed.
	opts = append(opts,
		chromedp.Flag("disable-features", casinoDisableFeatures(headless, lightweight)),
		chromedp.Flag("disable-site-isolation-trials", true),
	)
	if lightweight {
		if headless {
			opts = append(opts, headlessLightweightOptions()...)
			opts = append(opts, headlessPhase2LightweightOptions(true)...)
			opts = append(opts, headlessPhase3LightweightOptions()...)
			opts = append(opts, headlessPhase4LightweightOptions()...)
			opts = append(opts, headlessExperimentalLightweightOptions()...)
		} else {
			opts = append(opts, headedSafeLightweightOptions()...)
		}
	}
	if headless {
		if p, err := HeadlessShellPath(); err == nil {
			opts = append(opts, chromedp.ExecPath(p))
		}
	} else if p, err := Executable(); err == nil {
		opts = append(opts, chromedp.ExecPath(p))
	}
	return opts
}

// EnsureDirs creates the profile directory.
func EnsureDirs(profileDir string) error {
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return fmt.Errorf("create profile dir: %w", err)
	}
	return nil
}

// RemoveProfile removes a profile directory created for a failed start.
func RemoveProfile(profileDir string) {
	if profileDir == "" {
		return
	}
	_ = os.RemoveAll(profileDir)
}

// LogFile opens the worker log path, creating parent directories.
func LogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}
