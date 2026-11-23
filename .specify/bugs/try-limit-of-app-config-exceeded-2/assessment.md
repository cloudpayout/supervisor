# Bug Assessment: Headless Spribe config retry — chromedp background suppression flags

- **Slug**: try-limit-of-app-config-exceeded-2
- **Created**: 2026-11-22T23:52:00-07:00
- **Source**: pasted text (follow-up to `try-limit-of-app-config-exceeded`)
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

> In `headless` mode the game is failing with the following error message: Retry limit of app config exceeded. Gathered new root causes, and it may be caused by `--disable-background-networking`, `--disable-background-timer-throttling`, `--disable-backgrounding-occluded-windows`, or `--disable-renderer-backgrounding`

Prior fix (`try-limit-of-app-config-exceeded`) added `headless=new`, automation hygiene, SwiftShader WebGL, and `game_config` detection. **Live headless still shows the Spribe config retry error.**

## Symptom

Headless `supervisor start` reaches the 1win Aviator embed; Spribe displays **Retry limit of app config exceeded** (or **client** config variant). Headed on the same session reportedly works.

## Root Cause Hypothesis

`AllocatorOptions` starts from `chromedp.DefaultExecAllocatorOptions`, which enables Puppeteer-style **background throttling and background networking off** for all headless runs. Spribe’s config bootstrap likely depends on timely iframe network fetches and timers while the game surface is not “focused” in the usual headed sense; suppressing background networking/rendering may prevent config from loading before Spribe’s internal retry cap.

Headed runs use `VisibleWindow`, which does **not** remove these four flags—but headed also differs in other ways. The operator narrowed suspects to the four chromedp defaults that **only affect background work**, making them the best **headless-specific** lever without rebuilding the entire default set.

**Confidence**: medium (plausible mechanism, operator triage); needs live confirmation after change.

## Proposed Remediation

**Preferred** (headless branch only): after other headless overrides, set these flags to **false** in chromedp’s allocator so they are **not** passed to Chrome:

- `disable-background-networking`
- `disable-background-timer-throttling`
- `disable-backgrounding-occluded-windows`
- `disable-renderer-backgrounding`

Keep existing headless Spribe flags from `try-limit-of-app-config-exceeded`. Headed path unchanged.

**Files likely to change**: `internal/browser/chromedp.go`, `internal/browser/chromedp_test.go`, `specs/007-enter-aviator-onewin/contracts/aviator-entry.md`

**Tests**: unit test that headless argv omits the four flags; headed argv still includes them (chromedp default).

## Risks & Considerations

- Slightly higher background CPU/network in headless instances (acceptable for short casino sessions).
- If failure persists, next suspects remain provider bot checks or other chromedp defaults (`disable-extensions`, etc.).