# Feature Specification: Instance Log Files for Operators

**Feature Branch**: `008-log-files`

**Created**: 2026-11-21

**Status**: Draft

**Input**: User description: "introduce logs files to display errors, notice, messages, and so on." (slug: `log-files`)

## Clarifications

### Session 2026-11-21

- Q: After a crash or failed run removes an instance from `status`, how should an operator locate and open that run’s log? → A: `logs <instance-id>` reads the on-disk log when the file exists (exit 0, with a clear notice that the instance is no longer tracked); plus `logs --recent` lists instance IDs that have on-disk logs but are not in the registry so operators can find IDs they did not save.
- Q: Besides the instance ID, what should each row of `logs --recent` show? → A: Instance ID, last log activity time (UTC), and the message text of the most recent ERROR line when one exists (otherwise a clear “no error recorded” placeholder for that column).
- Q: How many orphaned logs should `logs --recent` show in one run? → A: At most the 50 newest orphaned logs; if more exist on disk, print a notice stating how many additional orphaned logs were omitted.
- Q: What canonical shape should each structured log line use in the file? → A: Single-line RFC 3339 UTC timestamp, space, uppercase level (`ERROR` \| `NOTICE` \| `INFO`), space, message text; no embedded newlines in one entry.
- Q: For raw text the worker still writes today (stdout/stderr capture), what should this feature do? → A: Wrap every non-empty raw output line at ingest as a structured `INFO` line with a fresh UTC timestamp; message body is the raw line text (subject to secret redaction).
- Q: When `logs <id>` applies `--level`, what should it print for each matching line? → A: Each matching line is printed exactly as stored in the file (full canonical `timestamp LEVEL message` line).
- Q: What is the single canonical place on disk where every instance log file must live? → A: `{supervisor_state}/logs/{instance-id}.log` where `supervisor_state` is `XDG_STATE_HOME/supervisor`, or `~/.local/state/supervisor` when `XDG_STATE_HOME` is unset (same state root as `instances.json`; not under profile/data dirs).
- Q: For a tracked instance, should `logs <id>` always read the file at the canonical path from option A, even if `log_path` in the registry points somewhere else? → A: At registration, registry `log_path` MUST equal that canonical path; while tracked, `logs` opens `log_path`, and if that file is missing, falls back to the computed canonical path for the same instance ID.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read what happened for one instance (Priority: P1)

An operator needs to understand why a supervised browser run failed, stalled, or behaved unexpectedly without attaching a debugger. After `start` (or while an instance is still running), they open that instance’s log and read human-readable lines grouped by severity so errors stand out from routine notices and informational messages.

**Why this priority**: Without a single, operator-facing log channel, diagnosis depends on raw process output or guessing; this is the minimum observability upgrade beyond `status`.

**Independent Test**: Start an instance (headless in tests), trigger a known failure or status change in the worker, run the log viewer for that instance ID once, and verify labeled ERROR, NOTICE, and INFO (or equivalent) lines appear in time order with no secrets.

**Acceptance Scenarios**:

1. **Given** a successfully started instance with ID `abc`, **When** the operator requests logs for `abc` with no follow flag, **Then** the CLI prints the contents of that instance’s log file from the beginning through the current end of file, each structured line showing timestamp, severity, and message, and the command exits successfully without holding the terminal.
2. **Given** an instance whose worker recorded at least one error, **When** the operator requests logs for that ID, **Then** every error-level line is visibly marked as an error and includes enough plain-language context for the operator to distinguish it from notices and informational messages.
3. **Given** a started instance, **When** the operator runs root help, **Then** `logs` appears alongside other operator commands (`start`, `status`, `stop`, `screenshot`) with a short description of viewing instance logs.
4. **Given** an instance was removed from the registry after crash, timeout, or post-`start` failure but its log file remains on disk, **When** the operator runs `logs <instance-id>`, **Then** the full log is printed, exit status is success, and output includes a brief notice that the instance is no longer tracked (it does not appear in `status`).

---

### User Story 2 - Discover logs for runs no longer in status (Priority: P1)

An operator’s instance crashed or failed during entry and disappeared from `status` before they copied the ID from a table view. They need to find which log files exist for past runs that are no longer registered.

**Why this priority**: Crashed and failed runs are exactly when logs matter most; registry removal must not hide the only diagnostic artifact.

**Independent Test**: Start an instance, force worker exit or entry failure so the registry row is removed, run `logs --recent`, and verify that ID appears in the list; then `logs <that-id>` prints the failure log.

**Acceptance Scenarios**:

1. **Given** at least one log file on disk whose instance ID is not in the registry, **When** the operator runs `logs --recent`, **Then** each orphaned log is listed as one row with instance ID, last activity time (UTC), and the latest ERROR message when present (newest activity first), the command exits successfully, and no instance ID argument is required.
2. **Given** every on-disk log corresponds to a currently tracked instance, **When** the operator runs `logs --recent`, **Then** the command exits successfully with an empty-state message (no orphaned logs).
3. **Given** more than 50 orphaned log files exist, **When** the operator runs `logs --recent`, **Then** exactly 50 rows are printed (newest first) and a notice states how many additional orphaned logs were not shown.
4. **Given** `logs --recent` printed an ID, **When** the operator runs `logs <that-id>` without follow, **Then** they see the same log file contents as if the instance had never been tracked.

---

### User Story 3 - Watch logs while a run is in progress (Priority: P2)

An operator watches a long-running automation (for example casino entry) and wants new log lines as they are written, similar to follow mode on `status`.

**Why this priority**: Live tailing shortens time-to-diagnosis when entry phases take tens of seconds.

**Independent Test**: With a running instance, run the log viewer in follow mode; confirm new lines appear within one refresh interval after the worker appends them; interrupt exits successfully.

**Acceptance Scenarios**:

1. **Given** a running instance, **When** the operator requests logs with follow enabled (`--follow` or `-f`), **Then** the CLI prints existing lines then continues printing new lines as they are appended until the operator interrupts, then exits successfully.
2. **Given** follow mode and an instance that stops while the operator is watching, **When** the log file reaches end of file and no further lines are appended, **Then** the command keeps waiting for new lines until interrupt (it does not exit solely because the instance stopped).

---

### User Story 4 - Focus on errors or notices (Priority: P3)

An operator debugging a failure wants to hide noisy informational lines and see only errors, or only errors and notices.

**Why this priority**: Reduces scroll fatigue on large logs; optional filter preserves full log as default.

**Independent Test**: Seed a log file with mixed severities; run with `--level error` and verify only error lines print; run with `--level notice` and verify error and notice lines print. (Severity filter applies to User Story 4 / `logs <id>` only, not to `--recent` row selection.)

**Acceptance Scenarios**:

1. **Given** a log file containing ERROR, NOTICE, and INFO lines, **When** the operator requests logs with a minimum severity of `error`, **Then** only error-level lines are printed, each line identical to the on-disk text, and notices and informational messages are omitted.
2. **Given** the same log file, **When** the operator requests logs with a minimum severity of `notice`, **Then** error and notice lines are printed and informational messages are omitted.
3. **Given** a filter that matches no lines, **When** the operator runs a one-shot log request, **Then** the command exits successfully with no lines (or a brief empty hint), not an error.

---

### Edge Cases

- Missing instance ID when neither `--recent` nor another discovery flag is set: usage error, non-zero exit, no stdin prompt (consistent with `stop`).
- Unknown instance ID with no log file on disk: clear not-found message and non-zero exit (registry row alone is not enough without a readable log file).
- Instance removed from registry after crash, reconciliation, `stop`, or post-`start` failure: log file MUST remain readable by ID via on-disk fallback; `logs <id>` exits successfully and states the instance is not tracked.
- `logs --recent` with a positional instance ID: usage error (mutually exclusive modes).
- Orphaned log with no ERROR lines in the file: `logs --recent` row still appears with last activity time and a placeholder in the error-summary column (not a blank row).
- Latest ERROR summary in `logs --recent` MUST be truncated for terminal width if the message is very long; full text remains in `logs <id>`.
- More than 50 orphaned logs: only the 50 most recently active are listed; omitted count is reported in a trailing notice line.
- Reconciliation removes registry rows but MUST NOT delete the instance log file as part of normal teardown (profile removal remains separate).
- Lines that do not match the structured pattern (e.g. log files written before this feature): passed through to `logs` output unchanged and excluded from `--level` filtering.
- Blank lines from raw stdout/stderr capture MUST NOT be written (ingest skips them).
- After this feature ships, newly written instance logs MUST contain only structured lines for all worker-captured stdout/stderr and for Supervisor-authored diagnostics.
- Structured log messages MUST NOT contain newline characters; multi-line data MUST be flattened or split into multiple structured lines.
- Severity filter `--level` applies only to structured lines matching the canonical pattern; unstructured lines appear only when no severity filter is set (or when a future “include raw” flag is added—out of scope for v1).
- Log file not yet created (instance still registering): clear message that the log is not available yet; exit non-zero.
- Secrets (session cookies, tokens, passwords): MUST never appear in log output; redaction rules align with prior features that forbid secrets in logs and registry.
- Very large logs: one-shot and follow modes MUST stream without loading the entire file into memory for display.
- Concurrent `logs --follow` and worker append: no duplicate lines; no lost lines at line boundaries under normal append patterns.
- Two operators following the same instance: both receive the same lines independently.
- Instance ID in a log filename MUST be a single safe path segment; log paths MUST NOT be constructed outside `{supervisor_state}/logs/`.
- Multiple CLI invocations on one machine with the same `XDG_STATE_HOME` (or default state home) MUST resolve the same log directory for a given instance ID.
- Registry row with `log_path` not equal to the canonical path for that ID: out of scope for operator repair in v1; `logs` still attempts registry path then canonical fallback per FR-009.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST maintain one append-only log file per browser instance, created when the instance is registered, at the **canonical log path** `{supervisor_state}/logs/{instance-id}.log`, where `{instance-id}` is the opaque instance ID and `{supervisor_state}` follows the XDG state layout in FR-001a.
- **FR-001a**: `{supervisor_state}` MUST be the `supervisor` subdirectory immediately under the host state home: `XDG_STATE_HOME` when set, otherwise `~/.local/state`. Chrome profiles and other mutable browser data MUST NOT share this directory; they remain under the separate data root from feature `001-supervisor-cli`.
- **FR-001b**: `logs --recent` MUST discover orphaned logs only by scanning `{supervisor_state}/logs/*.log`; it MUST NOT search profile directories or other paths.
- **FR-002**: Worker and Supervisor-owned automation code MUST write operator-relevant events as **structured log lines**, one per line, using this canonical pattern: `{RFC3339_UTC_timestamp} {LEVEL} {message}` where `{LEVEL}` is exactly `ERROR`, `NOTICE`, or `INFO` (uppercase), fields are separated by a single ASCII space, the timestamp uses UTC with `Z` suffix (fractional seconds permitted), and `{message}` is a single line with no embedded newline characters.
- **FR-002a**: The `logs` command, `--level` filter, and `--recent` ERROR summary MUST recognize structured lines by this pattern only; lines that do not match MUST be treated as unstructured legacy content.
- **FR-002b**: Worker log ingest MUST line-buffer stdout and stderr from the worker process and emit each non-empty raw line as one structured line with level `INFO`, ingest-time UTC timestamp, and message text equal to the raw line (after secret redaction). Explicit `ERROR` and `NOTICE` lines from Supervisor automation code MUST still use those levels directly and MUST NOT be downgraded to `INFO`.
- **FR-003**: ERROR lines MUST denote failures or abnormal termination paths the operator should act on. NOTICE lines MUST denote phase changes, reconciliation outcomes, or other salient state transitions that are not failures. INFO lines MUST denote routine progress or diagnostic detail suitable for full logs.
- **FR-004**: Supervisor MUST expose an operator CLI command `logs` that accepts either exactly one positional argument (instance ID) or `--recent` with no positional ID (mutually exclusive).
- **FR-005**: By default, `logs <instance-id>` MUST print the full log once to standard output and exit without requiring interactive input.
- **FR-006**: `logs` MUST support `--follow` / `-f` to tail new lines until interrupted, mirroring the follow semantics of `status` (successful exit on interrupt).
- **FR-007**: `logs` MUST support an optional minimum-severity filter (at least `error` and `notice` thresholds as in User Story 4). When omitted, all severities MUST be shown. When set, each emitted line MUST be the exact structured line from the file (no reformatting, no dropped timestamp or level).
- **FR-008**: When no instance ID is given and `--recent` is not set (or equivalent discovery flag is false), `logs` MUST exit with a usage error and MUST NOT read from standard input.
- **FR-009**: For `logs <instance-id>`, when the instance is tracked, Supervisor MUST open the registry `log_path`, which MUST equal the canonical path from FR-001 at registration time. If that path is not readable but the canonical `{supervisor_state}/logs/{instance-id}.log` exists, MUST fall back to the canonical file. When the instance is not tracked, MUST use only the canonical path. If a readable log file is found, MUST print it and exit successfully; if the instance is not tracked, MUST print a one-line notice that it is no longer in `status`.
- **FR-009a**: `logs --recent` MUST list each instance ID that has a log file under the standard logs directory and is **not** presently in the registry, ordered by most recent log activity first (e.g. file modification time), **up to 50 rows**. Each row MUST include: instance ID, last log activity timestamp in UTC, and the message portion of the most recent ERROR log entry in that file when one exists; if none exists, MUST show an explicit placeholder in the error-summary field (e.g. “no error recorded”). When more than 50 orphaned logs exist, MUST print a notice after the table stating how many additional orphaned logs were omitted.
- **FR-009c**: The error summary on `--recent` rows MUST be derived only from structured ERROR lines (or equivalent severity marker); it MUST NOT include secrets and MAY be truncated for display while the full line remains available via `logs <id>`.
- **FR-009b**: Lifecycle teardown (crash, failed start timeout, entry failure, `stop`, reconciliation) MUST NOT delete the instance log file; only the operator or out-of-scope retention policy may remove log files.
- **FR-010**: Session secrets, credentials, and cookie values MUST NOT be written to log lines; existing secret-handling obligations from casino automation features apply unchanged.
- **FR-011**: One-shot `status` MUST NOT be required to change for this feature; log viewing is exclusively via `logs` (planning MAY add log path hints to help text without changing default status columns).
- **FR-012**: Automated tests MUST cover: successful one-shot read, follow mode interruption, severity filter respecting the canonical line pattern, missing ID usage error, orphaned log read after registry removal (crash or failure path), `logs --recent` listing, parsing of structured ERROR lines for summaries, and absence of secrets in displayed log content for a representative start/worker flow.

### Key Entities *(include if feature involves data)*

- **Supervisor state root**: Host-specific directory holding `instances.json`, `logs/`, and runtime sockets; derived from `XDG_STATE_HOME` per FR-001a.
- **Instance log file**: Append-only file at `{supervisor_state}/logs/{instance-id}.log`; contains ordered log entries until removed by operator or retention policy outside this feature.
- **Structured log entry**: One line matching `{RFC3339Z} ERROR|NOTICE|INFO message`. **Unstructured line**: Any line in a pre-feature log file (or corrupt data) that does not match the pattern.
- **Log view request**: Operator invocation by instance ID and options (follow, minimum severity), or a discovery request (`--recent`) with no instance ID.
- **Orphaned log**: On-disk log whose instance ID is absent from the registry (typical after crash, failure, or `stop`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In automated scenarios, 100% of injected ERROR-level events appear in the log viewer output with error severity visible within the same one-shot command after the event is written.
- **SC-002**: Operators can retrieve a complete log for a test instance in one command without follow mode, in under 5 seconds for logs up to 10,000 lines in integration test environments.
- **SC-003**: Follow mode delivers appended lines to the operator within 3 seconds of the worker writing them in integration tests.
- **SC-004**: 100% of secret-redaction tests that apply to worker logs continue to pass; no credential material appears in log viewer output.
- **SC-005**: Documented operator help lists `logs` and describes instance ID, `--recent`, follow, and severity filter in plain language.
- **SC-006**: In integration tests where the registry row is removed after a worker failure, 100% of runs still allow `logs --recent` to surface that ID with matching last-ERROR summary text and `logs <id>` to return the failure log with success exit status.

## Assumptions

- Per-instance log files live only under `{supervisor_state}/logs/` per FR-001a; registry `log_path` from feature `001-supervisor-cli` records that same canonical path at registration.
- Log files are retained after `stop`, crash, or failed runs until the operator deletes state manually; automatic log deletion or rotation is out of scope for v1 unless planning adds it for disk safety.
- Linux and WSL2 remain the supported operator platforms; log file paths follow XDG state conventions already used by Supervisor.
- DEBUG or trace-level verbosity is out of scope for the operator CLI; developers rely on test hooks or future flags if needed.
- Log files created before this feature may be entirely unstructured; operators use unfiltered `logs <id>` for those files. New runs after this feature use structured-only ingest per FR-002b.

## Dependencies

- Existing instance lifecycle (`start`, `stop`, registry, worker process) and state directory layout from feature `001-supervisor-cli`.
- Secret non-disclosure rules from features `004-open-onewin-page` and `007-enter-aviator-onewin` apply to all log content.