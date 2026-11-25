# Quickstart: validate 014 optimizing browsers-2

**Branch**: `014-optimizing-browsers-2`  
**Spec**: [spec.md](./spec.md)  
**Prerequisite**: Spec **013** implemented on this branch.

## Prerequisites

- Linux or WSL2, Chrome on `PATH` or `CHROME_PATH`
- `go build -o supervisor ./cmd/supervisor`

## 1. Compare 013 vs 014 idle MEMORY (SC-001 manual)

Measure **013-only** (disable phase-2 slice temporarily or use pre-014 commit) and **default** headless on the same host, same procedure as [013 quickstart §1](../013-optimizing-browsers/quickstart.md).

Record both medians in [research.md](./research.md) Baseline Record (014). Target: **014 median ≥15% lower** than 013 median.

## 2. Unit allocator regression

```bash
go test ./internal/browser/... -run 'Phase2|Lightweight' -count=1
```

Expect phase-2 flags present; bug **-4** unset flags still honored (no `enable-features`, background flags unset).

## 3. Stand-in reliability (SC-003)

```bash
go test ./tests/integration/ -count=1 -timeout 30m -run 'OneWin|Aviator'
```

Must pass before lowering `js-flags` to 320 if gated.

## 4. Default suite

```bash
go test ./... -count=1
```

## 5. Memory limit sanity (unchanged from 013)

```bash
./supervisor start --headless --memory-limit 600MiB
./supervisor status
```

Expect `… MiB/600.0 MiB`; NOTE may show `threshold exceeded` while instance stays `running` — **not** capped at 600 MiB. See [operator-memory-limits.md](./contracts/operator-memory-limits.md).

## 6. Opt-out

```bash
./supervisor start --headless --no-lightweight
```

Neither 013 nor 014 lightweight slices apply.

## SC-001 outcome

On the reference host (2026-11-23), the full idle window did **not** meet the 15% bar. A same-day 013-only control median was **1125.3 MiB**. The 014 profile with a 250 MiB V8 heap and shared workers disabled was **1122.1 MiB** (about **0.3%** lower). Heap 320 alone was **1121.2 MiB**. Details are in [research.md](./research.md) Baseline Record (014).

Phase-2 flags stay on the default headless lightweight profile because they are safe for the OneWin/Aviator stand-ins. They did not reduce idle `about:blank` RSS on this host. A live casino page may still differ. `--memory-limit` remains a `status` warning policy, not a cap on Chrome RSS.

## Contracts

- Phase-2 allocator: [contracts/browser-allocator-phase2.md](./contracts/browser-allocator-phase2.md)
- Limits vs flags: [contracts/operator-memory-limits.md](./contracts/operator-memory-limits.md)