# Bug Fix: TestStopAllSIGINT TempDir cleanup after SIGINT

- **Slug**: tests-failling-aviator-exiting
- **Fixed**: 2026-11-22T22:50:00-07:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

`TestStopAllSIGINT` intentionally leaves a running headless instance after SIGINT, which left Chrome holding files under the test’s `t.TempDir()` profile tree. Added `t.Cleanup` plus `stopAllInstancesForTeardown` so leftover instances are stopped and profile directories are emptied before Go’s TempDir `RemoveAll` runs.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `tests/integration/stop_all_test.go` | modified | `t.Cleanup` on `TestStopAllSIGINT`; new `stopAllInstancesForTeardown` helper |

## Diff Highlights (optional)

`TestStopAllSIGINT` registers cleanup before starting instances:

```go
t.Cleanup(func() { stopAllInstancesForTeardown(t, bin, env, regPath) })
```

Teardown stops any registry rows, runs `stop all`, then polls until `profilesDir(env)` is empty (30s budget).

## Tests Added or Updated

- `tests/integration/stop_all_test.go::TestStopAllSIGINT` — same assertions; post-test teardown prevents `testing.go:1617` TempDir failures

## Local Verification

- Commands run: `go test ./tests/integration/ -run TestStopAllSIGINT -count=5 -timeout 3m` → **pass** (~29s)
- Manual checks: none

## Deviations from Assessment

None. Helper lives in `stop_all_test.go` rather than `start_worker_test.go` to keep interrupt teardown scoped to stop-all tests.

## Follow-ups

- Run full `go test -json ./...` on CI (ubuntu-latest + Chrome) to confirm default checks green.
- Optional: reuse `stopAllInstancesForTeardown` if other tests leave running browsers under `isolatedEnv`.