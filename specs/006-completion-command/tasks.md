---
description: "Task list for removing Cobra default completion command"
---

# Tasks: Remove Shell Completion Command

**Input**: Design documents from `specs/006-completion-command/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: FR-006 and SC-001/SC-002 require automated checks. Integration tests in `tests/integration/root_help_test.go` are written first and should fail until `internal/cli/root.go` is updated (T004). SC-003 is satisfied by full `go test ./...` (T006).

**Organization**: Tasks are grouped by user story so each story can be implemented and tested on its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Product code: `cmd/supervisor/`, `internal/cli/`
- Tests: `tests/integration/` using `supervisorBin`, `isolatedEnv`, and `exitCode` from `tests/integration/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and toolchain; no new dependencies.

- [X] T001 Confirm branch `fix/remove-completion-command`, feature dir `specs/006-completion-command/`, and leave `go.mod` on module `github.com/cloudpayout/supervisor` Go 1.27.1 with no new `require` entries per [plan.md](./plan.md).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Failing integration tests that encode the 006 CLI contract before the one-line Cobra change.

**⚠️ CRITICAL**: Do not set `DisableDefaultCmd` until T002 and T003 exist and fail for the expected reasons.

- [X] T002 [P] [US1] Add `TestRootHelpNoCompletion` in `tests/integration/root_help_test.go` using `supervisorBin(t)` and `isolatedEnv(t)`. Run `help` and `--help`. Assert stdout lists `start`, `status`, `stop`, `screenshot`, and `help`. Assert stdout does **not** contain a line advertising `completion` (FR-002, SC-001). Assert `worker` is not listed (data-model: `hidden_commands` MUST include `worker`; MUST NOT appear in root help).

- [X] T003 [P] [US2] In `tests/integration/root_help_test.go`, add `TestCompletionUnknownCommand`. For each of `completion`, `completion bash`, and `help completion`: assert exit code `2` via `exitCode(err)`, empty stdout, and stderr containing `unknown command` (FR-003, FR-004, SC-002, [contracts/cli.md](./contracts/cli.md)).

**Checkpoint**: `go test ./tests/integration/ -run RootHelp -v` fails on missing `DisableDefaultCmd`.

---

## Phase 3: User Story 1 - See only supported operator commands (Priority: P1) 🎯 MVP

**Goal**: Root help exposes only operator commands plus `help`; no `completion`.

**Independent Test**: `supervisor help` omits `completion` and includes `start`, `status`, `stop`, `screenshot`.

### Implementation for User Story 1

- [X] T004 [US1] In `internal/cli/root.go` `newRoot()`, after constructing `root := &cobra.Command{...}` and before `AddCommand` calls, set `root.CompletionOptions.DisableDefaultCmd = true` per [research.md](./research.md#r2-mechanism-to-remove-the-command) (FR-001).

**Checkpoint**: `TestRootHelpNoCompletion` passes. MVP for discovery-only operators is done.

---

## Phase 4: User Story 2 - Completion invocation is rejected (Priority: P1)

**Goal**: `supervisor completion` and variants behave as unknown commands with exit `2` and no script output.

**Independent Test**: `TestCompletionUnknownCommand` passes after T004.

### Implementation for User Story 2

- [X] T005 [US2] No separate code task: unknown-command behavior for `completion` is delivered by T004 via existing `Execute()` and `isUsageError` in `internal/cli/root.go`. Run `go test ./tests/integration/ -run TestCompletionUnknownCommand -v` after T004; fix only if tests fail.

**Checkpoint**: User Stories 1 and 2 both satisfied.

---

## Phase 5: User Story 3 - Lifecycle commands remain unaffected (Priority: P2)

**Goal**: `start`, `status`, `stop`, `screenshot`, and hidden `worker` unchanged except absence of `completion`.

**Independent Test**: Full default test suite passes (SC-003).

### Tests for User Story 3

- [X] T006 [US3] Run `go test ./...` from repository root. All packages pass with no changes required to `internal/cli/start.go`, `internal/cli/status.go`, `internal/cli/stop.go`, `internal/cli/screenshot.go`, or `internal/cli/worker.go` unless a failure proves a regression (FR-005, SC-003).

**Checkpoint**: Ordinary confirmation and existing integration suites green.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Docs and manual validation per spec edge cases.

- [X] T007 [P] Search the repository for operator-facing `supervisor completion` instructions (e.g. `README.md`, `specs/*/quickstart.md` outside 006). Remove or correct any that tell operators to run completion (spec edge case: documentation outside binary).

- [X] T008 Run the commands in [quickstart.md](./quickstart.md) against a built binary and confirm outcomes match the guide.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Phase 1 — **blocks** T004
- **User Story 1 (Phase 3)**: Depends on Phase 2 tests being present
- **User Story 2 (Phase 4)**: Depends on T004
- **User Story 3 (Phase 5)**: Depends on T004
- **Polish (Phase 6)**: Depends on T006 passing

### User Story Dependencies

| Story | Priority | Depends on | Independent test |
|-------|----------|------------|------------------|
| US1 | P1 | T002, T004 | `TestRootHelpNoCompletion` |
| US2 | P1 | T003, T004 | `TestCompletionUnknownCommand` |
| US3 | P2 | T004 | `go test ./...` |

US1 and US2 share one implementation task (T004). US3 is regression-only.

### Parallel Opportunities

- **T002** and **T003** can run in parallel (same new file, different test functions — coordinate one PR/file).
- **T007** can run in parallel with **T006** after T004 lands.

### Parallel Example: Foundational tests

```bash
# Add both tests in root_help_test.go, then run:
go test ./tests/integration/ -run 'TestRootHelpNoCompletion|TestCompletionUnknownCommand' -v
# Expect FAIL before T004, PASS after T004
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. T001 → T002 → T004 → validate `TestRootHelpNoCompletion`
2. Stop if only help hygiene is needed for a quick merge; add T003/T005/T006 before merge per FR-006

### Incremental Delivery

1. Phase 1–2: Contract tests (fail)
2. T004: Enable US1 + US2
3. T005–T006: Confirm rejection + no regression
4. T007–T008: Docs and quickstart

### Suggested merge order

Single small PR: T001 through T008 in one branch `fix/remove-completion-command`.

---

## Notes

- Cite spec scenarios in commit messages (e.g. FR-002 / User Story 1).
- Do not add custom `completion` command or operator-facing shell docs.
- `excluded_commands` MUST include `completion`; do not register it (data-model validation rules).

---

## Phase 7: Convergence

**Purpose**: Close doc and test gaps between artifacts and current behavior (FR-004 `help completion` semantics; FR-002 no-arg root help).

- [X] T009 Update `specs/006-completion-command/contracts/cli.md` and `specs/006-completion-command/quickstart.md` so `supervisor help completion` documents exit `0`, empty stdout, and stderr `Unknown help topic` (not exit `2` / `unknown command`), consistent with FR-004 and `tests/integration/root_help_test.go` (partial)

- [X] T010 [P] Extend `TestRootHelpNoCompletion` in `tests/integration/root_help_test.go` to run `supervisor` with no arguments and assert the same operator command list rules as `help` and `--help` per FR-002 and US1/AC1 (partial)