---
description: "Task list for headless viewport 1280×720"
---

# Tasks: Headless Viewport 1280×720

**Input**: Design documents from `specs/012-headless-mode-viewport/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/cli.md](./contracts/cli.md), [contracts/worker-viewport.md](./contracts/worker-viewport.md), [quickstart.md](./quickstart.md)

**Tests**: Spec FR-006 and constitution delivery require automated headless dimension coverage. Integration tests in `tests/integration/headless_viewport_test.go` are added in Phase 2 and must fail until Phases 3–5 land.

**Organization**: Tasks grouped by user story; Phase 2 adds red tests before implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no unfinished dependencies)
- **[Story]**: US1, US2, US3
- Include exact file paths in descriptions

## Path Conventions

- Viewport constants: `HeadlessViewportWidth` = **1280**, `HeadlessViewportHeight` = **720** ([data-model.md](./data-model.md))
- Product code: `internal/browser/chromedp.go`, `internal/worker/run.go`, `internal/cli/start.go`
- Tests: `tests/integration/headless_viewport_test.go`, `tests/integration/start_help_test.go`, `internal/browser/chromedp_test.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Align implementer with contracts before code changes.

- [X] T001 Confirm branch `012-headless-mode-viewport`, feature dir `specs/012-headless-mode-viewport/`, and review [contracts/worker-viewport.md](./contracts/worker-viewport.md) (allocator `WindowSize` when headless, `EmulateViewport` before navigate, no viewport on capture child path) and [contracts/cli.md](./contracts/cli.md) (headless screenshot 1280×720, no new flags).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Red integration/unit tests for viewport policy. Do not ship implementation until these exist and fail for missing behavior.

- [X] T002 [P] [US1] Add `tests/integration/headless_viewport_test.go` with `TestHeadlessViewportScreenshot`: `requireChrome`, `isolatedEnv`, `supervisor start --headless`, wait `running`, `screenshot <id>`, `png.Decode` width **1280** height **720** (FR-003, FR-006, SC-005; research R4).

- [X] T003 [P] [US1] In `tests/integration/headless_viewport_test.go`, add `TestHeadlessViewportTwoInstances`: start two `--headless` instances; screenshot each; both PNGs decode to **1280×720** (User Story 1 acceptance scenario 3).

- [X] T004 [P] [US2] In `tests/integration/headless_viewport_test.go`, add `TestHeadedStartRecordsHeaded`: `start` without `--headless`; assert registry row does not record headless true (reuse patterns from `tests/integration/start_flags_test.go`) (FR-002, User Story 2).

- [X] T005 [P] [US3] In `tests/integration/start_help_test.go`, extend `TestStartHelp`: combined output for `start --help` and `start --headless --help` contains **1280** and **720** (or `1280×720`) describing headless layout (FR-007).

- [X] T006 [P] [US2] Add `internal/browser/chromedp_test.go` with `TestAllocatorOptionsHeadlessWindowSize`: when `AllocatorOptions(dir, true)` is used, headless path includes `WindowSize(1280, 720)`; when `AllocatorOptions(dir, false)`, does not force 1280×720 window size and still uses headed visible-window path per [research.md](./research.md) R3.

**Checkpoint**: `go test ./tests/integration/... -run 'HeadlessViewport|StartHelp' -v` fails on dimensions/help; `AllocatorOptions` unit test fails until constants and options exist.

---

## Phase 3: User Story 1 - Predictable headless layout (Priority: P1) 🎯 MVP

**Goal**: Every `--headless` instance uses 1280×720 before navigation; screenshots match (FR-001, FR-003, FR-004).

**Independent Test**: T002 and T003 pass.

### Implementation for User Story 1

- [X] T007 [US1] In `internal/browser/chromedp.go`, add exported constants `HeadlessViewportWidth = 1280` and `HeadlessViewportHeight = 720`; when `headless == true` in `AllocatorOptions`, append `chromedp.WindowSize(HeadlessViewportWidth, HeadlessViewportHeight)` ([contracts/worker-viewport.md](./contracts/worker-viewport.md), FR-001).

- [X] T008 [US1] In `internal/worker/run.go`, when registry `headless` is true, prepend `chromedp.EmulateViewport(browser.HeadlessViewportWidth, browser.HeadlessViewportHeight)` before cookie and `Navigate` actions in the initial `chromedp.Do` open batch—order: emulate → cookies → navigate (FR-001, edge case before `StateRunning`).

- [X] T009 [US1] Verify capture callback in `internal/worker/run.go` does **not** add `EmulateViewport` or `Navigate` on screenshot child context (003 contract; research R6)—change only if regression introduced.

**Checkpoint**: T002, T003, T006 pass. MVP headless viewport complete.

---

## Phase 4: User Story 2 - Headed mode unchanged (Priority: P2)

**Goal**: Headed starts behave as before; no fixed 1280×720 override (FR-002).

**Independent Test**: T004, T006 pass; existing headed integration tests still green.

### Implementation for User Story 2

- [X] T010 [US2] Re-read `AllocatorOptions` and worker open path in `internal/browser/chromedp.go` and `internal/worker/run.go`: confirm `EmulateViewport` and `WindowSize` run only when `headless == true`; headed branch unchanged (`chromedp.VisibleWindow`, no forced metrics) (FR-002).

- [X] T011 [P] [US2] Run `go test ./tests/integration/... -run 'StartFlags|StartLaunch|start_worker' -count=1` and fix only regressions caused by viewport change (SC-002).

**Checkpoint**: Headed lifecycle tests pass; T004 green.

---

## Phase 5: User Story 3 - No new operator flags (Priority: P3)

**Goal**: Existing `--headless` invocation gains viewport; help documents layout (FR-004, FR-007).

**Independent Test**: T005 passes; no new Cobra flags in `start`.

### Implementation for User Story 3

- [X] T012 [US3] Update `internal/cli/start.go`: `--headless` flag `Usage` and/or command `Long` string state headless runs use a **1280×720** layout; do not add new flags (FR-004, FR-007, [contracts/cli.md](./contracts/cli.md)).

- [X] T013 [US3] Confirm `newStartCmd()` in `internal/cli/start.go` still accepts only existing args/flags; run `tests/integration/start_help_test.go` including T005 assertions.

**Checkpoint**: T005 green; operator surface unchanged aside from help text.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Full suite, docs, manual validation.

- [X] T014 [P] Run `go test ./... -count=1` from repository root (constitution delivery; default run without live casino).

- [X] T015 Run manual scenarios in [quickstart.md](./quickstart.md) Scenarios 1–4 when Chrome is available.

- [X] T016 [P] Optional: add `tests/unit/browser_viewport_constants_test.go` asserting `HeadlessViewportWidth == 1280` and `HeadlessViewportHeight == 720` if not covered by T006—skip if redundant.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: Start immediately.
- **Phase 2**: After T001; blocks Phases 3–5 implementation.
- **Phase 3 (US1)**: After Phase 2; MVP.
- **Phase 4 (US2)**: After T007–T008 (headed behavior is the negation of headless paths).
- **Phase 5 (US3)**: Can run in parallel with Phase 4 after US1 code lands; help test T005 needs T012.
- **Phase 6**: After Phases 3–5.

### User Story Dependencies

- **US1 (P1)**: Foundational tests T002–T003 → T007–T009.
- **US2 (P2)**: T004, T006, T010–T011; independent of US3.
- **US3 (P3)**: T005 → T012–T013; depends on US1 for meaningful screenshot help context but help text can ship with US1.

### Parallel Opportunities

```text
# Phase 2 (together after T001):
T002, T003, T004, T005, T006

# After US1 implementation:
T011 (headed regression) || T012 (help text)  # different files

# Phase 6:
T014 || T016
```

---

## Parallel Example: User Story 1

```bash
# Red tests first (Phase 2):
go test ./tests/integration/... -run TestHeadlessViewport -v   # expect FAIL

# Implement T007 + T008, then:
go test ./tests/integration/... -run TestHeadlessViewport -v   # expect PASS
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1–2 (red tests).
2. Complete Phase 3 (T007–T009).
3. **STOP and VALIDATE**: T002–T003, T006.
4. Proceed to US2/US3 and polish.

### Incremental Delivery

1. US1 → headless 1280×720 + screenshot proof (MVP).
2. US2 → headed regression guard.
3. US3 → help text + no new flags.

---

## Notes

- Registry: no `schema_version` bump; no viewport fields (FR-005).
- Do not persist viewport in profile; each headless run reapplies policy.
- If PNG bounds differ by ±1px on a host, document in test with tolerance only if Chrome evidence requires it—default expectation is exact **1280×720**.