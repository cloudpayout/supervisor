# Implementation Plan: Start command syntax

**Branch**: `005-start-command-syntax` | **Date**: 2026-11-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/005-start-command-syntax/spec.md`

## Summary

Operator `start` takes an optional positional session file path and optional `--headless`. Casino comes from the file's `casino` field. Without a path, behavior stays a blank page. Remove `--headed`, `--casino`, and `--session-file` from operator help and argv. `--help` shows help without launching; `-h` is a usage error. Session parsing gains required `casino`. Hidden worker keeps `--session-file <path>` on spawn.

## Technical Context

**Language/Version**: Go 1.27.1 (`go.mod`, constitution)

**Primary Dependencies**: `github.com/spf13/cobra`, chromedp stack (unchanged from 004)

**Storage**: JSON registry v1 with optional `casino`; profile dirs unchanged

**Testing**: `go test ./...` on Linux/WSL2 with Chrome; migrate 004 onewin tests to positional `start <path> --headless`; add usage cases for removed flags and `-h`. Visible window (SC-003): manual [quickstart.md](./quickstart.md#headed-default-manual-not-in-default-go-test) validation (T017); default test run uses registry `headless` as the automated proxy (T004, T008), not window visibility in CI.

**Target Platform**: Linux and WSL2 only

**Project Type**: CLI (`cmd/supervisor`)

**Performance Goals**: No change from 001/004

**Constraints**: Cookie values never in registry, logs, or errors; operator path may appear on argv; default tests avoid live casino

**Scale/Scope**: CLI syntax and session parse shape only; no new casinos or games

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec Before Implementation | Pass | Spec and clarifications in `005-start-command-syntax` |
| II. CLI as Operator Interface | Pass | Change is operator `start` only |
| III. Multiple Browser Instances | Pass | Unchanged isolation |
| IV. Supervisor Owns the Run | Pass | Lifecycle unchanged |
| V. Named Casinos and Games Only | Pass | `casino` in file validated via `casino.Lookup` |
| Technical: Go, chromedp, secrets | Pass | Module path unchanged |
| Delivery: tests before complete | Pass | quickstart updates integration and usage tests |

**Gate result**: Pass.

**Post–Phase 1 re-check**: Pass. Contracts and data model keep casino out of argv; worker internal flag does not expand operator surface.

## Project Structure

### Documentation (this feature)

```text
specs/005-start-command-syntax/
├── plan.md
├── spec.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── cli.md
│   └── session-file.md
├── checklists/
│   └── requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
internal/cli/start.go       # positional path, --headless only, help/-h behavior
internal/session/file.go    # required casino field; return casino + cookies
internal/worker/run.go      # drop onewin default when file present; use registry casino
tests/integration/          # onewin_* and start_* tests updated for new syntax
```

**Structure Decision**: Single Go module; no new packages.

## Design

### Start (`internal/cli/start.go`)

1. `Args`: at most one positional path.
2. Flags: `--headless` only among display flags. No `--headed`, `--casino`, `--session-file`.
3. Help: disable `-h` shorthand on this command; `--help` must not start Chrome.
4. `len(args)==0`: blank start (existing flow); `useHeadless := headless`.
5. `len(args)==1`: parse session file (casino + cookies + file on disk); `casino.Lookup`; set `inst.Casino`; spawn worker with `--session-file` path.
6. Reject legacy flags via Cobra (unknown flag → exit `2`).
7. Rollback and profile removal unchanged on failure after row creation.

### Session (`internal/session/file.go`)

1. JSON requires `casino` and `cookies` when parsing for a casino start.
2. Export struct or return pair `(casinoID string, cookies []Cookie, err error)`.
3. Update unit tests for missing `casino` and invalid identifier messaging without leaking cookie values.

### Worker (`internal/worker/run.go`)

1. When `sessionFile` non-empty, use `inst.Casino` from registry (set by start). Remove implicit `onewin` if registry casino empty (should not happen after start fix).
2. Cookie load still from file via extended parse or cookies-only re-read consistent with registry casino.

### Tests

1. Replace `start --casino onewin --session-file` with `start <path>` in integration tests.
2. Session JSON fixtures add `"casino":"onewin"`.
3. `onewin_usage_test.go`: assert `--session-file`, `--casino`, `--headed`, `-h` exit `2`.
4. `start_worker_test.go`: blank `start` and `start --headless` unchanged.
5. Optional: `start --help` does not create registry row.

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
| 1 | CLI and session-file contracts | [contracts/](./contracts/) |
| 1 | Validation guide | [quickstart.md](./quickstart.md) |