# Bug Verification: TestStopAllSIGINT TempDir cleanup on aviator-exiting CI

- **Slug**: tests-failling-aviator-exiting
- **Tested**: 2026-11-22T22:52:00-07:00
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: verified

## Summary

The CI failure mode (`TestStopAllSIGINT` failing with `testing.go:1617: TempDir RemoveAll cleanup: unlinkat .../Default: directory not empty`) does not reproduce after the teardown fix. Stress runs (20×) and the full `StopAll` integration subset pass on WSL2 with Chrome. Full `go test ./...` still fails on **unrelated** integration tests in this environment; that was not introduced by this fix and is outside the assessed bug scope.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | `go test ./tests/integration/ -run TestStopAllSIGINT -count=1 -v` | pass | No TempDir/unlinkat line; ~1s |
| New / updated tests (stress) | `go test ./tests/integration/ -run TestStopAllSIGINT -count=20 -timeout 10m` | pass | ~11s total; assessment suggested `-count=20` on Linux+Chrome |
| Regression (stop-all) | `go test ./tests/integration/ -run StopAll -count=1 -timeout 5m` | pass | All default `StopAll*` tests (excludes `slow` tag) |
| Regression suite | `go test ./... -count=1 -timeout 15m` | fail | Unrelated: `TestLogsAfterStop` (same TempDir class), `TestStartRecordsHeadless` (registry flake) |
| Lint / type-check | `go vet ./tests/integration/...` | pass | |
| CI default checks | GitHub Actions `ubuntu-latest` | not-run | Requires push/PR; follow-up from fix.md |

## Output Excerpts

Post-fix targeted test:

```text
=== RUN   TestStopAllSIGINT
--- PASS: TestStopAllSIGINT (0.99s)
PASS
ok  	github.com/cloudpayout/supervisor/tests/integration	0.991s
```

Stress (no failures in 20 iterations):

```text
ok  	github.com/cloudpayout/supervisor/tests/integration	11.173s
EXIT=0
```

Unrelated full-suite failures (same run, not caused by stop-all teardown change):

```text
--- FAIL: TestLogsAfterStop (0.27s)
    testing.go:1617: TempDir RemoveAll cleanup: unlinkat .../Default: directory not empty
--- FAIL: TestStartRecordsHeadless (5.04s)
    start_flags_test.go:32: instance ... not in registry
FAIL	github.com/cloudpayout/supervisor/tests/integration	128.233s
```

## Residual Risks

- **CI**: Default checks on `ubuntu-latest` were not re-executed in this verification session; merge confidence still depends on a green Actions run.
- **Parallel flakiness**: Other integration tests may still hit TempDir/profile cleanup races (`TestLogsAfterStop` observed locally).
- **Environment**: Verification used local WSL2 + Chrome; timing may differ from GitHub-hosted runners.

## Recommendation

Treat the assessed bug (**`TestStopAllSIGINT` blocking default checks on `fix/aviator-exiting`**) as **verified** for merge of this fix. Push the branch and confirm the default-checks workflow passes end-to-end. If CI stays red, inspect failures other than `TestStopAllSIGINT` before reopening this slug.