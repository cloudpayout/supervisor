# Research: Lightweight Browser Instances (013)

**Date**: 2026-11-23  
**Spec**: [spec.md](./spec.md)  
**Input**: Operator-recommended Chrome flags (plan message) + existing `internal/browser/chromedp.go` / `headless-chrome-flags.txt`

## R1: Baseline measurement procedure (FR-007, SC-001)

**Decision**: Capture **median aggregated RSS (MiB)** from `metrics.Aggregate` for one headless instance on `about:blank`, `running` ≥ 60s, sample every 2s for 5 minutes (150 samples), discard first 30s after `running`. Record **pre-change** on current `main`/parent commit **before** allocator edits; **post-change** on same host, same `CHROME_PATH`, single instance.

**Rationale**: Matches `status` MEMORY basis (owner + Chrome tree RSS / 1024²). Idle blank page isolates launcher overhead from casino variance.

**Alternatives**: `ps` manual sampling (rejected: diverges from product metric); multi-instance baseline (rejected: confounds fleet effects).

## R2: Lightweight Chrome flags — adopt / reject matrix

Operator suggested flags evaluated against **Spribe iframe / WebGL / canvas / fonts** guardrails (spec 012, 007, `headlessUnset*` arrays).

| Flag | Verdict | Notes |
|------|---------|-------|
| `--force-device-scale-factor=1` | **Adopt** (headless default profile) | Stabilizes layout vs DPR; aligns with fixed 1280×720. |
| `--disable-gpu` | **Reject** | Conflicts with `--enable-webgl` + `--use-angle=swiftshader` (FR-003). Breaks or degrades game canvas paths. |
| `--disable-software-rasterizer` | **Reject** | SwiftShader is the software rasterizer path for headless WebGL. |
| `--renderer-process-limit=1` | **Adopt** (headless) | Caps renderer processes; acceptable for single-tab automation; monitor iframe regressions in tests. |
| `--process-per-site` | **Adopt** (headless) | Reduces process fan-out; casino still uses cross-site iframes with site isolation already relaxed. |
| `--disk-cache-size=5242880` | **Adopt** (headless) | 5 MiB cap; low risk for session-driven runs. |
| `--media-cache-size=1` | **Adopt** (headless) | Minimal media cache. |
| `--disable-application-cache` | **Adopt** (headless) | Legacy appcache; harmless for current sites. |
| `--js-flags=--max-old-space-size=250 …` | **Defer / soften** | `250` MiB V8 heap is aggressive for embedded games; **start with `--js-flags=--max-old-space-size=384`** only (drop `--expose-gc --optimize-for-size --flush-bytecode --gc-global` in v1 — brittle across Chrome versions). Revisit if stand-in entry passes and baseline still misses SC-001. |
| `--v8-cache-options=none` | **Adopt** (headless) | Trades cold-start CPU for lower idle memory; acceptable for long-lived instances. |
| `--disable-plugins` | **Adopt** (headless) | NPAPI removed; no-op safety. |
| `--disable-notifications` | **Adopt** (headless) | No casino dependency. |
| `--disable-permissions-api` | **Adopt** (headless) | Automation does not drive permission prompts. |
| `--disable-speech-api` | **Adopt** (headless) | — |
| `--disable-webrtc` | **Adopt** (headless) | Aviator does not require WebRTC. |
| `--disable-logging` | **Adopt** (headless) | Minor; worker logs remain on disk. |
| `--disable-crash-reporter` | **Adopt** (headless) | — |
| `--disable-in-process-stack-traces` | **Adopt** (headless) | — |
| `--disable-component-extensions-with-background-pages` | **Adopt** (headless) | Fewer background extension processes. |
| `--disable-translate` | **Skip** | `chromedp.DefaultExecAllocatorOptions` already sets `disable-features=Translate,…`. |

**Rationale**: Maximize **memory** wins on headless fleet without reversing explicit Spribe/casino overrides (`headlessUnset*`, WebGL, site-isolation flags).

**Alternatives**: Apply full operator bundle verbatim (rejected: GPU/software-rasterizer flags contradict existing game profile); no new flags (rejected: fails SC-001 intent).

## R3: Lightweight tiers — headed, headless, opt-out (FR-001, FR-002, FR-018)

**Decision**: Three allocator behaviors driven by registry **`lightweight`** (from `start`; default lightweight **on** unless `--no-lightweight`):

| Tier | When | Options appended |
|------|------|------------------|
| **Full headless** | `headless && lightweight` | Full table in [browser-allocator.md](./contracts/browser-allocator.md) *Adopted flags (headless full)* |
| **Headed-safe** | `!headless && lightweight` | **`force-device-scale-factor=1` only** — no `renderer-process-limit`, no V8/cache/cache-size bundle |
| **None** | `!lightweight` | **No** 013 lightweight slices; keep existing Spribe/WebGL/012 headless headed paths only |

**Rationale**: Clarification session 2026-11-23 (option C): same `--no-lightweight` flag for headed and headless; SC-001 and FR-008 regression use default headless **with** lightweight, not `--no-lightweight`.

**Alternatives**: Headless-only opt-out flag (rejected: user chose C); always apply full bundle on headed (rejected: FR-002 / US4).

## R3b: Registry `lightweight` field (FR-018)

**Decision**: Persist **`lightweight: bool`** on each instance at `start`. Worker passes it into `AllocatorOptions`. `false` when `--no-lightweight`; `true` when headed-safe or full headless tier applies. Independent of `--memory-limit`.

**Rationale**: CLI process exits after `start`; worker must not guess profile from flags not on disk.

## R4: `--memory-limit` parsing (FR-011, MiB)

**Decision**: New `internal/memlimit` package: parse suffixes `M`/`MiB` (×1024² bytes), `G`/`GiB` (×1024³); reject `MB`/`GB`, bare numbers, ≤0. Store **`MemoryLimitBytes`** on registry row (integer bytes) for exact comparisons; display in MiB with one decimal.

**Rationale**: Binary mebibytes end-to-end; avoids float drift in warning line (`threshold% × limit`).

**Alternatives**: Store MiB float (rejected: rounding in warnings); decimal MB (rejected: spec).

## R5: Registry persistence (FR-014, FR-018)

**Decision**: Add optional JSON fields on `registry.Instance`:

- `memory_limit_bytes` (uint64, omitempty)
- `memory_threshold_percent` (int, omitempty, 1–100; default 100 when limit set and flag omitted)
- `lightweight` (bool, omitempty in JSON; **write explicit** `true`/`false` on 013 `start` rows)

Bump **`schema_version` to 2** when any 013 field is first written on a document (memory policy and/or new instances with explicit `lightweight`); readers accept version 1 rows (implicit: no memory fields; treat missing `lightweight` as legacy pre-013 behavior until stop).

**Rationale**: `status` must evaluate warnings across refreshes; worker must honor opt-out without re-parsing CLI.

## R6: Status NOTE message (FR-015)

**Decision**: Constant string **`threshold exceeded`** in NOTE column (ASCII, no punctuation required in v1). Empty string when no message.

**Rationale**: Fixed text per spec; plan leaves copy tweak to help text only.

## R7: MEMORY column formatting (FR-005)

**Decision**:

| Condition | MEMORY |
|-----------|--------|
| No limit | `{current} MiB` (unchanged, e.g. `245.3 MiB`) |
| Limit, `starting` | `unavailable/{limit} MiB` |
| Limit, `running` | `{current} MiB/{limit} MiB` |

Limit side formatted from stored bytes via same `formatMiBOneDecimal`.

**Rationale**: Clarifications require MiB on both sides; space before `MiB` matches existing current column.

## R8: Testing memory warning without live casino

**Decision**: Integration test `TestNoteThresholdExceeded` starts with `--memory-limit 1M` and `--threshold 50` (warning line 0.5 MiB) on blank headless; poll `status` output for up to **120s** until NOTE contains `threshold exceeded`. Test **MUST NOT** skip in default CI; fail if timeout.

**Rationale**: FR-008 without mocking casino; blank headless RSS typically exceeds 0.5 MiB quickly.

**Alternatives**: RSS injection hook (rejected unless timeout failures exceed 10% on CI hosts).

## R10: Baseline Record vs CI (SC-001, FR-008)

**Decision**: **Baseline Record** subsection in this file holds pre/post median MiB (T009/T014). **SC-001** percent improvement is validated there or in PR notes, not in `go test ./...`. Default CI uses **`TestHeadlessLightweightAllocatorFlags`** plus a contrasting test that **`--no-lightweight` headless** omits the headless bundle (FR-008, FR-018) in `internal/browser/chromedp_test.go` (task T006+).

**Rationale**: Host-specific RSS; CI still proves flags landed and opt-out works.

## R9: `disable-features` merging

**Decision**: Supervisor already appends `disable-features=IsolateOrigins,site-per-process`. Lightweight profile **must not** add a second conflicting `disable-features` flag; extend existing merge in code (single flag string) if Translate must remain from defaults.

**Rationale**: Chrome last-wins behavior is error-prone.

## Quickstart validation (T029, T045)

Recorded 2026-11-23 on this host:

- §1 post-change idle median: **1120.0 MiB** (see Baseline Record). Pre-change median was not captured.
- §2–§3, §6–§7: covered by `go test ./...` and `go test ./tests/integration/ -run 'OneWin|Aviator'` (both passed earlier this session).
- §4–§5: `--threshold` without `--memory-limit` exited 2; `--memory-limit 512MB` exited 2. Limit formatting and NOTE are covered by integration tests.
- §8: after `stop`, the sampled instance id was absent from the next `status`.

## SC-004 stand-in run (T042)

`go test ./tests/integration/ -count=1 -timeout 30m -run 'OneWin|Aviator'` passed on 2026-11-23 (stand-in pages, no live casino). No allocator guardrail change required.

## Baseline Record

*(Populate at implementation: T009 pre-change, T014 post-change.)*

| Field | Value |
|-------|-------|
| host | Linux Cedano 6.18.58.1-microsoft-standard-WSL2 x86_64 |
| chrome_binary | /usr/bin/google-chrome |
| procedure | idle headless `about:blank`, sample every 2s × 5m after 30s warmup (quickstart §1) |
| pre_median_mib | unavailable — lightweight allocator was already on this tree, so a pre-change sample cannot be taken here (T009) |
| post_median_mib | 1120.0 (149 samples, 2026-11-23, default `start --headless`) |
| sc001_pass | not evaluated — no pre-change median to compare |