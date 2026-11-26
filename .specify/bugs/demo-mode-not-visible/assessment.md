# Bug Assessment: Demo mode confirmation modal blocks Aviator view

- **Slug**: demo-mode-not-visible
- **Created**: 2026-11-26T20:26:00-07:00
- **Source**: https://github.com/cloudpayout/supervisor/issues/19 (allowlisted host: `github.com`) + operator-attached HTML (`aviator-demo-mode-not-visible.html`, `aviator.html`)
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

**GitHub issue #19 — [Start: Modal appears when playing in demo mode](https://github.com/cloudpayout/supervisor/issues/19)**

> When executing an instance in demo mode, and taking a screenshot or watching it in real-time, a modal appears on top of everything. Expected to see the aviator iframe just like when the demo mode is desactivated.
>
> **Steps:** `supervisor start <data-json> --headless` (or headed), wait until `status` is `ready`, then `screenshot` or `watch`.
>
> **Expected:** Aviator iframe visible. **Actual:** A modal covers the page.

Environment: Windows 11, issue includes screenshot evidence.

**Attached HTML** (saved from `https://1win.com/es-MX/casino/play/v_spribe:aviator`): page shows an open overlay `data-testid="game-demo-mode-modal"` with title **“Estás en modo demo”**, subtitle about switching to real money, and buttons **“Jugar con dinero”** / **“Seguir en demo”**, plus `data-testid="modal-header-button-close"`. The Aviator iframe (`aviator-demo.spribegaming.com`) is present in the DOM beneath the modal.

## Symptom

With `"demo": true` in the session file, supervisor reaches `entry=ready`, but screenshots and `watch` capture the outer 1win **demo-mode confirmation modal** instead of the game iframe. Real-money play does not show this modal after entry.

## Reproduction

1. Use a valid `onewin` session JSON with `"demo": true` (and usual cookies).
2. Run `supervisor start <session-file> --headless` (or without `--headless`).
3. Wait until `supervisor status` shows `ready` for the instance.
4. Run `supervisor screenshot <instance-id>` or `supervisor watch <instance-id>`.

Observed: full-page modal (“Estás en modo demo”) obscures the game.

[NEEDS CLARIFICATION: Exact session file used in the issue reporter’s run — likely includes `demo: true` given the modal type.]

## Suspected Code Paths

- `internal/worker/aviator.go` — `EnterAviator()` → `activatePlay(ctx, prefs.Demo, …)` clicks `game-play-card-demo-button` / “Modo demo” but has **no step** to dismiss the post-launch `game-demo-mode-modal`.
- `internal/worker/aviator.go` — `aviatorReadyEvaluate` / `waitReady()`: may mark **Aviator ready** while the demo confirmation modal is still open (iframe can load under `z-index-modal-layout: 1000` overlay). Readiness checks `game-play-card-*` buttons absent and iframe markers, not `[data-testid="game-demo-mode-modal"]` with `desktop-modal_open`.
- `internal/worker/run.go` — sets `entry=ready` immediately after `EnterAviator` succeeds, so operators trust `ready` while UI is still obstructed.

## Root Cause Hypothesis

1win added (or exposes in this locale) a **second modal** after activating demo play: `game-demo-mode-modal`, asking the user to stay in demo or switch to real money. Supervisor’s entry automation was built for the currency/play-card flow only (`game-currency-modal`, `game-play-card-demo-button`). After clicking demo, entry does not click **“Seguir en demo”** (`data-testid="game-demo-mode-modal-stay-in-demo-button"`) or close the dialog, so the instance is logically in demo with the game iframe loaded, but every viewport capture includes the modal backdrop.

**Confidence:** high (confirmed in attached live DOM snapshot; no handling in `aviator.go`).

## Proposed Remediation

**Preferred:** Extend demo entry in `EnterAviator` (after `activatePlay` when `prefs.Demo`):

1. Poll for `[data-testid="game-demo-mode-modal"]` (or visible `desktop-modal_open` within that subtree).
2. If present, click `[data-testid="game-demo-mode-modal-stay-in-demo-button"]` (fallback: `modal-header-button-close`, or text match “Seguir en demo” / “Stay in demo”).
3. Wait until the modal is gone (backdrop not `desktop-modal_open`, or testid removed/hidden) before `waitReady`.
4. Tighten `aviatorReadyEvaluate` (demo-only or always): return **not ready** if `game-demo-mode-modal` is open and visible, so `ready` cannot be set while the overlay blocks the game.

**Alternatives:**

- Click “Jugar con dinero” only when `demo: false` — wrong for this bug; demo sessions must dismiss with “stay in demo”.
- Dismiss modal inside screenshot/watch only — would fix captures but leave misleading `ready` semantics and any future automation touching the page.

**Files likely to change:**

- `internal/worker/aviator.go` — new `dismissDemoModeModal` (or inline step) + readiness guard.
- `tests/integration/onewin_helpers_test.go` — stand-in should render a `game-demo-mode-modal` after demo click to regression-test dismissal.
- `tests/integration/onewin_aviator_prefs_test.go` (or new test) — assert modal cleared before ready / screenshot path.

**Tests to add or update:**

- Integration: `demo: true` → stand-in shows demo confirmation modal → entry clears it → ready; optional screenshot assertion that modal testid absent in top document.
- Unit/JS evaluation test for readiness script rejecting open `game-demo-mode-modal`.

## Risks & Considerations

- 1win may A/B test or locale-variant copy; prefer `data-testid` selectors over Spanish-only labels.
- Modal may appear asynchronously; use existing poll pattern (`aviatorPollInterval`) with budget inside `aviatorEntryBudget`.
- Real-money path must not require demo modal dismissal (guard with `prefs.Demo` or presence detection).
- Clicking wrong button (“Jugar con dinero”) would violate FR-006 for demo sessions — only dismiss via “stay in demo” or close.

## Open Questions

- [NEEDS CLARIFICATION: Does the demo confirmation modal appear in all locales or only some (e.g. `es-MX`)?]
- [NEEDS CLARIFICATION: Should `aviatorReadyEvaluate` treat any open `desktop-modal` as not ready, or only `game-demo-mode-modal`?]

## URL fetch record

| URL | Host | Policy |
|-----|------|--------|
| https://github.com/cloudpayout/supervisor/issues/19 | github.com | allowlisted (fetched via `gh issue view`) |