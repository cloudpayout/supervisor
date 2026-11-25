# Data Model: Aggressive Chrome RAM Reduction (015)

**Spec**: [spec.md](./spec.md)

## Scope

No new registry fields, CLI flags, or `status` columns. Phase-3 affects only Chrome launch options for `headless && lightweight`, appended after 013 + 014 slices.

## Entities (unchanged from 013/014)

| Entity | 015 change |
|--------|------------|
| `lightweight` registry field | Unchanged; `true` still enables full headless tier **including** phase-3 when implemented |
| Lightweight profile | Semantic extension: 013 + 014 + optional 015 adopt slice |
| Memory policy / NOTE | Unchanged (013) |

## Allocator tiers (updated semantics)

| Condition | Chrome extras |
|-----------|----------------|
| `headless && lightweight` | … + `headlessPhase3LightweightOptions()` **if** research adopt list non-empty |
| `headed && lightweight` | Headed-safe only; **no** 015 |
| `!lightweight` | No 013/014/015 slices |

## Baseline artifact

| Field | Location |
|-------|----------|
| 014 vs 015 medians | [research.md](./research.md) Baseline Record (015) |