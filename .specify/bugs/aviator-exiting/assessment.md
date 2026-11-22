# Bug Assessment: Aviator session exits shortly after play page opens (live controls)

- **Slug**: aviator-exiting
- **Created**: 2026-11-22T20:50:00-07:00
- **Source**: pasted text + session log attachment (`ffqb6q7i3hfzo.log`)
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

> When executing `supervisor start` with a valid JSON data, the Aviator game is opened, but after a couple of seconds it exists. Take into account the casino opens with an user as expected. Currency modal doesn't display when the browser gets closed. Attached you the session's logs file

**Operator reproduction (terminal)**:

```text
./supervisor start ~/.supervisor/data.json
ffqb6q7i3hfzo
./supervisor logs ffqb6q7i3hfzo
```

**Instance log** (`ffqb6q7i3hfzo`):

```text
2026-11-22T20:43:01.611743789Z NOTICE worker started
2026-11-22T20:43:02.930069719Z NOTICE entry entering
2026-11-22T20:43:05.952811643Z ERROR controls
```

`logs --recent` also shows a prior run (`woezp5efwwzus`) with the same last error: `controls`.

**Live DOM capture** (operator-provided save: `Juego Aviator 1win - ¡Vuela alto y gana en grande!.html`, es-MX play page): the game-balance **play card** is already on the top document when saved. Relevant nodes:

| Control | Selector (stable) | Notes |
|--------|-------------------|--------|
| Currency opener | `[data-testid="game-play-card-currency-button"]` | `div.variant-1_currencyButton-icW8w`; shows MXN + chevron; nested arrow is a child `<button>`, not the ancestor of the opener |
| Real play | `[data-testid="game-play-card-play-button"]` | Label `Jugar` |
| Demo | `[data-testid="game-play-card-demo-button"]` | Label `Modo demo` |

Copy includes `Moneda del saldo juego`. Class prefix drift from RobotCasino (`game-play-card_currencyButton-bVSBd`) to `variant-1_currencyButton-icW8w`, but `[class*="currencyButton"]` still matches **when the node exists**.

**Live DOM capture (currency picker open)** — operator save: `Juego Aviator 1win - … - with currency modal opened.html`. Clicking `game-play-card-currency-button` opens a **desktop modal** (not an inline popover only):

| Step / UI | Selector | Notes |
|-----------|----------|--------|
| Modal root | `[data-testid="game-currency-modal"]` | Rendered under `modal-layout`; title **Elegir moneda** (`modal-header-title`) |
| Search (optional) | `[data-testid="game-currency-modal-search"]` | |
| Pick currency | `[data-testid="game-currency-modal-currency-{CODE}"]` | Each row is a `<button type="button">`; e.g. MXN shows title *Peso mexicano* and subtitle `MXN` in `list-item_subtitle` |
| Close | `[data-testid="modal-header-button-close"]` | |

This differs from `chooseCurrency` today, which scans generic `button` nodes for three-letter text and never waits for `game-currency-modal`. After a successful open click, entry must **wait for the modal**, then click `game-currency-modal-currency-{code}` (or the first list button if `currency` is omitted). Missing code → `currency` error should use absence of that `data-testid`, not unrelated header buttons (`header-balance-currency-text`).

## Symptom

After a successful-looking `start`, Chrome opens the Aviator play URL with an authenticated session, but the browser closes within a few seconds and the instance disappears from `status`. The operator never sees the currency modal stay open because background entry fails during the first automation step (**open currency control**), not because entry reached currency selection.

**Expected**: Entry continues through the outer-page currency modal, play/demo activation, and **Aviator ready** within ~60s, leaving the instance running.

**Observed**: ~3s after `entry entering`, the worker logs `ERROR controls`, tears down the browser (defer in `Run`), and removes the registry row.

## Reproduction

1. Build supervisor: `go build -o supervisor ./cmd/supervisor/main.go`
2. Ensure no instances: `./supervisor status` → `No instances are running.`
3. Run headed start with a valid `onewin` session file: `./supervisor start ~/.supervisor/data.json`
4. Note instance id (e.g. `ffqb6q7i3hfzo`); observe play page briefly, then browser closes.
5. `./supervisor logs <id>` → last line `ERROR controls`; stderr notes instance no longer in status.

**Environment**: WSL2 Linux, live `1win.com` play URL (no `SUPERVISOR_PAGE_URL` override). [NEEDS CLARIFICATION: exact `locale`, `currency`, and `demo` fields in `~/.supervisor/data.json` — not present in workspace.]

## Suspected Code Paths

- `internal/worker/run.go:144-148` — `EnterAviator` error triggers log + return; defer cancels Chrome and `remove()` clears the instance.
- `internal/worker/run.go:65-74` — defer explains why the visible window closes even though `start` already exited successfully (FR-009).
- `internal/worker/aviator.go:29-54` — `openCurrency`: 3s wait for test-only `[data-supervisor-currency-open]`, then a **single** `Evaluate` looking for `[class*="currencyButton"]`; any failure becomes `fmt.Errorf("controls")`.
- `internal/worker/aviator.go:15-26` — `EnterAviator` sequence; `controls` at ~3s implies failure in `openCurrency` before `chooseCurrency` (which would log `currency`).
- `tests/integration/onewin_helpers_test.go:97-109` — stand-in exposes `data-supervisor-*` hooks; live site does not, so only the brittle class fragment path runs in production.

## Root Cause Hypothesis

The worker fails live Aviator entry in `openCurrency` because the live play page does not expose `[data-supervisor-currency-open]`, and the live fallback performs a **single immediate evaluate** for `[class*="currencyButton"]` after the 3s stand-in probe—without polling until the Vue play card mounts. Saved live HTML shows the control is present with `data-testid="game-play-card-currency-button"` and a matching `currencyButton` class fragment, so **selector drift alone is unlikely**; **timing / hydration** is the leading cause of `missing` on that one shot.

The log timing (`entering` at +1.3s, `controls` at +4.3s from worker start) matches “3s `WaitVisible` on missing hook + one failed JS click,” not a currency mismatch (`currency`) or ready timeout (`timeout`).

The operator’s note that the “currency modal doesn’t display” likely means either the play card disappears when the browser is torn down (~4s), or the **currency dropdown** (after clicking MXN) never opens—because `openCurrency` never succeeds.

**Confidence**: high for failure locus and timing hypothesis; high that `data-testid` selectors are correct on current es-MX live DOM.

## Proposed Remediation

**Preferred**: Harden `openCurrency` for live 1win:

1. After navigation, poll the top document until the entry deadline (shared with `EnterAviator`, up to 60s) for `[data-testid="game-play-card-currency-button"]` (fallback: `[class*="currencyButton"]`)—mirror the loop pattern in `waitReady` rather than one-shot `chromedp.Run` after the 3s stand-in timeout.
2. Use the same `data-testid` hooks for play/demo: `game-play-card-play-button`, `game-play-card-demo-button` (keep `Jugar` / `Modo demo` text fallback).
3. Click the opener `div` itself (`closest('button')` must not replace the opener with an unrelated header control). Then wait for `[data-testid="game-currency-modal"]` and select via `[data-testid="game-currency-modal-currency-{CODE}"]` (subtitle/title fallbacks only if testid missing).
4. If session `currency` already matches the visible play-card code and no change is required, skip open+modal and go straight to **Jugar** / demo — [NEEDS CLARIFICATION: product rule when `currency` matches default on card].

Keep error tokens stable (`controls`, `currency`, `timeout`) for logging/CI. Do not log cookie values.

**Alternatives**:

- Add a short post-navigate settle (network idle or `WaitReady`) before entry automation — simpler but may still fail if selectors are wrong.
- Operator-only workaround: use `SUPERVISOR_PAGE_URL` stand-in — not acceptable for production live entry.

**Files likely to change**:

- `internal/worker/aviator.go` (`openCurrency`, possibly `activatePlay` / `chooseCurrency` for consistent polling)
- `tests/integration/onewin_helpers_test.go` or new test — delayed appearance of currency control (not just `slowEntry` on iframe)
- `specs/007-enter-aviator-onewin/contracts/aviator-entry.md` — document live wait/retry behavior if product-visible

**Tests to add or update**:

- Integration: stand-in where `[data-supervisor-currency-open]` appears after ~4s delay; expect `ready`, not `controls`.
- Integration or unit: assert live fallback path retries until control exists (can mock via evaluate script returning `missing` then `ok`).

## Risks & Considerations

- Live DOM changes again — selectors need periodic re-discovery; document operator headed verification.
- Longer polling increases time-to-fail when session is invalid; still bounded by 60s entry budget.
- Clicking wrong control could activate play without currency — keep fail-fast on missing requested `currency` after modal is open.
- Cross-origin iframe rules unchanged; automation must stay on top document per contract.

## Open Questions

- [NEEDS CLARIFICATION: contents of `~/.supervisor/data.json` — especially `locale`, `currency`, `demo` — to rule out locale-specific DOM variants.]
- **Resolved (HTML capture)**: The play card (“Moneda del saldo juego”, MXN chip, Jugar, Modo demo) is on the top document without a separate full-page modal; opening the currency **list** still requires clicking `game-play-card-currency-button` when `currency` must be changed.
- [NEEDS CLARIFICATION: whether `woezp5efwwzus` used the same session file or a different configuration.]