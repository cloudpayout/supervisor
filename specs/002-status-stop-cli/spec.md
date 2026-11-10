# Feature Specification: Status and Stop CLI Behavior

**Feature Branch**: `002-status-stop-cli`

**Created**: 2026-11-09

**Status**: Draft

**Input**: User description: "Modify two commands: `status` and `stop`. `status` takes control over the terminal when executed. It should now print the status of the instances in that moment. The current behavior should be used when a new flag `--follow` or `-f` is being used. `stop` takes control over the terminal to prompt the user the instance ID. It should now receive the ID directly as a parameter. The current behavior should be removed, if no parameter is passed an error should be raised."

## Clarifications

### Session 2026-11-09

- Q: When `status` runs without `--follow` on an interactive terminal, should it print the table once without clearing the screen (full-screen refresh only in follow mode)? → A: One-shot never clears the screen; follow mode keeps TTY full-screen refresh between refresh cycles (same as today's loop on a TTY).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Snapshot instance status (Priority: P1)

An operator runs a status check from a script or another terminal and needs a single answer: which instances exist right now and how they look, without blocking the shell or waiting for Ctrl+C.

**Why this priority**: Default `status` is the most common automation and scripting path; holding the terminal was the main friction.

**Independent Test**: With zero or more registered instances, run status once; verify one table (or empty message) is printed and the command exits immediately with success.

**Acceptance Scenarios**:

1. **Given** one or more active instances in the registry, **When** the operator runs status with no follow flag, **Then** the system prints one status table (ID, state, CPU, memory, optional age) reflecting the current moment, does not clear the terminal screen, and exits successfully without waiting for user interrupt.
2. **Given** no active instances, **When** the operator runs status with no follow flag, **Then** the system prints the same plain-language empty message used today and exits successfully.
3. **Given** instances in `starting` state, **When** the operator runs a one-shot status, **Then** those rows appear with CPU and memory shown as unavailable, consistent with existing status semantics.

---

### User Story 2 - Live status monitoring (Priority: P2)

An operator wants to watch instance health over time in an interactive session, as today.

**Why this priority**: Preserves the existing monitoring experience for humans at a terminal, opt-in via a flag.

**Independent Test**: Run status with follow enabled; observe repeated refreshes until interrupt; verify exit is successful after interrupt.

**Acceptance Scenarios**:

1. **Given** active instances, **When** the operator runs status with `--follow` or `-f`, **Then** on an interactive terminal the system clears and redraws the status table on the same interval as today (~2 seconds) until the operator sends interrupt (e.g. Ctrl+C), then exits successfully.
2. **Given** follow mode, **When** each refresh cycle runs, **Then** the system reads current registry state and reconciles stale rows using the same rules as the existing status command.
3. **Given** follow mode and no instances, **When** refresh cycles run, **Then** the system shows the empty-state message and continues refreshing until interrupt.

---

### User Story 3 - Stop by instance ID (Priority: P1)

An operator or script stops a run by passing the instance ID on the command line, without an interactive prompt.

**Why this priority**: Scripts and CI cannot drive stdin prompts; explicit IDs match how `start` already prints IDs.

**Independent Test**: Start an instance, run stop with that ID as an argument; verify the instance is torn down and registry updated. Run stop with no argument and verify a clear usage error.

**Acceptance Scenarios**:

1. **Given** a running instance with a known ID, **When** the operator runs stop with that ID as the sole required argument, **Then** the system reconciles the registry, stops the live instance if present, removes it from the registry on success, and exits successfully.
2. **Given** stop is invoked with no instance ID argument, **When** the command runs, **Then** the system does not read from standard input, prints a clear usage error stating that an instance ID is required, and exits with a non-zero code reserved for usage errors (same class as invalid flags today).
3. **Given** an ID that is not in the registry, **When** stop is run with that ID, **Then** the system exits successfully with the existing not-found message (distinct from already-stopped).
4. **Given** an ID that is already stopped or stale, **When** stop is run with that ID, **Then** the system exits successfully with the existing already-stopped message and prunes stale registry data as today.
5. **Given** a live instance that cannot be stopped due to an operational failure, **When** stop is run with that ID, **Then** the system exits with a general error code and does not claim success.

---

### Edge Cases

- One-shot status on a TTY: output is appended or printed in place without ANSI full-screen clear; follow on a TTY keeps full-screen clear between refreshes.
- Follow and one-shot status when the registry is temporarily locked or unreadable: the operator sees a clear error and a non-zero exit (same failure class as other registry I/O errors), not a hung command.
- One-shot status while instances are starting: rows appear immediately; metrics remain unavailable until running.
- Stop with extra unexpected positional arguments: treated as a usage error (ID must be exactly one argument).
- Stop with whitespace-only or empty string ID if the shell passes it: treated as missing/invalid ID with usage error.
- Piping or redirecting stdin no longer affects stop; behavior depends only on the ID argument.
- Documentation and help text for both commands reflects the new default behaviors and flags.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The status command MUST, by default (no follow flag), produce exactly one status output for all current instances, MUST NOT clear the terminal screen (including on a TTY), and MUST exit without entering a refresh loop or requiring interrupt.
- **FR-002**: The status command MUST accept `--follow` and `-f` as equivalent flags; when either is set, behavior MUST match the pre-change status command (periodic refresh until interrupt, TTY full-screen clear between refresh cycles, successful exit after interrupt).
- **FR-003**: One-shot and follow status MUST use the same table columns, state labels, empty-state message, metric rules (`unavailable` while starting), and reconciliation logic as the existing status command.
- **FR-004**: The stop command MUST require exactly one positional argument: the target instance ID.
- **FR-005**: The stop command MUST NOT prompt on standard input or block for operator input.
- **FR-006**: When stop is invoked without a valid instance ID argument, the system MUST fail with a usage error (non-zero exit in the usage-error class) and MUST NOT attempt termination.
- **FR-007**: When stop is invoked with a valid ID argument, reconciliation, stop, registry update, not-found, and already-stopped behaviors MUST remain as defined for the existing stop command (successful exit for not-found and already-stopped with distinct messages; general error when stop fails on a live instance).
- **FR-008**: Operator-facing help for status MUST describe default one-shot behavior and the follow flag; help for stop MUST describe the required ID argument and MUST NOT describe interactive prompting.
- **FR-009**: Integration tests that today drive stop via stdin MUST be updated to pass the ID as an argument; tests that assert follow-loop behavior MUST use the follow flag.

### Key Entities

- **Instance**: A supervised browser automation run identified by a stable ID, with lifecycle state and optional resource metrics shown in status output.
- **Status snapshot**: A point-in-time view of all instances eligible for display (starting and running, after reconciliation).
- **Stop request**: A non-interactive command identified by a single instance ID argument.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In 10 consecutive one-shot status runs with active instances, each run completes and returns control to the shell in under 5 seconds without operator interrupt.
- **SC-002**: With follow mode enabled and 10 registered instances, the operator sees at least one full refresh every 2 seconds until interrupt, matching prior monitoring expectations.
- **SC-003**: 100% of stop invocations without an ID argument fail immediately with a usage error and produce no registry change.
- **SC-004**: Operators and scripts can stop a known instance using only `stop <id>` with no stdin interaction; successful stop removes the instance from the next one-shot status output.
- **SC-005**: Help text review: zero references to interactive stop prompting; follow flag documented for status.

## Assumptions

- The refresh interval in follow mode remains approximately 2 seconds unless a separate performance spec changes it.
- Usage-error exit code remains distinct from general operational failures (aligned with existing CLI exit code contract).
- `start` continues to print instance IDs in the same format; stop accepts that same ID string.
- Breaking change to stop (removing prompt) is intentional for v1 of this feature; migration is updating scripts to pass the ID argument.
- Hidden `worker` and `start` commands are unchanged by this feature.
