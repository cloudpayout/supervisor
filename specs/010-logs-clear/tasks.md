---
description: "Task list for logs clear subcommand"
---

# Tasks: Clear Instance Log Files (`logs clear`)

**Input**: Design documents from `specs/010-logs-clear/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: SC-001–SC-003 and constitution delivery require automated checks. Integration tests in `tests/integration/logs_clear_test.go` are written in Phase 2 and must fail until `internal/cli/logs_clear.go` is implemented. Remove obsolete 009 tests in Phase 2.

**Organization**: Tasks grouped by user story; Phase 2 also covers 009 retirement.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1, US2, US3
- Include exact file paths in descriptions

## Path Conventions

- Product code: `internal/cli/`
- Tests: `tests/integration/`, `internal/cli/`
- Registry states protecting logs: `starting`, `running`, `stopping` (FR-001a; data-model eligibility table)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope before replacing 009.

- [X] T001 Confirm branch `010-logs-clear`, feature dir `specs/010-logs-clear/`, and `go.mod` unchanged (Go 1.27.1, no new `require`) per [plan.md](./plan.md).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Red tests for v10 contract and removal of 009 root `clear`. Do not register `logs clear` until T002–T007 exist and fail for expected reasons.

- [X] T002 [P] [US1] Add `tests/integration/logs_clear_test.go` with `TestLogsClearYesOrphanAndRunning`: seed `orphan01.log` and `active01.log` under `{XDG_STATE_HOME}/supervisor/logs/`; registry row `active01` state **`running`**; run `logs clear --yes`; assert orphan removed, `active01.log` remains, `instances.json` unchanged (SC-001, FR-001a).

- [X] T003 [P] [US1] In `tests/integration/logs_clear_test.go`, add `TestLogsClearProtectsStarting`: registry row state **`starting`** with matching `.log`; `logs clear --yes`; file must remain (clarification 2026-11-22).

- [X] T004 [P] [US1] In `tests/integration/logs_clear_test.go`, add `TestLogsClearYesOnlyProtected`: only active lifecycle logs on disk; `logs clear --yes`; exit `0`, no files deleted, message indicates nothing eligible (User Story 1 scenario 3).

- [X] T005 [P] [US3] In `tests/integration/logs_clear_test.go`, add `TestLogsClearNonTTYRequiresYes`: seeded orphan log; run `logs clear` with piped stdin and no `--yes`; exit `2`, file present (SC-003, FR-007).

- [X] T006 [P] [US2] In `tests/integration/logs_clear_test.go`, add `TestLogsClearHelp` and `TestLogsHelpListsClearSubcommand`: `logs clear --help` mentions eligible deletion, skip for `starting`/`running`/`stopping`, `--yes`; `logs --help` lists `clear` (FR-008).

- [X] T007 [P] [US2] Update `tests/integration/root_help_test.go`: root `help`/`--help` must **not** include top-level `clear`; remove or replace `TestRootHelpIncludesClear` from `tests/integration/clear_test.go` contract (v10 root catalog). Add `TestRootClearUnknownCommand`: `supervisor clear` exit `2` unknown command.

- [X] T008 [P] [US1] Move `TestConfirmYes` from `internal/cli/clear_test.go` to `internal/cli/confirm_test.go` (or `logs_clear_test.go`) targeting `confirmYes` after extraction; keep table cases for `y`/`yes`/decline (FR-004).

- [X] T009 Delete obsolete 009 artifacts when superseded by new tests: `tests/integration/clear_test.go`, `internal/cli/clear.go`, `internal/cli/clear_test.go` — **defer file deletion until T010–T012 land**; until then tests may fail on missing `logs clear`. Document in commit: retire 009.

**Checkpoint**: `go test ./tests/integration/ -run 'LogsClear|RootHelp|RootClear' -v` fails on missing subcommand / wrong root `clear`.

---

## Phase 3: User Story 1 - Reclaim disk after troubleshooting (Priority: P1) 🎯 MVP

**Goal**: Selective deletion of eligible `*.log` files; protect `starting`/`running`/`stopping`.

**Independent Test**: T002, T003, T004 pass.

### Implementation for User Story 1

- [X] T010 [US1] Add `internal/cli/confirm.go` with exported or package `confirmYes(line string) bool` moved from `internal/cli/clear.go` (same semantics as 009).

- [X] T011 [US1] Add `internal/cli/logs_clear.go`: `newLogsClearCmd()` with `--yes`; `runLogsClear` implements eligibility per [data-model.md](./data-model.md) (`protectedIDs` for states **`starting`**, **`running`**, **`stopping`**); glob `filepath.Join(layout.LogsDir(), "*.log")`; TTY prompt with eligible and skipped counts (FR-003, FR-006); decline exits `0` without `os.Remove` (FR-004); confirm removes eligible only; summary stdout (FR-005); missing logs dir → zero eligible (spec edge case).

- [X] T012 [US1] In `internal/cli/logs.go`, call `cmd.AddCommand(newLogsClearCmd())` and extend `Long` with `supervisor logs clear` (FR-001).

- [X] T013 [US1] Remove 009 root command: delete `internal/cli/clear.go`, remove `root.AddCommand(newClearCmd())` from `internal/cli/root.go`, delete `tests/integration/clear_test.go` and `internal/cli/clear_test.go` after T010.

**Checkpoint**: T002–T004, T008 pass. MVP selective clear complete.

---

## Phase 4: User Story 2 - Discover the use safely (Priority: P2)

**Goal**: Help text and discoverability under `logs`.

**Independent Test**: T006, T007 pass.

### Implementation for User Story 2

- [X] T014 [US2] In `internal/cli/logs_clear.go`, set `Short`/`Long` and flag help per FR-008 (registry unchanged, active lifecycle skip, `--yes`, top-level `*.log` only).

- [X] T015 [US2] In `internal/cli/logs.go`, ensure parent `Short`/`Long` still accurate when `clear` is a subcommand (no duplicate root-level wording).

**Checkpoint**: User Stories 1 and 2 satisfied.

---

## Phase 5: User Story 3 - Refuse unattended bulk deletion (Priority: P2)

**Goal**: Non-TTY requires `--yes`; same error class as 009.

**Independent Test**: T005 passes.

### Implementation for User Story 3

- [X] T016 [US3] In `internal/cli/logs_clear.go`, when `!isTTY(os.Stdin) && !yes`, return `fmt.Errorf("logs clear requires --yes when stdin is not a terminal")` (or equivalent matching [contracts/cli.md](./contracts/cli.md)); do not read stdin (FR-007).

- [X] T017 [US3] Extend `isUsageError` in `internal/cli/root.go` if needed so the logs-clear non-TTY message exits `2` (reuse `requires --yes` / `stdin is not a terminal` substrings).

- [X] T018 [US3] In `internal/cli/logs_clear.go`, reject parent view flags on clear subcommand per FR-009 (`logs clear --recent` etc. exit usage `2`); rely on Cobra flag inheritance rules or explicit checks in `logs_clear` `PreRunE` / separate command with no inherited flags.

**Checkpoint**: T005 and FR-009 combination tests (add `TestLogsClearRejectsRecentFlag` in `tests/integration/logs_clear_test.go` if not covered in T018).

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T019 Run `go test ./...` from repository root; full suite green (constitution delivery).

- [X] T020 [P] Run validation steps in [quickstart.md](./quickstart.md) against a built binary.

- [X] T021 [P] Grep repo for `supervisor clear` (not `logs clear`) in operator docs; update if any reference 009 registry wipe.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** (tests red) → **Phase 3** (T010–T013) → **Phases 4–5** (mostly same files as T011–T018) → **Phase 6**

### User Story Dependencies

| Story | Priority | Depends on | Independent test |
|-------|----------|------------|------------------|
| US1 | P1 | T002–T004, T010–T013 | `TestLogsClearYes*` |
| US2 | P2 | T006, T007, T012, T014 | `TestLogsClearHelp`, root help |
| US3 | P2 | T005, T016–T018 | `TestLogsClearNonTTY*` |

### Parallel Opportunities

- Phase 2: T002–T008 in parallel (different test cases/files)
- T020 ∥ T021 after T019

### Parallel Example: Phase 2

```bash
go test ./tests/integration/ -run LogsClear -v   # expect failures until T011–T012
go test ./internal/cli/ -run ConfirmYes -v
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Phase 1–2 (red tests + retire test expectations for root `clear`)
2. Phase 3 (T010–T013)
3. **STOP**: `go test ./tests/integration/ -run LogsClear -v`

### Incremental Delivery

1. US1 → selective `logs clear --yes`
2. US2 → help polish
3. US3 → non-TTY + FR-009 flag guards
4. Polish → full suite + quickstart

---

## Notes

- Cite FR IDs in commits where practical (constitution I).
- `registry.Store` is **read-only** for this feature; never `Update` in `logs_clear.go`.
- Interactive TTY affirm/decline: unit tests + quickstart manual; no PTY in CI.