# Bug Fix: Headless Spribe config retry — chromedp background suppression flags

- **Slug**: try-limit-of-app-config-exceeded-2
- **Fixed**: 2026-11-22T23:54:00-07:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

For headless instances only, Supervisor now clears chromedp’s four background-suppression Chrome flags so Spribe’s config requests and timers are not throttled while the game iframe loads.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `internal/browser/chromedp.go` | modified | `headlessUnsetBackgroundSuppression` |
| `internal/browser/chromedp_test.go` | modified | Assert four flags absent (headless) / present (headed) |
| `specs/007-enter-aviator-onewin/contracts/aviator-entry.md` | modified | Document headless background policy |

## Diff Highlights (optional)

Headless-only allocator overrides (chromedp `Flag(name, false)` → flag omitted from argv):

- `disable-background-networking`
- `disable-background-timer-throttling`
- `disable-backgrounding-occluded-windows`
- `disable-renderer-backgrounding`

## Tests Added or Updated

- `internal/browser/chromedp_test.go::TestAllocatorOptionsHeadlessWindowSize` — pins headless vs headed background flag behavior.

## Local Verification

- Commands run: `go test ./internal/browser/... ./tests/integration/ -run 'TestAllocatorOptionsHeadlessWindowSize|TestAviatorGameConfig|TestAviatorReadyOnStandIn' -count=1` → pass
- Manual checks: live headless 1win not run in workspace

## Deviations from Assessment

- Assessment written in-repo for this slug (no prior `/speckit-bug-assess` run); content matches user-supplied root-cause flags.
- No additional chromedp defaults changed beyond the four named flags.

## Follow-ups

- Operator: rebuild and `./supervisor start <session> --headless`; confirm game loads and `entry ready`.
- If error persists, compare full headless vs headed argv and consider `disable-extensions` next.