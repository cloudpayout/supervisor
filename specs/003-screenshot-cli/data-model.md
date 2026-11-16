# Data Model: Screenshot CLI (003)

**Date**: 2026-11-15

**Registry (unchanged)**: [001 registry schema](../001-supervisor-cli/contracts/registry-schema.json) and [001 data model](../001-supervisor-cli/data-model.md)

No new fields in `instances.json`. `schema_version` stays `1`. This feature adds an ephemeral socket and an operator-owned image file.

## Persistent model

Identical to 001 and 002:

- `Registry document` → `Instance` records (`id`, `state`, PIDs, `profile_dir`, timestamps).

`screenshot` does not write the registry and does not run reconciliation. `status` and `stop` keep their current triggers.

## Runtime entities

### Screenshot request (ephemeral)

One CLI invocation.

| Field | Rule |
|-------|------|
| `instance_id` | Required positional argument. Non-empty after trim. Must be a single path segment before it is used in a filesystem path (`filepath.Base(id) == id`). |
| `output_path` | Optional `--output` / `-o`. Empty means the default file name. Relative paths resolve against the current directory. Absolute paths are used as given. |
| `stdin` | Not read |

### Capture socket (ephemeral)

Not stored in the registry.

| Field | Rule |
|-------|------|
| Path | `{state dir}/supervisor/sockets/<id>.sock` via `Layout.CaptureSocket` |
| Lifetime | Created after the worker sets `state` to `running`. Removed when the worker shuts down. |
| Mode | Directory `0700`, socket `0600` |
| Relationship | One socket per running instance. The instance ID is the file name. |

### Screenshot file (operator output)

Not stored in the registry.

| Field | Rule |
|-------|------|
| Bytes | One PNG of that instance's current viewport |
| Default name | `screenshot-<instance-id>-<YYYYMMDD-HHMMSS>.png` in the current directory, local time. Collision suffix `-2`, `-3`, … before `.png` |
| Chosen name | The `--output` path, only when that final path does not already exist |
| Publish | Temporary file in the destination directory, then `os.Link` to the final name. The temporary file is removed on success and on failure |
| Sensitivity | May show whatever the page shows. Supervisor does not store, redact, or upload it |

## State transitions

No new instance states. Capture is allowed only from `running`.

```text
starting ──screenshot──► refused (not running), row unchanged
running  ──screenshot──► running (same page); PNG published or error with no final file
stopped / row removed ──screenshot──► not found when no row; not running when a non-running row remains
```

A capture failure does not move `running` to another state and does not remove the row.

## Validation rules

| Input | Rule |
|-------|------|
| Argument count | Exactly one positional ID |
| Blank ID or blank `--output` | Usage error, no file, no socket dial |
| ID with a path separator or `..` | Do not join it into the socket directory or the default file name |
| No registry row | Not found, no file |
| `state != running` | Not running, no file |
| Destination exists, is a directory, or parent is missing | Fail before capture, no file created, existing content unchanged |
| PNG publish fails | No final file, success path not printed |
| Other instances | Not modified by this command |

## Relationships

```text
Instance (1) ──while running──► (1) Capture socket
Screenshot request (1) ──targets──► (1) Instance
Screenshot request (1) ──publishes──► (0..1) Screenshot file
```
