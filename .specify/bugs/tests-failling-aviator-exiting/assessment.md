# Bug Assessment: Default checks fail on `fix/aviator-exiting` (TestStopAllSIGINT TempDir cleanup)

- **Slug**: tests-failling-aviator-exiting
- **Created**: 2026-11-22T22:43:00-07:00
- **Source**: https://github.com/cloudpayout/supervisor/actions/runs/45078677475/job/137381570238?pr=13
- **URL fetch**: host `github.com`, policy **allowlisted** (`gh run view 45078677475 --log-failed`, `gh api .../jobs/137381570238`)
- **Verdict**: valid
- **Severity**: medium

## Report (verbatim or summarized)

**CI job**: Workflow *Default checks*, job `default-checks`, branch `fix/aviator-exiting`, head SHA `2dab9b3` (*fix(worker): detect live Spribe iframe for Aviator ready*), PR [#13](https://github.com/cloudpayout/supervisor/pull/13).

**Failure excerpt** (step *Run the default checks once*):

```text
    testing.go:1617: TempDir RemoveAll cleanup: unlinkat /tmp/TestStopAllSIGINT962900100/001/data/supervisor/profiles/5u2tmhlvcwmcg/Default: directory not empty
--- FAIL: TestStopAllSIGINT (0.28s)
FAIL    github.com/cloudpayout/supervisor/tests/integration    118.015s
```

**Related runs on the same branch** (same failure signature):

| Run | Trigger | Failing test |
|-----|---------|----------------|
| [45078677475](https://github.com/cloudpayout/supervisor/actions/runs/45078677475) | push | `TestStopAllSIGINT` → `unlinkat .../Default: directory not empty` |
| [45072173931](https://github.com/cloudpayout/supervisor/actions/runs/45072173931) | pull_request | `TestStopAllSIGINT` → `unlinkat .../Default: directory not empty` |

**Note**: Despite the PR title and slug, **no Aviator integration test failed** in run `45078677475`. All `TestAviator*` cases in `tests/integration` reported `pass` (including `TestAviatorReadyLiveFrameShell`, `TestAviatorTimeoutRemovesInstance` ~63s, etc.). The job exits `1` solely because the integration package ends with one failed test.

## Symptom

**Expected**: `go test -json ./...` on `ubuntu-latest` with Chrome installed completes with exit `0` for PR #13.

**Observed**: `tests/integration` fails after ~118s. The only failing test is `TestStopAllSIGINT`; the failure message is Go’s `testing.T.TempDir` post-test cleanup, not an assertion inside the test body. Operators perceive this as “Aviator PR CI red” because it blocks merge on the `fix/aviator-exiting` branch.

## Reproduction

1. Checkout `fix/aviator-exiting` (or current `main` if `TestStopAllSIGINT` is unchanged).
2. Install Chrome; set `XDG_*` via test harness as in CI.
3. `go test ./tests/integration/ -run TestStopAllSIGINT -count=1 -v`
4. On CI (`ubuntu-latest`), failure manifests as TempDir `unlinkat .../profiles/<id>/Default: directory not empty` after the test function returns.
5. [NEEDS CLARIFICATION: local reproduction rate on WSL2 — may be timing-dependent; CI shows repeated failure on two consecutive PR runs.]

## Suspected Code Paths

- `tests/integration/stop_all_test.go:178-236` — `TestStopAllSIGINT`: starts two headless instances under `isolatedEnv` (`t.TempDir()`), runs `stop all` with `SUPERVISOR_STOP_ALL_PAUSE=3s`, sends `SIGINT` after one `stopped instance` line; **no `t.Cleanup` to stop remaining instances or wait for Chrome**.
- `tests/integration/harness_test.go:50-61` — `isolatedEnv` roots profiles under `t.TempDir()/data/supervisor/profiles/`.
- `internal/cli/stop.go:117-179` — `runStopAll`: on interrupt, breaks loop without stopping remaining IDs; at most one `stopOne` completes (removes one profile via `browser.RemoveProfile`).
- `internal/cli/stop.go:97-102` — `stopOne` terminates browser and removes profile for stopped IDs only.
- `internal/browser/chromedp.go:77-82` — `RemoveProfile` uses `os.RemoveAll`; ignored errors do not help TempDir cleanup if Chrome still holds the second profile.
- `tests/integration/status_concurrent_test.go:28-37` — `startHeadless`; contrast with `TestStatusReflectsConcurrentChanges`, which registers `t.Cleanup` → `stopID`.

## Root Cause Hypothesis

Confidence: **high**.

`TestStopAllSIGINT` is designed to leave at least one instance **still registered and running** after SIGINT (assertion: `stoppedCount >= 1`). That intentionally stops only the first ID in the bulk loop. The second headless Chrome process and its user-data directory (`.../profiles/<id2>/Default/`) remain on disk under the test’s `t.TempDir()` tree.

When the test function returns, Go 1.20+ runs `TempDir` cleanup via `os.RemoveAll`. Chrome (or recently exited Chrome) often still has files under `Default/`, so `unlinkat ... directory not empty` is reported at `testing.go:1617` and the test is marked **FAIL** even if all explicit assertions passed. This is a **harness / teardown defect**, not a regression in Aviator worker logic or the SIGINT behavior under test.

The Aviator-related commits on the branch lengthen integration runtime (~118s) but do not appear in the failure stack; they only increase exposure of an existing stop-all interrupt test that lacks post-test browser cleanup.

## Proposed Remediation

**Preferred**: Add deterministic test teardown after assertions in `TestStopAllSIGINT` (and optionally a shared helper for interrupt tests):

1. Register `t.Cleanup` (or defer before return) that:
   - Runs `stop all` with stdin drained **or** loops `stopID` over any instance IDs still in the registry snapshot.
   - Polls until `profilesDir(env)` is empty or times out (reuse patterns from `onewin_usage_test.go:68-71`).
2. Optionally `proc.KillByProfile` for any profile dirs still listed if `stop` fails due to prior interrupt state.

This preserves the SIGINT assertions while ensuring Chrome exits before `t.TempDir()` cleanup.

**Alternatives**:

- **Harden `browser.RemoveProfile`**: retry `RemoveAll` with short backoff after `KillByProfile` (product-wide; helps worker paths too, but does not replace test cleanup when stop was never called for id2).
- **Separate temp root for profiles**: use a non-`TempDir` directory cleaned in `t.Cleanup` with explicit logic (more complex, weaker than stopping browsers).
- **Skip on CI** (rejected): spec task T011 requires this test in default integration runs.

**Files likely to change**:

- `tests/integration/stop_all_test.go`
- Possibly `tests/integration/start_worker_test.go` or a small `cleanupAllInstances(t, bin, env)` helper

**Tests to add or update**:

- Re-run `go test ./tests/integration/ -run TestStopAllSIGINT -count=20` on Linux+Chrome.
- Full `go test -json ./...` as in `.github/workflows/default-checks.yml`.

## Risks & Considerations

- Cleanup must not mask real SIGINT bugs (e.g. still assert interrupt message and `stoppedCount >= 1 **before** cleanup runs).
- `stop all` in cleanup after an interrupted bulk command should be idempotent (empty registry → exit 0 per FR-003).
- Parallel integration tests share one `supervisor` binary build (`binOnce`); profile cleanup is per-test via isolated env (no cross-test leak if each test cleans its own temp root).

## Open Questions

- [NEEDS CLARIFICATION: Does `main` without the Aviator PR also fail `TestStopAllSIGINT` on current `ubuntu-latest`, or did longer Aviator timeouts correlate only by ordering/timing? Both failing runs were on `fix/aviator-exiting` only in sampled history.]
- [NEEDS CLARIFICATION: Should product `runStopAll` register deferred cleanup for profiles when exiting via interrupt (operator UX), or is test-only teardown sufficient?]