# Bug Verification: Demo mode confirmation modal

- **Slug**: demo-mode-not-visible
- **Tested**: 2026-11-26T13:40:00-07:00
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: partial

## Summary

The fix holds in automated verification: demo entry dismisses the stand-in `game-demo-mode-modal` before reaching `ready`, and the full `TestAviator*` integration suite passes. Live 1win `screenshot` / `watch` on Windows with a real demo session was not exercised in this run, so end-to-end confirmation against production UI remains outstanding.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | Stand-in automated flow: demo → modal → `demo-stay` → `ready` (via `TestAviatorPreferencesReachReady/demo`) | pass | Mirrors assessment DOM (`game-demo-mode-modal-stay-in-demo-button`) |
| Live reproduction (issue #19 steps) | `supervisor start` + `screenshot`/`watch` on live 1win `demo: true` | not-run | Requires operator credentials, live site, Windows env from issue |
| New / updated tests | `go test -count=1 -timeout 300s ./tests/integration/ -run 'TestAviator'` | pass | Includes updated demo prefs case |
| Regression suite | Same `TestAviator*` run (all Aviator integration tests) | pass | ~94.5s |
| Lint / type-check | `go vet ./internal/worker/...` | pass | No issues reported |

## Output Excerpts

```
ok  	github.com/cloudpayout/supervisor/tests/integration	94.458s
```

```
VET_EXIT:0
```

## Residual Risks

- Locale-specific copy or A/B layouts on live 1win may differ from stand-in `data-testid` selectors; manual screenshot on `es-MX` demo session recommended.
- Demo modal might appear with delay longer than exercised in stand-in; fix includes poll + `waitReady` retry path but not stress-tested against slow networks.
- Verification used a symlinked `chrome-headless-shell` from another worktree; CI/local clones without that binary still need `scripts/fetch-chrome-headless-shell.sh`.

## Recommendation

Treat the bug as **likely resolved** for merge and release: integration coverage explicitly models the blocking modal and blocks `ready` until it is cleared. Before closing GitHub issue #19, run one headed or headless demo session on the reporter's environment (`demo: true`, wait for `ready`, then `supervisor screenshot <id>`) and confirm the Aviator iframe is visible with no “Estás en modo demo” overlay. If that passes, upgrade this result to **verified** and close the issue.