# Bug Assessment: Headless Spribe config — User-Agent + unset safebrowsing / enable-features

- **Slug**: try-limit-of-app-config-exceeded-4
- **Created**: 2026-11-23T00:15:00-07:00
- **Source**: pasted text (follow-up to `try-limit-of-app-config-exceeded-3` failed live test)
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

> Change the User-Agent for headless mode. Let's try to remove these flags too: `--disable-background-networking`, `--safebrowsing-disable-auto-update`, and `--enable-features`

**Chain context**: Live headless 1win Aviator still shows **Retry limit of app config exceeded!** after slugs `-1`…`-3`. Instance `iwquviyzougi2` (post `-3` rebuild) logged `entry ready` while screenshot showed the error (`./test.md` on slug `-3`). Headed runs reportedly work on the same session.

Supervisor **does not** set `--user-agent` today; headless and headed rely on Chrome’s default strings (which may differ under `headless=new`).

## Symptom

`supervisor start <session> --headless` → Spribe client fails config bootstrap (retry-limit message). Headed start on the same host/session loads the game.

**Expected**: Headless presents a **headed-like** network fingerprint (UA + flags) so Spribe config loads; entry reaches true ready or fails with `game_config`, not silent false ready.

**Observed**: Config error UI persists; prior fixes to flags and ready probes did not resolve live load.

## Reproduction

1. `go build -o supervisor ./cmd/supervisor/main.go`
2. `./supervisor start ~/.supervisor/data.json --headless`
3. `./supervisor screenshot <id>` → Spribe **Retry limit of app config exceeded!**
4. Optional contrast: same steps without `--headless`

[NEEDS CLARIFICATION: Capture `navigator.userAgent` from headed vs headless on operator Chrome for before/after proof.]

## Suspected Code Paths

- `internal/browser/chromedp.go:69-98` — `AllocatorOptions` headless branch: no `chromedp.UserAgent`; inherits chromedp `DefaultExecAllocatorOptions` including `enable-features=NetworkService,NetworkServiceInProcess`, `safebrowsing-disable-auto-update`, and (by default) `disable-background-networking` **overridden to off** only on headless via `headlessUnsetBackgroundSuppression` (`-2`).
- `github.com/chromedp/chromedp` `DefaultExecAllocatorOptions` (`allocate.go:96-114`) — sources of `disable-background-networking`, `enable-features`, `safebrowsing-disable-auto-update`.
- `internal/worker/run.go` — open batch runs `EmulateViewport` headless only; **no** `network.SetUserAgentOverride` in worker.
- `internal/worker/aviator.go` — config/ready probes; false `entry ready` on live remains a **parallel** defect if Spribe rejects headless before error text is probe-able (see `-3` test.md).

## Root Cause Hypothesis

Spribe’s config API likely **rejects or mis-serves** headless clients that look automated or non-desktop: different **User-Agent** (including `HeadlessChrome` or missing platform tokens on some builds), combined with **Puppeteer-style Chrome flags** that headed `VisibleWindow` never strips but headless still inherits (`enable-features`, safe-browsing auto-update off, etc.).

`disable-background-networking` is **already unset** for headless in-tree (`Flag(..., false)`); repeating it in `-4` is **verification/idempotent** unless operator binary predates `-2`. New levers for this slug are **explicit desktop User-Agent** and unsetting **`safebrowsing-disable-auto-update`** and **`enable-features`** on headless only.

**Confidence**: **medium** for UA + remaining chromedp defaults as blockers; **low** until operator compares UA strings and live retest.

## Proposed Remediation

**Preferred** (headless-only in `AllocatorOptions`):

1. **User-Agent** — Append `chromedp.UserAgent(...)` with a **stable desktop Chrome on Linux** string (constant in `internal/browser`, e.g. aligned to operator’s headed `navigator.userAgent` once captured). Avoid `HeadlessChrome` token. Optional v2: env `SUPERVISOR_HEADLESS_USER_AGENT` override for experiments without rebuild.

2. **Unset flags** (same `Flag(name, false)` pattern as `headlessUnsetBackgroundSuppression`):
   - `disable-background-networking` — **already** in `headlessUnsetBackgroundSuppression`; keep (no headed change).
   - `safebrowsing-disable-auto-update` — add to headless unset list.
   - `enable-features` — add `Flag("enable-features", false)` so chromedp’s `NetworkService,NetworkServiceInProcess` pair is **not** passed on headless (Chrome uses its own defaults). **Do not** remove Supervisor’s separate `disable-features=IsolateOrigins,site-per-process` (casino iframe requirement).

3. **Tests** — Extend `TestAllocatorOptionsHeadlessWindowSize`:
   - Assert `--user-agent=` present with expected constant prefix (`Chrome/`).
   - Assert absent: `safebrowsing-disable-auto-update`, `enable-features`, and (existing) `disable-background-networking`.
   - Headed argv **unchanged** (still has chromedp background + enable-features + safebrowsing flags per current test).

4. **Docs** — One line in `specs/007-enter-aviator-onewin/contracts/aviator-entry.md` headless section: fixed desktop UA + flag unset list.

**Alternatives**:

- CDP `Network.setUserAgentOverride` in `worker/run.go` after navigate — later override; allocator flag is simpler and affects first request.
- Copy UA from headed instance at runtime — complex; constant + env override sufficient for v1.

**Files likely to change**:

- `internal/browser/chromedp.go`
- `internal/browser/chromedp_test.go`
- `specs/007-enter-aviator-onewin/contracts/aviator-entry.md`

**Tests to add or update**:

- Unit: headless argv UA + three flags absent.
- Live: operator headless restart (not CI); document in bug-test.

**Out of scope for this slug** (unless bundled by fix implementer): further `game_config` / target probe work from `-3` failed test — recommend separate follow-up if `-4` loads game but false ready remains.

## Risks & Considerations

- Removing `enable-features` on headless may affect Chrome networking stack; watch for regressions in blank `start --headless` integration tests.
- Hard-coded UA ages; document bump cadence or env override.
- Spribe may still block on other flags (`disable-extensions`, WebGL path); `-4` is one experiment, not guaranteed complete fix.
- Fake UA vs real Chrome version mismatch could trigger other checks — prefer UA matching operator’s headed string when known.

## Open Questions

- [NEEDS CLARIFICATION: Headed vs headless `navigator.userAgent` on operator WSL2 Chrome.]
- [NEEDS CLARIFICATION: Does live config request return 403/blocked in Network tab when error shows?]
- [NEEDS CLARIFICATION: Should UA match Windows if 1win expects Windows clients despite WSL2 host?]