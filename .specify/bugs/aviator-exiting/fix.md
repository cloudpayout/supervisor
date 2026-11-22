# Bug Fix: Aviator session exits shortly after play page opens (live controls)

- **Slug**: aviator-exiting
- **Fixed**: 2026-11-22T21:20:00-07:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

Live Aviator entry failed because `openCurrency` probed only once for live DOM after a 3s stand-in wait, so the play-card currency control was often still missing. Entry now polls until the 60s budget, uses 1win `data-testid` hooks for open/play/demo/currency modal selection, and keeps stand-in `data-supervisor-*` behavior for CI.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `internal/worker/aviator.go` | modified | Poll helpers; live testids; modal currency rows; 5s cap on skip-if-matched probe |
| `tests/integration/onewin_helpers_test.go` | (unchanged) | Already had `delayCurrencyMS` delayed opener |
| `tests/integration/onewin_aviator_slow_currency_test.go` | (unchanged) | Regression for delayed currency control |

## Diff Highlights (optional)

- `openCurrency`: poll script tries `[data-supervisor-currency-open]`, then `game-play-card-currency-button`, then `[class*="currencyButton"]`, clicking the opener `div` on live (not a nested chevron `button` via `closest`).
- `chooseCurrency`: stand-in inline list when `#currencies` is visible; else wait for `game-currency-modal` and click `game-currency-modal-currency-{CODE}`.
- `activatePlay`: poll for `game-play-card-play-button` / `game-play-card-demo-button` with label fallbacks.
- `EnterAviator`: skip open+choose when session `currency` already appears on the live play card (5s probe only; stand-in unaffected).

## Tests Added or Updated

- `tests/integration/onewin_aviator_slow_currency_test.go::TestAviatorReadyWithDelayedCurrencyControl` — pins polling until delayed `[data-supervisor-currency-open]` mount (~4.5s).
- Existing Aviator integration suite (currency fail, prefs, ready, entering, timeout) — all pass with updated worker.

## Local Verification

- Commands run: `go build ./...` → ok
- Commands run: `go test ./tests/integration/ -run Aviator -count=1 -timeout 5m` → ok (~87s)
- Commands run: `go test ./tests/unit/... -count=1` → ok
- Manual checks: none (live 1win headed run left to operator)

## Deviations from Assessment

- Did not edit `specs/007-enter-aviator-onewin/contracts/aviator-entry.md` (minimal scope; live selectors documented in assessment only).
- Skip-when-currency-matches uses a **5s** play-card probe so stand-in runs do not burn the full 60s entry budget waiting for a non-existent live testid.

## Follow-ups

- Operator headed verification on live `es-MX` with `~/.supervisor/data.json`.
- Optional stand-in HTML mimicking `game-currency-modal` testids for closer live parity without hitting 1win.