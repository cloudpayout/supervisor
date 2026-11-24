# Research: Headless Shell and Performance Flags (017 / Phase 4)

**Date**: 2026-11-24  
**Spec**: [spec.md](./spec.md)  
**Before profile**: [015 research](../015-optimizing-browsers-3/research.md) — **median_015_mib** ~1126.2 MiB; shipped **015** path (no 016 experiment)  
**Operator input**: [chrome-performance-flags.txt](../../chrome-performance-flags.txt)

## R1: Baseline procedure (FR-008, SC-001)

**Decision**: Same as 013–016: idle headless `about:blank`, 30s warmup, 2s × 150 samples, `status` MEMORY median. Compare **017** default vs **015** median on same host. Optional control: `start --headless --no-lightweight` labeled `performance_flags=none`.

**Success bar**: **≥10%** lower than **015** median, or document reject/ineffective outcome per **SC-001**.

## R2: Operator list — adopt / reject / skip (pre-smoke lock; implement may adjust after stand-ins)

| Flag / proposal | Verdict | Notes |
|-----------------|---------|-------|
| `disable-crash-reporter`, `disable-breakpad` | **Skip (013)** | Already in lightweight tier |
| `media-cache-size=1` | **Skip (013/014)** | Duplicate in operator file |
| `disable-application-cache` | **Skip (013)** | Already present |
| `renderer-process-limit=1`, `process-per-site` | **Skip (013)** | Already present |
| `disable-background-networking` | **Reject** | FR-003; bug **-4** |
| `no-sandbox` | **Reject (default)** | FR-004; security; spec default |
| `single-process` | **Reject (default)** | 016 segfault on Chrome 157 full binary; retry only on headless-shell smoke |
| `no-zygote` | **Reject (default)** | Pairs with single-process risk; smoke before reconsider |
| `disable-gpu-compositing` | **Adopt candidate** | Does not remove `enable-webgl`; GPU stack partial disable |
| `disable-gpu-rasterization` | **Adopt candidate** | Same |
| `disable-accelerated-2d-canvas` | **Adopt candidate** | Canvas game risk — stand-in gate |
| `disable-accelerated-video-decode` | **Adopt candidate** | Low game risk for Aviator path |
| `disable-component-update` | **Adopt candidate** | Reduces background work |
| `disable-features` += `OnDeviceModel`, `PaintHolding`, `AudioServiceOutOfProcess` | **Adopt candidate** | Merge into `casinoDisableFeatures` / phase-4 slice without dropping IsolateOrigins merge |

**Rationale**: **015** kept WebGL + SwiftShader; phase-4 targets compositor/raster/video paths, not wholesale `disable-webgl`. Forbidden networking and process hacks stay off unless smoke + stand-ins pass.

**Alternatives**: Ship full `chrome-performance-flags.txt` verbatim (rejected: includes forbidden `disable-background-networking` and high-risk process flags).

## R3: Revert spec 016 experimental (FR-011)

**Decision**: Delete use of `headlessExperimentalLightweightOptions()` on default headless lightweight. Restore headless base **with** `enable-webgl` + `use-angle=swiftshader` when `lightweight`. Restore phase-2 **`js-flags=--max-old-space-size=320`** (single bundle, no 128 experiment).

**Rationale**: Operator declared **015** successful; **016** raised idle RAM and disabled WebGL on default path.

## R4: chrome-headless-shell provisioning (FR-005, FR-006)

**Decision**: Add `scripts/fetch-chrome-headless-shell.sh` downloading **Chrome for Testing** `chrome-headless-shell-linux64.zip` pinned to **157.0.8103.78** (same major as reference host full Chrome **157.0.8103.78** used in 015/016 baselines). Extract to repo-root `bin/chrome-headless-shell` (mode `755`). Add `bin/chrome-headless-shell` to `.gitignore`. Record exact CfT build string in Baseline Record **headless_shell_version**.

**Runtime**: `browser.HeadlessShellPath()` in **`internal/browser/executable.go`** returns absolute path; `AllocatorOptions` uses `chromedp.ExecPath` from that function when `headless`, never headed `Executable()` that reads `CHROME_PATH`.

**Headed**: Unchanged `Executable()` (`CHROME_PATH` then PATH).

**Missing bundle**: `start --headless` fails preflight with message naming `scripts/fetch-chrome-headless-shell.sh` — **no** fallback to full Chrome.

**Rationale**: Spec FR-006/007; smaller headless footprint; reproducible CI path once script runs in setup.

**Alternatives**: Commit binary to repo (rejected: constitution secrets/size); use system `google-chrome --headless` only (rejected: spec requires shell artifact).

## R5: CHROME_PATH on headless (FR-007)

**Decision**: **Ignore** `CHROME_PATH` when resolving headless binary. If `CHROME_PATH` is set at headless `start`, emit one-line stderr notice: headless uses bundled shell only (plan UX; not a failure).

**Rationale**: Spec forbids selecting full Chrome via env on headless; operators commonly export `CHROME_PATH` for headed debugging.

## R6: Ephemeral profiles (FR-012, FR-013)

**Decision**: Keep current layout `profiles/<instanceID>` per start. **Verify** (do not reimplement) `RemoveProfile` on: worker defer, `stop`, failed start rollback, `reconcile` dead instance. Add integration test: start headless → assert profile dir exists while `running` → `stop` → assert dir absent; repeat **N=10** cycles for **SC-007**.

**Rationale**: Spec operator constraint; code path largely exists per worker `run.go` and `cli/stop.go`.

**In-run growth (FR-014)**: Rely on existing 013/014 cache limits (`disk-cache-size`, `disable-application-cache`, `aggressive-cache-discard`, etc.) plus phase-4 adopt candidates that reduce GPU/compositor work. Document in operator-flag matrix which adopts address FR-014; no mid-run profile wipe (out of scope).

**FR-012 verification**: Integration test asserts two sequential headless starts receive **different** `profile_dir` values in registry while each is active; paths must not be reused after stop.

## R7: Stand-ins and rollback (FR-009, SC-003)

**Decision**: Run `go test ./tests/integration/ -run 'OneWin|Aviator'` after shell switch + phase-4 adopts. If fail: drop last adopt batch or reject failing flags in matrix; do not ship headless shell default if stand-ins fail entirely (blocker).

## R8: CI vs baseline

**Decision**: Unit + integration in default `go test`; RAM median and stand-ins in **Baseline Record (017)**. **Mandatory local/CI prerequisite**: run `scripts/fetch-chrome-headless-shell.sh` before any headless `start` or headless integration test. Integration helpers **skip** with a clear message when `bin/chrome-headless-shell` is absent so `go test ./...` on headed-only subsets still works; headless-specific tests fail closed when shell required. Document in [quickstart.md](./quickstart.md) and root `README.md` Development section.

## Baseline Record (017)

*(Fill during implementation.)*

| Field | Value |
|-------|-------|
| host | WSL2 (same reference host as 015, 2026-11-24) |
| headless_shell_version | Google Chrome for Testing 157.0.8103.78 |
| headless_shell_path | `bin/chrome-headless-shell` (symlink to `bin/chrome-headless-shell-linux64/chrome-headless-shell` with ICU data) |
| procedure | R1 |
| median_015_mib | 1126.2 (import from 015) |
| median_017_mib | 604.6 (150 samples × 2s after 30s, instance `rckjzblycj7a4`; min 604.3, max 604.7) |
| median_017_no_lightweight_control_mib | 600.2 (150 samples × 2s after 30s, instance `5pj4ywrjoq32o`; min 600.0, max 601.2; `performance_flags=none`) |
| ram_improvement_pct | about **46%** lower than 015 (1126.2 → 604.6); meets the 10% bar |
| standins_pass | yes (`go test ./tests/integration -run 'OneWin|Aviator'` passed 2026-11-24 on chrome-headless-shell 157) |
| phase4_adopted_flags | disable-gpu-compositing, disable-gpu-rasterization, disable-accelerated-2d-canvas, disable-accelerated-video-decode, disable-component-update, disable-features += OnDeviceModel,PaintHolding (AudioServiceOutOfProcess already in 014). FR-014: those plus existing 013/014 cache limits. |
| sc006_sc007_pass | yes (`TestProfileRemovedAfterStop`, `TestProfileUniquePerInstance`, `TestProfileTenCycles` passed 2026-11-24) |