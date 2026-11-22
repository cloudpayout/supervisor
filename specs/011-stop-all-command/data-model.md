# Data Model: Stop All Instances Command

**Feature**: `011-stop-all-command` | **Date**: 2026-11-22  
**Registry (unchanged)**: [001 registry schema](../001-supervisor-cli/contracts/registry-schema.json) and [001 data model](../001-supervisor-cli/data-model.md)

No new fields in `instances.json`. Bulk stop reuses existing `Instance` records and reconciliation rules.

## Persistent model

Identical to 001–010. States `starting`, `running`, `stopping`, and reconciliation (60s `starting`, split-brain, dead PIDs) apply on:

| Trigger | Reconcile |
|---------|-----------|
| `stop <id>` | Before lookup (unchanged) |
| `stop all` | Once at bulk start (snapshot), then again per `stopInstance` call |

## Runtime / CLI entities

### Bulk stop request (ephemeral)

| Field | Rule |
|-------|------|
| `token` | Sole positional arg must equal `all` (case-sensitive) |
| `stdin` | Not used (FR-002) |
| `reserved` | `all` is not a valid opaque instance ID (spec assumption) |

### Stop target set (ephemeral)

| Field | Source |
|-------|--------|
| `ids` | Copy of `Instance.ID` for every row in registry document **after** first `registry.Apply` in bulk entry |
| Order | Document instance slice order at snapshot time |
| Stability | Fixed for this invocation; not refreshed mid-loop |

### Per-instance outcome (ephemeral)

Same enumeration as single-ID stop (002 data model), aggregated for bulk:

| Outcome | Registry effect | Stderr line | Bulk counter |
|---------|-----------------|-------------|--------------|
| Live stop success | Row removed | `stopped instance <id>` | `stopped++` |
| Not found | None | `instance not found: <id>` | `skipped++` |
| Already stopped | Row pruned if stale | `instance already stopped: <id>` | `skipped++` |
| Operational failure | Row may remain | Same distinct failure text as single-ID stop exit `1` | Set `failed` flag; **continue** loop; aggregate exit `1` at end |
| Registry I/O error on bulk snapshot | No per-ID loop | Error to stderr | Return error immediately (fail before loop) |

**Operational failure (FR-005)**: `stopInstance` returns non-nil `error` after a live row was present—registry update failure after terminate, or any path where single-ID `stop` would exit `1`. Not-found and already-stopped return `nil` error with outcome counters only.

### Bulk summary (ephemeral)

| Field | Rule |
|-------|------|
| `stopped_count` | Live stops succeeded |
| `skipped_count` | Not-found + already-stopped from snapshot |
| Line format | `stop all: stopped N, skipped M (already stopped or not found)` |

### Empty bulk result

When `len(ids)==0` after initial reconcile: message `No instances are running.` on stderr; no further stops; exit `0` unless initial `store.Update` failed.

## State transitions

No new transitions. Each live stop still drives terminate → profile removal → row removed, identical to `stop <id>`.

## Validation rules (CLI input)

| Input | Rule |
|-------|------|
| `stop` args | Exactly one non-blank token |
| Token `all` | Bulk mode |
| Token other than `all` | Single-ID mode (002 contract) |
| Token `All` / `ALL` | Treated as opaque ID → not found unless such ID exists |
| Extra args to `stop` | Two or more positionals (e.g. `stop all extra`) → Cobra usage error, exit `2`; no stops run |

## Interrupt handling

| Event | Behavior |
|-------|----------|
| SIGINT before next ID | Stop scheduling; finish current `stopInstance` |
| Exit | `1` with interrupted summary line |
| Registry | Reflects only fully completed stops |