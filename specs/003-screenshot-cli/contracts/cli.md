# CLI Contract: supervisor screenshot (003)

**Version**: v3 (003-screenshot-cli) — adds a command; does not change [002](../../002-status-stop-cli/contracts/cli.md) `start`, `status`, or `stop`

**Binary**: `supervisor`

**Module**: `github.com/cloudpayout/supervisor`

**Internal peer**: [capture-socket.md](./capture-socket.md)

## Delta summary

| Command | Before | v3 |
|---------|--------|----|
| `screenshot` | Absent | **New**. Required `<instance-id>`. Optional `--output` / `-o` |
| `start`, `status`, `stop`, hidden `worker` | 002 / 001 | Unchanged |

## `supervisor screenshot`

| Property | Value |
|----------|-------|
| Args | Exactly one: `<instance-id>` |
| Flags | `--output` / `-o` path (optional). Relative or absolute |
| Interaction | None. Standard input is ignored |
| Success stdout | One line: the output path. No other stdout |
| Success exit | `0` |
| Default file | `screenshot-<id>-<YYYYMMDD-HHMMSS>.png` in the current directory when `--output` is omitted. Local time. If that name exists, `-2`, `-3`, … before `.png` |
| Default stdout | The relative file name that was created |
| `--output` stdout | The trimmed flag value, unchanged |
| Missing / extra args | Usage error; exit `2`; no file |
| Blank ID | Usage error `instance ID required`; exit `2`; no file |
| Blank `--output` | Usage error; exit `2`; no file |
| Unknown ID | stderr `instance not found: <id>`; exit `1`; no file |
| State other than `running` | stderr `instance not running: <id>`; exit `1`; no file |
| Capture or write failure | stderr `screenshot failed: <id>: <reason>`; exit `1`; no final file; existing destination unchanged |
| Destination exists or is a directory | `screenshot failed` as above; exit `1`; no overwrite |
| Registry read | `View` only. No reconcile, no row write |
| Side effect on the instance | None. Page, size, and `running` state stay as they were |
| Deadline | Process exits with success or failure within 15 seconds |

**Invocation**:

```text
supervisor screenshot <instance-id>
supervisor screenshot <instance-id> --output path/to/page.png
supervisor screenshot <instance-id> -o /tmp/page.png
```

**Help** (`screenshot --help`) must describe the required ID, `--output` / `-o`, that the path may be relative or absolute, the default file location and name, and that the command does not prompt.

## Exit codes (v3 delta)

| Code | Meaning |
|------|---------|
| 0 | PNG published. Stdout is the path |
| 1 | Not found, not running, capture failure, or output-path failure. Also the existing general-error class |
| 2 | Usage: arity, blank ID, blank `--output`, unknown flag. Same class as 002 |

`stop` not-found remains exit `0`. `screenshot` not-found is exit `1`.

## Environment / platform

Unchanged from 001 (`XDG_STATE_HOME`, `XDG_DATA_HOME`, `HOME`). Official support remains Linux and WSL2.

The capture socket lives under the resolved state directory. See [capture-socket.md](./capture-socket.md). It is not an operator-facing path.

## Compatibility

- Scripts can take the path from stdout and the outcome from the exit code.
- `start`, `status`, and `stop` invocations from 002 keep working.
- Instances that were started by a worker without the capture listener fail this command with `screenshot failed` if their row is still `running`. Operators restart those instances to capture them. New `start` invocations after this feature listen automatically.
