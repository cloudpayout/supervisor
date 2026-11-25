# Shipped lightweight argv amendment (2026-11-24)

**Package**: `internal/browser/chromedp.go`  
**Applies**: `supervisor start --headless` (default lightweight on)  
**Authority**: Operator request after PR #16; amends original **017** plan that reverted **016**.

## Allocator append order (headless + lightweight)

1. chromedp defaults + headless base (viewport, UA, automation unset blocks; **no** `enable-webgl` / `use-angle` on lightweight)
2. `disable-features` merge (013/014/017 tokens)
3. **013** `headlessLightweightOptions` — includes explicit **`disable-breakpad`** (with `disable-crash-reporter`)
4. **014** `headlessPhase2LightweightOptions` — **without** phase-2 `js-flags=320` (see step 6)
5. **015** `memory-pressure-off`
6. **017** `headlessPhase4LightweightOptions` — GPU compositing/raster, accelerated canvas/video, `disable-component-update`
7. **016** `headlessExperimentalLightweightOptions` — GPU/WebGL/3D stack off, process model, experimental `js-flags`

## 016 experimental slice (re-shipped on lightweight)

| Flag | Value |
|------|-------|
| `disable-gpu` | true |
| `disable-webgl` | true |
| `disable-3d-apis` | true |
| `disable-software-rasterizer` | true |
| `single-process` | true |
| `no-zygote` | true |
| `no-sandbox` | true |
| `js-flags` | `--max-old-space-size=128 --expose-gc --optimize-for-size --flush-bytecode --gc-global` |

## Operator list from `chrome-performance-flags.txt` (remaining gap)

| Flag | Shipped on lightweight? |
|------|-------------------------|
| All phase-4 adopts | Yes |
| `disable-breakpad` | Yes (013 bundle, explicit) |
| `disable-crash-reporter` | Yes (013) |
| `media-cache-size=1`, `disable-application-cache`, `renderer-process-limit=1`, `process-per-site` | Yes (013) |
| `single-process`, `no-zygote`, `no-sandbox` | Yes (016 slice, operator amendment) |
| `disable-background-networking=true` | **No** — headless forces **`disable-background-networking=false`** |

## `--no-lightweight` headless

Headless base **with** `enable-webgl` + `use-angle=swiftshader`; **none** of 013/014/015/017/016 slices.

## Risks (documented, not spec-gated on this host)

- **`single-process` / `no-sandbox` / `no-zygote`**: stability and security trade-offs; historical segfault on full Chrome 157 without headless shell.
- **`disable-webgl`**: canvas games may need stand-in re-verification after deploy.

## Verification

- `go test ./internal/browser/... -run 'Lightweight|Phase4|Experimental' -count=1`
- Unit tests require experiment flags + `single-process` + `no-sandbox` + `no-zygote` + `disable-breakpad` on `AllocatorOptions(..., true, true)`; forbid them on `(..., true, false)`.