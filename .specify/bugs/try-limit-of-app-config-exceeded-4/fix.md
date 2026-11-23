# Bug Fix: Headless Spribe config — User-Agent + unset safebrowsing / enable-features

- **Slug**: try-limit-of-app-config-exceeded-4
- **Fixed**: 2026-11-23T00:20:00-07:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

Headless runs now send a fixed desktop Linux Chrome **User-Agent** (with `SUPERVISOR_HEADLESS_USER_AGENT` override) and omit chromedp’s **`safebrowsing-disable-auto-update`** and **`enable-features`** flags, alongside existing headless Spribe allocator tuning.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `internal/browser/chromedp.go` | modified | `DefaultHeadlessUserAgent`, `HeadlessUserAgent()`, `headlessUnsetChromedpDefaults` |
| `internal/browser/chromedp_test.go` | modified | UA + flag assertions |
| `specs/007-enter-aviator-onewin/contracts/aviator-entry.md` | modified | Headless UA + flag policy |

## Diff Highlights (optional)

- `--user-agent=<DefaultHeadlessUserAgent>` on headless only.
- `Flag("safebrowsing-disable-auto-update", false)` and `Flag("enable-features", false)` in headless unset block.
- `disable-background-networking` remains unset via existing `headlessUnsetBackgroundSuppression` (no headed change).

## Tests Added or Updated

- `internal/browser/chromedp_test.go::TestAllocatorOptionsHeadlessWindowSize` — desktop UA present; `enable-features` / safebrowsing absent on headless; headed keeps `enable-features`, no forced UA.
- Regression: `TestHeadlessViewport`, `TestAviator*` integration (~113s).

## Local Verification

- Commands run: `go test ./internal/browser/... ./tests/integration/ -run 'TestAllocatorOptionsHeadlessWindowSize|TestHeadlessViewport|TestAviator' -count=1 -timeout 5m` → pass
- Manual checks: live headless 1win not run in workspace

## Deviations from Assessment

- None.

## Follow-ups

- Operator: `go build` and headless `start`; compare game screenshot and optional `navigator.userAgent` in headed vs headless.
- If still failing, set `SUPERVISOR_HEADLESS_USER_AGENT` to match headed `navigator.userAgent` exactly before next assess cycle.