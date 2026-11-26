# Bug Assessment: Recurring demo-mode modal during live session

- **Slug**: demo-mode-not-visible-2
- **Created**: 2026-11-26T21:49:00-07:00
- **Revised**: 2026-11-26T21:59:00-07:00 (operator: do not gate `ready` on modal; persistent dismiss)
- **Source**: pasted text + operator screenshots (`screenshot-q5us2nt57nixo-20261126-152309.png`, `screenshot-q5us2nt57nixo-20261126-154831.png`); related: [GitHub #19](https://github.com/cloudpayout/supervisor/issues/19), prior fix `.specify/bugs/demo-mode-not-visible/`
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

> After the first fix, the modal is still appearing. It seems the modal appears at a given moment; the script that closes it should wait until the modal appears (when it is present). The script shouldn't interrupt other scripts running in the page.

**Operator refinement (same session):**

- **`ready` must not wait** for the bothering modal to appear. Entry status should follow the **normal** Aviator-ready workflow (game surface / iframe), not block on demo confirmation.
- The dismiss logic must **not interrupt** the instance’s normal workflow (no aggressive CDP that fights page boot, no lengthening entry solely for modal timing).
- The modal may appear **multiple times** while the session stays live; dismissal must **not stop after the first** successful click.

**Screenshot evidence (2026-11-26, instance `q5us2nt57nixo`):**

- Aviator already running behind the overlay (FUN MODE, bets, multipliers).
- Overlay: **Estás en modo demo** / **Seguir en demo** — `game-demo-mode-modal`.
- Captures at ~15:23 and ~15:48 — consistent with a modal that arrives late and/or **reappears** during an otherwise healthy session.

## Symptom

With `demo: true`, the instance can reach **ready** and play normally, but **screenshot** and **watch** often show the demo confirmation modal covering the game. Expected: the game stays the primary UI for operators; the modal should be cleared **whenever** it is shown, including after `ready` and across the session lifetime, without delaying or destabilizing entry.

## Reproduction

1. Start `onewin` with `"demo": true`.
2. Wait until `supervisor status` shows `ready` (should not require modal to have appeared).
3. Run `supervisor screenshot` or `watch` during the live session (minutes later is still valid).

Observed: modal still obscures captures.

## Suspected Code Paths

- `internal/worker/aviator.go` — `dismissDemoModeModal` runs only during **entry**, returns early after ~3s if modal absent, and is not invoked again after `EnterAviator` completes.
- `internal/worker/aviator.go` — `demoModalOpen()` inside `aviatorReadyEvaluate` **blocks** `ready` while modal visible (contradicts revised goal); may also encourage entry-time churn via `clickDemoModeStay` in `waitReady`.
- `internal/worker/run.go` — after `entry=ready`, no background task watches for `game-demo-mode-modal` before screenshots or while the worker keeps the browser alive.
- `internal/worker/run.go` / capture path — `chromedp.CaptureScreenshot()` uses whatever is painted at capture time; without a **living** dismiss loop, late or recurring modals still appear in output.
- `tests/integration/onewin_helpers_test.go` — synchronous one-shot modal on demo click; no test for **delayed** or **repeated** modal.

## Root Cause Hypothesis

The first fix treated demo-modal dismissal as a **finite entry step**: poll briefly, click once, optionally gate `ready` on no modal. Live 1win shows the overlay **after** the game is ready, and it can **come back** while the session runs. Entry-only dismissal with a **3s give-up** and **no worker-lifetime watcher** means most real captures still hit an open modal, even though the instance correctly reached **ready** on game criteria.

**Confidence:** high.

## Proposed Remediation

**Preferred:**

1. **Decouple `ready` from the demo confirmation modal**
   - Remove (or stop requiring) `demoModalOpen()` as a condition for Aviator ready when `demo: true`.
   - Remove entry-path behavior that **waits** for the modal to exist before proceeding to `ready`.
   - Keep `EnterAviator` focused on currency → demo/play activation → iframe ready, same as non-demo timing expectations.

2. **Run a persistent, low-impact dismiss loop for demo sessions**
   - Start when demo is activated (or when `prefs.Demo` and onewin entry succeeds), continue until worker shutdown / browser close.
   - On each tick (e.g. 250–500ms, not 100ms hammering): **read-only** check for visible `[data-testid="game-demo-mode-modal"]`; only then click `[data-testid="game-demo-mode-modal-stay-in-demo-button"]` or safe close — **one click per appearance**, no clicking when absent (does not interrupt other page scripts).
   - **Never arm a “already dismissed” stop flag** for the whole session; every new visible modal gets the same treatment.
   - Prefer a single injected page script using `MutationObserver` + idle scheduling (`requestAnimationFrame` / `queueMicrotask`) for attach events, with Go-side polling as fallback — minimizes CDP evaluate during boot.

3. **Trim entry-only dismiss**
   - Drop or narrow `dismissDemoModeModal`’s 3s “success if absent” shortcut; entry should not block on modal, but may kick off the lifetime watcher early.
   - Optional: dismiss **once** immediately before `CaptureScreenshot` if watcher is not yet running — belt-and-suspenders, not a substitute for (2).

4. **Tests**
   - Stand-in: `delayDemoModalMS` + ability to **re-show** modal after N seconds to simulate recurrence.
   - Integration: `ready` reached **before** delayed modal fires; watcher records multiple `demo-stay` (or equivalent) events over session lifetime.
   - Regression: `demo: false` → no watcher, no extra goroutines.

**Alternatives:**

- Dismiss only at screenshot time — fixes stills but not `watch`; acceptable as supplement only.
- Block `ready` until modal dismissed — **rejected** per operator requirement.

**Files likely to change:**

- `internal/worker/aviator.go` (ready script, extract shared dismiss probe/click)
- `internal/worker/run.go` (start/stop demo modal watcher with instance lifecycle)
- Possibly capture hook in `ServeCapture` / screenshot path for pre-frame dismiss
- `tests/integration/onewin_helpers_test.go`, new or extended aviator demo test

**Tests to add or update:**

- Demo: ready without modal present; modal appears late → dismissed without changing entry phase backwards.
- Demo: modal shown twice in stand-in → two dismiss actions; instance stays `ready`.
- Non-demo: unchanged `TestAviator*` behavior.

## Risks & Considerations

- Background loop must cancel on worker exit to avoid goroutine/callback leaks.
- Must never click **Jugar con dinero** (`game-demo-mode-modal-play-button`).
- Observer injection should be idempotent if `EnterAviator` retries.
- Slightly higher CDP load on demo instances — keep interval modest.
- Locale/testid drift — keep testids primary, text fallbacks secondary.

## Open Questions

- [NEEDS CLARIFICATION: Preferred interval cap — e.g. 500ms poll vs observer-only until first attach?]
- [NEEDS CLARIFICATION: Should pre-screenshot synchronous dismiss remain in addition to the background watcher?]

## Product constraints (from operator)

| Constraint | Implication |
|------------|-------------|
| Do not wait for modal before `ready` | No modal presence/absence gates in `waitReady` / registry `entry=ready`. |
| Do not interrupt instance workflow | Passive detection; click only when visible; no tight evaluate loops during entry boot. |
| Modal may appear multiple times | Session-long watcher; no single-shot “dismissed forever” flag. |

## URL fetch record

| URL | Host | Policy |
|-----|------|--------|
| (none in this report) | — | — |
| https://github.com/cloudpayout/supervisor/issues/19 | github.com | referenced from prior work; not re-fetched this run |