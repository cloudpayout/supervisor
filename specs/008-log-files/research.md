# Research: Instance Log Files (008-log-files)

**Date**: 2026-11-21

## 1. Structured line format and parsing

**Decision**: Canonical on-disk line: `{time.RFC3339Nano in UTC} {LEVEL} {message}` with `LEVEL` ∈ `ERROR`, `NOTICE`, `INFO`; single ASCII space separators; message has no `\n`. Parser regex: leading timestamp token, then level token, remainder is message. `ParseLine` returns `(structured bool, entry)`.

**Rationale**: Matches spec clarifications; `RFC3339Nano` satisfies fractional seconds; uppercase levels align with `--level error|notice` flags.

**Alternatives considered**:

- JSON Lines — rejected (operator grep/readability).
- Lowercase levels — rejected (spec locked uppercase).

## 2. Raw stdout/stderr ingest (FR-002b)

**Decision**: Replace direct `worker.Stdout/Stderr = logf` with an `io.Writer` that line-buffers, skips empty lines, applies cookie/session redaction (reuse pattern from integration `assertNoSecret`: scan known secret substrings from session file at worker start if available, else byte passthrough), and writes each line as `INFO` via the structured formatter.

**Rationale**: Parent `start` holds the log file handle briefly; worker reopens log path append-only on run (existing pattern or add worker flag for log path). Structured ingest must live in the worker process for chromedp subprocess output captured inside worker.

**Alternatives considered**:

- Wrap only in `start` for worker binary stdout — insufficient for in-process chromedp logs.
- Leave raw forever — rejected (spec FR-002b).

## 3. Worker diagnostic lines

**Decision**: Add `internal/logfmt` (or `internal/supervisorlog`) used by `worker` and future CLI diagnostics: `Notice`, `Info`, `Error` methods writing to the instance log file (opened once at worker run). Emit `NOTICE` for entry phase transitions (`entering`, `ready`); `ERROR` before teardown on entry failure (message names cause, never cookie values per 007).

**Rationale**: Satisfies FR-003 and SC-006; keeps secrets policy aligned with aviator errors.

**Alternatives considered**:

- Log only via fmt to stderr — would double through ingest as INFO; explicit levels preferred for `--level` and `--recent`.

## 4. `logs` command implementation

**Decision**: New `internal/cli/logs.go` registered in `root.go`. Subcommands/modes:

- `logs <id>` — resolve path via FR-009 (registry `log_path` then canonical fallback); stream file to stdout; optional `--follow` polling every **500ms** (tighter than status 2s for SC-003); optional `--level error|notice` filtering structured lines only; stderr notice when instance not tracked.
- `logs --recent` — scan `layout.LogsDir()` glob `*.log`; build registry ID set; orphaned = file basename without `.log` not in set; sort by `ModTime` desc; cap 50; for each file read backwards or forward-scan for last `ERROR` message; tabwriter columns ID, UTC mtime, error summary (truncate display ~80 runes).

**Rationale**: No new dependencies; mirrors `status` follow + interrupt handling; 500ms poll meets SC-003 without fsnotify.

**Alternatives considered**:

- `fsnotify` — extra complexity and edge cases on WSL2.
- `tail -F` external command — not portable, violates single-binary CLI.

## 5. Log file retention on teardown

**Decision**: Audit `reconcile`, `stop`, `worker` exit paths; remove any `os.Remove` on log paths if present; document that only `browser.RemoveProfile` runs. No code change if already absent (verified: only profiles removed).

**Rationale**: FR-009b.

## 6. Testing strategy

**Decision**:

- **Unit**: `logfmt` parse/format/filter; ingest writer redaction; path resolution helper.
- **Integration**: `tests/integration/logs_test.go` with isolated `XDG_STATE_HOME`; start headless blank instance; assert structured lines in file; `logs <id>` snapshot; `--level error`; follow interrupt; stop/crash simulation removing registry while log remains; `logs --recent`; secrets test mirroring `onewin_secrets_test.go` on log viewer output.

**Rationale**: Constitution default tests without live casino; FR-012 coverage map.

## 7. CLI contract versioning

**Decision**: New contract file `contracts/cli.md` as **008 amendment** atop latest operator CLI doc chain (reference 002/006/007 deltas); document `logs` flags, exit codes, and stderr notice strings.

**Rationale**: Spec Kit contract pattern for this repo.