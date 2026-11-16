# Implementation Plan: Remove Shell Completion Command

**Branch**: `fix/remove-completion-command` | **Date**: 2026-11-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/006-completion-command/spec.md`

## Summary

Cobra v1.10.2 injects a default `completion` subcommand on the root CLI. This feature sets `CompletionOptions.DisableDefaultCmd` on the root command so `completion` is neither listed nor executable. Unknown-command handling (exit `2`) covers legacy invocations. Add integration tests for help output and completion rejection; no changes to start/status/stop/screenshot/worker logic.

## Technical Context

**Language/Version**: Go 1.27.1 (`go.mod`, constitution)

**Primary Dependencies**: `github.com/spf13/cobra` v1.10.2 (completion default cmd only)

**Storage**: N/A (no persisted data)

**Testing**: `go test ./...`; new integration tests for root help and `completion` rejection (FR-006)

**Target Platform**: Linux and WSL2 only

**Project Type**: CLI (`cmd/supervisor`)

**Performance Goals**: N/A (one-time root init field)

**Constraints**: Operator command catalog per constitution II; default tests without live casino

**Scale/Scope**: Single-field change in `internal/cli/root.go` plus tests; optional doc grep for `supervisor completion` in operator docs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec Before Implementation | Pass | Spec `006-completion-command` complete |
| II. CLI as the Operator Interface | Pass | Narrows advertised operator surface to supported commands |
| III. Multiple Browser Instances | Pass | No browser change |
| IV. Supervisor Owns the Run | Pass | Lifecycle commands untouched |
| V. Named Casinos and Games Only | Pass | No casino/catalog change |
| Technical: Go, chromedp, secrets | Pass | No module or secret handling change |
| Delivery: tests before complete | Pass | New integration tests + full suite per quickstart |

**Gate result**: Pass.

**Post–Phase 1 re-check**: Pass. Contract documents catalog only; no new operator capabilities.

## Project Structure

### Documentation (this feature)

```text
specs/006-completion-command/
├── plan.md              # This file
├── spec.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli.md
├── checklists/
│   └── requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
internal/cli/root.go          # CompletionOptions.DisableDefaultCmd = true
tests/integration/            # new root_help_test.go (or equivalent)
```

**Structure Decision**: Single Go module; change isolated to CLI root construction.

## Design

### Root command (`internal/cli/root.go`)

In `newRoot()`, after creating `root := &cobra.Command{...}` and before `AddCommand` calls:

```go
root.CompletionOptions.DisableDefaultCmd = true
```

Effects:

- Cobra does not register default `completion` subcommand (FR-001).
- Root help omits `completion` (FR-002).
- `supervisor completion` becomes unknown command → existing `Execute()` usage path, exit `2`, no script (FR-003, FR-004).

Do not set `DisableFlags` or alter `worker` `Hidden` flag.

### Tests (`tests/integration/root_help_test.go`)

1. `TestRootHelpNoCompletion`: build binary, run `help` / `--help`; assert output lacks `\n  completion ` or equivalent; assert presence of `start`, `status`, `stop`, `screenshot`.
2. `TestCompletionUnknownCommand`: run `completion`, `completion bash`, `help completion`; assert exit `2`, empty stdout, stderr contains `unknown command`.
3. Use existing `supervisorBin(t)` and `isolatedEnv` patterns.

### Documentation

- Grep repo for `supervisor completion` in operator-facing README/quickstarts; remove if found (spec edge case). No change required to 001–005 contracts beyond 006 delta file.

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
| 1 | CLI contract amendment | [contracts/cli.md](./contracts/cli.md) |
| 1 | Validation guide | [quickstart.md](./quickstart.md) |