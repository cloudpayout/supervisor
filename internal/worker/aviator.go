package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/cloudpayout/supervisor/internal/session"
)

// EnterAviator completes the outer-page currency modal and waits until the
// Spribe launch iframe shows div.game-logo. Errors name a cause, never a cookie.
func EnterAviator(ctx context.Context, prefs session.PlayPreferences) error {
	deadline := time.Now().Add(60 * time.Second)
	if err := openCurrency(ctx, deadline); err != nil {
		return err
	}
	if err := chooseCurrency(ctx, prefs.Currency); err != nil {
		return err
	}
	if err := activatePlay(ctx, prefs.Demo); err != nil {
		return err
	}
	return waitReady(ctx, deadline)
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
	// Live 1win: the currency control is on the outer page. The class hash
	// changes; match the stable fragment from the prior Aviator workflow.
	script := `(() => {
		const el = document.querySelector('[class*="currencyButton"]');
		if (!el) return 'missing';
		(el.closest('button') || el).click();
		return 'ok';
	})()`
	wait, cancelWait := context.WithDeadline(ctx, deadline)
	defer cancelWait()
	result, evalErr := chromedp.Run(wait, chromedp.Evaluate[string](script))
	if evalErr != nil || result != "ok" {
		return fmt.Errorf("controls")
	}
	return nil
}

func chooseCurrency(ctx context.Context, code string) error {
	// code is empty or [A-Z]{3} from the session parser.
	script := fmt.Sprintf(`(() => {
		const code = %q;
		let buttons = [...document.querySelectorAll('[data-supervisor-currency]')];
		if (buttons.length === 0) {
			buttons = [...document.querySelectorAll('button')].filter((b) => /\b[A-Z]{3}\b/.test(b.textContent || ''));
		}
		if (buttons.length === 0) return 'none';
		if (!code) { buttons[0].click(); return 'ok'; }
		const hit = buttons.find((b) => (b.textContent || '').includes(code));
		if (!hit) return 'missing';
		hit.click();
		return 'ok';
	})()`, code)
	result, err := chromedp.Run(ctx, chromedp.Evaluate[string](script))
	if err != nil {
		return fmt.Errorf("controls")
	}
	if result == "missing" {
		return fmt.Errorf("currency")
	}
	if result != "ok" {
		return fmt.Errorf("controls")
	}
	return nil
}

func activatePlay(ctx context.Context, demo bool) error {
	sel := chromedp.CSS(`[data-supervisor-play]`)
	if demo {
		sel = chromedp.CSS(`[data-supervisor-demo]`)
	}
	probe, cancel := context.WithTimeout(ctx, 3*time.Second)
	err := chromedp.Do(probe,
		chromedp.WaitVisible(sel),
		chromedp.Click(sel),
	)
	cancel()
	if err == nil {
		return nil
	}
	labels := []string{"Play", "Jugar"}
	if demo {
		labels = []string{"Modo demo", "Demo"}
	}
	script := fmt.Sprintf(`(() => {
		const labels = %s;
		const buttons = [...document.querySelectorAll('button')];
		const hit = buttons.find((b) => labels.some((label) => (b.textContent || '').trim() === label));
		if (!hit) return 'missing';
		hit.click();
		return 'ok';
	})()`, jsStringArray(labels))
	result, evalErr := chromedp.Run(ctx, chromedp.Evaluate[string](script))
	if evalErr != nil || result != "ok" {
		return fmt.Errorf("controls")
	}
	return nil
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
		case <-time.After(100 * time.Millisecond):
		}
	}
	return fmt.Errorf("timeout")
}
