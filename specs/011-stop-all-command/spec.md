# Feature Specification: Stop All Instances Command

**Feature Branch**: `011-stop-all-command`

**Created**: 2026-11-22

**Status**: Draft

**Input**: User description: "a new command `stop all` to stop all running instances"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Tear down every tracked instance (Priority: P1)

An operator or CI job has several supervised browser instances still registered and wants to shut them all down without copying each ID from a status snapshot. They run stop all once and expect every instance that still has live work to stop the same way as a single-ID stop, with a concise summary.

**Why this priority**: Manual one-by-one stopping does not scale in scripts, multi-instance tests, or end-of-day cleanup. This is the main operator value.

**Independent Test**: Start two instances with distinct IDs, run stop all, then run a one-shot status and verify the empty state and that neither browser session remains supervised.

**Acceptance Scenarios**:

1. **Given** the registry lists two or more instances that are eligible to stop (for example `starting` or `running`), **When** the operator runs `supervisor stop all`, **Then** the system reconciles the registry, stops each remaining live instance using the same termination and registry-removal rules as `stop <instance-id>`, reports each stopped ID or already-stopped/not-found outcome on standard error, prints a one-line summary of how many instances were stopped versus skipped as already gone, and exits successfully when every attempted live stop succeeds.
2. **Given** one instance is running and another was already removed by a prior `stop <id>`, **When** the operator runs `supervisor stop all`, **Then** the live instance is stopped and removed from the registry, the absent ID is not treated as a usage error, and the command exits successfully.
3. **Given** `stop all` is invoked from a script, **When** the command runs, **Then** it does not read from standard input and completes without an interactive prompt.

---

### User Story 2 - No-op when nothing is tracked (Priority: P2)

An operator runs stop all after already stopping everything (or on a fresh state directory) and wants immediate success without an error.

**Why this priority**: Idempotent teardown in scripts should not require a prior status check.

**Independent Test**: With an empty registry (or zero rows after reconciliation), run `supervisor stop all`; verify exit success, a clear empty message aligned with one-shot status empty wording, and no registry change.

**Acceptance Scenarios**:

1. **Given** reconciliation leaves zero instance rows in the registry, **When** the operator runs `supervisor stop all`, **Then** the system prints a plain-language message that no instances are running, makes no registry writes, and exits successfully.
2. **Given** rows exist but every row is already stale (processes gone, pruned during reconciliation), **When** the operator runs `supervisor stop all`, **Then** the command exits successfully, reports already-stopped outcomes as appropriate, and leaves the registry with zero active rows (consistent with single-ID stop semantics).

---

### User Story 3 - Discover the command and preserve single-ID stop (Priority: P2)

An operator learns bulk stop from help text. Existing `supervisor stop <instance-id>` behavior remains unchanged.

**Why this priority**: Prevents confusion between bulk and targeted stop and avoids breaking scripts that pass opaque IDs.

**Independent Test**: Run `supervisor stop --help`; confirm documentation for `<instance-id>` and for `all`. Run `supervisor stop <id>` on a running instance and confirm behavior matches the 002 stop contract.

**Acceptance Scenarios**:

1. **Given** Supervisor is installed, **When** the operator runs `supervisor stop --help`, **Then** help documents stopping one instance by ID and stopping every tracked instance via `stop all`, and does not describe interactive prompting.
2. **Given** a running instance with a known ID, **When** the operator runs `supervisor stop <instance-id>`, **Then** behavior is unchanged from the status-and-stop CLI feature (required single ID argument, same messages and exit codes for not-found, already-stopped, and live failure).
3. **Given** an operator passes `all` as the sole argument to stop, **When** the command runs, **Then** it executes bulk stop and is not interpreted as a literal instance ID lookup first.

---

### Edge Cases

- One live instance fails to terminate while others succeed: the system continues attempting the remaining instances, reports which IDs failed, and exits with the same non-zero general-error class used when a single-ID live stop fails.
- Concurrent `start` or single-ID `stop` while `stop all` runs: registry updates follow existing locking conventions; no duplicate IDs and no torn registry; each instance is stopped at most once.
- Instance in `starting` state: included in bulk stop the same as `running` (tear down owner/browser and remove registry row when stop succeeds).
- Operator sends interrupt during bulk stop: the system stops scheduling further instances, leaves the registry consistent with instances already fully stopped (rows removed as in single-ID stop), reports how many completes before interrupt, and exits with a non-zero code indicating the command did not finish the full set.
- Reserved word `all`: only honored as the sole argument to `stop` for bulk teardown; opaque instance IDs generated by Supervisor are not the literal string `all` (see Assumptions).
- More than one positional argument to `stop` (for example `supervisor stop <id1> <id2>` or `supervisor stop all extra`): Cobra usage error exit `2`; no stop runs. The shell command `supervisor stop all` passes exactly one argument (`all`) to `stop`.
- `stop` with no arguments: unchanged usage error from 002 (instance ID required).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST accept `supervisor stop all` as an operator-facing invocation that stops every instance row remaining in the registry after reconciliation, using the same per-instance stop, profile cleanup, and registry-removal semantics as `supervisor stop <instance-id>`.
- **FR-002**: `stop all` MUST NOT prompt on standard input and MUST be suitable for non-interactive scripts.
- **FR-003**: When the registry has zero instance rows after reconciliation, `stop all` MUST print a message equivalent in meaning to the one-shot status empty state (`No instances are running.`), MUST NOT modify the registry, and MUST exit successfully.
- **FR-004**: For each instance targeted by `stop all`, outcomes MUST match single-ID stop: successful live stop removes the row and reports `stopped instance <id>`; unknown ID after reconciliation is skipped without usage error; already-stopped rows report the existing already-stopped wording and are pruned as today.
- **FR-005**: When at least one snapshotted instance hits an **operational failure**, `stop all` MUST still attempt the remaining snapshot IDs, MUST report each failure distinctly on standard error, and MUST exit with the general-error class (`1`) if any operational failure occurred (even when other IDs succeeded). **Operational failure** means `stopInstance` returns a non-nil error to the bulk loop—the same conditions that make single-ID `stop` exit `1` today (for example registry `store.Update` failure after a live terminate attempt), excluding not-found, already-stopped, and user interrupt.
- **FR-006**: When every snapshotted ID ends in live stop, already-stopped, or not-found (no operational failure) and the command was not interrupted, `stop all` MUST exit successfully.
- **FR-007**: `stop all` MUST print a one-line summary after processing that states how many instances were stopped and how many were already absent or already stopped (counts may be zero).
- **FR-008**: `supervisor stop <instance-id>` with any ID other than the reserved bulk token MUST remain unchanged from the 002 stop contract (exactly one argument, no stdin).
- **FR-009**: The reserved bulk token MUST be exactly `all` (case-sensitive per existing CLI conventions unless planning documents otherwise).
- **FR-010**: Operator help for `stop` MUST document both single-ID and `all` forms and MUST NOT claim that `all` is an instance ID format.
- **FR-011**: Integration and quickstart documentation that today say “stop all” informally MUST be updated to name `supervisor stop all` as the supported bulk teardown command.

### Key Entities

- **Bulk stop request**: A non-interactive `stop` invocation whose sole argument is the reserved token `all`.
- **Stop target set**: The ordered list of instance IDs read from the registry immediately after reconciliation at the start of `stop all` (snapshotted so mid-run registry changes from other terminals do not skip or double-stop IDs incorrectly).
- **Per-instance stop outcome**: The same success, not-found, already-stopped, or operational-failure classification used by single-ID stop, aggregated into the bulk summary.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With five concurrently registered running instances, one `stop all` invocation removes all five from the next one-shot status output within 60 seconds on a typical development host, without operator input.
- **SC-002**: With zero registered instances, `stop all` completes and returns control to the shell in under 3 seconds with a success exit code.
- **SC-003**: In 10 consecutive bulk-stop runs against an empty registry, 100% exit successfully with the empty-state message and no registry modification.
- **SC-004**: Scripts can replace N sequential `stop <id>` calls after status parsing with a single `stop all` in documented quickstart teardown flows.
- **SC-005**: Help text for `stop` mentions both single-ID and `all` forms with zero references to interactive stop prompting.

## Assumptions

- Instance IDs remain opaque, URL-safe strings that are not assigned the literal value `all`; bulk stop therefore does not collide with a real ID.
- Single-ID stop semantics, exit codes (including usage error for missing argument), reconciliation, profile removal, and log file retention rules from prior features remain the authority for each instance during bulk stop.
- `stop all` is the supported bulk lifecycle teardown: it stops processes and removes instance rows together. Supervisor no longer exposes a top-level `clear` command (removed in the logs-clear feature); `logs clear` only deletes eligible log files and does not modify the instance registry or stop browsers.
- Order of stopping among IDs is not operator-visible; only outcomes and counts matter for v1.
- Linux and WSL2 remain the supported operator environments (unchanged from 001).