package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/cloudpayout/supervisor/internal/session"
)

const aviatorPollInterval = 100 * time.Millisecond

// EnterAviator completes the outer-page currency modal and waits until the
// Spribe launch iframe shows div.game-logo. Errors name a cause, never a cookie.
func EnterAviator(ctx context.Context, prefs session.PlayPreferences) error {
	deadline := time.Now().Add(60 * time.Second)
	skipCurrency := false
	if prefs.Currency != "" {
		skipProbe := time.Now().Add(5 * time.Second)
		if skipProbe.After(deadline) {
			skipProbe = deadline
		}
		match, err := playCardShowsCurrency(ctx, prefs.Currency, skipProbe)
		if err != nil {
			return err
		}
		skipCurrency = match
	}
	if !skipCurrency {
		if err := openCurrency(ctx, deadline); err != nil {
			return err
		}
		if err := chooseCurrency(ctx, prefs.Currency, deadline); err != nil {
			return err
		}
	}
	if err := activatePlay(ctx, prefs.Demo, deadline); err != nil {
		return err
	}
	return waitReady(ctx, deadline)
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

func waitReady(ctx context.Context, deadline time.Time) error {
	script := `(() => {
		const frame = [...document.querySelectorAll('iframe')].find((f) => decodeURIComponent(f.getAttribute('src') || '').includes('spribelaunch.com/aviator'));
		if (!frame || !frame.contentDocument) return false;
		return !!frame.contentDocument.querySelector('div.game-logo');
	})()`
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return fmt.Errorf("timeout")
		}
		ready, err := chromedp.Run(ctx, chromedp.Evaluate[bool](script))
		if err != nil {
			return fmt.Errorf("unexpected")
		}
		if ready {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout")
		case <-time.After(aviatorPollInterval):
		}
	}
	return fmt.Errorf("timeout")
}