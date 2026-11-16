# Bug Fix: Operator start argv capture flakes on CI

- **Slug**: operator-argv-capture-flake
- **Fixed**: 2026-11-16T04:52:32-06:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

`TestOneWinSecretStaysOffTheRecord` no longer reads `/proc` for the short-lived operator `start` process. It waits until the instance is running, then checks the worker command line for the session path. Operator rejection of `--session-file` stays in the existing usage tests.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `tests/integration/onewin_secrets_test.go` | modified | Drop `waitCmdlineContains`; assert worker argv after `waitRunning` |

## Diff Highlights (optional)

Operator `start` now runs to completion. The session path is checked on the worker, which is still alive:

```go
workerArgv := cmdline(t, inst.OwnerPID)
if !bytes.Contains(workerArgv, []byte("--session-file")) || !bytes.Contains(workerArgv, []byte(file)) {
    t.Fatalf("worker argv missing the session path: %q", workerArgv)
}
```

`--session-file` here is the hidden worker argument in `internal/cli/start.go`. It is not an operator `start` flag. Cookie-value checks on worker, browser, log, registry, screenshot cmdline, and filename are unchanged. Operator stdout is also checked for the cookie value.

## Tests Added or Updated

- `tests/integration/onewin_secrets_test.go` `TestOneWinSecretStaysOffTheRecord` — proves the positional path reached the worker and that the cookie value stays off the record. Operator `/proc` polling is gone.

## Local Verification

- Commands run: `go test -count=20 -timeout 15m -run TestOneWinSecretStaysOffTheRecord ./tests/integration/` → pass (`ok`, 23.042s).
- Commands run: `go test -count=1 -timeout 10m ./...` → pass (integration 22.539s, unit 0.086s).

## Deviations from Assessment

- Did not add the optional operator-argv goroutine. The operator process is too short to observe reliably, and operator `--session-file` is already a usage error covered by `tests/integration/onewin_usage_test.go`.
- The remaining `--session-file` assertion is on the worker command line only, matching the internal spawn in `internal/cli/start.go`. It does not claim the operator flag still exists.

## Follow-ups

- Re-run the `default-checks` workflow on the pull request.
