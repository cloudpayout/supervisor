---
description: "Task list for one visible default-checks result on each change"
---

# Tasks: Record One Test Result on Each Change

**Input**: Design documents from `specs/003-record-test-results/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/, quickstart.md

**Tests**: No new test files. The feature runs the existing default suite once. Do not add `-tags=slow`, a live-casino check, or a second execution.

**Organization**: Tasks are grouped by user story so each story can be implemented and checked on its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

One workflow file at `.github/workflows/default-checks.yml`. Do not add `.github/CODEOWNERS` or a repository ruleset. The existing Go module stays at `cmd/supervisor/`, `internal/`, and `tests/`. Do not add `src/`.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the workflow directory the plan calls for

- [x] T001 Create the directory `.github/workflows/` at the repository root per `specs/003-record-test-results/plan.md`. Do not add packages under `cmd/` or `internal/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The job identity, triggers, permissions, cancellation, and spec citation that every story uses

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Create `.github/workflows/default-checks.yml` with workflow name `Default checks`, job id `default-checks`, and `runs-on: ubuntu-latest`. Begin the file with a comment that cites `specs/003-record-test-results/spec.md`. Trigger on `push` and `pull_request` with no `branches` filter and no `paths` filter. Set `permissions: contents: read` and leave every other permission unset. Do not reference secrets and do not use `pull_request_target`. Cancel an in-progress run when a newer commit arrives on the same ref. Constraints from `specs/003-record-test-results/data-model.md`: "The single execution uses this checkout. A newer commit on the ref cancels an in-progress result for the older commit." and "No secret, credential, or write permission is part of the result."

**Checkpoint**: The job exists, is read-only, cites the spec, and will run for every push and pull request. It does not install Go or run tests yet

---

## Phase 3: User Story 1 - See whether checks passed (Priority: P1) 🎯 MVP

**Goal**: An author can open a push or a pull request, including a fork and any branch, and see one pass or fail for the default checks without running them locally

**Independent Test**: Push a change on a branch that is not the default branch, and open a pull request from a fork. Each shows one recorded pass or fail for the default checks, without anyone running those checks locally

### Implementation for User Story 1

- [x] T003 [US1] In `.github/workflows/default-checks.yml`, check out the revision that triggered the job with `actions/checkout`. Constraint from `specs/003-record-test-results/data-model.md`: "The single execution uses this checkout. A newer commit on the ref cancels an in-progress result for the older commit."
- [x] T004 [US1] In `.github/workflows/default-checks.yml`, install Go with `actions/setup-go` and `go-version-file: go.mod` so `go` stays on `PATH` for the whole job, including the `go build` the integration harness runs. `go.mod` declares `go 1.27.1`
- [x] T005 [US1] In `.github/workflows/default-checks.yml`, run `go test -json ./...` once with no `-tags`. Do not pass `-tags=slow` and do not add a second test step. The job conclusion is the recorded pass or fail. Constraints from `specs/003-record-test-results/data-model.md`: "`passed` or `failed`. There is no second execution." "Every push and every pull request gets a result, including forks and branches other than the default branch." "Execution count is 0 (`did_not_run`) or 1. A second execution is invalid."

**Checkpoint**: User Story 1 is done when a push and a fork pull request each show one run of job `default-checks`. A skip may still exit 0; User Story 2 closes that gap

---

## Phase 4: User Story 2 - Trust a successful result (Priority: P2)

**Goal**: The job is successful only when the one execution passed and none of the four browser checks were skipped. That label does not decide the merge

**Independent Test**: Run the default checks once with a browser available, and once with it unavailable. Only the run where the checks passed and no browser-dependent check was skipped is recorded as success

### Implementation for User Story 2

- [x] T006 [US2] In `.github/workflows/default-checks.yml`, before the test step, install a real browser executable on `PATH` under one of these names: `google-chrome`, `google-chrome-stable`, `chromium`, `chromium-browser`, or `chrome`. Do not use an Ubuntu snap stub. Do not edit `internal/browser/chromedp.go`
- [x] T007 [US2] In `.github/workflows/default-checks.yml`, fail the single execution when the `go test -json` output contains `"Action":"skip"`. The four browser checks are `TestParallelStartRegistry`, `TestStartStatusStop`, `TestStartRegistersAndReachesRunning`, and `TestStatusReflectsConcurrentChanges`. Constraints from `specs/003-record-test-results/data-model.md`: "Each is `passed`, `failed`, `skipped`, or `not_run`." "`successful` requires that one execution `passed` and all four browser checks `passed` or `failed` (not `skipped`, not `not_run`)."

**Checkpoint**: User Story 2 is done when a green job means one passing execution and zero skips, and a skip fails the job

---

## Phase 5: User Story 3 - Leave the merge to the code owner (Priority: P3)

**Goal**: The result stays visible and does not decide the merge. A code owner who can already merge may merge a failure, a skip, a job that did not finish, or a pull request that changes the check. This feature adds no hold and no required approval

**Independent Test**: Open a pull request whose check failed, one whose browser-dependent check was skipped, and one whose check never completed. The failure and the skip stay visible as not successful. The check that never completed stays visible as did not run. A code owner merges each one. Open a pull request that changes the check itself. A code owner merges it without a separate approval created by this feature

### Implementation for User Story 3

- [x] T008 [US3] In `.github/workflows/default-checks.yml`, leave the job as a visible result only. Do not add an environment protection rule, a deployment gate, or any other setting in that file that would hold a pull request. A failed or skipped test still fails the job so the author can read it. Constraints from `specs/003-record-test-results/data-model.md`: "Always no. A code owner who can already merge may merge while Overall is `successful`, `not_successful`, or `did_not_run`, including when the pull request changes the check." "The result includes no required status check and no required approval."
- [x] T009 [P] [US3] Do not create `.github/CODEOWNERS`, and do not add a ruleset that requires job `default-checks` or a review of `.github/workflows/default-checks.yml`. The only file this feature adds under `.github/` is `.github/workflows/default-checks.yml`. Constraint from `specs/003-record-test-results/data-model.md`: "This result has no review field. A change to the check does not create a pending approval."

**Checkpoint**: User Story 3 is done when a code owner can merge a failed, skipped, unfinished, or rewritten check, and the result stays visible on the change

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Confirm the contract and stop if the suite cannot run without a product change

- [x] T010 Review `.github/workflows/default-checks.yml` and confirm the opening comment cites `specs/003-record-test-results/spec.md`, `pull_request_target` is absent, the only permission is `contents: read`, no secret name appears, the file contains one `go test -json ./...` step, and the file does not declare a required status check
- [x] T011 Run scenarios A through F in `specs/003-record-test-results/quickstart.md` against `.github/workflows/default-checks.yml`. Confirm `.github/CODEOWNERS` was not added by this feature
- [x] T012 If a run of `.github/workflows/default-checks.yml` records `"Action":"skip"` for `TestParallelStartRegistry`, `TestStartStatusStop`, `TestStartRegistersAndReachesRunning`, or `TestStatusReflectsConcurrentChanges`, edit only that test file under `tests/` so the check runs. If the only remaining fix would change what an operator does with start, status, or stop, stop without editing `internal/` or `cmd/`, without adding `NoSandbox`, and without adding a second test execution. Follow the Stop section of `specs/003-record-test-results/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup. Blocks all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational. Edits `.github/workflows/default-checks.yml`
- **User Story 2 (Phase 4)**: Depends on User Story 1 because it extends the same test step in `.github/workflows/default-checks.yml`
- **User Story 3 (Phase 5)**: T008 edits the same workflow file, so it follows T007. T009 can start after Foundational because it adds no file
- **Polish (Phase 6)**: Depends on the stories you intend to ship

### User Story Dependencies

- **User Story 1 (P1)**: Starts after Foundational. No dependency on US2 or US3
- **User Story 2 (P2)**: Starts after US1. Same workflow file, so it cannot be written in parallel with US1
- **User Story 3 (P3)**: The absence of a CODEOWNERS file does not depend on the test steps. The workflow reminder in T008 follows the finished job

### Within Each User Story

- T003 before T004 before T005 (checkout, then Go, then the one test that needs both)
- T006 before T007 (browser present, then a skip fails that same execution)
- T009 does not wait on T008
- No new test files come first. This feature does not use a red-then-green test task

### Parallel Opportunities

- T001 is the only setup task
- T002 is the only foundational task
- T003, T004, T005, T006, T007, and T008 all edit `.github/workflows/default-checks.yml` and stay sequential
- T009 [P] adds no file, so it can proceed beside US1 after T002
- T010 and T011 wait until T008 and T009 are done

---

## Parallel Example: User Story 3 beside User Story 1

```bash
# After T002, different files:
Task: "T003–T005 Add checkout, Go, and one go test run in .github/workflows/default-checks.yml"
Task: "T009 Do not create .github/CODEOWNERS or a required check for job default-checks"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: A push and a fork pull request each show one run of job `default-checks` without a local test run
5. Demo that recorded pass or fail

### Incremental Delivery

1. Setup + Foundational → the read-only job is defined and cites the spec
2. User Story 1 → authors can see one result (MVP)
3. User Story 2 → a green result means one pass and zero skips
4. User Story 3 → the result stays visible and a code owner can still merge
5. Polish → quickstart A–F, then stop if the suite needs an operator-facing change

### Parallel Team Strategy

1. One person completes T001 and T002
2. Then:
   - One person edits `.github/workflows/default-checks.yml` through US1, US2, and T008 (same file)
   - Another person confirms in T009 that this feature adds no CODEOWNERS file and no required check
3. A code owner can merge at every increment. The job does not decide that

---

## Notes

- [P] tasks use different files and do not wait on an unfinished edit of the same file
- [US1], [US2], and [US3] map to the stories in `specs/003-record-test-results/spec.md`
- User Story 1 is independently shippable. User Story 2 is the point where a skip stops counting as success
- Do not add a second `go test` step. The data model says a second execution is invalid
- Do not add a required status check or `.github/CODEOWNERS`. The data model says the result includes no required status check and no required approval
- A recorded skip of one of the four browser checks may be fixed only in `tests/`. Stop when the only remaining fix would change what an operator does with start, status, or stop
- Commit after each task or logical group

## Phase 7: Convergence

- [x] T013 Set `cache: false` on `actions/setup-go` in `.github/workflows/default-checks.yml` so a cache save cannot fail the job under `permissions: contents: read` per FR-004 (partial)
