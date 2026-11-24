# Browser allocator contract: phase 4 (017)

**Package**: `internal/browser`  
**Amends**: [015 browser-allocator-phase3.md](../../015-optimizing-browsers-3/contracts/browser-allocator-phase3.md)  
**Coexists with**: [016 browser-allocator-experimental.md](../../016-optimizing-browsers-experimental/contracts/browser-allocator-experimental.md) — 016 slice **re-shipped after** phase-4 per [shipped-lightweight-argv-amendment.md](./shipped-lightweight-argv-amendment.md).

## When applied

| `headless` | `lightweight` | Phase-4 slice |
|------------|---------------|---------------|
| true | true | **Yes** (only **adopt** flags from [operator-flag-matrix.md](./operator-flag-matrix.md)) |
| true | false | **No** |
| false | * | **No** |

## Prerequisite profile (013–015)

Before phase-4 append:

- 013 (incl. explicit `disable-breakpad`), 014 cache slice, 015 `memory-pressure-off` present.
- On **amended** shipped profile: headless lightweight **omits** WebGL enable; 014 omits **320** `js-flags` (016 128 bundle applies later).

## Phase-4 flags (when adopted)

Emitted from `headlessPhase4LightweightOptions()` per research R2 final matrix (initial adopt candidates: GPU compositing/rasterization, accelerated canvas/video decode off, `disable-component-update`, merged `disable-features` entries).

## Must not appear on default headless lightweight

- `disable-background-networking=true`

## Shipped after phase-4 (operator amendment)

See [shipped-lightweight-argv-amendment.md](./shipped-lightweight-argv-amendment.md): 016 experimental + `single-process` + `no-zygote` + `no-sandbox`.

## Exec path

- `chromedp.ExecPath` MUST be project **chrome-headless-shell** when `headless` (see [headless-shell-binary.md](./headless-shell-binary.md)).

## Verification

- Unit tests: phase-4 flags present; 016 + process flags present on default lightweight per amendment; WebGL enable only on `--no-lightweight`.
- Forbidden flags test: no `disable-background-networking=true`.