# Bug Fix: Aviator entry hits 60s timeout — live game iframe not matched by ready probe

- **Slug**: aviator-iframe-ready
- **Fixed**: 2026-11-22T21:52:00-07:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

Expanded `waitReady` to discover live 1win game iframes (bundled `aviator.html`, Aviaport / Spribe hosts, largest visible frame) and to accept Aviaport-style ready markers (`.header__game-logo`, `app-game`, `app-header`) in addition to legacy `div.game-logo`. Added a stand-in integration mode that mirrors live DOM without `spribelaunch` in the iframe URL.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `internal/worker/aviator.go` | modified | `aviatorReadyEvaluate` JS probe; `gameReadyInChildFrames` with short timeouts and `chromedp.FromNode` for same-origin frames |
| `tests/integration/onewin_helpers_test.go` | modified | `liveReadyShell` stand-in; `?frame=live` serves Aviaport-like HTML |
| `tests/integration/onewin_aviator_live_frame_test.go` | added | Regression for live-style iframe + ready marker |
| `specs/007-enter-aviator-onewin/contracts/aviator-entry.md` | modified | Document live + legacy ready probes |

## Diff Highlights (optional)

- Iframe `src` matching: `spribelaunch.com/aviator`, `spribegaming.com`, `aviaport`, `/aviator` / `aviator.html`, plus largest visible non-tracking iframe.
- Ready markers: `div.game-logo`, `.header__game-logo`, `app-game`, `app-header`, nested `gameIFrame` canvas/`app-root`.

## Tests Added or Updated

- `tests/integration/onewin_aviator_live_frame_test.go::TestAviatorReadyLiveFrameShell` — relative iframe URL, no `spribelaunch` in `src`, `header__game-logo` inside frame.

## Local Verification

- Commands run: `go test ./tests/integration/ -run Aviator -count=1 -timeout 5m` → pass (~96s)
- Commands run: `go test ./tests/integration/ -run 'TestAviatorReady(LiveFrameShell|OnStandIn)$' -count=1 -timeout 3m` → pass (~5s)

## Deviations from Assessment

- Did not implement concurrent `waitReady` during play activation (optional item); single shared 60s deadline unchanged.
- Cross-origin Aviaport iframes still need `chromedp.WithTargetID` if `contentDocument` is null; same-origin path covers bundled 1win shell and loopback tests. Documented as follow-up.

## Follow-ups

- Headed live verification: `./supervisor start ~/.supervisor/data.json` → logs should show `entry ready` instead of `timeout`.
- If cross-origin launch URLs return, add target-based frame attach for ready probe when `ContentDocument` is nil.