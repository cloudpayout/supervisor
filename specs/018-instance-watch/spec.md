# Feature Specification: Instance Watch (Local Live View)

**Feature Branch**: `feature/instance-watch`

**Feature Directory**: `specs/018-instance-watch`

**Created**: 2026-11-25

**Status**: Draft

**Input**: User description: "use handoff `.specify/assessments/real-time-streaming/decision.md` to develop the feature" (assessment slug `real-time-streaming`)

**Assessment handoff**: Operators cannot continuously see what a running instance is showing without repeated still captures or tools outside Supervisor. **Go** on worker-served **loopback** live view: new operator command, local browser URL (default port **8080**, next free if busy), default **1** frame per second up to **30**, ends on interrupt or when the instance is no longer running; **`screenshot`** stays separate and may run while a watch is active; headless and headed both supported; sparse updates on static pages at low cadence are acceptable.

This feature **adds** a command; it does not change `start`, `status`, `stop`, `screenshot`, or `logs` except where this spec explicitly amends behavior.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Watch a running instance in the local browser (Priority: P1)

An operator who started an instance wants to see what that browser is showing over time without running still capture on every glance. They start watch with the instance ID, open the printed local address in a browser, and see the page update while the instance keeps running.

**Why this priority**: Continuous local observation is the core outcome of the assessment.

**Independent Test**: Start one headless instance on a stand-in page, run watch with that ID, open the printed URL, confirm a live multipart image stream appears, leave watch running at least five minutes, then interrupt watch and confirm the instance is still running.

**Acceptance Scenarios**:

1. **Given** an instance that status shows as running, **When** the operator runs watch with that instance ID, **Then** the command prints one local HTTP URL as a single line on standard output, the operator can open that URL in a browser on the same machine and see a live view of that instance's current page, and the command keeps running until the operator interrupts it or the instance is no longer running.
2. **Given** an active watch, **When** the operator views the URL, **Then** the view shows only that instance's page (current viewport), not another instance's page and not the operator's whole desktop.
3. **Given** one running headless instance and one running headed instance, **When** the operator watches each by its own ID on a machine where each mode is available, **Then** each watch shows that instance's page. The headless case gates the default automated check; headed is a manual check when a display is available.
4. **Given** a successful watch session, **When** the operator checks status after interrupting watch, **Then** the instance is still running with the same owner and browser process as before watch, unless the instance ended on its own during watch.
5. **Given** the operator asks for watch help, **When** help is shown, **Then** it describes the watch command, the required instance ID, the `--fps` flag, that the printed URL is local only, and that interrupting the command ends the watch session.

---

### User Story 2 - Refuse watch that cannot succeed (Priority: P1)

An operator or script may pass a missing, unknown, or not-yet-running ID, or watch may be impossible because nothing is serving the stream. The command must fail clearly.

**Why this priority**: Scripts must distinguish usage errors, missing instances, and not-running instances from a successful watch.

**Independent Test**: Run watch with no ID, an unknown ID, a not-running ID, and a running instance started by an older binary without watch support; confirm clear failures and no misleading URL.

**Acceptance Scenarios**:

1. **Given** watch is run with no instance ID, more than one ID, or a blank ID, **When** the command runs, **Then** it does not prompt on standard input, explains that an instance ID is required, and fails as a usage error.
2. **Given** an instance ID that status does not list, **When** the operator runs watch with that ID, **Then** the command fails, reports that the instance was not found, and does not print a watch URL.
3. **Given** a tracked row that status shows as not running (including starting, stopping, or stopped), **When** the operator runs watch with that ID, **Then** the command fails, reports that the instance is not running, and does not print a watch URL.
4. **Given** a running instance whose worker cannot start a watch stream, **When** the operator runs watch, **Then** the command fails with a clear reason, does not print a success URL, and does not leave a listen port serving unrelated content.

---

### User Story 3 - Control update cadence (Priority: P2)

An operator wants a slower default update rate for light monitoring or a higher rate when more motion detail is needed, within a fixed upper limit.

**Why this priority**: Cadence control is a stated handoff requirement; default low rate limits load.

**Independent Test**: Watch the same running instance with default flags and with an explicit higher `--fps` within the allowed maximum; confirm the stream updates at the configured rate within reasonable tolerance on a stand-in page that changes over time.

**Acceptance Scenarios**:

1. **Given** a running instance, **When** the operator runs watch without `--fps`, **Then** the stream targets **1** frame per second on average over a one-minute window on a stand-in page that changes at least once per second.
2. **Given** a running instance, **When** the operator runs watch with `--fps` set to an integer between **1** and **30** inclusive, **Then** the stream targets that rate within the same tolerance rules documented for success criteria.
3. **Given** watch with `--fps` outside **1–30** or non-integer, **When** the command runs, **Then** it fails as a usage error and does not start a stream.
4. **Given** a headless instance on a page that does not change, **When** the operator watches at default cadence, **Then** fewer visual updates than the configured rate are acceptable; the stream does not fail solely because the page is static.

---

### User Story 4 - Still capture while watching (Priority: P2)

An operator who is watching still needs a one-off still image for evidence without stopping watch.

**Why this priority**: Assessment requires `screenshot` to remain available in parallel.

**Independent Test**: Start watch on a running instance, open the URL, run screenshot on the same ID in another terminal, confirm a PNG is produced and watch continues until interrupted.

**Acceptance Scenarios**:

1. **Given** an active watch on a running instance, **When** the operator runs screenshot with that instance ID, **Then** screenshot completes per spec 003 or fails with a clear error, and watch continues if the instance is still running and the operator has not interrupted watch.
2. **Given** screenshot and watch are both active, **When** either command finishes or fails, **Then** the other instance state and unrelated instances are unchanged except as defined in specs 003 and this spec.

---

### User Story 5 - Watch ends when the run ends (Priority: P1)

An operator must not keep watching a stale picture after the supervised browser has stopped.

**Why this priority**: Immediate end on teardown is an explicit handoff rule.

**Independent Test**: Start watch, stop the instance from another terminal without interrupting watch first; confirm watch command exits within a few seconds and the browser stream stops.

**Acceptance Scenarios**:

1. **Given** an active watch, **When** the operator stops that instance with `stop`, **Then** the watch command exits within **5** seconds, the stream at the printed URL stops delivering new frames, and the exit is not reported as watch success unless the spec defines a distinct success class for natural end.
2. **Given** an active watch, **When** the instance disappears because of supervised failure teardown (for example entry failure that removes the instance), **Then** watch ends within **5** seconds without requiring the operator to interrupt watch first.
3. **Given** an active watch, **When** the operator sends interrupt (for example Ctrl+C), **Then** watch ends, the stream stops, and the instance remains running if it was running before interrupt.

---

### Edge Cases

- Watch never reads standard input and never prompts for an instance ID.
- A whitespace-only ID is a usage error.
- Only **one** local watch session per instance at a time; a second watch against the same ID while one is active fails with a clear reason.
- The printed URL uses loopback only; it is not reachable from remote machines.
- No authentication is required to open the URL; any local process that can reach loopback may view the stream (stakeholder accepted).
- The stream may show sensitive page content; the command does not redact or record a gallery.
- Default local port is **8080**; if it is already in use by another program or another watch, the product chooses the next available port on loopback and prints the actual URL.
- Multiple running instances may each be watched at once from separate terminals; each receives its own URL and port when ports are available.
- Platforms other than Linux and WSL2 are out of scope, consistent with the existing operator surface.
- Default automated tests use headless stand-in pages and do not require a live casino site.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST provide a `watch` command that offers a continuous local live view of a running browser instance identified by its instance ID.
- **FR-002**: The command MUST take the instance ID as its only positional argument and MUST NOT prompt on standard input.
- **FR-003**: When the ID is missing, blank, or not the sole positional argument, the command MUST fail as a usage error.
- **FR-004**: When status shows the instance as running and watch can start, the command MUST print exactly one local HTTP URL on standard output as a single line and MUST keep running until the operator interrupts it or the instance is no longer running.
- **FR-005**: While watch is active, the URL MUST serve a browser-viewable live image stream of that instance's current viewport for a single local viewer at a time.
- **FR-006**: Watch MUST leave the instance running and MUST NOT navigate, click, resize, stop, or replace that instance solely to produce the stream.
- **FR-007**: When no tracked row exists for the ID, the command MUST report that the instance was not found. When a tracked row exists but its state is not running, the command MUST report that the instance is not running. Both cases MUST NOT print a watch URL.
- **FR-008**: When watch cannot start or the stream fails irrecoverably, the command MUST fail with a clear reason and MUST NOT print a misleading success URL.
- **FR-009**: The command MUST accept `--fps` with an integer from **1** through **30** inclusive. When omitted, the default MUST be **1**. Invalid values MUST be usage errors.
- **FR-010**: When the instance is no longer running, watch MUST end within **5** seconds without requiring a second operator action.
- **FR-011**: When the operator interrupts watch while the instance is still running, watch and the stream MUST stop and the instance MUST remain running.
- **FR-012**: While watch is active on an instance, `screenshot` for that same running instance MUST remain available per spec 003.
- **FR-013**: At most one active watch per instance ID; a concurrent second watch for the same ID MUST fail clearly.
- **FR-014**: The default listen port MUST be **8080** on loopback; if unavailable, the product MUST select the next available loopback port and print the URL that reflects the port actually used.
- **FR-015**: Operator help MUST describe `watch`, the required instance ID, `--fps`, default cadence, the local URL behavior, and that interrupt ends watch.
- **FR-016**: Root help MUST list `watch` alongside existing instance commands with a short description.

### Key Entities

- **Instance**: An existing supervised browser session with a stable ID. Only a running instance may be watched. The ID is the same value `start` printed and `status` lists.
- **Watch session**: One operator-initiated continuous local view tied to one instance ID, one loopback URL, and one active stream until interrupt or instance teardown.
- **Live view**: The sequence of viewport images delivered to the operator's browser at the configured cadence; it is not a stored recording in v1.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In at least **19** of **20** trials on a headless stand-in instance, the operator can open the printed URL and see a live view within **10** seconds of starting watch.
- **SC-002**: In **10** consecutive five-minute watch sessions on a stand-in instance, the operator is not required to run still capture commands to refresh the view; the stream remains available until interrupt or instance stop.
- **SC-003**: In **100**% of trials where the instance is stopped while watch is active, the watch command exits within **5** seconds.
- **SC-004**: **100**% of invocations without exactly one instance ID fail as usage errors and do not print a watch URL.
- **SC-005**: **100**% of attempts against an unknown ID or a not-running row fail without a watch URL and without changing other instances.
- **SC-006**: On a stand-in page with visible periodic change, watch with default `--fps` delivers an average of **0.8–1.2** updates per second over **60** seconds; with `--fps 10`, an average of **8–12** updates per second over **60** seconds, unless the page is static (headless sparse updates allowed per FR-009 acceptance).
- **SC-007**: In at least **9** of **10** trials, screenshot succeeds while watch is active on the same running instance and watch continues afterward until interrupted.
- **SC-008**: After watch ends by operator interrupt only, status still shows the instance as running with the same processes as before watch in at least **19** of **20** trials.

## Assumptions

- The command name is `watch`, matching assessment "watch" language; renaming requires a spec amendment.
- "Running" matches what `status` shows as running; starting instances are not watchable.
- The live view matches the same viewport semantics as still capture for headless instances (**1280×720** layout per spec 012) unless a later spec amends layout.
- Stream format is a browser-viewable multipart image stream over HTTP on loopback (MJPEG-style); operators use a normal browser; no in-terminal video player is required.
- v1 supports one local viewer per watch URL; no viewer authentication.
- Workers started before this feature may lack watch support; operators restart those instances to watch them, analogous to spec 003 capture socket behavior.
- A pre-implementation spike may record p50/p95 update latency and bytes per second on stand-in pages; planning uses existing single-capture time budgets until measured.
- Watch does not write files, upload streams, or add registry fields unless a later spec says otherwise.