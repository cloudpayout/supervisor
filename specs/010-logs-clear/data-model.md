# Data Model: Clear Instance Log Files

**Feature**: `010-logs-clear` | **Date**: 2026-11-22

**Log layout**: [008 data-model](../../008-log-files/data-model.md)  
**Registry**: [001 registry](../../001-supervisor-cli/data-model.md) — **read-only** for this feature

## Scope

No new registry fields. Ephemeral classification of on-disk log files into **eligible** vs **protected** for deletion.

## Entity: Instance log file (on disk)

| Attribute | Rule |
|-----------|------|
| Path | `{supervisor_state}/logs/{instance-id}.log` |
| Identity | Basename without `.log` equals instance `id` |
| Deletion | Allowed only when **eligible** |

### Eligibility rules (FR-001a)

| Condition | Eligible for `logs clear` |
|-----------|---------------------------|
| No registry row for `id` | Yes (orphan) |
| Registry row `state` = `stopped` (if present) | Yes |
| Registry row `state` ∈ `starting`, `running`, `stopping` | **No** (protected) |

Protected files MUST remain after successful clear.

## Entity: Logs directory

| Field | Description |
|-------|-------------|
| `path` | `registry.Layout.LogsDir()` |
| `candidates` | Regular files matching `*.log` at top level only |

Non-`.log` files and subdirectories are out of scope for deletion.

## Entity: Clear operation (ephemeral)

| Field | Source |
|-------|--------|
| `eligible_count` | Glob + eligibility filter |
| `skipped_count` | Protected active lifecycle logs |
| `confirmed` | TTY prompt or `--yes` |

### Transition (filesystem only)

```text
[logs/*.log on disk]
  └── logs clear + confirm
        ├── eligible → removed
        └── protected (starting|running|stopping) → unchanged
[instances.json unchanged]
```

## CLI catalog delta

| Invocation | Effect |
|------------|--------|
| `supervisor logs clear` | Eligible log file deletion with confirmation |
| `supervisor clear` (009) | **Removed** from operator catalog in this feature |

## Relationships

```text
Operator
  └── logs clear [--yes]
        ├── Store.View → protected instance IDs
        ├── Glob → *.log candidates
        └── os.Remove → eligible paths only
```