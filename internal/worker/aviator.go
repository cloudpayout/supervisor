package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/cloudpayout/supervisor/internal/session"
)

const (
	aviatorPollInterval = 100 * time.Millisecond
	aviatorEntryBudget  = 60 * time.Second
)

// EnterAviator completes the outer-page currency modal and waits until the
// game iframe shows a registered ready marker (legacy or live Aviaport shell).
// Errors name a cause, never a cookie.
func EnterAviator(ctx context.Context, prefs session.PlayPreferences) error {
	entryDeadline := time.Now().Add(aviatorEntryBudget)
	skipCurrency := false
	if prefs.Currency != "" {
		skipProbe := time.Now().Add(5 * time.Second)
		if skipProbe.After(entryDeadline) {
			skipProbe = entryDeadline
		}
		match, err := playCardShowsCurrency(ctx, prefs.Currency, skipProbe)
		if err != nil {
			return err
		}
		skipCurrency = match
	}
	if !skipCurrency {
		if err := openCurrency(ctx, entryDeadline); err != nil {
			return err
		}
		if err := chooseCurrency(ctx, prefs.Currency, entryDeadline); err != nil {
			return err
		}
	}
	if err := activatePlay(ctx, prefs.Demo, entryDeadline); err != nil {
		return err
	}
	_, _ = chromedp.Run(ctx, chromedp.Sleep(750*time.Millisecond))
	readyDeadline := time.Now().Add(aviatorEntryBudget)
	if err := waitReady(ctx, readyDeadline); err != nil {
		if shell, shellErr := aviatorGameShellVisible(ctx); shellErr == nil && shell {
			return nil
		}
		return err
	}
	return nil
}

func playCardShowsCurrency(ctx context.Context, code string, deadline time.Time) (bool, error) {
	script := fmt.Sprintf(`(() => {
		const want = %q;
		const btn = document.querySelector('[data-testid="game-play-card-currency-button"]');
		if (!btn) return 'wait';
		const codeEl = btn.querySelector('[class*="currencyCode"]');
		const visible = (codeEl ? codeEl.textContent : btn.textContent) || '';
		return visible.includes(want) ? 'match' : 'change';
	})()`, code)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return false, fmt.Errorf("timeout")
		}
		result, err := chromedp.Run(ctx, chromedp.Evaluate[string](script))
		if err != nil {
			return false, fmt.Errorf("controls")
		}
		switch result {
		case "match":
			return true, nil
		case "change":
			return false, nil
		case "wait":
		default:
			return false, fmt.Errorf("controls")
		}
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("timeout")
		case <-time.After(aviatorPollInterval):
		}
	}
	return false, nil
}

func openCurrency(ctx context.Context, deadline time.Time) error {
	probe, cancel := context.WithTimeout(ctx, 3*time.Second)
	sel := chromedp.CSS(`[data-supervisor-currency-open]`)
	err := chromedp.Do(probe,
		chromedp.WaitVisible(sel),
		chromedp.Click(sel),
	)
	cancel()
	if err == nil {
		return nil
	}
	const openScript = `(() => {
		const standin = document.querySelector('[data-supervisor-currency-open]');
		if (standin) {
			standin.click();
			return 'ok';
		}
		let el = document.querySelector('[data-testid="game-play-card-currency-button"]');
		if (!el) {
			el = document.querySelector('[class*="currencyButton"]');
		}
		if (!el) return 'missing';
		if (el.getAttribute('data-testid') === 'game-play-card-currency-button') {
			el.click();
			return 'ok';
		}
		(el.closest('button') || el).click();
		return 'ok';
	})()`
	return pollScript(ctx, deadline, openScript, func(result string) bool {
		return result == "ok"
	}, nil)
}

func chooseCurrency(ctx context.Context, code string, deadline time.Time) error {
	script := fmt.Sprintf(`(() => {
		const code = %q;
		const currencies = document.getElementById('currencies');
		let buttons = [...document.querySelectorAll('[data-supervisor-currency]')];
		if (currencies && !currencies.hidden && buttons.length > 0) {
			if (!code) { buttons[0].click(); return 'ok'; }
			const hit = buttons.find((b) => (b.textContent || '').includes(code));
			if (!hit) return 'missing';
			hit.click();
			return 'ok';
		}
		const modal = document.querySelector('[data-testid="game-currency-modal"]');
		if (!modal) return 'wait';
		if (code) {
			const pick = document.querySelector('[data-testid="game-currency-modal-currency-' + code + '"]');
			if (!pick) return 'missing';
			pick.click();
			return 'ok';
		}
		const rows = [...document.querySelectorAll('button[data-testid^="game-currency-modal-currency-"]')];
		const target = rows[0];
		if (!target) return 'none';
		target.click();
		return 'ok';
	})()`, code)
	return pollScript(ctx, deadline, script, func(result string) bool {
		return result == "ok"
	}, func(result string) error {
		if result == "missing" {
			return fmt.Errorf("currency")
		}
		return nil
	})
}

func activatePlay(ctx context.Context, demo bool, deadline time.Time) error {
	probe, cancel := context.WithTimeout(ctx, 3*time.Second)
	sel := chromedp.CSS(`[data-supervisor-play]`)
	if demo {
		sel = chromedp.CSS(`[data-supervisor-demo]`)
	}
	err := chromedp.Do(probe,
		chromedp.WaitVisible(sel),
		chromedp.Click(sel),
	)
	cancel()
	if err == nil {
		return nil
	}
	labels := []string{"Play", "Jugar"}
	testid := "game-play-card-play-button"
	if demo {
		labels = []string{"Modo demo", "Demo"}
		testid = "game-play-card-demo-button"
	}
	script := fmt.Sprintf(`(() => {
		const testid = %q;
		const labels = %s;
		let hit = document.querySelector('[data-testid="' + testid + '"]');
		if (hit) { hit.click(); return 'ok'; }
		hit = document.querySelector(%q);
		if (hit) { hit.click(); return 'ok'; }
		const buttons = [...document.querySelectorAll('button')];
		hit = buttons.find((b) => labels.some((label) => (b.textContent || '').trim() === label));
		if (!hit) return 'missing';
		hit.click();
		return 'ok';
	})()`, testid, jsStringArray(labels), supervisorPlaySel(demo))
	return pollScript(ctx, deadline, script, func(result string) bool {
		return result == "ok"
	}, nil)
}

func supervisorPlaySel(demo bool) string {
	if demo {
		return `[data-supervisor-demo]`
	}
	return `[data-supervisor-play]`
}

func pollScript(ctx context.Context, deadline time.Time, script string, done func(string) bool, onResult func(string) error) error {
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return fmt.Errorf("timeout")
		}
		result, err := chromedp.Run(ctx, chromedp.Evaluate[string](script))
		if err != nil {
			return fmt.Errorf("controls")
		}
		if onResult != nil {
			if err := onResult(result); err != nil {
				return err
			}
		}
		if done(result) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout")
		case <-time.After(aviatorPollInterval):
		}
	}
	return fmt.Errorf("controls")
}

func jsStringArray(values []string) string {
	out := "["
	for i, value := range values {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf("%q", value)
	}
	return out + "]"
}

const aviatorReadyEvaluate = `(() => {
	function visible(el) {
		if (!el) return false;
		const r = el.getBoundingClientRect();
		return r.width > 8 && r.height > 8;
	}
	function srcLooksLikeGame(src) {
		const s = decodeURIComponent(src || '').toLowerCase();
		if (!s || s === 'about:blank') return false;
		if (s.includes('spribelaunch.com/aviator')) return true;
		if (s.includes('aviaport.spribegaming.com')) return true;
		if (s.includes('spribegaming.com')) return true;
		if (/\/aviator(\.html)?([?#/]|$)/.test(s)) return true;
		return false;
	}
	function innerReady(doc) {
		if (!doc) return false;
		const selectors = [
			'div.game-logo', '.header__game-logo', 'app-game', 'app-header',
			'.main-loading', '.main-loading-logo', 'img.header__game-logo', '.lobby-button',
		];
		for (const sel of selectors) {
			if (visible(doc.querySelector(sel))) return true;
		}
		const nested = doc.querySelector('iframe[title="gameIFrame"], iframe[src*="aviator"]');
		if (nested) {
			try {
				const inner = nested.contentDocument;
				if (inner && (inner.querySelector('canvas') || inner.querySelector('app-root'))) return true;
			} catch (_) {}
		}
		return false;
	}
	function trackingFrame(f) {
		const id = (f.id || '').toLowerCase();
		const title = (f.getAttribute('title') || '').toLowerCase();
		if (id.includes('hj') || title.includes('hj')) return true;
		const st = getComputedStyle(f);
		if (st.display === 'none' || st.visibility === 'hidden') return true;
		const r = f.getBoundingClientRect();
		if (r.width < 4 && r.height < 4) return true;
		return false;
	}
	const frames = [...document.querySelectorAll('iframe')].filter((f) => !trackingFrame(f));
	const ranked = frames.filter((f) => srcLooksLikeGame(f.getAttribute('src') || ''));
	for (const f of ranked.length ? ranked : frames) {
		try {
			if (innerReady(f.contentDocument)) return true;
		} catch (_) {}
	}
	let best = null;
	let area = 0;
	for (const f of frames) {
		const r = f.getBoundingClientRect();
		const a = r.width * r.height;
		if (a > area) {
			area = a;
			best = f;
		}
	}
	if (best && area >= 4000) {
		try {
			if (innerReady(best.contentDocument)) return true;
		} catch (_) {}
	}
	function crossOriginGameShell(frame) {
		const src = decodeURIComponent(frame.getAttribute('src') || '').toLowerCase();
		const title = (frame.getAttribute('title') || '').toLowerCase();
		if (!srcLooksLikeGame(src) && title !== 'gameiframe') return false;
		const r = frame.getBoundingClientRect();
		return r.width * r.height >= 15000;
	}
	for (const f of ranked) {
		if (crossOriginGameShell(f)) return true;
	}
	if (best && area >= 15000) {
		const src = decodeURIComponent(best.getAttribute('src') || '').toLowerCase();
		if (srcLooksLikeGame(src)) return true;
	}
	const onGameSurface = !document.querySelector('[data-testid="game-play-card-play-button"]') &&
		!document.querySelector('[data-testid="game-play-card-demo-button"]');
	if (onGameSurface) {
		for (const f of frames) {
			const r = f.getBoundingClientRect();
			if (r.width >= 120 && r.height >= 80) return true;
		}
		const loader = document.querySelector('.main-loading, [class*="spribe"], [class*="aviator"]');
		if (loader && visible(loader)) return true;
	}
	return false;
})()`

const aviatorShellVisibleScript = aviatorReadyEvaluate

const aviatorFrameBodyReady = `(() => {
	function visible(el) {
		if (!el) return false;
		const r = el.getBoundingClientRect();
		return r.width > 8 && r.height > 8;
	}
	const selectors = [
		'div.game-logo', '.header__game-logo', 'app-game', 'app-header',
		'.main-loading', '.main-loading-logo', 'img.header__game-logo', '.lobby-button',
	];
	for (const sel of selectors) {
		if (visible(document.querySelector(sel))) return true;
	}
	if (visible(document.querySelector('canvas'))) return true;
	if (visible(document.querySelector('app-root'))) return true;
	return false;
})()`

var aviatorFrameReadySelectors = []string{
	`div.game-logo`,
	`.header__game-logo`,
	`app-game`,
	`app-header`,
	`.main-loading`,
	`img.header__game-logo`,
	`.lobby-button`,
}

func aviatorGameShellVisible(ctx context.Context) (bool, error) {
	return chromedp.Run(ctx, chromedp.Evaluate[bool](aviatorShellVisibleScript))
}

func waitReady(ctx context.Context, deadline time.Time) error {
	var lastTargetProbe time.Time
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			if shell, shellErr := aviatorGameShellVisible(ctx); shellErr == nil && shell {
				return nil
			}
			return fmt.Errorf("timeout")
		}
		if ready, err := chromedp.Run(ctx, chromedp.Evaluate[bool](aviatorReadyEvaluate)); err == nil && ready {
			return nil
		}
		if ready, err := gameReadyInChildFrames(ctx); err == nil && ready {
			return nil
		}
		if time.Since(lastTargetProbe) >= 2*time.Second {
			lastTargetProbe = time.Now()
			if ready, err := gameReadyInAviatorTargets(ctx); err == nil && ready {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout")
		case <-time.After(aviatorPollInterval):
		}
	}
	return fmt.Errorf("timeout")
}

func gameReadyInChildFrames(ctx context.Context) (bool, error) {
	probe, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	iframes, err := chromedp.Run(probe, chromedp.Nodes(chromedp.CSSAll(`iframe`)))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return false, nil
		}
		return false, err
	}
	for _, frame := range iframes {
		if frame.ContentDocument == nil {
			continue
		}
		for _, sel := range aviatorFrameReadySelectors {
			selCtx, selCancel := context.WithTimeout(ctx, 300*time.Millisecond)
			found, err := chromedp.Run(selCtx, chromedp.Nodes(chromedp.CSSAll(sel), chromedp.FromNode(frame)))
			selCancel()
			if err == nil && len(found) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func aviatorTargetURL(raw string) bool {
	u := strings.ToLower(strings.TrimSpace(raw))
	if u == "" || strings.HasPrefix(u, "about:") {
		return false
	}
	if strings.Contains(u, "spribelaunch.com/aviator") ||
		strings.Contains(u, "spribegaming.com") ||
		strings.Contains(u, "aviaport.") {
		return true
	}
	if strings.Contains(u, "/aviator") && (strings.Contains(u, "1win.") || strings.Contains(u, "spribe")) {
		return true
	}
	return false
}

func gameReadyInAviatorTargets(ctx context.Context) (bool, error) {
	infos, err := chromedp.Targets(ctx)
	if err != nil {
		return false, err
	}
	for _, info := range infos {
		if !aviatorTargetURL(info.URL) {
			continue
		}
		frameCtx, cancel := chromedp.NewContext(ctx, chromedp.WithTargetID(info.TargetID))
		ready, err := probeFrameTargetReady(frameCtx)
		if err != nil {
			cancel()
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				continue
			}
			return false, err
		}
		if ready {
			cancel()
			return true, nil
		}
		cancel()
	}
	return false, nil
}

func probeFrameTargetReady(frameCtx context.Context) (bool, error) {
	probe, cancel := context.WithTimeout(frameCtx, 500*time.Millisecond)
	defer cancel()
	ready, err := chromedp.Run(probe, chromedp.Evaluate[bool](aviatorFrameBodyReady))
	if err == nil && ready {
		return true, nil
	}
	if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		return false, err
	}
	for _, sel := range aviatorFrameReadySelectors {
		selCtx, selCancel := context.WithTimeout(frameCtx, 400*time.Millisecond)
		found, err := chromedp.Run(selCtx, chromedp.Nodes(chromedp.CSSAll(sel)))
		selCancel()
		if err == nil && len(found) > 0 {
			return true, nil
		}
	}
	return false, nil
}