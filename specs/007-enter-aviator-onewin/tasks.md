---
description: "Task list for entering Aviator on onewin"
---

# Tasks: Enter Aviator on onewin

**Input**: Design documents from `specs/007-enter-aviator-onewin/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: The specification requires the ordinary confirmation on the Aviator stand-in in the default `go test ./...` run. Test tasks are listed before implementation where they gate a story.

**Organization**: Tasks grouped by user story (US1–US5 from [spec.md](./spec.md)).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks in the same batch)
- **[Story]**: US1…US5
- Paths are under repository root unless noted

## Path Conventions

- Product: `internal/casino/`, `internal/session/`, `internal/registry/`, `internal/worker/`, `internal/cli/`
- Tests: `tests/unit/`, `tests/integration/` (`harness_test.go`, `onewin_helpers_test.go`)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No new module dependencies; align comments with feature 007.

- [X] T001 [P] Update `internal/casino/catalog.go` package comment so `aviator` is entered after `start` with an `onewin` session (remove “no command starts the game” wording). No behavior change in this task.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Play URL, session preferences, registry `entry`, and Aviator stand-in. **No user story work until this checkpoint.**

- [X] T002 [P] Add `PlayURL(locale string) string` in `internal/casino/catalog.go`: empty `locale` → `https://1win.com/casino/play/v_spribe:aviator`; non-empty → `https://1win.com/{locale}/casino/play/v_spribe:aviator`. Extend `tests/unit/catalog_test.go` for both forms; refuse invalid identifiers unchanged.
- [X] T003 [P] Extend `internal/session/file.go`: parse optional `currency` (if present exactly `[A-Z]{3}`), `demo` (bool, default false), `locale` (string, non-empty, no `/`). Unknown top-level fields still fail. Export `PlayPreferences`. Extend `tests/unit/session_file_test.go` for valid/invalid optional fields and backward-compatible files without them.
- [X] T004 [P] Add optional `Entry string` on `internal/registry/types.go` (`entering` | `ready`). Bump handling in load/save if schema requires; document in `specs/007-enter-aviator-onewin/data-model.md` only if schema version changes. Update `internal/cli/status.go` to print `entry` when set for an instance.
- [X] T005 Add `newAviatorStandIn` in `tests/integration/aviator_standin_test.go` (or extend `onewin_helpers_test.go`): loopback `http://127.0.0.1:<port>/` HTML with `[data-supervisor-currency-open]`, `[data-supervisor-currency]` options labeled with `USD`/`MXN`, `[data-supervisor-play]`, `[data-supervisor-demo]`, iframe `src` containing `spribelaunch.com/aviator` and inner `div.game-logo` after successful flow per [contracts/aviator-entry.md](./contracts/aviator-entry.md).

**Checkpoint**: Parse preferences, resolve play URL, registry can hold `entry`, stand-in HTML exists.

---

## Phase 3: User Story 1 - Reach Aviator ready after onewin start (Priority: P1) MVP

**Goal**: Cookies, play URL, modal automation, **Aviator ready** within 60s; instance stays running with `entry=ready`.

**Independent Test**: `start` with stand-in URL and session → `status` shows `ready`; no manual clicks (spec US1).

### Tests for User Story 1

- [X] T006 [P] [US1] Add `tests/integration/onewin_aviator_ready_test.go`: `SUPERVISOR_PAGE_URL` = Aviator stand-in; session with `casino` `onewin` and cookie for `127.0.0.1`; `start` exits 0; poll `status` until `ready` or fail within 70s; assert stand-in received navigation. No `1win.com` request.

### Implementation for User Story 1

- [X] T007 [US1] Add `internal/worker/aviator.go`: implement entry per [contracts/aviator-entry.md](./contracts/aviator-entry.md) (modal, play/demo, iframe + `div.game-logo`, 60s deadline). No cookie values in errors.
- [X] T008 [US1] Wire `internal/worker/run.go`: for `onewin` with session, navigate to `SUPERVISOR_PAGE_URL` if set else `casino.PlayURL(prefs.Locale)`; set `entry=entering` on registry; run `aviator.Entry`; on success set `entry=ready`; do not block `start` return (worker-only).
- [X] T009 [US1] Pass `PlayPreferences` from `internal/session/file.go` through `internal/cli/start.go` / worker session read path (path only on argv, not JSON body).

**Checkpoint**: US1 integration test passes; happy path reaches `ready` on stand-in.

---

## Phase 4: User Story 2 - See failure when entry does not complete (Priority: P2)

**Goal**: Currency fail-fast and timeout teardown; instance removed; no CLI failure reason (spec clarify).

**Independent Test**: Session with `"currency":"ZZZ"` on stand-in → instance absent from `status` (US2).

### Tests for User Story 2

- [X] T010 [P] [US2] Add `tests/integration/onewin_aviator_currency_fail_test.go`: stand-in without `ZZZ`; `start` then poll until instance gone; `assertStatusGone`; no secret in registry/logs/cmdline checks mirroring `onewin_secrets_test.go`.
- [X] T011 [P] [US2] Add `tests/integration/onewin_aviator_timeout_test.go`: stand-in variant that never exposes `div.game-logo`; instance removed within ~65s; profile cleaned.

### Implementation for User Story 2

- [X] T012 [US2] In `internal/worker/aviator.go`, on currency not found return typed failure; in `internal/worker/run.go` on entry failure remove registry row and profile (same path as navigation failure). Do not add CLI diagnostic strings.

**Checkpoint**: US2 tests pass.

---

## Phase 5: User Story 3 - Recorded confirmation on a local Aviator stand-in (Priority: P3)

**Goal**: Full `start` → `status` → `screenshot` → `stop` on stand-in; zero live casino (FR-012, SC-003).

**Independent Test**: quickstart ordinary confirmation (US3).

### Tests for User Story 3

- [X] T013 [US3] Add `tests/integration/onewin_aviator_confirmation_test.go`: full chain on Aviator stand-in; assert PNG pixel or DOM marker for post-ready page; `stop`; profile removed; no live host in stand-in hit log.

### Implementation for User Story 3

- [X] T014 [US3] Migrate or narrow existing `tests/integration/onewin_start_test.go` (and related) so `onewin` session starts exercise Aviator entry against stand-in, not balance-only homepage assertion, or split legacy balance checks to explicit non-aviator env if needed per plan.

**Checkpoint**: US3 confirmation test in default `go test ./...`.

---

## Phase 6: User Story 4 - Capture the page after entry (Priority: P4)

**Goal**: `screenshot` after `ready` within 15s; instance still running.

**Independent Test**: After US1 path, `screenshot` succeeds (US4).

### Tests for User Story 4

- [X] T015 [US4] Extend `tests/integration/onewin_screenshot_test.go` (or add `onewin_aviator_screenshot_test.go`): wait for `ready`, then `screenshot`; exit 0 within 15s; instance still `running`; filename must not contain cookie value.

**Checkpoint**: US4 test passes; no change required in `internal/cli/screenshot.go` unless capture timing forces a documented wait.

---

## Phase 7: User Story 5 - Session preferences without new command surface (Priority: P5)

**Goal**: `currency`, `demo`, `locale` defaults and overrides (FR-016).

**Independent Test**: Parse + stand-in runs for each combination (US5).

### Tests for User Story 5

- [X] T016 [P] [US5] Add `tests/integration/onewin_aviator_prefs_test.go`: (a) omit all optional keys → default currency + real play + unprefixed URL path on stand-in; (b) `"demo":true` clicks demo path; (c) `"locale":"es-MX"` requests stand-in path or logged path containing `es-MX` segment when using live URL builder in test hook.

### Implementation for User Story 5

- [X] T017 [US5] Ensure `internal/worker/aviator.go` demo branch activates `[data-supervisor-demo]` on stand-in and localized play labels on live (stub selectors ok until live tuning).

**Checkpoint**: US5 prefs tests pass.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [X] T018 [P] Run `go test ./...` per [quickstart.md](./quickstart.md); fix regressions in `tests/integration/start_*` from play URL change.
- [X] T019 [P] Verify `tests/integration/onewin_secrets_test.go` and `onewin_stop_test.go` still pass with Aviator entry (secrets never in argv, registry, log, screenshot name).
- [X] T020 Update `README.md` only if it still claims homepage-only `onewin` open; cite spec 007 scenario.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1–2**: Block all user stories
- **US1 (P1)**: Blocks MVP; US2–US5 build on entry + stand-in
- **US2**: Needs US1 worker path
- **US3–US4**: Need US1 ready path
- **US5**: Needs US1 + parser from Phase 2

### User Story Dependencies

| Story | Depends on |
|-------|------------|
| US1 | Phase 2 |
| US2 | US1 (`aviator.go`, `run.go`) |
| US3 | US1, stand-in |
| US4 | US1 |
| US5 | Phase 2 T003, US1 |

### Parallel Opportunities

- **Phase 2**: T002, T003, T004 in parallel after T001
- **Tests**: T006, T010, T011, T016 can be authored in parallel (fail until impl)
- **Polish**: T018, T019 parallel

### Parallel Example: Phase 2

```text
T002 catalog PlayURL + unit tests
T003 session PlayPreferences + unit tests  
T004 registry entry + status output
```

---

## Implementation Strategy

### MVP First

1. Phase 1–2
2. Phase 3 (US1) only → validate `onewin_aviator_ready_test.go`
3. Add US2–US5 incrementally

### Suggested MVP scope

**User Story 1** (T006–T009) plus Phase 2.

---

## Task Summary

| Phase | Story | Task IDs | Count |
|-------|-------|----------|-------|
| 1 Setup | — | T001 | 1 |
| 2 Foundational | — | T002–T005 | 4 |
| 3 US1 P1 | US1 | T006–T009 | 4 |
| 4 US2 P2 | US2 | T010–T012 | 3 |
| 5 US3 P3 | US3 | T013–T014 | 2 |
| 6 US4 P4 | US4 | T015 | 1 |
| 7 US5 P5 | US5 | T016–T017 | 2 |
| 8 Polish | — | T018–T020 | 3 |
| **Total** | | **T001–T020** | **20** |

**Format validation**: All tasks use a checkbox, task ID, optional `[P]`, story label where required, and file paths.

## Phase 9: Convergence

- [X] T021 Update `tests/integration/onewin_start_test.go` and `tests/integration/onewin_stop_test.go` to `waitEntry(..., "ready")` before balance, screenshot, and stop assertions so legacy onewin tests validate Aviator entry completion per T014 and US1 (partial)
- [X] T022 Extend `tests/integration/onewin_aviator_prefs_test.go` and stand-in hooks to assert (a) default currency plus real play on unprefixed URL, (b) `demo:true` activates demo control, and (c) `locale:es-MX` navigates a path containing `es-MX` per T016 and US5 (partial)
- [X] T023 Assert internal currency failure in `tests/integration/onewin_aviator_currency_fail_test.go` (worker log or debug channel) while keeping CLI silent, satisfying SC-004 and FR-010 distinction per US2 (partial)
- [X] T024 Add integration coverage that `status` shows `entering` before `ready` for a successful onewin Aviator start per FR-009a and quickstart step 4 (missing)
- [X] T025 Cross-link `specs/005-start-command-syntax/contracts/session-file.md` and `data-model.md` to optional `currency`, `demo`, and `locale` per [007 session-file contract](./contracts/session-file.md) and FR-016 (partial)