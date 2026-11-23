# Bug Fix: Headless Spribe config error — false `entry ready` + next Chrome flags

- **Slug**: try-limit-of-app-config-exceeded-3
- **Fixed**: 2026-11-23T00:10:00-07:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

Headless Chrome no longer passes `disable-remote-fonts` or `disable-reading-from-canvas`. Aviator entry probes Spribe config errors in CDP game targets and drops iframe-size-only ready heuristics, so `entry ready` cannot succeed while the retry-limit message is visible.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `internal/browser/chromedp.go` | modified | `headlessUnsetCanvasAndFonts` |
| `internal/browser/chromedp_test.go` | modified | Assert font/canvas flags absent |
| `internal/worker/aviator.go` | modified | Target config probe; stricter `waitReady`; no shell fallback |
| `specs/007-enter-aviator-onewin/contracts/aviator-entry.md` | modified | Ready + headless flag policy |

## Diff Highlights (optional)

- `aviatorGameConfigFailedInTargets` / `aviatorGameConfigFailedAny` — config regex in Spribe CDP targets.
- Removed `crossOriginGameShell`, large-iframe-only, and small-iframe-size ready paths from `aviatorReadyEvaluate`.
- `EnterAviator` final `aviatorGameConfigFailedAny` gate; removed timeout→shell→`return nil` bypass.
- Headless: `Flag("disable-remote-fonts", false)`, `Flag("disable-reading-from-canvas", false)`.

## Tests Added or Updated

- `internal/browser/chromedp_test.go::TestAllocatorOptionsHeadlessWindowSize` — six unset headless flags.
- Regression: `TestAviatorGameConfigErrorRemovesInstance`, `TestAviatorReadyOnStandIn`, full `TestAviator*` suite.

## Local Verification

- Commands run: `go test ./internal/browser/... ./internal/worker/... -count=1` → pass
- Commands run: `go test ./tests/integration/ -run 'TestAviator|TestAllocatorOptions' -count=1 -timeout 5m` → pass (~104s)
- Manual checks: live headless 1win not run in workspace

## Deviations from Assessment

- `aviatorGameShellVisible` left defined but unused (no callers); removal deferred to avoid unrelated diff.

## Follow-ups

- Operator: rebuild; headless `start` — expect `entry ready` only with playable UI, or `ERROR game_config` when badge visible.
- If config error persists without false ready, next assess: remaining chromedp defaults (`disable-extensions`, etc.).