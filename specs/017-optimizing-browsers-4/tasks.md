---
description: "Task list for headless shell, phase-4 flags, ephemeral profiles (017)"
---

# Tasks: Headless Shell and Performance Flags (Phase 4)

**Input**: Design documents from `specs/017-optimizing-browsers-4/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md); specs **013–015** implemented; **016** may be on tree to remove.

**Tests**: **SC-003**–**SC-007** and constitution delivery rules require allocator regression, headless-shell preflight, ephemeral profile integration, and stand-in coverage; foundational red tests before reversions and new slices.

**Organization**: User stories from [spec.md](./spec.md): **US1–US4** are **P1**; **US5** is **P2**. Phase numbers **3–7** follow dependency order, not numeric story IDs (e.g. **Phase 5 = US4**, **Phase 6 = US3**).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallel when different files and no unfinished dependencies
- **[Story]**: US1–US5
- Exact file paths in every task

## Path Conventions

- Allocator: `internal/browser/chromedp.go`, `internal/browser/chromedp_test.go`, `internal/browser/executable.go`, `internal/browser/executable_test.go`
- Shell: `scripts/fetch-chrome-headless-shell.sh`, `bin/chrome-headless-shell` (gitignored)
- CLI: `internal/cli/start.go`
- Worker / stop: `internal/worker/run.go`, `internal/cli/stop.go`, `internal/registry/reconcile.go`
- Integration: `tests/integration/`
- Docs: `headless-chrome-flags.txt`, `lightweight-profile-flags.txt`
- Baseline: `specs/017-optimizing-browsers-4/research.md` **Baseline Record (017)**; **median_015_mib** from [015 research](../015-optimizing-browsers-3/research.md)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm feature context and contracts.

- [x] T001 Confirm branch `017-optimizing-browsers-4`, feature dir `specs/017-optimizing-browsers-4/`, read [contracts/headless-shell-binary.md](./contracts/headless-shell-binary.md), [contracts/browser-allocator-phase4.md](./contracts/browser-allocator-phase4.md), [contracts/ephemeral-profile.md](./contracts/ephemeral-profile.md), [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md), and current `internal/browser/chromedp.go` (016 experimental slice if present).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Provision headless shell, import baseline, red tests. Blocks all user stories until complete.

**⚠️ CRITICAL**: Do not remove 016 or add phase-4 until T004–T008, T032–T033 exist and fail appropriately on current tree. Do **not** delete `TestHeadlessExperimental*` until **T010** (US1 revert).

- [x] T002 [P] Add `scripts/fetch-chrome-headless-shell.sh` per [research.md](./research.md) R4; pin **157.0.8103.78** chrome-headless-shell-linux64; extract to `bin/chrome-headless-shell`; append `bin/chrome-headless-shell` to `.gitignore` (**FR-005**).

- [x] T003 Run `scripts/fetch-chrome-headless-shell.sh` on reference host; record pinned version in `specs/017-optimizing-browsers-4/research.md` Baseline Record (017) **headless_shell_version** (blocking for local implement).

- [x] T004 [P] Copy **median_015_mib** into `specs/017-optimizing-browsers-4/research.md` **Baseline Record (017)** from [015 research](../015-optimizing-browsers-3/research.md) (**FR-008**).

- [x] T005 [P] Add `TestHeadlessPhase4AllocatorFlags` in `internal/browser/chromedp_test.go` **alongside** existing 016 tests (remove 016 tests only in **T010**); assert adopt flags from [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md) and **no** 016 experiment keys (`disable-gpu`/`disable-webgl` wholesale, 128 js bundle); assert `enable-webgl` present on default headless lightweight (**FR-011**, **US1**).

- [x] T006 [P] Add `TestHeadlessPhase4ForbiddenFlags` and `TestNoLightweightOmitsPhase4Tiers` in `internal/browser/chromedp_test.go`: no `disable-background-networking=true`; no default `no-sandbox`, `single-process`, `no-zygote` unless matrix marks adopt; `AllocatorOptions(..., true, false)` omits phase-4 keys and 013/014/015 RAM tiers (**FR-002**, **FR-003**, **FR-004**, US1 AC4).

- [x] T007 [P] Add `TestHeadlessShellExecPath` in `internal/browser/executable_test.go`: `AllocatorOptions` with `headless` uses bundled shell path from `HeadlessShellPath()` in `internal/browser/executable.go` when file exists (**SC-002**).

- [x] T008 [P] Add `TestHeadlessIgnoresChromePath` in `tests/integration/headless_shell_binary_test.go`: with `CHROME_PATH=/usr/bin/google-chrome`, headless start uses `chrome-headless-shell` in process command line or allocator smoke (**FR-007**, **SC-002**).

- [x] T009 [P] Add `TestProfileRemovedAfterStop` in `tests/integration/profile_ephemeral_test.go`: start headless → profile dir exists while `running` → `stop` → directory absent (**SC-006**, [contracts/ephemeral-profile.md](./contracts/ephemeral-profile.md)).

- [x] T032 [P] Add `TestProfileUniquePerInstance` in `tests/integration/profile_ephemeral_test.go`: two sequential headless start/stop cycles; second `profile_dir` MUST differ from first (**FR-012**).

- [x] T033 [P] Add `tests/integration/headless_shell_require.go` helper: skip or fail fast when `bin/chrome-headless-shell` missing; use in headless integration tests; document in `README.md` Development and [quickstart.md](./quickstart.md) Prerequisites (**R8**, research R8).

**Checkpoint**: `go test ./internal/browser/... -run 'Phase4|HeadlessShell' -count=1` and new integration tests **fail** until US1–US2–US4 land; run **T003** before headless integration.

---

## Phase 3: User Story 1 - Adopt safe performance flags on 015 baseline (Priority: P1) 🎯 MVP

**Goal**: Revert **016**; restore **015** profile; append phase-4 adopt slice only (**FR-001**, **FR-002**, **FR-011**).

**Independent Test**: T005–T006 pass; argv matches [contracts/browser-allocator-phase4.md](./contracts/browser-allocator-phase4.md).

### Implementation for User Story 1

- [x] T010 [US1] Remove `headlessExperimentalLightweightOptions()` and its call from `internal/browser/chromedp.go`; remove obsolete `TestHeadlessExperimental*` from `internal/browser/chromedp_test.go`; restore `enable-webgl` + `use-angle=swiftshader` when `headless && lightweight`; restore phase-2 `js-flags=--max-old-space-size=320` in `headlessPhase2LightweightOptions()` (**FR-011**).

- [x] T011 [US1] Finalize adopt list in `specs/017-optimizing-browsers-4/contracts/operator-flag-matrix.md` and `specs/017-optimizing-browsers-4/research.md` R2 after Chrome headless-shell smoke on candidate flags (**FR-001**).

- [x] T012 [US1] Implement `headlessPhase4LightweightOptions()` in `internal/browser/chromedp.go` with **adopt** switches only; merge extra `disable-features` tokens via `casinoDisableFeatures` or dedicated flags without breaking IsolateOrigins merge; note FR-014 adopts in [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md) and research R2 (**FR-002**, **FR-014**).

- [x] T013 [US1] Append `headlessPhase4LightweightOptions()` in `AllocatorOptions` when `headless && lightweight`, after `headlessPhase3LightweightOptions()`; never when `!lightweight` ([data-model.md](./data-model.md)).

- [x] T014 [P] [US1] Update `headless-chrome-flags.txt` **SPEC 017** section and `lightweight-profile-flags.txt` for post-017 default headless argv (015 + phase-4; no 016) (**SC-005**).

**Checkpoint**: US1 allocator tests green; default headless matches 015 + adopts.

---

## Phase 4: User Story 2 - Default headless uses chrome-headless-shell (Priority: P1)

**Goal**: Bundled shell only for headless; ignore `CHROME_PATH` for full Chrome (**FR-006**, **FR-007**).

**Independent Test**: T007–T008 pass; missing bundle fails preflight.

### Implementation for User Story 2

- [x] T015 [US2] Add `HeadlessShellPath()` and headed `Executable()` in `internal/browser/executable.go` only (move from `chromedp.go` if duplicated); resolve repo `bin/chrome-headless-shell`, validate executable (**FR-005**, [contracts/headless-shell-binary.md](./contracts/headless-shell-binary.md)).

- [x] T016 [US2] In `AllocatorOptions`, set `chromedp.ExecPath` from `HeadlessShellPath()` when `headless`; headed keeps `Executable()` + `CHROME_PATH` behavior (**FR-006**).

- [x] T017 [US2] In `internal/cli/start.go` headless preflight: require bundled shell; error if missing with fetch script hint; if `CHROME_PATH` set, log stderr that headless ignores it for binary selection (**FR-007**, **SC-005**).

- [x] T018 [P] [US2] Add integration test in `tests/integration/start_launch_failure_test.go` (or extend): missing `bin/chrome-headless-shell` → start headless fails, no instance ID (**SC-004**).

**Checkpoint**: US2 tests green; headed start still uses full Chrome path.

---

## Phase 5: User Story 4 - Ephemeral browser profile (Priority: P1) *(spec US4; after US2 for working headless starts)*

**Goal**: Verify and document profile teardown (**FR-012**, **FR-013**, **SC-006**, **SC-007**).

**Independent Test**: T009, T032 pass; 10-cycle test pass.

### Implementation for User Story 4

- [x] T019 [US4] Audit `internal/worker/run.go`, `internal/cli/stop.go`, `internal/cli/start.go` rollback, and `internal/registry/reconcile.go` for `browser.RemoveProfile` on all paths in [contracts/ephemeral-profile.md](./contracts/ephemeral-profile.md); add missing calls if any (**FR-013**).

- [x] T020 [P] [US4] Extend `tests/integration/profile_ephemeral_test.go`: **N=10** consecutive headless start/stop; assert zero leftover profile dirs when status empty (**SC-007**).

**Checkpoint**: US4 integration tests green.

---

## Phase 6: User Story 3 - Games and stand-ins keep working (Priority: P1) *(spec US3)*

**Goal**: Stand-ins on headless shell + phase-4 profile (**FR-009**, **SC-003**).

**Independent Test**: `go test ./tests/integration/ -run 'OneWin|Aviator'`.

### Implementation for User Story 3

- [x] T021 [US3] Audit `headlessUnsetBackgroundSuppression` and related unset slices in `internal/browser/chromedp.go` after T012 — confirm phase-4 did not enable `disable-background-networking=true` (**FR-003**).

- [x] T022 [P] [US3] Run `go test ./tests/integration/ -count=1 -timeout 30m -run 'OneWin|Aviator'`; set **standins_pass** in `specs/017-optimizing-browsers-4/research.md` Baseline Record (017); if fail, update matrix rejects and drop failing adopts (**SC-003**, research R7).

- [x] T023 [P] [US3] Run `go test ./tests/integration/ -count=1 -run 'HeadlessViewport|Screenshot|StartLaunch|five_headless|headless_memory'`; fix 017 regressions only (**SC-004**).

**Checkpoint**: US3 recorded; merge gate = stand-ins pass or adopts rolled back in matrix.

---

## Phase 7: User Story 5 - Transparent RAM outcome (Priority: P2)

**Goal**: Baseline + operator docs (**FR-008**, **SC-001**, **SC-005**).

**Independent Test**: Baseline Record (017) filled per [quickstart.md](./quickstart.md) §2.

### Implementation for User Story 5

- [x] T024 [US5] Measure idle `about:blank` default `start --headless` (150 samples, 30s warmup); fill **median_017_mib** and **ram_improvement_pct** in `specs/017-optimizing-browsers-4/research.md` Baseline Record (017); document if &lt;10% vs 015 (**SC-001**).

- [x] T025 [P] [US5] Optional control: `start --headless --no-lightweight`; record median with `performance_flags=none` in same Baseline Record (**FR-008**).

- [x] T026 [US5] Update `Long` on `newStartCmd()` in `internal/cli/start.go`: headless shell, `CHROME_PATH` ignored on headless, ephemeral profiles, phase-4 tuning summary (**SC-005**, **FR-010**).

- [x] T027 [P] [US5] Add `## SC-001 outcome` section to `specs/017-optimizing-browsers-4/quickstart.md` after measurement (015-style).

**Checkpoint**: US5 docs and baseline complete.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Full suite and handoff.

- [x] T028 Run `go test ./... -count=1` after **T003** (bundled shell present); fix failures attributable to 017 (**SC-004**).

- [x] T029 [P] Execute [quickstart.md](./quickstart.md) §1–§7; set **sc006_sc007_pass** in Baseline Record (017).

- [x] T030 Update final **phase4_adopted_flags** list in `specs/017-optimizing-browsers-4/research.md` Baseline Record (017) and cross-check argv vs `lightweight-profile-flags.txt`.

- [x] T031 Run `/speckit-converge` after `/speckit-implement` completes.

---

## Dependencies & Execution Order

### Phase Dependencies

| Phase | Depends on |
|-------|----------------|
| 1 Setup | — |
| 2 Foundational | 1; **T003** blocks shell-dependent tests on other machines |
| 3 US1 | 2 |
| 4 US2 | 2 (T015–T017 after T002); may parallel US1 after T010 if coordinated on `chromedp.go` |
| 5 US4 | 2 (T009); best after US2 headless starts work |
| 6 US3 | 3 + 4 |
| 7 US5 | 3 + 4 + 6 recommended |
| 8 Polish | desired stories complete |

### User Story Dependencies

| Story | Notes |
|-------|--------|
| US1 | 015 baseline + phase-4; blocks meaningful stand-ins |
| US2 | Independent of flags but shares `chromedp.go` — serialize edits |
| US4 | Mostly test/audit; after headless start works |
| US3 | Needs US1 + US2 |
| US5 | After US3 optional but needs shipped profile |

### Parallel Examples

```bash
# Foundational (different files):
# T002 script, T004 baseline copy, T005–T007 unit tests, T008–T009 integration stubs

# After US1+US2:
# T022 stand-ins parallel T023 non-game integration (same package — prefer sequential if flake)
```

### Suggested MVP scope

**MVP** = Phase 1–4 (T001–T018): **015** profile + phase-4 adopts + headless shell + allocator tests.

**Merge-ready** = MVP + US4 (T019–T020) + US3 (T022) + T028.

---

## Task Summary

| Phase | Tasks | Story |
|-------|-------|-------|
| 1 Setup | T001 | — |
| 2 Foundational | T002–T009, T032–T033 | — |
| 3 US1 | T010–T014 | US1 |
| 4 US2 | T015–T018 | US2 |
| 5 US4 | T019–T020 | US4 |
| 6 US3 | T021–T023 | US3 |
| 7 US5 | T024–T027 | US5 |
| 8 Polish | T028–T031 | — |

**Total**: **33 tasks** — US1: 5, US2: 4, US3: 3, US4: 2, US5: 4; setup/foundational/polish: 15.

**Format validation**: All tasks use `- [x]`, sequential `T001`–`T033`, story labels on US phases, explicit file paths.

---

## Notes

- Do **not** use `CHROME_PATH` for headless full Chrome (**FR-007**).
- Do **not** ship `disable-background-networking=true` on headless (**FR-003**).
- Remove **016** experimental slice before declaring US1 complete (**FR-011**).
- Stand-ins (**T022**) gate default adopts per constitution **V**.
- Run `scripts/fetch-chrome-headless-shell.sh` before headless tests (**T003**, **T033**).

## Phase 9: Convergence

- [x] T034 Create `headless-chrome-flags.txt` at repo root with **SPEC 017** section documenting chrome-headless-shell default, 015+phase-4 argv, and `--no-lightweight` policy; cross-check with `internal/browser/chromedp.go` per **T014**, **SC-005** (`partial`)
- [x] T035 Measure `start --headless --no-lightweight` idle median and record **median_017_no_lightweight_control_mib** (`performance_flags=none`) in `specs/017-optimizing-browsers-4/research.md` Baseline Record (017) per **FR-008**, **T025** (`partial`)
- [x] T036 Add `TestStartHeadlessFailsWithoutBundledShell` in `tests/integration/start_launch_failure_test.go` (temporarily rename `bin/chrome-headless-shell`); assert headless start fails with no instance ID per [contracts/headless-shell-binary.md](./contracts/headless-shell-binary.md), **T018** (`partial`)
- [x] T037 Extend `TestNoLightweightOmitsExperimentalAndLightweightTiers` in `internal/browser/chromedp_test.go` to include phase-4 keys (`disable-gpu-compositing`, etc.) and rename comment from 016 to 017 per **FR-002** (`partial`)