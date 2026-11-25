# Feature Specification: Aggressive RAM Experiment (Default Headless)

**Feature Branch**: `016-optimizing-browsers-experimental`

**Created**: 2026-11-24

**Status**: Draft

**Input**: User description: "Previous specification rejected some flags due concerns about game's performance. In this spec we're going to implement them as a experiment to see if they work as expected: to reduce the RAM usage." (slug: optimizing-browsers-experimental)

**Experiment flag set** (default `start --headless` only; **not** on `--no-lightweight`):

- `--disable-gpu`
- `--disable-webgl`
- `--disable-3d-apis`
- `--disable-software-rasterizer`
- `--single-process` — **omit** when the reference Chrome cannot reach `running` with it (recorded in research; Google Chrome 157 exits during startup). The other five switches stay required.
- `--js-flags=--max-old-space-size=128 --expose-gc --optimize-for-size --flush-bytecode --gc-global`

Flags from earlier specs that are **not** part of this experiment include `disable-background-networking` (headless unset policy unchanged).

**Operator constraints**:

- No new experimental CLI flag.
- **`--no-lightweight`** MUST NOT apply any Chrome switches whose purpose is RAM or performance tuning (including the experiment flag set).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Default headless tries rejected RAM flags (Priority: P1)

An operator on Linux or WSL2 wants to test **previously rejected** Chrome launch choices from spec **015** on the usual `supervisor start --headless` path (lightweight default on). They compare MEMORY in `status` against a recorded **pre-016 / 015** baseline on the same host and accept possible game-compatibility risk on that default path.

**Why this priority**: No extra flag; the experiment answers whether aggressive switches help RAM on the path operators already use for headless.

**Independent Test**: Start default headless before and after the feature; confirm post-change default includes the experiment flag set and that `--no-lightweight` does not.

**Acceptance Scenarios**:

1. **Given** `supervisor start --headless` with **no** `--no-lightweight`, **When** the instance is `running`, **Then** the **experiment flag set** above is present on the Chrome command line, except `--single-process` when research records that this Chrome cannot start with it (with plan-defined override order for conflicts with **015**, such as `enable-webgl` vs `disable-webgl`).
2. **Given** `supervisor start --headless --no-lightweight`, **When** the instance is `running`, **Then** **no** 013/014/015 lightweight RAM tiers and **no** experiment flag set and **no** other RAM/performance optimization flags are applied—only headless baseline, casino/iframe tuning, and chromedp defaults documented at plan time.
3. **Given** a **headed** start with `--no-lightweight`, **When** the instance is `running`, **Then** behavior matches **013** (no headed-safe lightweight extras; no RAM/performance optimization flags from this spec).
4. **Given** any instance, **When** `status` runs, **Then** lightweight on/off uses **existing** registry semantics only (no new experimental markers).

---

### User Story 2 - Measure whether RAM actually drops (Priority: P1)

The operator records whether the new default headless profile reduces idle aggregated MEMORY (MiB) compared to the **015** median on the same host, using the same sampling procedure as specs **013–015**.

**Why this priority**: The experiment exists to answer whether rejected flags help RAM.

**Independent Test**: Run idle measurement on default headless after change vs documented **median_015_mib**; optionally measure headless `--no-lightweight` as a **non-tuned control** (not expected to be lower RAM from optimization flags).

**Acceptance Scenarios**:

1. **Given** default headless on a reference host after ship, **When** idle measurement completes, **Then** **median_016_default_mib** and comparison to **median_015_mib** are recorded.
2. **Given** default median is **not** lower than 015 by at least **10%**, **When** the feature closes, **Then** documentation records the outcome; rollback to **015-only** default profile is planned as a follow-up task if stand-ins fail or RAM does not improve.

---

### User Story 3 - Game and stand-in compatibility is measured, not assumed (Priority: P1)

Stand-in flows (for example OneWin Aviator) run on **default** `start --headless` after the experiment flag set ships on default headless. Failures are recorded; games are not claimed supported on the new default unless stand-ins pass.

**Why this priority**: **015** rejected these flags for compatibility; changing the default is high risk.

**Independent Test**: Stand-in suite on default headless post-change vs **015** reference pass rate.

**Acceptance Scenarios**:

1. **Given** default headless after change, **When** stand-in tests run, **Then** pass/fail is recorded.
2. **Given** stand-ins fail attributable to aggressive flags, **When** the feature closes, **Then** help and plan artifacts describe rollback to **015** profile on default headless (follow-up implementation).
3. **Given** `start --headless --no-lightweight`, **When** stand-ins run, **Then** results are optional control data; **`--no-lightweight` is not required to pass** for feature completeness.

---

### User Story 4 - `--no-lightweight` remains a performance-neutral opt-out (Priority: P1)

Operators use `--no-lightweight` to skip **all** RAM/performance Chrome tuning and run a minimal profile—for example to debug Spribe issues or compare against a tuned default.

**Why this priority**: Mixing aggressive RAM flags with `--no-lightweight` contradicts the flag’s meaning (“skip lightweight extras”) and confuses baselines.

**Independent Test**: Allocator regression tests show headless and headed `--no-lightweight` omit 013/014/015 **and** the experiment flag set and any other performance-only switches defined in plan.

**Acceptance Scenarios**:

1. **Given** headless `--no-lightweight`, **When** Chrome args are inspected, **Then** none of the experiment flag set switches from **FR-003** appear, and none of the 013/014/015 performance switches appear.
2. **Given** help for `start`, **When** an operator reads it, **Then** `--no-lightweight` is described as skipping **all** lightweight and RAM-tuning flags, not enabling aggressive RAM.

---

### Edge Cases

- Aggressive default may **disable WebGL** while **015** enabled it: plan defines override order; stand-ins are the gate.
- Experiment disables WebGL/GPU while **015** enabled WebGL + SwiftShader: stand-ins are the gate; plan documents final argv.
- **`single-process`** instability: omitted on default headless when the reference Chrome cannot start with it; **`--no-lightweight` stays without it**.
- Experiment does **not** add `disable-background-networking`; Spribe/iframe unset policy from prior specs remains on all headless paths.
- Operator runs default headless and headless `--no-lightweight` side by side: distinct MEMORY; no-lightweight is not a “faster” tuning arm.
- Chrome rejects a switch: baseline records effective argv.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: **`supervisor start --headless --no-lightweight`** MUST NOT apply **any** Chrome flag whose documented purpose is RAM reduction or performance tuning, including: the **013/014/015** lightweight tiers, the **experiment flag set** (**FR-003**), and any future performance-only slices unless a later spec explicitly amends this rule.
- **FR-002**: **`supervisor start` with `--no-lightweight` (headed or headless)** MUST behave as a **performance-neutral** opt-out: minimal applicable chromedp/supervisor flags only, with headless Spribe-safe unset overrides unchanged from pre-016 behavior.
- **FR-003**: **`supervisor start --headless`** without `--no-lightweight` (default lightweight on) MUST apply the experiment flag set from the feature Input: `disable-gpu`, `disable-webgl`, `disable-3d-apis`, `disable-software-rasterizer`, one `js-flags` value `--max-old-space-size=128 --expose-gc --optimize-for-size --flush-bytecode --gc-global`, and `single-process` **unless** research records that the reference Chrome cannot reach `running` with `single-process` (then that switch is omitted and the other five remain). Plan MUST define how these merge with remaining **013/014/015** lightweight switches without duplicate keys and how conflicting switches (for example prior `enable-webgl` / `use-angle=swiftshader` or phase-2 `js-flags=320`) are resolved so the experiment values win.
- **FR-004**: Supervisor MUST NOT add a new CLI flag, environment toggle, or registry field for “experimental mode.”
- **FR-005**: `status`, NOTE, `--memory-limit`, and `--threshold` semantics from **013** remain; only Chrome launch tuning changes on default headless per **FR-003**.
- **FR-006**: Record **median_015_mib** (pre-016 reference) and **median_016_default_mib** (post-change default headless) using the **013–015** idle procedure; optionally record headless `--no-lightweight` as **control** with label `performance_flags=none`.
- **FR-007**: Stand-in suite on **default** `start --headless` before feature close.
- **FR-008**: `start --help` MUST state: default headless uses aggressive RAM switches (compatibility risk); **`--no-lightweight` skips all RAM/performance tuning**; `--memory-limit` is NOTE-only.

### Key Entities

- **015 baseline profile**: Pre-016 default headless lightweight; RAM comparison reference.
- **Experiment flag set**: The Input switches on **default headless only** (`single-process` omitted when research records a startup failure); excludes `disable-background-networking`.
- **No-lightweight profile**: Performance-neutral; `lightweight=false`; no RAM tuning flags.
- **Baseline record (016)**: `median_015_mib`, `median_016_default_mib`, optional control median for `--no-lightweight`, stand-in results.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Headless `--no-lightweight` argv contains **zero** switches from the experiment flag set and **zero** 013/014/015 performance switches.
- **SC-002**: Default `start --headless` includes the experiment flag set from Input without a new CLI flag, omitting `single-process` only when research records that the reference Chrome cannot start with it.
- **SC-003**: Baseline artifact compares 015 vs post-change default idle medians.
- **SC-004**: Stand-in results on default headless are recorded.
- **SC-005**: Default integration tests for **non-game** headless flows pass, or failures are documented with rollback plan.
- **SC-006**: Help text matches **FR-001** and **FR-008**.

## Assumptions

- **015** shipped as-is is the “before” reference until **016** lands.
- Default headless may regress games; rollback spec/task is acceptable if stand-ins fail.
- Headed starts are unchanged except **`--no-lightweight`** clarity in help (**FR-002**).
- Hard RSS caps remain out of scope.

## Dependencies

- **Spec 015**: Reject matrix and before baseline.
- **Specs 013–014**: Defines what counts as “lightweight RAM tiers” excluded from `--no-lightweight`.
- **Bug try-limit-of-app-config-exceeded-4**: Experiment does not add `disable-background-networking`; headless unset policy unchanged.
- **Constitution**: CLI surface; default tests without live casino.