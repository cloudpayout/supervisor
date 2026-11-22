---
description: "Task list for stop all CLI command"
---

# Tasks: Stop All Instances Command

**Input**: Design documents from `specs/011-stop-all-command/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/cli.md](./contracts/cli.md), [quickstart.md](./quickstart.md)

**Tests**: Constitution delivery requires CLI `stop` coverage. Integration tests in `tests/integration/stop_all_test.go` are written in Phase 2 and must fail until `internal/cli/stop.go` implements bulk mode (research R7).

**Organization**: Tasks grouped by user story; Phase 2 adds red integration tests before implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no unfinished dependencies)
- **[Story]**: US1, US2, US3
- Include exact file paths in descriptions

## Path Conventions

- Product code: `internal/cli/stop.go`
- Tests: `tests/integration/stop_all_test.go`, `internal/cli/stop_test.go` (unit), `tests/integration/stop_lifecycle_test.go`
- Reserved bulk token: exactly `all` (case-sensitive; FR-009; data-model validation table)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Align implementer with v11 contract before code changes.

- [X] T001 Confirm branch `011-stop-all-command`, feature dir `specs/011-stop-all-command/`, and review operator deltas in [contracts/cli.md](./contracts/cli.md) and [research.md](./research.md) (keyword dispatch, snapshot, empty branch in `runStopAll`, summary line, SIGINT exit `1`, operational failure R8).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Red integration tests for bulk stop. Do not ship implementation until these exist and fail for missing `stop all` behavior.

- [X] T002 [P] [US1] Add `tests/integration/stop_all_test.go` with `TestStopAllTwoInstances`: start two `--headless` instances in isolated env; run `supervisor stop all` with stdin from `os.DevNull` or pipe (no hang, FR-002); assert exit `0`, stderr contains both `stopped instance <id>` lines, summary `stop all: stopped 2, skipped 0 (already stopped or not found)`, and one-shot `status` shows `No instances are running.` (FR-001, FR-002, FR-004, FR-007; SC-001 proxy with two instances).

- [X] T003 [P] [US2] In `tests/integration/stop_all_test.go`, add `TestStopAllEmptyRegistry`: fresh state; `stop all` exits `0` within 3 seconds; stderr contains `No instances are running.` (FR-003, SC-002).

- [X] T004 [P] [US2] In `tests/integration/stop_all_test.go`, add `TestStopAllEmptyRegistryRepeated`: run `stop all` ten consecutive times on empty state; every run exits `0` with `No instances are running.` and identical registry JSON (SC-003).

- [X] T005 [P] [US2] In `tests/integration/stop_all_test.go`, add `TestStopAllAllStaleRows`: seed registry with dead-PID rows per pattern in `tests/integration/stop_lifecycle_test.go`; run `stop all`; exit `0`; stderr contains `instance already stopped:` lines; summary shows `skipped` ≥1 and `stopped` 0; one-shot `status` empty (User Story 2 scenario 2, FR-004).

- [X] T006 [P] [US3] In `tests/integration/stop_all_test.go`, add `TestStopHelpDocumentsAll`: `supervisor stop --help` mentions `all` and `<instance-id>`; does not mention interactive prompt (FR-010, SC-005).

- [X] T007 [P] [US1] In `tests/integration/stop_all_test.go`, add `TestStopAllUsageErrors`: `supervisor stop` (no args) exit `2`; `supervisor stop all extra` exit `2` (Cobra `ExactArgs(1)` receives `all` and `extra`); reserved token `All` → `instance not found: All` when no such row (FR-009, [contracts/cli.md](./contracts/cli.md) extra positionals).

- [X] T008 [P] [US3] Ensure `go test ./tests/integration/... -run StopLifecycle -count=1` passes before refactor; re-run after T013 and T019 to confirm single-ID contract (FR-008).

- [X] T009 [P] [US1] In `tests/integration/stop_all_test.go`, add `TestStopAllFiveInstances` with build tag `slow`: start five `--headless` instances; one `stop all`; assert all removed from status within 60s (SC-001). Default `go test ./...` excludes `slow`; quickstart Scenario 5 covers manual path.

- [X] T010 [P] [US1] In `tests/integration/stop_all_test.go`, add `TestStopAllOperationalFailureContinues`: start two `--headless` instances; implement bulk so `stopInstance` returns error when post-terminate `store.Update` fails; test uses read-only registry file or injected fault after first ID succeeds—second ID attempted, exit `1`, first ID remains removed from registry (FR-005, research R8). If injection is impractical in integration, add `internal/cli/stop_test.go` `TestRunStopAllAggregatesStopInstanceErrors` that drives `runStopAll` via a test hook or documents same loop semantics.

- [X] T011 [P] [US1] In `tests/integration/stop_all_test.go`, add `TestStopAllSIGINT`: start two `--headless` instances; run `stop all` and send `SIGINT` while command runs; assert exit `1`, interrupted summary on stderr, at least one instance fully stopped in registry (spec interrupt edge case; must not `t.Skip` in default integration run).

**Checkpoint**: `go test ./tests/integration/... -run 'StopAll' -v` fails on missing bulk implementation; `StopLifecycle` still passes.

---

## Phase 3: User Story 1 - Tear down every tracked instance (Priority: P1) 🎯 MVP

**Goal**: `supervisor stop all` stops every snapshotted registry ID using shared per-instance logic; summary and per-ID stderr; continue on operational failure (FR-001, FR-004–FR-007).

**Independent Test**: T002, T007 pass; T010 passes after error plumbing.

### Implementation for User Story 1

- [X] T012 [US1] Refactor `internal/cli/stop.go`: extract `stopInstance(id string) (outcome stopOutcome, err error)` with the same `store.Update` → `registry.Apply` → terminate → `browser.RemoveProfile` → row removal and stderr lines as today; return non-nil `err` on paths where single-ID `stop` would exit `1` (research R2, R8, FR-004).

- [X] T013 [US1] Add `const stopAllToken = "all"` and route in `newStopCmd` `RunE`: when `strings.TrimSpace(args[0]) == stopAllToken`, call `runStopAll()`; otherwise call `stopInstance` and map outcome to existing exit behavior (research R1, FR-008).

- [X] T014 [US1] Implement `runStopAll()` in `internal/cli/stop.go`: one `store.Update` with `registry.Apply`, snapshot instance IDs; **if `len(snapshot)==0`**, print `No instances are running.` on stderr and return `nil` (FR-003, merges former T011); else loop `stopInstance` for each ID; increment counters per [data-model.md](./data-model.md); print summary on stderr after loop (FR-007).

- [X] T015 [US1] In `runStopAll`, on non-nil `err` from `stopInstance`, record failure, continue remaining snapshot IDs, return aggregated error at end so exit is `1` (FR-005). On all outcomes skipped or stopped with no errors, return `nil` (FR-006).

**Checkpoint**: T002, T005, T007, T014 path pass. MVP bulk stop + empty + stale complete (T003–T005).

---

## Phase 4: User Story 3 - Discover the command and preserve single-ID stop (Priority: P2)

**Goal**: Help documents bulk form; single-ID path unchanged.

**Independent Test**: T006, T008 pass.

### Implementation for User Story 3

- [X] T016 [US3] Update `Use`, `Short`, and `Long` on `newStopCmd()` in `internal/cli/stop.go` to document `stop <instance-id|all>`, reserved token `all`, non-interactive behavior (FR-010).

- [X] T017 [US3] Verify `stopInstance` for IDs other than `all` preserves 002 exit codes and stderr; no stdin (FR-008); run T008 (`StopLifecycle`).

**Checkpoint**: Phase 2 tests green except T009 (`slow`), T010 (until error plumbing), T011 (until SIGINT).

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Interrupt handling, docs, full suite.

- [X] T018 In `runStopAll()` in `internal/cli/stop.go`, register `signal.Notify` on `os.Interrupt`; on SIGINT stop scheduling further IDs, finish current `stopInstance`, print `stop all interrupted: completed K of L` on stderr, return error for exit `1` (research R6).

- [X] T019 [P] Update informal “stop all” wording to `supervisor stop all` in `specs/001-supervisor-cli/quickstart.md` and `specs/002-status-stop-cli/quickstart.md` (FR-011).

- [X] T020 Run `go test ./...` from repository root; full suite green without `-tags=slow` (constitution delivery).

- [X] T021 Run manual scenarios in [quickstart.md](./quickstart.md) including Scenario 5 (`slow` or manual) and Scenario 6 (interrupt).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: Start immediately.
- **Phase 2**: After T001; blocks Phases 3–4.
- **Phase 3**: After Phase 2 checkpoint.
- **Phase 4**: Can start after T013 (help) in parallel with T014–T015 tail.
- **Phase 5**: After T014–T017; T011 depends on T018 for SIGINT handling.

### User Story Dependencies

- **US1 (P1)**: T012–T015.
- **US2 (P2)**: Empty and stale covered in T014 + tests T003–T005 (no separate implementation phase).
- **US3 (P2)**: T016–T017.

### Parallel Opportunities

- T002–T011 in Phase 2 (same file — apply in order or one PR).
- T009 (`slow`) and T019 in Phase 5.

---

## Parallel Example: User Story 1

```bash
go test ./tests/integration/... -run TestStopAllTwoInstances -v
go test ./tests/integration/... -tags=slow -run TestStopAllFiveInstances -v
```

---

## Implementation Strategy

### MVP First (User Story 1 + US2 behavior)

1. Complete Phase 1–2 (red tests).
2. Complete Phase 3 (T012–T015).
3. **STOP and VALIDATE**: T002, T003, T004, T005, T007 pass.
4. Phase 4–5 for help, SIGINT (T018), T011, T010, docs, suite.

### Task Count Summary

| Phase | Tasks | Story focus |
|-------|-------|-------------|
| 1 Setup | 1 | — |
| 2 Foundational | 10 | US1–US3 tests |
| 3 US1 | 4 | P1 MVP + empty in T014 |
| 4 US3 | 2 | Help / regression |
| 5 Polish | 4 | SIGINT, docs, suite |
| **Total** | **21** | |

**MVP scope**: Phases 1–3 through **T015** (empty branch in **T014**, not a later phase).

---

## Notes

- Do not change registry schema, `start`, `worker`, or `logs clear`.
- Log files remain on disk after stop (008 FR-009b).
- Contract reference: [contracts/cli.md](./contracts/cli.md) v11.
- Operational failure definition: [data-model.md](./data-model.md) and research R8.

## Phase 6: Convergence

- [X] T022 Update `specs/011-stop-all-command/quickstart.md` Scenario 2 step 1 to say `./supervisor stop all` instead of informal `stop all` per FR-011 (partial)
- [X] T023 Align bulk operational-failure stderr in `internal/cli/stop.go` with [contracts/cli.md](./contracts/cli.md) (per-ID failures must match single-ID stop; remove extra `stop failed:` line or emit only what `stop <id>` would) per FR-005 / data-model operational failure row (partial)