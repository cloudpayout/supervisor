# CLI Contract: start command syntax (005 amendment)

**Version**: v5 (005-start-command-syntax)  
**Amends**: [001 v1](../../001-supervisor-cli/contracts/cli.md), [004 onewin](../../004-open-onewin-page/contracts/cli.md)  
**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

## Delta summary

| Area | Feature 004 | Feature 005 |
|------|-------------|-------------|
| Session path | `--session-file <path>` | Single optional positional `<path>` |
| Casino | `--casino onewin` with file | `casino` field in session file only |
| Display | `--headed` / `--headless`, default headed | `--headless` only, default headed |
| Blank start | `start` with no casino flags | `start` or `start --headless` with no positional |
| Help | (Cobra default) | `--help` only; `-h` is usage error |

`status`, `stop`, and `screenshot` are unchanged (002/003).

## `supervisor start`

| Property | Value |
|----------|-------|
| Args | Zero or one positional: path to [session file](./session-file.md). More than one → exit `2`, no row. |
| Flags kept | `--headless` (boolean; default false) |
| Flags removed (usage error if passed) | `--headed`, `--casino`, `--session-file` (any spelling) |
| Help | `--help` prints help, exit `0`, no instance. `-h` → exit `2`, no instance. |
| No positional | Blank page (`about:blank`). Exit `0`, prints instance id. |
| With positional | Parse file; apply cookies; open registered page for file's `casino`. Exit `0`, prints id. |
| Missing path on disk | Non-zero, no stdout id, no row, no profile left behind. |
| Bad session file | Non-zero, no id, no row, no profile (same as 004 parse rules plus `casino`). |
| Success stdout | Instance `id` (one line) |
| Usage exit | `2` for Cobra usage (unknown flags, `-h`, extra args, removed flags) |

Successful operator invocations:

```text
supervisor start
supervisor start --headless
supervisor start <path>
supervisor start <path> --headless
supervisor start --headless <path>
supervisor start --help
```

Internal worker spawn (unchanged):

```text
supervisor worker run --id <id> [--headless] --session-file <path>
```

`--session-file` on the worker is internal only. It is not an operator flag after this amendment.

## Hidden worker

No operator-facing change. Worker still receives `--session-file` from detached `start` when the instance has a session path. Registry row `casino` is set from the parsed file before spawn.

## Registry

When a positional path is used, row includes `casino` from the file (e.g. `onewin`). Blank start omits `casino`. Schema version stays `1`. No cookie values or JSON body on the row.

## Exit codes

Unchanged from 001: `0` success; `2` usage; `1` other failures with no id on synchronous `start` failure.