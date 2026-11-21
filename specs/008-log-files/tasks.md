---
description: "Task list for instance log files and supervisor logs command"
---

# Tasks: Instance Log Files for Operators

**Input**: Design documents from `specs/008-log-files/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: FR-012 and spec user stories require integration and unit tests; test tasks are listed before implementation per story.

**Organization**: Tasks grouped by user story (US1–US4 from [spec.md](./spec.md)). US1 and US2 are both P1; US1 is MVP.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks in the same batch)
- **[Story]**: US1…US4
- Paths are under repository root unless noted

## Path Conventions

- Product: `internal/logfmt/`, `internal/cli/`, `internal/worker/`, `internal/registry/`
- Tests: `tests/unit/`, `tests/integration/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm feature docs and shared layout helpers referenced by all stories.

- [X] T001 [P] Add `LogsDir() string` on `internal/registry/paths.go` returning `filepath.Join(l.StateDir, "logs")`; keep existing `LogPath(id)` unchanged per [data-model.md](./data-model.md).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Structured line format, ingest writer, and worker log I/O. **No user story work until this checkpoint.**

- [X] T002 [P] Create `internal/logfmt/format.go`: `Format(level, msg string) string` using UTC `time.RFC3339Nano` with `Z` suffix, level exactly `ERROR`|`NOTICE`|`INFO`, single ASCII spaces, message with no embedded `\n`; export `ParseLine` returning structured match per FR-002.
- [X] T003 [P] Create `tests/unit/logfmt_test.go`: round-trip `Format`/`ParseLine`; reject multiline messages; filter cases for levels.
- [X] T004 [P] Create `internal/logfmt/filter.go`: `MinLevel` type `error`|`notice`; `KeepLine(line string, min MinLevel) bool` (structured lines only; unstructured excluded when min set).
- [X] T005 [P] Create `internal/logfmt/redact.go`: `Redact(msg string, secrets []string) string` replacing secret substrings with `[REDACTED]` (empty secrets no-op).
- [X] T006 Create `internal/logfmt/ingest.go`: `NewIngestWriter(w io.Writer, secrets []string) io.Writer` line-buffering, skip blank lines, wrap each line as structured `INFO` via `Format` after `Redact` (FR-002b).
- [X] T007 Add hidden `--log-path` on `internal/cli/worker.go` `run` subcommand; `MarkFlagRequired` with `id` and `log-path` when worker starts.
- [X] T008 Update `internal/cli/start.go`: pass `--log-path` with canonical `layout.LogPath(id)` to worker; remove direct `worker.Stdout`/`Stderr` assignment to parent-owned file (worker owns structured ingest).
- [X] T009 Update `internal/worker/run.go`: open `log-path` append-only; set `logfmt.IngestWriter` as `os.Stderr` sink for panics/logging helper; load cookie values from session file into redaction list when session present; add `logWriter` helper emitting `Notice`/`Info`/`Error` via `Format`.

**Checkpoint**: Worker appends only structured lines to `{supervisor_state}/logs/{id}.log`.

---

## Phase 3: User Story 1 - Read what happened for one instance (Priority: P1) MVP

**Goal**: `supervisor logs <id>` prints full log once; structured lines visible; root help lists `logs`.

**Independent Test**: `start --headless` → `logs <id>` shows `TIMESTAMP LEVEL message`; spec US1.

### Tests for User Story 1

- [X] T010 [P] [US1] Add `tests/integration/logs_read_test.go`: isolated `XDG_STATE_HOME`; `start --headless`; `logs <id>` exit 0; stdout lines match structured grammar; `supervisor help` output contains `logs`.

### Implementation for User Story 1

- [X] T011 [US1] Create `internal/cli/logs.go`: `newLogsCmd()` with `Use: "logs <instance-id>"`; `resolveLogPath(store, layout, id)` per FR-009 (registry `log_path` then canonical `{supervisor_state}/logs/{id}.log`); `streamFile(path, follow false, level none)` print line-by-line to stdout.
- [X] T012 [US1] Register `newLogsCmd()` in `internal/cli/root.go`; extend `isUsageError` in `internal/cli/root.go` for logs usage strings if needed.
- [X] T013 [P] [US1] In `internal/worker/run.go` and `internal/worker/aviator.go`, emit `NOTICE` at entry phase transitions and `ERROR` on entry failure messages (no cookie values in message text per FR-010).

**Checkpoint**: T010 passes; operator can read live instance logs.

---

## Phase 4: User Story 2 - Discover logs for runs no longer in status (Priority: P1)

**Goal**: `logs --recent` lists ≤50 orphaned logs with UTC activity and last ERROR summary; `logs <id>` after stop exit 0 with not-tracked notice.

**Independent Test**: Stop or fail instance → ID in `--recent`; spec US2.

### Tests for User Story 2

- [X] T014 [P] [US2] Add `tests/integration/logs_recent_test.go`: start → `stop` → `logs --recent` shows ID with columns; `logs <id>` exit 0; stderr contains not-tracked notice; seed >50 log files optional skip or subtest for omitted count.

### Implementation for User Story 2

- [X] T015 [P] [US2] Create `internal/logfmt/summary.go`: `LastErrorInFile(path string) (message string, found bool)` scanning structured `ERROR` only; placeholder when none.
- [X] T016 [US2] Extend `internal/cli/logs.go`: `--recent` flag mutually exclusive with positional ID; glob `layout.LogsDir()/*.log`; filter IDs not in registry; sort by `ModTime` desc; cap 50 rows; tabwriter `ID`, `LAST_ACTIVITY_UTC`, `LAST_ERROR` truncated ~80 runes; trailing omitted count (FR-009a).
- [X] T017 [US2] Audit `internal/registry/reconcile.go`, `internal/cli/stop.go`, `internal/worker/run.go`: confirm no `os.Remove` on `*.log`; add comment if already satisfied (FR-009b).

**Checkpoint**: T014 passes; crashed/stopped runs discoverable without `status`.

---

## Phase 5: User Story 3 - Watch logs while a run is in progress (Priority: P2)

**Goal**: `logs <id> --follow` tails until SIGINT, exit 0.

**Independent Test**: Follow during running instance; interrupt; spec US3.

### Tests for User Story 3

- [X] T018 [P] [US3] Add `tests/integration/logs_follow_test.go`: start instance; run `logs <id> -f` with timeout/context cancel after observing ≥1 line; exit 0.

### Implementation for User Story 3

- [X] T019 [US3] Extend `internal/cli/logs.go`: `--follow`/`-f`; poll new content every ~500ms from file offset; SIGINT/TERM exits 0 like `internal/cli/status.go`; do not exit when instance stops but follow until interrupt.

**Checkpoint**: T018 passes.

---

## Phase 6: User Story 4 - Focus on errors or notices (Priority: P3)

**Goal**: `--level error` and `--level notice` filter structured lines; output exact on-disk lines.

**Independent Test**: Log with mixed severities; filtered output; spec US4.

### Tests for User Story 4

- [X] T020 [P] [US4] Add `tests/integration/logs_level_test.go`: instance with NOTICE/ERROR/INFO lines; `logs <id> --level error` only ERROR; `--level notice` ERROR+NOTICE; lines byte-identical to file.

### Implementation for User Story 4

- [X] T021 [US4] Extend `internal/cli/logs.go`: `--level` flag accepting `error`|`notice` only; wire `logfmt.KeepLine` in `streamFile`; invalid value usage exit 2.

**Checkpoint**: T020 passes.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: FR-012 coverage, help contract, quickstart validation.

- [X] T022 [P] Add `tests/integration/logs_orphan_failure_test.go`: headless start with worker failure path removing registry row while log remains; `logs --recent` and `logs <id>` per SC-006 (may reuse stand-in pattern from `onewin_aviator_currency_fail_test.go` without live casino).
- [X] T023 [P] Extend `tests/integration/onewin_secrets_test.go` or `tests/integration/logs_secrets_test.go`: session secret MUST NOT appear in `logs <id>` stdout (FR-010).
- [X] T024 [P] Extend `tests/integration/root_help_test.go`: assert `logs` listed with `start`, `status`, `stop`, `screenshot` per contracts/cli.md.
- [X] T025 Run [quickstart.md](./quickstart.md) scenarios manually or via `go test ./...` and fix gaps.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** → **US1 (Phase 3)** → **US2 (Phase 4)** → **US3 (Phase 5)** → **US4 (Phase 6)** → **Polish (Phase 7)**
- US2 depends on US1 `logs` command skeleton (T011–T012); US3/US4 extend same file `internal/cli/logs.go` sequentially or merged carefully.

### User Story Dependencies

| Story | Depends on |
|-------|------------|
| US1 | Phase 2 |
| US2 | US1 (`logs.go` base) |
| US3 | US1 streaming |
| US4 | US1 streaming |

### Parallel Opportunities

- T002–T005, T010, T014, T018, T020, T022–T024 marked `[P]` can run in parallel when their prerequisites are done.
- After Phase 2, US1 implementation T013 can parallel T011 if `logs.go` created first as stub.

### Parallel Example: Foundational

```bash
# After T001:
# Parallel: logfmt format, filter, redact unit tests (T002–T005)
```

### Parallel Example: User Story 1

```bash
# T010 integration test (red) while T011–T012 implement logs command
# T013 worker NOTICE/ERROR in parallel if logs.go exists
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Complete Phases 1–2 (structured logs on disk).
2. Complete Phase 3 (US1): `logs <id>` + help text.
3. **STOP**: `go test ./...` with T010 green.

### Incremental Delivery

1. US1 → US2 (orphans + post-stop read) — both P1 operator value.
2. US3 follow, US4 filter.
3. Polish: secrets, failure orphan, help contract.

---

## Notes

- Canonical path: `{XDG_STATE_HOME or ~/.local/state}/supervisor/logs/{id}.log` (FR-001a).
- Do not delete log files on `stop`, reconcile, or worker failure (FR-009b).
- Pre-feature unstructured logs: passthrough only when `--level` unset (FR-002a).
- Cite spec FR/scenario in commit messages when implementing.

## Phase 8: Convergence

- [X] T026 Extend `tests/integration/onewin_aviator_currency_fail_test.go` (or `logs_orphan_failure_test.go`) so after registry removal, `logs --recent` lists the instance ID with last-ERROR summary containing `currency`, and `logs <id>` exits 0 with stdout containing structured `ERROR currency`, per SC-006 and T022 intent (partial)
- [X] T027 Add `tests/integration/logs_usage_test.go`: `supervisor logs` with no args and `supervisor logs --level error` with no id exit usage code 2, per FR-008 and FR-012 (missing)
- [X] T028 Add `tests/integration/logs_usage_test.go` case: unknown instance ID with no log file exits non-zero and stderr mentions log not found, per spec edge case and contracts/cli.md (missing)
- [X] T029 Change untracked-instance stderr in `internal/cli/logs.go` to match contracts/spec wording (e.g. instance no longer in status), per contracts/cli.md and US1 AC4 (partial)
- [X] T030 Add `tests/integration/logs_level_test.go` case: `--level error` on a log with only NOTICE/INFO exits 0 with empty stdout, per US4 AC3 (partial)