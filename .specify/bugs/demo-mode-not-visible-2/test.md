# Bug Verification: Recurring demo-mode modal

- **Slug**: demo-mode-not-visible-2
- **Tested**: 2026-11-26T15:35:00-07:00
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: partial

## Summary

Automated checks pass: demo reaches `ready` without waiting on the modal, the `repeat` stand-in case records two `demo-stay` dismissals after ready, and the full `TestAviator*` suite is green. Live 1win screenshot/watch with a delayed or recurring modal was not run in this environment, so operator confirmation is still needed.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | Stand-in: `ready` before delayed modal; `repeat` → 2× `demo-stay` | pass | Via `TestAviatorPreferencesReachReady/demo` and `/repeat` |
| Live reproduction (screenshots) | Headless/headed demo on 1win after ready | not-run | Requires operator session and live site |
| New / updated tests | `go test -count=1 -timeout 300s ./tests/integration/ -run 'TestAviator'` | pass | ~97.7s |
| Regression suite | Same `TestAviator*` run (all Aviator integration tests) | pass | Includes prefs, timeout, config error, etc. |
| Lint / type-check | `go vet ./internal/worker/...` | pass | Run in same shell invocation as tests |

## Output Excerpts

```
ok  	github.com/cloudpayout/supervisor/tests/integration	97.742s
```

## Residual Risks

- In-page `MutationObserver` + 500ms interval behavior on production 1win (Vue timing, locale, re-open rules) is untested outside the stand-in.
- Screenshot pre-capture `clickDemoModeStay` is a belt-and-suspenders path; watch/MJPEG relies mainly on the injected watcher.
- Chrome for tests used a symlinked `chrome-headless-shell` from another worktree.

## Recommendation

Merge is reasonable on integration evidence. Before closing [#19](https://github.com/cloudpayout/supervisor/issues/19) or the follow-up, run one live `"demo": true` session: wait for `ready` (should not hang on the modal), take `screenshot` while the game is running, and confirm no **Estás en modo demo** overlay. If the modal appears again minutes later, wait and screenshot again to confirm the page script dismisses it without blocking workflow. Upgrade this report to **verified** after that manual check passes.