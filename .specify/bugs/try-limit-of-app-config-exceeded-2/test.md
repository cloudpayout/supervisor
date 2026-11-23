# Bug Verification: Headless Spribe config retry — chromedp background suppression flags

- **Slug**: try-limit-of-app-config-exceeded-2
- **Tested**: 2026-11-22T23:58:00-07:00
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: partial

## Summary

Automated checks confirm headless Chrome no longer receives the four background-suppression flags and related Aviator/headless integration tests pass. Live headless reproduction on 1win was **not** run in this environment, so whether the Spribe “Retry limit of app config exceeded” symptom is gone remains **operator-unverified**.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | `./supervisor start <session> --headless` on live 1win Aviator | skipped | Requires operator session, live casino; not run in workspace |
| New / updated tests | `go test ./internal/browser/... -count=1 -v` | pass | `TestAllocatorOptionsHeadlessWindowSize` passes |
| Regression suite (browser/worker/unit) | `go test ./internal/worker/... ./tests/unit/... -count=1` | pass | No regressions in adjacent packages |
| Regression suite (integration Aviator/headless) | `go test ./tests/integration/ -run 'TestHeadlessViewport\|TestAviator' -count=1 -timeout 5m` | pass | ~99s; includes stand-in Aviator paths |
| Lint / type-check | `go test` compile via above | pass | Go 1.27 module |

## Output Excerpts

```text
--- PASS: TestAllocatorOptionsHeadlessWindowSize (0.00s)
PASS
ok  	github.com/cloudpayout/supervisor/internal/browser	0.008s

ok  	github.com/cloudpayout/supervisor/tests/integration	98.787s
```

## Residual Risks

- Fix only removes four chromedp flags; Spribe may still fail headless for other defaults (`disable-extensions`, provider bot checks, WSL2 GPU).
- Headed vs headless behavioral parity was not re-tested live.
- Prior slug `try-limit-of-app-config-exceeded` changes (WebGL, `game_config`) are assumed still in tree; this verification does not re-run that slug’s live checklist.

## Recommendation

Hold full closure until the operator confirms on the same machine that failed before:

1. `go build -o supervisor ./cmd/supervisor/main.go`
2. `./supervisor start ~/.supervisor/data.json --headless`
3. Expect game UI (not config retry banner) and worker log `entry ready` (not `ERROR game_config`).

If live still fails, capture logs and re-run `/speckit-bug-assess` with network/HAR evidence or the next flag suspects from the assessment.