package worker

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/cloudpayout/supervisor/internal/browser"
	"github.com/cloudpayout/supervisor/internal/casino"
	"github.com/cloudpayout/supervisor/internal/logfmt"
	"github.com/cloudpayout/supervisor/internal/proc"
	"github.com/cloudpayout/supervisor/internal/registry"
	"github.com/cloudpayout/supervisor/internal/session"
)

// Run owns one chromedp session until the browser exits or SIGTERM arrives.
// sessionFile is empty for a blank start, which still opens about:blank (FR-010 of feature 001).
func Run(reg registry.Store, id, sessionFile, logPath string) error {
	var inst registry.Instance
	err := reg.View(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i < 0 {
			return fmt.Errorf("instance %s not in registry", id)
		}
		inst = doc.Instances[i]
		return nil
	})
	if err != nil {
		return err
	}
	if logPath == "" {
		logPath = inst.LogPath
	}
	sink, err := logfmt.Open(logPath, cookieSecrets(sessionFile))
	if err != nil {
		return err
	}
	defer sink.Close()
	if _, err := logfmt.AttachProcessStreams(sink); err != nil {
		return err
	}
	sink.Notice("worker started")
	headless := false
	if inst.Headless != nil {
		headless = *inst.Headless
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, browser.AllocatorOptions(inst.ProfileDir, headless, inst.LightweightEnabled())...)
	taskCtx, taskCancel := chromedp.NewContext(allocCtx)

	var browserPID int
	defer func() {
		// FR-007: cancel the browser, wait until it has exited, then delete the profile.
		allocCancel()
		taskCancel()
		deadline := time.Now().Add(5 * time.Second)
		for proc.Alive(browserPID) && time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
		}
		browser.RemoveProfile(inst.ProfileDir)
		_ = remove(reg, id)
	}()

	page, cookies, prefs, err := openTarget(inst.Casino, sessionFile)
	if err != nil {
		sink.Error(err.Error())
		return err
	}
	open := openActions(page, cookies)
	if headless {
		open = append(
			[]chromedp.Action[chromedp.Void]{
				chromedp.EmulateViewport(browser.HeadlessViewportWidth, browser.HeadlessViewportHeight),
			},
			open...,
		)
	}
	if err := chromedp.Do(taskCtx, open...); err != nil {
		if sessionFile == "" {
			sink.Error(err.Error())
			return fmt.Errorf("open about:blank: %w", err)
		}
		sink.Error("open page")
		return fmt.Errorf("open page")
	}

	browserPID, err = findBrowserPID(inst.ProfileDir)
	if err != nil {
		sink.Error(err.Error())
		return err
	}
	layout, err := registry.Resolve()
	if err != nil {
		return err
	}
	socketPath := layout.CaptureSocket(id)

	// The row becomes running before the socket exists, so a starting instance
	// has no listener (FR-010). The CLI retries the dial inside its 15-second budget.
	owner := os.Getpid()
	now := time.Now().UTC()
	err = reg.Update(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i < 0 {
			return fmt.Errorf("instance %s disappeared", id)
		}
		doc.Instances[i].State = registry.StateRunning
		doc.Instances[i].OwnerPID = owner
		doc.Instances[i].BrowserPID = browserPID
		doc.Instances[i].LastSeenAt = &now
		if sessionFile != "" && inst.Casino == "onewin" {
			doc.Instances[i].Entry = "entering"
		}
		return nil
	})
	if err != nil {
		sink.Error(err.Error())
		return err
	}
	if sessionFile != "" && inst.Casino == "onewin" {
		sink.Notice("entry entering")
	}

	listenCtx, listenCancel := context.WithCancel(context.Background())
	captureDone := make(chan struct{})
	var watchGate WatchGate
	go func() {
		defer close(captureDone)
		if err := ServeCapture(listenCtx, socketPath, CaptureHandlers{
			Gate: &watchGate,
			Capture: func(context.Context) ([]byte, error) {
				// Child of taskCtx: this timeout does not close the browser (003 FR-005, FR-009).
				shotCtx, cancel := context.WithTimeout(taskCtx, 10*time.Second)
				defer cancel()
				if prefs.Demo {
					_, _ = clickDemoModeStay(shotCtx)
				}
				return chromedp.Run(shotCtx, chromedp.CaptureScreenshot())
			},
			Watch: func(conn net.Conn, fps int) {
				// 018 FR-004: loopback MJPEG until CLI disconnect or worker shutdown.
				RunWatchSession(conn, taskCtx, nil, fps, &watchGate)
			},
		}); err != nil {
			sink.Error(err.Error())
		}
	}()
	if err := waitForSocket(socketPath, captureDone); err != nil {
		sink.Error(err.Error())
		stopCapture(listenCancel, captureDone, socketPath)
		return err
	}

	if sessionFile != "" && inst.Casino == "onewin" {
		if err := EnterAviator(taskCtx, prefs); err != nil {
			sink.Error(err.Error())
			stopCapture(listenCancel, captureDone, socketPath)
			return err
		}
		if prefs.Demo {
			if err := InstallDemoModeDismiss(taskCtx); err != nil {
				sink.Error(err.Error())
				stopCapture(listenCancel, captureDone, socketPath)
				return err
			}
		}
		if err := reg.Update(func(doc *registry.Document) error {
			i := doc.Find(id)
			if i < 0 {
				return fmt.Errorf("instance %s disappeared", id)
			}
			doc.Instances[i].Entry = "ready"
			return nil
		}); err != nil {
			sink.Error(err.Error())
			stopCapture(listenCancel, captureDone, socketPath)
			return err
		}
		sink.Notice("entry ready")
	}

	<-taskCtx.Done()
	stopCapture(listenCancel, captureDone, socketPath)
	return nil
}

func cookieSecrets(sessionFile string) []string {
	if sessionFile == "" {
		return nil
	}
	_, cookies, _, err := session.Parse(sessionFile, time.Now())
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(cookies))
	for _, c := range cookies {
		if c.Value != "" {
			out = append(out, c.Value)
		}
	}
	return out
}

func openTarget(casinoID, sessionFile string) (string, []session.Cookie, session.PlayPreferences, error) {
	if sessionFile == "" {
		return "about:blank", nil, session.PlayPreferences{}, nil
	}
	fileCasino, cookies, prefs, err := session.Parse(sessionFile, time.Now())
	if err != nil {
		return "", nil, session.PlayPreferences{}, err
	}
	// specs/005-start-command-syntax: casino comes from the registry row set by start.
	if casinoID == "" {
		return "", nil, session.PlayPreferences{}, fmt.Errorf("session file requires a casino on the instance")
	}
	if fileCasino != casinoID {
		return "", nil, session.PlayPreferences{}, fmt.Errorf("session file casino %q does not match instance %q", fileCasino, casinoID)
	}
	if _, err := casino.Lookup(casinoID); err != nil {
		return "", nil, session.PlayPreferences{}, err
	}
	page := casino.PlayURL(prefs.Locale)
	if override := os.Getenv("SUPERVISOR_PAGE_URL"); override != "" {
		// Keep the loopback origin and still expose the locale path segment
		// so stand-in tests can see FR-003 without contacting the live site.
		page = standInPage(override, prefs.Locale)
	}
	return page, cookies, prefs, nil
}

func standInPage(override, locale string) string {
	if locale == "" {
		return override
	}
	u, err := url.Parse(override)
	if err != nil || u.Host == "" {
		return override
	}
	u.Path = "/" + locale + "/"
	u.RawQuery = ""
	return u.String()
}

func openActions(page string, cookies []session.Cookie) []chromedp.Action[chromedp.Void] {
	actions := make([]chromedp.Action[chromedp.Void], 0, len(cookies)+1)
	for i, cookie := range cookies {
		cookie := cookie
		index := i
		actions = append(actions, chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
			if err := setCookie(ctx, t, page, cookie); err != nil {
				return fmt.Errorf("set cookie %d", index)
			}
			return nil
		}))
	}
	actions = append(actions, chromedp.Navigate(page))
	return actions
}

func setCookie(ctx context.Context, target *chromedp.Target, page string, cookie session.Cookie) error {
	params := network.SetCookieParams{
		Name:  cookie.Name,
		Value: cookie.Value,
		URL:   page,
	}
	if cookie.Domain != "" {
		params.Domain = cookie.Domain
	}
	if cookie.Path != "" {
		params.Path = cookie.Path
	}
	if cookie.Secure {
		v := true
		params.Secure = &v
	}
	if cookie.HTTPOnly {
		v := true
		params.HTTPOnly = &v
	}
	if cookie.SameSite != "" {
		params.SameSite = network.CookieSameSite(cookie.SameSite)
	}
	if cookie.Expires != nil {
		params.Expires = cdp.TimeSinceEpoch(float64(cookie.Expires.UnixNano()) / 1e9)
	}
	if cookie.Partitioned {
		params.PartitionKey = &network.CookiePartitionKey{TopLevelSite: topLevelSite(page)}
	}
	_, err := cdp.Call(ctx, target, network.SetCookie, params)
	return err
}

func topLevelSite(page string) string {
	u, err := url.Parse(page)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return page
	}
	return u.Scheme + "://" + u.Host
}

// stopCapture closes the listener and unlinks the socket before the caller
// removes the registry row (FR-010).
func stopCapture(cancel context.CancelFunc, done <-chan struct{}, socketPath string) {
	cancel()
	select {
	case <-done:
	case <-time.After(12 * time.Second):
	}
	_ = os.Remove(socketPath)
}

func waitForSocket(path string, failed <-chan struct{}) error {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		select {
		case <-failed:
			return fmt.Errorf("capture socket failed")
		case <-time.After(20 * time.Millisecond):
		}
	}
	return fmt.Errorf("capture socket not ready")
}

func remove(reg registry.Store, id string) error {
	return reg.Update(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i < 0 {
			return nil
		}
		doc.Instances = append(doc.Instances[:i], doc.Instances[i+1:]...)
		return nil
	})
}

func findBrowserPID(profile string) (int, error) {
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		procs, err := process.Processes()
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		for _, p := range procs {
			cmd, err := p.Cmdline()
			if err != nil || !strings.Contains(cmd, profile) {
				continue
			}
			if strings.Contains(cmd, "--type=") {
				continue
			}
			return int(p.Pid), nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return 0, fmt.Errorf("chrome process for profile %s not found", profile)
}
