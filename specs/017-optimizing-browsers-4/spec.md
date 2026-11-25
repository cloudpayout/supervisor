# Feature Specification: Headless Shell and Performance Flags (Phase 4)

**Feature Branch**: `017-optimizing-browsers-4`

**Created**: 2026-11-24

**Status**: Draft

**Input**: User description: "@specs/015-optimizing-browsers-3/spec.md was successful, help me to implement these flags too @chrome-performance-flags.txt. Also let's try to install `chrome-headless-shell` into the project to use it by default." (slug: optimizing-browsers-4). Additional: "the browser profile shouldn't be kept, that file can grow as the automation is executing."

**Operator performance flag proposals** (evaluate each for adopt / reject / already present on the **015** default headless lightweight profile; apply only **adopt** on default `start --headless`):

- `disable-gpu-compositing`
- `disable-gpu-rasterization`
- `disable-accelerated-2d-canvas`
- `disable-accelerated-video-decode`
- `disable-component-update`
- Additional `disable-features` entries: `OnDeviceModel`, `PaintHolding`, `AudioServiceOutOfProcess` (merge with existing casino/feature disables without contradicting supported games)
- `no-sandbox`
- `no-zygote`
- `single-process`
- (Duplicates against **013–015** such as `disable-crash-reporter`, `disable-breakpad`, `media-cache-size=1`, `disable-application-cache`, `renderer-process-limit=1`, `process-per-site` are **already present** candidates—not re-added)

Proposals that remain **forbidden by policy** unless a spec amends with measured proof:

- `disable-background-networking` on headless (Spribe / iframe bootstrap policy from prior specs)

**Headless browser selection (operator constraint)**:

- `supervisor start --headless` MUST launch only the project **chrome-headless-shell** binary.
- The `CHROME_PATH` environment variable MUST NOT be used on headless starts to select full Chrome, Chromium, or any binary other than **chrome-headless-shell** (ignore `CHROME_PATH` on headless, or treat a set `CHROME_PATH` as an error if it is not the headless-shell artifact—plan chooses UX; headless never falls back to PATH discovery of full Chrome).

**Ephemeral browser profile (operator constraint)**:

- Each instance MUST use a **fresh, dedicated** on-disk user-data profile only for the life of that run. The profile MUST NOT be retained, reused, or archived after the instance stops, crashes, or fails to start.
- While automation is `running`, Chrome may still write cache and storage under that profile; Supervisor MUST treat that directory as **disposable** and remove it when the run ends so disk use does not accumulate across sessions.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Adopt safe performance flags on the 015 baseline (Priority: P1)

An operator on Linux or WSL2 runs default headless after spec **015** shipped (`memory-pressure-off` and the 013/014 tiers without the **016** aggressive GPU/WebGL experiment). Idle MEMORY is still high (~1.1 GiB class on the reference host). They want the switches listed in `chrome-performance-flags.txt` evaluated and merged using the same adopt/reject discipline as **015**.

**Why this priority**: Extends a profile that already passed stand-ins; avoids repeating rejected networking and GPU kills without measurement.

**Independent Test**: Publish a per-flag matrix; default headless includes only **adopt** flags on top of the **015** profile; compare idle median MiB to the documented **015** median on the same host.

**Acceptance Scenarios**:

1. **Given** the operator list in this spec Input, **When** planning completes, **Then** every listed switch has **adopt**, **reject**, or **already present** with a testable reason.
2. **Given** `supervisor start --headless` with no opt-out, **When** the instance is `running`, **Then** only **adopt** flags from the list appear (no duplicates; forbidden switches absent).
3. **Given** a proposal conflicts with casino automation policy (background networking suppression, disabling capabilities required for canvas/WebGL games, or **015**/**016** documented rejects), **When** the default profile is built, **Then** that proposal is **reject**ed.
4. **Given** `supervisor start --headless --no-lightweight`, **When** the instance is `running`, **Then** no phase-4 performance extras and no 013/014/015 lightweight RAM tiers apply (unchanged **013** opt-out).

---

### User Story 2 - Default headless uses project chrome-headless-shell (Priority: P1)

The operator wants headless automations to start using **chrome-headless-shell** supplied with the project, without manually installing full Chrome for headless-only hosts. Full browser remains available for **headed** runs only (`CHROME_PATH` may still apply there per prior specs).

**Why this priority**: Headless shell is a smaller process footprint and matches the headless-only automation path.

**Independent Test**: Fresh checkout provisions the headless shell artifact; every headless `start` launches that binary even if `CHROME_PATH` points at full Chrome.

**Acceptance Scenarios**:

1. **Given** the project bundle step completed on a supported platform, **When** the operator runs `supervisor start --headless`, **Then** the running instance uses **chrome-headless-shell** from the project (not the system full Chrome).
2. **Given** a headed `supervisor start`, **When** the instance is `running`, **Then** the full Chrome/Chromium discovery path is used (headless shell is not forced on headed starts; `CHROME_PATH` behavior for headed is unchanged from pre-017 specs).
3. **Given** `CHROME_PATH` points at full Chrome (or any non-headless-shell binary), **When** the operator runs `supervisor start --headless`, **Then** Supervisor still launches **chrome-headless-shell** from the project and does not use `CHROME_PATH` to pick that other binary.
4. **Given** the headless shell bundle is missing or the platform is unsupported, **When** headless `start` runs, **Then** the operator sees a clear error with provisioning steps; headless MUST NOT silently fall back to full Chrome via `CHROME_PATH` or PATH.

---

### User Story 3 - Games and stand-ins keep working (Priority: P1)

Supported casino stand-ins (for example OneWin Aviator) must pass on the **adopt**ed phase-4 profile and on **chrome-headless-shell** before the feature closes.

**Why this priority**: Performance and shell swaps can break iframe games even when idle RAM improves.

**Independent Test**: Re-run documented stand-in suite and default headless integration tests after adoptions and after switching default headless binary.

**Acceptance Scenarios**:

1. **Given** phase-4 flags and headless shell are active on default headless, **When** stand-in tests run, **Then** results are recorded and match or exceed the **015** reference pass rate for failures not attributable to environment drift.
2. **Given** `single-process`, `no-zygote`, or `no-sandbox` would ship on the default profile, **When** stand-ins or startup smoke fail, **Then** those switches are **reject**ed for default headless (partial adoption allowed).
3. **Given** a flag regresses stand-ins, **When** the matrix is updated, **Then** that flag is **reject**ed even if idle RAM improves.

---

### User Story 4 - Ephemeral browser profile (Priority: P1)

An operator runs long casino automations. Chrome’s user-data profile on disk can grow during execution (cache, IndexedDB, service data), wasting disk and sometimes contributing to memory pressure. They want Supervisor **not to keep** profile directories after a run and not to reuse profiles across instances.

**Why this priority**: Unbounded profile growth undermines RAM tuning goals and clutters the data directory.

**Independent Test**: Start headless, let automation run, `stop` the instance; verify the instance profile path no longer exists. Repeat starts get new profile paths. Orphan/crash paths still trigger cleanup per existing lifecycle specs.

**Acceptance Scenarios**:

1. **Given** a new `supervisor start --headless`, **When** the instance reaches `running`, **Then** it uses a newly created profile directory unique to that instance ID (no reuse of a prior instance’s profile).
2. **Given** a `running` instance, **When** the operator runs `supervisor stop <id>` and the worker exits, **Then** that instance’s profile directory is removed from disk.
3. **Given** a worker crash, failed start rollback, or reconcile of a dead instance, **When** cleanup runs, **Then** the associated profile directory is removed and does not remain under the supervisor data layout.
4. **Given** ten sequential start/stop cycles on the same host, **When** no instances are `running`, **Then** no orphaned profile directories from those instances remain (only active runs may have a profile on disk).

---

### User Story 5 - Transparent RAM outcome (Priority: P2)

If idle MEMORY does not drop, the operator can read what was tried, what shipped on the **015** baseline, and why.

**Why this priority**: **015** and **016** already showed idle `about:blank` may not move with tuning alone.

**Independent Test**: Baseline artifact records **015** vs phase-4 median on the same procedure; docs list the matrix and headless-shell version.

**Acceptance Scenarios**:

1. **Given** phase-4 completes, **When** the operator reads feature docs, **Then** they see **median_015_mib**, **median_017_mib**, and **adopt/reject** outcomes.
2. **Given** SC-001 is not met, **When** docs are published, **Then** they state that remaining proposals were rejected or ineffective without implying `--memory-limit` caps RSS.

---

### Edge Cases

- **016 experimental** GPU/WebGL disables may still be on tree: phase 4 **replaces** default headless tuning with **015 + adopt** list; overlapping **016** experiment flags are removed from default headless unless the matrix explicitly readopts them with stand-in proof.
- Headless shell version skew vs full Chrome on the same host: pin versions in the bundle step; record in baseline.
- `no-sandbox` weakens isolation: default **reject** for operator-facing default profile unless plan documents a constrained dev-only exception (not default).
- `single-process` / `no-zygote` caused startup failures on reference Chrome in **016**: default **reject** until smoke and stand-ins pass on headless shell.
- Duplicate `media-cache-size` or feature-list entries in the operator file: dedupe on argv.
- WSL2 without sandbox support: failures surface at preflight with actionable message.
- Operator sets `CHROME_PATH` for headed debugging: headless starts MUST ignore it for binary selection (document in help/docs).
- Games requiring GPU features headless shell lacks: stand-ins **reject** related flags or block shell default until proof exists.
- Long automation sessions: in-run profile growth is expected; **stop** and worker teardown MUST still delete the profile so only **running** instances may have profile data on disk.
- Operator inspects data directory while instances run: profile size may grow; after **stop**, that subdirectory MUST be gone.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST evaluate every switch in this spec’s operator performance list against the **015** default headless lightweight profile and publish an adopt/reject/**already present** matrix before implementation.
- **FR-002**: Only **adopt** switches MAY be added on default `supervisor start --headless` (lightweight on); headed and `--no-lightweight` paths MUST stay performance-neutral per **013** (no phase-4 RAM extras on opt-out).
- **FR-003**: `disable-background-networking=true` MUST NOT appear on the default headless profile; headless unset/bootstrap policy from prior specs remains.
- **FR-004**: Proposals for `no-sandbox`, `single-process`, and `no-zygote` MUST default to **reject** on the operator-facing default profile until stand-in and startup smoke on **chrome-headless-shell** pass; security-sensitive `no-sandbox` MUST NOT ship on default without an explicit spec amendment for production use.
- **FR-005**: The project MUST include a repeatable way to obtain **chrome-headless-shell** for supported developer platforms (bundle, download script, or documented install step checked into the repo) with a pinned version recorded in feature artifacts.
- **FR-006**: Headless launches MUST use only the project **chrome-headless-shell** (provisioned per **FR-005**); headed launches MUST NOT use headless shell by default.
- **FR-007**: On `supervisor start --headless`, Supervisor MUST NOT use `CHROME_PATH` to select full Chrome, Chromium, or any binary other than **chrome-headless-shell**. On headed starts, `CHROME_PATH` and PATH discovery behavior remain as in specs **001** and **013** (full browser only).
- **FR-008**: Record **median_015_mib** (reference) and **median_017_mib** (post-change default headless) using the **013–015** idle procedure (30s warmup, 150 samples); optionally record headless `--no-lightweight` as control labeled `performance_flags=none`.
- **FR-009**: Stand-in suite on default headless MUST run before feature close; results recorded in the baseline artifact.
- **FR-010**: `status`, NOTE, `--memory-limit`, `--threshold`, and registry semantics from **013** MUST remain unchanged; this spec changes launch binary selection and Chrome tuning only.
- **FR-011**: Default headless profile MUST remove **016 experimental** GPU/WebGL/js bundle overrides from the default lightweight path unless the adopt matrix explicitly reintroduces an equivalent switch with stand-in proof.
- **FR-012**: Each instance MUST be assigned a **new** browser user-data profile directory at `start`; profiles MUST NOT be shared across instance IDs or reused after `stop`.
- **FR-013**: Supervisor MUST delete the instance profile directory on normal `stop`, worker shutdown, failed-start rollback, and orphan reconcile paths (headed and headless) so profile data is **not kept** after the run ends.
- **FR-014**: Phase-4 launch tuning (adopted performance flags plus **015** baseline) MUST include measures that limit unnecessary on-disk profile growth **during** automation where compatible with stand-ins (for example cache and application-cache policy already in lightweight tiers); plan documents any additional Chrome switches adopted for this purpose.

### Key Entities

- **015 baseline profile**: Shipped default headless lightweight after spec **015** (RAM comparison reference).
- **Operator performance proposals**: Switches from `chrome-performance-flags.txt` and this spec Input.
- **Adopt/reject matrix (017)**: Per-switch decision with rationale and test links.
- **Headless shell bundle**: Project-local **chrome-headless-shell** binary (or platform-specific artifact) and version pin.
- **Baseline record (017)**: Medians, stand-in pass/fail, headless-shell version, effective argv notes.
- **Ephemeral instance profile**: Per-instance user-data directory; exists only while the instance is starting, running, or stopping; removed after teardown.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On the documented idle procedure, median aggregated MEMORY (MiB) with the phase-4 default profile is at least **10% lower** than the documented **015** median on the **same reference host**, **or** the matrix and baseline artifact record that remaining proposals were **reject**ed or ineffective (measured outcome documented).
- **SC-002**: **100%** of headless starts use **chrome-headless-shell** from the project, including when `CHROME_PATH` is set to a full Chrome binary in the same shell session.
- **SC-003**: **100%** of documented OneWin/Aviator stand-in scenarios that passed under **015** still pass with phase-4 adoptions and default headless shell (or failing switches are **reject**ed and documented).
- **SC-004**: **100%** of default integration tests for headless start, status, stop, and screenshot pass without a live casino site.
- **SC-005**: Operator-facing documentation lists each proposed performance switch with adopt/reject/**already present**, headless-shell provisioning steps, and explicit rule that headless ignores `CHROME_PATH` for full-Chrome selection (`CHROME_PATH` remains for headed only).
- **SC-006**: **100%** of integration scenarios that start and then `stop` an instance leave **no** profile directory for that instance ID under the supervisor profiles layout after cleanup completes.
- **SC-007**: In a documented test of **N** consecutive start/stop cycles (plan defines **N**, minimum 5) with no instances left `running`, **zero** leftover profile directories from those instance IDs remain on disk.

## Operator amendments (shipped 2026-11-24)

After initial **017** implement, the operator re-shipped the **016** experimental slice on default headless lightweight, plus explicit **`disable-breakpad`**, **`single-process`**, **`no-zygote`**, and **`no-sandbox`**, while keeping **`disable-background-networking=false`**. This amends the original **FR-011** revert for branch **017-optimizing-browsers-4**. Canonical argv: [contracts/shipped-lightweight-argv-amendment.md](./contracts/shipped-lightweight-argv-amendment.md).

## Assumptions

- Spec **015** remains the RAM comparison reference; default shipped argv may combine **015** tiers, **017** phase-4, and **016** experiment per operator amendment above.
- Operators develop on Linux or WSL2; bundle step targets those platforms first (other platforms may document manual install).
- **chrome-headless-shell** is distributed under Google Chrome terms; the project documents license and download source without committing proprietary blobs unless policy allows (plan chooses vendor dir vs script-only).
- Several operator proposals are already on **013–015** profiles; phase 4 treats duplicates as **already present**.
- Headed casino debugging may still require full Chrome with WebGL (`CHROME_PATH` applies to headed only); headless always uses **chrome-headless-shell**, never full Chrome via env override.
- Hard RSS caps and automatic stop/restart remain **out of scope**.
- Session cookies and in-run state live in the ephemeral profile until `stop`; persistence across runs is **out of scope** (operators restart sessions via session files as today).

## Dependencies

- **Spec 015**: Adopt/reject methodology and shipped default profile reference.
- **Specs 013–014**: Lightweight tiers and `--no-lightweight` semantics.
- **Spec 016** (if merged): Experimental slice must not contradict **FR-011** without matrix proof.
- **Compatibility policy**: Bug try-limit-of-app-config-exceeded-4 (headless networking/canvas).
- **Constitution**: CLI operator surface; default tests without live casino; chromedp browser control.
- **Specs 001 / stop & worker lifecycle**: Existing stop and worker cleanup behavior; phase 4 makes ephemeral profiles an explicit operator requirement and validates it in tests (**FR-012**, **FR-013**).