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
| `disable-crash-reporter`, `disable-breakpad` | Already present (013) | Skip |
| `media-cache-size=1` | Already present | Skip duplicate |
| `disable-application-cache` | Already present (013) | Skip |
| `renderer-process-limit=1`, `process-per-site` | Already present (013) | Skip |
| `disable-background-networking` | **Reject** | FR-003 / bug **-4** |
| `no-sandbox` | **Reject (default)** | FR-004 security |
| `single-process` | **Reject (default)** | 016 startup failure pattern |
| `no-zygote` | **Reject (default)** | Process model risk |

Implementers update this table with **final** adopt/reject after Chrome headless-shell smoke and stand-ins. Only **Adopt** rows ship in `headlessPhase4LightweightOptions()`.

**FR-014 (in-run profile growth)**: Existing **013/014** cache switches already limit disk growth; phase-4 GPU/compositor/component-update adopts are incremental. Note in Baseline Record which adopts were counted toward FR-014.