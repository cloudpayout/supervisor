# Data Model: Check Result

One check result belongs to one revision of one change. A push and a pull request for the same revision each have their own result. The repository does not store a new database. The result is the workflow job GitHub shows on that change. The result does not decide whether the change can be merged.

## Check result

| Field | Meaning | Rule |
|-------|---------|------|
| Revision | The commit the job checked out | The single execution uses this checkout. A newer commit on the ref cancels an in-progress result for the older commit. |
| Event | `push` or `pull_request` | Every push and every pull request gets a result, including forks and branches other than the default branch. |
| Execution | The one run of the default checks | `passed` or `failed`. There is no second execution. |
| Browser checks | The four default checks that skip when no browser is installed | Each is `passed`, `failed`, `skipped`, or `not_run`. They are `TestParallelStartRegistry`, `TestStartStatusStop`, `TestStartRegistersAndReachesRunning`, and `TestStatusReflectsConcurrentChanges`. |
| Overall | What the author reads | See states below. |
| Decides merge | Whether this result changes who may merge | Always no. A code owner who can already merge may merge while Overall is `successful`, `not_successful`, or `did_not_run`, including when the pull request changes the check. |

## States

```text
did_not_run ──job finishes──► not_successful
     │
     └──────── one execution passed and zero browser checks skipped ──► successful
```

| Overall | When | Merge |
|---------|------|-------|
| `did_not_run` | The job has not started, is still running, was cancelled, or did not finish | Unchanged. A code owner may merge. |
| `not_successful` | The execution failed, or any of the four browser checks was skipped | Unchanged. A code owner may merge. The result stays visible. |
| `successful` | The one execution passed, and none of the four browser checks was skipped | Unchanged. The result adds no merge step. |

A skip is never folded into `successful`. The optional ten-instance check and any live-casino check are not fields on this result. This result has no review field. A change to the check does not create a pending approval.

## Validation

- Execution count is 0 (`did_not_run`) or 1. A second execution is invalid.
- `successful` requires that one execution `passed` and all four browser checks `passed` or `failed` (not `skipped`, not `not_run`).
- No secret, credential, or write permission is part of the result.
- The result includes no required status check and no required approval.
