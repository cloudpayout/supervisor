---
description: "Task list for Session Instance Info (019)"
---

# Tasks: Session Instance Info

**Input**: Design documents from `specs/019-session-data/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/cli.md](./contracts/cli.md), [quickstart.md](./quickstart.md)

**Tests**: Included. FR-011 and the constitution require automated CLI coverage before the behavior is complete. Default tests must not open a live casino. Write each story's tests so they fail before that story's implementation tasks.

**Organization**: By user story (US1 read metadata, US2 refuse bad input, US3 help and discovery).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no unfinished dependencies)
- **[Story]**: US1–US3

## Path Conventions

Single Go module at repository root: `internal/cli/`, `internal/registry/`, `internal/casino/`, `tests/unit/`, `tests/integration/`, `cmd/supervisor/`. Branch `019-session-data`. Feature directory `specs/019-session-data`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Lock operator contract before coding

- [x] T001 Review `specs/019-session-data/contracts/cli.md`, `specs/019-session-data/research.md`, and `specs/019-session-data/data-model.md`: four stdout lines `casino` / `game` / `strategy` / `date` in that order; literal `null`; `date` UTC `YYYY-MM-DD HH:MM` via `2006-01-02 15:04`; success stderr empty; failure stdout empty; exit **2** usage / **1** not-found; `store.Update` + `registry.Apply`; game from `casino.Lookup`; no registry schema change

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Command tree shared by every story

**⚠️ CRITICAL**: No user story work begins until this phase is complete

- [x] T002 [P] Add `internal/cli/session.go` with `newSessionCmd()` — parent `Use: "session"`, `Short` describes instance session metadata, no `Run` (subcommands only); add child placeholder wiring target for `info` (FR-001)
- [x] T003 Register `newSessionCmd()` from `internal/cli/root.go` via `root.AddCommand(newSessionCmd())` after existing operator commands (FR-010 partial)
- [x] T004 Add `internal/cli/session_info.go` with `newSessionInfoCmd()`: `Use: "info [instance-id]"`, `Args: cobra.ExactArgs(1)`, trim ID — empty after trim → `fmt.Errorf("instance ID required")` for `isUsageError` exit **2**; `RunE` calls stub `runSessionInfo` returning `not implemented` until US1 (FR-002)

**Checkpoint**: `supervisor session` and `supervisor session info` appear in Cobra tree; `supervisor session info` with no args exits **2**. User stories can start.

---

## Phase 3: User Story 1 — Read session metadata (Priority: P1) 🎯 MVP

**Goal**: `supervisor session info <id>` prints four label lines for a reconciled registry row; blank start shows `casino: null` and `game: null`; onewin stand-in shows `onewin` / `aviator`; `strategy: null`; `date: YYYY-MM-DD HH:MM` UTC (FR-005–FR-009, SC-001–SC-003).

**Independent Test**: `start --headless`, capture id, `session info <id>` → four lines, exit **0**, stderr empty. Onewin session file path → `casino: onewin`, `game: aviator`.

### Tests for User Story 1

- [x] T005 [P] [US1] Add `tests/unit/session_info_format_test.go` for a testable formatter (e.g. `formatSessionInfo(registry.Instance)` in `internal/cli/session_info.go`): blank `Casino` → `casino: null` and `game: null`; `Casino: onewin` → `game: aviator`; `strategy: null` always; `CreatedAt` → `date: 2026-11-26 09:15` style (`2006-01-02 15:04`, UTC, minute truncate)
- [x] T006 [US1] Add `TestSessionInfoBlankStart` in `tests/integration/session_info_test.go` using `isolatedEnv` and `supervisorBin`: `start --headless`, `session info <id>`, assert stdout exactly four lines, stderr empty, exit **0**, no `{` on stdout (SC-001, quickstart Scenario 1)

### Implementation for User Story 1

- [x] T007 [US1] Implement `formatSessionInfo` / `writeSessionInfo` in `internal/cli/session_info.go`: `casino` from row or `null`; `game` via `casino.Lookup` or `null`; `strategy: null`; `date: ` + `inst.CreatedAt.UTC().Format("2006-01-02 15:04")`; one `fmt.Fprintf(os.Stdout, "%s: %s\n", …)` per field in fixed order (FR-005–FR-009)
- [x] T008 [US1] Implement `runSessionInfo` in `internal/cli/session_info.go`: `mustStore()`, `store.Update` with `registry.Apply(doc, time.Now().UTC())`, `Find(id)`; missing row → `fmt.Errorf("instance not found: %s", id)` (defer stderr via `main`, exit **1**); success calls formatter only to stdout (FR-003, FR-004 partial). Any lifecycle state allowed if row exists (research R4)
- [x] T009 [US1] Add onewin stand-in test `TestSessionInfoOnewinStandIn` in `tests/integration/session_info_test.go` reusing existing onewin session helpers: after `start` with session file, `session info <id>` shows `casino: onewin`, `game: aviator`, `strategy: null`, and valid `date:` line (User Story 1 scenarios 2–4; quickstart Scenario 2)

**Checkpoint**: `go test ./tests/integration/ -run TestSessionInfoBlankStart` passes. Unit format test passes. MVP delivers metadata for a running blank instance.

---

## Phase 4: User Story 2 — Refuse invalid or missing targets (Priority: P1)

**Goal**: Usage and not-found leave stdout empty; errors on stderr; no four-line success block (FR-004, FR-005b, SC-004).

**Independent Test**: `session info` with no ID, extra arg, `missing-id` → empty stdout; stderr usage or `instance not found`; exits **2** / **1**.

### Tests for User Story 2

- [x] T010 [P] [US2] Add refusal tests in `tests/integration/session_info_test.go` (or `tests/integration/session_info_refuse_test.go`): no args, extra arg, blank id → exit **2**, `len(stdout)==0`; `missing-id` → exit **1**, stderr contains `instance not found: missing-id`, `len(stdout)==0` (mirror `watch_refuse_test` / `screenshot_refuse_test`; contracts/cli.md failure table)

### Implementation for User Story 2

- [x] T011 [US2] Verify `internal/cli/session_info.go` and `internal/cli/root.go` `Execute` / `isUsageError` paths: no `fmt.Print` to stdout on error returns; Cobra usage only via stderr pattern used by `screenshot` (FR-005b; research R8). Fix any stdout leakage from usage printing

**Checkpoint**: Refusal integration tests pass without Chrome.

---

## Phase 5: User Story 3 — Discover command and keep lifecycle commands unchanged (Priority: P2)

**Goal**: Root help lists `session`; `session info --help` documents four fields and example; no regressions (FR-010, User Story 3).

**Independent Test**: `supervisor help` contains `session` not `worker`; `session info --help` mentions `casino`, `game`, `strategy`, `date`, and plain-text example.

### Tests for User Story 3

- [x] T012 [P] [US3] Extend `tests/integration/root_help_test.go`: root help lists `session` alongside `start`, `status`, `stop`, `screenshot`, `watch`, `logs` (User Story 3 scenario 1)
- [x] T013 [US3] Add `tests/integration/session_info_help_test.go`: `session info --help` mentions required instance ID, four field names, `strategy` always `null`, and a `label: value` example without JSON (User Story 3 scenario 2)

### Implementation for User Story 3

- [x] T014 [US3] Set `Short` and `Long` on `internal/cli/session.go` and `internal/cli/session_info.go` with example block from `specs/019-session-data/contracts/cli.md` (`date: 2026-11-26 09:15`); ensure parent lists `info` subcommand in default help (FR-010)

**Checkpoint**: Help tests pass. `go test ./...` default suite still passes for pre-existing lifecycle tests (SC-005).

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end validation and spec traceability

- [x] T015 [P] Manually or in a comment in `specs/019-session-data/quickstart.md` cross-check: run Scenarios 1–4 from quickstart against built `./supervisor` in isolated env; confirm stdout/stderr matrix
- [x] T016 Run `go test ./...` from repo root; fix any failures; cite spec scenarios in test names or doc comments where missing (FR-011)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — blocks all user stories
- **User Stories (Phase 3–5)**: Depend on Phase 2
- **Polish (Phase 6)**: Depends on US1–US3

### User Story Dependencies

- **US1 (P1)**: After Phase 2 — MVP, no dependency on US2/US3
- **US2 (P1)**: After Phase 2 — tests assume `session info` exists (T008 from US1 or T004 stub); complete US1 core path first for meaningful success/failure contrast
- **US3 (P2)**: After Phase 2 — can parallelize with US2 once US1 registers command

### Parallel Opportunities

```bash
# After Phase 2:
# T005 (unit format tests) || T002 (session.go skeleton) — T005 needs formatter signature from T007 ideally; run T005 after T007 or test pure helpers first

# T010 (refuse tests) || T012 (root help) after US1 implementation

# T012 || T013 in US3
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Phase 1 → Phase 2 → Phase 3 (T005–T009)
2. Stop and validate blank-start `session info`
3. Add US2 refusal tests, then US3 help

### Task Count Summary

| Phase | Tasks | Story |
|-------|-------|-------|
| Setup | T001 | — |
| Foundational | T002–T004 | — |
| US1 | T005–T009 | P1 MVP |
| US2 | T010–T011 | P1 |
| US3 | T012–T014 | P2 |
| Polish | T015–T016 | — |

**Total**: 16 tasks  
**MVP scope**: Phases 1–3 (T001–T009) — blank start session info with four stdout lines

---

## Notes

- Do not add `game` to `registry.Instance` (research R2).
- Do not require `running` state (research R4).
- No worker or chromedp changes.