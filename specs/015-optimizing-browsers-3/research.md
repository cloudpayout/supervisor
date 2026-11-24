# Research: Aggressive Chrome RAM Reduction (015 / Phase 3)

**Date**: 2026-11-23  
**Spec**: [spec.md](./spec.md)  
**Depends on**: [014 research](../014-optimizing-browsers-2/research.md) Baseline Record (014)  
**Operator input**: Flag list from spec Input line

## R1: Baseline procedure (FR-006, SC-001)

**Decision**: Same as 013/014: idle headless `about:blank`, 30s warmup discard, 2s × 150 samples, `status` MEMORY MiB median. Compare **phase-3 (015)** default profile vs **014** median on same host (**use 1122.1 MiB shipped 014 reference** or re-measure 014 same day before 015 lands).

**Rationale**: SC-001 is vs 014, not 013.

## R2: Operator list — adopt / reject / skip

| Flag / proposal | Verdict | Notes |
|-----------------|---------|-------|
| `process-per-site` | **Skip (013)** | Already in `headlessLightweightOptions()` |
| `disable-application-cache` | **Skip (013)** | Already present |
| `v8-cache-options=none` | **Skip (013)** | Already present |
| `disable-gpu` | **Reject** | Conflicts with headless `enable-webgl` + ANGLE SwiftShader path (013/014 guardrail) |
| `disable-webgl` | **Reject** | Contradicts `chromedp.Flag("enable-webgl", true)`; breaks canvas games (FR-004) |
| `disable-3d-apis` | **Reject** | Same as WebGL policy for supported games |
| `disable-background-networking` | **Reject** | Bug **-4** / Spribe; must keep unset on headless |
| `single-process` | **Reject** | Stability, iframe, WSL2 (013/014 research) |
| `js-flags=--max-old-space-size=128 --expose-gc --optimize-for-size --flush-bytecode --gc-global` | **Reject (default profile)** | FR-005; brittle; 014 showed heap steps did not move idle RSS; stand-in risk |
| `memory-pressure-off` | **Adopt (phase-3 slice)** | Only net-new operator flag not in 013/014; low game risk; verify Chrome accepts switch |

**Rationale**: Spec FR-003–FR-005; 014 idle baseline showed safe flags barely move RSS — phase 3 must not trade game entry for unproven switches.

**Alternatives**: Ship full operator bundle verbatim (rejected: violates FR-004/FR-005); experimental opt-in CLI (out of scope for 015).

## R3: Interaction with 014 shipped profile

**Decision**: Phase-3 **appends** after phase-2; do **not** remove `enable-webgl` / `use-angle=swiftshader`. Do **not** add flags that duplicate 013/014 keys.

**Rationale**: Operator list would disable GPU stack that product explicitly enables.

## R4: Bug **-4** blocklist (unchanged)

Phase-3 MUST NOT pass as enabled disables: `disable-background-networking`, background throttling quartet, `disable-remote-fonts`, `disable-reading-from-canvas`, `safebrowsing-disable-auto-update`, `enable-features=NetworkService,...`.

## R5: SC-001 when all proposals reject except one small adopt

**Decision**: If after 015 ship median idle MiB is still within noise of 014 (**&lt;15%** improvement), set `sc001_pass=no` and satisfy alternate SC-001 wording: matrix documents compatibility-driven rejects + measured outcome ([spec SC-001](./spec.md)).

**Outcome (2026-11-23)**: `sc001_pass=no`. Median **1126.2 MiB** vs 014 **1122.1 MiB**. Rejected proposals stay off the default profile (FR-003–FR-005). `memory-pressure-off` remains the only adopt because Chrome accepts it and it does not touch the WebGL or Spribe networking path; it did not cut idle `about:blank` RSS on this host.

## R6: CI vs baseline

**Decision**: `TestHeadlessPhase3AllocatorFlags` (or extend Phase2 test) in default `go test`; SC-001 percent in Baseline Record (015), not CI gate.

## Baseline Record (015)

*(Fill during implementation.)*

| Field | Value |
|-------|-------|
| host | WSL2 linux (same reference host as 014, 2026-11-23) |
| chrome_binary | Google Chrome 157.0.8103.78 (`/usr/bin/google-chrome`); `--memory-pressure-off` accepted (`dump-dom about:blank` exit 0, no unrecognized-switch error) |
| procedure | 014 R1 |
| median_014_mib | 1122.1 (from [014 Baseline](../014-optimizing-browsers-2/research.md), shipped heap 320) |
| median_015_mib | 1126.2 (150 samples × 2s after 30s warmup, 2026-11-23, instance `x6qejddpf73ta`; min 965.5, max 1127.0). About **0.4% higher** than 1122.1, not ≥15% lower. |
| sc001_pass | no |
| phase3_adopted | `memory-pressure-off` |
| validation | `go test ./... -count=1` passed 2026-11-23 (integration 162.7s), including OneWin/Aviator stand-ins. Opt-out covered by `TestHeadlessPhase3AllocatorFlags` / lightweight tier tests. |