# Bug Verification: Aviator session exits shortly after play page opens (live controls)

- **Slug**: aviator-exiting
- **Tested**: 2026-11-22T21:26:00-07:00
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: partial

## Summary

Automated Aviator integration tests pass, including the delayed currency-control regression that models the original timing failure. The operator’s post-fix live runs no longer fail with `ERROR controls` after ~4s; recent logs show `timeout` at ~60s instead, so early teardown from missing currency controls appears resolved, but **Aviator ready** on live 1win was not confirmed in this verification pass.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix, live) | Operator `./supervisor start` + `logs --recent` (terminal history) | partial | Pre-fix: `controls` ~4s (`ffqb6q7i3hfzo`). Post-fix: `timeout` at 60s (`7sbivjbuc5kr4`, `r44wyuyxp6kwg`) — no longer immediate exit on `controls` |
| New / updated tests | `go test ./tests/integration/ -run Aviator -count=1 -timeout 5m` | pass | ~86s |
| Regression (unit) | `go test ./tests/unit/... -count=1` | pass | |
| Build | `go build ./...` | pass | |
| Lint / type-check | (not configured separately) | skipped | `go build` suffices for this Go module |

## Output Excerpts

```text
go test ./tests/integration/ -run Aviator -count=1 -timeout 5m
ok  	github.com/cloudpayout/supervisor/tests/integration	85.985s
```

Operator `logs --recent` (2026-11-22, post-fix):

```text
7sbivjbuc5kr4   2026-11-22T21:24:25Z    timeout
r44wyuyxp6kwg   2026-11-22T21:23:22Z    timeout
6oaiqs3kg3kw4   2026-11-22T20:53:32Z    controls   # likely pre-rebuild
```

## Residual Risks

- Live entry may still fail at **Aviator ready** (`timeout`) — separate from the assessed `controls` failure at currency open.
- No automated test exercises real `game-currency-modal` / `data-testid` DOM (stand-in only).
- Session file fields (`locale`, `currency`, `demo`) on operator machine not inspected here.

## Recommendation

Treat the **immediate `controls` exit** as likely fixed; keep the bug open for live **ready** until a headed run shows `entry ready` in logs and the instance stays in `status`. If `timeout` persists, run `/speckit-bug-assess` with fresh logs from a post-fix start (new slug or append evidence) focusing on iframe ready / play activation after currency modal. Close this slug only after operator confirms ready on live `~/.supervisor/data.json`.