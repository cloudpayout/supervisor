# Feature Specification: Aggressive Chrome RAM Reduction (Phase 3)

**Feature Branch**: `015-optimizing-browsers-3`

**Created**: 2026-11-23

**Status**: Draft

**Input**: User description: "Still the RAM usage is too high. Let's add the following flags to the lightweight profile. They may work: `--disable-gpu`, `--disable-webgl`, `--disable-3d-apis`, `--js-flags=--max-old-space-size=128 --expose-gc --optimize-for-size --flush-bytecode --gc-global`, `--process-per-site`, `--disable-application-cache`, `--v8-cache-options=none`, `--single-process`, `--memory-pressure-off`, `--disable-background-networking`" (slug: optimizing-browsers-3)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Evaluate and adopt safe RAM flags (Priority: P1)

An operator runs headless supervised instances on Linux or WSL2. After specs **013** and **014**, default `start --headless` still reports on the order of **1 GiB** aggregated idle MEMORY on `about:blank`. They propose a **third wave** of Chrome launch choices (the operator list above) to cut RAM further.

**Why this priority**: 014 did not meet the idle RAM improvement target on the reference host; the operator wants to try more aggressive options.

**Independent Test**: Document an adopt/reject matrix for **each** proposed flag; apply only adopted flags to the default headless lightweight profile; compare median idle MEMORY (MiB) to the documented **014** baseline using the same measurement procedure.

**Acceptance Scenarios**:

1. **Given** the operator flag list, **When** the feature is planned, **Then** every listed flag has a recorded **adopt**, **reject**, or **already present** decision with a testable reason.
2. **Given** `supervisor start --headless` with no opt-out, **When** the instance is `running`, **Then** only flags marked **adopt** appear on the default headless lightweight profile (no duplicate switches for flags already in 013/014).
3. **Given** a proposed flag conflicts with supported casino automation or established compatibility policy (for example disabling capabilities required for canvas/WebGL/iframe games, or re-enabling background networking suppression that breaks Spribe bootstrap), **When** the default profile is built, **Then** that flag is **reject**ed and **not** present on default `start --headless`.
4. **Given** `supervisor start --headless --no-lightweight`, **When** the instance is `running`, **Then** **no** phase-3 lightweight extras are applied (same as 013/014 for opt-out).

---

### User Story 2 - Automations must not regress (Priority: P1)

Supported casino stand-in and integration flows (for example OneWin Aviator paths) must keep passing when only **adopt**ed phase-3 flags are enabled.

**Why this priority**: Aggressive GPU/WebGL/process flags can silently break games; savings are worthless if entry fails.

**Independent Test**: Re-run documented stand-in and headless integration suites after each adopted flag group; fail the feature if a regression is attributable solely to phase-3 tuning.

**Acceptance Scenarios**:

1. **Given** a flag would set **disable-webgl** or **disable-3d-apis** while games require canvas/WebGL, **When** evaluating the operator list, **Then** that flag is **reject**ed unless stand-in proof shows equivalent game readiness (default: **reject**).
2. **Given** a flag would enable **disable-background-networking** on headless, **When** evaluating against Spribe/iframe compatibility policy, **Then** that flag is **reject**ed on the default headless profile.
3. **Given** phase-3 adoptions are merged, **When** stand-in tests run, **Then** pass/fail matches the pre-phase-3 baseline for failures not attributable to environment drift.

---

### User Story 3 - Operator visibility unchanged (Priority: P2)

`status`, NOTE, `--memory-limit`, `--threshold`, MiB display, and `--no-lightweight` behavior remain as defined in **013**; phase-3 only changes launch tuning on the default lightweight path.

**Independent Test**: Existing memory/status integration tests pass without change to CLI contracts.

**Acceptance Scenarios**:

1. **Given** any instance with phase-3 profile, **When** `status` runs, **Then** MEMORY and NOTE columns behave as in 013/014.
2. **Given** `start --headless --memory-limit 512MiB`, **When** usage exceeds the warning line, **Then** NOTE shows the threshold message and the instance stays `running` (soft warning only).

---

### User Story 4 - Transparent outcomes when RAM does not drop (Priority: P2)

If aggressive flags are rejected or idle RAM does not improve, the operator can read **what was tried**, **what shipped**, and **why** idle MEMORY may still read ~1 GiB.

**Why this priority**: 014 already showed idle `about:blank` RSS may not move with safe flags alone; phase-3 must not repeat confusion between flags and `--memory-limit`.

**Independent Test**: Plan/research artifacts list operator proposals, final profile deltas, and baseline comparison vs 014.

**Acceptance Scenarios**:

1. **Given** SC-001 is not met after phase-3, **When** the operator reads feature docs, **Then** they see measured medians and which proposals were rejected for compatibility.
2. **Given** a flag is **already present** in 013/014 (for example process-per-site, disable-application-cache, v8-cache-options=none), **When** documented, **Then** it is marked **skip (duplicate)** not added again.

---

### Edge Cases

- Operator list includes **contradictory** pairs (for example current product enables WebGL for games while proposals disable WebGL/GPU): compatibility policy wins; conflicting proposals are **reject**ed.
- **single-process** and aggressive V8 bundles may improve RSS briefly but destabilize iframes or Chrome on WSL2: default **reject** unless stand-ins and idle measurement both pass.
- Flags already in the lightweight profile must not be duplicated on the command line.
- Chrome version drift changes absolute MiB; improvement is judged on the **same host** vs **014** median.
- Partial adoption (only **memory-pressure-off** or tighter **js-flags** without the full operator bundle) is acceptable if the bundle fails stand-ins.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST evaluate the operator’s phase-3 flag list against the current **013 + 014** default headless lightweight profile and publish an adopt/reject/**already present** matrix before implementation.
- **FR-002**: Only flags marked **adopt** in that matrix MAY be added to the default headless lightweight launch profile; headed and `--no-lightweight` paths MUST remain as in 013/014.
- **FR-003**: Flags that conflict with supported casino automation, documented Spribe/iframe compatibility policy, or prior **reject** decisions (for example headless must not suppress background networking required for config bootstrap) MUST NOT appear on the default headless profile.
- **FR-004**: Proposals that disable GPU, WebGL, or 3D APIs MUST NOT ship on the default profile unless stand-in tests prove supported games still enter and reach readiness; absent proof, they MUST be **reject**ed.
- **FR-005**: Proposals for **single-process** or the full aggressive **js-flags** bundle (expose-gc, optimize-for-size, flush-bytecode, gc-global) MUST be **reject**ed on the default profile unless a later spec explicitly opts into experimental instability risk after measured proof.
- **FR-006**: The feature MUST record a **014-vs-phase-3** idle MEMORY baseline on a reference host (same procedure as 013/014) and set **`sc001_pass`** for whether median idle MiB improved by at least **15%** vs the documented **014** median.
- **FR-007**: Per-instance isolation, `status` semantics, registry fields, and existing CLI from **013** MUST remain satisfied; this spec does not add hard memory enforcement or automatic stop/restart.
- **FR-008**: Default automated tests MUST continue to pass without a live casino site; stand-in suites MUST be re-run after any adoption batch.

### Key Entities

- **014 baseline profile**: Current default headless lightweight bundle (comparison “before” for phase 3).
- **Operator phase-3 proposals**: The user-supplied Chrome switch list in the feature input.
- **Adopt/reject matrix**: Per-flag decision with rationale and links to stand-in/baseline results.
- **Baseline record (015)**: 014 median and phase-3 median idle MiB on the same host.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On the documented idle procedure, median aggregated MEMORY (MiB) on `about:blank` with the **phase-3** default profile is at least **15% lower** than the documented **014** median on the **same reference host**, **or** the matrix documents that all remaining operator proposals were **reject**ed for compatibility and measured idle RSS did not improve (outcome recorded in the baseline artifact).
- **SC-002**: 100% of default integration tests for headless start, status, stop, and screenshot pass without a live casino site.
- **SC-003**: 100% of documented OneWin/Aviator stand-in scenarios that passed under **014** still pass with the **adopt**ed phase-3 profile only.
- **SC-004**: Operator-facing documentation lists each proposed flag with adopt/reject/**already present** and explains that `--memory-limit` remains a warning policy, not an RSS cap.

## Assumptions

- Specs **013** and **014** are implemented; comparison baselines use **014** as the immediate “before” profile for RAM measurements.
- Several operator proposals are **already present** in 013/014 (`process-per-site`, `disable-application-cache`, `v8-cache-options=none`); phase 3 treats them as duplicates, not new work.
- Current headless product policy **enables WebGL with a software renderer path** for casino canvas games; proposals to disable GPU/WebGL/3D APIs are assumed **incompatible** until stand-ins prove otherwise.
- Headless must **not** enable Puppeteer-style **disable-background-networking** (compatibility with Spribe iframe config) unless a future spec amends that policy with measured proof.
- **`memory-pressure-off`** is net-new and may be adopted if stand-ins and baseline support it; it is the primary candidate not already in 014 among safe switches.
- Tighter V8 heap (for example **128 MiB**) may be tried only in controlled steps after stand-ins; the full operator **js-flags** bundle remains high risk.
- Hard cgroups or process-level RSS caps are **out of scope** (deferred).

## Dependencies

- **Spec 014**: Phase-2 profile and baseline record.
- **Spec 013**: Lightweight tiers, `--no-lightweight`, memory `status` semantics.
- **Spec 012**: Headless viewport dimensions.
- **Compatibility policy**: Prior bug assessment for headless networking/canvas/fonts (try-limit-of-app-config-exceeded-4).
- **Constitution**: CLI-only operator surface; default tests without live casino.