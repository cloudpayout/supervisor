# Bug Assessment: Headless Aviator shows Spribe “Retry limit of app config exceeded!”

- **Slug**: try-limit-of-app-config-exceeded
- **Created**: 2026-11-22T23:37:00-07:00
- **Source**: pasted text + operator screenshot (1win play page, game canvas area)
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

> When using `--headless` mode, the aviator game doesn't start; the following error message appears: Retry limit of app config exceeded!

**Operator screenshot** (2026-11-22): Authenticated 1win session (`es-MX`), Aviator play route loaded. Game container is black with centered red badge text **“Retry limit of app config exceeded!”** (variant of Spribe’s common **“client config”** retry message seen when the embedded client cannot load its remote configuration). 1win chrome (balance, deposit, notification prompt) is visible; failure is inside the game surface, not at login.

**Unverified external context**: Third-party Aviator demo troubleshooting pages record the same family of message when launching Spribe inside automated Chrome (mobile emulation / headless), before betting controls appear — consistent with provider-side config fetch or environment checks failing in automation, not with a Supervisor string in-repo.

## Symptom

With `supervisor start <session> --headless`, navigation and 1win shell succeed, but the Spribe Aviator client inside the game iframe never reaches a playable state. The operator sees (via `screenshot` or headed observation of the same session) a red error: **Retry limit of app config exceeded!** instead of the game UI.

**Expected**: Headless runs behave like headed runs after feature 012 (1280×720 layout): Spribe config loads, game renders, entry automation can reach **ready** (per `007-enter-aviator-onewin`).

**Observed**: Game client exhausts its internal config retry loop and surfaces the error; Aviator does not start in headless mode.

## Reproduction

1. Build: `go build -o supervisor ./cmd/supervisor/main.go`
2. `./supervisor start ~/.supervisor/data.json --headless` (valid `onewin` session with cookies).
3. Wait for play page / entry (`status` → `running` or `entering`).
4. `./supervisor screenshot <id>` → black game area with red **Retry limit of app config exceeded!** badge.

**Contrast**: Repeat steps 2–4 **without** `--headless` on the same host and session file; game loads and plays (operator-reported; not re-verified in this workspace).

**Environment**: WSL2 Linux, live `1win.com` (screenshot shows `MXN` balance and Spanish notification copy). [NEEDS CLARIFICATION: Chrome version, `CHROME_PATH`, and whether headed vs headless share the same profile.]

## Suspected Code Paths

- `internal/browser/chromedp.go:54-74` — `AllocatorOptions`: headless path keeps chromedp’s default **headless** allocator (`DefaultExecAllocatorOptions` includes `chromedp.Headless`); only adds `WindowSize(1280,720)`. Does **not** apply `VisibleWindow` removals (`enable-automation`, etc.) that headed mode gets. No GPU / `headless=new` / anti-automation tuning for casino embeds.
- `internal/worker/run.go:83-89` — Headless prepends single `EmulateViewport(1280,720)` before first `openActions` batch; no re-apply after in-game navigations (unlikely root cause for config API failures, but relevant to layout).
- `internal/worker/aviator.go:246-336` — `aviatorReadyEvaluate` / `waitReady`: can treat large visible game iframes or `[class*="spribe"]` / `[class*="aviator"]` nodes as “shell visible” without detecting Spribe fatal error text → risk of false **ready** or timeout while UI shows config error.
- `github.com/chromedp/chromedp` `DefaultExecAllocatorOptions` (`allocate.go:90-116`) — Baseline automation flags (`enable-automation`, `disable-extensions`, legacy `--headless=true`) differ materially from headed `VisibleWindow` path; Spribe client may reject config under that fingerprint.
- `specs/012-headless-mode-viewport/spec.md` — SC-003 allows documenting headless Aviator failures unrelated to viewport size; this bug is **not** fixed by 1280×720 alone.

## Root Cause Hypothesis

The error string is emitted by **Spribe’s in-iframe client**, not Supervisor. In `--headless`, Chrome is launched via chromedp defaults (legacy headless, automation infobar flag, no visible window), while headed starts strip several of those flags via `VisibleWindow`. Spribe’s bootstrap likely fetches “app” or “client” config over XHR/WebSocket; in headless/automation that request chain fails or is blocked until an internal retry cap is hit — surfacing **Retry limit of app config exceeded!**

Viewport standardization (012) addresses undersized layout but does not change headless Chrome’s rendering stack, automation bit, or network/user-agent surface that providers often gate. Headed success on the same operator account strongly implicates **browser mode / automation environment**, not session cookies or 1win entry steps.

**Confidence**: **medium** for environment/automation as primary cause (consistent with operator report, screenshot, and external Spribe demo failures in automated Chrome); **low** without captured network logs from a failing headless run proving which config URL fails (4xx, CORS, TLS, or empty body).

## Proposed Remediation

**Preferred**: Harden headless Chrome for embedded Spribe in `AllocatorOptions` (headless branch only), then tighten ready detection.

1. **Allocator (headless)**:
   - Prefer `Flag("headless", "new")` over boolean `true` (replace default `Headless` option ordering so the modern headless path is used).
   - Mirror headed hygiene where safe: delete `enable-automation`; add `Flag("disable-blink-features", "AutomationControlled")`.
   - Enable software/WebGL stack commonly required by canvas games: e.g. `--enable-webgl`, `--use-angle=swiftshader` or platform-appropriate GPU flags; validate on WSL2 CI and operator host.
   - Document required Chrome version in quickstart if a minimum is discovered.

2. **Ready probe** (`aviator.go`):
   - Treat visible text matching `/retry limit of (app|client) config exceeded/i` inside game iframes as a hard failure with stable token `game_config` (distinct from `timeout` / `unexpected`).
   - Avoid counting bare `[class*="spribe"]` loaders as ready when the error badge is visible (narrow the fallback in `aviatorReadyEvaluate`).

3. **Operator docs**: Note that live Aviator in headless may require the flag set above; headed remains the verification path per `007` spec edge case.

**Alternatives**:

- **Xvfb / headed-on-virtual-display** instead of true headless for casino workers — higher ops cost, best compatibility.
- **No product change**: document “Aviator requires headed Chrome” — rejects FR/scenarios that assume `--headless` automation on 1win.

**Files likely to change**:

- `internal/browser/chromedp.go`
- `internal/browser/chromedp_test.go` (assert new flags present only when `headless`)
- `internal/worker/aviator.go`
- `specs/007-enter-aviator-onewin/contracts/aviator-entry.md` or `specs/012-headless-mode-viewport/quickstart.md`

**Tests to add or update**:

- Unit: headless `AllocatorOptions` includes `headless=new` (or chosen policy) and omits `enable-automation`.
- Integration (stand-in): HTML fixture with error node text “Retry limit of app config exceeded!” → entry must **not** reach `ready`; logs or error token `game_config` if wired through worker.
- Manual / optional gated live test: headless `start` + screenshot hash or DOM probe for error string (off by default in CI).

## Risks & Considerations

- Spribe may change error copy (`app` vs `client` config); match case-insensitively.
- GPU/software GL flags vary by OS; WSL2 may need different angles than bare Linux.
- Stripping automation flags can affect other sites; scope changes to headless allocator only.
- False **ready** if only iframe size is checked — error detection reduces silent success.
- Provider-side bot blocking may remain; remediation should log config endpoint failures when CDP Network is enabled for debugging.

## Open Questions

- [NEEDS CLARIFICATION: Does headed `start` with the same session reach `entry ready` in logs, or only “looks playable” to the operator?]
- [NEEDS CLARIFICATION: Failing network request URL/status for Spribe config in headless (HAR or worker Network logging).]
- [NEEDS CLARIFICATION: Whether error appears immediately after play click or only after prior 012 viewport work — i.e. regression vs long-standing Spribe+headless limitation.]