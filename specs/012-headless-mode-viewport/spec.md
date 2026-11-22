# Feature Specification: Headless Viewport 1280×720

**Feature Branch**: `012-headless-mode-viewport`

**Created**: 2026-11-22

**Status**: Draft

**Input**: User description: "increment headless viewport to 1280×720"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Predictable headless layout for automations (Priority: P1)

An operator starts a supervised browser with `--headless` to run casino automations (for example OneWin Aviator). Today the visible page area in headless mode follows browser defaults, which are smaller than a typical desktop layout and can hide or shrink embedded game UIs. After this change, every headless instance presents a consistent 1280×720 page area so layout-sensitive scripts and readiness checks behave the same on every run and host.

**Why this priority**: Game entry and iframe probes depend on elements having enough layout space; an undersized default viewport causes flaky or permanent failures in headless CI and scripts.

**Independent Test**: Start one headless instance, navigate to a stand-in page that reports its layout size (or capture dimensions via the existing screenshot path), and verify width 1280 and height 720 before any operator-specific casino session is required.

**Acceptance Scenarios**:

1. **Given** an operator runs `supervisor start --headless` and the instance reaches `running`, **When** the page layout size is observed (via a documented test hook or screenshot metadata), **Then** the layout width is 1280 pixels and the layout height is 720 pixels.
2. **Given** a headless instance is already `running`, **When** the operator captures a screenshot of that instance, **Then** the image dimensions are 1280×720 (same as the live viewport, per existing screenshot semantics: current view only).
3. **Given** two headless instances started on the same host, **When** both are running, **Then** each uses the same 1280×720 layout independently (isolation unchanged).

---

### User Story 2 - Headed mode unchanged (Priority: P2)

An operator sometimes runs without `--headless` to watch the browser window. They expect window sizing and user-resize behavior to remain as today; only headless runs adopt the fixed viewport.

**Why this priority**: Avoids surprising operators who rely on a visible Chrome window for debugging.

**Independent Test**: Start without `--headless`, confirm the registry records headed mode, and verify no forced 1280×720 layout is applied when headless is false.

**Acceptance Scenarios**:

1. **Given** an operator runs `supervisor start` without `--headless`, **When** the instance is `running`, **Then** the browser is shown in a visible window and viewport sizing is not overridden to 1280×720.
2. **Given** registry rows for headed and headless instances, **When** status is inspected, **Then** the `headless` flag continues to distinguish them; viewport policy applies only where `headless` is true.

---

### User Story 3 - No new operator flags for v1 (Priority: P3)

Scripts and quickstarts that already pass `--headless` should gain the larger viewport without new arguments or config files.

**Why this priority**: Keeps CLI surface area stable and matches constitution “CLI as the operator interface” with minimal new concepts.

**Independent Test**: Run existing integration tests that use `--headless` without changing invocation syntax; they pass after expected dimension assertions are added or updated where tests encode pixel geometry.

**Acceptance Scenarios**:

1. **Given** documented quickstart flows that use `start --headless`, **When** an operator follows them unchanged, **Then** instances still start successfully and automations see the enlarged layout.
2. **Given** help text for `start`, **When** the operator reads `--headless` documentation, **Then** it states (or references) that headless runs use a 1280×720 layout; no new sizing flags are required for v1.

---

### Edge Cases

- Instance transitions from `starting` to `running`: viewport must already be 1280×720 before game entry scripts run (set at browser open, not lazily after navigation).
- Blank `about:blank` headless starts (no session file): same 1280×720 layout applies so screenshot and worker defaults stay consistent.
- Device pixel ratio: layout size is 1280×720 CSS/layout pixels; screenshot byte dimensions match that layout unless a future spec defines otherwise.
- Headless instance stopped and restarted: each new headless run applies 1280×720 again; no persistence of custom sizes in the profile for v1.
- Mixed headed/headless fleet: bulk `stop all` and per-ID stop behavior unchanged; viewport policy does not affect lifecycle commands.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When an instance is started with headless mode enabled (`--headless` recorded on the registry row), Supervisor MUST configure the browser page layout to width 1280 pixels and height 720 pixels before substantive navigation for that run (including the initial blank or casino page).
- **FR-002**: Headed instances (headless not enabled) MUST NOT apply the fixed 1280×720 override; visible-window behavior remains unchanged from prior releases.
- **FR-003**: Screenshot capture for a headless instance MUST reflect the 1280×720 viewport (image width 1280, height 720) when capture succeeds under normal conditions.
- **FR-004**: The headless viewport dimensions MUST be fixed constants for v1 (1280 width, 720 height); operators MUST NOT need a new CLI flag or config file to obtain this size.
- **FR-005**: Registry storage for instances MUST continue to record headless as today; viewport size MUST NOT require a new persisted field for v1 (dimensions are implied by headless policy).
- **FR-006**: Automated tests that start `--headless` instances MUST include at least one check that the effective layout or screenshot size is 1280×720, without requiring a live casino site in the default test run.
- **FR-007**: `start --help` (or equivalent operator documentation) MUST mention that headless runs use a 1280×720 layout so operators understand screenshot and layout behavior.

### Key Entities

- **Headless layout policy**: The rule that headless=true implies layout 1280×720 for the browser tab Supervisor controls for that instance.
- **Layout observation**: A verifiable signal of width and height (test page reporting, screenshot decode, or equivalent) used in acceptance and integration tests.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In 10 consecutive headless starts on a typical development host, 100% report layout or screenshot dimensions of 1280×720 before stop.
- **SC-002**: Headed starts show zero regression in existing start/stop/status integration tests that do not assert headless dimensions.
- **SC-003**: Aviator or other layout-sensitive headless entry paths that previously failed solely due to undersized default viewport succeed in documented stand-in or integration scenarios at 1280×720 (or document remaining failures as unrelated).
- **SC-004**: Operators need no new command-line arguments beyond existing `--headless` to obtain the enlarged viewport.
- **SC-005**: Screenshot PNGs from headless instances have decoded size 1280×720 in the default integration screenshot test path.

## Assumptions

- “Increment” means raising headless layout from browser defaults (commonly smaller than 1280×720) to a fixed 1280×720 HD-ish layout suitable for embedded casino games and iframes.
- 1280×720 is width × height in layout/CSS pixels, landscape orientation.
- Per-operator configurable viewport sizes, mobile emulation, and dynamic resize after start are out of scope for v1.
- Supervisor continues to own browser lifecycle per project constitution; this spec does not change casino identifiers, session files, or stop/start contracts.
- Linux and WSL2 remain supported operator environments; dimension checks in tests may use local stand-in pages and the existing screenshot CLI.