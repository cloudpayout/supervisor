# Bug Fix: Headless Aviator shows Spribe “Retry limit of app config exceeded!”

- **Slug**: try-limit-of-app-config-exceeded
- **Fixed**: 2026-11-22T23:45:00-07:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

Hardened headless Chrome allocation for embedded Spribe (`headless=new`, no automation infobar, WebGL via SwiftShader) and taught Aviator entry to fail fast with `game_config` when the Spribe config retry error is visible, instead of treating a broken game shell as ready.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `internal/browser/chromedp.go` | modified | Headless-only Chrome flags for casino embeds |
| `internal/browser/chromedp_test.go` | modified | Assert `headless=new`, no `enable-automation` |
| `internal/worker/aviator.go` | modified | Config error probe; narrowed ready heuristics |
| `tests/integration/onewin_helpers_test.go` | modified | Stand-in `?frame=configerr` fixture |
| `tests/integration/onewin_aviator_game_config_test.go` | added | Integration regression for `game_config` |
| `specs/007-enter-aviator-onewin/contracts/aviator-entry.md` | modified | Document `game_config` + headless allocator |
| `specs/012-headless-mode-viewport/quickstart.md` | modified | Live headless Aviator verification scenario |

## Diff Highlights (optional)

Headless allocator append:

- `--headless=new`, `enable-automation` disabled, `AutomationControlled` blink feature off, `enable-webgl` + `use-angle=swiftshader`.

`waitReady` polls `aviatorConfigErrorScript` and returns `game_config` before timeout when the Spribe message is visible in the page or same-origin iframes.

## Tests Added or Updated

- `internal/browser/chromedp_test.go::TestAllocatorOptionsHeadlessWindowSize` — headless=new and no automation flag.
- `tests/integration/onewin_aviator_game_config_test.go::TestAviatorGameConfigErrorRemovesInstance` — stand-in config error → log `ERROR game_config`, instance removed.

## Local Verification

- Commands run: `go test ./internal/browser/... ./internal/worker/... -count=1` → pass
- Commands run: `go test ./tests/integration/ -run 'TestAllocatorOptionsHeadlessWindowSize|TestAviatorGameConfigError|TestAviatorReadyOnStandIn' -count=1 -timeout 5m` → pass (~5s)
- Manual checks: none (live 1win headless not run in workspace)

## Deviations from Assessment

- Did not add optional gated live headless screenshot test (assessment listed as optional / off in CI).
- GPU path uses `use-angle=swiftshader` only; no per-OS branching yet.

## Follow-ups

- Operator verification: `./supervisor start ~/.supervisor/data.json --headless` on WSL2 live 1win; confirm game loads and logs show `entry ready`.
- If `game_config` still appears live, capture failing config network requests via CDP for the next assess cycle.