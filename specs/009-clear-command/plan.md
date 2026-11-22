# Implementation Plan: Clear Instance Registry Command

**Branch**: `009-clear-command` | **Date**: 2026-11-22 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/009-clear-command/spec.md`

## Summary

Add operator command `supervisor clear` that resets `instances.json` to an empty valid registry (`schema_version: 1`, `instances: []`) after explicit confirmation. On a TTY, prompt with instance count and process warning; default is decline. When stdin is not a TTY, require `--yes` or exit `2` without reading stdin. Implementation is a new `internal/cli/clear.go`, one line in `root.go`, optional `isUsageError` extension, and integration/unit tests. No schema change; no process termination.

## Technical Context

**Language/Version**: Go 1.27.1 (`go.mod`, constitution)

**Primary Dependencies**: `github.com/spf13/cobra` (existing); stdlib `bufio`, `fmt`, `os`, `strings`

**Storage**: Existing JSON registry `{XDG_STATE_HOME}/supervisor/instances.json` via `internal/registry.Store`

**Testing**: `go test ./...`; new `tests/integration/clear_test.go`; unit tests for confirmation parsing in `internal/cli`

**Target Platform**: Linux and WSL2 only

**Project Type**: CLI (`cmd/supervisor`)

**Performance Goals**: Single registry read + optional write; interactive completion under 30s (SC-004)

**Constraints**: Same flock/atomic write as other registry mutators; default test suite without live casino

**Scale/Scope**: One new command file (~120 LOC), root registration, ~3 integration tests, small unit test file

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec Before Implementation | Pass | Spec `009-clear-command` complete |
| II. CLI as the Operator Interface | Pass | Adds supported operator command with help |
| III. Multiple Browser Instances | Pass | Does not change instance isolation model |
| IV. Supervisor Owns the Run | Pass | Does not replace `stop`; documents orphan risk |
| V. Named Casinos and Games Only | Pass | No catalog change |
| Technical: Go, chromedp, secrets | Pass | Registry-only; no secrets in command |
| Delivery: tests before complete | Pass | Integration + unit tests per research R6 |

**Gate result**: Pass.

**Post–Phase 1 re-check**: Pass. Contract and data model describe registry reset only; no new operator lifecycle bypass.

## Project Structure

### Documentation (this feature)

```text
specs/009-clear-command/
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
internal/cli/clear.go           # newClearCmd(), confirm prompt, --yes
internal/cli/root.go          # AddCommand(newClearCmd()); isUsageError tweak
internal/cli/clear_test.go    # confirmYes unit tests (optional co-locate)
tests/integration/clear_test.go # --yes, non-TTY, corrupt registry, help
```

**Structure Decision**: Single Go module; follows existing `stop` / `logs` Cobra layout and `openStore()` / `mustStore()` helpers.

## Design

### Command flow (`internal/cli/clear.go`)

```text
openStore()
  → View: load doc, count = len(Instances)  [fail → exit 1, no prompt]
  → if non-TTY && !yesFlag → usage error exit 2
  → if TTY && !yesFlag → print warning + count + prompt; if !affirmative → exit 0
  → Update: doc.Instances = []; normalize schema; write
  → print success (cleared vs already empty)
  → exit 0
```

Reuse `isTTY(os.Stdin)` from `internal/cli/status.go` (same package).

### Flags and help

- `--yes` / `BoolVar`: skip prompt when set (required for non-TTY).
- `Long` text: registry wipe, confirmation, `--yes`, no process stop.

### Registry write

```go
store.Update(func(d *registry.Document) error {
    d.Instances = []registry.Instance{}
    d.SchemaVersion = registry.SchemaVersion
    return nil
})
```

### Usage errors (`internal/cli/root.go`)

Extend `isUsageError` to recognize the non-TTY clear message (e.g. substring `requires --yes` or `stdin is not a terminal`).

### Tests (`tests/integration/clear_test.go`)

1. `TestClearYesEmptyRegistry` — seed JSON, `clear --yes`, assert empty list + status empty.
2. `TestClearNonTTYRequiresYes` — pipe stdin, no flag, exit 2, registry intact.
3. `TestClearCorruptRegistry` — invalid JSON, `clear --yes`, exit 1, no overwrite.

Optional: extend root help test to assert `clear` listed.

### Documentation

- CLI contract: [contracts/cli.md](./contracts/cli.md) (v9 amendment).
- No change to 001 registry schema file.

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