# CLI Contract: supervisor watch (018)

**Version**: v1 (018-instance-watch) — adds a command. Does not change [003](../../003-screenshot-cli/contracts/cli.md) `screenshot` or earlier `start`, `status`, `stop`, `logs`.

**Internal peer**: [watch-socket.md](./watch-socket.md)

## Command set delta

| Command | Before | After |
|---------|--------|-------|
| `watch` | Absent | **New**. Required `<instance-id>`. Optional `--fps` |

## `supervisor watch`

| Rule | Behavior |
|------|----------|
| ID | Exactly one positional argument. No prompt. Blank or extra arguments are a usage error (exit 2): `instance ID required` |
| `--fps` | Optional integer. Default `1`. Allowed `1`–`30`. Anything else is a usage error (exit 2): `invalid --fps` |
| Lookup | `View` only. No row → stderr `instance not found: <id>`, exit 1. Row not `running` → stderr `instance not running: <id>`, exit 1 |
| Success start | Stdout is one line: `http://127.0.0.1:<port>/`. Exit stays open |
| Success end | Operator interrupt, or the instance is no longer running: exit `0`. No extra stdout |
| In progress failure | After dial, worker `ERR` or no `OK` line: stderr `watch failed: <id>: <reason>`, exit 1. No URL line |
| Second watch | Same ID while one watch is active: `watch failed: <id>: already watching`, exit 1. The first watch continues |
| Side effects | Does not stop, navigate, resize, or replace the instance. Does not write a file |
| Concurrent screenshot | Allowed. Screenshot contract is unchanged |

**Help** (`watch --help`) must describe the required ID, `--fps` and its default and maximum, that the printed URL is loopback-only, and that interrupting the command ends the watch.

**Root help** must list `watch` with a short description of a live local view.

## Examples

```text
supervisor watch <instance-id>
supervisor watch <instance-id> --fps 1
supervisor watch <instance-id> --fps 30
```

## Exit codes

| Code | When |
|------|------|
| 0 | URL was printed, then the operator interrupted watch or the instance ended |
| 1 | Not found, not running, watch failed, or a second watch |
| 2 | Usage: arguments, flags, invalid `--fps` |

## Compatibility

- Workers started before this feature have no `WATCH` handler. `watch` fails with `watch failed` if the row is `running` but the worker rejects the verb. Operators restart those instances. New `start` invocations listen for both `CAPTURE` and `WATCH`.
- Platforms other than Linux and WSL2 are out of scope.
