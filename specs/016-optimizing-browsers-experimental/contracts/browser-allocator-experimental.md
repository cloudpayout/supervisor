# Browser allocator contract: experimental RAM (016)

**Package**: `internal/browser`  
**Amends**: [015 browser-allocator-phase3.md](../../015-optimizing-browsers-3/contracts/browser-allocator-phase3.md)  
**Entry**: `headlessExperimentalLightweightOptions()` when `headless && lightweight`, **after** `headlessPhase3LightweightOptions()`.

## When applied

| `headless` | `lightweight` | Experimental slice |
|------------|---------------|---------------------|
| true | true | **Yes** (default `start --headless`) |
| true | false | **No** |
| false | * | **No** |

## Experiment flags (required on argv)

| Chrome flag | Value |
|-------------|-------|
| `disable-gpu` | true |
| `disable-webgl` | true |
| `disable-3d-apis` | true |
| `disable-software-rasterizer` | true |
| `single-process` | **Omitted** when research records the reference Chrome cannot start (Chrome 157 segfault). See research R3 and spec FR-003. |
| `js-flags` | `--max-old-space-size=128 --expose-gc --optimize-for-size --flush-bytecode --gc-global` |

## Must not appear on experimental path

- `enable-webgl` / `use-angle=swiftshader` (conflict with experiment GPU stack)
- Second `js-flags` (phase-2 **320** must be suppressed when experimental is active)
- `disable-background-networking=true` (bug **-4** policy)

## Coexistence with 013/014/015

013/014/015 slices remain appended **before** experimental except:

- Phase-2 `js-flags` is **not** emitted when experimental slice is used (single winning `js-flags`).

## Verification

- `TestHeadlessExperimentalAllocatorFlags`: the five shipped switches (`disable-gpu`, `disable-webgl`, `disable-3d-apis`, `disable-software-rasterizer`, one experimental `js-flags`) are present on `AllocatorOptions(dir, true, true)`; `single-process` is absent on Chrome 157.
- `TestHeadlessExperimentalForbiddenFlags`: no `enable-webgl`, no duplicate `js-flags`, no `disable-background-networking=true`.
- `TestNoLightweightOmitsExperimentalAndLightweightTiers`: headless and headed `--no-lightweight` omit the experiment set and 013/014/015 RAM tiers.
- `go test ./tests/integration/ -run 'OneWin|Aviator'` after implementation.