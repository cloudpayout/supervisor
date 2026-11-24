# Data Model: Aggressive RAM Experiment (016)

**Spec**: [spec.md](./spec.md)

## Scope

No new registry fields, CLI flags, or `status` columns (**FR-004**, **FR-005**). Behavior is entirely in Chrome launch options and `start` help text.

## Entities (unchanged shape)

| Entity | 016 change |
|--------|------------|
| `lightweight` registry field | Unchanged; `true` → 013+014+015+**016 experimental** on headless; `false` → performance-neutral profile (**FR-001**) |
| `headless` | Unchanged |
| Memory policy / NOTE | Unchanged (**013**) |

## Allocator semantics (updated)

| Condition | Chrome lightweight / experiment |
|-----------|----------------------------------|
| `headless && lightweight` | 013 + 014 + 015 + **016 experimental**; headless base **without** WebGL enable when lightweight (research R1) |
| `headless && !lightweight` | **No** 013/014/015/016; no RAM tuning flags ([no-lightweight-neutral.md](./contracts/no-lightweight-neutral.md)) |
| `headed && lightweight` | Headed-safe 013 only; **no** 016 |
| `headed && !lightweight` | No headed-safe scale; **no** 016 |

## Baseline artifact

| Field | Location |
|-------|----------|
| 015 vs 016 medians, stand-ins | [research.md](./research.md) Baseline Record (016) |