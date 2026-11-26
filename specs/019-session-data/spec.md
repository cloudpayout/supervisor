# Feature Specification: Session Instance Info

**Feature Branch**: `019-session-data`

**Feature Directory**: `specs/019-session-data`

**Created**: 2026-11-26

**Status**: Draft

**Input**: User description: "a new command named as `session info <instance-id>` that will display the following information about the instance: 'casino'; 'game'; 'strategy', null by default (next specs will be cover it); 'date', when it started (date and time)". Follow-ups: "No JSON output"; `date` format `{year}-{month}-{day} {hour}:{minute}` (UTC).

This feature **adds** the operator command `supervisor session info <instance-id>`. It does not change `start`, `status`, `stop`, `screenshot`, `logs`, or `watch` except where this spec explicitly amends help text to list the new command.

## Clarifications

### Session 2026-11-26

- Q: Should successful output be JSON on standard output? → A: No. Output is plain text only: exactly four lines, one per field, in the form `label: value`. Output is not JSON and MUST NOT begin with `{` or `[`.
- Q: For the `date:` line, what exact date-and-time format should be shown? → A: Registration instant in **UTC**, formatted as `YYYY-MM-DD HH:MM` (zero-padded year, month, day, hour, and minute; one ASCII space between date and time; **colon** between hour and minute; no seconds and no timezone suffix). Example: `date: 2026-11-26 09:15`.
- Q: When `session info` fails, must standard output be completely empty? → A: Yes. On every non-zero exit (usage, not found, registry I/O), stdout MUST be empty. Operator-readable errors go to stderr only.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read session metadata for a known instance (Priority: P1)

An operator has an instance ID from `start` or `status` and needs a compact answer about what that run represents (casino, game, strategy placeholder, and when it began) without parsing the full status table or reading registry files by hand.

**Why this priority**: Session metadata is the core deliverable; without it the command has no operator value.

**Independent Test**: Start an instance with a session file for a supported casino and game entry (for example onewin Aviator on the stand-in), note the printed ID, run `supervisor session info <id>`, and verify the four labeled lines on standard output match the run’s registered casino, known game, null strategy, and start timestamp.

**Acceptance Scenarios**:

1. **Given** a registry row for instance ID `I` that is still tracked after reconciliation, **When** the operator runs `supervisor session info I`, **Then** the command exits successfully, writes exactly four lines to standard output in fixed order (`casino`, then `game`, then `strategy`, then `date`), each line using the form `label: value` with a single space after the colon, ends output with a single trailing newline, writes no JSON, writes nothing to standard error, and writes no session cookies, credentials, or file paths to standard output or standard error.
2. **Given** instance `I` was started with a session file whose casino is a supported lowercase identifier (for example `onewin`), **When** session info runs for `I`, **Then** the `casino` line shows that identifier as the value.
3. **Given** instance `I` was started with no session file (blank start), **When** session info runs for `I`, **Then** the `casino` line shows the literal value `null`.
4. **Given** instance `I` is a supported onewin run that entered the registered Aviator flow, **When** session info runs for `I`, **Then** the `game` line shows the registered game identifier `aviator` (lowercase ASCII only).
5. **Given** instance `I` is a blank start or a run where no registered game is yet associated with the row, **When** session info runs for `I`, **Then** the `game` line shows the literal value `null`.
6. **Given** any tracked instance in scope of this feature, **When** session info runs, **Then** the `strategy` line shows the literal value `null` until a future feature defines strategy values.
7. **Given** instance `I` has registry start time `T`, **When** session info runs for `I`, **Then** the `date` line shows `date: ` followed by `T` formatted in UTC as `YYYY-MM-DD HH:MM` (for example `date: 2026-11-26 09:15`).

---

### User Story 2 - Refuse invalid or missing targets (Priority: P1)

An operator or script may omit the ID, pass multiple arguments, or refer to an ID that is not in the registry after reconciliation. The command must fail clearly without emitting the four-line success output.

**Why this priority**: Scripts must distinguish usage errors from not-found outcomes, consistent with `stop`, `logs`, and `watch`.

**Independent Test**: Run session info with no ID, two IDs, a blank ID, and an ID that never existed; confirm usage or not-found failures and no success field lines on standard output.

**Acceptance Scenarios**:

1. **Given** session info is invoked with no instance ID, more than one positional ID, or a whitespace-only ID, **When** the command runs, **Then** it does not read from standard input, prints a clear usage message that an instance ID is required on standard error, fails as a usage error, leaves standard output empty, and does not print the four success field lines on standard output.
2. **Given** an instance ID that reconciliation does not list in the registry, **When** the operator runs session info for that ID, **Then** the command fails with a clear not-found message on standard error (same outcome class as `stop <id>` or `logs <id>` for unknown IDs), exits non-zero, leaves standard output empty, and does not print the four success field lines on standard output.
3. **Given** an ID that was removed from the registry after a failed background entry, **When** session info runs, **Then** the outcome is the same as scenario 2 (not found).

---

### User Story 3 - Discover the command and keep lifecycle commands unchanged (Priority: P2)

An operator learns about session info from root help and command help. Existing lifecycle commands behave as before.

**Why this priority**: CLI discoverability and non-regression of constitution-mandated start/status/stop coverage.

**Independent Test**: Run `supervisor --help` and `supervisor session --help` and `supervisor session info --help`; confirm the command is listed and documents the required ID. Re-run default integration tests for start, status, and stop without modification to operator workflows.

**Acceptance Scenarios**:

1. **Given** Supervisor is installed, **When** the operator runs root help, **Then** help lists `session` as an operator-facing command (not hidden like `worker`).
2. **Given** the operator runs `supervisor session info --help`, **When** help is shown, **Then** it describes the required instance ID, names the four output fields, states that `strategy` is always `null` in this version, and shows the `label: value` line format with an example (plain text, not JSON).
3. **Given** the default automated test suite without live casino checks, **When** it runs after this feature lands, **Then** existing start, status, stop, and screenshot scenarios that passed before still pass.

---

### Edge Cases

- Session info for a row in `starting`, `running`, `stopping`, or `stopped` state: the command succeeds if the row exists after reconciliation; lifecycle state is not included (operators use `status` for state).
- Registry I/O or lock failure: clear error on standard error, non-zero exit, standard output empty, no success field lines (same failure class as other registry-backed commands).
- `date` always reflects registration time at `start` in UTC, not last status refresh or last page navigation. Sub-minute precision from `created_at` is truncated (seconds are not shown).
- Success output is exactly four lines plus one trailing newline; no extra blank lines, headers, or banners.
- Values that are identifiers contain only lowercase ASCII letters; the literal `null` is used for absent casino, game, and for `strategy` in this version.
- The command must not log or print cookie values, session file paths, or profile directories.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST expose `supervisor session info <instance-id>` as an operator command (nested under a new top-level `session` command with `info` as its subcommand).
- **FR-002**: The command MUST require exactly one positional argument: the target instance ID. Zero or more than one positional ID MUST be a usage error with no standard-input prompt.
- **FR-003**: Before reading instance fields, the command MUST reconcile the registry using the same rules as other instance-targeting operator commands (`stop`, `logs`, `screenshot`).
- **FR-004**: When the instance ID is not found after reconciliation, the command MUST fail with a not-found outcome consistent with other instance commands, MUST write the error to standard error only, MUST leave standard output empty, and MUST NOT print the four success field lines on standard output.
- **FR-005**: On success, the command MUST print exactly four lines on standard output, in order: `casino: …`, `game: …`, `strategy: …`, `date: …`. Output MUST NOT be JSON. On success, standard error MUST be empty.
- **FR-005b**: On every non-zero exit (usage, not found, registry I/O), standard output MUST be empty; errors and usage text MUST go to standard error only.
- **FR-006**: `casino` MUST be the registered lowercase casino identifier stored on the registry row when present; MUST be the literal `null` when the row has no casino (for example blank start).
- **FR-007**: `game` MUST be the registered lowercase game identifier when Supervisor has associated one with the instance (for example `aviator` for onewin Aviator entry per feature 007); MUST be the literal `null` when no game is associated.
- **FR-008**: `strategy` MUST always show the literal `null` in this feature; assigning or displaying other strategy values is explicitly deferred to a future spec.
- **FR-009**: `date` MUST be the instance registration timestamp (`created_at` on the registry row), converted to UTC and formatted as `YYYY-MM-DD HH:MM` with zero-padded fields, one space between date and time, and a colon between hour and minute.
- **FR-010**: Root and subcommand help MUST document `session info` and MUST NOT expose internal worker-only flags.
- **FR-011**: Automated tests MUST cover success on a stand-in instance with session file, success on blank start (`null` casino and game), `strategy: null`, usage errors, and not-found, without live casino sites in the default run.

### Key Entities

- **Session info record**: Ephemeral view returned to the operator for one instance — `casino` (identifier or `null`), `game` (identifier or `null`), `strategy` (always `null` here), `date` (start instant in UTC as `YYYY-MM-DD HH:MM`).
- **Instance**: Existing supervised browser run identified by a stable ID in the registry; same ID value as `start`, `status`, `stop`, and `logs`.
- **Registered identifiers**: Casino and game names MUST obey constitution principle V (lowercase ASCII letters only when non-null).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In 100% of default test runs, `supervisor session info <valid-id>` exits `0` and standard output is exactly four `label: value` lines in the mandated order, with no JSON syntax.
- **SC-002**: In 100% of default test runs, the `strategy` line’s value is the literal `null` for every success case.
- **SC-003**: Operators can read casino, game, and start time for a running stand-in instance in one command invocation, without opening registry files or using follow-mode status; the `date` line matches the `YYYY-MM-DD HH:MM` UTC format in 100% of default test runs.
- **SC-004**: In 100% of default test runs for invalid invocations (missing ID, unknown ID), the command exits non-zero, prints an operator-readable error on standard error, standard output is completely empty, and stdout does not contain the four success field lines.
- **SC-005**: The default automated test suite completes with no regressions in start, status, stop, or screenshot behavior covered before this feature.

## Assumptions

- Operators prefer human-readable plain text (`label: value` lines) on standard output for success only, consistent with other Supervisor CLI output, not JSON; failures follow the empty-stdout / stderr-errors pattern used by `screenshot` and `watch`.
- “When it started” means Supervisor’s instance registration time at successful `start`, not worker process start skew or first navigation finish.
- The `date` line is always **UTC** even though the format omits a `Z` or offset suffix.
- Game association for display follows product registration: onewin session starts that run Aviator entry expose `aviator`; other casinos or blank starts show `null` until future specs add game fields to the registry or derivation rules.
- Strategy selection, persistence, and non-null display are **out of scope** for this feature; a later spec will define them.
- Stopped rows remain queryable with session info until reconciliation removes them, matching other ID-based commands that read the registry.
- Secrets remain out of band: session info never surfaces cookies, session file paths, or credential material.