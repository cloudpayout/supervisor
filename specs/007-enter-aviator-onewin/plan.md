# Implementation Plan: Enter Aviator on onewin

**Branch**: `007-enter-aviator-onewin` | **Date**: 2026-11-21 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/007-enter-aviator-onewin/spec.md`. Planning reference: RobotCasino Aviator workflow (currency modal, concurrent iframe wait, `spribelaunch.com/aviator`, inner `game-logo`). User slug `onewin-in-aviator` maps to this feature / assessment `aviator-in-onewin`.

## Summary

For `onewin`, the worker navigates to the Aviator **play URL** (locale from session), automates the outer-page currency modal, activates play or demo, and waits for **Aviator ready** (Spribe launch iframe + inner `game-logo`). `start` returns after spawn; entry runs in the worker. Optional session fields `currency`, `demo`, `locale`. Registry gains `entry` phase `entering` → `ready`. Failures tear down silently on CLI. Integration tests use a loopback Aviator HTML stand-in with stable `data-supervisor-*` hooks; live site uses selectors informed by RobotCasino but verified during implementation.

## Technical Context

**Language/Version**: Go 1.27.1 (`go.mod`, constitution)

**Primary Dependencies**: cobra, chromedp v0.23.0, cdproto v0.159.0, gopsutil v4; `net/http` stand-in

**Storage**: JSON registry schema bump or optional `entry` field; session file on operator path; profiles under XDG data dir

**Testing**: `go test ./...`, headless, Chrome. Aviator stand-in on `127.0.0.1` with `SUPERVISOR_PAGE_URL`. Currency fail-fast and ready-path cases. No live `1win.com` in default run.

**Target Platform**: Linux / WSL2

**Project Type**: CLI (`cmd/supervisor`)

**Performance Goals**: Entry complete within 60s (SC-002). Overlap modal + iframe wait when possible (RobotCasino pattern).

**Constraints**: No secrets in registry/logs/argv. No in-round play. No CLI failure taxonomy. No reload after cookies. chromedp only.

**Scale/Scope**: Single casino `onewin`, game `aviator` entry only.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec Before Implementation | Pass | Spec 007 + clarify session |
| II. CLI as the Operator Interface | Pass | Still `start`, `status`, `screenshot`, `stop`; `SUPERVISOR_PAGE_URL` remains test seam |
| III. Multiple Browser Instances | Pass | One chromedp profile per instance |
| IV. Supervisor Owns the Run | Pass | Worker entry is lifecycle-owned; no external game scripts |
| V. Named Casinos and Games Only | Pass | `onewin` + `aviator`; identifiers lowercase; game **entry** now explicit (constitution 1.2.0 allows run as separate decision — taken in assessment **go**) |
| Technical: Go, chromedp, secrets | Pass | |
| Delivery: tests; default off live casino | Pass | quickstart + stand-in |

**Gate result**: Pass.

**Post–Phase 1 re-check**: Pass. Data model and contracts keep secrets out of registry; live play external to default tests.

## Project Structure

### Documentation (this feature)

```text
specs/007-enter-aviator-onewin/
├── plan.md
├── spec.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── cli.md
│   ├── session-file.md
│   └── aviator-entry.md
├── checklists/
│   └── requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
cmd/supervisor/
internal/casino/catalog.go       # PlayURL(locale), game URL registration
internal/session/file.go         # currency, demo, locale parse + validation
internal/registry/types.go       # optional Entry phase
internal/worker/run.go           # invoke entry after navigate
internal/worker/aviator.go       # modal + iframe ready (new)
internal/cli/status.go             # print entry phase when set
tests/integration/
  onewin_helpers_test.go         # existing balance stand-in
  aviator_standin_test.go        # new modal + iframe HTML (new)
  onewin_aviator_*.go              # entry, currency, timeout, secrets (new/extend)
tests/unit/
  session_file_test.go           # optional fields
  catalog_test.go                # play URLs
```

**Structure Decision**: Entry logic isolated in `internal/worker/aviator.go` to keep `run.go` orchestration readable. Stand-in HTML served from test helper, not production binary.

## Design

### Session parse

`Parse` returns `(casinoID, cookies, PlayPreferences, error)`. Validate `currency` length 3, `locale` non-empty without slashes. Return typed defaults for omitted keys.

### Catalog

`PlayURL(locale string) string` implements URL rules in research. `openTarget` uses play URL for `onewin` when starting with session (not homepage only). `SUPERVISOR_PAGE_URL` override unchanged.

### Worker flow (`onewin` + session)

1. Set cookies against effective page URL.
2. Navigate play URL.
3. Set registry `entry=entering` (if not already).
4. Run `aviator.Entry(ctx, prefs)`:
   - Modal automation per [contracts/aviator-entry.md](./contracts/aviator-entry.md)
   - Concurrent iframe wait where applicable
   - 60s deadline
5. Success: `entry=ready`, keep browser running.
6. Failure: cancel, remove profile, remove registry row (no cookie values in errors).

`start` does not wait on step 4–6.

### Status

Include `entering` / `ready` in text output for instances with `entry` set.

### Integration tests

- Extend or parallel `newStandIn` → `newAviatorStandIn` with modal DOM and iframe.
- Tests: happy path ready, default currency, `demo` true, `locale` prefix via URL assertion on stand-in request path, currency ZZZ fail, timeout, secrets sweep.
- Migrate existing onewin tests: either still use balance stand-in with override to **homepage** behavior removed — **004 tests must expect play URL + entry** or separate test files. Plan: update onewin suite to Aviator stand-in for session starts (breaking change aligned with spec amendment).

### Live selectors

Implementation discovers live DOM (RobotCasino: `game-play-card_currencyButton-bVSBd`, `Play`, `USD`; spec: `Jugar`, `Modo demo`, `MXN`). Selectors live in `aviator.go` constants, tuned manually; not committed screenshots.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Phase Artifacts

| Phase | Output | Path |
|-------|--------|------|
| 0 | Research decisions | [research.md](./research.md) |
| 1 | Data model | [data-model.md](./data-model.md) |
| 1 | Contracts | [contracts/](./contracts/) |
| 1 | Validation guide | [quickstart.md](./quickstart.md) |