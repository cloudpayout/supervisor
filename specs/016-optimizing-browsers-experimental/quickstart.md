# Quickstart: validate 016 optimizing browsers-experimental

**Branch**: `016-optimizing-browsers-experimental`  
**Spec**: [spec.md](./spec.md)  
**Prerequisite**: Spec **015** implemented on tree.

## Prerequisites

- Linux or WSL2, Chrome on `PATH` or `CHROME_PATH`
- `go build -o supervisor ./cmd/supervisor`

## 1. Compare 015 vs 016 default idle MEMORY (SC-003)

Same procedure as [015 quickstart §1](../015-optimizing-browsers-3/quickstart.md) (150 samples, 30s warmup).

| Arm | Command |
|-----|---------|
| Reference 015 | Pre-016 binary or documented **median_015_mib** in [015 research](../015-optimizing-browsers-3/research.md) |
| 016 default | `./supervisor start --headless` |
| Control (optional) | `./supervisor start --headless --no-lightweight` |

Record in [research.md](./research.md) Baseline Record (016).

## 2. Allocator regression

```bash
go test ./internal/browser/... -run 'Experimental|NoLightweight|Phase3|Lightweight' -count=1
```

Expect the five shipped experiment switches on default headless lightweight (`disable-gpu`, `disable-webgl`, `disable-3d-apis`, `disable-software-rasterizer`, experimental `js-flags`); **no** `single-process` on Chrome 157; **no** experiment flags on `--no-lightweight`.

## 3. Stand-in reliability (SC-004, FR-007)

```bash
go test ./tests/integration/ -count=1 -timeout 30m -run 'OneWin|Aviator'
```

Run against **default** `start --headless` (integration helpers use default lightweight). Record pass/fail in Baseline Record (016).

## 4. Default non-game suite

```bash
go test ./... -count=1
```

## 5. Help text (SC-006)

```bash
./supervisor start --help
```

Expect: default headless uses aggressive RAM switches; `--no-lightweight` skips all RAM tuning.

## 6. Opt-out control

```bash
./supervisor start --headless --no-lightweight
```

Must omit 013/014/015/016 performance flags ([no-lightweight-neutral.md](./contracts/no-lightweight-neutral.md)).

## SC-003 outcome

On the reference host (2026-11-24, Google Chrome 157.0.8103.78), default headless did **not** meet the 10% bar versus **015**. The documented **015** median is **1126.2 MiB**. The **016** default window (150 samples after 30s) was **1158.4 MiB** (min 989.5, max 1162.0), about **2.9% higher**. The control arm `start --headless --no-lightweight` (`performance_flags=none`) measured **1123.9 MiB** (min 963.7, max 1146.4) on the same procedure.

Shipped default argv uses the five switches that start on this Chrome: GPU, WebGL, 3D APIs, and software rasterizer disabled, plus one `js-flags` bundle at old-space 128. `--single-process` stays off (startup segfault). Stand-in tests including OneWin and Aviator passed. Because idle RAM did not improve, a follow-up to restore the **015-only** default is planned in research §R5 and is not executed in this feature. Details are in [research.md](./research.md) Baseline Record (016).

## Contracts

- Experimental allocator: [contracts/browser-allocator-experimental.md](./contracts/browser-allocator-experimental.md)
- No-lightweight: [contracts/no-lightweight-neutral.md](./contracts/no-lightweight-neutral.md)