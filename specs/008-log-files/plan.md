# Implementation Plan: Instance Log Files for Operators

**Branch**: `008-log-files` | **Date**: 2026-11-21 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/008-log-files/spec.md` (slug: `log-files`)

## Summary

Add operator command `supervisor logs` with instance snapshot/follow, severity filter, and `--recent` orphaned-log discovery. Formalize on-disk log lines (`RFC3339Z LEVEL message`), wrap worker stdout/stderr into structured `INFO` at ingest, and emit `NOTICE`/`ERROR` from worker lifecycle (especially entry failures). Preserve log files under `{supervisor_state}/logs/{id}.log` after registry removal. Implementation centers on new `internal/logfmt` (format/parse/filter), worker ingest writer, `internal/cli/logs.go`, and integration tests without live casino for blank-page paths.

## Technical Context

**Language/Version**: Go 1.27.1 (`go.mod`, constitution)

**Primary Dependencies**: `spf13/cobra` (existing); standard library `bufio`, `regexp`, `time`, `sort`, `path/filepath`

**Storage**: Append-only files `{XDG_STATE_HOME or ~/.local/state}/supervisor/logs/<id>.log`; registry `log_path` unchanged

**Testing**: `go test ./...`; unit tests for logfmt; integration tests for `logs` CLI and orphan discovery

**Target Platform**: Linux and WSL2 only

**Project Type**: CLI (`cmd/supervisor`, `internal/cli`, `internal/worker`)

**Performance Goals**: Stream logs without loading full file (FR large logs); follow latency ≤3s (SC-003) via ~500ms poll

**Constraints**: No secrets in logs; chromedp-only browser; CLI-only operator surface; default tests without live casino

**Scale/Scope**: One new top-level command; ~4 new Go packages/files; no registry schema change

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec Before Implementation | Pass | spec + plan + forthcoming tasks |
| II. CLI as the Operator Interface | Pass | `logs` is operator-facing |
| III. Multiple Browser Instances | Pass | Per-instance log files unchanged |
| IV. Supervisor Owns the Run | Pass | Logging accompanies existing lifecycle |
| V. Named Casinos and Games Only | Pass | No new casino identifiers |
| Technical: Go, chromedp, secrets | Pass | Redaction at write; no committed secrets |
| Delivery: tests before complete | Pass | quickstart + FR-012 test map |

**Gate result**: Pass.

**Post–Phase 1 re-check**: Pass. Contracts document CLI only; no web UI or alternate operator surface.

## Project Structure

### Documentation (this feature)

```text
specs/008-log-files/
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
internal/logfmt/              # Format, ParseLine, Filter, tail helpers
internal/worker/
  run.go                      # Open log append; ingest + Notice/Error calls
  aviator.go                  # ERROR/NOTICE at phase boundaries (no cookie in msg)
internal/cli/
  logs.go                     # logs command
  root.go                     # AddCommand(newLogsCmd)
internal/registry/
  paths.go                    # LogsDir() helper if useful for --recent scan
internal/cli/start.go         # Optional: pass log path to worker; ensure logs/ mkdir
tests/unit/logfmt_test.go
tests/integration/logs_test.go
```

**Structure Decision**: Single Go module; logging library isolated for unit tests; CLI thin.

## Design

### Log formatting (`internal/logfmt`)

- `Format(level, msg string) string` — UTC `time.Now()`, trims newlines in msg.
- `ParseLine(line string) (ok bool, ts time.Time, level, msg string)`.
- `FilterLevel(lines, min Level) []string` — structured only.
- `LastErrorMessage(path string) (summary string, ok bool)` for `--recent`.
- `Redact(msg string, secrets []string) string` — substring replacement with `[REDACTED]`.

### Worker ingest

- On `worker run`, open `Layout.LogPath(id)` append-only (flag `--log-path` optional internal flag or derive from store).
- Replace direct writes to raw file with `logfmt.IngestWriter` for os.Stdout/os.Stderr if worker attaches subprocess, and use `logfmt` for explicit events.
- Entry failures: write `ERROR` line before registry row removal (007 compatibility).

### `logs` command (`internal/cli/logs.go`)

- Cobra: `Use: "logs"`, validate args per contract.
- `resolveLogPath(store, id)` — FR-009.
- `streamLog(path, follow, level)` — line-by-line stdout.
- `runRecent(store)` — glob, registry ID set, sort, cap 50, tabwriter.
- Follow: loop read from offset, sleep 500ms, SIGINT → exit 0 (match status).
- Usage errors → exit via existing `Execute()` usage path (`2`).

### Start command touch

- Ensure `logs/` directory created at start (may already happen via `browser.LogFile` MkdirAll).
- Worker child currently inherits closed log FD; worker must reopen log file (document in tasks).

### Tests

| Test | Covers |
|------|--------|
| `logfmt_test.go` | Parse, format, filter, last error |
| `logs_test.go` | start → logs; --level; follow; stop → --recent; not-found |
| Extend secrets | logs output vs session cookie |

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