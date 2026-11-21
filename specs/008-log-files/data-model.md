# Data Model: Instance Log Files

**Feature**: `008-log-files`

## Storage layout (FR-001, FR-001a)

| Path element | Rule |
|--------------|------|
| State home | `XDG_STATE_HOME` or `~/.local/state` |
| `{supervisor_state}` | `{state_home}/supervisor` |
| Registry | `{supervisor_state}/instances.json` (unchanged) |
| Logs directory | `{supervisor_state}/logs/` |
| Instance log | `{supervisor_state}/logs/{instance-id}.log` |

**Invariants**:

- `{instance-id}` is a single path-safe segment (existing ID generator).
- `log_path` on registry `Instance` MUST equal `Layout.LogPath(id)` at registration.
- Teardown MUST NOT delete `*.log` (FR-009b).

## Registry instance (amends 001)

No new JSON fields. Existing field usage enforced:

| Field | Role in this feature |
|-------|----------------------|
| `log_path` | Open first when tracked; must match canonical path at `start` |
| `id` | Filename stem for canonical path and `logs` argument |

## Structured log entry (on disk)

| Part | Constraint |
|------|------------|
| Timestamp | RFC 3339 UTC with `Z` suffix; fractional seconds allowed |
| Level | `ERROR` \| `NOTICE` \| `INFO` |
| Message | Remainder of line; no `\n`; secrets redacted at write time |

**Line grammar** (informal):

```text
TIMESTAMP SP LEVEL SP MESSAGE_EOL
```

Unstructured line: any file line not matching the grammar (pre-feature logs).

## Severity filter (read path)

| Flag value | Emits levels |
|------------|----------------|
| (none) | Structured: all; unstructured: included |
| `error` | `ERROR` only |
| `notice` | `ERROR`, `NOTICE` |

Unstructured lines: excluded whenever `--level` is set.

## Orphaned log (derived)

| Attribute | Source |
|-----------|--------|
| `instance_id` | basename of `*.log` without extension |
| `last_activity_utc` | file `ModTime` in UTC |
| `last_error_summary` | message from last structured `ERROR` line, or placeholder `no error recorded` |

**Orphaned** iff `instance_id` ∉ current registry IDs (after reconcile read).

## Log view modes (operator)

| Mode | Args | Output |
|------|------|--------|
| Snapshot | `logs <id>` | File lines to stdout |
| Follow | `logs <id> -f` | Snapshot then tail |
| Filtered | `logs <id> --level L` | Passthrough lines matching filter |
| Recent | `logs --recent` | ≤50 orphaned rows + optional omitted count |

## Lifecycle

```text
start → create/append log file → worker structured + INFO-wrapped capture
     → logs reads while tracked (registry path)
crash/stop/failure → registry row removed → log file remains
     → logs reads canonical path; --recent lists orphan
```

## Secret redaction

Cookie/session values from the session file MUST NOT appear in structured messages or wrapped INFO lines. Redaction applies at write time (ingest and explicit `Error`/`Notice`/`Info` call sites).