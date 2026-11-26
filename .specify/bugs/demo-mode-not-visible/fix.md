# Bug Fix: Dismiss demo-mode confirmation modal

- **Slug**: demo-mode-not-visible
- **Fixed**: 2026-11-26T13:31:00-07:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

Demo entry now clicks “Seguir en demo” (or the modal close control) on `game-demo-mode-modal` and refuses Aviator ready while that overlay is still visible, so screenshots and watch show the game iframe.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `internal/worker/aviator.go` | modified | `dismissDemoModeModal` after demo activation; `clickDemoModeStay` inside `waitReady`; ready script returns false while the modal is shown |
| `tests/integration/onewin_helpers_test.go` | modified | Stand-in shows `game-demo-mode-modal` after demo click and reveals the iframe only after “Seguir en demo” |
| `tests/integration/onewin_aviator_prefs_test.go` | modified | Demo case asserts a `demo-stay` event before ready |

## Diff Highlights (optional)

After `activatePlay` when `prefs.Demo` is true, entry polls for `[data-testid="game-demo-mode-modal"]` and clicks `[data-testid="game-demo-mode-modal-stay-in-demo-button"]` (fallbacks: header close, “Seguir en demo” / “Stay in demo”). `aviatorReadyEvaluate` treats a visible demo modal as not ready. `waitReady` clicks stay-in-demo again if the modal appears later.

## Tests Added or Updated

- `tests/integration/onewin_aviator_prefs_test.go` `TestAviatorPreferencesReachReady/demo` — demo click must be followed by dismissing the confirmation (`demo-stay`) before the instance can reach ready.

## Local Verification

- Commands run: `go test -count=1 -timeout 180s ./tests/integration/ -run 'TestAviatorPreferencesReachReady'` → pass (`ok`, ~10.8s) using a local `chrome-headless-shell` symlink that was removed afterward.
- Manual checks: none against live 1win.

## Deviations from Assessment

- No separate unit test of the readiness JavaScript string. The integration stand-in covers the same guard: ready cannot succeed until the modal is hidden and the iframe is revealed.
- Dismissal also runs inside `waitReady` if the modal appears after the initial 3s grace window, so a late overlay cannot mark the instance ready.

## Follow-ups

- Confirm on a live `es-MX` (and one other locale) demo session that `screenshot` shows the Aviator iframe after this change (`/speckit-bug-test`).
