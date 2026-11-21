# CLI Contract: supervisor (008 amendment)

**Version**: v8 (008-log-files) — amends operator catalog after 006/007  
**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

## Delta summary

| Item | Prior | v8 (008) |
|------|-------|----------|
| Operator commands | `start`, `status`, `stop`, `screenshot`, `help` | Adds **`logs`** |
| Worker log on disk | Raw stdout/stderr | Structured lines + INFO-wrapped capture (FR-002b) |
| Log retention | Implicit | Files kept after registry removal (FR-009b) |

Commands not listed here remain unchanged.

## `supervisor logs`

| Property | Value |
|----------|-------|
| Modes | **A**: `logs <instance-id>` **B**: `logs --recent` (mutually exclusive with positional ID) |
| Args | Zero args with `--recent`, or exactly one `<instance-id>` |
| Flags | `--follow`, `-f` (instance mode only); `--level error\|notice` (instance mode only) |
| stdin | Ignored |
| Missing args | Usage error; exit `2` |
| `logs --recent` + ID | Usage error; exit `2` |

### Instance mode: `logs <instance-id>`

| Condition | stdout | stderr | Exit |
|-----------|--------|--------|------|
| Success, tracked | Log file lines (filtered if `--level`) | (empty) | `0` |
| Success, not tracked | Log file lines | One-line notice: instance not in `status` | `0` |
| Log not found | (empty) | `log not found: <id>` (or equivalent) | non-zero (general) |
| Registry I/O error | (empty) | `supervisor: …` | `1` |
| Follow + SIGINT | Printed lines so far | (empty) | `0` |

**Filtering**: When `--level` set, only structured lines matching spec grammar with minimum severity; each printed line is exact on-disk bytes including timestamp and level.

### Discovery mode: `logs --recent`

| Condition | stdout | Exit |
|-----------|--------|------|
| Orphans exist | Table: `ID`, `LAST_ACTIVITY_UTC`, `LAST_ERROR` (≤50 rows, newest activity first) | `0` |
| >50 orphans | Same + notice line with omitted count | `0` |
| No orphans | Empty-state message | `0` |
| Logs dir missing/empty | Empty-state message | `0` |

`LAST_ERROR`: last structured `ERROR` message, or `no error recorded`.

## Log file format (persisted)

```text
2006-01-02T15:04:05.123Z LEVEL message text
```

`LEVEL` is `ERROR`, `NOTICE`, or `INFO`.

## Path resolution

| Instance state | File opened |
|----------------|-------------|
| Tracked | Registry `log_path`; if unreadable, `{state}/supervisor/logs/<id>.log` |
| Not tracked | Canonical path only |

`{state}/supervisor/logs/` is the only directory scanned for `--recent`.

## Environment / platform

Unchanged: `XDG_STATE_HOME`, Linux and WSL2.

## Exit codes (008 delta)

| Code | Meaning |
|------|---------|
| 0 | Successful read, empty recent list, follow interrupted |
| 1 | Registry/log I/O failure (same class as status) |
| 2 | Usage: missing ID, both `--recent` and ID, invalid `--level` |

## Compatibility

- Operators use `logs <id>` instead of `cat` on log paths.
- Scripts use `logs --recent` to recover IDs after failed runs.
- Pre-feature unstructured log files: `logs <id>` without `--level` still prints raw lines.