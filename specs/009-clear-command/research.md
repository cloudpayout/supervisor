# Research: Clear Instance Registry Command

**Feature**: `009-clear-command` | **Date**: 2026-11-22

## R1: Registry reset mechanism

**Decision**: Use existing `registry.Store.Update` to set `Document.Instances` to an empty slice and normalize `SchemaVersion` to `registry.SchemaVersion` (1). Persistence goes through `writeDoc` (truncate, rewrite, `Sync`) under the same exclusive `flock` as `start` / `stop` / `status`.

**Rationale**: Reuses the only supported registry writer; avoids hand-deleting `instances.json` or bypassing locks. Empty-file and missing-file cases already yield an empty in-memory `Document` via `readDoc` (`internal/registry/registry.go`).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `os.Remove` registry path | Skips lock; races with concurrent CLI; may leave parent dir inconsistent |
| New `Store.Clear()` API only used here | Thin wrapper is optional in implementation; behavior is one `Update` callback |
| Wipe entire state directory | Out of spec scope (profiles, logs must remain) |

## R2: Confirmation and TTY policy

**Decision**:

- **Interactive (stdin is a TTY)**: Print a warning line with instance count, warn that browsers/workers are not stopped, then prompt `Clear instance registry? [y/N]: ` (default decline). Read one line with `bufio.Reader`; affirmative iff trimmed lowercase is `y` or `yes`; anything else including empty line → decline (exit `0`, no write).
- **Non-interactive (stdin not a TTY)**: Do not read stdin. Require `--yes` (boolean, default `false`); if absent, return a usage-class error (exit `2`) with a message that `--yes` is required when stdin is not a terminal.
- **SIGINT during prompt**: Treat as decline; exit `0`; registry unchanged (read may return `io.EOF` / interrupt; map to declined).

**Rationale**: Matches FR-002, FR-003, FR-005, FR-006 and mirrors industry practice for destructive CLIs. Default `[y/N]` avoids accidental wipe. Non-TTY guard prevents CI/pipe accidents (spec User Story 3).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Always prompt even on pipes | Allows `echo yes \| clear` in scripts without explicit opt-in; violates FR-006 |
| `--force` that also kills processes | Out of scope; spec says registry-only reset |
| Type-to-confirm phrase (`DELETE`) | Heavier UX than spec requires |

## R3: Read-before-write and corrupt registry

**Decision**: Before prompting, call `store.View` to load the document and compute `len(doc.Instances)`. If `View` fails (I/O, decode error on non-empty invalid JSON), print `supervisor: …` on stderr and exit `1` without prompting and without writing (FR-008). If the registry file is missing or zero-length, count is `0`; after confirmation, `Update` still writes a valid empty document.

**Rationale**: Operator must see accurate count on the prompt. Overwriting corrupt JSON without a successful read would mask data-loss recovery issues.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Prompt first, read inside `Update` | Operator confirms without knowing count or read failure class |
| Delete corrupt file without prompt | Violates FR-008 |

## R4: Usage errors and exit codes

**Decision**: Non-TTY without `--yes` returns `fmt.Errorf("clear requires --yes when stdin is not a terminal")` (or equivalent). Extend `isUsageError` in `internal/cli/root.go` to treat that message (and optionally a stable `usage:` substring) as usage class → exit `2`, consistent with `stop` missing ID and `logs` usage errors.

**Rationale**: Reuses existing `Execute()` path; no new exit-code family.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Exit `1` for non-TTY | Breaks spec FR-006 / SC-003 (usage-error class) |
| Cobra-only flag validation | Does not cover TTY detection logic |

## R5: Operator discovery and help

**Decision**: Register `newClearCmd()` in `newRoot()` after existing operator commands. `Use: "clear"`, `Short`/`Long` describe registry wipe, confirmation, `--yes`, and that processes are not stopped. Root help automatically lists `clear` when added via `AddCommand`.

**Rationale**: FR-007; same Cobra pattern as `logs`, `stop`.

## R6: Testing strategy

**Decision**:

| Area | Approach |
|------|----------|
| `--yes` with seeded registry | Integration: write fixture JSON to `isolatedEnv` registry path, run `clear --yes`, assert `instances` empty and `status` empty message |
| Non-TTY without flag | Integration: run `clear` with stdin from pipe/null, assert exit `2` and registry unchanged |
| Decline / affirmative parsing | Unit tests in `internal/cli` for `confirmYes(line string) bool` |
| Root help | Extend or add integration test asserting `clear` appears in help output |
| Invalid JSON | Integration: corrupt file, assert exit `1`, content unchanged |

Full `go test ./...` for regression (constitution delivery workflow).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Only manual quickstart for TTY prompt | Leaves FR-002 / SC-004 untested in CI; unit tests cover parser |
| creack/pty in default tests | Extra dependency; optional manual quickstart covers TTY path |