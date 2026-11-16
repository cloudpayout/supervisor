# Research: Remove Shell Completion Command

**Feature**: `006-completion-command` | **Date**: 2026-11-16

## R1: Why `completion` appears today

**Decision**: Cobra v1.10.2 (project `go.mod`) auto-registers a default top-level `completion` subcommand on the root command unless explicitly disabled.

**Rationale**: Running `go run ./cmd/supervisor help` lists `completion` alongside operator commands even though `internal/cli` never calls `AddCommand` for completion. Cobra’s `initDefaultCompletionCmd` runs during command initialization when `CompletionOptions.DisableDefaultCmd` is false (default).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Leave `completion` but mark Hidden | Still executable; violates FR-003 and misleads operators who discover it via muscle memory |
| Document completion as unsupported | Does not meet FR-001/FR-002; command would still emit scripts |
| Replace with custom no-op command | Extra surface; spec requires unknown-command semantics |

## R2: Mechanism to remove the command

**Decision**: Set `root.CompletionOptions.DisableDefaultCmd = true` on the root `cobra.Command` in `internal/cli/root.go` immediately after `newRoot()` constructs the command (before `Execute()`).

**Rationale**: This is the supported Cobra API (`completions.go` documents “prevents Cobra from creating a default `completion` command”). One field, no new dependencies, no change to operator subcommand definitions.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `cobra.CompletionOptions{HiddenDefaultCmd: true}` | Hides from help but command remains (fails FR-003) |
| Remove Cobra and rewrite CLI | Violates constitution scope; massive unrelated change |
| Pin older Cobra without default completion | Fragile; disables security fixes in dependency |

## R3: Unknown-command behavior for `supervisor completion`

**Decision**: After disable, `supervisor completion` is handled like any unknown top-level subcommand: Cobra error containing `unknown command`, caught by existing `Execute()` usage-error path → message on stderr, `Usage()` printed, **exit code 2**.

**Rationale**: Matches `isUsageError` in `internal/cli/root.go` (`strings.Contains(msg, "unknown command")`). Same class as invalid flags (spec: “usage or unknown commands”).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Custom error message “completion not supported” | Unnecessary; unknown command is clear and consistent |
| Exit 1 for unknown completion | Breaks existing usage-error convention (exit 2) |

## R4: Hidden `__complete` shell hooks

**Decision**: Do not customize or remove Cobra’s internal `__complete` hidden commands unless they appear in operator help or respond to `supervisor completion`. Scope is only the default `completion` subcommand.

**Rationale**: Spec FR-005/edge cases require `worker` stay hidden; internal shell protocol commands are not operator-facing and are not listed in root help today.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Disable all Cobra completion machinery | Over-scoped; `DisableDefaultCmd` suffices for spec |

## R5: Testing strategy

**Decision**: Add `tests/integration/root_help_test.go` (or extend an existing help test) asserting: (1) root help output does not contain a `completion` command line; (2) `supervisor completion` and `supervisor completion bash` exit 2 with empty stdout; (3) optional `supervisor help completion` exits non-zero with unknown-command semantics. Run full `go test ./...` for SC-003 regression.

**Rationale**: FR-006 requires automated assertion; integration harness already builds `supervisor` binary (`supervisorBin(t)` pattern).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Manual-only quickstart | Does not satisfy FR-006 |
| Unit test of Cobra config only | Does not validate compiled binary behavior |