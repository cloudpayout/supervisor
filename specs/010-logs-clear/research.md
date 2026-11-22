# Research: Clear Instance Log Files (`logs clear`)

**Feature**: `010-logs-clear` | **Date**: 2026-11-22

## R1: Command surface (subcommand vs root)

**Decision**: Implement `supervisor logs clear` as a **Cobra child** of the existing `logs` command (`cmd.AddCommand` on `newLogsCmd()`). **Remove** the mis-scoped top-level `supervisor clear` from `internal/cli/clear.go` and `root.go` (shipped under 009).

**Rationale**: Spec and clarifications require log cleanup under the logs operator surface; registry reset was the wrong product intent. Subcommands keep `logs --recent` and view flags separate from bulk delete (FR-009).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Keep root `clear` and add `logs clear` | Conflicts with operator mental model and spec assumptions |
| `logs --clear` flag on parent | Collides with `--recent` / `--level`; FR-009 forbids |

## R2: Eligibility (active lifecycle skip)

**Decision**: Before delete, `store.View` builds a set of instance IDs where `state ∈ {starting, running, stopping}`. Glob `{LogsDir}/*.log`; for each file, `id := strings.TrimSuffix(base, ".log")`; skip delete if `id` is in the protected set (FR-001a).

**Rationale**: Matches clarified spec; basename matches 008 canonical naming. Rows in other states or absent from registry are eligible (orphans).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Delete all then recreate running logs | Violates FR-001a; races with workers |
| Protect only `running` | Contradicts clarification session 2026-11-22 |
| Parse `log_path` instead of basename | Orphans and canonical layout use `{id}.log` under `LogsDir` |

## R3: Confirmation and non-TTY

**Decision**: Reuse the pattern from `internal/cli/clear.go` (009): TTY prompt with eligible/skipped counts; `confirmYes` for `y`/`yes`; `--yes` on the **clear subcommand** for non-TTY; usage error via `fmt.Errorf("… requires --yes when stdin is not a terminal")` and existing `isUsageError` in `root.go`.

**Rationale**: Consistent operator safety across destructive commands; SC-003 alignment.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| New flag name `--force` | Spec documents `--yes` |
| Shared package for confirm | Optional; may inline or move `confirmYes` to `internal/cli/confirm.go` when deleting `clear.go` |

## R4: Enumeration and delete

**Decision**: `filepath.Glob(layout.LogsDir() + "/*.log")`. Missing directory: treat as zero candidates (spec edge case). Unreadable dir or `Glob` failure: exit `1` without deleting. On confirm, `os.Remove` each eligible path only; **no** `registry.Store.Update`.

**Rationale**: FR-002 forbids registry writes; 008 retention policy unchanged except operator-driven eligible deletes.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `RemoveAll` on logs dir | Would remove non-`.log` files if added later |
| Truncate instead of unlink | Spec says delete files |

## R5: Parent `logs` command routing

**Decision**: When `len(args)==1 && args[0]=="clear"`, parent `RunE` must not fire the "instance ID required" path. Prefer registering `clear` as a subcommand so Cobra dispatches `supervisor logs clear` to the child; parent `Use` remains `logs [instance-id]` for the default `RunE` only when no subcommand matches.

**Rationale**: Cobra dispatches subcommands before parent `RunE` when subcommand is registered (standard pattern).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Special-case `clear` in parent `RunE` | Subcommand is idiomatic and helps `logs clear --help` |

## R6: Retire 009 integration tests

**Decision**: Remove or replace `tests/integration/clear_test.go` and `internal/cli/clear_test.go` with `logs_clear` tests; remove `clear` from `TestRootHelpIncludesClear` / root help expectations.

**Rationale**: Top-level command removal; avoid false contract tests.

## R7: Testing strategy

**Decision**:

| Test | Intent |
|------|--------|
| `TestLogsClearYesOrphanAndActive` | Seed orphan + registry row `running` with logs; `logs clear --yes`; orphan gone, active kept; registry unchanged |
| `TestLogsClearNonTTYRequiresYes` | Piped stdin, no flag → exit `2`, no deletes |
| `TestLogsClearDecline` | Unit `confirmYes` (move tests to `logs_clear_test.go` or shared `confirm_test.go`) |
| `TestLogsClearHelp` | Subcommand help lists active skip + `--yes` |
| `TestRootHelpNoClear` | Root help must **not** list top-level `clear` |

Use `isolatedEnv`, `supervisorBin`, fixture registry + log files under `XDG_STATE_HOME/supervisor/logs/`.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| PTY for interactive affirm | Same as 009; `--yes` + unit parser suffice in CI |