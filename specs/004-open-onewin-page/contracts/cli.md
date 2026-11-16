# CLI Contract: onewin session start

**Feature**: `004-open-onewin-page`
**Extends**: `specs/001-supervisor-cli/contracts/cli.md`
**Binary**: `supervisor`

`status` is unchanged. `stop <instance-id>` is unchanged except that ending an instance also removes that instance's profile directory after the browser process has exited.

## `supervisor start`

| Property | Value |
|----------|-------|
| Flags kept | `--headed`, `--headless`. Mutually exclusive. Default headed when both are omitted. Both set → exit `2`, no registry row. |
| Flags added | `--casino`, `--session-file` |
| Neither new flag | Blank page, as in feature 001. |
| Both new flags | Required together. One without the other → exit `2`, no registry row. |
| `--casino` | `onewin` only. Must match `^[a-z]+$`. `Onewin`, `1win.com`, and any other identifier → exit `2`. |
| `--session-file` | Path to a [session file](./session-file.md). The JSON text is not a legal argument. |
| Page | `https://1win.com/` from the catalog. The operator has no page flag. |
| Success | Exit `0`. Stdout is the instance id. The process exits while the instance keeps running. Cookies are applied, then the page opens. |
| Bad file | Missing, unreadable, not JSON, unknown field, empty `cookies`, empty name or value, or illegal `sameSite` → non-zero, no instance id, no registry row, no profile left behind. The error text does not include a cookie value. |

Spawn, internal only:

```text
supervisor worker run --id <id> [--headless] --session-file <path>
```

`--session-file` is present only when `--casino` was set. The path may appear in the command line. The file's contents must not.

## Hidden worker

`supervisor worker run` gains `--session-file`. When the instance's `casino` is `onewin`, the worker reads that file, sets every cookie, then navigates. `Network.setCookie` errors fail the worker, which removes the registry row and the profile.

`SUPERVISOR_PAGE_URL`, when set, replaces the catalog page for that process. It is not listed in operator help. The default test sets it to the stand-in. When it is unset, `onewin` opens `https://1win.com/`.

## `supervisor screenshot`

Unchanged from `specs/003-screenshot-cli/contracts/cli.md`. This feature adds no flag and no new failure text.

When the instance is an `onewin` row in `running`, the image is the page the worker opened after setting cookies. The command does not navigate, does not read the session file, and does not remove the profile. The printed file name must not contain a cookie value.

## What is not a command

There is no flag or subcommand that starts `aviator`, Spaceman, Dino, or a casino other than `onewin`.

## Registry

An `onewin` row includes `casino` with value `onewin`. It does not include cookies, the session file path, or the page URL. Schema version stays `1`.
