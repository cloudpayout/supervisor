# Bug Fix: Integration tests fail after 005 start-command-syntax

- **Slug**: integration-tests-005-start-syntax
- **Fixed**: 2026-11-16T04:27:49-06:00
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

Stabilized the two integration tests that failed in CI after the 005 start-command syntax. The operator CLI was left unchanged. The secrets test now waits briefly for `/proc` argv to be populated, and the headless-flag test stops each browser before starting the next one.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `tests/integration/onewin_secrets_test.go` | modified | Poll `/proc/<pid>/cmdline` up to 500ms for the session path |
| `tests/integration/start_flags_test.go` | modified | Stop each instance before the next start; headed session still uses `waitPresent` |

## Diff Highlights (optional)

Operator argv is no longer read once immediately after `Start()`:

```go
operatorArgv := waitCmdlineContains(t, startCmd.Process.Pid, file, 500*time.Millisecond)
```

`waitCmdlineContains` returns when cmdline contains the session path or its base name, and fails with the last cmdline if the budget expires without a match.

`TestStartRecordsHeadless` calls `stopID` after each assertion so blank, headless, and session starts do not keep Chrome processes alive together.

## Tests Added or Updated

- `tests/integration/onewin_secrets_test.go` `TestOneWinSecretStaysOffTheRecord` — still requires the positional session path on the operator argv, rejects `--session-file`, and rejects the cookie value. The read retries through the empty cmdline window between fork and exec.
- `tests/integration/start_flags_test.go` `TestStartRecordsHeadless` — same headless-field assertions, with instances stopped between cases so a later headed session start is not competing with earlier browsers.

## Local Verification

- Commands run: `go test -count=10 -timeout 8m -run 'TestOneWinSecretStaysOffTheRecord|TestStartRecordsHeadless' ./tests/integration/` → pass (`ok`, 27.431s).
- An earlier `-count=10` run failed all ten times on `TestStartRecordsHeadless` when the headed session case used `waitRunning` (`instance … did not reach running`). That assertion was reverted before the passing run. See deviations.

## Deviations from Assessment

- Headed `start <file>` stays on `waitPresent`, not `waitRunning`. Locally that start does not reach `running` within the existing wait (headed Chrome without a display). The registry flag is written before the worker runs, and the assessment allows asserting the row is present. Each prior instance is still stopped before the next start, which is the contention fix.
- `waitCmdlineContains` accepts the session path or its base name, capped at 500ms. A missing path still fails the test.

## Follow-ups

- Re-run the `default-checks` workflow on the pull request. These tests were flaky under CI load and passed locally before the change as well.
