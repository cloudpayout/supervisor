# Quickstart: Start command syntax

Validation for `specs/005-start-command-syntax/spec.md`. Contracts: [cli.md](./contracts/cli.md), [session-file.md](./contracts/session-file.md). Fields: [data-model.md](./data-model.md).

## Prerequisites

- Linux or WSL2
- Go 1.27.1 from `go.mod`
- Google Chrome on `PATH`
- Feature 004 stand-in tests updated for positional `start`

## Ordinary confirmation

Run from the repository root:

```text
go test -json ./...
```

The onewin integration suite (from 004, updated for 005):

1. Stand-in on `127.0.0.1` with ephemeral port; marker `balance` when expected cookie present.
2. Temp session file with `"casino": "onewin"` and generated cookie for the stand-in domain.
3. `SUPERVISOR_PAGE_URL=http://127.0.0.1:<port>/`
4. `supervisor start <temp> --headless` → exit `0`, instance id; stand-in shows `balance`.
5. `status` → id running.
6. `screenshot <id>` → PNG within 15s; pixel matches shown color; id still running.
7. `stop <id>` → profile removed.
8. Repeat with file missing cookie → withheld color.
9. Secrets absent from worker log, registry, screenshot name; path may appear on `start` argv.
10. No request to `1win.com` in default run.

## Usage failures (no Chrome)

| Command | Result |
|---------|--------|
| `start --session-file <file>` | Exit `2`, no id |
| `start --session-file=<file>` | Exit `2`, no id |
| `start --casino onewin <file>` | Exit `2`, no id |
| `start --headed <file>` | Exit `2`, no id |
| `start -h` | Exit `2`, no id, no page |
| `start <file> extra` | Exit `2`, no id |
| `start /no/such/file.json` | Non-zero, no id, no profile |
| `start <file with no casino or bad casino>` | Non-zero, no id, no profile |

## Blank start

| Command | Result |
|---------|--------|
| `start` | Exit `0`, id, blank page headed |
| `start --headless` | Exit `0`, id, blank page headless |

## Help

| Command | Result |
|---------|--------|
| `start --help` | Exit `0`, help text, no instance |
| `start <path> --help` | Exit `0`, help text, no instance |

Help text lists positional path, omit-path blank behavior, and `--headless`. It does not list `--session-file`, `--headed`, or `--casino`.

## Headed default (manual, not in default `go test`)

Automated tests assert headed vs headless through the registry `headless` field on the instance row. This section is the check for a **visible window** (FR-006, SC-003).

1. Build or use a small valid session file with `"casino": "onewin"` and at least one cookie (or point `SUPERVISOR_PAGE_URL` at a local stand-in as in the ordinary confirmation).
2. Run `supervisor start <file>` with **no** `--headless`. Expect exit `0`, an instance id, and a **visible** Chrome window.
3. Run `supervisor stop <id>`.
4. Run `supervisor start <file> --headless`. Expect exit `0` and **no** visible window (or a window that does not appear for operator use—headless Chrome).

Do not add this to `go test ./...`.

## Live page (manual, not in default `go test`)

```text
supervisor start <operator-file>
```

Visible window; page `https://1win.com/` when `SUPERVISOR_PAGE_URL` unset. Operator judgment only; not gated on CI.