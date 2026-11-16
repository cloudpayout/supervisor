# Feature Specification: Remove Shell Completion Command

**Feature Branch**: `fix/remove-completion-command`

**Created**: 2026-11-16

**Status**: Draft

**Input**: User description: "remove `completion` command from supervisor"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See only supported operator commands (Priority: P1)

An operator runs `supervisor` with no arguments or runs `supervisor help` to learn what Supervisor can do. The listed commands must match the product’s supported operator surface (start, status, stop, screenshot, and built-in help)—not shell integration utilities.

**Why this priority**: Help text is the primary discovery path. A `completion` entry suggests an officially supported workflow that the project has chosen not to offer.

**Independent Test**: Run the root help command; verify every line under “Available Commands” is an operator-facing command and that `completion` does not appear.

**Acceptance Scenarios**:

1. **Given** Supervisor is installed, **When** the operator runs `supervisor` with no subcommand (or `supervisor help`), **Then** the command list includes `start`, `status`, `stop`, and `screenshot`, includes `help`, and does **not** list `completion`.
2. **Given** Supervisor is installed, **When** the operator runs `supervisor --help`, **Then** the output matches the same command list rules as scenario 1.
3. **Given** other operator commands are unchanged in scope, **When** the operator runs `supervisor start --help`, `supervisor status --help`, `supervisor stop --help`, or `supervisor screenshot --help`, **Then** each command’s help still describes the same operator behavior as before this change.

---

### User Story 2 - Completion invocation is rejected (Priority: P1)

An operator or script that still calls `supervisor completion` (or a shell-specific variant) must not receive autocompletion scripts from Supervisor. The CLI must treat these invocations as unsupported, with the same class of error as other unknown top-level commands.

**Why this priority**: Removing the command from help is not enough if the subcommand still works; scripts and muscle memory could keep depending on it.

**Independent Test**: Invoke `supervisor completion`, `supervisor completion bash`, and `supervisor help completion`; verify each fails with a clear unknown-command or usage outcome and produces no shell script on standard output.

**Acceptance Scenarios**:

1. **Given** Supervisor is installed, **When** the operator runs `supervisor completion`, **Then** the command exits with a non-zero code reserved for usage or unknown commands, prints a clear message that `completion` is not a known command, and writes no autocompletion script to standard output.
2. **Given** Supervisor is installed, **When** the operator runs `supervisor completion bash` (or `zsh`, `fish`, `powershell` if previously accepted), **Then** the outcome is the same as scenario 1: no script output and a clear unsupported/unknown command result.
3. **Given** Supervisor is installed, **When** the operator runs `supervisor help completion`, **Then** Supervisor does not show completion help and instead reports that `completion` is not available (consistent with unknown top-level commands).

---

### User Story 3 - Lifecycle commands remain unaffected (Priority: P2)

An operator who only uses start, status, stop, and screenshot must see no change in behavior, exit codes, or help for those commands after `completion` is removed.

**Why this priority**: This change is CLI surface hygiene; it must not regress automation lifecycle operations covered by the constitution and prior specs.

**Independent Test**: Run the existing default integration checks for start, status, and stop (and screenshot where already covered); all pass without modification to operator workflows.

**Acceptance Scenarios**:

1. **Given** the default automated test suite for Supervisor, **When** it runs without external casino checks, **Then** all tests that passed on the prior revision still pass after this change.
2. **Given** a running instance, **When** the operator uses only `start`, `status`, `stop`, and `screenshot` as documented, **Then** outcomes match pre-change behavior for those commands.

---

### Edge Cases

- Invoking `completion` with extra arguments or unsupported shell names must not partially succeed or print a script fragment.
- Root-level help must not mention installing shell completion or redirecting operators to a `completion` subcommand.
- The internal `worker` command must remain non-operator (hidden or absent from root help as today); this feature must not expose `worker` while hiding `completion`.
- Documentation outside the binary (README, quickstarts, prior spec contracts) must not instruct operators to run `supervisor completion`; any such references are updated as part of implementation follow-through.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST NOT register a top-level `completion` command on the root CLI.
- **FR-002**: Root help (`supervisor`, `supervisor help`, `supervisor --help`) MUST list only operator-facing commands plus `help`, and MUST NOT list `completion`.
- **FR-003**: Invoking `supervisor completion` or `supervisor completion <shell>` MUST fail with a clear unknown-command or usage error and MUST NOT emit shell autocompletion scripts on standard output.
- **FR-004**: Invoking `supervisor help completion` MUST fail or report that `completion` is unavailable, consistent with other unknown top-level commands.
- **FR-005**: Behavior of `start`, `status`, `stop`, `screenshot`, and hidden `worker` MUST remain unchanged except for the absence of `completion`.
- **FR-006**: Automated checks MUST assert that root help output does not contain a `completion` command entry (in addition to any existing CLI lifecycle tests).

### Key Entities

This feature does not introduce new persisted data. It constrains the **operator CLI command catalog**: the set of named subcommands advertised and executable at the root of the `supervisor` binary.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In 100% of runs, root help output contains zero lines advertising a `completion` subcommand.
- **SC-002**: In 100% of test runs, `supervisor completion` (with and without a shell argument) exits non-zero, produces no autocompletion script on stdout, and includes an operator-readable error that the command is not supported.
- **SC-003**: The default automated test suite completes successfully with no regressions in start, status, stop, or screenshot scenarios already covered before this change.
- **SC-004**: Operators reviewing help for the first time can identify all supported actions from a single root help screen without encountering shell-setup instructions.

## Assumptions

- Shell tab completion for Supervisor is **out of scope** for the product; operators rely on explicit flags, instance IDs, and documented command syntax.
- The `completion` command visible today comes from default CLI framework behavior, not a first-class product requirement in prior feature specs.
- No migration path or deprecation period is required: `completion` was never part of the documented operator contract in project specs.
- Implementation will align operator-facing contracts (e.g. CLI contract docs in prior specs) when those documents are next updated; this spec defines the required operator behavior regardless.