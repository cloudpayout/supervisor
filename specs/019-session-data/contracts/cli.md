# CLI Contract: supervisor session (019)

**Version**: v19 (019-session-data) — adds `session info`. Prior commands unchanged except root help lists `session`.

**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

## Command set delta

| Command | Before | After |
|---------|--------|-------|
| `session` | Absent | **New** parent (subcommands only) |
| `session info` | Absent | **New** `<instance-id>` |

## `supervisor session info`

| Property | Value |
|----------|-------|
| Invocation | `supervisor session info <instance-id>` |
| Args | Exactly one `<instance-id>`. Zero or >1 → usage error. Blank after trim → usage error `instance ID required` |
| Flags | None in v1 |
| stdin | Ignored |
| Lookup | `store.Update` + `registry.Apply`, then `Find`. No row → stderr `instance not found: <id>`, exit `1` |
| States | Any tracked state after reconcile may succeed (not limited to `running`) |

### Success stdout

Exactly four lines, in order, then one newline:

```text
casino: onewin
game: aviator
strategy: null
date: 2026-11-26 09:15
```

| Line | Value |
|------|-------|
| `casino` | Registry `casino` or literal `null` |
| `game` | Registered game for that casino (catalog) or literal `null` |
| `strategy` | Always literal `null` in v19 |
| `date` | `created_at` as UTC `YYYY-MM-DD HH:MM` (colon between hour and minute) |

Stdout MUST NOT be JSON (no `{` or `[` as first non-space byte). Stderr empty on success.

### Failure stdout/stderr

| Condition | stdout | stderr | Exit |
|-----------|--------|--------|------|
| Usage | (empty) | Usage / `instance ID required` | `2` |
| Not found | (empty) | `instance not found: <id>` | `1` |
| Registry I/O | (empty) | `supervisor: …` | `1` |

### Help

| Command | Must describe |
|---------|----------------|
| Root help | `session` among operator commands |
| `session --help` | Subcommand `info` |
| `session info --help` | Required ID, four fields, `strategy` always `null`, example output block above |

## Examples

```text
supervisor session info <instance-id>
supervisor session info --help
```

## Exit codes

| Code | When |
|------|------|
| 0 | Row found; four lines printed |
| 1 | Not found, registry I/O failure |
| 2 | Usage: arguments |

## Compatibility

- No worker or registry schema change. Instances started before v19 are readable the same way once the binary includes `session info`.
- Platforms other than Linux and WSL2 are out of scope.