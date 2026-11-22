# Bug Assessment: Aviator entry hits 60s timeout — live game iframe not matched by ready probe

- **Slug**: aviator-iframe-ready
- **Created**: 2026-11-22T21:45:00-07:00
- **Source**: pasted text + operator HTML capture (`c:\Users\oscar\Downloads\aviator-playing.html`) + instance logs (`vdlfewolhbydu`, `7sbivjbuc5kr4`, `suglmfegofaeq` under `~/.local/state/supervisor/logs/`)
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

> Once the automation has entered to aviator game, it closes after 60 seconds. Surely, the automation isn't detecting the iframe. Attached you this HTML page where you can analyse the iframe element: `c:\Users\oscar\Downloads\aviator-playing.html`, also attached you the instance's log file.

**Related prior work**: Bug slug `aviator-exiting` addressed early `ERROR controls` (~4s) on currency open. Post-fix live runs now reach the full **60s entry budget** and fail with `ERROR timeout` (or occasionally `ERROR unexpected`), which matches stalling in **Aviator ready** (`waitReady`) rather than currency controls.

## Symptom

After `supervisor start` with a valid onewin session, Chrome opens the Aviator play URL, the user can see the game loading or playing, but roughly **one minute** later the browser closes and the instance disappears from `status`. Logs show `NOTICE entry entering` followed by `ERROR timeout` (or `ERROR unexpected`), not `controls` or `currency`.

**Expected**: Entry completes within 60s, registry `entry` becomes `ready`, browser stays open.

**Observed**: Worker exhausts the shared 60s entry deadline in `waitReady`; defer tears down Chrome and removes the registry row.

## Reproduction

1. Build: `go build -o supervisor ./cmd/supervisor/main.go`
2. `./supervisor start ~/.supervisor/data.json` (valid live session; headed WSL2 Chrome).
3. Watch Aviator play page; after ~60s window closes.
4. `./supervisor logs <instance-id>` → `entry entering` then `ERROR timeout`.

**Evidence logs** (`~/.local/state/supervisor/logs/`):

```text
# vdlfewolhbydu — full 60s budget
2026-11-22T21:35:38.445134506Z NOTICE entry entering
2026-11-22T21:36:38.535124856Z ERROR timeout

# 7sbivjbuc5kr4 — timeout after partial budget (modal/play consumed earlier wall time)
2026-11-22T21:23:48.53517634Z NOTICE entry entering
2026-11-22T21:24:25.656742689Z ERROR timeout

# suglmfegofaeq — chromedp evaluate failure during ready poll
2026-11-22T21:35:11.580783834Z NOTICE entry entering
2026-11-22T21:35:34.573148426Z ERROR unexpected
```

**Environment**: WSL2, live `https://1win.com/es-MX/casino/play/v_spribe:aviator` (HTML capture saved from this URL). [NEEDS CLARIFICATION: exact `currency`, `demo`, and `locale` in `~/.supervisor/data.json`.]

## Live DOM capture (operator `aviator-playing.html`)

Page saved **while the game is playing** (`https://1win.com/es-MX/casino/play/v_spribe:aviator`).

| Observation | Detail |
|-------------|--------|
| Top-level play card | Absent (`game-play-card-*` testids not present — UI has transitioned to in-game layout). |
| Game container | Single visible game iframe (excluding analytics `_hjSafeContext_*` hidden frame): `<iframe src="./aviator-playing_files/aviator.html" allow="fullscreen; clipboard-write; web-share" allowfullscreen="">` |
| `spribelaunch.com/aviator` in top document | **Not present** in saved capture (no substring match under `aviator-playing*` tree). |
| Inner document (`aviator-playing_files/aviator.html`) | Angular Aviator shell: `app-game`, `app-header`, `header__game-logo` (CSS class on header logo). Count of literal `div.game-logo`: **0** as an element selector (only `header__game-logo` in styles/markup). Count of `spribelaunch`: **0**. |

This contradicts the worker’s ready probe, which only accepts an iframe whose `src` contains `spribelaunch.com/aviator` and an inner `div.game-logo`.

## Suspected Code Paths

- `internal/worker/aviator.go:232-255` — `waitReady`: polls until deadline with JS that finds `iframe[src*="spribelaunch.com/aviator"]` and `frame.contentDocument.querySelector('div.game-logo')`. Never succeeds on current live DOM → `fmt.Errorf("timeout")` at deadline (or `unexpected` on evaluate error).
- `internal/worker/aviator.go:17-42` — `EnterAviator`: single `deadline := time.Now().Add(60 * time.Second)` shared across currency, play activation, and `waitReady`; late steps get less wall time but still end in `timeout` when ready never fires.
- `internal/worker/run.go:144-148` — `EnterAviator` error → `sink.Error` → worker return → browser teardown.
- `tests/integration/onewin_helpers_test.go:133-139` — stand-in sets iframe `src` with query param `marker=spribelaunch.com/aviator` and inner `div.game-logo`; integration tests pass but **do not model** live 1win bundled `aviator.html` iframe.
- `specs/007-enter-aviator-onewin/contracts/aviator-entry.md:23-24` — contract still documents `spribelaunch` + `div.game-logo` only.

## Root Cause Hypothesis

Live 1win Aviator **does not expose** the launch iframe the worker waits for. After play activation, the top document embeds a **same-origin bundled** game document (`aviator.html` on the 1win origin), not an `iframe` whose `src` attribute contains `spribelaunch.com/aviator`. The saved playing-state capture is consistent with the operator’s belief that “the automation isn’t detecting the iframe”: the probe’s **iframe filter** misses the only game frame.

Even if the filter were broadened to that iframe, the **ready marker** is wrong: the loaded game uses Angular markup (`header__game-logo`, `app-game`) rather than a standalone `div.game-logo` from older RobotCasino / Spribe launch HTML. The worker never switches chromedp context into the frame (no `chromedp.FromNode` / target-based frame traversal); it relies on `contentDocument` from a top-document `Evaluate`, which is fragile for cross-origin launch URLs and irrelevant when the iframe is never selected.

Log timing (`timeout` at ~60s from worker start when currency path is fast; `unexpected` when evaluate throws) aligns with **ready detection failure**, not modal automation — matching post-fix operator runs in terminal history (`7sbivjbuc5kr4`, `vdlfewolhbydu`).

**Confidence**: **high** for iframe `src` / inner selector mismatch vs live capture; **medium** on whether an intermediate pre-load state still uses a `spribelaunch` URL before redirect to bundled `aviator.html` (capture is post-play only).

## Proposed Remediation

**Preferred**: Rework **Aviator ready** for live 1win while keeping stand-in compatibility.

1. **Iframe discovery** (top document, poll until deadline):
   - Keep legacy: `decodeURIComponent(src).includes('spribelaunch.com/aviator')`.
   - Add live paths: visible `iframe` whose `src` matches `aviator` / Spribe CDN patterns on `1win.com` (e.g. bundled `…/aviator.html`, `/resources/…/aviator`, or query tokens referencing Spribe).
   - Fallback: after successful play click, the largest visible non-tracking `iframe` (exclude hidden 1×1 analytics frames).

2. **Ready marker inside frame** (use chromedp iframe targeting, not only top-level `Evaluate`):
   - Prefer frame context: locate iframe node → `chromedp.FromNode` / frame executor → wait for any of: `div.game-logo`, `.header__game-logo`, `app-game`, or `app-header` with non-empty layout (tunable list).
   - Document chosen selectors in `contracts/aviator-entry.md`.

3. **Stand-in tests**: Add integration variant where iframe `src` is relative (`?frame=1`) **without** `spribelaunch` in the attribute, inner markup uses `header__game-logo` instead of `div.game-logo`, and entry still reaches `ready`.

4. **Optional**: Start `waitReady` concurrently with play activation (per `007` research) so slow game load does not eat modal time from the shared 60s deadline.

Keep stable error tokens: `timeout`, `unexpected`, `controls`, `currency`.

**Alternatives**:

- Ready = top-document signal only (e.g. play card hidden + game iframe present) — faster but weaker than inner game load.
- Operator workaround: extend contract only after manual headed verification each DOM release.

**Files likely to change**:

- `internal/worker/aviator.go` (`waitReady`, possibly shared iframe helpers)
- `tests/integration/onewin_helpers_test.go` (live-like stand-in branch)
- `specs/007-enter-aviator-onewin/contracts/aviator-entry.md`

**Tests to add or update**:

- Integration: stand-in `liveFrame` mode — relative iframe URL, inner `header__game-logo`, expect `entry ready`.
- Integration: ensure legacy `spribelaunch` + `div.game-logo` stand-in still passes (regression).

## Risks & Considerations

- Live DOM may change again; maintain a short list of probe selectors and headed verification steps.
- Cross-origin Spribe URLs, if they reappear in some locales, require chromedp frame targets — `contentDocument` from parent `Evaluate` may throw (`unexpected`).
- Broad iframe heuristics could false-positive on ads; gate on play-page route and post–play-click state.
- Tightening ready on `app-game` before bet UI is usable is acceptable per FR-008 (automation success ≠ bet placement).

## Open Questions

- [NEEDS CLARIFICATION: `src` of the game iframe **before** clicking Jugar on current live build — is `spribelaunch` ever in the attribute, or only bundled `aviator.html`?]
- [NEEDS CLARIFICATION: session `demo` vs real play — does iframe URL or inner ready DOM differ?]
- [NEEDS CLARIFICATION: whether `suglmefegofaeq` `unexpected` was a transient chromedp glitch or systematic cross-origin `contentDocument` access.]