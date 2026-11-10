# CLI Contract: supervisor (002 amendment)

**Version**: v2 (002-status-stop-cli) — amends [001 v1](../001-supervisor-cli/contracts/cli.md)  
**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

## Delta summary

| Command | v1 (001) | v2 (002) |
|---------|----------|----------|
| `status` | Loop ~2s until SIGINT | **Default**: one snapshot, exit `0`. **`--follow` / `-f`**: v1 loop behavior |
| `status` TTY | Clear screen each refresh | Clear **only** in follow mode on TTY |
| `stop` | Prompt on stdin | **Required arg**: `<instance-id>`; no stdin |
| `stop` empty input | Exit `0` abort | **Removed** — missing arg → usage exit `2` |

Commands not listed here (`start`, hidden `worker`) remain as 001.

## Commands (operator-facing)

### `supervisor status`

| Property | Value |
|----------|-------|
| Args | None |
| Flags | `--follow`, `-f` (boolean; default **false**) |
| Default output | Single table: `ID`, `STATE`, `CPU%`, `MEMORY` |
| Default TTY | No ANSI screen clear (FR-001, clarification 2026-11-09) |
| Follow output | Same columns; ~2s refresh until SIGINT |
| Follow TTY | Full-screen clear between refreshes when stdout is a TTY |
| States / metrics | Same as 001 (`starting` → `unavailable` metrics) |
| Reconcile | Each render (once for default; per cycle for follow) |
| Empty | `No instances are running.` then exit `0` |
| Default exit | `0` immediately after print |
| Follow exit | `0` after interrupt |
| Registry I/O error | Non-zero (general error class) |

### `supervisor stop`

| Property | Value |
|----------|-------|
| Args | **Exactly one**: `<instance-id>` |
| Flags | None |
| Interaction | None (stdin ignored) |
| Missing / extra args | Usage error; exit `2`; no registry change (FR-006) |
| Whitespace-only ID | Usage error; exit `2` |
| Reconciliation | Same as 001 before lookup |
| Success (live stop) | Exit `0`; row removed |
| Unknown ID | Exit `0`; `instance not found: <id>` |
| Already stopped | Exit `0`; `instance already stopped: <id>` |
| Live stop failure | Exit `1` |

**Invocation**:

```text
supervisor stop <instance-id>
```

## Exit codes (v2 delta)

| Code | Meaning |
|------|---------|
| 0 | Unchanged success paths (including stop not-found / already-stopped) |
| 1 | Unchanged general errors |
| 2 | Usage: Cobra validation, including **stop without ID**, extra stop args, invalid start flags |

**Removed**: exit `0` for interactive “stop aborted” / empty stdin (001 FR-006a).

## Environment / platform

Unchanged from 001 (`XDG_*`, Linux and WSL2).

## Compatibility

- Scripts should use `supervisor status` for snapshots and `supervisor status -f` for monitoring.
- Scripts must use `supervisor stop "$id"` instead of piping stdin.
- 001 contract remains historical reference for `start` and registry.
