# Feature Specification: Lightweight Browser Instances

**Feature Branch**: `013-optimizing-browsers`

**Created**: 2026-11-22

**Status**: Draft

**Input**: User description: "Optimize the resources being expent by the browsers. I want to make the instances as lightweight as possible." (slug: optimizing-browsers)

## Clarifications

### Session 2026-11-22

- Q: When an instance’s aggregated memory crosses a defined limit, what should Supervisor do? → A: `status` marks instances that exceed the configured warning line (state or field); no automatic `stop` or `start` refusal in v1. Automatic restart when over limit is deferred to a future feature.
- Q: How should the v1 per-instance memory threshold be chosen? → A: By default, no memory limit is set. `supervisor start` accepts optional `--memory-limit` (maximum memory for that instance) and optional `--threshold` (percentage of that maximum at which `status` shows a warning). Limits are per instance at start time, not global defaults.
- Q: How should operators write values for `--memory-limit`? → A: Suffixed human-readable size strings (e.g. `512M`, `1G`, `512MiB`) with documented parsing rules; parsed limit is compared to the same aggregated-memory basis as the `status` MEMORY column.
- Q: For instances started with `--memory-limit`, should `status` show the configured limit and threshold on that row? → A: `status` adds a **NOTE** column (empty by default). When a memory limit is configured, the **MEMORY** column shows `{current-memory-use}/{memory-limit}`; when aggregated memory is at or above the warning line, NOTE shows a fixed threshold-exceeded message (v1, no auto-stop).
- Q: What memory numeral system should Supervisor use for limits, comparisons, and `status` display? → A: **Mebibytes (MiB)** end-to-end: binary (1024-based) units only; `status` shows memory in MiB; `--memory-limit` accepts `M`/`MiB` and `G`/`GiB` as mebibyte/gibibyte shorthands; decimal **MB**/**GB** are not accepted.

### Session 2026-11-23

- Q: How should an operator skip the lightweight headless Chrome profile on `start`? → A: **`--no-lightweight`** on `start` for **both headed and headless** runs. **Headless** without the flag uses the full lightweight profile by default; with the flag, skip the lightweight bundle (pre-013–style headless allocator for resource extras). **Headed** without the flag may receive only **headed-safe** lightweight extras defined in the feature plan; with the flag, skip **all** 013 lightweight extras. Memory limit flags are independent.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Lower cost per headless automation instance (Priority: P1)

An operator runs several headless supervised browser instances on one Linux or WSL2 host to drive casino automations. Each instance today consumes noticeable memory (MiB) even when idle on a blank page, which limits how many runs they can keep alive and increases host cost. After this change, a newly started headless instance uses a documented lightweight launch profile by default so typical idle and steady-state **memory** drop materially compared to the pre-change baseline, while supported automations still reach `running` and pass existing default integration checks.

**Why this priority**: Headless fleet density is the main driver of host spend; measurable per-instance savings directly deliver the operator goal; optional memory flags on `start` are not required to obtain the default lightweight profile.

**Independent Test**: On a reference host, record median idle memory (MiB) for one headless instance on `about:blank` before and after the change (same Chrome availability, same headless viewport policy). Start one headless instance, wait until `status` shows `running` with available memory, and verify RAM meets the success thresholds while default integration start/stop/status tests pass.

**Acceptance Scenarios**:

1. **Given** an operator runs `supervisor start --headless` with no extra resource flags, **When** the instance reaches `running` on a blank page, **Then** the aggregated memory reported for that instance in `status` is at or below the post-change targets defined in Success Criteria versus the documented baseline.
2. **Given** a headless instance is `running`, **When** the operator leaves it idle for five minutes, **Then** memory does not climb unbounded compared to baseline idle behavior (no new runaway growth attributable to the lightweight profile).
3. **Given** default integration tests for start, status, and stop with `--headless`, **When** the test suite runs without a live casino site, **Then** all such tests pass with no new failures attributable to the optimization profile.
4. **Given** `supervisor start --headless --no-lightweight`, **When** the instance is `running`, **Then** the instance uses the non-lightweight headless allocator (no 013 lightweight bundle) and still passes start/stop/status lifecycle tests.
5. **Given** two headless instances, one with default lightweight and one with `--no-lightweight`, **When** both are `running`, **Then** each retains its own profile independently (no cross-instance leakage).

---

### User Story 2 - Automations remain reliable (Priority: P1)

An operator depends on layout-sensitive, iframe-heavy casino automations (for example OneWin Aviator entry) that already required careful browser configuration. They need confidence that making instances lightweight does not silently break game entry, readiness probes, screenshots, or session lifecycle. Optimizations MUST be gated by compatibility guardrails so required capabilities for explicitly supported games remain available.

**Why this priority**: Resource savings are worthless if automations fail; constitution requires real-time monitoring while instances run supported games.

**Independent Test**: Re-run documented stand-in or integration scenarios that exercise headless viewport, screenshot capture, and casino entry paths that existed before this feature; confirm pass criteria unchanged except where this spec explicitly updates expectations.

**Acceptance Scenarios**:

1. **Given** a headless instance at the fixed headless layout size already defined for Supervisor, **When** the operator captures a screenshot, **Then** capture succeeds with the same dimensional semantics as before this feature (current view, documented width and height).
2. **Given** headless automation entry checks that succeeded on the reference stand-in path before optimization, **When** those checks run again after the lightweight profile is applied, **Then** readiness outcomes match pre-change pass/fail (no new failures solely due to resource tuning).
3. **Given** an optimization would disable a capability required for a supported game, **When** Supervisor applies the default lightweight profile, **Then** that capability remains enabled for that profile (compatibility overrides win over maximum savings).

---

### User Story 3 - Optional memory limits and status warnings (Priority: P1)

An operator uses `supervisor status` to decide when to stop idle instances or whether the host can accept another `start`. By default, instances have **no** configured memory limit: the MEMORY column shows current usage only and the NOTE column is empty. When they pass `--memory-limit` (and optionally `--threshold`) on `start`, MEMORY shows **`{current}/{limit}`** and NOTE stays empty until aggregated memory reaches the warning line, then NOTE shows a threshold-exceeded message. Alerts are informational in v1—no automatic stop, restart, or `start` refusal.

**Why this priority**: Clarified product focus is instance memory and limits; constitution mandates resource monitoring; operators tune limits per run before future automatic restart behavior.

**Independent Test**: Start one instance without memory flags (MEMORY current-only, NOTE empty). Start a second with `--memory-limit` and `--threshold`, confirm MEMORY uses `{current}/{limit}`, drive memory above the warning line, confirm NOTE shows the threshold message while the instance stays `running`. Run `status` on a mixed fleet and confirm formatting applies per row.

**Acceptance Scenarios**:

1. **Given** `supervisor start` without `--memory-limit`, **When** the instance is `running`, **Then** `status` MEMORY shows current aggregated memory only (not `current/limit`) and NOTE is empty.
2. **Given** `supervisor start --memory-limit <size>` where `<size>` is valid (for example `512MiB` or `512M`), **When** the instance is `running`, **Then** MEMORY displays `{current-memory-use}/{memory-limit}` with both sides in **MiB** (same precision rules, e.g. one decimal and `MiB` suffix), and NOTE is empty while aggregated memory is below the warning line.
3. **Given** an instance started with `--memory-limit` and `--threshold <percent>`, **When** aggregated memory is strictly below `<percent>`% of the parsed limit, **Then** NOTE is empty.
4. **Given** the same instance when aggregated memory is at or above the warning line, **When** the operator runs `status`, **Then** NOTE shows a fixed threshold-exceeded message (documented in help), MEMORY still shows `{current}/{limit}`, and the instance remains `running` until operator `stop` or crash—no auto-stop in v1.
5. **Given** two concurrent instances, only one started with `--memory-limit`, **When** the operator runs `supervisor status`, **Then** the limited row uses `current/limit` in MEMORY when running; the other row shows MEMORY without a limit suffix; NOTE is empty on the unconfigured row unless it has its own warning state (it cannot, without a limit).
6. **Given** an instance in `starting` with `--memory-limit` configured, **When** status refreshes, **Then** MEMORY is `unavailable` or shows `unavailable/{limit-in-MiB}` consistently for all rows with a limit; NOTE is empty while metrics are unavailable.
7. **Given** the operator runs `start` with `--threshold` but without `--memory-limit`, **When** `start` is invoked, **Then** Supervisor exits non-zero with a clear usage error and does not register an instance.
8. **Given** the operator runs `start` with `--memory-limit` set to an unparseable or non-positive size string (including decimal `MB`/`GB`), **When** `start` is invoked, **Then** Supervisor exits non-zero with a clear usage error citing invalid memory limit syntax and does not register an instance.
9. **Given** `512M` and `512MiB` limits, **When** both are started on separate instances, **Then** both configure the same 512 MiB maximum and MEMORY denominators match.

---

### User Story 4 - Headed debugging stays usable (Priority: P3)

An operator sometimes starts a headed instance to watch behavior. They accept modest savings where safe but MUST retain a visible window and debugging ergonomics; aggressive headless-only optimizations MUST NOT apply in ways that break headed mode.

**Why this priority**: Headed runs are fewer but critical for diagnosis; scope is headless-first per assumptions.

**Independent Test**: Start without `--headless`, confirm visible window, registry `headless` false, and no regression in headed start integration tests.

**Acceptance Scenarios**:

1. **Given** `supervisor start` without `--headless`, **When** the instance is `running`, **Then** a visible browser window is shown and lifecycle commands behave as today.
2. **Given** headed and headless instances on one host, **When** both are running, **Then** headless rows show larger relative savings than headed rows where optimizations are headless-specific.
3. **Given** `supervisor start --no-lightweight` without `--headless`, **When** the instance is `running`, **Then** a visible window is shown and **no** 013 lightweight extras are applied (legacy headed allocator behavior for this feature).
4. **Given** `supervisor start` without `--no-lightweight` and without `--headless`, **When** the instance is `running`, **Then** at most **headed-safe** lightweight extras from the feature plan apply; headless-only bundle flags MUST NOT be present.

---

### User Story 5 - Opt out of lightweight profile (Priority: P2)

An operator debugging game failures or comparing memory behavior needs to start a browser **without** the 013 lightweight extras. They use **`--no-lightweight`** on `start` for headed or headless runs; the choice is recorded for that instance until `stop`.

**Why this priority**: Reliability debugging must not require a separate binary or config file; same flag works for headed and headless.

**Independent Test**: Start headless with and without `--no-lightweight`; allocator tests (or integration checks) prove the lightweight bundle is present only when the flag is omitted on headless. Repeat headed with `--no-lightweight` vs default.

**Acceptance Scenarios**:

1. **Given** `start --help`, **When** the operator reads lightweight options, **Then** it documents `--no-lightweight`, defaults for headed vs headless, and independence from `--memory-limit`.
2. **Given** `supervisor start --headless --no-lightweight`, **When** the worker opens Chrome, **Then** the headless lightweight bundle from the feature plan is **not** applied.
3. **Given** `supervisor start --no-lightweight` (headed), **When** the instance is `running`, **Then** no 013 lightweight extras are applied.

---

### Edge Cases

- Host at capacity: `start` still fails clearly with no phantom registry row when the browser cannot launch; lightweight profile must not mask launch failures.
- Instance owner crash with Chrome still running: reconciliation and orphan cleanup behave as today; optimizations do not skip process termination.
- Mixed fleet (headed, headless, with and without session files): each instance gets the profile appropriate to its mode; stopping one does not affect others.
- Long-running automation with periodic network activity: memory may spike during heavy pages; success criteria compare idle blank-page baseline and documented steady-state probes; warning markers may appear during legitimate peaks without triggering auto-stop in v1.
- Instance above its configured warning line (`threshold`% of `--memory-limit`) for a sustained period: NOTE remains populated; operator must intervene manually in v1; automatic restart when over limit is out of scope and planned as a separate future feature.
- Instance at or above 100% of its configured `--memory-limit`: MEMORY may show current above limit (for example `600.0 MiB/512.0 MiB`); NOTE shows threshold message when warning line is crossed; no auto-stop in v1; future restart spec may treat at-cap differently.
- `start` with `--memory-limit` only (no `--threshold`): warning uses default 100% of the configured maximum (warn when aggregated memory is at or above the full limit).
- Mixed fleet: some rows with `current/limit` MEMORY, some with current-only; NOTE empty unless that row’s warning line is crossed.
- **NOTE** column present on every status row (one-shot and follow); empty string when no message applies.
- Ambiguous or unsupported `--memory-limit` suffixes (for example bare numbers with no unit when documentation requires a suffix): `start` fails with usage error; no registry row.
- Equivalent limits with different spellings (for example `512M` vs `512MiB`) MUST parse to the same configured maximum when both forms are supported.
- Screenshot or capture errors: unrelated capture failures still surface clearly; lightweight profile must not remove capture capability.
- Operator runs many parallel starts: distinct IDs and profiles per instance; no shared leakage of cookies or storage across profiles.
- Game vendor updates that need canvas, fonts, or background timers: compatibility guardrails documented in the plan must list which capabilities stay enabled even if they cost memory.
- Baseline drift across hosts: acceptance uses a documented reference measurement procedure; absolute MiB may vary by host but relative improvement thresholds apply.
- `--memory-limit 512MB` (decimal megabyte suffix): rejected at `start` with usage error; operators must use `MiB`/`M` or `GiB`/`G`.
- Fleet mix: instances with and without `--no-lightweight` on one host; each worker uses the profile stored for that instance ID.
- `--no-lightweight` with `--memory-limit`: memory policy applies; lightweight skip does not disable limits or NOTE logic.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST apply the **lightweight browser launch profile** by default on **`--headless` starts** when `--no-lightweight` is omitted. Headed starts without `--no-lightweight` MUST apply at most **headed-safe** lightweight extras from the feature plan; headless-only bundle flags MUST NOT apply to headed runs. When `--no-lightweight` is set, Supervisor MUST NOT apply any 013 lightweight extras for that instance (headed or headless). Memory flags `--memory-limit` / `--threshold` are independent of lightweight profile selection.
- **FR-002**: The lightweight profile MUST prioritize headless instances for **SC-001**: baseline and savings claims use default **`start --headless`** (lightweight **on**). Headless starts with `--no-lightweight` are out of scope for SC-001 measurement. Headed starts MUST NOT adopt headless-only restrictions that remove the visible window or break headed integration tests.
- **FR-018**: `supervisor start` MUST accept `--no-lightweight`. When present, the instance MUST be started with lightweight profile **disabled** per session 2026-11-23. Supervisor MUST persist whether lightweight profile is enabled for that instance (for example `lightweight: true|false` on the registry row) so the worker applies the same choice at browser open. Default: **headless → lightweight enabled**; **headed → headed-safe tier only** unless `--no-lightweight` (then none).
- **FR-003**: The lightweight profile MUST NOT remove capabilities required for explicitly supported casino automations and documented stand-in entry paths; where a saving conflicts with a required capability, the capability MUST be kept and the conflict MUST be recorded in the feature plan as a compatibility guardrail.
- **FR-004**: Per-instance isolation MUST remain unchanged: one dedicated browser profile and process tree per instance; optimizations MUST NOT merge or share browsers between instances.
- **FR-005**: `supervisor status` MUST include a **NOTE** column on every row (empty when no message applies). MEMORY MUST continue to reflect per-instance aggregated memory for the owner and that instance’s browser process tree only (CPU reporting unchanged from prior specs). All memory quantities in MEMORY MUST be expressed in **mebibytes (MiB)** with the same display convention: one decimal, a space, then `MiB` (e.g. `245.3 MiB`; limited rows `245.3 MiB/512.0 MiB`). When the instance has no configured `--memory-limit`, MEMORY shows current usage only. When `--memory-limit` is configured, MEMORY MUST use `{current-mib}/{limit-mib}` with that spacing. While `starting`, current memory is `unavailable` (denominator may still show the configured limit, e.g. `unavailable/512.0 MiB`). Refresh and concurrency behavior from prior specs MUST NOT regress except for this column contract.
- **FR-011**: `supervisor start` MUST accept an optional `--memory-limit` argument as a **suffixed binary size string** that sets the configured maximum in **mebibytes**. Parsing rules MUST be documented in `start --help` and MUST accept `M` or `MiB` as mebibytes and `G` or `GiB` as gibibytes (1024 MiB per GiB). Decimal-suffixed **MB** and **GB** MUST be rejected. Invalid, non-positive, or bare unsuffixed values MUST fail `start` with a clear usage error and no registry row. The parsed limit MUST be compared to the same aggregated RSS basis displayed as MiB in MEMORY. When omitted, the instance has no configured memory maximum and no memory warning logic applies.
- **FR-017**: Baseline measurement, warning-line arithmetic, and integration assertions for memory MUST use MiB consistently; documentation MUST state that Supervisor memory figures are mebibytes, not decimal megabytes.
- **FR-012**: `supervisor start` MUST accept an optional `--threshold` argument: a percentage (1–100) of the instance’s configured `--memory-limit` at which `status` populates NOTE with the threshold-exceeded message. `--threshold` MUST NOT be accepted without `--memory-limit`; invalid combinations MUST fail `start` with a clear usage error and no registry row.
- **FR-013**: When `--memory-limit` is set without `--threshold`, Supervisor MUST treat the warning line as 100% of the configured maximum (warn when aggregated memory is at or above the full limit).
- **FR-014**: For each instance with a configured `--memory-limit`, Supervisor MUST persist the limit and effective threshold percentage on the registry row for that instance so `status` can evaluate warnings across refreshes until `stop`.
- **FR-015**: When a `running` instance’s aggregated memory is at or above its effective warning line (threshold% × memory-limit), `status` MUST set NOTE to a fixed threshold-exceeded message documented in `status --help`; when below the warning line, NOTE MUST be empty. The instance MUST remain `running` and MUST NOT be stopped or restarted automatically in v1. The configured maximum is not hard-enforced in v1 (current may exceed limit in MEMORY; NOTE reflects threshold state).
- **FR-016**: `start` MUST NOT be refused solely because other instances are in a memory warning state or above their configured limits, unless the host cannot launch a browser for an existing documented reason (unchanged start-failure rules).
- **FR-006**: Start, stop, status, screenshot, logging, and registry semantics defined in earlier Supervisor specs MUST remain satisfied unless this spec explicitly supersedes a narrower behavior. This spec supersedes prior **status** table shape for MEMORY and adds NOTE as defined in FR-005. Viewport size for headless is unchanged.
- **FR-007**: The feature MUST include a documented baseline measurement method (idle headless instance on blank page, duration, sample count) and post-change verification using the same method so operators and reviewers can reproduce savings claims.
- **FR-008**: Default automated tests MUST run without a live casino site and MUST include checks that headless instances still start, appear in status with memory when running, stop cleanly, include an automated **lightweight-profile regression** assertion for default `start --headless` (allocator includes adopted flags) and a contrasting check that `start --headless --no-lightweight` does **not** include the headless lightweight bundle (FR-018), verify starts without memory flags show MEMORY without a limit suffix and empty NOTE, verify NOTE column exists on all rows, verify `--threshold` without `--memory-limit` fails usage validation, verify invalid `--memory-limit` strings (including `MB`/`GB`) fail usage validation, verify `512M` and `512MiB` on separate starts yield the same limit denominator and MiB `current/limit` formatting, verify `start` succeeds while another instance is in memory warning state (FR-016), verify five concurrent headless blank starts succeed when the host can launch browsers (SC-005), and verify NOTE shows the threshold message when aggregated memory is driven above the warning line in a controlled test scenario without skip.
- **FR-009**: Operator-facing help for `start` MUST document `--no-lightweight`, default lightweight behavior for headed vs headless, `--memory-limit` binary suffix grammar (`M`/`MiB`, `G`/`GiB`, rejection of `MB`/`GB`), `--threshold`, default 100% when limit set alone, and that all memory is in **MiB**; `status` help MUST document MEMORY MiB `current/limit` formatting, the NOTE column, the threshold-exceeded message text, and when NOTE is empty.
- **FR-010**: If a particular optimization causes start failure on a supported environment, Supervisor MUST fail `start` with a clear message and MUST NOT leave a registry row for that failed attempt, consistent with existing start failure rules.

### Key Entities

- **Lightweight profile**: The named set of browser launch choices Supervisor applies at instance creation to reduce memory and background work while honoring compatibility guardrails; **headless full bundle** by default, **headed-safe** subset only when headed; skippable via `--no-lightweight`.
- **Compatibility guardrail**: A documented rule that forbids a specific optimization when a supported automation or probe requires the corresponding browser capability.
- **Resource snapshot**: One `status` row: ID, state, entry, CPU%, MEMORY (current only or `current/limit`), NOTE (optional message), aggregating only that instance’s processes for the current side of MEMORY.
- **Instance memory policy** (optional): Per-instance settings set at `start`: configured maximum (`--memory-limit`) and warning percentage (`--threshold`); absent unless operator passes `--memory-limit`.
- **Warning line**: Aggregated memory level equal to `(threshold ÷ 100) × memory-limit` for that instance; at or above this line triggers the `status` memory warning in v1.
- **Threshold NOTE message**: Fixed operator-facing text in the NOTE column when aggregated memory is at or above the warning line; empty otherwise.
- **Baseline measurement**: A repeatable procedure and recorded median idle memory for pre-change Supervisor on a reference host, used to judge optimization (SC-001); independent of optional per-instance limits.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On the documented baseline procedure, median idle memory (MiB) per headless instance on `about:blank` after the change is at least 20% lower than the recorded pre-change median on the same reference host (same instance count, same measurement window). Verification is recorded in the feature **Baseline Record** (research/plan); it is not required to gate the default CI run unless the repository later adds a documented reference-host job.
- **SC-002**: Operator documentation describes `--no-lightweight`, optional `--memory-limit` and `--threshold`, MEMORY `current/limit` format, and NOTE threshold message; default integration tests cover lightweight on/off, no-limit starts, invalid flag combination, `current/limit` MEMORY when limited, and NOTE populated when above the warning line.
- **SC-003**: 100% of default integration tests for headless start, status, stop, screenshot, and registry lifecycle pass without requiring a live casino site.
- **SC-004**: 100% of documented stand-in or integration scenarios that validated supported game entry and iframe readiness before this feature still pass after the lightweight profile ships, with no new failures attributable solely to resource tuning.
- **SC-005**: An operator running five concurrent headless blank instances on the reference host completes start for all five without artificial cap errors when the host could launch five instances before the change (no reduction in achievable concurrent count attributable to optimization).
- **SC-006**: Within one refresh cycle of `status`, stopped optimized instances disappear from the list. For a single default headless instance on the reference host, `status` MEMORY (MiB) after optimization SHOULD be **observably lower** than the **Baseline Record** pre-change median for the same procedure (same as SC-001 comparison); validated manually via [quickstart.md](./quickstart.md) §8 / task T029—not a default CI numeric gate.

## Assumptions

- Primary optimization target is headless automation instances with lightweight **on** by default; operators may use `--no-lightweight` on headed or headless runs. Headed instances without the flag receive only headed-safe lightweight extras defined in the feature plan.
- The existing fixed headless layout dimensions from spec 012 remain in force; this feature does not change viewport width or height.
- Supported casino and game identifiers and CLI-only operator surface remain per project constitution; no web UI or alternate control plane.
- Chrome or Chromium is provided on the host (Linux/WSL2); absolute MiB will vary by binary version and host, but relative improvement thresholds in Success Criteria apply.
- All operator-facing memory figures, limits, and comparisons use **mebibytes (MiB)**; shorthand `M`/`G` on `--memory-limit` means MiB/GiB, not decimal MB/GB.
- Some vendor embeds require graphics, canvas, fonts, or background network behavior; the plan phase will enumerate guardrails rather than maximizing savings at any functional cost.
- Per-game memory policies and fleet-wide default limits in config files are out of scope for v1; only per-`start` optional flags apply.
- **Automatic restart** (or any automatic remediation) when memory exceeds the configured maximum or warning line is **out of scope for v1**; this spec delivers optimization, optional limit flags, and `status` warnings only. A follow-on feature will define restart policy, timing, and operator impact (including behavior at 100% of `--memory-limit`).
- Primary success metric for this feature is **memory**; CPU may improve collaterally but is not a gating success criterion in v1.
- Secrets, cookies, and session files continue to be supplied at runtime only; lightweight tuning must not log credentials.
- Baseline capture occurs during planning/implementation on the same repository version line immediately before applying optimization changes, stored in the feature plan or research artifact for auditability.