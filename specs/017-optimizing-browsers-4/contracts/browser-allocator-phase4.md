# Browser allocator contract: phase 4 (017)

**Package**: `internal/browser`  
**Amends**: [015 browser-allocator-phase3.md](../../015-optimizing-browsers-3/contracts/browser-allocator-phase3.md)  
**Supersedes**: [016 browser-allocator-experimental.md](../../016-optimizing-browsers-experimental/contracts/browser-allocator-experimental.md) on default headless — **016 slice removed**.

## When applied

| `headless` | `lightweight` | Phase-4 slice |
|------------|---------------|---------------|
| true | true | **Yes** (only **adopt** flags from [operator-flag-matrix.md](./operator-flag-matrix.md)) |
| true | false | **No** |
| false | * | **No** |

## Prerequisite profile (015 baseline)

Before phase-4 append:

- Headless base includes `enable-webgl` + `use-angle=swiftshader` when `lightweight`.
- 013, 014 (incl. `js-flags` 320), 015 `memory-pressure-off` present.
- **No** 016 experimental GPU/WebGL/128-js flags.

## Phase-4 flags (when adopted)

Emitted from `headlessPhase4LightweightOptions()` per research R2 final matrix (initial adopt candidates: GPU compositing/rasterization, accelerated canvas/video decode off, `disable-component-update`, merged `disable-features` entries).

## Must not appear on default headless lightweight

- `disable-background-networking=true`
- 016 experiment set (`disable-gpu`, `disable-webgl`, `disable-3d-apis`, experimental 128 `js-flags`, etc.) unless matrix amended with stand-in proof
- **Reject** defaults: `no-sandbox`, `single-process`, `no-zygote` unless research records smoke + stand-in pass on headless shell

## Exec path

- `chromedp.ExecPath` MUST be project **chrome-headless-shell** when `headless` (see [headless-shell-binary.md](./headless-shell-binary.md)).

## Verification

- Unit tests: phase-4 flags present/absent per matrix; 016 flags absent; WebGL enable present on 015 path.
- Forbidden flags test: no `disable-background-networking=true`.