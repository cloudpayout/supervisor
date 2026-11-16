---
description: "Task list for start command syntax (positional session path, casino in file)"
---

# Tasks: Start command syntax

**Input**: Design documents from `specs/005-start-command-syntax/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: The specification requires the ordinary confirmation in the default test run (`start <path> --headless`, then `status`, `screenshot`, `stop` against a local stand-in). Per [plan.md](./plan.md) Testing: SC-003 visible window is validated manually in quickstart (T017); default `go test` uses registry `headless` as the automated proxy (T004, T008). Test tasks below are written first and are expected to fail until the implementation task for that story is done.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested on its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Product code: `cmd/supervisor/`, `internal/cli/`, `internal/worker/`, `internal/session/`, `internal/registry/`, `internal/casino/`
- Tests: `tests/unit/`, `tests/integration/` using `supervisorBin`, `isolatedEnv`, `requireChrome`, and `writeSession` from `tests/integration/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No new dependencies. Confirm the module is unchanged.

- [X] T001 Leave `go.mod` on module `github.com/cloudpayout/supervisor` and Go 1.27.1 with no new requirement.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Session file shape with required `casino`. No user story starts until parse rules are decided and unit-tested.

- [X] T002 [P] Extend `tests/unit/session_file_test.go` for `internal/session/file.go`. Top-level `casino` is required when parsing a file for `start`. `cookies` length is at least 1. Missing, empty, or unknown `casino` values fail. `Onewin` and `1win.com` fail. Cookie rules from feature 004 stay unchanged. Unknown top-level fields fail. `value` is never echoed in errors.

**Checkpoint**: Unit tests describe the parser contract before `start` changes.

---

## Phase 3: User Story 1 - Start from a session file that names the casino (Priority: P1) 🎯 MVP

**Goal**: `start <path>` applies cookies and opens the registered page for the file's `casino`. `start` with no path stays a blank page. Removed flags `--session-file`, `--casino`, and `--headed` are usage errors.

**Independent Test**: Stand-in on `127.0.0.1` shows `balance` when the file has `"casino": "onewin"` and the expected cookie. `start <path> --headless` succeeds. Files without the cookie withhold `balance`. No live casino.

### Tests for User Story 1

- [X] T003 [P] [US1] Rewrite `tests/integration/onewin_usage_test.go` without Chrome. `start --session-file <file>`, `start --session-file=<file>`, `start --casino onewin <file>`, `start --casino Onewin <file>`, `start --headed <file>`, `start <file> extra`, and `start --casino onewin` exit `2` with no instance id. `start <missing>` exits non-zero with no id and no profile. `start <file>` with no `casino` field or bad `casino` exits non-zero with no id. Cookie value never appears in output.
- [X] T004 [P] [US1] Update `tests/integration/onewin_start_test.go` before `internal/cli/start.go` changes. Session fixtures include `"casino": "onewin"`. Use `start <temp> --headless` instead of `--casino` and `--session-file`. Expect balance shown and withheld cases, two-cookie case, blank `start` and `start --headless` with no path unchanged. Run the stand-in twice for the same file: `start <temp> --headless` and `start --headless <temp>` must each show `balance` (FR-001 flag order). After `start <temp>` without `--headless`, read the registry and assert `headless` is unset or false on that row (headed default). Registry row has `casino` `onewin` when a path was passed and omits `casino` on blank start. Update `tests/integration/onewin_helpers_test.go` `writeSession` default body to include `"casino": "onewin"` when callers rely on it.

### Implementation for User Story 1

- [X] T005 [US1] Implement `internal/session/file.go` to satisfy `tests/unit/session_file_test.go`. Export parse result with casino identifier and cookies. Validate `casino` through `internal/casino.Lookup`. Do not log or store cookie `value`.
- [X] T006 [US1] In `internal/cli/start.go`, implement the full operator `start` surface in one pass (includes US2 help and display behavior; do not defer). Zero or one positional path. Remove flags `--casino`, `--session-file`, and `--headed`; keep `--headless` only; `useHeadless := headless`. Clear the shorthand `h` on the default help flag so `start -h` is an unknown shorthand usage error. Set `Short`, `Long`, and flag usage per `specs/005-start-command-syntax/contracts/cli.md`. On path present: parse file before preflight, set registry `casino` from file, spawn `worker run --id <id> [--headless] --session-file <path>`. On no path: blank start unchanged. Extra positionals exit `2` with no row. Legacy flags absent from Cobra → unknown-flag usage errors.
- [X] T007 [US1] In `internal/worker/run.go` `openTarget`, when `sessionFile` is non-empty, require `casinoID` from the registry instance only. Remove the fallback that sets `casinoID` to `onewin` when empty. Keep `about:blank` when `sessionFile` is empty.

**Checkpoint**: `start <path> --headless` opens the stand-in. Blank `start` still works. Legacy flags fail.

---

## Phase 4: User Story 2 - Open a visible window unless headless is requested (Priority: P2)

**Goal**: Default headed when `--headless` is omitted. `--headed` and `--session-file` are refused. `start --headless <path>` matches `start <path> --headless`. `--help` prints help without starting; `-h` is usage error.

**Independent Test**: In `go test`, registry `headless` is unset or false for `start` and `start <path>` without `--headless`, and true when `--headless` is passed (T004, T008). A visible Chrome window for headed starts is validated manually in quickstart [Headed default](./quickstart.md#headed-default-manual-not-in-default-go-test) (T017), not gated on CI. `start --session-file <path>`, `start --headed <path>`, and `start -h` exit `2` with no id. `start --help` exits `0` with no instance.

### Tests for User Story 2

- [X] T008 [P] [US2] Update `tests/integration/start_flags_test.go`. Remove mutual `--headed`/`--headless` test. Add `start --headed <file>` exit `2`. With Chrome: `start` with no args leaves registry `headless` unset or false; `start --headless` with no args sets `headless` true on the row (extends `tests/integration/start_worker_test.go` patterns). With a minimal valid session file (includes `"casino": "onewin"` and one cookie), `start <file>` without `--headless` leaves `headless` unset or false; `start <file> --headless` sets `headless` true (SC-003 automated proxy for visible window). Actual window visibility is manual per quickstart, not asserted in this test.
- [X] T009 [P] [US2] Add `tests/integration/start_help_test.go` without Chrome. `start --help` and `start <file> --help` exit `0`, print help, and create no registry row. `start -h` exits `2` with no id. Help text does not contain `--session-file`, `--headed`, or `--casino`, and mentions positional path and blank start.

### Implementation for User Story 2

- [X] T010 [US2] No separate code task: `internal/cli/start.go` behavior for FR-006 and FR-007 is delivered in T006. Run `go test` for `start_flags_test.go` and `start_help_test.go` after T006; fix only if those tests fail.

**Checkpoint**: Display mode and help match FR-006 and FR-007.

---

## Phase 5: User Story 3 - Keep the rest of the run (Priority: P3)

**Goal**: After positional `start`, `status`, `screenshot`, and `stop` behave as in features 002–004. Cookie secrets stay out of logs and registry. Session path may appear on `start` argv as the positional argument.

**Independent Test**: After `start <path> --headless`, run `status`, `screenshot`, `stop`; profile removed. Secrets test passes with updated command line.

### Tests for User Story 3

- [X] T011 [P] [US3] Update `tests/integration/onewin_stop_test.go` to use `start <path> --headless` instead of `--casino` and `--session-file`.
- [X] T012 [P] [US3] Update `tests/integration/onewin_secrets_test.go` to use `start <path> --headless`. Expect the positional path may appear on the `start` process command line but not `--session-file` on operator `start`. Generated cookie value still absent from worker log, registry, and screenshot name.
- [X] T013 [P] [US3] Update `tests/integration/onewin_screenshot_test.go` if it invokes `start` with legacy flags; use `start <path> --headless`.

### Implementation for User Story 3

- [X] T014 [US3] In `internal/cli/start.go` and `internal/worker/run.go`, confirm errors for bad files and failed cookie set still omit cookie `value`. No change to `internal/cli/screenshot.go`, `internal/cli/stop.go`, or `internal/cli/status.go` unless a test proves a regression. Re-run `tests/integration/onewin_stop_test.go` (covers `status` after positional `start`) to confirm FR-010 list/capture/stop path still passes.

**Checkpoint**: Full ordinary confirmation in `go test ./...` passes per [quickstart.md](./quickstart.md).

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Spec trace, docs in code, default test run.

- [X] T015 [P] Add spec citations in `internal/cli/start.go` and `internal/session/file.go` comments referencing `specs/005-start-command-syntax` scenarios (constitution I).
- [X] T016 Run `go test ./...` from the repository root and fix any remaining tests that still call `start` with `--casino` or `--session-file` (for example `tests/integration/onewin_start_test.go` blank-start section, `tests/integration/registry_concurrent_test.go`, `tests/integration/status_concurrent_test.go`).
- [X] T017 Validate [quickstart.md](./quickstart.md) manually: usage table, blank start, help, ordinary confirmation, and [Headed default](./quickstart.md#headed-default-manual-not-in-default-go-test)—confirm a visible Chrome window for `supervisor start <session-file>` (no `--headless`) against the stand-in or a throwaway file; confirm `start <path> --headless` does not show a window (FR-006 / SC-003).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: Start immediately.
- **Phase 2**: After Phase 1; blocks all user stories.
- **Phase 3 (US1)**: After Phase 2; MVP.
- **Phase 4 (US2)**: After T006 (`start.go` includes FR-006/FR-007); T010 is test verification only.
- **Phase 5 (US3)**: After US1; test updates can run in parallel once US1 passes.
- **Phase 6**: After US1–US3.

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 parser tests and T005.
- **US2 (P2)**: Depends on US1 `start.go` rework.
- **US3 (P3)**: Depends on US1 command shape; no new product code except verification.

### Parallel Opportunities

- T002 with T001.
- T003, T004 in parallel after T002.
- T008, T009 parallel after T006.
- T011, T012, T013 parallel after US1.
- T015 parallel with T016.

### Parallel Example: User Story 1

```bash
# Tests first (expect fail):
tests/integration/onewin_usage_test.go
tests/integration/onewin_start_test.go

# Then implementation:
internal/session/file.go
internal/cli/start.go
internal/worker/run.go
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Complete Phase 1–2.
2. T003–T004 (failing tests).
3. T005–T007.
4. Stop and run stand-in start tests only if needed.

### Incremental Delivery

1. US1 → positional casino start + blank start.
2. US2 → headless default and help.
3. US3 → full lifecycle and secrets tests updated.
4. Polish → full `go test ./...`.

---

## Notes

- Hidden worker keeps `--session-file`; only operator `start` changes.
- Feature 004 contracts are amended by `specs/005-start-command-syntax/contracts/`; do not edit 004 task files.
- Do not commit real session files or live-casino screenshots.

---

## Phase 7: Convergence

- [X] T018 Extend `tests/integration/start_help_test.go` to run `start --headless --help` and `start <file> --headless --help`; each must exit `0`, print help, and leave no registry row per FR-007 (partial)
- [X] T019 Extend `tests/integration/onewin_secrets_test.go` to capture the operator `start` argv (positional session path allowed, no `--session-file` on operator `start`) per FR-009 and T012 (partial)