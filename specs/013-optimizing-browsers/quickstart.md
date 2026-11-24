# Quickstart: validate 013 optimizing browsers

**Branch**: `013-optimizing-browsers`  
**Spec**: [spec.md](./spec.md)

## Prerequisites

- Linux or WSL2, Chrome/Chromium on `PATH` or `CHROME_PATH`
- Built `supervisor` from this branch (`go build -o supervisor ./cmd/supervisor`)

## 1. Baseline memory (pre or post comparison)

```bash
./supervisor start --headless
# wait until status shows running
./supervisor status
# record MEMORY (MiB) every 2s for 5m; compute median
./supervisor stop <id>
```

Use **default** headless (lightweight **on**). Do **not** pass `--no-lightweight` for SC-001. Document median in PR notes; post-change median should be ≥20% lower vs pre-change on same host ([SC-001](./spec.md#success-criteria-mandatory)).

## 2. Lightweight headless starts (regression)

```bash
go test ./tests/integration/... -run 'Headless|Start|Status|Lightweight' -count=1
go test ./internal/browser/... -run 'Lightweight' -count=1
```

Expect pass with no live casino. Contrast:

```bash
./supervisor start --headless --no-lightweight
# worker must omit headless lightweight bundle; lifecycle still succeeds
./supervisor stop <id>
```

## 3. Headed lightweight vs opt-out

```bash
./supervisor start
# headed-safe tier only (visible window)

./supervisor start --no-lightweight
# no 013 lightweight extras; visible window
```

Compare allocator tests or flag introspection per [browser-allocator.md](./contracts/browser-allocator.md).

## 4. Memory limit + NOTE column

```bash
./supervisor start --headless --memory-limit 512MiB --threshold 80
./supervisor status
```

Expect header includes **NOTE**; MEMORY like `NNN.N MiB/512.0 MiB` (space before each `MiB`); NOTE empty while below 80% of limit.

Combine with opt-out (independent flags):

```bash
./supervisor start --headless --no-lightweight --memory-limit 512MiB
```

Low-limit smoke (integration test pattern):

```bash
./supervisor start --headless --memory-limit 1M --threshold 50
./supervisor status
# NOTE may show "threshold exceeded" once RSS > 0.5 MiB
```

## 5. Invalid flags

```bash
./supervisor start --headless --threshold 90
# expect usage error, no ID

./supervisor start --headless --memory-limit 512MB
# expect rejection of decimal MB
```

## 6. Mixed fleet

Start one default headless, one `--no-lightweight` headless, and one headed; confirm each row keeps its profile until `stop` (no cross-instance leakage).

## 7. Stand-in game entry (if available)

Re-run existing OneWin Aviator stand-in / integration checks from spec 007 after allocator change ([SC-004](./spec.md)).

## 8. Status refresh and SC-006 (manual)

Verifies [SC-006](./spec.md#success-criteria-mandatory): stopped rows vanish after one `status`; running row MEMORY aligns with baseline story.

```bash
./supervisor start --headless
ID=<id from output>
./supervisor status
# note MEMORY for running row (MiB)
./supervisor stop "$ID"
./supervisor status
# expect stopped instance absent from table within one refresh
```

Compare running MEMORY from the first `status` to **Baseline Record** `pre_median_mib` in [research.md](./research.md): post-change idle SHOULD be lower when measured per §1 (T009/T014). Record pass/fail in PR notes (task T029).

## Contracts

- CLI: [contracts/cli.md](./contracts/cli.md)
- Allocator: [contracts/browser-allocator.md](./contracts/browser-allocator.md)
- Registry: [contracts/registry-schema.json](./contracts/registry-schema.json)