# Browser allocator contract: phase-2 headless RAM (014)

**Package**: `internal/browser`  
**Amends**: [013 browser-allocator.md](../../013-optimizing-browsers/contracts/browser-allocator.md)  
**Entry**: Appended when `lightweight && headless` **after** `headlessLightweightOptions()`.

## Phase-2 adopted flags (v1)

| Chrome flag | Value | Note |
|-------------|-------|------|
| `aggressive-cache-discard` | true | New in 014 |
| `disable-back-forward-cache` | true | New in 014 |
| `disable-domain-reliability` | true | New in 014 |
| `disk-cache-size` | `1048576` | **Replaces** 013 `5242880` (implement one value total) |
| `disable-features` | merge | Append `BackForwardCache,MediaRouter,AudioServiceOutOfProcess` to existing Supervisor/casino merge; keep `IsolateOrigins,site-per-process`; do not pass `enable-features` |
| `js-flags` | `--max-old-space-size=320` | Kept. Heap 250 did not lower idle RSS and was reverted (T030) |

## Must not appear (phase-2 or any tier)

Same as 013 **plus** bug **-4** blocklist:

- `disable-gpu`, `disable-software-rasterizer`
- `disable-background-networking`, `disable-background-timer-throttling`, `disable-backgrounding-occluded-windows`, `disable-renderer-backgrounding` as **enabled** (must stay unset on headless)
- `disable-remote-fonts`, `disable-reading-from-canvas` as **enabled**
- `safebrowsing-disable-auto-update` as **enabled**
- `enable-features=NetworkService,NetworkServiceInProcess`

## Verification

- `TestHeadlessPhase2AllocatorFlags`: adopted phase-2 keys present; 013 core keys still present; forbidden keys absent; `headlessUnset*` behavior unchanged.
- Stand-in: `go test ./tests/integration/ -run 'OneWin|Aviator'`.