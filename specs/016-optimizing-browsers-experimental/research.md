# Research: Aggressive RAM Experiment (016)

**Date**: 2026-11-24  
**Spec**: [spec.md](./spec.md)  
**Before profile**: [015 research](../015-optimizing-browsers-3/research.md) Baseline Record (015) — **median_015_mib** ~1126.2 (reference) / shipped 015 path

## R1: Merge 013/014/015 + experiment (FR-003)

**Decision**: **Retain** 013, 014, and 015 lightweight slices on default headless. **Append** `headlessExperimentalLightweightOptions()` **last** when `headless && lightweight`.

**Rationale**: Spec requires experiment set **with** merge, not a wholesale replacement of prior tiers.

**Override rules**:

| Conflict | Resolution |
|----------|------------|
| `enable-webgl` + `use-angle=swiftshader` (headless base) vs experiment GPU/WebGL/3D disables | Do **not** pass `enable-webgl` / `use-angle` when `headless && lightweight` (**016** path only). |
| Phase-2 `js-flags=--max-old-space-size=320` vs experiment **128** bundle | Emit **one** `js-flags`: experimental value only (skip phase-2 `js-flags` when experimental slice is applied). |
| `disable-software-rasterizer` | Allowed on experiment path (explicit in spec Input); was forbidden on **015** default — **016** supersedes for default headless only. |
| `disable-background-networking` | **Not** in experiment; headless unset policy unchanged. |

**Alternatives considered**: Replace 013–015 with experiment-only (rejected: spec merge language); stack experiment on top while keeping WebGL enable (rejected: contradicts Input).

## R2: `--no-lightweight` performance-neutral profile (FR-001, FR-002)

**Decision**: When `!lightweight`, allocator appends **none** of: `headlessLightweightOptions`, phase 2, phase 3, experimental. Headless still applies: viewport, user-agent, unset blocks (bug **-4**), casino `disable-features` (without phase-2 extra features merge when `!lightweight` — existing **013** behavior).

**Rationale**: Operator requested no RAM/performance flags on `--no-lightweight`; use as measurement control.

**Alternatives**: Aggressive on `--no-lightweight` (rejected in spec edits).

## R3: Chrome smoke (six switches)

**Decision**: Smoke on Google Chrome 157.0.8103.78. The full set including `--single-process` makes the browser process **segmentation fault** during chromedp `Log.enable` (`browser process exited: signal: segmentation fault`). Without `--single-process`, default headless reaches `running`. **Shipped argv omits `--single-process`** so Supervisor can open pages. The other five experiment switches stay on default headless lightweight.

**Rationale**: Shipping the segfaulting switch makes every headless start fail (stand-ins and blank pages). Research R3 allowed recording a switch the reference Chrome cannot run; stand-ins cannot be authoritative if the browser never starts.

**Alternatives considered**: Keep `--single-process` and fail all headless tests (rejected: product cannot start).

## R4: Baseline procedure (FR-006, SC-003)

**Decision**: Same as 013–015: idle headless `about:blank`, 30s warmup, 2s × 150 samples, `status` MEMORY median.

| Arm | Command |
|-----|---------|
| Before (015) | Pre-016 binary or documented **median_015_mib** |
| After (016 default) | `supervisor start --headless` |
| Control (optional) | `supervisor start --headless --no-lightweight` — label `performance_flags=none` |

**Success bar**: **≥10%** lower median than 015 on same host (spec US2); document shortfall if not met.

## R5: Stand-ins and rollback (FR-007, SC-004)

**Decision**: Run `go test ./tests/integration/ -run 'OneWin|Aviator'` on default headless **after** 016 lands. If failures attribute to experiment flags, document **rollback plan**: revert experimental slice and restore headless `enable-webgl` + **015** profile on default path (follow-up task or hotfix branch).

**US2 AC2 follow-up (2026-11-24)**: **median_016_default_mib** is not at least 10% below **median_015_mib** (about 2.9% higher). Stand-ins passed, so this feature does **not** revert the allocator. Planned follow-up (separate spec, not executed here): remove `headlessExperimentalLightweightOptions` from the default headless path and restore the **015** profile (`enable-webgl`, `use-angle=swiftshader`, phase-2 `js-flags` at 320, no GPU/WebGL/3D/software-rasterizer disables). Until that follow-up ships, the tree keeps the five switches Chrome 157 can run, with `--single-process` omitted per R3.

**Rationale**: Spec US2 and US3; Constitution principle **V** conditional in plan.

**Alternatives**: Automatic revert inside 016 when RAM misses the bar (rejected: stand-ins pass and the feature’s job is to record the experiment). Ship with no follow-up plan (rejected: US2 AC2 requires a planned 015-only rollback when RAM does not improve).

## R6: CI vs baseline

**Decision**: Allocator unit tests (`TestHeadlessExperimental*`, `TestNoLightweightNeutral`) in default `go test`; RAM percent and stand-in pass/fail in **Baseline Record (016)**, not CI hard gate unless team adds later.

## Baseline Record (016)

*(Fill during implementation.)*

| Field | Value |
|-------|-------|
| host | WSL2 (same reference host as 015, 2026-11-24) |
| chrome_binary | Google Chrome 157.0.8103.78 (`/usr/bin/google-chrome`) |
| procedure | R4 |
| median_015_mib | 1126.2 (imported from [015 Baseline](../015-optimizing-browsers-3/research.md)) |
| median_016_default_mib | 1158.4 (150 samples × 2s after 30s, instance `bngjpzehytd7q`; min 989.5, max 1162.0) |
| median_016_no_lightweight_control_mib | 1123.9 (`performance_flags=none`; 150 samples × 2s after 30s, instance `5bba6xomdu5l4`; min 963.7, max 1146.4) |
| ram_improvement_pct | about **-2.9%** (higher than 015, not ≥10% lower) |
| standins_pass | yes (`go test ./tests/integration` passed 2026-11-24, including OneWin/Aviator, after omitting `single-process`) |
| rollback_recommended | no for the five shipped switches (stand-ins pass). `--single-process` must stay omitted on this Chrome (segfault). |
| experiment_argv_notes | one `js-flags` (128 bundle); no `enable-webgl` / `use-angle` on default headless lightweight; no `single-process` |