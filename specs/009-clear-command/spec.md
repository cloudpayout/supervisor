# Feature Specification: Clear Instance Registry Command

**Feature Branch**: `009-clear-command`

**Created**: 2026-11-21

**Status**: Draft

**Input**: User description: "A new command named as `clear` to clear `instances.json` file. It should prompt for confirmation."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reset registry after troubleshooting (Priority: P1)

An operator has stopped or abandoned test runs and wants Supervisor to forget all tracked instance rows so `status` shows a clean slate. They run `clear`, read what will happen, confirm deliberately, and the instance registry becomes empty while the command reports success.

**Why this priority**: Without a supported reset path, operators may hand-edit or delete the registry file and risk corrupt JSON or torn writes. A confirmed `clear` is the safe operator workflow.

**Independent Test**: Seed the registry with one or more instance rows (or use instances left from prior tests), run `clear`, answer affirmatively at the prompt, then run a one-shot `status` and verify the empty state. Inspect the on-disk registry document and verify it is valid JSON with an empty instance list.

**Acceptance Scenarios**:

1. **Given** the registry contains one or more instance records, **When** the operator runs `clear` on an interactive terminal and answers affirmatively at the confirmation prompt, **Then** the registry document is rewritten to a valid empty registry (current schema version, zero instances), the command exits successfully, and a one-shot `status` shows the same empty state as when no instances were ever registered.
2. **Given** the registry contains instance records, **When** the operator runs `clear` and answers negatively or cancels at the prompt (including empty input interpreted as decline), **Then** the registry document is unchanged, no instance row is added or removed, and the command exits successfully without claiming the registry was cleared.
3. **Given** the registry is already empty, **When** the operator runs `clear` and answers affirmatively at the prompt, **Then** the registry remains a valid empty document and the command exits successfully with a clear message that nothing needed clearing.

---

### User Story 2 - Discover and use the command safely (Priority: P2)

An operator learns what Supervisor can do from root help and from `clear` help. The command is clearly labeled as destructive to the instance registry and explains that confirmation is required before any write.

**Why this priority**: Registry reset is irreversible from Supervisor’s perspective (instance IDs and tracking metadata are lost). Help and messaging must prevent accidental use.

**Independent Test**: Run root help and `clear` help; verify `clear` is listed among operator commands and help text describes registry reset and confirmation.

**Acceptance Scenarios**:

1. **Given** Supervisor is installed, **When** the operator runs root help (`supervisor`, `supervisor help`, or `supervisor --help`), **Then** `clear` appears in the available commands list alongside existing operator commands.
2. **Given** Supervisor is installed, **When** the operator runs `supervisor clear --help`, **Then** help describes that the command clears the instance registry file, that confirmation is required before writing, and that declining leaves the registry unchanged.

---

### User Story 3 - Refuse unattended destructive runs (Priority: P2)

An operator or script runs `clear` when standard input is not an interactive terminal (for example in CI or a pipeline). Supervisor must not clear the registry based on implicit or missing input; it must fail with a clear usage-style message unless the operator has opted into non-interactive confirmation.

**Why this priority**: Destructive registry wipes must not happen in scripts by accident when stdin is closed or piped.

**Independent Test**: Run `clear` with stdin not a TTY and without a non-interactive confirmation flag; verify non-zero exit, no registry change, and a message that explains interactive confirmation or an explicit flag is required.

**Acceptance Scenarios**:

1. **Given** the registry contains records and stdin is not a TTY, **When** the operator runs `clear` without a documented non-interactive confirmation flag, **Then** the command exits with the usage-error class code used for invalid CLI usage today, prints a clear message, and does not modify the registry.
2. **Given** stdin is not a TTY, **When** the operator runs `clear` with the documented non-interactive confirmation flag, **Then** behavior matches an affirmative interactive confirmation (registry cleared per User Story 1) without reading from stdin.

---

### Edge Cases

- Registry file missing: after affirmative confirmation, Supervisor creates or replaces the registry with a valid empty document at the same location other commands use (consistent with product state-dir conventions).
- Registry file present but unreadable or not valid JSON: command exits with a registry I/O or corruption error and does not overwrite the file with an empty document until the operator can read a valid registry or explicitly clears after fixing permissions (failure message explains the problem).
- Concurrent `start`, `stop`, or `status` while `clear` runs: registry updates use the same locking conventions as other CLI commands; one invocation wins without torn JSON (same failure/success classes as existing registry writers).
- Registry cleared while browser or worker processes from prior instances still run on the host: `clear` removes tracking rows only; it does not terminate processes. Confirmation prompt MUST state that clearing does not stop browsers and that orphan processes may remain until the operator stops them manually or via `stop` where IDs are still known.
- Operator sends interrupt during the confirmation prompt: treat as decline; registry unchanged; exit successfully without clearing.
- Parallel `clear` invocations: at most one succeeds in clearing; others see an empty registry or lock contention with clear operator messaging (no corrupt registry).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST expose a top-level CLI command `clear` that resets the instance registry document (`instances.json` in the supervisor state directory) to a valid empty registry matching the current registry schema (schema version and empty instance list).
- **FR-002**: Before writing any change to the registry, `clear` MUST prompt the operator on an interactive terminal for explicit confirmation and MUST describe that the instance registry will be wiped.
- **FR-003**: When the operator declines or cancels confirmation (including negative answer or empty input treated as decline), `clear` MUST NOT modify the registry and MUST exit successfully.
- **FR-004**: When the operator confirms on an interactive terminal, `clear` MUST write the empty registry atomically under the same locking rules as `start`, `stop`, and `status`, then exit successfully with a concise success message.
- **FR-005**: The confirmation prompt MUST include a count of instance rows currently in the registry (zero included) and MUST warn that clearing does not stop running browsers or worker processes.
- **FR-006**: When standard input is not a TTY, `clear` MUST NOT read an implicit confirmation from stdin; it MUST require a documented boolean flag (for example `--yes`) to proceed without a prompt, or MUST exit with a usage error and leave the registry unchanged.
- **FR-007**: Root help and `clear` help MUST document the command, confirmation behavior, non-interactive flag, and the warning that process cleanup is not performed by `clear`.
- **FR-008**: When the registry cannot be read due to I/O or invalid JSON, `clear` MUST exit with a non-zero error code consistent with other registry failures and MUST NOT replace the file unless the read succeeded or the file is missing and the operator confirmed (missing file treated as empty for prompting purposes only).

### Key Entities

- **Instance registry document**: The persisted JSON file listing tracked instances (`schema_version`, `instances` array). This feature only resets that document; it does not define new fields.
- **Confirmation decision**: Ephemeral operator input at the prompt (or the non-interactive flag) gating whether a registry write occurs.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In 100% of test runs where an operator confirms `clear` with a non-empty registry, a subsequent one-shot `status` shows the empty state and the on-disk registry contains zero instance rows with valid schema.
- **SC-002**: In 100% of test runs where an operator declines `clear`, the registry byte-for-byte content (or logical instance list) is unchanged from before the command.
- **SC-003**: In 100% of test runs with non-TTY stdin and without the non-interactive confirmation flag, `clear` exits with the usage-error class and leaves the registry unchanged.
- **SC-004**: Operators can complete an interactive clear (confirm and see success) in under 30 seconds when the registry is readable.

## Assumptions

- Primary use is development, test cleanup, or recovery when the operator accepts losing all tracked instance IDs in the registry.
- Clearing the registry is **not** a substitute for `stop`; operators who need processes torn down MUST run `stop` (or otherwise end processes) before or after clear as appropriate.
- Affirmative confirmation accepts common yes answers (for example `y` and `yes`, case-insensitive); anything else at the prompt is treated as decline unless planning documents a different exact set.
- The non-interactive flag name and spelling are defined in planning and CLI contracts; this spec requires only that such a flag exists for scripts when stdin is not interactive.
- Chrome profile directories and log files under the supervisor state tree are **out of scope** for this feature; only the registry document content is reset.