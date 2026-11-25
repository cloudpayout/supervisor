# Data Model: Lightweight Browser Instances (013)

**Spec**: [spec.md](./spec.md)

## Registry

### Document

| Field | Type | Change |
|-------|------|--------|
| `schema_version` | int | **1** default; bump to **2** on first write of any instance with `memory_limit_bytes` **or** explicit 013 `lightweight` on a new instance row (read accepts 1 or 2; v1 rows without 013 fields remain valid) |
| `instances` | []Instance | — |

### Instance (extensions)

| Field | JSON | Type | Required | Validation |
|-------|------|------|----------|------------|
| Memory limit | `memory_limit_bytes` | uint64 | No | >0 when present; set only when `start --memory-limit` parsed |
| Threshold | `memory_threshold_percent` | int | No | 1–100 when present; default **100** at write time if limit set and `--threshold` omitted |
| Lightweight profile | `lightweight` | bool | No* | **`true`**: apply headed-safe (headed) or full bundle (headless). **`false`**: `--no-lightweight` — no 013 allocator extras. *013 `start` MUST write explicit `true` or `false`. |

Existing fields (`id`, `state`, `headless`, `casino`, `entry`, PIDs, etc.) unchanged.

### Warning line (derived, not stored)

```
warning_bytes = floor(memory_limit_bytes * memory_threshold_percent / 100)
```

Compare to `metrics.Aggregate(...).MemoryRSS` for NOTE and tests.

## CLI start inputs (transient → persisted)

| Flag | Parsed type | Persisted |
|------|-------------|-----------|
| `--memory-limit` | suffix size → bytes | `memory_limit_bytes` |
| `--threshold` | int percent | `memory_threshold_percent` |
| `--no-lightweight` | presence bool | `lightweight: false` (omit flag ⇒ `lightweight: true`) |

Invalid: `--threshold` without `--memory-limit`; `MB`/`GB` suffixes; non-positive; bad suffix.

**Independence**: `--no-lightweight` does not clear or block memory fields; both may be set on one row.

## Allocator inputs (worker, from registry)

| Registry | Allocator tier |
|----------|----------------|
| `headless=true`, `lightweight=true` | Append `headlessLightweightOptions()` |
| `headless=false`, `lightweight=true` | Append `headedSafeLightweightOptions()` only |
| `lightweight=false` | Skip all 013 lightweight option slices |

## Status row (ephemeral view model)

| Column | Source |
|--------|--------|
| ID, STATE, ENTRY | registry.Instance |
| CPU% | metrics.Aggregate |
| MEMORY | format per [research.md](./research.md) R7 |
| NOTE | empty, or `threshold exceeded` when RSS ≥ warning_bytes and state `running` |

Lightweight tier is **not** shown as a status column in v1 (registry + tests only).

## Baseline artifact (plan-local, not registry)

Stored in `research.md` or `docs/baseline-013.md` at implementation time:

| Field | Meaning |
|-------|---------|
| `host` | machine id / uname |
| `chrome` | resolved binary |
| `pre_median_mib` | baseline median idle |
| `post_median_mib` | after lightweight profile |
| `samples` | count, window |

Measurement for SC-001: default `start --headless` with `lightweight=true` (no `--no-lightweight`).

## Lightweight profile (registry-backed)

Chrome allocator option sets documented in [contracts/browser-allocator.md](./contracts/browser-allocator.md); effective set = f(`headless`, `lightweight`).