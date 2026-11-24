# Data Model: Further Chrome RAM Reduction (014)

**Spec**: [spec.md](./spec.md)

## Scope

**No new registry fields, CLI flags, or status columns.** Phase-2 affects only Chrome launch options for `headless && lightweight` (same as 013 tier gate).

## Entities (unchanged from 013)

| Entity | 014 change |
|--------|------------|
| Registry `lightweight` | Unchanged; `true` still enables headless full tier **including** phase-2 slice |
| `Lightweight profile` | Semantic extension: 013 bundle + 014 slice (documented in contracts) |
| Resource snapshot / memory policy | Unchanged (013 `status` NOTE, `--memory-limit`) |

## Allocator tiers (updated semantics)

| Condition | Chrome extras |
|-----------|----------------|
| `headless && lightweight` | Spribe/WebGL baseline + `headlessUnset*` + 013 `headlessLightweightOptions` + **014 `headlessPhase2LightweightOptions`** |
| `headed && lightweight` | Headed-safe only (013); **no** 014 |
| `!lightweight` | No 013 or 014 slices |

## Baseline artifact

| Field | Location |
|-------|----------|
| 013 vs 014 medians | [research.md](./research.md) Baseline Record (014) |