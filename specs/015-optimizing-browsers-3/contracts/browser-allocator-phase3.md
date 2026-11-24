# Browser allocator contract: phase-3 headless RAM (015)

**Package**: `internal/browser`  
**Amends**: [014 browser-allocator-phase2.md](../../014-optimizing-browsers-2/contracts/browser-allocator-phase2.md)  
**Entry**: Appended when `lightweight && headless` **after** `headlessPhase2LightweightOptions()`.

## Phase-3 adopted flags (planned v1)

| Chrome flag | Value | Note |
|-------------|-------|------|
| `memory-pressure-off` | true | Only expected adopt from operator list (research R2) |

*If Chrome rejects the switch on reference binary, adopt set is **empty** and phase-3 slice is not appended (document in research).*

## Must remain on headless (any tier)

- `enable-webgl=true`
- `use-angle=swiftshader`
- Bug **-4** unset policy (no `disable-background-networking` as enabled disable)

## Must not appear (phase-3 or default profile)

From operator list and prior specs:

- `disable-gpu`, `disable-webgl`, `disable-3d-apis`
- `disable-background-networking` (enabled)
- `single-process`
- `js-flags` bundle with `expose-gc`, `optimize-for-size`, `flush-bytecode`, `gc-global` on default profile
- Duplicates of 013/014 flags (process-per-site, disable-application-cache, v8-cache-options=none)

## Verification

- Tests assert adopted phase-3 keys (if any) and forbidden keys absent
- `go test ./tests/integration/ -run 'OneWin|Aviator'` after implementation