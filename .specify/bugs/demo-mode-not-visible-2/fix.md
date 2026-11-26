# Bug Fix: Recurring demo-mode modal

- **Slug**: demo-mode-not-visible-2
- **Fixed**: 2026-11-26T15:30:00-07:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

Demo entry no longer waits on or blocks `ready` for the confirmation modal. After the game is ready, a one-time in-page watcher clicks **Seguir en demo** whenever that modal is visible, including later appearances, without a Go poll loop on the browser connection.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `internal/worker/aviator.go` | modified | Removed the entry-time dismiss and the ready-script modal gate. Added `InstallDemoModeDismiss` (MutationObserver + 500ms interval, click only when visible). |
| `internal/worker/run.go` | modified | Installs the dismiss watcher after Aviator entry when `demo` is true, before `entry=ready` is recorded. Screenshot capture still tries one stay-in-demo click if the modal is up. Capture socket failures are now written to the instance log. |
| `tests/integration/onewin_helpers_test.go` | modified | Demo click reveals the game immediately. Modal appears after a delay and can be shown a second time. |
| `tests/integration/onewin_aviator_prefs_test.go` | modified | `repeat` subtest expects `ready` and two `demo-stay` events. |

## Diff Highlights (optional)

`aviatorReadyEvaluate` no longer returns false because `game-demo-mode-modal` is open. `EnterAviator` no longer calls `dismissDemoModeModal`. `InstallDemoModeDismiss` sets `window.__supervisorDemoDismiss` once and clicks `game-demo-mode-modal-stay-in-demo-button` (or the header close / “Seguir en demo” / “Stay in demo”) only while the modal is visible.

## Tests Added or Updated

- `tests/integration/onewin_aviator_prefs_test.go` `TestAviatorPreferencesReachReady/demo` — demo still reaches ready and records `demo-stay` after the delayed modal.
- `tests/integration/onewin_aviator_prefs_test.go` `TestAviatorPreferencesReachReady/repeat` — ready is reached, then the modal is dismissed twice, and entry stays `ready`.

## Local Verification

- Commands run: `go test -count=1 -timeout 180s ./tests/integration/ -run TestAviatorPreferencesReachReady` → pass (`ok`, ~15.5s) with a local `chrome-headless-shell` symlink removed afterward.
- Manual checks: none against live 1win.

## Deviations from Assessment

- Dismissal is an in-page `MutationObserver` plus a 500ms `setInterval`, installed with one CDP call after entry, instead of a Go goroutine that calls `chromedp` for the life of the worker. A Go loop would share the browser connection with entry and screencast. The page script still clicks only when the modal is visible and does not stop after the first dismissal.
- Capture-socket bind failures are logged. A long subtest name produced `bind: invalid argument` on the Unix socket path and previously exited with no log line. The regression subtest is named `repeat` so the path stays short.

## Follow-ups

- Confirm on a live demo session that a modal appearing after `ready`, and again later, is cleared in `screenshot` / `watch` (`/speckit-bug-test slug=demo-mode-not-visible-2`).
