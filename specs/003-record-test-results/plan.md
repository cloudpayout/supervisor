# Implementation Plan: Record One Test Result on Each Change

**Branch**: `003-record-test-results` | **Date**: 2026-11-13 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/003-record-test-results/spec.md`

## Summary

Every push and every pull request, including forks and any branch, gets one recorded result for the project's default checks. The result is a single GitHub-hosted job, `default-checks`, on `ubuntu-latest` (Linux, `ubuntu-24.04`). The job installs Go from `go.mod` (1.27.1 today) and a real browser the product already knows how to find, then runs `go test -json ./...` once on that checkout. A skip or a failed execution fails the job, so the author sees a result that is not successful. A second execution is not added. The job is information on the change. This feature does not require the job as a status check, does not add a code-owner review, and does not otherwise hold a merge. A code owner who can already merge may merge a failure, a skip, a job that did not finish, or a pull request that changes the workflow file. The job requests read-only contents permission and no secrets. The workflow file cites this spec. Product behavior is unchanged unless a recorded run shows a browser check skipped, and even then only that skip may be fixed.

## Technical Context

**Language/Version**: Go 1.27.1, taken from `go.mod` at job time (`actions/setup-go` `go-version-file`)

**Primary Dependencies**: `actions/checkout`, `actions/setup-go`, a browser binary named `google-chrome`, `google-chrome-stable`, `chromium`, `chromium-browser`, or `chrome`. Existing module dependencies are unchanged (`chromedp` v0.23.0, `cobra`, `gopsutil`).

**Storage**: N/A. The result is the workflow job GitHub shows on the change. See [data-model.md](./data-model.md).

**Testing**: `go test -json ./...` once, no build tags. Any `"Action":"skip"` fails that execution. The `slow` ten-instance test is not part of the run. A second execution is out of scope.

**Target Platform**: GitHub-hosted `ubuntu-latest` (Linux x64, 4 CPU, 16 GB RAM). macOS, Windows, and privately operated runners are out of scope.

**Project Type**: CI check added to the existing CLI repository. No new operator command. No repository ruleset.

**Performance Goals**: No tighter time budget than the spec gives. The job is one default-suite run on a standard hosted runner, not a single-CPU runner (those are capped at 15 minutes).

**Constraints**: Read-only token, no secrets, no `pull_request_target`, no product sandbox change (`NoSandbox` stays off), no second test execution. The job must not be a required status check, and this feature must not add an approval rule for edits to the workflow file.

**Scale/Scope**: One workflow, one job, one execution, every push and every pull request. The merge decision stays with the people who already have it.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec Before Implementation | Pass | Spec `003-record-test-results` is the source. The workflow file must cite that spec. This plan does not start implementation. |
| II. CLI as the Operator Interface | Pass | No new operator surface. The check runs the existing CLI tests. |
| III. Multiple Browser Instances | Pass | Tests keep using chromedp. Allocator options are not changed. |
| IV. Supervisor Owns the Run | Pass | No change to who starts, monitors, or stops a run. |
| V. Named Casinos and Games Only | Pass | Default `go test ./...` has no external casino check. The ten-instance check stays out. |
| Technical constraints | Pass | Go comes from `go.mod` (1.27.1). Module path unchanged. No secrets committed. Feature directory is `003`. |
| Delivery workflow | Pass | Plan follows the spec. The check runs the start, status, and stop tests that already exist, without a live casino. |

**Post–Phase 1 re-check**: Pass. The contract, data model, and quickstart add one read-only job whose result is visible and does not decide the merge. They do not add a required status check, a code-owner review, a second execution, a second operator interface, a shared browser profile, a casino check, or a secret in the repository.

## Project Structure

### Documentation (this feature)

```text
specs/003-record-test-results/
├── plan.md              # This file
├── research.md          # Phase 0
├── data-model.md        # Phase 1
├── quickstart.md        # Phase 1
├── contracts/
│   └── check-result.md  # Phase 1
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
.github/
└── workflows/
    └── default-checks.yml   # new: the visible check

tests/
├── integration/             # unchanged unless a recorded run skips
└── unit/

cmd/supervisor/              # unchanged
internal/                    # unchanged
```

**Structure Decision**: The repository stays a single Go CLI module. This feature adds one workflow file. It does not add `src/`, a service, a second test tree, a CODEOWNERS rule, or a repository ruleset. Test or product edits are not part of the structure unless the stop rule in [quickstart.md](./quickstart.md) says a recorded run requires a test-only fix. A second execution is not part of the structure.

## Complexity Tracking

No constitution violations. Nothing to justify.
