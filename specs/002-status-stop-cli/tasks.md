---
description: "Task list for status/stop CLI behavior (002)"
---

# Tasks: Status and Stop CLI Behavior

**Input**: Design documents from `specs/002-status-stop-cli/`  
**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/cli.md](./contracts/cli.md)

**Organization**: By user story (US1 one-shot status, US2 follow, US3 stop by ID). FR-009 requires integration test updates.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no unfinished dependencies)
- **[Story]**: US1, US2, or US3

## Path Conventions

Single Go module at repository root: `internal/cli/`, `tests/integration/`, `cmd/supervisor/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Align implementer with v2 contract before code changes

- [x] T001 Review operator deltas in specs/002-status-stop-cli/contracts/cli.md and decisions in specs/002-status-stop-cli/research.md (exit `2` for stop usage; `renderStatus(clearScreen)` split)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared status rendering hook for US1 and US2

**⚠️ CRITICAL**: Complete before user story work on `status`

- [x] T002 Add `clearScreen bool` to `renderStatus` in internal/cli/status.go; emit ANSI full-screen clear only when `clearScreen && isTTY(os.Stdout)` per clarification 2026-11-09 and research R2

**Checkpoint**: US1 and US2 can proceed on `status.go`

---

## Phase 3: User Story 1 — Snapshot instance status (Priority: P1) 🎯 MVP

**Goal**: Default `supervisor status` prints one reconciled table (or empty message) and exits without clearing the screen (FR-001, FR-003).

**Independent Test**: `./supervisor status` returns within seconds with no Ctrl+C; output lists instances or `No instances are running.`

### Tests for User Story 1 (FR-009)

- [x] T003 [P] [US1] Simplify `runStatusOnce` in tests/integration/status_refresh_test.go to run `status` without SIGINT/kill wait (one-shot should exit on its own)
- [x] T004 [P] [US1] Add `TestStatusOneShotExitsImmediately` in tests/integration/status_oneshot_test.go asserting `status` completes within 5s and exit code `0` (SC-001)

### Implementation for User Story 1

- [x] T005 [US1] Change default `runStatus` path in internal/cli/status.go to call `renderStatus(store, false)` once and return `nil` (no ticker/signals on default path) (FR-001)
- [x] T006 [US1] Update status `Short` help in internal/cli/status.go to describe default snapshot behavior (FR-008)

**Checkpoint**: `go test ./tests/integration/... -run StatusEmptyAndStarting` passes with one-shot `status`; MVP shippable for scripting

---

## Phase 4: User Story 2 — Live status monitoring (Priority: P2)

**Goal**: `supervisor status --follow` / `-f` preserves ~2s refresh loop with TTY clear between cycles (FR-002).

**Independent Test**: `./supervisor status -f` until Ctrl+C → exit `0`; table refreshes ~every 2s.

### Tests for User Story 2 (FR-009)

- [x] T007 [US2] Update follow-loop coverage in tests/integration/status_refresh_test.go to invoke `status` with `-f` (or `--follow`) where SIGINT and refresh behavior is asserted

### Implementation for User Story 2

- [x] T008 [US2] Register `BoolP("follow", "f", false, ...)` on the status command in internal/cli/status.go (FR-002, research R1)
- [x] T009 [US2] When follow is true, run existing signal+ticker loop in internal/cli/status.go calling `renderStatus(store, isTTY(os.Stdout))` each cycle (FR-002, FR-003)
- [x] T010 [P] [US2] Extend status `Long` help in internal/cli/status.go to document `--follow` / `-f` (FR-008)

**Checkpoint**: Follow mode matches pre-change monitoring; one-shot default unchanged

---

## Phase 5: User Story 3 — Stop by instance ID (Priority: P1)

**Goal**: `supervisor stop <instance-id>` only; no stdin; missing/extra args → usage exit `2` (FR-004–FR-006); preserve not-found/already-stopped/stop failure semantics (FR-007).

**Independent Test**: `stop <id>` tears down instance; `stop` alone exits `2`; unknown ID still exit `0` with not-found message.

### Tests for User Story 3 (FR-009)

- [x] T011 [P] [US3] Change `stopID` in tests/integration/start_worker_test.go to `exec.Command(bin, "stop", id)` with no `Stdin`
- [x] T012 [US3] Update `TestStopLifecycleMessages` in tests/integration/stop_lifecycle_test.go: require argv ID, assert missing-arg exit `2`, remove empty-stdin/`stop aborted` cases; keep not-found and already-stopped cases with argv

### Implementation for User Story 3

- [x] T013 [US3] Set `Use: "stop <instance-id>"`, `Args: cobra.ExactArgs(1)`, wire `id := strings.TrimSpace(args[0])` in internal/cli/stop.go; remove bufio stdin prompt and abort branches (FR-004, FR-005, research R3)
- [x] T014 [US3] Return usage-class error when trimmed ID is empty in internal/cli/stop.go (FR-006; data-model: exactly one non-blank `instance_id`)
- [x] T015 [US3] Update stop `Short`/`Long` help in internal/cli/stop.go for required positional ID; remove interactive prompting text (FR-008)

**Checkpoint**: `TestStartStatusStop` and `TestStopLifecycleMessages` pass; all cleanups use `stop <id>`

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: FR-009 sweep and constitution test gate

- [x] T016 [P] Grep `tests/integration` for `exec.Command(bin, "stop")` without instance ID and update callers (e.g. tests/integration/status_concurrent_test.go, status_ten_instances_test.go via `stopID`)
- [x] T017 Run `go test ./...` from repository root and fix any remaining failures (constitution Delivery Workflow)
- [x] T018 [P] Walk through specs/002-status-stop-cli/quickstart.md scenarios and confirm outcomes match SC-001–SC-005

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** → **US1 (Phase 3)** → **US2 (Phase 4)** requires T002 + US1 default path (T005) before follow flag (T008) is wired cleanly
- **US3 (Phase 5)** can start after Phase 2 in parallel with US1/US2 (different file: internal/cli/stop.go)
- **Polish** after US1–US3

### User Story Dependencies

| Story | Depends on | Independent test |
|-------|------------|------------------|
| US1 | T002 | One-shot `status` |
| US2 | T002, T005–T009 | `status -f` + SIGINT |
| US3 | None (stop.go only) | `stop <id>` / usage error |

### Parallel Opportunities

```bash
# After T002, parallel tracks:
# Track A: T005–T010 (status.go + status tests)
# Track B: T011–T015 (stop.go + stop tests)

# Within US1 after T005:
# T003 + T004 in parallel (different test files)
```

---

## Parallel Example: User Story 3

```text
T011 stopID helper in tests/integration/start_worker_test.go
T012 stop_lifecycle_test.go (after T013 if tests run against new binary — implement T013–T014 first, then T011–T012)
```

Recommended order for US3: T013 → T014 → T015 → T011 → T012 (or T011 immediately after T013).

---

## Implementation Strategy

### MVP First (User Story 1)

1. T001 → T002 → T005–T006 → T003–T004  
2. Validate: `./supervisor status` snapshot + integration tests  
3. Optional early ship before follow/stop if only scripting status needed (not recommended — stop still prompts until US3)

### Incremental Delivery

1. US1: scriptable status  
2. US2: operator monitoring  
3. US3: scriptable stop  
4. Polish: full suite green + quickstart

### Suggested merge order

US1 + US3 (both P1) → US2 → Polish

---

## Notes

- Cite spec IDs in commit messages (e.g. FR-001, User Story 1).
- Do not change registry schema, `start`, or `worker` (spec assumptions).
- Total tasks: **18** (Setup 1, Foundational 1, US1 4, US2 4, US3 5, Polish 3).

## Phase 7: Convergence

- [x] T019 Update slow integration test in tests/integration/status_ten_instances_test.go to use `status --follow` and assert at least one full refresh within a 2s window with ten instances per SC-002 (partial)
- [x] T020 Set `Args: cobra.NoArgs` on the status command in internal/cli/status.go and extend integration or CLI tests so extra positional args yield usage exit `2`, per specs/002-status-stop-cli/data-model.md and contracts/cli.md (partial)
