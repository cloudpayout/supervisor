---
description: "Task list for clear instance registry command"
---

# Tasks: Clear Instance Registry Command

**Input**: Design documents from `specs/009-clear-command/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: Success criteria SC-001–SC-003 and constitution delivery workflow require automated checks. Integration tests in `tests/integration/clear_test.go` are written in Phase 2 and should fail until `internal/cli/clear.go` is implemented. Unit tests in `internal/cli/clear_test.go` cover decline/affirmative parsing (SC-002) without a TTY.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested on its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Product code: `cmd/supervisor/`, `internal/cli/`
- Registry: `internal/registry/` (read/write via `Store.View` / `Store.Update` only; no schema change)
- Tests: `tests/integration/` using `supervisorBin`, `isolatedEnv`, and exit-code helpers from `tests/integration/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and toolchain; no new dependencies.

- [X] T001 Confirm branch `009-clear-command`, feature dir `specs/009-clear-command/`, and leave `go.mod` on module `github.com/cloudpayout/supervisor` Go 1.27.1 with no new `require` entries per [plan.md](./plan.md).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Failing integration and unit tests that encode the 009 CLI contract before `clear` is implemented.

**⚠️ CRITICAL**: Do not register `clear` on the root command until T002–T005 exist and fail for the expected reasons.

- [X] T002 [P] [US1] Add `TestClearYesEmptyRegistry` in `tests/integration/clear_test.go` using `supervisorBin(t)` and `isolatedEnv(t)`. Seed `instances.json` with `schema_version: 1` and at least one instance row. Run `clear --yes` with isolated `XDG_STATE_HOME`. Assert exit `0`, registry JSON has `"instances": []` and `schema_version` 1, then run `status` once and assert empty message (SC-001, FR-001, [contracts/cli.md](./contracts/cli.md)).

- [X] T003 [P] [US3] In `tests/integration/clear_test.go`, add `TestClearNonTTYRequiresYes`. With seeded registry, run `clear` with stdin from a pipe (`echo yes | …`) and **without** `--yes`. Assert exit `2`, stderr mentions `--yes` or non-interactive requirement, registry row still present (FR-006, SC-003).

- [X] T004 [P] [US1] In `tests/integration/clear_test.go`, add `TestClearCorruptRegistry`. Write invalid JSON to registry path, run `clear --yes`, assert exit `1`, file content unchanged (FR-008).

- [X] T005 [P] [US1] Add `internal/cli/clear_test.go` table tests for affirmative parsing: trimmed `y` and `yes` (any case) → affirmative; empty, `n`, `no`, other → decline (spec Assumptions; FR-003).

- [X] T014 [P] [US1] In `tests/integration/clear_test.go`, add `TestClearYesAlreadyEmpty`. With missing or zero-instance registry, run `clear --yes`; assert exit `0` and valid empty registry (User Story 1 scenario 3).

**Checkpoint**: `go test ./tests/integration/ -run Clear -v` and `go test ./internal/cli/ -run Clear -v` fail until implementation lands.

---

## Phase 3: User Story 1 - Reset registry after troubleshooting (Priority: P1) 🎯 MVP

**Goal**: Confirmed `clear` writes empty registry; decline leaves file unchanged; success messages for cleared vs already-empty.

**Independent Test**: `TestClearYesEmptyRegistry`, `TestClearYesAlreadyEmpty`, and unit tests in `internal/cli/clear_test.go` pass.

### Implementation for User Story 1

- [X] T006 [US1] Create `internal/cli/clear.go` with `newClearCmd()`: `Use: "clear"`, no args. `RunE` calls `openStore()`; `store.View` to load doc and `count := len(doc.Instances)` — on View/decode error print `supervisor: %v` to stderr and exit `1` without prompting (FR-008). If TTY and not `--yes`: print count, warning that browsers/workers are **not** stopped (FR-005), prompt `Clear instance registry? [y/N]: `, read one line; if not affirmative per T005 parser, exit `0` without write (FR-002, FR-003). On confirm or `--yes`: `store.Update` set `Instances` to `[]registry.Instance{}`, `SchemaVersion` to `registry.SchemaVersion`; print success (distinct message when `count == 0`) and exit `0` (FR-001, FR-004). Reuse `isTTY(os.Stdin)` from `internal/cli/status.go`.

- [X] T007 [US1] Register `newClearCmd()` in `internal/cli/root.go` via `root.AddCommand` with other operator commands (FR-001).

**Checkpoint**: T002, T004, T005, T014 pass. MVP for scripted registry reset is done.

---

## Phase 4: User Story 2 - Discover and use the command safely (Priority: P2)

**Goal**: `clear` appears in root help; `clear --help` documents wipe, confirmation, `--yes`, and no process stop.

**Independent Test**: Root help and `clear --help` integration assertions pass (FR-007).

### Tests for User Story 2

- [X] T008 [P] [US2] Extend `tests/integration/root_help_test.go` or `tests/integration/clear_test.go` with `TestRootHelpIncludesClear`: `help` and `--help` stdout must contain `clear` alongside `start`, `status`, `stop`, `screenshot`, `logs` (FR-007, [data-model.md](./data-model.md) catalog delta).

### Implementation for User Story 2

- [X] T009 [US2] In `internal/cli/clear.go`, set `Short` and `Long` help strings to describe registry wipe, TTY confirmation, `--yes` for scripts, decline behavior, and that `clear` does not stop processes (FR-007). Add `--yes` bool flag with help text. Ensure `cobra` help lists the flag.

- [X] T010 [US2] Add `TestClearHelp` in `tests/integration/clear_test.go`: run `clear --help`, assert stdout mentions instance registry, confirmation, and `--yes` (FR-007).

**Checkpoint**: User Stories 1 and 2 satisfied for operator discovery.

---

## Phase 5: User Story 3 - Refuse unattended destructive runs (Priority: P2)

**Goal**: Non-TTY without `--yes` → usage exit `2`; non-TTY with `--yes` clears without reading stdin.

**Independent Test**: `TestClearNonTTYRequiresYes` and `TestClearYesEmptyRegistry` with `--yes` only (no stdin).

### Implementation for User Story 3

- [X] T011 [US3] In `internal/cli/clear.go`, when `!isTTY(os.Stdin)` and `!yesFlag`, return `fmt.Errorf("clear requires --yes when stdin is not a terminal")` (wording per [contracts/cli.md](./contracts/cli.md)); do not read stdin (FR-006).

- [X] T012 [US3] Extend `isUsageError` in `internal/cli/root.go` to treat the clear non-TTY message as usage class (e.g. substring `requires --yes` or `stdin is not a terminal`) so `Execute()` exits `2` (research R4).

**Checkpoint**: T003 passes; piping `yes` without `--yes` never clears registry.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Regression suite and manual validation.

- [X] T013 [US3] Run `go test ./...` from repository root. All packages pass; no regressions in `start`, `status`, `stop`, `screenshot`, `logs`, or `worker` (constitution delivery workflow).

- [X] T015 [P] Execute validation steps in [quickstart.md](./quickstart.md) against a built binary and confirm outcomes match the guide.

- [X] T016 [P] Grep operator-facing docs for registry hand-edit instructions; optionally mention `supervisor clear` where troubleshooting guides reference deleting `instances.json` manually (only if such docs exist).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Phase 1 — **blocks** T006–T007
- **User Story 1 (Phase 3)**: Depends on Phase 2 tests
- **User Story 2 (Phase 4)**: Depends on T007 (command must exist for help tests); T008 can be written in Phase 2 in parallel
- **User Story 3 (Phase 5)**: Depends on T006; T011–T012 can follow T006
- **Polish (Phase 6)**: Depends on T013-worthy implementation complete

### User Story Dependencies

| Story | Priority | Depends on | Independent test |
|-------|----------|------------|------------------|
| US1 | P1 | T002, T004, T005, T006, T007 | `TestClearYes*`, `clear_test.go` unit |
| US2 | P2 | T007, T009 | `TestRootHelpIncludesClear`, `TestClearHelp` |
| US3 | P2 | T003, T011, T012 | `TestClearNonTTYRequiresYes` |

US1 delivers core value. US2 and US3 can proceed in parallel after T007.

### Parallel Opportunities

- Phase 2: T002, T003, T004, T005, T014 — all different files or isolated tests
- Phase 4: T008 parallel with T009 once T007 done
- Phase 6: T015, T016 parallel after T013

### Parallel Example: Phase 2

```bash
# Write all failing tests before implementation:
go test ./tests/integration/ -run 'TestClear' -v   # expect failures
go test ./internal/cli/ -run Affirmative -v        # expect failures until parser exists
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1–2 (tests red)
2. Complete Phase 3 (T006, T007)
3. **STOP and VALIDATE**: `go test ./tests/integration/ -run Clear -v` and unit tests green
4. Proceed to US2/US3 for full contract compliance

### Incremental Delivery

1. Setup + Foundational → red tests
2. US1 → scripted `clear --yes` + TTY prompt path → MVP
3. US2 → help discovery
4. US3 → non-TTY guard + exit `2`
5. Polish → full suite + quickstart

### Parallel Team Strategy

- Developer A: `internal/cli/clear.go` (T006, T009, T011)
- Developer B: `tests/integration/clear_test.go` (Phase 2)
- Developer C: `root.go` + `root_help_test.go` (T007, T012, T008)

---

## Notes

- Interactive decline on a real TTY is covered by unit parser (T005) and [quickstart.md](./quickstart.md); no PTY required in CI.
- Do not delete profile dirs or log files in `clear` (spec Assumptions).
- Cite FR IDs in commit messages where practical per constitution I.