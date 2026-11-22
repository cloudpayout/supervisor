# Implementation Plan: Clear Instance Log Files

**Branch**: `010-logs-clear` | **Date**: 2026-11-22 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/010-logs-clear/spec.md`

## Summary

Add `supervisor logs clear` to delete **eligible** `*.log` files under `{supervisor_state}/logs/`, skipping logs for instances in `starting`, `running`, or `stopping`, with TTY confirmation and `--yes` for scripts. Do not touch `instances.json`. **Remove** the mis-scoped top-level `supervisor clear` (registry wipe from 009). Implementation: Cobra subcommand on `logs`, eligibility via registry `View` + glob, reuse confirmation patterns from `internal/cli/clear.go` then delete that file.

## Technical Context

**Language/Version**: Go 1.27.1 (`go.mod`, constitution)

**Primary Dependencies**: `github.com/spf13/cobra` (existing); stdlib `bufio`, `fmt`, `os`, `path/filepath`, `strings`

**Storage**: Read-only `registry.Store.View`; filesystem deletes under `registry.Layout.LogsDir()`

**Testing**: `go test ./...`; replace `tests/integration/clear_test.go` with `logs_clear_test.go`; unit tests for `confirmYes` (relocate from `clear_test.go`)

**Target Platform**: Linux and WSL2 only

**Project Type**: CLI (`cmd/supervisor`)

**Performance Goals**: Single registry view + glob + N removes; SC-004 under 30s interactive

**Constraints**: Default tests without live casino; protect active lifecycle logs (clarification 2026-11-22)

**Scale/Scope**: New `internal/cli/logs_clear.go` (~150 LOC); modify `logs.go`; delete `clear.go`; update root help tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec Before Implementation | Pass | Spec `010-logs-clear` + clarifications |
| II. CLI as the Operator Interface | Pass | Extends `logs`; removes wrong root command |
| III. Multiple Browser Instances | Pass | Preserves active instance logs |
| IV. Supervisor Owns the Run | Pass | No lifecycle change; does not stop instances |
| V. Named Casinos and Games Only | Pass | No catalog change |
| Technical: Go, chromedp, secrets | Pass | No secrets in delete path |
| Delivery: tests before complete | Pass | Integration tests per research R7 |

**Gate result**: Pass.

**Post–Phase 1 re-check**: Pass. Contract v10 documents selective delete and 009 retirement.

## Project Structure

### Documentation (this feature)

```text
specs/010-logs-clear/
├── plan.md
├── spec.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
internal/cli/logs.go           # AddCommand(logs clear); update Long to mention clear
internal/cli/logs_clear.go     # newLogsClearCmd, runLogsClear, eligibility
internal/cli/confirm.go        # optional: confirmYes shared (moved from clear.go)
internal/cli/root.go           # Remove AddCommand(newClearCmd)
internal/cli/clear.go          # DELETE (009)
internal/cli/clear_test.go     # DELETE or merge → confirm_test.go
tests/integration/logs_clear_test.go   # new
tests/integration/clear_test.go        # DELETE
tests/integration/root_help_test.go    # Remove clear from expectations; drop TestRootHelpIncludesClear if merged
```

**Structure Decision**: Single Go module; logs subtree owns bulk delete.

## Design

### Eligibility algorithm

```text
layout ← registry.Resolve()
store.View → protectedIDs = { inst.ID | inst.State ∈ {starting,running,stopping} }
paths ← Glob(layout.LogsDir() + "/*.log")
for each path:
  id ← TrimSuffix(Base(path), ".log")
  if id ∈ protectedIDs → skipped++
  else → eligible++
```

### `newLogsClearCmd()` (`internal/cli/logs_clear.go`)

- `Use: "clear"`, parent `logs` → `supervisor logs clear`
- Flag `--yes`
- `RunE`: compute counts → non-TTY guard → TTY prompt → remove eligible files → print summary (FR-005, FR-006)
- Decline: exit `0`, no `os.Remove`

### Register subcommand

In `newLogsCmd()`, after building parent:

```go
cmd.AddCommand(newLogsClearCmd())
```

Update parent `Long` with `supervisor logs clear` line (FR-008).

### Retire root `clear`

- Remove `root.AddCommand(newClearCmd())` from `internal/cli/root.go`
- Delete `internal/cli/clear.go`
- Keep `isUsageError` strings for `requires --yes` (reuse on logs clear)
- Unknown `supervisor clear` → exit `2` unknown command

### Tests (`tests/integration/logs_clear_test.go`)

1. Orphan + `running` log: `--yes` deletes only orphan; registry intact.
2. `starting` row: same protection (table test or second case).
3. Non-TTY without `--yes`: exit `2`, files intact.
4. `logs clear --help` content checks.
5. Root help lacks `clear` subcommand at top level.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Phase Artifacts

| Phase | Output | Path |
|-------|--------|------|
| 0 | Research decisions | [research.md](./research.md) |
| 1 | Data model | [data-model.md](./data-model.md) |
| 1 | CLI contract v10 | [contracts/cli.md](./contracts/cli.md) |
| 1 | Validation guide | [quickstart.md](./quickstart.md) |