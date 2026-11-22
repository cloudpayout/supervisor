# CLI Contract: supervisor (010 amendment)

**Version**: v10 (010-logs-clear) — amends [008 logs](../../008-log-files/contracts/cli.md); retires mistaken [009](../../009-clear-command/contracts/cli.md) root `clear` if present in tree  
**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

## Delta summary

| Item | Prior | v10 (010) |
|------|-------|-----------|
| `supervisor clear` (root) | May exist from 009 (registry wipe) | **Removed** — not registered |
| `supervisor logs` | View + `--recent` only | Adds subcommand **`clear`** |
| Log files on disk | Retained until manual delete | Eligible `*.log` may be bulk-deleted via `logs clear` |

## Root catalog

Operator root help MUST list: `start`, `status`, `stop`, `screenshot`, `logs`, `help`.  
MUST **not** list top-level `clear`.

## `supervisor logs` (parent)

View modes unchanged from 008 (`<instance-id>`, `--recent`, `--follow`, `--level`).  
Parent flags MUST NOT be used on the clear subcommand (FR-009).

## `supervisor logs clear`

| Property | Value |
|----------|-------|
| Invocation | `supervisor logs clear` |
| Args | None |
| Flags | `--yes` (boolean; default **false**) — required when stdin is **not** a TTY |
| Registry | **No write** |
| Scope | Top-level `*.log` in `{supervisor_state}/logs/` |
| Protection | Skip file when basename ID has registry row in `starting`, `running`, or `stopping` |

### Confirmation (TTY, `--yes` false)

Prompt MUST include:

1. Count of **eligible** files to delete.
2. Count of **skipped** active lifecycle logs.
3. Default-decline question (e.g. `Clear eligible log files? [y/N]: `).

| Input | Files | Exit |
|-------|-------|------|
| `y` / `yes` (case-insensitive) | Eligible removed; protected kept | `0` |
| Other / empty / cancel | Unchanged | `0` |

### Non-interactive

| Condition | Exit | Files |
|-----------|------|-------|
| Non-TTY, no `--yes` | `2` (usage) | Unchanged |
| Non-TTY, `--yes` | `0` on success | As confirm row |

### Success output

Summary on stdout: deleted count, skipped count; if `eligible_count == 0`, message that nothing eligible was cleared.

### Errors

| Condition | Exit |
|-----------|------|
| Logs dir unreadable / glob failure | `1` |
| `logs clear` with extra args or parent view flags | `2` |

### Invalid combinations (FR-009)

| Invocation | Exit |
|------------|------|
| `logs clear --recent` | `2` |
| `logs clear --follow` | `2` |
| `logs clear <instance-id>` | `2` |
| `logs clear --level error` | `2` |

## Compatibility

- Scripts: `supervisor logs clear --yes` for bulk cleanup without TTY.
- Operators: active runs keep logs; use `stop` before expecting all files eligible.
- Replaces operator workflow of deleting files under `logs/` by hand.