# Implementation Plan: Open the onewin page

**Branch**: `004-open-onewin-page` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/004-open-onewin-page/spec.md`

## Summary

`start` gains `--casino onewin` and `--session-file`. The worker reads that file, sets every cookie through chromedp, then opens the catalog page `https://1win.com/`. A blank `start` stays a blank page. The instance row stores the casino identifier and no secret. The profile directory is removed after the browser has exited. The ordinary test is a Go HTTP stand-in on `127.0.0.1`, then `status`, the existing `screenshot` command, and `stop`. `aviator` is the catalog's game name and has no command.

## Technical Context

**Language/Version**: Go 1.27.1 (`go.mod`, constitution)

**Primary Dependencies**: `github.com/spf13/cobra`, `github.com/chromedp/chromedp` v0.23.0, `github.com/chromedp/cdproto` v0.159.0, `github.com/shirou/gopsutil/v4`. The stand-in uses the standard library `net/http`.

**Storage**: Existing JSON registry, schema version `1`, plus optional `casino`. Profiles under `$XDG_DATA_HOME/supervisor/profiles/<id>/` (or `~/.local/share` when unset). The session file stays on the path the operator passes.

**Testing**: `go test ./...` on Linux and WSL2, with Chrome. The new confirmation is in that default run. It sets `SUPERVISOR_PAGE_URL` to the stand-in, captures that page with `screenshot`, and does not open `1win.com`.

**Target Platform**: Linux and WSL2 only, as in `specs/001-supervisor-cli`

**Project Type**: CLI binary (`cmd/supervisor`)

**Performance Goals**: No new target. Feature 001's `start` return and `status` refresh targets still apply. Cookie setup is one `Network.setCookie` call per cookie in the file, then one navigation.

**Constraints**: Session values only at runtime. The JSON document is not a process argument. Cookie values stay out of the registry, the worker log, error text, and the screenshot file name. Default tests do not open a live casino. Identifiers are `onewin` and `aviator`. A rejected live cookie stops the feature rather than expanding it.

**Scale/Scope**: One casino. One named game that is not started. Headed and headless keep their current meaning. `stop <instance-id>` and `screenshot <instance-id>` keep their current arguments.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec Before Implementation | Pass | Spec `004-open-onewin-page` is the source for this plan |
| II. CLI as the Operator Interface | Pass | Operator behavior is `start`, `status`, `screenshot`, and `stop`. `SUPERVISOR_PAGE_URL` is an undocumented test seam, not a supported operator interface |
| III. Multiple Browser Instances | Pass | One chromedp browser and one profile per instance |
| IV. Supervisor Owns the Run | Pass | `start`, `status`, and `stop` still own the instance. No game script is added |
| V. Named Casinos and Games Only | Pass | Constitution 1.2.0 allows a casino to be supported before any of its games can be run. The only casino is `onewin`. The only game name is `aviator`. Nothing starts that game. Both identifiers match `^[a-z]+$`. `1win.com` is the page, not an identifier |
| Technical: Go, chromedp, module path, secrets | Pass | Product and the stand-in stay Go 1.27.1. Module path unchanged. Browser control stays chromedp. Session cookies are not committed. The assessment's FastAPI stand-in is not used |
| Delivery: tests before complete; default run off a live casino | Pass | [quickstart.md](./quickstart.md) puts `start`, `status`, `screenshot`, and `stop` in `go test ./...` against `127.0.0.1`. The live page is a manual check and does not gate the default run |

**Gate result**: Pass.

**Post–Phase 1 re-check**: Pass. The catalog, the session file, the registry field, the stand-in, and the reused `screenshot` command match the table above. No second operator interface was added. No cookie value is in the data model or the contracts. The screenshot file is operator output, not a registry field.

## Project Structure

### Documentation (this feature)

```text
specs/004-open-onewin-page/
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
└── tasks.md             # /speckit-tasks (the earlier file is the blocked list)
```

### Source Code (repository root)

```text
cmd/supervisor/
internal/cli/start.go       # --casino and --session-file; blank start unchanged
internal/cli/worker.go      # pass --session-file through to the worker
internal/cli/stop.go        # remove the profile after the browser has exited
internal/cli/screenshot.go  # unchanged; captures the page the worker is showing
internal/worker/run.go      # read the file, set cookies, then navigate
internal/browser/chromedp.go
internal/registry/types.go  # optional casino field
internal/registry/reconcile.go
internal/casino/catalog.go  # onewin, https://1win.com/, aviator
internal/session/file.go    # parse the session file; no cookie value is logged or stored
tests/integration/          # stand-in, start, status, screenshot, stop
```

**Structure Decision**: Same module as feature 001. The catalog is a new package so the page address and the game name are not flags. The session parser is `internal/session/file.go`. The stand-in lives in the integration test, not in the operator binary.

## Design

### Start

1. Neither `--casino` nor `--session-file`: today's blank start.
2. Exactly one of those flags: exit `2`, no registry write.
3. `--casino` must be `onewin`. Anything else, including `Onewin` and `1win.com`, exits `2`.
4. Parse the session file. A bad document exits non-zero before preflight. The error does not contain a cookie value.
5. Existing preflight, registry append, and detached worker. The worker arguments add `--session-file` and the path. They do not add the document.
6. The new row sets `casino` to `onewin`. Failed spawn still rolls back the row and calls `RemoveProfile`.

### Worker

When `casino` is `onewin`, read `--session-file` again, call `Network.setCookie` for each cookie, then navigate. The URL is `SUPERVISOR_PAGE_URL` when that variable is set, otherwise `https://1win.com/`. A set or navigation error removes the row and the profile. On any exit, cancel the browser allocator, wait until the browser PID is not alive, then `RemoveProfile`, then drop the row. Logs and errors omit cookie values.

### Stop and reconcile

`proc.Terminate` already waits and then sends `SIGKILL`. After that and `KillByProfile` return, call `RemoveProfile`. Do this on the live `stop` path and inside `registry.Apply`, so a browser that exits outside `stop` loses its profile too. `RemoveProfile` remains safe to call more than once.

### Screenshot

No change to `internal/cli/screenshot.go` or the capture socket. The worker's navigation is what the existing command captures. The test samples one pixel from the PNG with the standard-library decoder. That pixel matches the stand-in's shown color or its withheld color. The file name and the screenshot command line do not contain the cookie value. The instance stays `running` until `stop`.

### Ordinary test

A `net/http` server on `127.0.0.1` shows the marker `balance` and one full-viewport color only for the cookie generated in that test. Without that cookie it uses a different full-viewport color and omits the marker. The test runs `start` (headless), `status`, `screenshot`, and `stop`, with `SUPERVISOR_PAGE_URL` pointed at the server. It checks the marker, the PNG pixel, the missing profile, and the absence of the generated value from the worker command line, the screenshot command line, the worker log, the registry, and the file name. A second run expects the withheld color. Usage failures in the quickstart do not launch Chrome.

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
