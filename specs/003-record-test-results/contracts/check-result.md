# Contract: Default checks result

**Feature**: `specs/003-record-test-results`  
**Workflow name**: `Default checks`  
**Job id**: `default-checks`  
**Workflow file**: `.github/workflows/default-checks.yml`

Authors read this job on the change. The job does not decide the merge. The fields and states are in [data-model.md](../data-model.md).

## When it runs

| Trigger | Scope |
|---------|--------|
| `push` | Every branch. No path filter. |
| `pull_request` | Every pull request, including those opened from a fork. No path filter. |

A new commit on the same ref cancels the in-progress job for that ref. The replacement job checks the new revision once.

## Permissions

| Permission | Value |
|------------|--------|
| `contents` | `read` |
| Every other permission | unset |
| Secrets | none |
| `pull_request_target` | not used |

A fork pull request cannot push to this repository and cannot read secrets. The job must not request a write token. The job itself is the result the author reads.

The workflow file begins with a comment that cites `specs/003-record-test-results/spec.md`.

## What the job does

1. Check out the revision that triggered the job.
2. Install the Go version declared in `go.mod` (1.27.1 today) and leave `go` on `PATH`.
3. Install a real browser executable on `PATH` under one of these names: `google-chrome`, `google-chrome-stable`, `chromium`, `chromium-browser`, `chrome`.
4. Run `go test -json ./...` once, with no extra build tags. Fail the job if the command exits non-zero or the JSON contains `"Action":"skip"`.

There is no second test execution. `go test ./...` does not include the `slow` ten-instance check and does not include an external casino check. Product browser flags are unchanged.

## Outcomes the author sees

| Job | Overall | Merge |
|-----|---------|-------|
| Job queued, running, cancelled, or absent | `did_not_run` | Unchanged. A code owner may merge. |
| Job failed, or the JSON contains a skip | `not_successful` | Unchanged. The result stays visible. A code owner may merge. |
| Job succeeded with no skip | `successful` | Unchanged. This job adds no merge step. |
| The workflow file changed | Whatever the job on that revision reports | Unchanged. This feature adds no approval that must happen first. |

## What this feature does not configure

The repository's existing permission to merge stays as it is. This contract does not add:

- A required status check for job `default-checks`.
- A CODEOWNERS rule.
- A review that must approve edits to `.github/workflows/default-checks.yml`.

## Out of contract

- A second execution of the default checks.
- A hold on the merge, or an approval created by this feature.
- The optional ten-instance check.
- Live casino or other external checks.
- macOS, Windows, and privately operated runners.
- Operator-facing changes to start, status, or stop.
- A billing report.
