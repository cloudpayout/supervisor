# CLI Contract: supervisor (009 amendment)

**Version**: v9 (009-clear-command) — amends operator catalog after 008  
**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

**Prior contracts**: [001](../../001-supervisor-cli/contracts/cli.md), [002](../../002-status-stop-cli/contracts/cli.md), [003](../../003-screenshot-cli/contracts/cli.md), [005](../../005-start-command-syntax/contracts/cli.md), [006](../../006-completion-command/contracts/cli.md), [008](../../008-log-files/contracts/cli.md). Behavior for existing commands is unchanged unless noted.

## Delta summary

| Item | Prior | v9 (009) |
|------|-------|----------|
| Operator commands | `start`, `status`, `stop`, `screenshot`, `logs`, `help` | Adds **`clear`** |
| `instances.json` | Mutated by start/stop/reconcile | May be reset to empty list via **`clear`** |

## Root catalog (operator-facing)

Root help MUST list: `start`, `status`, `stop`, `screenshot`, `logs`, **`clear`**, `help`.

MUST NOT list: `completion`, `worker` (unchanged from 006/008).

## `supervisor clear`

| Property | Value |
|----------|-------|
| Args | None |
| Flags | `--yes` (boolean; default **false**) — required when stdin is **not** a TTY |
| stdin | Read **only** on a TTY when `--yes` is false (single-line confirmation) |
| Registry | Resets `{XDG_STATE_HOME}/supervisor/instances.json` to `schema_version: 1`, `instances: []` |
| Process cleanup | **None** — running workers/browsers not stopped (FR-005) |

### Confirmation (TTY, `--yes` false)

Prompt MUST include:

1. Number of instance rows currently in the registry.
2. Warning that clearing does **not** stop browsers or worker processes.
3. Question defaulting to decline, e.g. `Clear instance registry? [y/N]: `

| Operator input | Registry | stdout | Exit |
|----------------|----------|--------|------|
| `y` or `yes` (case-insensitive) | Cleared if write succeeds | Short success message | `0` |
| Any other input, empty line, or EOF | Unchanged | No success message (may print cancelled/declined hint) | `0` |
| Interrupt during read (SIGINT) | Unchanged | — | `0` (decline) |

### Non-interactive (`stdin` not a TTY)

| Condition | Registry | stderr | Exit |
|-----------|----------|--------|------|
| `--yes` not set | Unchanged | Message: requires `--yes` when stdin is not a terminal | `2` (usage) |
| `--yes` set | Cleared if write succeeds | — | `0` |

Piping `yes` into `clear` without `--yes` MUST NOT clear (stdin not a TTY).

### Registry errors

| Condition | Registry | Exit |
|-----------|----------|------|
| Read/decode failure (corrupt JSON) | Unchanged | `1` (general) |
| Write failure after confirm | Undefined torn state prevented by flock; on failure operator sees error | `1` |

### Success messages

| Prior registry | Message (informative) |
|----------------|------------------------|
| `N > 0` | Confirms registry cleared (wording implementation-defined) |
| `N == 0` | Confirms already empty / nothing to clear |

Subsequent `supervisor status` (default one-shot): `No instances are running.` exit `0`.

## Exit codes (v9 delta)

| Code | New / changed meaning for `clear` |
|------|-----------------------------------|
| 0 | Success, declined/cancelled, or already-empty success |
| 1 | Registry I/O or decode error before/after confirm |
| 2 | Non-TTY without `--yes` |

Other commands: unchanged from 002/008.

## Environment / platform

Unchanged: `XDG_STATE_HOME`, Linux and WSL2 only.

## Compatibility

- Scripts MUST use `supervisor clear --yes` to reset the registry; MUST NOT rely on stdin prompts in CI.
- Operators SHOULD run `stop` for each instance before `clear` when processes must end; `clear` alone drops tracking only.