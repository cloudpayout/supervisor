# Feature Specification: Further Chrome RAM Reduction (Phase 2)

**Feature Branch**: `014-optimizing-browsers-2`

**Created**: 2026-11-23

**Status**: Draft

**Input**: User description: "After specs/013-optimizing-browsers spec was implemented, still the resources taken by the instances are too much. We should add more flags to Chrome to reduce the RAM usage as much as we can." (slug: optimizing-browsers-2)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Lower idle RAM beyond spec 013 (Priority: P1)

An operator runs headless supervised instances on Linux or WSL2. After spec **013**, default `start --headless` already applies a lightweight Chrome profile, but a single idle instance on `about:blank` can still report on the order of **1 GiB** aggregated MEMORY in `status`. They need a **second wave** of Chrome launch choices that further reduces typical idle and steady-state **memory (MiB)** without turning off monitoring or breaking supported automations.

**Why this priority**: Host cost and fleet density remain the top operator pain; 013 was necessary but insufficient for their hardware budget.

**Independent Test**: On a reference host, compare median idle MEMORY (MiB) for one headless instance using the **013** profile versus the **014** profile using the same measurement procedure as spec 013. Default integration start/status/stop with `--headless` still pass without a live casino site.

**Acceptance Scenarios**:

1. **Given** `supervisor start --headless` with no opt-out flags, **When** the instance is `running` on `about:blank` after five minutes idle, **Then** aggregated MEMORY in `status` is materially lower than the documented **013 post-change** median on the same host (Success Criteria).
2. **Given** an instance started with the 014 profile, **When** the operator leaves it idle for five minutes, **Then** memory does not show runaway growth attributable solely to the new launch choices.
3. **Given** default integration tests for headless start, status, and stop, **When** the suite runs without a live casino site, **Then** all such tests pass with no new failures attributable to phase-2 flags.
4. **Given** `supervisor start --headless --no-lightweight`, **When** the instance is `running`, **Then** **neither** 013 **nor** 014 lightweight extras are applied (behavior unchanged from 013 for opt-out).

---

### User Story 2 - Automations stay reliable (Priority: P1)

The operator still runs iframe-heavy casino automations (for example OneWin Aviator stand-in paths). Phase-2 flags MUST NOT trade away capabilities required for supported games. Any candidate flag that conflicts with canvas, WebGL, fonts, or Spribe background behavior remains **off**, with the conflict recorded in the feature plan.

**Why this priority**: Constitution and 013 guardrails still apply; additional savings are worthless if entry or readiness regress.

**Independent Test**: Re-run documented stand-in and integration scenarios that passed after 013 (headless viewport, screenshot, OneWin/Aviator stand-in suite). No new failures solely due to phase-2 tuning.

**Acceptance Scenarios**:

1. **Given** headless instance at the fixed layout size from spec 012, **When** screenshot is requested, **Then** capture succeeds with the same dimensional semantics as after 013.
2. **Given** stand-in paths that validated game entry and iframe readiness after 013, **When** they run with the 014 profile, **Then** outcomes match pre-014 pass/fail (no new failures attributable only to RAM flags).
3. **Given** a proposed flag would disable a capability required for a supported game, **When** Supervisor applies the default 014 headless profile, **Then** that flag is **not** present (compatibility overrides maximum savings).

---

### User Story 3 - Operator visibility unchanged (Priority: P2)

The operator continues to use `status`, optional `--memory-limit` / `--threshold`, NOTE, and MiB display exactly as defined in **013**. Phase-2 changes affect **how much** RAM Chrome uses, not **how limits are enforced** (soft warnings only in v1).

**Why this priority**: Avoid scope creep into hard enforcement or new CLI surface unless needed.

**Independent Test**: Start with `--memory-limit 600MiB`; observe MEMORY `current/limit` and NOTE when over threshold; instance remains `running`.

**Acceptance Scenarios**:

1. **Given** `start --headless --memory-limit 512MiB`, **When** aggregated memory exceeds the warning line, **Then** NOTE shows the threshold message and the instance stays `running` (same as 013).
2. **Given** any instance started with 014 profile, **When** `status` is run, **Then** MEMORY is still expressed in MiB with the same column layout as 013 (including NOTE).

---

### User Story 4 - Documented adopt/reject matrix (Priority: P2)

An operator or reviewer needs to see **which additional Chrome flags** were considered, which were adopted in the default headless profile, and which remain rejected (with reasons), so expectations about “600 MiB limit” vs “actual RSS” stay clear: limits **warn**; flags **reduce** typical use where safe.

**Why this priority**: Reduces confusion between `--memory-limit` policy and launch-time optimization.

**Independent Test**: Feature plan or research lists every new flag with adopt/reject and links to idle RAM measurement results.

**Acceptance Scenarios**:

1. **Given** the feature ships, **When** an operator reads operator-facing docs for this change, **Then** they can distinguish **status memory warnings** from **Chrome launch optimization** and see that decimal MB/GB are not used for limits.
2. **Given** a flag is rejected for game compatibility, **When** the operator inspects the plan artifact, **Then** the rejection reason is stated (for example WebGL, SwiftShader, Spribe background fetches).

---

### Edge Cases

- Operator expected `--memory-limit` to **cap** Chrome near 600 MiB: v1 still only **warns** in `status`; phase-2 flags **lower typical RSS** but do not guarantee staying under an arbitrary limit (document clearly).
- Chrome version on host differs from reference: relative improvement vs 013 matters more than absolute MiB.
- `start --no-lightweight`: no 014 flags; measurement comparisons for SC-001 use **default** headless **without** `--no-lightweight`.
- Mix of instances: some started before upgrade (013 only in registry) until stopped; new starts get 014 profile.
- Flag combination causes start failure: `start` fails clearly with no registry row (same as 013).
- Long-running pages with heavy media: idle blank-page wins may not apply; stand-in steady-state probes included in plan.
- WSL2 vs bare Linux: baseline recorded per host; CI does not gate absolute MiB.
- Rejected candidates from 013 (for example `disable-gpu`, `disable-software-rasterizer`) remain out of scope for adoption unless a future spec amends guardrails with measured proof.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST extend the **default headless lightweight launch profile** (spec 013 full tier) with **phase-2 Chrome launch choices** chosen to reduce aggregated idle memory (MiB) as far as compatibility guardrails allow. Headed and `--no-lightweight` behavior MUST remain as in 013 (no additional 014 flags on those paths).
- **FR-002**: Phase-2 choices MUST be documented in a feature **adopt/reject matrix** (plan/research). Each adopted flag MUST map to an expected memory or process-behavior rationale; each rejected flag MUST cite a guardrail (game entry, WebGL, canvas, fonts, background networking, or prior 013 rejection).
- **FR-003**: Phase-2 MUST NOT remove capabilities required for explicitly supported casino automations and documented stand-in entry paths; where a saving conflicts with a required capability, the capability MUST be kept.
- **FR-004**: Per-instance isolation, `status` MEMORY/NOTE semantics, `--memory-limit`, `--threshold`, `--no-lightweight`, and registry fields from **013** MUST remain satisfied unless this spec explicitly states otherwise (this spec does **not** add hard memory enforcement or automatic stop/restart).
- **FR-005**: Default automated tests MUST run without a live casino site and MUST include regression checks that phase-2 flags are present on default headless `start` and absent when `--no-lightweight`; existing 013 integration coverage for memory flags, NOTE, and lightweight tiers MUST NOT regress.
- **FR-006**: The feature MUST define a **013-vs-014 baseline procedure** (same idle headless `about:blank` method as 013) and record median idle MiB for both profiles on a reference host for audit (manual gate acceptable; same policy as SC-001 in 013).
- **FR-007**: Operator-facing help or release notes MUST clarify that `--memory-limit` sets a **warning policy** in `status`, not a hard cap on Chrome RSS, and that phase-2 flags aim to reduce typical usage.
- **FR-008**: If a phase-2 flag causes `start` failure on a supported environment, Supervisor MUST fail `start` with a clear message and MUST NOT leave a registry row for that attempt.

### Key Entities

- **013 lightweight profile**: Default headless bundle from spec 013 (baseline for comparison).
- **Phase-2 profile**: Additional launch choices appended to 013 headless full tier only.
- **Compatibility guardrail**: Same concept as 013; phase-2 MUST not violate existing guardrails without spec amendment.
- **Baseline record (014)**: 013 median and 014 median idle MiB on the same host, same procedure.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On the documented procedure, median idle MEMORY (MiB) per headless instance on `about:blank` with the **014** default profile is at least **15% lower** than the median for the **013** profile on the **same reference host** (same sample window). Recorded in the feature baseline artifact; not required to gate default CI unless a documented reference job is added later.
- **SC-002**: 100% of default integration tests for headless start, status, stop, screenshot, and registry lifecycle pass without a live casino site (constitution).
- **SC-003**: 100% of documented stand-in or integration scenarios that passed after **013** still pass with the **014** profile, with no new failures attributable solely to phase-2 flags.
- **SC-004**: Operator documentation lists phase-2 adopted and rejected flags and explains the difference between MEMORY limits (warnings) and launch optimization (typical RSS).
- **SC-005**: An operator can still run at least as many concurrent headless blank instances on the reference host as with 013 (no artificial reduction in achievable concurrent count attributable to phase-2 flags).

## Assumptions

- Spec **013** is implemented on the branch before **014** work begins; comparison baselines use 013 as the “before” profile for RAM measurements.
- Primary target is **headless** fleet automation; headed remains 013 headed-safe tier only.
- Chrome/Chromium is provided on the host; absolute MiB varies by binary version.
- Memory figures remain **mebibytes (MiB)** end-to-end per 013.
- Some operator-requested flags from the original RAM backlog remain **permanently rejected** unless this feature’s research proves safe stand-in results (for example GPU/software rasterizer disables with SwiftShader WebGL).
- Hard cgroups, OOM kill, or automatic restart when over `--memory-limit` are **out of scope** (deferred features).
- A follow-on may tighten V8 heap or renderer limits further if SC-001 is not met with “safe” flags alone.

## Dependencies

- **Spec 013**: Lightweight tiers, guardrails, `status` NOTE/MiB, `--memory-limit`, `--no-lightweight`.
- **Spec 012**: Headless viewport dimensions unchanged.
- **Constitution**: CLI-only operator surface; chromedp; default tests without live casino.