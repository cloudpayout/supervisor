---
description: "Task list for aggressive RAM experiment on default headless (016)"
---

# Tasks: Aggressive RAM Experiment (Default Headless)

**Input**: Design documents from `specs/016-optimizing-browsers-experimental/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/browser-allocator-experimental.md](./contracts/browser-allocator-experimental.md), [contracts/no-lightweight-neutral.md](./contracts/no-lightweight-neutral.md), [quickstart.md](./quickstart.md); specs **013–015** implemented.

**Tests**: **FR-007**, **SC-004**, **SC-005** require allocator regression and stand-in/integration coverage; foundational red tests before US1 implementation.

**Organization**: User stories from [spec.md](./spec.md): **US1–US4** are all **P1**.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallel when different files and no unfinished dependencies
- **[Story]**: US1–US4
- Exact file paths in every task

## Path Conventions

- Allocator: `internal/browser/chromedp.go`, `internal/browser/chromedp_test.go`, `headless-chrome-flags.txt`, `lightweight-profile-flags.txt`
- CLI help: `internal/cli/start.go`
- Tests: `tests/integration/`
- Baseline: `specs/016-optimizing-browsers-experimental/research.md` **Baseline Record (016)**; **median_015_mib** from [015 research](../015-optimizing-browsers-3/research.md)
- Guardrails: bug **-4** unset slices in `internal/browser/chromedp.go`; no `disable-background-networking=true` on any headless path

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm feature context and contracts before code changes.

- [x] T001 Confirm branch `016-optimizing-browsers-experimental`, feature dir `specs/016-optimizing-browsers-experimental/`, specs **015** on tree; read [contracts/browser-allocator-experimental.md](./contracts/browser-allocator-experimental.md), [contracts/no-lightweight-neutral.md](./contracts/no-lightweight-neutral.md), [research.md](./research.md) §R1–R3, and current `AllocatorOptions` in `internal/browser/chromedp.go`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Red allocator tests, Chrome smoke, baseline import. Blocks US1 until tests exist and fail.

**⚠️ CRITICAL**: Do not land experimental slice until T002–T005 are written and fail on current code.

- [x] T002 [P] Add `TestHeadlessExperimentalAllocatorFlags` in `internal/browser/chromedp_test.go`: for `AllocatorOptions(dir, true, true)` assert all six experiment switches per [contracts/browser-allocator-experimental.md](./contracts/browser-allocator-experimental.md); assert `js-flags` contains `--max-old-space-size=128`, `expose-gc`, `optimize-for-size`, `flush-bytecode`, `gc-global`; assert headed `(false, true)` and opt-out `(true, false)` omit experiment keys (**SC-002**).

- [x] T003 [P] Add `TestHeadlessExperimentalForbiddenFlags` in `internal/browser/chromedp_test.go`: default headless lightweight MUST NOT contain `enable-webgl`, `use-angle=swiftshader`; MUST have at most one `js-flags`; MUST NOT pass `disable-background-networking=true`; experiment path must still honor bug **-4** unset list from `TestHeadlessPhase2ForbiddenFlags` (research R1, contract).

- [x] T004 [P] Add or extend `TestNoLightweightOmitsExperimentalAndLightweightTiers` in `internal/browser/chromedp_test.go`: `AllocatorOptions(..., true, false)` and `(..., false, false)` omit experiment six flags and 013/014/015 keys (`renderer-process-limit`, `aggressive-cache-discard`, `memory-pressure-off`, etc.) per [contracts/no-lightweight-neutral.md](./contracts/no-lightweight-neutral.md) (**SC-001**, **FR-001**).

- [x] T005 [P] Smoke Chrome on reference binary with experiment six-switch subset (minimal headless argv); record accept/ignore in `specs/016-optimizing-browsers-experimental/research.md` §R3 (plan Phase A).

- [x] T006 Copy **median_015_mib** into `specs/016-optimizing-browsers-experimental/research.md` **Baseline Record (016)** from [015 research](../015-optimizing-browsers-3/research.md) before post-change measurement (**FR-006**).

**Checkpoint**: `go test ./internal/browser/... -run 'Experimental|NoLightweight' -count=1` — experimental tests **fail** until US1.

---

## Phase 3: User Story 1 - Default headless tries rejected RAM flags (Priority: P1) 🎯 MVP

**Goal**: Default `start --headless` applies experiment flag set on top of 013/014/015 with R1 overrides (**FR-003**).

**Independent Test**: T002–T004 pass; default headless argv matches contract.

### Implementation for User Story 1

- [x] T007 [US1] Add `headlessExperimentalLightweightOptions()` in `internal/browser/chromedp.go` with the six switches from [spec.md](./spec.md) Input (**FR-003**).

- [x] T008 [US1] In `AllocatorOptions` in `internal/browser/chromedp.go`, append `headlessExperimentalLightweightOptions()` when `headless && lightweight`, after `headlessPhase3LightweightOptions()`; never append when `!lightweight` (**FR-003**, data-model).

- [x] T009 [US1] When `headless && lightweight`, omit `enable-webgl` and `use-angle=swiftshader` from headless base options; when `headless && !lightweight`, keep pre-016 headless WebGL path unchanged per research R2 (**FR-001** vs default experiment path).

- [x] T010 [US1] Ensure exactly one effective `js-flags` on default headless lightweight: suppress `js-flags=--max-old-space-size=320` from `headlessPhase2LightweightOptions()` when experimental slice is active (research R1, contract).

- [x] T011 [P] [US1] Update `headless-chrome-flags.txt` with **SPEC 016** section; update `lightweight-profile-flags.txt` at repo root to reflect post-016 default headless argv (**SC-002**).

**Checkpoint**: T002–T004 green; MVP = default headless includes experiment bundle.

---

## Phase 4: User Story 2 - Measure whether RAM actually drops (Priority: P1)

**Goal**: Record **median_016_default_mib** vs **median_015_mib** (**FR-006**, **SC-003**).

**Independent Test**: Baseline Record (016) filled per [quickstart.md](./quickstart.md) §1.

### Implementation for User Story 2

- [x] T012 [US2] Measure idle headless `about:blank` for default `start --headless` (150 samples, 30s warmup); fill **median_016_default_mib** and **ram_improvement_pct** in `specs/016-optimizing-browsers-experimental/research.md` **Baseline Record (016)**; document if &lt;10% vs 015 (spec US2).

- [x] T013 [P] [US2] Optional control: measure `start --headless --no-lightweight`; record **median_016_no_lightweight_control_mib** with label `performance_flags=none` in same Baseline Record (research R4).

**Checkpoint**: US2 artifact complete.

---

## Phase 5: User Story 3 - Game and stand-in compatibility (Priority: P1)

**Goal**: Stand-ins on default headless; rollback narrative if fail (**FR-007**, **SC-004**, research R5).

**Independent Test**: `go test ./tests/integration/ -run 'OneWin|Aviator'` outcome recorded in Baseline Record.

### Implementation for User Story 3

- [x] T014 [US3] Audit `headlessUnsetBackgroundSuppression`, `headlessUnsetChromedpDefaults`, `headlessUnsetCanvasAndFonts` in `internal/browser/chromedp.go` after T008—confirm experiment did not add `disable-background-networking=true` or other bug **-4** forbidden enables.

- [x] T015 [P] [US3] Run `go test ./tests/integration/ -count=1 -timeout 30m -run 'OneWin|Aviator'`; set **standins_pass** and **rollback_recommended** in `specs/016-optimizing-browsers-experimental/research.md` **Baseline Record (016)**; if fail, append rollback steps to research §R5 (revert experimental slice, restore WebGL on default headless lightweight).

- [x] T016 [P] [US3] Run `go test ./tests/integration/ -count=1 -run 'HeadlessViewport|Screenshot|StartLaunch|five_headless|headless_memory'`; fix only 016 regressions (**SC-005**).

**Checkpoint**: US3 recorded; feature complete only if rollback path documented when stand-ins fail.

---

## Phase 6: User Story 4 - `--no-lightweight` performance-neutral (Priority: P1)

**Goal**: Help text and tests prove opt-out skips all RAM tuning (**FR-001**, **FR-008**, **SC-006**).

**Independent Test**: T004 passes; `supervisor start --help` matches FR-008.

### Implementation for User Story 4

- [x] T017 [US4] Update `Long` (and flag help if needed) on `newStartCmd()` in `internal/cli/start.go`: default headless uses aggressive RAM switches (compatibility risk); `--no-lightweight` skips **all** lightweight and RAM tuning including experiment; `--memory-limit` is NOTE-only (**FR-008**, **SC-006**).

- [x] T018 [US4] Cross-check shipped argv in `internal/browser/chromedp.go` against [contracts/browser-allocator-experimental.md](./contracts/browser-allocator-experimental.md) and [contracts/no-lightweight-neutral.md](./contracts/no-lightweight-neutral.md); fix drift in `research.md` or `headless-chrome-flags.txt`.

**Checkpoint**: US4 docs and help match code.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Full suite and quickstart validation.

- [x] T019 Run `go test ./... -count=1` and fix failures attributable to 016 (**SC-005**).

- [x] T020 [P] Execute [quickstart.md](./quickstart.md) §2–§6; note results in Baseline Record or PR comment.

- [x] T021 Run `/speckit-converge` or manually verify **FR-001–FR-008** against tree; open follow-up spec for 015 rollback if stand-ins fail and research R5 rollback is executed.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** → **Phase 3 (US1)** → **Phase 4 (US2)** ∧ **Phase 5 (US3)** after US1 → **Phase 6 (US4)** → **Phase 7**

### User Story Dependencies

| Story | Depends on | Notes |
|-------|------------|-------|
| US1 | Phase 2 | Allocator core |
| US2 | US1 T008 | Measurement needs shipped profile |
| US3 | US1 T008 | Stand-ins on default headless |
| US4 | US1 | Help describes shipped behavior |

### Parallel Opportunities

- **Phase 2**: T002–T006 in parallel.
- **US1**: T011 after T010.
- **US2**: T013 ∥ T012 after T012 starts (different arms).
- **US3**: T015 ∥ T016 after T014.
- **US4**: T017 ∥ T018 after US1 green.

### Parallel Example: User Story 1

```bash
go test ./internal/browser/... -run Experimental -count=1
# In parallel after T010:
# T011 headless-chrome-flags.txt + lightweight-profile-flags.txt
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Phase 1–2 (red tests + median_015 copy).
2. Phase 3 through T010.
3. `go test ./internal/browser/... -run 'Experimental|NoLightweight'`.
4. Then US2 measurement and US3 stand-ins before declaring merge-ready.

### Incremental Delivery

1. Ship allocator + unit tests (US1).
2. Measure RAM (US2) and stand-ins (US3); rollback per R5 if needed.
3. Help + doc polish (US4).

### Suggested MVP scope

**MVP** = T001–T011 + T014 + T016 (non-game integration); **merge-ready** requires T015 stand-ins + T012 + T017–T021.

---

## Task Summary

| Phase | Tasks | Story |
|-------|-------|-------|
| 1 Setup | T001 | — |
| 2 Foundational | T002–T006 | — |
| 3 US1 | T007–T011 | US1 |
| 4 US2 | T012–T013 | US2 |
| 5 US3 | T014–T016 | US3 |
| 6 US4 | T017–T018 | US4 |
| 7 Polish | T019–T021 | — |

**Total**: **21 tasks** — US1: 5, US2: 2, US3: 3, US4: 2; infrastructure/polish: 9.

**Format validation**: All tasks use `- [x]`, sequential `T001`–`T021`, story labels on US phases, and explicit file paths.

---

## Notes

- Do **not** add a new CLI flag or registry field (**FR-004**).
- Do **not** put experiment flags on `--no-lightweight` (**FR-001**).
- Do **not** enable `disable-background-networking` on any headless path.
- Default headless **changes** production profile; stand-ins (**T015**) gate completion per plan Constitution **V**.

## Phase 8: Convergence

- [x] T022 Amend `specs/016-optimizing-browsers-experimental/spec.md` Input, **FR-003**, **SC-002**, and US1 AC1 to document Chrome-version conditional omission of `single-process` (shipped off on Chrome 157 per research §R3) so spec matches allocator and `TestHeadlessExperimentalAllocatorFlags` per **FR-003** (`contradicts`)
- [x] T023 Measure idle `about:blank` for `supervisor start --headless --no-lightweight` (same R4 procedure); set **median_016_no_lightweight_control_mib** with `performance_flags=none` in `specs/016-optimizing-browsers-experimental/research.md` Baseline Record (016) per **FR-006** / T013 (`partial`)
- [x] T024 Add `## SC-003 outcome` to `specs/016-optimizing-browsers-experimental/quickstart.md` (015-style): medians, ~2.9% RAM shortfall vs 015, five-flag effective argv, stand-in pass; link Baseline Record per **SC-003** and US2 AC1 (`partial`)
- [x] T025 Update `specs/016-optimizing-browsers-experimental/contracts/browser-allocator-experimental.md` Verification to require five shipped experiment flags and assert `single-process` absent on Chrome 157 (align with tests, not stale “six present”) per plan contract (`partial`)
- [x] T026 Per US2 AC2, when **median_016_default_mib** is not ≥10% below 015, add an explicit follow-up plan in `specs/016-optimizing-browsers-experimental/research.md` §R5 (revert experimental slice → 015-only default) **or** document a dated retention decision why default keeps 016 despite RAM shortfall and passing stand-ins (`partial`)
- [x] T027 Fix `specs/016-optimizing-browsers-experimental/quickstart.md` §2 allocator regression text (“six flags”) to match effective shipped set after T022/T025 per **SC-002** (`partial`)