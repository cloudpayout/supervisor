# Feature Specification: Clear Instance Log Files

**Feature Branch**: `010-logs-clear`

**Created**: 2026-11-22

**Status**: Draft

**Input**: User description: "Corrected: the new command will be `logs clear`, not `clear`; it won't clear `instances.json`, it will clear all logs files in `{XDG_STATE_HOME}/supervisor/logs/*.log`."

## Clarifications

### Session 2026-11-22

- Q: Should `logs clear` delete log files for instances that are currently running? → A: No; log files for **running** instances MUST NOT be removed.
- Q: Should `logs clear` also keep log files for instances in `starting` or `stopping` registry state? → A: Yes; protect **`running`**, **`starting`**, and **`stopping`** the same way (skip deletion for all three).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reclaim disk after troubleshooting (Priority: P1)

An operator has finished diagnosing failed or abandoned runs and wants to remove accumulated instance log files without touching the instance registry, browser profiles, or logs for runs that are still on the lifecycle (`starting`, `running`, or `stopping`). They run `supervisor logs clear`, review what will be deleted and what will be skipped, confirm, and eligible log files are removed while active instances keep their logs.

**Why this priority**: Log files can grow large across many test runs; operators need a supported wipe path instead of manually deleting files and risking the wrong directory or cutting off live diagnostics.

**Independent Test**: Place one log file for an instance in registry state **running** (or **starting**) and one for an orphaned (untracked) instance; run `logs clear`, confirm; verify the orphaned log is gone, the active instance’s log remains, and `instances.json` is unchanged.

**Acceptance Scenarios**:

1. **Given** log files exist for both an orphaned instance and a tracked instance in an active lifecycle state (`starting`, `running`, or `stopping`), **When** the operator runs `logs clear` on an interactive terminal and answers affirmatively, **Then** the orphaned (and other eligible) log files are deleted, each active instance’s log file is **not** deleted, the command exits successfully, the registry is unchanged, and a summary distinguishes deleted vs skipped counts.
2. **Given** log files exist, **When** the operator runs `logs clear` and answers negatively or cancels (including empty input treated as decline), **Then** every log file remains on disk, the registry is unchanged, and the command exits successfully without claiming logs were cleared.
3. **Given** only log files tied to active lifecycle instances (no eligible files), **When** the operator confirms `logs clear`, **Then** no log file is deleted, the command exits successfully with a clear message that nothing eligible needed clearing, and the registry is unchanged.

---

### User Story 2 - Discover the command safely (Priority: P2)

An operator explores what the `logs` command can do via `supervisor logs --help` and `supervisor logs clear --help`. The clear subcommand is described as destructive to eligible on-disk log files, requires confirmation, skips active lifecycle instances, and does not modify the registry.

**Why this priority**: Bulk log deletion is irreversible; discovery text must prevent confusion with viewing logs (`logs <id>`) or listing orphans (`logs --recent`).

**Independent Test**: Run `logs --help` and `logs clear --help`; verify `clear` is listed and help describes selective file deletion, active-instance skip, confirmation, and registry safety.

**Acceptance Scenarios**:

1. **Given** Supervisor is installed, **When** the operator runs `supervisor logs --help`, **Then** `clear` appears as a subcommand of `logs` with a short description of removing eligible log files.
2. **Given** Supervisor is installed, **When** the operator runs `supervisor logs clear --help`, **Then** help states that eligible `*.log` files under the supervisor logs directory are removed after confirmation, that logs for instances in `starting`, `running`, or `stopping` are never removed, that the registry is not modified, and how non-interactive use works.

---

### User Story 3 - Refuse unattended bulk deletion (Priority: P2)

A script or CI job must not accidentally delete all log files when standard input is not an interactive terminal. The operator must pass an explicit non-interactive confirmation flag to proceed without a prompt.

**Why this priority**: Same safety model as other destructive Supervisor CLI actions; piped input must not count as confirmation.

**Independent Test**: With eligible and active-instance log files present, run `logs clear` with non-TTY stdin and without the non-interactive flag; verify usage-class failure and no deleted files.

**Acceptance Scenarios**:

1. **Given** log files exist and stdin is not a TTY, **When** the operator runs `logs clear` without the non-interactive confirmation flag, **Then** the command exits with the usage-error class, prints a clear message, and leaves all log files in place.
2. **Given** stdin is not a TTY, **When** the operator runs `logs clear` with the non-interactive confirmation flag, **Then** eligible log files are deleted and active-instance logs are skipped, matching User Story 1, without reading stdin.

---

### Edge Cases

- Logs directory missing: treated as zero eligible files for prompting; after confirmation, exit successfully with nothing-to-clear semantics.
- Logs directory exists but is not readable: exit with a general error, no partial deletion without successful enumeration.
- Only non-`.log` files in the logs directory: left untouched; only top-level `*.log` files are candidates.
- Log filename `{instance-id}.log` where `instance-id` is in the registry with state **`starting`**, **`running`**, or **`stopping`**: MUST NOT be deleted even if other files are cleared.
- Log file with no matching registry row (orphan), or for a tracked instance in **`stopped`** with no active lifecycle (if such a row appears transiently): eligible for deletion.
- `logs clear` with positional instance ID or with `--recent` / `--follow` / `--level`: usage error; clear is a separate subcommand, not a flag combination on view modes.
- Registry rows and `log_path` fields: unchanged; `logs <id>` for an active instance continues to read the preserved log after clear removed other files.
- Concurrent `logs clear` while a worker appends to an active instance log: skip that file; deletion of other files may proceed; preserved log remains valid for `logs --follow`.
- Concurrent `logs clear` invocations: no corrupt state; skipped active logs remain consistent across both invocations.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST expose a `logs` subcommand `clear` (invoked as `supervisor logs clear`) that deletes eligible regular files matching `*.log` in `{supervisor_state}/logs/`, where `{supervisor_state}` is the supervisor state directory under the host state home (`XDG_STATE_HOME/supervisor` when set, otherwise `~/.local/state/supervisor`).
- **FR-001a**: A log file is **ineligible** for deletion when its basename (without `.log`) equals the instance ID of a registry row whose `state` is **`starting`**, **`running`**, or **`stopping`**; such files MUST be skipped and MUST remain on disk after `logs clear` completes.
- **FR-002**: `logs clear` MUST NOT modify the instance registry (`instances.json`) or any Chrome profile or data directories.
- **FR-003**: Before deleting any log file, `logs clear` MUST prompt the operator on an interactive terminal for explicit confirmation and MUST state how many eligible log files will be removed and that logs for active lifecycle instances (`starting`, `running`, `stopping`) will be kept.
- **FR-004**: When the operator declines or cancels confirmation (including negative answer or empty input treated as decline), `logs clear` MUST NOT delete any log file and MUST exit successfully.
- **FR-005**: When the operator confirms, `logs clear` MUST delete every **eligible** in-scope log file, MUST skip every ineligible log file per FR-001a, then exit successfully with a concise summary (counts deleted, skipped, and nothing-to-clear when applicable).
- **FR-006**: The confirmation prompt MUST include the number of eligible `*.log` files to delete and the number of active-instance logs that will be skipped (each may be zero).
- **FR-007**: When standard input is not a TTY, `logs clear` MUST NOT read implicit confirmation from stdin; it MUST require a documented boolean flag (for example `--yes`) to proceed without a prompt, or MUST exit with a usage error and leave log files unchanged.
- **FR-008**: `logs --help` and `logs clear --help` MUST document the subcommand, confirmation, non-interactive flag, active lifecycle skip rule (FR-001a), registry safety, and that only eligible top-level `*.log` files in the canonical logs directory are affected.
- **FR-009**: Invoking `logs clear` with arguments or with parent `logs` view flags (`--recent`, `--follow`, `--level`) MUST produce a usage error and MUST NOT delete files.

### Key Entities

- **Instance log file**: One append-only diagnostic file per instance at `{supervisor_state}/logs/{instance-id}.log` (per feature 008); deleted only when eligible for clear (not tied to a registry instance in `starting`, `running`, or `stopping`).
- **Logs directory**: `{supervisor_state}/logs/` containing zero or more `*.log` files.
- **Confirmation decision**: Ephemeral operator input or non-interactive flag gating deletion.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In 100% of test runs where an operator confirms `logs clear` with at least one eligible log file present, every eligible `*.log` is removed, every log file for an instance in `starting`, `running`, or `stopping` remains, and `instances.json` is unchanged.
- **SC-002**: In 100% of test runs where an operator declines `logs clear`, every log file that existed before the command is still present.
- **SC-003**: In 100% of test runs with non-TTY stdin and without the non-interactive confirmation flag, `logs clear` exits with the usage-error class and leaves all log files unchanged.
- **SC-004**: Operators can complete an interactive clear (confirm and see success) in under 30 seconds when the logs directory is readable.

## Assumptions

- This feature corrects an earlier mis-scoped top-level `clear` idea (registry reset); operator intent here is **log file** cleanup only. Retirement or replacement of any mistakenly shipped top-level `clear` command is handled in planning, not in this spec.
- Affirmative confirmation accepts `y` and `yes` case-insensitively; any other response at the prompt is decline.
- Deletion is permanent; there is no archive or trash step in v1.
- Registry states **`starting`**, **`running`**, and **`stopping`** protect the matching log file; orphaned logs and logs without an active registry row are eligible.
- Feature 008 rules that reconciliation must not delete logs automatically remain unchanged; only explicit `logs clear` removes eligible files.
- `logs clear` does not stop workers or change instance state; it only removes eligible files on disk.