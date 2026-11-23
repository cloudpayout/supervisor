# Bug Assessment: Headless Spribe config error — false `entry ready` + next Chrome flags

- **Slug**: try-limit-of-app-config-exceeded-3
- **Created**: 2026-11-23T00:05:00-07:00
- **Source**: pasted text + live verification (`p7csnu2sphg4g`, `screenshot-p7csnu2sphg4g-20261122-175524.png`, `./test.md`)
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

> Focus on tighten ready / game_config so this error cannot produce `entry ready`. Next headless Chrome flags to try and remove these two flags: `--disable-remote-fonts` and `--disable-reading-from-canvas`.

**Prior attempts** (`try-limit-of-app-config-exceeded`, `-2`): `headless=new`, automation/WebGL tuning, four background-suppression flags cleared. **Live headless still shows** “Retry limit of app config exceeded!”; worker logged **`entry ready`** in ~9s (`p7csnu2sphg4g`).

## Symptom

Headless `supervisor start ~/.supervisor/data.json --headless` leaves a **running** instance with registry **`entry ready`**, but screenshots show the Spribe fatal config badge. Headed runs reportedly load the game.

**Expected**: `entry ready` only when the game is past config bootstrap (no retry-limit message); otherwise `ERROR game_config` and teardown.

**Observed**: Broken game UI + **`entry ready`** (false positive).

## Reproduction

1. `go build -o supervisor ./cmd/supervisor/main.go`
2. `./supervisor start ~/.supervisor/data.json --headless`
3. `./supervisor screenshot <id>` → red **Retry limit of app config exceeded!**
4. `./supervisor logs <id>` → `entry ready` without `ERROR game_config`

**Evidence**: Instance `p7csnu2sphg4g` (2026-11-22); PNG in repo root. Headed contrast not re-run in workspace.

## Suspected Code Paths

- `internal/worker/aviator.go:50-59` — On `waitReady` **success**, `EnterAviator` returns **without** a final `game_config` check; failure path checks `game_config` only after `waitReady` **error**.
- `internal/worker/aviator.go:341-354` — `aviatorReadyEvaluate`: `crossOriginGameShell` / large `best` iframe area can return **ready true** when inner `contentDocument` is null (cross-origin) — matches visible error in frame Supervisor cannot read from the top document.
- `internal/worker/aviator.go:418-424` — `waitReady` returns **nil** immediately when `aviatorReadyEvaluate` or `gameReadyInChildFrames` is true, **without** re-probing `game_config` on that success path (race: ready fires before error paints, or ready heuristics ignore unreadable error).
- `internal/worker/aviator.go:249-262` — `aviatorConfigErrorScript` only sees iframe text via `contentDocument`; **cross-origin** game hosts skip the error string → `game_config` never fires live.
- `internal/worker/aviator.go:444-468` — `gameReadyInAviatorTargets` probes Spribe targets via CDP but **does not** run config-error text in those frame contexts before declaring ready.
- `internal/browser/chromedp.go` — Headless overrides exist; **`disable-remote-fonts`** and **`disable-reading-from-canvas`** are **not** set in repo or in `chromedp@v0.23.0` `DefaultExecAllocatorOptions` (grep). They may appear from **Chrome’s own headless** defaults on the operator build — treat as headless-only `Flag(name, false)` overrides once confirmed on argv (same pattern as `-2` background flags).

## Root Cause Hypothesis

**Two coupled problems:**

1. **Detection gap**: Spribe’s error text lives in a **cross-origin** game frame. Top-document `Evaluate` + `iframe.contentDocument` cannot read it; `waitReady` still hits **geometry/shell heuristics** (`crossOriginGameShell`, large iframe, canvas/app-root in readable frames) and exits success → `entry ready` while the operator sees the error overlay.

2. **Loader/config environment**: Removing four background flags did not fix config bootstrap. Operator suspects **`disable-remote-fonts`** and **`disable-reading-from-canvas`** (observed on real Chrome headless invocations) interfere with canvas/text rendering Spribe needs before config succeeds. Not present in chromedp defaults in-tree; must be **unset headless-only** if they appear on the launched command line.

**Confidence**: **high** for false `entry ready` mechanism (documented code paths + live log/screenshot mismatch); **medium** for the two new Chrome flags as config fix (not in chromedp source; need argv diff headed vs headless on operator Chrome).

## Proposed Remediation

**Preferred** (two tracks in one fix):

### A. Tighten `game_config` / `waitReady` (P0 — stops lying `entry ready`)

1. **Before any successful return** from `waitReady` and from `EnterAviator`, require `aviatorGameConfigFailed` is false; if true → `game_config`.
2. Extend config-error probing to **chromedp frame targets** (reuse `gameReadyInAviatorTargets` target selection): `Evaluate` the same regex in each Aviator/Spribe target’s `frameBody`, not only top `contentDocument`.
3. Narrow `aviatorReadyEvaluate`: remove or gate `crossOriginGameShell` / “large iframe only” ready when config cannot be read — require inner ready marker **or** target-frame body check, not iframe area alone.
4. Remove or fix `EnterAviator` shell fallback (`waitReady` error → `aviatorGameShellVisible` → `return nil`) unless config-error probe passes in targets.
5. Integration: stand-in with config error in **child target** (or `?frame=configerr` cross-origin simulate) must log `game_config` and never set `entry ready`.

### B. Headless Chrome flags (P1 — may fix underlying load)

In `AllocatorOptions` headless branch, append (same `Flag(..., false)` pattern as `headlessUnsetBackgroundSuppression`):

- `disable-remote-fonts` → false (omit from argv)
- `disable-reading-from-canvas` → false (omit from argv)

Extend `chromedp_test.go` to assert these flags are **absent** on headless argv (after building opts). Document that headed path unchanged. If operator argv still lists them, investigate Chrome version adding them implicitly.

**Alternatives**:

- Post-ready screenshot OCR in worker — heavy; not v1.
- Headed-only Aviator — rejects headless automation goal.

**Files likely to change**:

- `internal/worker/aviator.go`
- `internal/browser/chromedp.go`
- `internal/browser/chromedp_test.go`
- `tests/integration/onewin_helpers_test.go`, `onewin_aviator_game_config_test.go`
- `specs/007-enter-aviator-onewin/contracts/aviator-entry.md`

**Tests to add or update**:

- Unit/integration: config error inside Spribe-like target must yield `game_config`, not `ready`.
- Unit: headless allocator omits `disable-remote-fonts` and `disable-reading-from-canvas`.
- Regression: existing `TestAviatorReadyOnStandIn` still passes.

## Risks & Considerations

- Stricter ready may increase `timeout` on slow loads — keep 60s budget; prefer target-frame probes with short timeouts.
- Unsetting font/canvas flags is headless-scoped; minor perf/rendering cost.
- If Chrome never passes those two flags, tests still pass (absent flags).

## Open Questions

- [NEEDS CLARIFICATION: Operator `chrome --version` and actual headless argv slice showing `disable-remote-fonts` / `disable-reading-from-canvas`.]
- [NEEDS CLARIFICATION: Exact iframe origin of error text on live 1win — Aviaport vs bundled `aviator.html` — for target URL matching.]
- [NEEDS CLARIFICATION: Does headed `start` still reach `entry ready` with playable UI on same `data.json` after stricter ready rules?]