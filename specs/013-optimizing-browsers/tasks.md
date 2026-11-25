---
description: "Task list for lightweight browser instances (013)"
---

# Tasks: Lightweight Browser Instances

**Input**: Design documents from `specs/013-optimizing-browsers/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/cli.md](./contracts/cli.md), [contracts/browser-allocator.md](./contracts/browser-allocator.md), [contracts/registry-schema.json](./contracts/registry-schema.json), [quickstart.md](./quickstart.md)

**Tests**: Spec **FR-008** and constitution require default integration coverage for start/status/stop, lightweight on/off, memory flags, and NOTE column. Red tests in Phase 2; implement per user story.

**Organization**: User stories from [spec.md](./spec.md): US1–US3 P1; **US5** P2 (`--no-lightweight`); US4 P3.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallel when different files and no unfinished dependencies
- **[Story]**: US1–US5
- Exact file paths in every task

## Path Conventions

- Allocator: `internal/browser/chromedp.go`, `headless-chrome-flags.txt`
- Memory parse/format: `internal/memlimit/`
- Registry: `internal/registry/types.go`, `internal/registry/registry.go`
- CLI: `internal/cli/start.go`, `internal/cli/status.go`
- Worker: `internal/worker/run.go`
- Tests: `tests/unit/`, `tests/integration/`
- Registry constraints: `schema_version` **2** when any 013 field written; `memory_limit_bytes` **minimum 1**; `memory_threshold_percent` **1–100**; **`lightweight` explicit `true`/`false` on every 013 `start`** ([data-model.md](./data-model.md))

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm feature context before code changes.

- [x] T001 Confirm branch `013-optimizing-browsers`, feature dir `specs/013-optimizing-browsers/`, and read [contracts/browser-allocator.md](./contracts/browser-allocator.md) (headless full, headed-safe, opt-out tiers) and [contracts/cli.md](./contracts/cli.md) (`--no-lightweight`, NOTE column, MiB `current/limit`).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared packages, registry schema, and red tests. Blocks US3; US1 may start T006 after T001.

**⚠️ CRITICAL**: Complete registry + `memlimit` before US3 implementation.

- [x] T002 [P] Add `internal/memlimit/parse.go` with `ParseLimit(string) (uint64, error)`: accept `M`/`MiB` (×1024²), `G`/`GiB` (×1024³); reject `MB`/`GB`, bare numbers, ≤0 (FR-011, research R4).

- [x] T003 [P] Add `internal/memlimit/format.go` with `FormatMiBOneDecimal(bytes uint64) string` for status MEMORY/N (FR-005, research R7).

- [x] T004 Add optional fields to `internal/registry/types.go`: `MemoryLimitBytes uint64` `json:"memory_limit_bytes,omitempty"`, `MemoryThresholdPercent int` `json:"memory_threshold_percent,omitempty"`, `Lightweight bool` `json:"lightweight,omitempty"` (013 `start` MUST write explicit `true` or `false`). In `internal/registry/registry.go` and `types.go`, set document `schema_version` to **2** when writing any instance with memory policy and/or explicit `lightweight` on a new row; on read accept **1 or 2**; missing `lightweight` on v1 rows treated as legacy until stop (FR-018, research R5).

- [x] T005 [P] Add `tests/unit/memlimit_test.go`: `512M` == `512MiB`; `1G`; reject `512MB`, `512`, `0M` (FR-008).

- [x] T006 [P] Extend `internal/browser/chromedp_test.go` with `TestLightweightAllocatorTiers`: `headless && lightweight` includes full adopted flags (`force-device-scale-factor`, `renderer-process-limit`, `disk-cache-size`); `!headless && lightweight` includes **only** `force-device-scale-factor=1`; `!lightweight` includes **neither** tier; absent rejected `disable-gpu`, `disable-software-rasterizer` (FR-002, FR-008, FR-018, research R3, R10).

- [x] T007 [P] Add `tests/integration/memory_flags_usage_test.go`: `start --threshold 90` without `--memory-limit` exits non-zero and prints no new ID; `start --memory-limit 512MB` rejected; after each failed `start`, registry `instances` count unchanged (no phantom row) (FR-012, FR-008, FR-010).

- [x] T008 [P] Add `tests/integration/status_columns_test.go`: assert `status` header contains `NOTE`; row **without** `--memory-limit` has empty NOTE and running MEMORY matches `^\d+\.\d+ MiB$` (no `/` limit suffix); update any test that hard-codes old header only (`tests/integration/status_ten_instances_test.go` if needed) (FR-005, FR-008).

- [x] T032 [P] Add `TestStartInvalidMemoryLimitNoRegistryRow` in `tests/integration/memory_flags_usage_test.go`: capture registry path/instance count before `start --headless --memory-limit 512MB`; assert count unchanged and no new ID in stderr (FR-010; complements T007).

- [ ] T009 Record **pre-change** idle median MiB in `specs/013-optimizing-browsers/research.md` **Baseline Record** table **before** lightweight allocator lands (FR-007, research R10). Use default `start --headless` (lightweight on); do **not** use `--no-lightweight` for SC-001 baseline.

**Checkpoint**: `go test ./tests/unit/... -run Memlimit -v` green for parse; T006 and allocator integration tests **fail** until US1 lands.

---

## Phase 3: User Story 1 - Lower headless memory (Priority: P1) 🎯 MVP

**Goal**: Default headless **full** lightweight Chrome profile (`lightweight=true`); median idle MiB ↓ vs baseline (FR-001, FR-002, SC-001).

**Independent Test**: Headless `start`/`status`/`stop` pass; T006 headless+lightweight branch passes; post-change median documented vs T009 (≥20% reduction target).

### Implementation for User Story 1

- [x] T010 [US1] In `internal/browser/chromedp.go`, add `headlessLightweightOptions()` and `headedSafeLightweightOptions()` per [contracts/browser-allocator.md](./contracts/browser-allocator.md); change `AllocatorOptions(profileDir string, headless bool, lightweight bool)` to append the correct tier **after** WebGL/SwiftShader and `headlessUnset*` blocks (FR-001, FR-003).

- [x] T011 [US1] Confirm **rejected** flags are absent in all tiers: `disable-gpu`, `disable-software-rasterizer`, full operator `js-flags` bundle; headless full tier uses `js-flags=--max-old-space-size=384` only (research R2).

- [x] T012 [P] [US1] Update `headless-chrome-flags.txt` to list full bundle vs headed-safe vs opt-out and reference spec 013.

- [x] T013 [P] [US1] Add `tests/integration/headless_memory_smoke_test.go`: `start --headless`, wait `running`, `status` shows ` MiB` memory; no panic (SC-003 smoke).

- [x] T014 [US1] Append **post-change** median to research **Baseline Record**; verify ≥20% vs T009 or document shortfall in PR (SC-001 manual gate, not CI).

- [x] T035 [P] [US1] Add `tests/integration/five_headless_starts_test.go`: `TestFiveConcurrentHeadlessBlankStarts` — five `start --headless` without session, distinct IDs, all reach `running` when Chrome available (SC-005).

- [x] T036 [US1] In `internal/worker/run.go`, pass registry `inst.Lightweight` into `AllocatorOptions` (default **`true`** when field absent on legacy rows) with `inst.Headless` (FR-018).

- [x] T037 [US1] **(MVP interim only)** In `internal/cli/start.go` `runStart`, persist `Lightweight: true` on every new instance (explicit) and set `schema_version` **2** when first 013 field written (T004). **Superseded by T040** before feature ships: spec **FR-018** is not satisfied until `--no-lightweight` exists.

**Checkpoint**: T006, T013, T035 pass. **MVP** = default lightweight headless only—not feature-complete 013 (US5/T040 required for FR-018).

---

## Phase 4: User Story 2 - Automations remain reliable (Priority: P1)

**Goal**: No regressions on viewport, screenshot, stand-in entry paths (FR-003, SC-004).

**Independent Test**: Existing integration suites for headless viewport, screenshot, start/stop still pass with default lightweight profile.

### Implementation for User Story 2

- [x] T015 [P] [US2] Run `go test ./tests/integration/... -run 'HeadlessViewport|Screenshot|StartLaunch|start_worker' -count=1`; fix only regressions from lightweight flags (SC-004).

- [x] T016 [US2] Re-read `headlessUnsetBackgroundSuppression`, `headlessUnsetCanvasAndFonts`, site-isolation flags in `internal/browser/chromedp.go` and `internal/worker/run.go` headless branch—ensure lightweight block did not reintroduce chromedp defaults that break Spribe (FR-003).

- [x] T017 [P] [US2] If spec 007 stand-in tests exist in repo, run them headless after T010; file issue or adjust guardrails in `research.md` if fail (SC-004).

**Checkpoint**: US2 green with US1 profile enabled.

---

## Phase 5: User Story 3 - Memory limits & status NOTE (Priority: P1)

**Goal**: `--memory-limit`, `--threshold` on `start`; registry persistence; `status` MEMORY `current/limit` + NOTE `threshold exceeded` (FR-011–FR-016, FR-005).

**Independent Test**: T007, T008 extended; limited instance shows `MiB/MiB`; NOTE when RSS ≥ warning bytes.

### Tests for User Story 3

- [x] T018 [P] [US3] In `tests/integration/memory_limit_test.go`, add `TestMemoryLimitParsingOnStart`: `start --headless --memory-limit 512MiB` registers row with limit; status MEMORY matches spaced format (e.g. `\d+\.\d+ MiB/512\.0 MiB`) when running ([contracts/cli.md](./contracts/cli.md), FR-005, FR-014).

- [x] T019 [P] [US3] In `tests/integration/status_note_test.go`, add `TestNoteThresholdExceeded`: start with `--memory-limit 1M --threshold 50`; poll `status` up to **120s** until NOTE contains `threshold exceeded`; **must not skip** (FR-015, research R8).

- [x] T033 [P] [US3] In `tests/integration/memory_limit_test.go`, add `TestMemoryLimitEquivalentSuffixes`: start instance A with `--memory-limit 512M`, instance B with `--memory-limit 512MiB`; both status rows show same limit denominator `512.0 MiB` when running (spec US3 #9, FR-008).

- [x] T034 [US3] Add `tests/integration/memory_start_under_warning_test.go` with `TestStartSucceedsWhilePeerInMemoryWarning`: start A with `--memory-limit 1M --threshold 50` until NOTE warns; `start --headless` B succeeds with new ID (FR-016).

### Implementation for User Story 3

- [x] T020 [US3] In `internal/cli/start.go`, add `--memory-limit` and `--threshold` flags; validate threshold requires limit; default threshold **100** when limit set alone; parse via `memlimit.ParseLimit`; on any validation/parse error exit **before** registry write (no row) (FR-012, FR-013, FR-010).

- [x] T021 [US3] In `internal/cli/start.go` `runStart`, persist `MemoryLimitBytes` and `MemoryThresholdPercent` on new `registry.Instance`; preserve `Lightweight` from T037 when writing memory policy; set `doc.SchemaVersion = 2` when any 013 field written ([data-model.md](./data-model.md)).

- [x] T022 [US3] In `internal/cli/status.go`, add `NOTE` column; implement `formatMemory(inst, snap)` and `noteFor(inst, snap)` using `warning_bytes = limit * threshold / 100` and constant NOTE text `threshold exceeded` (FR-005, FR-015, research R6–R7).

- [x] T023 [US3] Update `internal/cli/status.go` header to `ID\tSTATE\tENTRY\tCPU%\tMEMORY\tNOTE`; `starting` rows use `unavailable/{limit} MiB` when limit configured.

- [x] T024 [P] [US3] Update `internal/cli/start.go` and `internal/cli/status.go` command `Long` and flag help for MiB suffix grammar (FR-017: state figures are **mebibytes**, not decimal MB/GB), default 100% threshold, MEMORY spacing (`245.3 MiB/512.0 MiB`), NOTE `threshold exceeded`, and when NOTE is empty. `--no-lightweight` help completed in T041 (FR-009).

**Checkpoint**: T007, T008, T018, T019, T033, T034 pass; quickstart §4–§5 runnable.

---

## Phase 6: User Story 5 - Opt out of lightweight profile (Priority: P2)

**Goal**: `--no-lightweight` on `start` for headed and headless; registry `lightweight: false`; worker honors opt-out (FR-018, US1 scenarios 4–5).

**Independent Test**: Headless with/without flag; headed with flag; mixed fleet; T006 contrasting `!lightweight` branch.

### Tests for User Story 5

- [x] T038 [P] [US5] Add `tests/integration/lightweight_flag_test.go`: `start --headless --no-lightweight` lifecycle pass; parallel default headless instance retains full bundle; no cross-instance leakage. **Requires T039 green** (allocator contrast—not optional smoke) (spec US1 #4–#5, US5).

- [x] T039 [P] [US5] Extend `internal/browser/chromedp_test.go` with `TestAllocatorNoLightweightOmitsFullTier`: `AllocatorOptions(..., headless=true, lightweight=false)` omits `renderer-process-limit` and other full-tier-only flags; default headless+lightweight still includes them (FR-008, FR-018).

### Implementation for User Story 5

- [x] T040 [US5] In `internal/cli/start.go`, add `--no-lightweight` flag: when set persist `Lightweight: false`; when omitted persist `Lightweight: true`; independent of `--memory-limit` / `--threshold` (FR-018). Replace T037 hardcoded-always-true with this logic.

- [x] T041 [P] [US5] Update `internal/cli/start.go` `--help` / `Long`: document `--no-lightweight`, defaults for headed vs headless, independence from memory flags (FR-009, SC-002).

**Checkpoint**: quickstart §2–§3 and §6 mixed fleet; US5 acceptance scenarios 1–3 pass.

---

## Phase 7: User Story 4 - Headed debugging usable (Priority: P3)

**Goal**: Headed default uses **headed-safe** tier only; headed `--no-lightweight` skips all 013 extras; optional memory flags (FR-002, User Story 4).

**Independent Test**: T006 headed branches; visible window; no headless-only flags on headed runs.

### Implementation for User Story 4

- [x] T025 [US4] Verify in `internal/browser/chromedp_test.go` or T006: `AllocatorOptions(..., headless=false, lightweight=true)` does **not** include `renderer-process-limit` or V8/cache bundle; `lightweight=false` adds no 013 extras (spec US4 #3–#4).

- [x] T026 [P] [US4] Run `go test ./tests/integration/... -run 'StartFlags|Headed' -count=1` (User Story 4 acceptance).

- [x] T027 [P] [US4] Add or extend integration in `tests/integration/lightweight_flag_test.go`: headed `start --no-lightweight` shows window; headed default `start` headed-safe only; headed `start --memory-limit 512M` shows limit formatting without full headless bundle (mixed fleet).

**Checkpoint**: Headed + headless fleet per spec edge cases.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [x] T028 [P] Run `go test ./... -count=1` (constitution default suite, no live casino); confirms FR-004 via existing multi-instance tests (e.g. `tests/integration/status_ten_instances_test.go`); FR-010 covered by T007/T032.

- [x] T029 Execute [quickstart.md](./quickstart.md) scenarios **1–8** including §8 SC-006 (stop removes row; compare running MEMORY to Baseline Record); note baseline in PR.

- [x] T030 [P] Align `specs/001-supervisor-cli/contracts/registry-schema.json` or document supersession by `specs/013-optimizing-browsers/contracts/registry-schema.json` if monolith schema is maintained; if a repo-level operator doc exists, add one line that Supervisor memory units are **MiB** (FR-017).

- [x] T031 [P] Extend `tests/integration/start_help_test.go` (or add `status_help_test.go`): assert `start --help` mentions `--no-lightweight`, `--memory-limit`, `--threshold`, MiB; `status --help` mentions NOTE and `threshold exceeded` (FR-009, SC-002).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** → **US1 (Phase 3)** can overlap Phase 2 test authoring
- **US3 (Phase 5)** requires **T002–T004** (memlimit + registry)
- **US2 (Phase 4)** after **T010** (allocator tiers)
- **US5 (Phase 6)** after **US1** T010, T036, T037 (allocator + worker + default lightweight write)
- **US4 (Phase 7)** after **US5** (headed-safe vs opt-out both need CLI flag)
- **Phase 8** last

### User Story Dependencies

| Story | Depends on | Independent test |
|-------|------------|------------------|
| US1 | T001, T006 red→green via T010 | T013, T035, baseline T009/T014 |
| US2 | US1 | T015–T017 |
| US3 | T002–T004, T007–T008 | T018–T019, T033–T034, T020–T024 |
| US5 | US1 | T038–T041 |
| US4 | US5 | T025–T027 |

### Parallel Examples

```bash
# Phase 2 parallel:
T002 memlimit/parse.go + T003 format.go + T005 unit tests
T006 chromedp_test.go + T007 memory_flags_usage_test.go + T008 status_columns_test.go

# US3 tests before impl:
T018 memory_limit_test.go + T019 status_note_test.go + T033

# US5 parallel:
T038 lightweight_flag_test.go + T039 allocator contrast + T041 help text
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. T001 → T006 (red) → **T010–T014, T035–T037** (lightweight headless + worker/default registry)
2. **STOP**: `go test` + baseline SC-001
3. Add US2 → US3 → **US5 (mandatory for FR-018)** → US4 → Polish

**Feature-complete 013** requires through **T040** (`--no-lightweight`), not MVP alone.

### Suggested delivery order

**T001 → T002–T009,T032 → T010–T014,T035–T037 (MVP) → T015–T017 → T018–T019,T033,T034 → T020–T024 → T038–T041 → T025–T027 → T028–T031**

### Task counts

| Phase | Tasks | Story |
|-------|-------|-------|
| Setup | 1 | — |
| Foundational | 9 | — |
| US1 | 8 | P1 MVP |
| US2 | 3 | P1 |
| US3 | 9 | P1 |
| US5 | 4 | P2 |
| US4 | 3 | P3 |
| Polish | 4 | — |
| **Total** | **41** | |

**Parallel opportunities**: T002–T003, T005–T008, T032, T012–T013, T035, T015+T017, T018–T019+T033, T024+T038+T039+T041, T026+T027, T028+T030+T031

---

## Notes

- SC-001 and T009/T014 use default `start --headless` only (not `--no-lightweight`).
- **T037** = MVP interim only; **T040** is required before merge/feature sign-off (FR-018).
- US1 acceptance #4–#5 and US5 share behavior; integration coverage is **T038+T039** (not duplicated scenarios).
- Task IDs T001–T041 are sequential (T032 = FR-010 registry guard).
- All tasks use checklist format `- [ ] Tnnn ...` with file paths for LLM execution.

---

## Phase 9: Convergence

- [x] T042 Run `go test ./tests/integration/ -count=1 -timeout 30m -run 'Onewin|onewin'` on reference host with Chrome; fix allocator guardrails in `internal/browser/chromedp.go` and note outcome in `specs/013-optimizing-browsers/research.md` if any fail (SC-004, partial)
- [x] T043 In `tests/integration/status_ten_instances_test.go`, assert `status -f` / one-shot output includes **NOTE** column header alongside CPU% and MEMORY (FR-005, partial)
- [ ] T044 Complete open baseline work: capture `pre_median_mib` and `post_median_mib` per quickstart §1 into `specs/013-optimizing-browsers/research.md` Baseline Record (`host`, `chrome_binary`, `sc001_pass`); requires default `start --headless` without `--no-lightweight` (FR-007, SC-001, missing)
- [x] T045 Execute `specs/013-optimizing-browsers/quickstart.md` scenarios 1–8 including §8 SC-006; record results in PR notes (T029, partial)