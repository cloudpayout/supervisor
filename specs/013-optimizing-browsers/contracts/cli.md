# CLI Contract: optimizing browsers (013)

**Version**: v13 (013-optimizing-browsers)  
**Amends**: [002 status](../../002-status-stop-cli/contracts/cli.md), [005 start](../../005-start-command-syntax/contracts/cli.md)  
**Binary**: `supervisor`

## Delta summary

| Area | Prior | 013 |
|------|-------|-----|
| Default headless Chrome | WebGL + Spribe overrides | Adds **full lightweight allocator profile** when lightweight on (see [browser-allocator.md](./browser-allocator.md)) |
| Default headed Chrome | Spribe/shared flags | **Headed-safe** lightweight only (`force-device-scale-factor=1`) when lightweight on |
| Lightweight opt-out | — | **`--no-lightweight`** skips **all** 013 lightweight extras (headed or headless) |
| `start` flags | `--headless`, session path | **`--memory-limit`**, **`--threshold`**, **`--no-lightweight`** |
| `status` columns | ID, STATE, ENTRY, CPU%, MEMORY | + **NOTE**; MEMORY may be `current MiB/limit MiB` |
| Memory units | MiB display for current | **MiB only**; limits binary suffixes |
| Registry | schema 1 | schema **2** when 013 fields written; includes **`lightweight`** |

## `supervisor start`

### New flags

| Flag | Type | Required with | Semantics |
|------|------|---------------|-----------|
| `--no-lightweight` | bool (presence) | — | When set, instance starts with **no** 013 lightweight allocator extras. Registry `lightweight: false`. Default: lightweight **on** (`lightweight: true`) — full bundle if `--headless`, headed-safe tier if headed. Independent of memory flags. |
| `--memory-limit` | string | — | Max memory (mebibytes/gibibytes): `M`, `MiB`, `G`, `GiB`. Rejects `MB`, `GB`, bare numbers. |
| `--threshold` | int 1–100 | `--memory-limit` | Warning at `threshold%` of limit. Omitted with limit ⇒ **100**. |

**Persistence**: `memory_limit_bytes`, `memory_threshold_percent` (when limit set), **`lightweight`** on registry row (schema v2).

**Failures**: Invalid suffix/combination → exit non-zero, **no registry row** (FR-010, FR-012).

### Defaults (lightweight)

| Invocation | `headless` on row | `lightweight` | Allocator tier |
|------------|-----------------|---------------|----------------|
| `start --headless` | true | true | Full headless lightweight bundle |
| `start --headless --no-lightweight` | true | false | Pre-013-style headless (no 013 bundle) |
| `start` (headed) | false | true | Headed-safe only |
| `start --no-lightweight` | false | false | No 013 lightweight extras |

SC-001 baseline and default FR-008 regression: first row only.

### Unchanged

Session path, headed/headless mutual semantics, start failure / ID rules.

## `supervisor status` / `status --follow`

### Header

```text
ID    STATE    ENTRY    CPU%    MEMORY    NOTE
```

### MEMORY

| Instance | Format | Example |
|----------|--------|---------|
| No limit, running | one decimal + MiB | `245.3 MiB` |
| Limit, running | current + limit MiB | `245.3 MiB/512.0 MiB` |
| Limit, starting | unavailable/limit | `unavailable/512.0 MiB` |
| No limit, starting | `unavailable` | — |

### NOTE

| Condition | Value |
|-----------|-------|
| Default | *(empty)* |
| RSS ≥ warning line, running | `threshold exceeded` |
| starting / metrics unavailable | *(empty)* |

**Comparison**: `metrics.Aggregate` RSS bytes vs `warning_bytes` ([data-model.md](../data-model.md)).

## Help text (FR-009)

- `start --help`: document **`--no-lightweight`**, default lightweight behavior for headed vs headless, new memory flags, MiB/GiB grammar, default threshold 100%, independence of lightweight vs memory limits.
- `status --help`: NOTE column + threshold message.

## Exit codes

No change to global exit code tables except new **usage errors** for bad `--memory-limit` / `--threshold` on `start`.