# Bug Verification: Headless Spribe config — User-Agent + unset safebrowsing / enable-features

- **Slug**: try-limit-of-app-config-exceeded-4
- **Tested**: 2026-11-23T00:25:00-07:00
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: verified

## Summary

After the `-4` fix (desktop **User-Agent**, unset `safebrowsing-disable-auto-update` and `enable-features`), live headless 1win Aviator **loads and shows playable UI** (bet panels, multiplier history). The Spribe **Retry limit of app config exceeded!** badge no longer appears on instance `jwbghxipdo7h2`. Automated tests pass.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | `go build`; `./supervisor start ~/.supervisor/data.json --headless`; `screenshot jwbghxipdo7h2` | **pass** | Game UI visible in `screenshot-jwbghxipdo7h2-20261122-230909.png` |
| Registry / logs | `logs jwbghxipdo7h2`, `status` | pass | `entry ready` ~18s after entering |
| New / updated tests | `go test ./internal/browser/... -run TestAllocatorOptionsHeadlessWindowSize -v` | pass | UA + flag assertions |
| Regression suite | `go test ./tests/integration/ -run 'TestHeadlessViewport\|TestAviator' -count=1 -timeout 5m` | pass | ~122s |
| Lint / type-check | via `go test` | pass | |

## Output Excerpts

Operator log (`jwbghxipdo7h2`):

```text
2026-11-23T05:08:07Z NOTICE worker started
2026-11-23T05:08:11Z NOTICE entry entering
2026-11-23T05:08:29Z NOTICE entry ready
```

```text
--- PASS: TestAllocatorOptionsHeadlessWindowSize (0.00s)
ok  	github.com/cloudpayout/supervisor/internal/browser	0.009s
ok  	github.com/cloudpayout/supervisor/tests/integration	122.459s
```

Live screenshot: Aviator play surface with **Apuesta** controls and multiplier strip (not config retry error).

## Residual Risks

- 1win notification modal still overlays the page; unrelated to Spribe config.
- Hard-coded UA Chrome/131 may drift from operator Chrome; use `SUPERVISOR_HEADLESS_USER_AGENT` if regressions appear after browser upgrade.
- Prior slug `-3` false-ready concerns should be spot-checked on future builds; this run shows **legitimate** ready aligned with visible game UI.

## Recommendation

**Close `-4` for live headless load** on operator WSL2 session tested. Optionally run `/speckit-bug-test` on earlier slugs as historical only, or archive failed `-3` test as pre-`-4` baseline. Re-verify after major Chrome upgrades.