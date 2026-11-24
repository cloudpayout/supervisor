# Contract: performance-neutral `--no-lightweight`

**Spec**: [spec.md](../spec.md) **FR-001**, **FR-002**  
**Package**: `internal/browser` `AllocatorOptions(profileDir, headless, lightweight)`

## Rule

When `lightweight == false`, the allocator MUST NOT append:

- `headlessLightweightOptions()` (013)
- `headlessPhase2LightweightOptions()` (014)
- `headlessPhase3LightweightOptions()` (015)
- `headlessExperimentalLightweightOptions()` (016)

Nor any other Chrome switch whose **documented purpose** in specs 013–016 is RAM or performance tuning.

## Headless `--no-lightweight` still includes

- Headless viewport, user-agent, automation tweaks
- `headlessUnset*` blocks (Spribe / bug **-4**)
- Casino `disable-features` base merge (without 014 lightweight feature additions when `!lightweight` — existing behavior)

## Headed `--no-lightweight`

- No `headedSafeLightweightOptions()` (no `force-device-scale-factor=1`)
- No experimental or lightweight RAM slices

## Verification

- `TestNoLightweightOmitsExperimentalAndLightweightTiers`: `AllocatorOptions(dir, true, false)` has **none** of the experiment six flags and **none** of `renderer-process-limit`, `aggressive-cache-discard`, `memory-pressure-off`, etc.
- Same for `AllocatorOptions(dir, false, false)` vs headed-safe flags.