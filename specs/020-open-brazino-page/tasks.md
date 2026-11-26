---
description: "Task list for opening the Brazino home page with a session"
---

# Tasks: Open the Brazino page

**Input**: Design documents from `specs/020-open-brazino-page/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: The specification requires a `brazino` ordinary confirmation in the default test run (`start`, `status`, `screenshot`, `stop` against the shared local stand-in from `tests/integration/onewin_helpers_test.go`). Test tasks are written before implementation where they define new failure modes.

**Organization**: Tasks are grouped by user story. This feature is additive on top of `specs/004-open-onewin-page`; do not regress `onewin` (SC-006).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- Product code: `internal/casino/catalog.go`, `internal/worker/run.go`, `internal/session/file.go`, `internal/cli/session_info.go`
- Tests: `tests/unit/`, `tests/integration/` using `supervisorBin`, `isolatedEnv`, `requireChrome`, `newStandIn`, and `startCasino` from `tests/integration/onewin_helpers_test.go` and `tests/integration/onewin_start_test.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No new dependencies or packages.

- [x] T001 Leave `go.mod` on module `github.com/cloudpayout/supervisor` and Go 1.27.1 with no new requirement. Reuse `net/http` stand-in helper only.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Catalog and worker navigation for `brazino`. No user story completes until this phase is done.

- [x] T002 Add `brazino` to `internal/casino/catalog.go`. Identifier `brazino`. One or more lowercase ASCII letters. Page `https://www.brazino777.mx/`. Game field empty (no registered game). `Lookup` continues to return `onewin` unchanged. Unsupported identifiers keep today's usage error.
- [x] T003 In `internal/casino/catalog.go`, add casino-aware start navigation (e.g. `StartURL(casinoID, locale string)`): when catalog `Game` is empty, return registered `Page`; for `onewin`, preserve today's Aviator `PlayURL(locale)` behavior. Refuse unknown `casinoID`.
- [x] T004 In `internal/worker/run.go`, change `openTarget` to use the navigation helper from T003 instead of calling `casino.PlayURL(locale)` unconditionally. `SUPERVISOR_PAGE_URL` override and `standInPage` locale rewrite behave as for `onewin`.

**Checkpoint**: `Lookup("brazino")` resolves; worker navigates to stand-in or `https://www.brazino777.mx/` when unset; `onewin` play URL unchanged.

---

## Phase 3: User Story 1 - Open the Brazino home page with a session (Priority: P1)

**Goal**: `supervisor start <session-file>` with `"casino":"brazino"` applies cookies and opens the registered home page (or stand-in under test).

**Independent Test**: Shared `newStandIn` shows `balance` only with the expected cookie; `brazino` session file on `127.0.0.1`; no navigation to `www.brazino777.mx` in default tests.

### Tests for User Story 1

- [x] T005 [P] [US1] Extend `tests/unit/catalog_test.go`. `brazino` resolves to page `https://www.brazino777.mx/` and empty `Game`. `Brazino`, `brazino777`, and unsupported ids are refused. Assert start navigation URL for `brazino` is the home page, not an Aviator path.
- [x] T006 [P] [US1] Extend `tests/unit/session_file_test.go`. Accept `{"casino":"brazino",...}` when `Lookup` succeeds. Keep refusal for unknown `casino` values. `casino` field required; must match supported catalog entry.
- [x] T007 [P] [US1] Add `tests/integration/brazino_usage_test.go` without Chrome. Bad session paths and documents with unsupported casino strings exit non-zero with no instance id and no profile, consistent with `tests/integration/onewin_usage_test.go`.
- [x] T008 [US1] Add `tests/integration/brazino_start_test.go` using `requireChrome`, `newStandIn`, and `startCasino`. Set `SUPERVISOR_PAGE_URL`. Session file `casino` is `brazino`. Cookie domain `127.0.0.1`. Assert shown/withheld `balance`, registry `casino: brazino`, and no live Mex host in test. Do not add a second stand-in implementation file.

### Implementation for User Story 1

- [x] T009 [US1] After T002–T004, run and fix `tests/integration/brazino_start_test.go` until green. Blank `start` without a session path remains unchanged. No new operator flags.

**Checkpoint**: Brazino page opens on the shared stand-in with a valid session file.

---

## Phase 4: User Story 2 - Capture the page that start opened (Priority: P2)

**Goal**: `screenshot <id>` captures the `brazino` home/stand-in page; instance stays running.

**Independent Test**: Pixel check matches shown/withheld colors within 15 seconds; `status` still lists the instance.

### Tests for User Story 2

- [x] T010 [US2] Add `tests/integration/brazino_screenshot_test.go` mirroring `tests/integration/onewin_screenshot_test.go` with `casino: brazino`. No changes to `internal/cli/screenshot.go`.

**Checkpoint**: Screenshot command works for `brazino` instances without new flags.

---

## Phase 5: User Story 3 - See the run and end it (Priority: P3)

**Goal**: `status`, `stop`, and profile removal for `brazino` match `onewin` lifecycle (FR-007).

**Independent Test**: After start, `status` lists id; after `screenshot`, still listed; after `stop`, id and profile gone.

### Tests for User Story 3

- [x] T011 [US3] Add `tests/integration/brazino_stop_test.go` mirroring `tests/integration/onewin_stop_test.go` with `casino: brazino`. Include failed-start profile cleanup case.

**Checkpoint**: Ending a `brazino` run removes browser data.

---

## Phase 6: User Story 4 - Keep the session out of the project and the command (Priority: P4)

**Goal**: Cookie secrets absent from argv, worker log, registry, screenshot filename (FR-005).

**Independent Test**: Runtime-generated secret not recoverable from those surfaces.

### Tests for User Story 4

- [x] T012 [US4] Add `tests/integration/brazino_secrets_test.go` mirroring `tests/integration/onewin_secrets_test.go` with `casino: brazino`. Registry row has `casino: brazino` only; no cookie or page URL stored.

**Checkpoint**: Secrets handling matches 004 for the second casino.

---

## Phase 7: Session info (FR-014)

**Goal**: `session info <id>` prints `casino: brazino` and `game: null` for a Brazino run.

- [x] T013 [P] [US1] Add `tests/integration/brazino_session_info_test.go`. After stand-in start, `session info` stdout is exactly four lines with `game: null` per `specs/019-session-data`.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Quickstart alignment, full suite green, spec trace comments, non-regression.

- [x] T014 [P] Walk `specs/020-open-brazino-page/quickstart.md` against `specs/020-open-brazino-page/contracts/cli.md` and `specs/020-open-brazino-page/contracts/session-file.md`. Fix test or contract mismatches under `tests/integration/brazino_*`.
- [x] T015 Run `go test ./...` from the repository root. All `tests/integration/onewin_*` and Aviator tests pass unchanged (SC-006). New `brazino_*` tests pass. Default run does not open `https://www.brazino777.mx/` or `https://1win.com/` (SC-007).
- [x] T016 [P] Add spec/task citations in the existing FR-comment style on new or changed behavior in `internal/casino/catalog.go` and `internal/worker/run.go` (spec `020-open-brazino-page`, tasks T002–T004).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: T001 — no dependencies.
- **Phase 2**: T002 → T003 → T004 (same files; sequential).
- **Phase 3**: Depends on Phase 2. T005–T007 parallel tests may be written before T002; they fail until T002–T004 land. T008 after T005–T007 drafted. T009 after T004 and T008.
- **Phase 4–6**: Depend on Phase 3 checkpoint (running `brazino` instance). T010, T011, T012 can be authored in parallel after T009.
- **Phase 7**: T013 depends on T009 (running instance).
- **Phase 8**: T015 after all tests and implementation; T014 and T016 in parallel with T015 prep.

### User Story Dependencies

- **US1 (P1)**: Foundational catalog/navigation only.
- **US2 (P2)**: Needs US1 start on stand-in.
- **US3 (P3)**: Needs US1; independent of screenshot pixel details.
- **US4 (P4)**: Needs US1 and US2 (screenshot filename check).

### Parallel Opportunities

- T005, T006, T007, T013 (tests) — different files after Phase 2.
- T010, T011, T012 — different files after US1.
- T014, T016 — different files during polish.

---

## Parallel Example: User Story 1 (tests)

```text
T005 tests/unit/catalog_test.go
T006 tests/unit/session_file_test.go
T007 tests/integration/brazino_usage_test.go
```

Then T008 `tests/integration/brazino_start_test.go`, then foundational T002–T004, then T009 to green.

---

## Implementation Strategy

### MVP First (User Story 1)

1. Complete Phase 1–2 (catalog + worker navigation).
2. Complete Phase 3 through T009.
3. Stop and run `go test ./tests/integration -run Brazino -count=1` (or package-wide when ready).

### Incremental Delivery

1. US1: open Brazino page on stand-in.
2. US2: screenshot that page.
3. US3: stop and profile removal.
4. US4: secrets audit.
5. T013: session info lines.
6. Phase 8: full `go test ./...`.

### Non-Goals (this feature)

- No new stand-in server type (reuse `newStandIn`).
- No changes to `specs/004-open-onewin-page` artifacts.
- No live `www.brazino777.mx` in default CI.

---

## Notes

- Every task line uses checkbox, ID, optional `[P]`, story label on story phases only, and a file path.
- Operator CLI is `start [path]`; session file carries `casino: brazino` (005 syntax), not `--casino`.
- Live home page check in quickstart stays manual and outside `go test ./...`.