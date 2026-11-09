# Data Model: Supervisor CLI (001-supervisor-cli)

**Date**: 2026-11-09  
**Registry contract**: [contracts/registry-schema.json](./contracts/registry-schema.json)

## Registry document (`instances.json`)

Top-level object with schema version and instance list.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `schema_version` | integer | yes | Registry format version; start at `1` |
| `instances` | array | yes | All tracked instances (including `starting`) |

### Instance record

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Opaque unique ID (stable for run lifetime) |
| `state` | enum | yes | `starting` \| `running` \| `stopping` \| `stopped` |
| `created_at` | string (RFC3339) | yes | Registration time |
| `profile_dir` | string | yes | Absolute path to Chrome user-data directory |
| `owner_pid` | integer | no | Worker process PID; set when worker starts |
| `browser_pid` | integer | no | Root Chrome PID for metrics tree; set after launch |
| `log_path` | string | no | Worker log file path |
| `headless` | boolean | no | `true` if launched with `--headless`; omitted or `false` for headed (FR-001d) |
| `last_seen_at` | string (RFC3339) | no | Updated by worker heartbeat or status reconciliation |

### Validation rules

- `id` MUST be unique within `instances`.
- While `state` is `running`, `owner_pid` and `browser_pid` MUST be present and > 0.
- `profile_dir` MUST be exclusive to that `id` (no two records share the same path).
- Records in `stopped` SHOULD be removed from the array on write (registry contains only active or `starting`/`stopping`).

## Runtime entities (in-memory)

### Instance (aggregate)

Mirrors registry record plus ephemeral **resource snapshot**:

| Field | Source |
|-------|--------|
| CPU percent | gopsutil aggregate |
| Memory bytes | gopsutil aggregate RSS |
| Metrics status | `ok` \| `unavailable` |

### Resource snapshot

Point-in-time metrics for display in `status`; not persisted to JSON (recomputed each refresh).

## State transitions

```text
(start CLI)     starting ──worker publishes PIDs──► running
(running)       running ──stop SIGTERM────────────► stopping ──reap──► (removed)
(running)       running ──worker/browser exit────► (reconcile remove)
(starting)      starting ──worker fail────────────► (remove, no phantom ID)
(starting)      starting ──>60s from created_at──► (remove, cleanup FR-013)
```

### Reconciliation rules

Triggered on `status` refresh, `stop`, and worker exit:

1. If `state` is `starting` and `created_at` + **60s** elapsed → remove as failed start; tear down owner/Chrome if present (FR-013).
2. If owner dead but Chrome alive for that profile → remove row; terminate orphan Chrome tree (FR-009a).
3. If Chrome dead but owner alive → remove row; terminate owner (FR-009b).
4. If both dead while row remains → remove row (FR-007a, FR-009).
5. Duplicate `id` MUST NOT be written (lock prevents concurrent corruption).

## Relationships

```text
Registry (1) ──contains──► (N) Instance records
Instance (1) ──has──► (1) Worker process
Instance (1) ──has──► (1) Chrome profile directory
Instance (1) ──has──► (1) Chrome process tree (root browser_pid)
```

## IDs and paths

- **ID generation**: cryptographically random string; collision re-roll under lock.
- **Profile path**: `${DATA_ROOT}/supervisor/profiles/${id}/` created before worker spawn.
- **No secrets** in registry (constitution); credentials are future runtime-only concerns.
