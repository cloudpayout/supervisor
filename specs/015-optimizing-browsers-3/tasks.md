---
description: "Task list for aggressive Chrome RAM reduction (015 phase 3)"
---

# Tasks: Aggressive Chrome RAM Reduction (Phase 3)

**Input**: Design documents from `specs/015-optimizing-browsers-3/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/browser-allocator-phase3.md](./contracts/browser-allocator-phase3.md), [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md), [quickstart.md](./quickstart.md); spec **014** implemented on tree.

**Tests**: Spec **FR-008**, **SC-002**, **SC-003** require default integration coverage and allocator regression tests; foundational red tests before US1 implementation.

**Organization**: User stories from [spec.md](./spec.md): **US1** + **US2** P1; **US3** + **US4** P2.

**Note**: `/speckit-tasks` invoked with `slug=optimizing-browsers-4`; active feature directory is **`015-optimizing-browsers-3`** per `.specify/feature.json`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallel when different files and no unfinished dependencies
- **[Story]**: US1–US4
- Exact file paths in every task

## Path Conventions

- Allocator: `internal/browser/chromedp.go`, `internal/browser/chromedp_test.go`, `headless-chrome-flags.txt`
- Tests: `tests/integration/`
- Baseline: `specs/015-optimizing-browsers-3/research.md` **Baseline Record (015)**
- Comparison baseline: `specs/014-optimizing-browsers-2/research.md` **Baseline Record (014)** (`median_014_mib` ~1122.1 MiB shipped)
- Guardrails: `.specify/bugs/try-limit-of-app-config-exceeded-4/`, `headlessUnset*` in `internal/browser/chromedp.go`
- No registry/CLI schema changes ([data-model.md](./data-model.md))

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm feature context and contracts before code changes.

- [x] T001 Confirm branch `015-optimizing-browsers-3`, feature dir `specs/015-optimizing-browsers-3/`, spec **014** phase-2 slice present in `internal/browser/chromedp.go`; read [contracts/browser-allocator-phase3.md](./contracts/browser-allocator-phase3.md), [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md), [research.md](./research.md) §R2–R4, and `headlessPhase2LightweightOptions()` wiring in `internal/browser/chromedp.go`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Red allocator tests, Chrome adopt smoke, baseline inputs. Blocks US1 until tests exist and fail.

**⚠️ CRITICAL**: Do not land phase-3 flags until T002–T005 are written and fail on current code.

- [x] T002 [P] Add `TestHeadlessPhase3AllocatorFlags` in `internal/browser/chromedp_test.go`: for `AllocatorOptions(dir, true, true)` assert adopted phase-3 keys per [contracts/browser-allocator-phase3.md](./contracts/browser-allocator-phase3.md) (`memory-pressure-off=true` when adopt set non-empty); assert 013+014 keys still present; assert `headless && !lightweight` and `headed && lightweight` omit phase-3 keys (FR-001, FR-002).

- [x] T003 [P] Add `TestHeadlessPhase3ForbiddenFlags` in `internal/browser/chromedp_test.go`: default headless lightweight args MUST NOT contain operator **reject** flags (`disable-gpu`, `disable-webgl`, `disable-3d-apis`, `single-process`, `disable-background-networking=true`, aggressive `js-flags` substrings `expose-gc`, `optimize-for-size`, `flush-bytecode`, `gc-global`); MUST still include `enable-webgl` and `use-angle=swiftshader`; bug **-4** unset policy unchanged (FR-003, FR-004, FR-005, research R4).

- [x] T004 [P] Extend `TestLightweightAllocatorTiers` or `TestAllocatorNoLightweightOmitsFullTier` in `internal/browser/chromedp_test.go`: `AllocatorOptions(..., true, false)` omits 013, 014, **and** phase-3 keys; headed lightweight omits all headless RAM slices (data-model, FR-002).

- [x] T005 [P] Smoke Chrome on reference binary: launch once with `--memory-pressure-off=true` (plus minimal headless args); if Chrome exits non-zero or ignores switch, record **empty adopt set** in `specs/015-optimizing-browsers-3/research.md` §R2 row for `memory-pressure-off` and update [contracts/browser-allocator-phase3.md](./contracts/browser-allocator-phase3.md) accordingly (plan Phase A).

- [x] T006 Copy **median_014_mib** (1122.1 or same-day re-measure) into `specs/015-optimizing-browsers-3/research.md` **Baseline Record (015)** before phase-3 measurement (FR-006, SC-001).

**Checkpoint**: `go test ./internal/browser/... -run 'Phase3|Lightweight' -count=1` — phase-3 tests **fail** until US1.

---

## Phase 3: User Story 1 - Evaluate and adopt safe RAM flags (Priority: P1) 🎯 MVP

**Goal**: Append `headlessPhase3LightweightOptions()` after 014 bundle when `headless && lightweight`; only **adopt** rows from operator matrix (expected: `memory-pressure-off`).

**Independent Test**: T002–T004 pass; [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md) matches shipped args; idle median recorded vs 014 in Baseline Record (015).

### Implementation for User Story 1

- [x] T007 [US1] Add `headlessPhase3LightweightOptions()` in `internal/browser/chromedp.go` returning adopt flags from T005 (default: `chromedp.Flag("memory-pressure-off", true)` only); return empty slice if adopt set empty (FR-001, contracts).

- [x] T008 [US1] In `AllocatorOptions` in `internal/browser/chromedp.go`, append `headlessPhase3LightweightOptions()` **only** when `headless && lightweight`, immediately after `headlessPhase2LightweightOptions()`; do not append for headed or `!lightweight` (FR-002, data-model).

- [x] T009 [US1] Ensure no duplicate Chrome keys across 013/014/015 slices (do not re-add skip rows: `process-per-site`, `disable-application-cache`, `v8-cache-options=none`) (FR-001, research R3).

- [x] T010 [P] [US1] Update `headless-chrome-flags.txt` with **SPEC 015** section listing phase-3 adopt flag(s), reject/skip summary, and note headed/`--no-lightweight` omit 015 (SC-004).

- [x] T011 [P] [US1] Run `tests/integration/five_headless_starts_test.go` and `tests/integration/headless_memory_smoke_test.go`; fix only phase-3 regressions (SC-002).

- [x] T012 [US1] Measure idle headless `about:blank` per [quickstart.md](./quickstart.md) §1; fill **median_015_mib** and **sc001_pass** in `specs/015-optimizing-browsers-3/research.md` **Baseline Record (015)** (≥15% lower than **median_014_mib** or `no` with matrix rationale per research R5) (FR-006, SC-001).

**Checkpoint**: T002, T007–T009 green; MVP = default headless start includes 013+014+015 adopt slice when non-empty.

---

## Phase 4: User Story 2 - Automations must not regress (Priority: P1)

**Goal**: Stand-ins and Spribe guardrails unchanged with phase-3 adoptions only (FR-003, SC-003).

**Independent Test**: `go test ./tests/integration/ -run 'OneWin|Aviator|HeadlessViewport|Screenshot'` passes; `headlessUnset*` unchanged.

### Implementation for User Story 2

- [x] T013 [US2] Audit `headlessUnsetBackgroundSuppression`, `headlessUnsetChromedpDefaults`, `headlessUnsetCanvasAndFonts` in `internal/browser/chromedp.go` after T008—confirm phase-3 did not reintroduce bug **-4** forbidden enables (FR-003, research R4).

- [x] T014 [P] [US2] Run `go test ./tests/integration/ -count=1 -timeout 30m -run 'OneWin|Aviator'`; if failure attributable to phase-3, remove adopt flag(s) from `headlessPhase3LightweightOptions()` in `internal/browser/chromedp.go`, sync tests and `headless-chrome-flags.txt`, document in `specs/015-optimizing-browsers-3/research.md` §R2 (SC-003, spec Edge Cases partial adoption).

- [x] T015 [P] [US2] Run `go test ./tests/integration/ -count=1 -run 'HeadlessViewport|Screenshot|StartLaunch|start_worker'`; fix only phase-3 regressions (SC-002, SC-003).

**Checkpoint**: US2 green with phase-3 profile enabled (or documented empty adopt after stand-in failure).

---

## Phase 5: User Story 3 - Operator visibility unchanged (Priority: P2)

**Goal**: `status`, NOTE, `--memory-limit`, `--threshold`, `--no-lightweight` behave exactly as 013/014 (FR-007).

**Independent Test**: Existing memory/status integration tests pass unchanged.

### Implementation for User Story 3

- [x] T016 [P] [US3] Run `go test ./tests/integration/ -count=1 -run 'Memory|NOTE|Status|memory_flags|status_columns|NoteThreshold'`; no code changes unless regression from allocator (FR-007).

- [x] T017 [US3] Assert in `internal/browser/chromedp_test.go` that `AllocatorOptions(..., true, false)` omits phase-3 keys alongside 013/014 (extend T004 or dedicated test) (spec US3 acceptance scenario 4).

**Checkpoint**: US3 verified; no new CLI fields.

---

## Phase 6: User Story 4 - Transparent outcomes when RAM does not drop (Priority: P2)

**Goal**: Operators see adopt/reject/skip matrix, shipped deltas, and baseline vs 014 (FR-001, SC-004, SC-001 alt).

**Independent Test**: `research.md`, contracts, and `headless-chrome-flags.txt` agree with `internal/browser/chromedp.go`.

### Implementation for User Story 4

- [x] T018 [P] [US4] Set `phase3_adopted` column in `specs/015-optimizing-browsers-3/research.md` **Baseline Record (015)** to final shipped adopt list (may be empty) (FR-001).

- [x] T019 [US4] Cross-check shipped flags in `internal/browser/chromedp.go` against `specs/015-optimizing-browsers-3/research.md` §R2 and [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md); fix doc drift in `research.md` or `headless-chrome-flags.txt` (SC-004).

- [x] T020 [P] [US4] Add **SC-001 outcome** subsection to `specs/015-optimizing-browsers-3/quickstart.md` after T012: median vs 014, `sc001_pass`, link Baseline Record; note `--memory-limit` remains warning-only (mirror `specs/014-optimizing-browsers-2/quickstart.md` pattern) (SC-001, FR-007).

**Checkpoint**: US4 docs match code and measured outcome.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Full suite and quickstart validation.

- [x] T021 Run `go test ./... -count=1` and fix any failures attributable to phase-3 (SC-002).

- [x] T022 [P] Execute validation scenarios in [quickstart.md](./quickstart.md) §2–§6; note results in research baseline comment or PR.

- [x] T023 Run `/speckit-converge` or manually verify spec FR-001–FR-008 against tree; file follow-up spec if experimental operator bundle needs opt-in CLI (out of scope for 015).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** → **Phase 3 (US1)** → **Phase 4 (US2)** → **Phase 5–6** can overlap after US2 gate → **Phase 7**

### User Story Dependencies

| Story | Depends on | Notes |
|-------|------------|-------|
| US1 | Phase 2 | Core phase-3 allocator + baseline |
| US2 | US1 T008 | Stand-ins before declaring adopt final |
| US3 | US1 | Regression-only |
| US4 | US1 T012 | Docs need measured medians |

### Parallel Opportunities

- **Phase 2**: T002, T003, T004, T005, T006 in parallel (tests vs smoke vs doc).
- **US1**: T010 ∥ T011 after T009.
- **US2**: T014 ∥ T015 after T013.
- **US3**: T016 parallel with US4 T018–T020 once US2 green.

### Parallel Example: User Story 1

```bash
# After T009 lands:
go test ./internal/browser/... -run Phase3 -count=1
# In parallel:
# T010 headless-chrome-flags.txt
# T011 integration smoke
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Complete Phase 1–2 (red tests + Chrome smoke + 014 baseline copy).
2. Complete Phase 3 through T009 (allocator + docs).
3. Validate T002–T004, integration smoke (T011).
4. Measure T012; only then lock US2 stand-ins.

### Incremental Delivery

1. Ship `memory-pressure-off` only if T005 + T014 pass.
2. If stand-ins fail, ship **empty** phase-3 slice and document in research (valid SC-001 alt).
3. US3–US4 polish without behavior change.

### Suggested MVP scope

**MVP** = T001–T011 with adopt set as proven by T005/T014; feature-complete requires US2–US4 and Phase 7.

---

## Task Summary

| Phase | Tasks | Story |
|-------|-------|-------|
| 1 Setup | T001 | — |
| 2 Foundational | T002–T006 | — |
| 3 US1 | T007–T012 | US1 |
| 4 US2 | T013–T015 | US2 |
| 5 US3 | T016–T017 | US3 |
| 6 US4 | T018–T020 | US4 |
| 7 Polish | T021–T023 | — |

**Total**: **23 tasks** — US1: 6, US2: 3, US3: 2, US4: 3; infrastructure/polish: 9.

**Format validation**: All tasks use `- [x]`, sequential `T001`–`T023`, story labels on US phases, and explicit file paths.

---

## Notes

- Do **not** ship operator **reject** flags on default profile (FR-003–FR-005).
- Do **not** duplicate 013/014 flags (research §R2 skip rows).
- SC-001 is manual in **Baseline Record (015)**, not CI (research R6).
- Reject GPU/WebGL/background-networking/single-process/full js bundle even if idle RSS unchanged (compatibility policy).