# Quickstart: validate 015 optimizing browsers-3

**Branch**: `015-optimizing-browsers-3`  
**Spec**: [spec.md](./spec.md)  
**Prerequisite**: Spec **014** implemented.

## Prerequisites

- Linux or WSL2, Chrome on `PATH` or `CHROME_PATH`
- `go build -o supervisor ./cmd/supervisor`

## 1. Compare 014 vs 015 idle MEMORY (SC-001 manual)

Same procedure as [014 quickstart §1](../014-optimizing-browsers-2/quickstart.md) (150 samples, 30s warmup). Record medians in [research.md](./research.md) Baseline Record (015). Reference **014** median from [014 research](../014-optimizing-browsers-2/research.md).

## 2. Review operator matrix

Read [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md) — most aggressive proposals should **not** appear on the command line.

## 3. Unit allocator regression

```bash
go test ./internal/browser/... -run 'Phase3|Phase2|Lightweight' -count=1
```

Expect phase-3 adopt flag(s) if any; `enable-webgl` and `use-angle=swiftshader` still present; forbidden flags absent.

## 4. Stand-in reliability (SC-003)

```bash
go test ./tests/integration/ -count=1 -timeout 30m -run 'OneWin|Aviator'
```

## 5. Default suite

```bash
go test ./... -count=1
```

## 6. Opt-out

```bash
./supervisor start --headless --no-lightweight
```

No 013/014/015 lightweight slices.

## SC-001 outcome

On the reference host (2026-11-23, Google Chrome 157.0.8103.78), the phase-3 profile did **not** meet the 15% bar. The documented **014** median is **1122.1 MiB**. The **015** window (150 samples after 30s) was **1126.2 MiB** (min 965.5, max 1127.0), about **0.4% higher**. Details are in [research.md](./research.md) Baseline Record (015).

The only adopted switch is `--memory-pressure-off`. GPU, WebGL, 3D, background-networking, single-process, and the aggressive `js-flags` bundle stay rejected. `--memory-limit` remains a `status` warning policy, not a cap on Chrome RSS.

## Contracts

- Phase-3 allocator: [contracts/browser-allocator-phase3.md](./contracts/browser-allocator-phase3.md)
- Operator matrix: [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md)