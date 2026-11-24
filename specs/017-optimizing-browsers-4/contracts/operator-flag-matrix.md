# Operator flag matrix: phase 4 (017)

**Source**: [chrome-performance-flags.txt](../../../chrome-performance-flags.txt), [research.md](../research.md) R2  
**Baseline**: **015** default headless lightweight ([015 research](../../015-optimizing-browsers-3/research.md))

| Switch | Verdict (plan) | Rationale |
|--------|----------------|-----------|
| `disable-gpu-compositing` | **Adopt** | Shipped in `headlessPhase4LightweightOptions` |
| `disable-gpu-rasterization` | **Adopt** | Shipped |
| `disable-accelerated-2d-canvas` | **Adopt** | Shipped; stand-in gate |
| `disable-accelerated-video-decode` | **Adopt** | Shipped |
| `disable-component-update` | **Adopt** | Shipped |
| `OnDeviceModel`, `PaintHolding`, `AudioServiceOutOfProcess` (features) | **Adopt** | OnDeviceModel + PaintHolding added; AudioServiceOutOfProcess already in 014 merge |
| `disable-crash-reporter` | Already present (013) | Skip duplicate |
| `disable-breakpad` | **Adopt** | Explicit in 013 `headlessLightweightOptions` (operator amendment 2026-11-24) |
| `media-cache-size=1` | Already present | Skip duplicate |
| `disable-application-cache` | Already present (013) | Skip |
| `renderer-process-limit=1`, `process-per-site` | Already present (013) | Skip |
| `disable-background-networking` | **Reject** | FR-003 / bug **-4** |
| `no-sandbox` | **Adopt (amendment)** | Shipped in `headlessExperimentalLightweightOptions`; operator/dev host only — see [shipped-lightweight-argv-amendment.md](./shipped-lightweight-argv-amendment.md) |
| `single-process` | **Adopt (amendment)** | Same; pairs with `no-zygote` |
| `no-zygote` | **Adopt (amendment)** | Same |

**016 experimental** (`disable-gpu`, `disable-webgl`, `disable-3d-apis`, `disable-software-rasterizer`, 128 `js-flags`): **re-shipped** on default headless lightweight per operator amendment (supersedes original **017** FR-011 revert for this branch).

Implementers update this table with **final** adopt/reject after Chrome headless-shell smoke and stand-ins. Phase-4 **Adopt** rows ship in `headlessPhase4LightweightOptions()`; process-model and 016 rows ship in `headlessExperimentalLightweightOptions()`.

**FR-014 (in-run profile growth)**: Existing **013/014** cache switches already limit disk growth; phase-4 GPU/compositor/component-update adopts are incremental. Note in Baseline Record which adopts were counted toward FR-014.