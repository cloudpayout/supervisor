# Research: Record One Test Result on Each Change

Decisions for one visible result. The spec removes any hold on the merge and any extra approval when the check changes. Earlier research that chose a required status check or a code-owner review of the workflow file is not in force.

## 1. Where the check runs

**Decision**: One job on GitHub-hosted `ubuntu-latest`, which GitHub documents as Linux `ubuntu-24.04` (standard x64, 4 CPU, 16 GB RAM, 14 GB storage).

**Rationale**: The spec requires the hosted Linux environment selected for this feature. The assessment selected the current hosted Ubuntu label, and that label is Linux. A standard runner is outside the 15-minute cap that applies to single-CPU runners.

**Alternatives considered**: A privately operated runner, macOS, and Windows. The spec puts all three out of scope. A single-CPU runner was rejected because of the 15-minute job cap.

## 2. Which Go the job uses

**Decision**: `actions/setup-go` with `go-version-file: go.mod`. Today that file declares `go 1.27.1`. The action leaves `go` on `PATH` for later steps, including the `go build` the integration harness runs.

**Rationale**: The constitution requires the Go version in `go.mod` (1.27.1 at ratification). Reading the file at job time follows that file if a later spec amends the version. The public versions manifest lists 1.27.1 as stable.

**Alternatives considered**: A hardcoded `go-version: '1.27.1'` string. It would drift from `go.mod`. The starter workflow's `go-version: '1.20'` is an example pin, not this module's version.

## 3. How many times the checks run

**Decision**: One `go test -json ./...` with no `-tags`. A newer commit on the same ref cancels the in-progress job. The replacement job checks the new revision once.

**Rationale**: The spec says one execution is the success rule. A second execution is invalid. The ten-instance test is behind `//go:build slow` and stays out. There is no `//go:build external` test, and a live-casino check must not gate this run.

**Alternatives considered**: Repeating the suite before a result counts. The spec records that earlier assessment text as not in force.

## 4. What a skip means on the visible result

**Decision**: The job fails when `go test` exits non-zero or the JSON contains `"Action":"skip"`. The four browser checks are `TestParallelStartRegistry`, `TestStartStatusStop`, `TestStartRegistersAndReachesRunning`, and `TestStatusReflectsConcurrentChanges`. A real browser executable is installed on `PATH` under one of `google-chrome`, `google-chrome-stable`, `chromium`, `chromium-browser`, or `chrome`. An Ubuntu snap stub is not that executable.

**Rationale**: The spec marks a result successful only when the one execution passed and those checks finished as pass or fail. A skip is not a success. The product already discovers a browser by those names. Failing the job on any skip keeps a skipped browser check from appearing as a green result.

**Alternatives considered**: Treating a skip as success when the process exits 0. That is the local behavior the spec calls out as the baseline to leave behind.

## 5. What stays unchanged in the product

**Decision**: Do not change `AllocatorOptions`. `NoSandbox` stays off. If a recorded run still skips a browser check, a later task may adjust only the test that skipped. If the only remaining fix would change what an operator does with start, status, or stop, stop.

**Rationale**: chromedp's default allocator options already include headless mode and `disable-dev-shm-usage`. `NoSandbox` is a separate option and is not in those defaults. The spec forbids an operator-facing change and a second execution.

**Alternatives considered**: Adding `NoSandbox` so Chrome starts inside the runner. That would change browser flags the product uses for operators. The spec says to stop instead.

## 6. The job does not decide the merge

**Decision**: Do not mark job `default-checks` as a required status check. Do not add `.github/CODEOWNERS`. Do not add a ruleset that reviews edits to `.github/workflows/default-checks.yml`. The workflow file is still the check this feature adds, and a pull request may change it. Merge rights stay whatever the repository already grants. A code owner who can merge may merge a red job, a skipped job, a job that did not finish, or a pull request that rewrites the workflow.

**Rationale**: The spec says the recorded result leaves the merge decision with the people who already have it. It also says this feature adds no hold and no required approval when the check itself changes. A required check would hold the pull request. A code-owner rule on the workflow file would add an approval. Both contradict the updated spec.

**Alternatives considered**: Requiring the status check, and requiring a code owner to approve workflow edits. Those were the previous design. The spec now withdraws them. Leaving the job visible still gives the author one pass or fail to read before someone merges.

## 7. What the job is allowed to touch

**Decision**: `permissions: contents: read` only. No secrets. No `pull_request_target`. The workflow file begins with a comment that cites `specs/003-record-test-results/spec.md`.

**Rationale**: A fork pull request must not be able to change this repository or read secrets. The constitution says code that lands must cite the spec, and secrets must not be committed. `pull_request` runs the workflow from the pull request with the read-only token. `pull_request_target` would run a privileged workflow and is not used.

**Alternatives considered**: A write token, or `pull_request_target`, so a fork check could update the repository. The spec forbids that.

## 8. Out of scope

**Decision**: No second execution, no ten-instance run, no live-casino check, no macOS or Windows runner, no privately operated runner, no billing report, no new operator command, no required status check, and no approval rule for changes to the check.

**Rationale**: Each item is either withdrawn by the updated spec or already outside the default checks.
