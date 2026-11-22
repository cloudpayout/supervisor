# CLI Contract: supervisor (011 amendment)

**Version**: v11 (011-stop-all-command) — amends [010](../010-logs-clear/contracts/cli.md) / [002 stop](../002-status-stop-cli/contracts/cli.md)  
**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

## Delta summary

| Item | Prior (002/010) | v11 (011) |
|------|-----------------|-----------|
| `supervisor stop <instance-id>` | Required single ID | **Unchanged** when arg ≠ `all` |
| `supervisor stop all` | N/A (informal teardown only) | **New** — bulk stop all snapshotted registry IDs |
| Reserved token | — | `all` (case-sensitive) triggers bulk, not ID lookup |

## Commands (operator-facing)

### `supervisor stop` (single ID — unchanged from 002)

| Property | Value |
|----------|-------|
| Args | Exactly one: `<instance-id>` where ID ≠ `all` |
| Flags | None |
| Interaction | None |
| Outcomes | Same as [002 contract](../002-status-stop-cli/contracts/cli.md) |

### `supervisor stop all`

| Property | Value |
|----------|-------|
| Args | Exactly one: literal `all` |
| Flags | None |
| Interaction | None (stdin ignored) |
| Snapshot | After one `registry.Apply`, copy all instance IDs; iterate snapshot only |
| Per ID | Same terminate / profile / registry removal as single stop |
| Stderr (per ID) | Same lines as single stop (`stopped instance`, `instance not found`, `instance already stopped`, failures) |
| Stderr (summary) | `stop all: stopped N, skipped M (already stopped or not found)` after loop |
| Empty snapshot | `No instances are running.` then exit `0` |
| Extra positionals | `stop` accepts exactly one arg: `supervisor stop all` is valid; `supervisor stop all extra` or `supervisor stop id1 id2` → usage exit `2`, no stops |
| Success | Exit `0` when no operational failures (including empty and all-skipped) and no SIGINT |
| Partial failure | Continue remaining snapshot IDs; exit `1` if any `stopInstance` returned operational error (same class as single-ID stop failure) |
| Interrupt (SIGINT) | Stop scheduling further IDs; complete current stop; `stop all interrupted: completed K of L`; exit `1` |
| Registry lock | Same `store.Update` / flock rules as single stop |

**Invocation**:

```text
supervisor stop all
```

## Exit codes (v11 delta)

| Code | Meaning |
|------|---------|
| 0 | Unchanged; bulk success including empty registry and all skipped |
| 1 | Unchanged general errors; **adds** bulk incomplete (any operational `stopInstance` error or SIGINT before full snapshot processed) |
| 2 | Unchanged usage (missing arg, extra args, invalid flags) |

## Help

- `supervisor stop --help` MUST document both `<instance-id>` and `all`.
- MUST NOT describe interactive prompting.

## Environment / platform

Unchanged from 001 (Linux, WSL2, `XDG_*`).

## Compatibility

- Scripts: prefer `supervisor stop all` over parsing `status` for N× `stop <id>`.
- `supervisor stop all` is distinct from `supervisor logs clear` (log files only, no registry stop).