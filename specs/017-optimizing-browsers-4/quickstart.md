# Quickstart: validate 017 optimizing browsers-4

**Branch**: `017-optimizing-browsers-4`  
**Spec**: [spec.md](./spec.md)  
**Prerequisite**: Spec **015** on tree; **016** experimental slice removed per plan.

## Prerequisites

- Linux or WSL2  
- `go build -o supervisor ./cmd/supervisor`  
- **Required before headless `start` or headless integration tests** (see research R8): provision **chrome-headless-shell** pinned to **157.0.8103.78**:

```bash
./scripts/fetch-chrome-headless-shell.sh
test -x ./bin/chrome-headless-shell && ./bin/chrome-headless-shell --version
```

Without the bundled shell, headless `start` fails preflight. Headed tests and non-headless packages in `go test ./...` may still run; headless integration tests skip or fail fast per `tests/integration/headless_shell_require.go`.

## 1. Headless uses bundled shell (SC-002, FR-007)

```bash
export CHROME_PATH=/usr/bin/google-chrome   # must NOT affect headless
./supervisor start --headless
# Confirm process argv uses .../bin/chrome-headless-shell (plan: ps or integration test)
./supervisor stop <id>
```

Headed control (still respects `CHROME_PATH`):

```bash
export CHROME_PATH=/usr/bin/google-chrome
./supervisor start
./supervisor stop <id>
```

## 2. Compare 015 vs 017 idle MEMORY (SC-001)

Same procedure as [015 quickstart §1](../015-optimizing-browsers-3/quickstart.md) (150 samples, 30s warmup).

| Arm | Command |
|-----|---------|
| Reference 015 | Documented **median_015_mib** in [015 research](../015-optimizing-browsers-3/research.md) |
| 017 default | `./supervisor start --headless` |
| Control (optional) | `./supervisor start --headless --no-lightweight` |

Record in [research.md](./research.md) Baseline Record (017).

## 3. Allocator regression

```bash
go test ./internal/browser/... -run 'Phase4|Lightweight|HeadlessShell' -count=1
```

Expect **013–015** + **017 phase-4** + **016 experimental** (incl. `single-process`, `no-sandbox`, `no-zygote`, `disable-breakpad`); **no** `disable-background-networking=true`. See [contracts/shipped-lightweight-argv-amendment.md](./contracts/shipped-lightweight-argv-amendment.md).

## 4. Ephemeral profile (SC-006, SC-007)

```bash
go test ./tests/integration/ -count=1 -run 'ProfileEphemeral|ProfileRemoved' -timeout 10m
```

Manual spot-check:

```bash
# After stop, profile path for instance ID must be gone under XDG_DATA_HOME/.../profiles/
```

## 5. Stand-in reliability (SC-003)

```bash
go test ./tests/integration/ -count=1 -timeout 30m -run 'OneWin|Aviator'
```

## 6. Default suite

```bash
go test ./... -count=1
```

## 7. Help text (SC-005)

```bash
./supervisor start --help
```

Expect: headless uses bundled **chrome-headless-shell**; `CHROME_PATH` does not select full Chrome for headless; ephemeral profiles.

## SC-001 outcome

On the reference host (2026-11-24, chrome-headless-shell 157.0.8103.78), default headless idle MEMORY median was **604.6 MiB** (min 604.3, max 604.7) versus **015** **1126.2 MiB**, about **46% lower**, which meets the 10% bar. Details are in [research.md](./research.md) Baseline Record (017).

## Contracts

- Phase-4 allocator: [contracts/browser-allocator-phase4.md](./contracts/browser-allocator-phase4.md)  
- Headless shell: [contracts/headless-shell-binary.md](./contracts/headless-shell-binary.md)  
- Ephemeral profile: [contracts/ephemeral-profile.md](./contracts/ephemeral-profile.md)  
- Flag matrix: [contracts/operator-flag-matrix.md](./contracts/operator-flag-matrix.md)  