# Quickstart: Record One Test Result on Each Change

Validate the visible job in [contracts/check-result.md](./contracts/check-result.md). Fields and states are in [data-model.md](./data-model.md). These scenarios check the workflow file and the absence of a merge rule. They do not add a second test execution.

## Prerequisites

- The workflow file `.github/workflows/default-checks.yml` is on the revision under test.
- The repository has no required status check for job `default-checks` and no CODEOWNERS rule added by this feature.
- A code owner already has permission to merge. This feature does not grant or remove that permission.
- Go on the runner comes from `go.mod`. A browser binary is installed under one of the names in the contract.

## Scenario A — One result on a branch that is not the default

1. Push a commit to a branch other than the default branch.
2. Open the check run for that commit.

**Expected**: One run of job `default-checks`. The log contains one `go test -json ./...` and no `-tags=slow`. The log does not name `TestStatusTenInstances`. The run is not followed by a second test execution.

## Scenario B — A successful result adds no merge step

1. Open a pull request whose job succeeded and whose JSON has no `"Action":"skip"`.
2. Confirm `.github/workflows/default-checks.yml` may have changed or may be unchanged. Either way, look for an approval that this feature added.

**Expected**: Overall is `successful`. The pull request has no required check and no extra approval created by this feature. A code owner can merge it.

## Scenario C — A failure, a skip, or a missing job stays visible

1. Open three pull requests: one whose `go test` exits non-zero, one whose JSON contains `"Action":"skip"`, and one whose job was cancelled or never finished.
2. Read each result, then merge each pull request as a code owner.

**Expected**: The first two are `not_successful`. The third is `did_not_run`. Each result stays visible after the merge. Each merge completes. The log of a finished job shows one execution.

## Scenario D — A fork pull request cannot change the repository

1. Open a pull request from a fork.
2. Read the job's permissions and the checkout.

**Expected**: `contents: read`. No secret names. `pull_request_target` is absent. The job cannot push to this repository. One result is visible on the pull request.

## Scenario E — A new commit replaces the in-progress result

1. Push a commit that starts job `default-checks`.
2. Push a second commit to the same ref before the first job finishes.

**Expected**: The first job is cancelled. The second job checks out the second commit and runs `go test -json ./...` once.

## Scenario F — Changing the check does not add an approval

1. Open a pull request that replaces the test step in `.github/workflows/default-checks.yml`.
2. Look for a required review of that file, then merge as a code owner.

**Expected**: This feature adds no CODEOWNERS entry and no required review of the workflow file. The code owner can merge. The job on that pull request shows whatever the edited workflow ran. The result does not have to be `successful` for the merge to complete.

## Stop

If the browser checks cannot finish without a change to what an operator does with start, status, or stop, stop. Do not edit `internal/` or `cmd/` to force the job green. Do not add `NoSandbox`. Do not add a second execution or the ten-instance check.
