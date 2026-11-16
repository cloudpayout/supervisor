# CLI Contract: supervisor (006 amendment)

**Version**: v3-surface (006-completion-command) — amends operator catalog only  
**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

**Prior contracts**: [001 v1](../../001-supervisor-cli/contracts/cli.md), [002 status/stop](../../002-status-stop-cli/contracts/cli.md), [003 screenshot](../../003-screenshot-cli/contracts/cli.md), [005 start](../../005-start-command-syntax/contracts/cli.md). Command behavior for `start`, `status`, `stop`, `screenshot`, and hidden `worker` is unchanged except as already amended in those specs.

## Delta summary

| Topic | Before 006 | After 006 |
|-------|------------|-----------|
| Root `completion` subcommand | Present (Cobra default); listed in help | **Absent** — not registered |
| `supervisor completion [shell]` | Emits shell script on stdout | **Unknown command**; exit `2`; no script on stdout |
| `supervisor help completion` | Shows completion help | **Unavailable topic**; exit `0`; stderr `Unknown help topic`; empty stdout |
| Root help command list | Included `completion` | **Excludes** `completion` |

## Root catalog (operator-facing)

Invoking `supervisor`, `supervisor help`, or `supervisor --help` MUST list these top-level commands only:

| Command | Role |
|---------|------|
| `start` | Per 005 |
| `status` | Per 002 |
| `stop` | Per 002 |
| `screenshot` | Per 003 |
| `help` | Built-in Cobra help |

MUST **not** list:

| Name | Requirement |
|------|-------------|
| `completion` | FR-001, FR-002; never advertised or executable |
| `worker` | Hidden internal (001); unchanged |

## Explicitly unsupported commands

### `supervisor completion` (removed)

| Property | Value |
|----------|-------|
| Availability | **Not supported** after 006 |
| Args | N/A — command does not exist |
| stdout | Empty for rejection cases (no autocompletion script) |
| stderr | Unknown-command message from Cobra |
| Exit | `2` (usage / unknown command class, same as other unknown subcommands) |

Same rejection for `supervisor completion bash`, `zsh`, `fish`, and `powershell` (exit `2`, unknown command).

### `supervisor help completion` (unavailable topic)

| Property | Value |
|----------|-------|
| Availability | No `completion` command to document (FR-004) |
| stdout | Empty (no autocompletion script) |
| stderr | Contains `Unknown help topic` for `completion` |
| Exit | `0` (Cobra help for unknown topic; still reports unavailable) |

## Unchanged commands

Refer to linked 001/002/003/005 contract files for:

- `supervisor start` (positional session file, `--headless`, help semantics)
- `supervisor status` (default snapshot; `--follow` / `-f`)
- `supervisor stop` (required instance ID argument)
- `supervisor screenshot`
- `supervisor worker run` (hidden)

## Exit codes

No change to 001/002 exit code table. Unknown top-level command (including former `completion` invocations) remains **exit `2`**.

## Environment

No new variables. Existing `SUPERVISOR_*` and registry layout unchanged.