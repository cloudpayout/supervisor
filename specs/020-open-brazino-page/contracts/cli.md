# CLI Contract: Brazino session start

**Feature**: `020-open-brazino-page`
**Extends**: `specs/004-open-onewin-page/contracts/cli.md`, `specs/005-start-command-syntax/contracts/cli.md`
**Binary**: `supervisor`

Adds `brazino` as a supported session `casino` value. Commands otherwise unchanged.

## `supervisor start`

| Property | Value |
|----------|-------|
| Invocation | `supervisor start [path]` — optional path to a session file |
| No path | Blank page (feature 001), unchanged |
| With path | Parse session file; open catalog page for `doc.casino` |
| Supported `casino` in file | `onewin` (004), `brazino` (this feature) |
| `brazino` page | `https://www.brazino777.mx/` from catalog |
| Refused casino strings | Unsupported ids, uppercase, digits, hosts (e.g. `brazino777`, `www.brazino777.mx`) at parse/Lookup |
| Mismatch | File `casino` must equal registry `casino` on the row |
| Success | Exit `0`, instance id on stdout, cookies applied then navigation |
| Bad file | Same failure classes as 004: missing file, bad JSON, empty cookies, unknown fields |

Worker spawn (unchanged):

```text
supervisor worker run --id <id> [--headless] --session-file <path>
```

## Test seam

`SUPERVISOR_PAGE_URL`, when set in the worker environment, replaces the catalog navigation URL for **any** supported casino. Not documented in operator help. Default `go test` sets it to the shared `127.0.0.1` stand-in.

## `supervisor session info`

For a running `brazino` instance:

```text
casino: brazino
game: null
strategy: null
date: <UTC YYYY-MM-DD HH:MM>
```

Unchanged failure and stdout rules from `specs/019-session-data`.

## `supervisor status`, `screenshot`, `stop`

Unchanged from 004 and 003. Instance lifecycle and screenshot rules apply to `brazino` rows the same way as `onewin` rows.

## What is not a command

No flag or subcommand starts a Brazino game. No operator flag selects `www.brazino777.mx`; only identifier `brazino` via the session file.

## Registry

A `brazino` row includes `casino: "brazino"`. No game field. No cookie, session path, or page URL stored. Schema version stays `1`.

## Relationship to 004

All 004 `onewin` behaviors and tests remain valid. This contract only adds rows in the "supported casino" column.