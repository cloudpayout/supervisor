# Data Model: Status and Stop CLI Behavior (002)

**Date**: 2026-11-09  
**Registry (unchanged)**: [001 registry schema](../001-supervisor-cli/contracts/registry-schema.json) and [001 data model](../001-supervisor-cli/data-model.md)

No new fields in `instances.json`. This feature only changes **how** the CLI reads and acts on existing records.

## Persistent model

Identical to 001:

- `Registry document` → `Instance` records (`id`, `state`, PIDs, `profile_dir`, timestamps).

Reconciliation rules (60s `starting`, split-brain, dead PIDs) still run on:

| Trigger | Reconcile + display |
|---------|---------------------|
| `status` (default) | Once per invocation |
| `status --follow` | Each refresh cycle (~2s) |
| `stop <id>` | Before lookup (unchanged) |

## Runtime / CLI entities

### Status snapshot (ephemeral)

| Field | Source |
|-------|--------|
| Rows | Registry instances in `starting` or `running` after `registry.Apply` |
| CPU / memory columns | `internal/metrics` aggregate; `unavailable` while `starting` |
| `clearScreen` | `false` (default status); `isTTY(stdout)` when `--follow` |

One snapshot is produced per one-shot exit; follow mode produces a sequence of snapshots until interrupt.

### Stop request (ephemeral)

| Field | Rule |
|-------|------|
| `instance_id` | Required positional arg; non-empty after trim |
| `stdin` | Not used (FR-005) |

Outcomes unchanged from 001: live stop (remove row), not-found, already-stopped, operational failure.

## State transitions

No new transitions. Stop by ID still drives `running` → terminate → row removed; reconciliation paths unchanged.

## Validation rules (CLI input)

| Input | Rule |
|-------|------|
| `status` args | None accepted (usage error if extras) |
| `status --follow` | Compatible with no args only |
| `stop` args | Exactly one non-blank `instance_id` |
| Unknown `instance_id` | Not-found message, exit `0` (FR-007) |
