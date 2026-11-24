# Browser allocator contract: lightweight profiles (013)

**Package**: `internal/browser`  
**Entry**: `AllocatorOptions(profileDir string, headless bool, lightweight bool)`  
**Source of `lightweight`**: registry row written at `start` (FR-018)

## Scope

When `lightweight == true` and `headless == true`, append **full headless lightweight profile** after existing headless + Spribe/casino flags (012 WebGL, `headlessUnset*`, site-isolation relaxations).

When `lightweight == true` and `headless == false`, append **headed-safe** options only (FR-002).

When `lightweight == false` (`start --no-lightweight`), append **no** 013 lightweight slices — headed or headless use only pre-013 Supervisor flags for that mode.

## Adopted flags — headless full tier (`lightweight && headless`)

Applied via `chromedp.Flag(...)` unless noted:

| Chrome flag | Value |
|-------------|-------|
| `force-device-scale-factor` | `1` |
| `renderer-process-limit` | `1` |
| `process-per-site` | true |
| `disk-cache-size` | `5242880` |
| `media-cache-size` | `1` |
| `disable-application-cache` | true |
| `js-flags` | `--max-old-space-size=384` |
| `v8-cache-options` | `none` |
| `disable-plugins` | true |
| `disable-notifications` | true |
| `disable-permissions-api` | true |
| `disable-speech-api` | true |
| `disable-webrtc` | true |
| `disable-logging` | true |
| `disable-crash-reporter` | true |
| `disable-in-process-stack-traces` | true |
| `disable-component-extensions-with-background-pages` | true |

## Adopted flags — headed-safe tier (`lightweight && !headless`)

| Chrome flag | Value |
|-------------|-------|
| `force-device-scale-factor` | `1` |

**Must NOT** apply headless-only rows from the full tier table to headed runs.

## Opt-out tier (`!lightweight`)

No rows from either table above. Spribe/casino/shared flags unchanged per prior specs.

## Explicitly NOT applied (any tier — guardrails)

| Flag | Reason |
|------|--------|
| `disable-gpu` | Breaks WebGL path with `--use-angle=swiftshader` |
| `disable-software-rasterizer` | SwiftShader requires software GL |
| `disable-remote-fonts` / `disable-reading-from-canvas` | Spribe/canvas probes (kept unset via `headlessUnset*`) |
| `disable-background-*` / `disable-renderer-backgrounding` | Spribe config fetches |
| Full operator `js-flags` bundle (`expose-gc`, etc.) | Chrome version fragility (see [research.md](../research.md) R2) |

## Unchanged obligations (prior specs)

- `headless=new`, `window-size=1280,720`, UA, automation blink, **enable-webgl**, **use-angle=swiftshader** (headless path)
- `disable-features=IsolateOrigins,site-per-process`, `disable-site-isolation-trials`
- Worker still runs `EmulateViewport(1280, 720)` first on headless (012)

## Verification

- `internal/browser/chromedp_test.go`: full bundle when `headless && lightweight`; headed-safe when `!headless && lightweight`; neither bundle when `!lightweight`.
- Integration: parallel instances with mixed `lightweight` on one host (no cross-leakage).
- Integration: memory baseline + existing headless viewport / start-stop tests (no live casino).