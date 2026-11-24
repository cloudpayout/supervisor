---
description: "Task list for further Chrome RAM reduction (014 phase 2)"
---

# Tasks: Further Chrome RAM Reduction (Phase 2)

**Input**: Design documents from `specs/014-optimizing-browsers-2/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/browser-allocator-phase2.md](./contracts/browser-allocator-phase2.md), [contracts/operator-memory-limits.md](./contracts/operator-memory-limits.md), [quickstart.md](./quickstart.md); spec **013** implemented on branch.

**Tests**: Spec **FR-005**, **SC-002**, **SC-003** require default integration coverage and allocator regression tests; foundational red tests before US1 implementation.

**Organization**: User stories from [spec.md](./spec.md): **US1** + **US2** P1; **US3** + **US4** P2.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallel when different files and no unfinished dependencies
- **[Story]**: US1–US4
- Exact file paths in every task

## Path Conventions

- Allocator: `internal/browser/chromedp.go`, `internal/browser/chromedp_test.go`, `headless-chrome-flags.txt`
- CLI (help only): `internal/cli/start.go`
- Tests: `tests/integration/`
- Baseline: `specs/014-optimizing-browsers-2/research.md` **Baseline Record (014)**
- Guardrails: `.specify/bugs/try-limit-of-app-config-exceeded-4/`, `headlessUnset*` in `internal/browser/chromedp.go`
- No registry/CLI schema changes ([data-model.md](./data-model.md))

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm feature context and contracts before code changes.

- [x] T001 Confirm branch `014-optimizing-browsers-2`, feature dir `specs/014-optimizing-browsers-2/`, spec **013** present on tree; read [contracts/browser-allocator-phase2.md](./contracts/browser-allocator-phase2.md), [research.md](./research.md) §R2–R4 (skip/forbid lists), and existing `headlessLightweightOptions()` in `internal/browser/chromedp.go`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Red allocator tests and baseline inputs. Blocks US1 implementation until tests exist and fail.

**⚠️ CRITICAL**: Do not land phase-2 flags until T002–T005 are written and fail on current code.

- [x] T002 [P] Add `TestHeadlessPhase2AllocatorFlags` in `internal/browser/chromedp_test.go`: for `AllocatorOptions(dir, true, true)` assert adopted phase-2 keys per [contracts/browser-allocator-phase2.md](./contracts/browser-allocator-phase2.md) (`aggressive-cache-discard`, `disable-back-forward-cache`, `disable-domain-reliability`, merged `disable-features` contains `BackForwardCache`, single `disk-cache-size=1048576`); assert `headless && !lightweight` and `headed && lightweight` omit phase-2 keys (FR-001, FR-005).

- [x] T003 [P] Extend `TestLightweightAllocatorTiers` in `internal/browser/chromedp_test.go`: after phase-2 lands, expect `js-flags=--max-old-space-size=320` **or** `384` per research gate (document constant in test comment); headed-safe and `--no-lightweight` paths still omit 013/014 bundles (FR-001, FR-005).

- [x] T004 [P] Add `TestHeadlessPhase2ForbiddenFlags` in `internal/browser/chromedp_test.go`: default headless lightweight args MUST NOT contain `disable-gpu`, `disable-software-rasterizer`, `disable-background-networking=true`, `enable-features=NetworkService`; MUST still apply bug **-4** unset overrides (`disable-background-networking` passed as false or absent per existing test helpers) (FR-003, research R4).

- [x] T005 Record **median_013_mib** in `specs/014-optimizing-browsers-2/research.md` **Baseline Record (014)** from `specs/013-optimizing-browsers/research.md` post-change row **or** re-measure idle headless `about:blank` with phase-2 slice disabled (FR-006, SC-001).

**Checkpoint**: `go test ./internal/browser/... -run 'Phase2|Lightweight' -count=1` — phase-2 tests **fail** until US1.

---

## Phase 3: User Story 1 - Lower idle RAM beyond spec 013 (Priority: P1) 🎯 MVP

**Goal**: Append `headlessPhase2LightweightOptions()` after 013 bundle when `headless && lightweight`; achieve measurable idle MiB reduction (FR-001, SC-001).

**Independent Test**: T002–T004 pass; idle median in Baseline Record shows ≥15% improvement vs 013 on reference host; default headless integration smoke still passes.

### Implementation for User Story 1

- [x] T006 [US1] Add `headlessPhase2LightweightOptions()` in `internal/browser/chromedp.go` with net-new flags from [contracts/browser-allocator-phase2.md](./contracts/browser-allocator-phase2.md): `aggressive-cache-discard`, `disable-back-forward-cache`, `disable-domain-reliability` (FR-001).

- [x] T007 [US1] In `AllocatorOptions` in `internal/browser/chromedp.go`, append `headlessPhase2LightweightOptions()` **only** when `headless && lightweight`, immediately after `headlessLightweightOptions()`; do not append for headed or `!lightweight` (FR-001, data-model).

- [x] T008 [US1] Ensure **one** effective `disk-cache-size` on the command line: set `1048576` via phase-2 slice and remove or override 013 `5242880` so the final allocator args expose a single `disk-cache-size=1048576` (research R3).

- [x] T009 [US1] Merge phase-2 `disable-features` additions (`BackForwardCache`, `MediaRouter`, `AudioServiceOutOfProcess`) with existing casino `IsolateOrigins,site-per-process` and chromedp defaults without passing `enable-features=NetworkService,NetworkServiceInProcess` (research R3, bug **-4**).

- [x] T010 [P] [US1] Update `headless-chrome-flags.txt` with **SPEC 014** section listing phase-2 flags, replaced `disk-cache-size`, and note that headed/`--no-lightweight` omit 014 (FR-002, SC-004).

- [x] T011 [P] [US1] Re-run `tests/integration/five_headless_starts_test.go` and `tests/integration/headless_memory_smoke_test.go`; fix only phase-2 regressions (SC-005).

- [x] T012 [US1] Fill **median_014_mib** and **sc001_pass** in `specs/014-optimizing-browsers-2/research.md` **Baseline Record (014)** using [quickstart.md](./quickstart.md) §1 procedure; if &lt;15% vs 013, document shortfall and defer `js-flags` step-down per research R3 (SC-001 manual gate).

**Checkpoint**: T002, T006–T009 green; MVP = default headless start includes 013+014 slices.

---

## Phase 4: User Story 2 - Automations stay reliable (Priority: P1)

**Goal**: No regressions on viewport, screenshot, Spribe guardrails, stand-in entry (FR-003, SC-003).

**Independent Test**: `go test ./tests/integration/ -run 'OneWin|Aviator|HeadlessViewport|Screenshot'` passes; `headlessUnset*` unchanged.

### Implementation for User Story 2

- [x] T013 [US2] Audit `headlessUnsetBackgroundSuppression`, `headlessUnsetChromedpDefaults`, `headlessUnsetCanvasAndFonts` in `internal/browser/chromedp.go` after T007—confirm phase-2 did not reintroduce bug **-4** forbidden enables (FR-003, research R4).

- [x] T014 [P] [US2] Run `go test ./tests/integration/ -count=1 -timeout 30m -run 'OneWin|Aviator'`; fix allocator only if new failures attributable to phase-2 flags (SC-003).

- [x] T015 [P] [US2] Run `go test ./tests/integration/ -count=1 -run 'HeadlessViewport|Screenshot|StartLaunch|start_worker'`; fix only phase-2 regressions (SC-002, SC-003).

- [x] T016 [US2] If T014 passes, set `js-flags=--max-old-space-size=320` in phase-2 slice (replace 013 `384` in final args); update T003 expectation and `js_flags_final` in research **Baseline Record**; if T014 fails, keep `384` and note in research (research R3 gate).

**Checkpoint**: US2 green with phase-2 profile enabled.

---

## Phase 5: User Story 3 - Operator visibility unchanged (Priority: P2)

**Goal**: `status`, NOTE, `--memory-limit`, `--threshold`, `--no-lightweight` behave exactly as 013 (FR-004).

**Independent Test**: Existing memory/status integration tests pass unchanged.

### Implementation for User Story 3

- [x] T017 [P] [US3] Run `go test ./tests/integration/ -count=1 -run 'Memory|NOTE|Status|memory_flags|status_columns|NoteThreshold'`; no code changes unless regression from allocator (FR-004, FR-005).

- [x] T018 [US3] Add or extend test in `internal/browser/chromedp_test.go` that `AllocatorOptions(..., true, false)` omits **both** 013 `renderer-process-limit` **and** phase-2 keys (`aggressive-cache-discard`, etc.)—align with `TestAllocatorNoLightweightOmitsFullTier` (FR-001 acceptance scenario 4).

**Checkpoint**: US3 verified; no new CLI fields.

---

## Phase 6: User Story 4 - Documented adopt/reject matrix (Priority: P2)

**Goal**: Operators can distinguish launch optimization vs limit warnings (FR-007, SC-004).

**Independent Test**: Help text and contracts readable; research R2–R4 match shipped code.

### Implementation for User Story 4

- [x] T019 [P] [US4] Extend `Long` (and optionally `Short`) text on `newStartCmd()` in `internal/cli/start.go` to state that `--memory-limit` is a **status warning policy**, not a Chrome RSS cap, and that default headless includes phase-2 RAM flags on top of 013 (FR-007); reference [contracts/operator-memory-limits.md](./contracts/operator-memory-limits.md).

- [x] T020 [US4] Cross-check shipped flags in `internal/browser/chromedp.go` against `specs/014-optimizing-browsers-2/research.md` §R2–R3; fix any doc drift in `research.md` or `headless-chrome-flags.txt` (FR-002, SC-004).

**Checkpoint**: US4 docs match code.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Full suite and quickstart validation.

- [x] T021 Run `go test ./... -count=1` and fix any failures attributable to phase-2 (SC-002).

- [x] T022 [P] Execute validation scenarios in [quickstart.md](./quickstart.md) §2–§6; note results in PR or research baseline comment.

- [x] T023 Run `/speckit-converge` or manually verify spec FR-001–FR-008 against tree; open follow-up tasks in research if `disable-shared-workers` or `js-flags=250` needed (deferred items).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** → **Phase 3 (US1)** → **Phase 4 (US2)** → **Phase 5–6** can overlap after US2 gate → **Phase 7**

### User Story Dependencies

| Story | Depends on | Notes |
|-------|------------|-------|
| US1 | Phase 2 | Core phase-2 allocator |
| US2 | US1 T007 | Stand-ins before `js-flags` 320 (T016) |
| US3 | US1 | Regression-only |
| US4 | US1 | Docs reference shipped flags |

### Parallel Opportunities

- **Phase 2**: T002, T003, T004, T005 in parallel (different test sections / doc).
- **US1**: T010 ∥ T011 after T009.
- **US2**: T014 ∥ T015 after T013.
- **US3**: T017 parallel with US4 T019 once US2 green.

### Parallel Example: User Story 1

```bash
# After T009 lands:
go test ./internal/browser/... -run Phase2 -count=1   # T002–T004
# In parallel:
# T010 headless-chrome-flags.txt
# T011 integration smoke + five headless
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Complete Phase 1–2 (red tests + 013 baseline).
2. Complete Phase 3 through T009 (flags without aggressive `js-flags` step-down).
3. Validate T002, integration smoke, five headless.
4. Measure T012; only then US2 T016 for 320 MiB heap.

### Incremental Delivery

1. Phase-2 flags except V8 step → measure (may meet SC-001).
2. T016 V8 step-down if stand-ins pass.
3. US3–US4 polish without behavior change.

### Suggested MVP scope

**MVP** = T001–T011 + T012 with `js-flags` still **384** if stand-ins not yet run; feature-complete requires US2 T014–T016 and Phase 7.

---

## Task Summary

| Phase | Tasks | Story |
|-------|-------|-------|
| 1 Setup | T001 | — |
| 2 Foundational | T002–T005 | — |
| 3 US1 | T006–T012 | US1 |
| 4 US2 | T013–T016 | US2 |
| 5 US3 | T017–T018 | US3 |
| 6 US4 | T019–T020 | US4 |
| 7 Polish | T021–T023 | — |

**Total**: **23 tasks** — US1: 7, US2: 4, US3: 2, US4: 2; infrastructure/polish: 8.

**Format validation**: All tasks use `- [ ]`, sequential `T001`–`T023`, story labels on US phases, and explicit file paths.

---

## Notes

- Do **not** duplicate 013 flags already in `headlessLightweightOptions()` (research §R2).
- Do **not** add bug **-4** forbidden enables (research §R4).
- SC-001 is manual in research, not CI, unless a reference job is added later (research R6).

## Phase 8: Convergence

**Purpose**: Gaps found by `/speckit-converge` (2026-11-23) — complete before closing the feature.

- [x] T024 Move `disk-cache-size` and `js-flags` out of `headlessLightweightOptions()` into `headlessPhase2LightweightOptions()` in `internal/browser/chromedp.go` so default headless lightweight emits exactly one `--disk-cache-size=1048576` and one `--js-flags=--max-old-space-size=320` on the Chrome command line per plan T008 / FR-001 (partial)

- [x] T025 Add assertions in `internal/browser/chromedp_test.go` (`TestLightweightAllocatorTiers` or new helper) that headless lightweight allocator args contain at most one `disk-cache-size` and one `js-flags` entry after T024 (partial; plan: single winning value)

## Phase 9: Convergence

**Purpose**: Gaps found by `/speckit-converge` (2026-11-23, pass 2). Baseline measurement remains **open T012** (do not duplicate).

- [x] T026 If `sc001_pass` is `no` after **T012**, set `js-flags=--max-old-space-size=250` in `internal/browser/chromedp.go` `headlessPhase2LightweightOptions`, update `internal/browser/chromedp_test.go` and `headless-chrome-flags.txt`, rerun `go test ./tests/integration/ -run 'OneWin|Aviator'`, remeasure using T012 procedure, and update Baseline Record (014) per SC-001 (partial; research R3)

- [x] T027 If SC-001 still fails after T026, adopt `disable-shared-workers=true` in `headlessPhase2LightweightOptions` with stand-in re-run, document in `specs/014-optimizing-browsers-2/research.md` §R3, and remeasure per T012 (partial; research R3 defer)

## Phase 10: Convergence

**Purpose**: Gaps found by `/speckit-converge` (2026-11-23, pass 3). All T001–T027 checked; **SC-001 outcome `no`** on reference host.

- [x] T028 Re-measure **013-only** idle median on the reference host (temporary disable `headlessPhase2LightweightOptions()` append or `--no-lightweight` is wrong—need 013 bundle without 014; use git stash/swap or build flag) with the same 150-sample procedure and Chrome binary as the final 014 window; update `median_013_mib` and recompute `sc001_pass` in `specs/014-optimizing-browsers-2/research.md` Baseline Record (014) per SC-001 (partial)

- [x] T029 Add an **SC-001 outcome** subsection to `specs/014-optimizing-browsers-2/quickstart.md`: record that idle blank-page RSS did not improve ≥15% on the reference host; link Baseline Record; note stand-ins still pass and phase-2 may help non-idle workloads; no change to `--memory-limit` semantics per FR-007 (partial; SC-004)

- [x] T030 If `sc001_pass` remains `no` after T028, revert aggressive steps with no idle RSS win: set `js-flags=--max-old-space-size=320`, remove `disable-shared-workers` from `internal/browser/chromedp.go`, sync `internal/browser/chromedp_test.go`, `headless-chrome-flags.txt`, and `specs/014-optimizing-browsers-2/contracts/browser-allocator-phase2.md`, rerun `go test ./tests/integration/ -run 'OneWin|Aviator'`, and update `specs/014-optimizing-browsers-2/research.md` §R3 verdicts (partial; spec Assumptions follow-on / harm reduction)