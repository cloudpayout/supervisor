# Implementation Plan: Open the Brazino page

**Branch**: `020-open-brazino-page` | **Date**: 2026-11-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/020-open-brazino-page/spec.md`

## Summary

Register a second casino, `brazino`, with home page `https://www.brazino777.mx/`. An operator runs `supervisor start <session-file>` where the file's `casino` field is `brazino`; the worker sets cookies and opens that home page (or `SUPERVISOR_PAGE_URL` in tests). No game is registered for `brazino`; `session info` shows `game: null`. The change is additive: `onewin` and Aviator behavior from `specs/004-open-onewin-page` stays as-is. Default tests add `brazino` confirmations on the same `newStandIn` HTTP helper already used for `onewin`, not a second stand-in implementation.

## Technical Context

**Language/Version**: Go 1.27.1 (`go.mod`, constitution)

**Primary Dependencies**: Unchanged from feature 004: `cobra`, `chromedp` v0.23.0, `cdproto` v0.159.0, `gopsutil/v4`. Stand-in remains `net/http` in integration tests.

**Storage**: Unchanged JSON registry (schema version `1`, optional `casino`). Profiles under `$XDG_DATA_HOME/supervisor/profiles/<id>/`. Session files stay on the operator path.

**Testing**: `go test ./...` on Linux/WSL2 with Chrome. New `brazino_*` integration tests mirror the `onewin_start`, `onewin_screenshot`, `onewin_stop`, `onewin_secrets`, and `session_info` patterns using the shared `newStandIn` in `tests/integration/onewin_helpers_test.go`. Each test still binds one ephemeral `127.0.0.1` listener per case (same as today); the suite does not add a Brazino-only server type or a permanent second stand-in process (SC-007).

**Target Platform**: Linux and WSL2 only (`specs/001-supervisor-cli`)

**Project Type**: CLI binary (`cmd/supervisor`)

**Performance Goals**: No new targets. Cookie setup remains one `Network.setCookie` per cookie, then one navigation.

**Constraints**: Session secrets only at runtime. Default tests never open `www.brazino777.mx` or `1win.com`. Identifiers `brazino` and `onewin` match `^[a-z]+$`. `session file` `casino` must match the registered row and `casino.Lookup`. Worker navigation must use the **home page** for `brazino`, not an Aviator-style play path.

**Scale/Scope**: One additional casino, zero new games. Headed/headless unchanged. `screenshot`, `status`, `stop`, and `session info` unchanged except catalog coverage.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec Before Implementation | Pass | Spec `020-open-brazino-page` with clarifications is the source |
| II. CLI as the Operator Interface | Pass | Operator path is `start [session-file]`, `status`, `screenshot`, `stop`, `session info`. `SUPERVISOR_PAGE_URL` stays an undocumented test seam |
| III. Multiple Browser Instances | Pass | One chromedp browser and one profile per instance |
| IV. Supervisor Owns the Run | Pass | No automation script for Brazino games |
| V. Named Casinos and Games Only | Pass | Adds `brazino` with no runnable game. `Brazino`, `brazino777`, hosts as identifiers refused via existing validation |
| Technical: Go, chromedp, module path, secrets | Pass | Module path unchanged; no committed cookies |
| Delivery: tests; default run off live casino | Pass | [quickstart.md](./quickstart.md) adds `brazino` cases on shared stand-in helper; SC-006 non-regression on `onewin` tests |

**Gate result**: Pass.

**Post–Phase 1 re-check**: Pass. Catalog entry, navigation URL rule, session `casino` field, registry `casino`, shared stand-in, and `session info` null game are reflected in [data-model.md](./data-model.md) and [contracts/](./contracts/).

## Project Structure

### Documentation (this feature)

```text
specs/020-open-brazino-page/
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
internal/casino/catalog.go     # add brazino; navigation helper for home vs play URL
internal/session/file.go       # Lookup already validates casino; extend unit cases
internal/worker/run.go         # openTarget uses casino-specific start URL
internal/cli/session_info.go   # already shows game null when Game == ""
tests/unit/catalog_test.go     # brazino catalog + URL tests
tests/integration/
  onewin_helpers_test.go       # shared stand-in (unchanged type)
  brazino_start_test.go        # new
  brazino_screenshot_test.go   # new
  brazino_stop_test.go         # new
  brazino_secrets_test.go      # new
  brazino_session_info_test.go # new
  brazino_usage_test.go        # new (refused identifiers)
```

**Structure Decision**: Same module as 004. No new packages. Prefer extending `internal/casino` with a small `StartURL(casinoID, locale string)` (or equivalent) so `openTarget` no longer calls a global `PlayURL` that only knows `onewin`.

## Design

### Catalog and navigation

1. `Lookup("brazino")` returns `ID: brazino`, `Page: https://www.brazino777.mx/`, `Game: ""`.
2. `Lookup` continues to return `onewin` with `Game: aviator` unchanged.
3. Worker `openTarget` resolves the navigation URL:
   - `brazino` (empty `Game`): registered `Page` (apply `standInPage` locale suffix only when `SUPERVISOR_PAGE_URL` is set, same seam as 004).
   - `onewin`: keep today's Aviator play URL behavior via existing `PlayURL(locale)`.
4. Refused identifiers: any unsupported `casino` in the file or registry fails with the existing usage error shape.

### Start (operator)

1. No session path: blank start unchanged.
2. With session path: `session.Parse` requires `casino: "brazino"` in the file; registry row gets `casino: brazino`.
3. Mismatch between file `casino` and registry (worker re-parse) already fails in `openTarget`.
4. Spawn unchanged: `worker run --id <id> [--headless] --session-file <path>`.

### Session info

For `brazino` rows, `session info` prints `casino: brazino` and `game: null` because catalog `Game` is empty (existing logic in `session_info.go`).

### Tests

1. Copy the smallest 004 confirmation slice: cookie shown / withheld, `status`, `screenshot`, pixel check, `stop`, secrets absent from argv/log/registry/filename.
2. Use `{"casino":"brazino", ...}` session files and `startCasino` helper (may rename to `startWithSession` in a follow-up task; not required for behavior).
3. Do not fork the stand-in into `brazino_stand_in.go`; reuse `newStandIn`.
4. Run existing `TestOneWin*` unchanged for SC-006.

### Screenshot, stop, profile

No command changes. Same lifecycle and FR-007 profile removal as 004.

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