# Feature Specification: Supervisor CLI (start, status, stop)

**Feature Branch**: `001-supervisor-cli`

**Created**: 2026-11-07

**Status**: Draft

**Input**: User description: "Develop a CLI program named `supervisor`, it will run automations that run in `chromedp` instances. Initially, `supervisor` will have three commands: `start`, `status` and `stop`. Commands: `start`: No parameters (for now). It starts a `chromedp` instance in a blank page. Stores the instance's ID to monitor it. `status`: Displays all running `chromedp` instances in real-time. It shows the RAM and CPU usage in that moment. `stop`: Asks for an instance ID to kill the `chromedp` process."

## Clarifications

### Session 2026-11-07

- Q: After `supervisor start` succeeds, should the command exit immediately while the browser (chromedp) instance keeps running, or stay running in that terminal until the instance stops? → A: `start` exits after launch; each instance is owned by a detached background process (or agent) that holds the chromedp session until `stop` or crash.
- Q: When several instances run at once, must each instance use its own isolated Chrome browser (separate OS process and separate profile/storage), rather than sharing one browser with multiple tabs or one profile? → A: Each instance is exactly one dedicated Chrome browser with its own isolated profile/storage; no sharing between instances.
- Q: Should Supervisor refuse a new `start` once a fixed maximum number of concurrent instances is already running? → A: No enforced limit in v1; concurrent instances are allowed until the host cannot launch another browser.
- Q: If two operators run `supervisor start` at the same time in different terminals, must both calls be allowed to succeed with distinct instance IDs whenever the host can launch both browsers? → A: Yes; parallel `start` commands may succeed independently, each with a unique ID, when browsers can launch.
- Q: For each instance row in `status`, should CPU and memory include the instance's background helper and that instance's Chrome processes (including Chrome child processes), and exclude other instances? → A: Per instance, aggregate CPU and RAM for the background helper and the full Chrome process tree for that instance only.

### Session 2026-11-09

- Q: If an instance never reaches a running state after `start` has registered it, how long may it stay in the registry before Supervisor removes it and treats the start as failed? → A: Remove after 60 seconds if still not running.
- Q: While an instance is still coming up (registered but not yet running), should `status` list it with a visible non-running state and CPU/RAM as unavailable until it is running? → A: List `starting` instances in `status` with a distinct state; show CPU/RAM as `unavailable` until running.
- Q: When the operator runs `stop` for an instance ID that is already stopped or no longer on the machine, should Supervisor treat that as a successful stop with an informational message, or as an error? → A: Treat as success; exit `0` with a clear already-stopped message; prune stale registry when applicable.
- Q: For v1 of this feature, must Supervisor officially support only Linux and WSL2 environments, with other operating systems explicitly out of scope? → A: v1 supports Linux and WSL2 only; macOS and Windows are out of scope.
- Q: For v1 `start`, must the browser open in a visible window by default, or run headless with no window unless a future flag turns headed mode on? → A: Operator chooses headed vs headless on each `start` via a v1 flag.
- Q: How should the operator choose headed vs headless on `supervisor start` in v1? → A: Mutually exclusive flags `--headed` and `--headless`; default headed when both are omitted; error if both are set.
- Q: When the operator aborts `stop` by submitting an empty instance ID or canceling at the prompt, should the command exit with code 0 or a non-zero error code? → A: Exit `0`; no instance terminated.
- Q: For `stop`, should "not found" mean the ID is absent from the registry, and "already stopped" mean a registry row exists but processes are gone? → A: Both cases exit `0` with distinct operator messages; not found = no registry row; already stopped = row exists but processes gone (prune row).
- Q: When reconciliation detects the owner is dead but Chrome is still running for that instance, should Supervisor only remove the registry row, or also try to terminate the orphan Chrome process? → A: Remove the registry row and terminate the orphan Chrome process tree.
- Q: While `status` is running, if another terminal runs `start` or `stop`, must the next refresh show the updated instance list without duplicate IDs and without omitting IDs that were just added or removed? → A: Within one refresh cycle, the list matches the registry with no duplicate or stale IDs for concurrent `start`/`stop`.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start a supervised browser instance (Priority: P1)

An operator runs the Supervisor CLI to bring up a new isolated browser session on a blank page so a future automation can attach to it. Supervisor assigns a stable instance identifier and records the session as active.

**Why this priority**: Without the ability to start and register instances, monitoring and stopping have nothing to manage. This is the minimum viable lifecycle entry point.

**Independent Test**: Run `supervisor start`, confirm a new identifier is returned or shown, confirm the browser session opens to a blank page, and confirm a subsequent status listing includes that identifier while the session remains running.

**Acceptance Scenarios**:

1. **Given** Supervisor is installed and no start failure occurred, **When** the operator runs `supervisor start` with exactly one of `--headed` or `--headless` or neither (default headed), **Then** a new browser automation instance opens on a blank page, Supervisor exposes a new unique instance ID, the `start` command exits while the instance continues in the background, and `status` may show that ID as `starting` with `unavailable` metrics until running.
2. **Given** one instance is already running in the background, **When** the operator runs `supervisor start` again (same or different terminal), **Then** a second instance is created with a different ID, each with its own Chrome process and profile, both instances keep running concurrently, and both appear in `status`.
3. **Given** the host cannot launch another browser process (resource exhaustion or OS error), **When** the operator runs `supervisor start`, **Then** Supervisor exits non-zero with a clear failure message, does **not** print an instance ID on stdout, and does **not** leave a registry row for that attempt (FR-011). Failures detected synchronously in `start` (e.g. Chrome binary missing, profile path error, detached worker process cannot be started) are handled in the CLI before success output.
4. **Given** an instance remains in a non-running state for 60 seconds after registration, **When** reconciliation runs, **Then** the instance is removed from the tracked set and any leftover browser or owner processes for that ID are torn down where possible.

---

### User Story 2 - Observe running instances and resource usage (Priority: P2)

An operator runs `supervisor status` to see every browser automation instance Supervisor is currently tracking, along with each instance's current CPU and memory usage, updated continuously until the operator ends the view.

**Why this priority**: Real-time resource visibility is the product's core monitoring value and supports decisions about scaling or stopping runs.

**Independent Test**: With at least one instance started, run `supervisor status` and verify all tracked IDs appear with CPU and RAM fields that change over time while workloads run; verify an empty list when no instances are tracked.

**Acceptance Scenarios**:

1. **Given** two instances are running and registered, **When** the operator runs `supervisor status`, **Then** both instance IDs are listed with CPU and memory usage for each, and each row reflects only that instance's owner plus Chrome tree (not the other instance's processes).
2. **Given** no instances are registered, **When** the operator runs `supervisor status`, **Then** Supervisor indicates that no instances are running (empty state message).
3. **Given** an instance exits outside Supervisor, **When** the operator views status, **Then** that instance is no longer listed or is clearly marked as no longer running within one refresh cycle.

---

### User Story 3 - Stop a supervised instance (Priority: P3)

An operator ends a specific browser automation instance by ID when a run should finish or must be reclaimed. Supervisor prompts for the ID when not provided on the command line (initial scope).

**Why this priority**: Completes the lifecycle Supervisor owns: start, monitor, stop.

**Independent Test**: Start an instance, run `supervisor stop`, enter its ID at the prompt, confirm the browser process ends and the instance disappears from status.

**Acceptance Scenarios**:

1. **Given** a running instance with a known ID, **When** the operator runs `supervisor stop` and enters that ID at the prompt, **Then** the corresponding browser process terminates and the instance is removed from the tracked set.
2. **Given** no instance matches the entered ID in the registry, **When** the operator confirms stop, **Then** Supervisor reports that the ID was not found, exits `0`, and does not affect other instances.
3. **Given** the entered ID matches a record but owner/browser processes are already gone, **When** the operator confirms stop, **Then** Supervisor exits `0`, reports already stopped (distinct message from not found), and removes any stale registry entry for that ID.
4. **Given** the operator cancels or submits empty input at the prompt, **When** they abort the stop flow, **Then** no instance is terminated and `stop` exits `0`.

---

### Edge Cases

- What happens when the operator starts many instances in sequence from the same terminal—each `start` must return without waiting for prior instances to stop; each must receive a unique ID, its own browser profile, and remain independently stoppable; there is no artificial cap beyond host capacity.
- What happens when two instances are running—cookies, storage, and sessions MUST NOT leak across profiles; stopping one MUST NOT terminate the other’s browser.
- What happens when a background instance owner crashes while Chrome is still running—or the reverse—registry and `status` must reconcile within one refresh cycle; if the owner is dead but Chrome remains, remove the row and terminate the orphan Chrome tree for that instance profile.
- What happens when Chrome is dead but the owner remains—reconciliation MUST remove the registry row and terminate the owner process if it is still running.
- How does Supervisor handle an instance whose browser process crashed or was killed externally—status must not show it as healthy indefinitely.
- What happens when `status` is run while `start` or `stop` is in progress—the next refresh (within one cycle) MUST reflect registry changes with no duplicate IDs and no omitted IDs that were added or removed by concurrent CLI commands.
- What happens when two `start` commands overlap in time—both MUST register distinct IDs without corrupting the shared registry; neither MUST block the other except for normal OS resource limits.
- How does Supervisor behave when CPU or memory metrics are temporarily unavailable for a process—show a clear placeholder or “unavailable” rather than silent failure.
- What happens on `stop` if the process is already gone—exit `0` with an already-stopped message and prune stale registry rows; unknown ID with no registry row also exits `0` with a not-found message (distinct wording from already stopped).
- What happens when an instance stays non-running after `start`—within 60 seconds of registration it must be removed from the registry and must not count as a successful long-lived instance.
- What happens when the operator passes both `--headed` and `--headless` on one `start`—Supervisor MUST exit non-zero with a clear usage error and MUST NOT register an instance.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST expose a CLI command `start` that launches one new browser automation instance navigated to a blank page.
- **FR-001d**: `start` MUST expose mutually exclusive boolean flags `--headed` and `--headless`; when both are omitted, default MUST be **headed** (documented in operator help); when both are set on one invocation, MUST exit non-zero with a usage error and MUST NOT register an instance.
- **FR-001a**: On successful `start`, the CLI MUST exit promptly after registration; it MUST NOT block the terminal for the lifetime of the instance.
- **FR-001b**: Each successfully started instance MUST be kept alive by a detached background owner process (or equivalent agent) that holds the chromedp session until `stop`, owner failure, or browser exit.
- **FR-001c**: Each concurrent instance MUST launch a dedicated Chrome browser with an isolated profile/storage directory; instances MUST NOT share a single browser process or profile.
- **FR-002**: Supervisor MUST assign each successfully started instance a unique, opaque instance ID suitable for copy/paste at the stop prompt.
- **FR-003**: Supervisor MUST persist knowledge of active instances it started so `status` and `stop` operate on the same registry for the Supervisor process lifetime (and across CLI invocations on the same machine per product conventions in planning).
- **FR-004**: Supervisor MUST expose a CLI command `status` that lists every tracked instance that is `starting` or `running`, each with its ID, lifecycle state, current CPU usage, and current memory (RAM) usage.
- **FR-004b**: For instances in `starting`, CPU and RAM MUST display as `unavailable` until the instance reaches `running`.
- **FR-004a**: For each instance, CPU and RAM shown in `status` MUST aggregate the instance owner (background chromedp helper) and that instance's dedicated Chrome process tree (including child processes); metrics for one instance MUST NOT include processes belonging to another instance.
- **FR-005**: The `status` command MUST update displayed metrics on a recurring basis until the operator exits the command (live monitoring view), not only a single static snapshot.
- **FR-005a**: Each `status` refresh MUST read the current registry under lock and MUST show adds/removals from concurrent `start`/`stop` within one refresh cycle without duplicate or stale IDs.
- **FR-006**: Supervisor MUST expose a CLI command `stop` that prompts the operator to enter an instance ID and terminates the browser process associated with that ID when it exists.
- **FR-006a**: When the operator submits empty input or cancels the stop prompt without entering an ID, `stop` MUST exit `0` and MUST NOT terminate any instance.
- **FR-007**: When stop succeeds, Supervisor MUST remove the instance from the tracked set so it no longer appears in `status`.
- **FR-007a**: When `stop` targets an ID whose processes are already gone but a stale registry entry may remain, Supervisor MUST exit `0`, report already stopped, and remove that registry entry.
- **FR-007b**: When `stop` targets an ID that is not in the registry (never valid or unknown), Supervisor MUST exit `0` with a not-found message that is clearly distinct from the already-stopped message in FR-007a.
- **FR-008**: Supervisor MUST produce operator-readable messages for start success (including ID), empty status, unknown stop ID, and start/stop failures.
- **FR-009**: Stopping or losing an instance MUST end Supervisor's ownership of that run; no orphan browser processes should remain registered under that ID.
- **FR-009a**: On reconciliation, if the instance owner is dead but Chrome for that instance is still running, Supervisor MUST remove the registry row and MUST terminate that orphan Chrome process tree (scoped to the instance profile).
- **FR-009b**: On reconciliation, if Chrome is dead but the owner is still running, Supervisor MUST remove the registry row and MUST terminate the owner process.
- **FR-010**: This feature MUST NOT require casino, game, or automation script selection at `start` (blank page only); game/casino automation attachment is out of scope for this spec slice.
- **FR-014**: v1 MUST support operator use on **Linux and WSL2** only; **macOS and Windows** are out of scope for this feature slice unless a future spec extends platform support.
- **FR-011**: Supervisor MUST NOT enforce a fixed maximum concurrent-instance count in v1; operators MAY start additional instances until the host fails to launch a browser, at which point `start` MUST exit non-zero with a clear message, MUST NOT print a new instance ID on stdout, and MUST NOT persist a registry record for that failed attempt. This applies to failures handled **synchronously in `start`** before the command reports success (including Chrome preflight, filesystem setup, and failure to start the detached worker process). A row in `starting` is allowed only after the worker process has been started successfully; if Chrome later fails in the worker before `running`, the worker MUST remove that row (FR-013/reconcile) even though the operator may already have seen the ID from a prior successful `start` exit.
- **FR-012**: The shared instance registry MUST support concurrent `start` registrations from multiple CLI invocations; each successful parallel `start` MUST yield a unique ID without global serialization of the `start` command.
- **FR-013**: If an instance has not reached a running state within **60 seconds** of registration, Supervisor MUST remove it from the tracked set, treat the start as failed for lifecycle purposes, and reconcile any orphaned owner or browser processes associated with that ID.

### Key Entities

- **Instance**: A single supervised browser automation session; attributes include unique ID, lifecycle state (`starting`, `running`, etc.), association to a background owner, a dedicated Chrome process, an isolated profile path, and OS-level process metrics (owner + browser tree) when running.
- **Instance owner**: The detached background helper process (one per instance) that holds the chromedp session and keeps the browser runnable until teardown; distinct from the Chrome browser process tree.
- **Instance registry**: The set of instances Supervisor currently tracks between start and stop (or detected exit), used by `status` and `stop`; shared across CLI invocations on the same machine.
- **Resource snapshot**: Point-in-time CPU and memory measurements for one instance in `status`, covering the instance owner and that instance's Chrome process tree combined.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator can start a new instance and obtain its ID in one command with the shell prompt returned in under 30 seconds on a typical developer workstation; when headed mode is selected, the blank page is visible in a Chrome window.
- **SC-006**: An operator can run three consecutive `start` commands from one terminal without waiting; all three instance IDs appear in `status` within 10 seconds and remain runnable until stopped.
- **SC-007**: When two `start` commands are issued concurrently from separate terminals, both complete with distinct IDs in under 30 seconds whenever the host can launch both browsers, and both appear in `status` without duplicate IDs.
- **SC-002**: With 10 concurrent instances (representative load test, not a product cap), `status` lists every instance with CPU and RAM fields refreshing at least once per 2 seconds while the command runs.
- **SC-003**: In 95% of manual test trials, stopping a valid ID removes the instance from `status` within 5 seconds and the browser window/process is no longer running.
- **SC-004**: Operators can complete the full cycle (start → observe status → stop) without editing configuration files or using unsupported interfaces.
- **SC-005**: When no instances are running, `status` communicates the empty state in plain language without appearing as an error.
- **SC-008**: Instances that never become running within 60 seconds of registration do not remain in `status` as healthy instances beyond one refresh cycle after the timeout.

## Assumptions

- Browser automation instances are backed by chromedp-driven browser processes, per project constitution and product README; this spec describes operator-facing behavior, not library layout.
- Instance IDs are generated by Supervisor and are stable for the life of that run.
- “Blank page” means a neutral empty document (e.g., `about:blank` or equivalent) with no casino or game URL loaded.
- `stop` uses an interactive prompt for the instance ID in v1; non-interactive flags may be planned later.
- Resource metrics per instance are the sum of the instance owner and that instance's dedicated Chrome tree; other instances' processes are never included in a given row.
- Each concurrent instance has its own background owner and dedicated Chrome browser with a distinct profile; owners run independently so multiple isolated chromedp sessions can be active at once without shared cookies or storage.
- Multiple CLI invocations share the same instance registry on the local machine (mechanism chosen in planning); operators expect `supervisor start` in one terminal and `supervisor status` in another to see the same instances while owners keep sessions alive.
- Authentication, credentials, and named game/casino automation are future features; this slice establishes only instance lifecycle and monitoring on a blank page.
- Default automated tests for this feature run without live casino sites; tests may use mocks or headless browser fixtures per constitution delivery workflow.
- v1 does not define a configurable concurrent-instance limit; a future spec may add one if operators need guardrails.
- Failed-start timeout is **60 seconds** from registration to running; reconciliation may run on `status`, `stop`, or background owner exit.
- Target platforms for v1 are Linux and WSL2; other OS behavior is not a requirement for this slice.
- Headed vs headless is selected per `start` via `--headed` or `--headless` (default headed); automated tests may use `--headless` while manual checks may use default or `--headed`.
