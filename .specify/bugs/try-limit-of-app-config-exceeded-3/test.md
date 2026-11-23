# Bug Verification: Headless Spribe config error — false `entry ready` + next Chrome flags

- **Slug**: try-limit-of-app-config-exceeded-3
- **Tested**: 2026-11-23T00:12:00-07:00
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: failed

## Summary

After rebuilding with the `-3` fix, live headless 1win **still** shows **Retry limit of app config exceeded!** in screenshots, and the worker still logs **`entry ready`** without `ERROR game_config`. Automated tests pass. The primary symptom (broken game) and the false-ready defect **both persist** on instance `iwquviyzougi2`.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | `go build`; `./supervisor start ~/.supervisor/data.json --headless`; `screenshot iwquviyzougi2` | **fail** | Same Spribe error badge in PNG `screenshot-iwquviyzougi2-20261122-180812.png` |
| Registry / logs | `status`, `logs iwquviyzougi2` | **fail** | `entry ready` at 00:07:33 (~18s after entering); no `game_config` |
| New / updated tests | `go test ./internal/browser/... ./internal/worker/... -count=1` | pass | |
| Regression suite | `go test ./tests/integration/ -run 'TestAviator\|TestAllocatorOptions' -count=1 -timeout 5m` | pass | ~98s |
| Lint / type-check | via `go test` | pass | |

## Output Excerpts

Operator log (`iwquviyzougi2`, post-rebuild):

```text
2026-11-23T00:07:14Z NOTICE worker started
2026-11-23T00:07:15Z NOTICE entry entering
2026-11-23T00:07:33Z NOTICE entry ready
```

```text
ok  	github.com/cloudpayout/supervisor/internal/browser	0.007s
ok  	github.com/cloudpayout/supervisor/internal/worker	0.004s
ok  	github.com/cloudpayout/supervisor/tests/integration	98.204s
```

## Residual Risks

- Config error text may render in a CDP target whose URL does not match `aviatorTargetURL`, so `aviatorGameConfigFailedInTargets` never runs there.
- Error may be canvas/shadow-only (not in `body.innerText`), evading regex probe while looser ready paths still match.
- Headless Chrome flag unsetting did not restore Spribe config load; further argv diff vs headed required.
- Stale instance `p7csnu2sphg4g` also showed `ready` in `status` during operator session (pre-fix instance still running).

## Recommendation

**Reopen** — run `/speckit-bug-assess` (e.g. slug `try-limit-of-app-config-exceeded-4`) with: (1) CDP `Target.getTargets` / URL list during error state; (2) screenshot-based or frame-target selector probe for Spribe error node; (3) block `entry ready` until inner game markers pass **and** config probe in **every** child target attached to the game iframe. Stop operator `p7csnu2sphg4g` when testing to avoid confusion. Consider headed argv dump vs headless for next flag experiments (`disable-extensions`, etc.).