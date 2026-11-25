# Research: Further Chrome RAM Reduction (014)

**Date**: 2026-11-23  
**Spec**: [spec.md](./spec.md)  
**Depends on**: [013 research](../013-optimizing-browsers/research.md), [013 Baseline Record](../013-optimizing-browsers/research.md#baseline-record)  
**Plan filters**: Skip flags already in 013 or chromedp defaults; never add flags undone by [bug -4](../../../.specify/bugs/try-limit-of-app-config-exceeded-4/fix.md).

## R1: Baseline procedure (FR-006, SC-001)

**Decision**: Same as 013 R1 (idle headless `about:blank`, 30s warmup, 2s × 150 samples, `status` MEMORY MiB). Record two medians on one host: **profile 013-only** (git parent before 014 merge or temporary disable phase-2 slice) and **profile 013+014** (default `start --headless`, no `--no-lightweight`).

**Rationale**: SC-001 compares 013 vs 014, not pre-013 main.

## R2: 013 already present — do not add again (operator list filter)

These are **already** in `headlessLightweightOptions()`; phase-2 MUST NOT duplicate:

| Flag | 013 value |
|------|-----------|
| `force-device-scale-factor` | `1` |
| `renderer-process-limit` | `1` |
| `process-per-site` | true |
| `disk-cache-size` | `5242880` (014 **replaces** value, see R3) |
| `media-cache-size` | `1` |
| `disable-application-cache` | true |
| `js-flags` | `--max-old-space-size=384` (014 **replaces**, see R4) |
| `v8-cache-options` | `none` |
| `disable-plugins` | true |
| `disable-notifications` | true |
| `disable-permissions-api` | true |
| `disable-speech-api` | true |
| `disable-webrtc` | true |
| `disable-logging` | true |
| `disable-crash-reporter` | true |
| `disable-in-process-stack-traces` | true |
| `disable-component-extensions-with-background-pages` | true |

**Skip — chromedp defaults already on** (no second `Flag` needed): `disable-extensions`, `disable-sync`, `disable-breakpad`, `disable-default-apps`, `disable-dev-shm-usage`, `disable-hang-monitor`, `disable-ipc-flooding-protection`, `disable-popup-blocking`, `disable-prompt-on-repost`, `disable-client-side-phishing-detection`, `metrics-recording-only`, `disable-features=Translate,BlinkGenPropertyTrees` (from defaults; casino `IsolateOrigins` remains separate Supervisor flag).

## R3: Phase-2 adopt matrix (net-new or replaced)

| Flag / change | Verdict | Notes |
|---------------|---------|-------|
| `disk-cache-size=1048576` | **Adopt (replace)** | 1 MiB cap vs 5 MiB (013); single `disk-cache-size` on command line |
| `aggressive-cache-discard` | **Adopt** | Evict cache under pressure; low game risk on session-driven runs |
| `disable-back-forward-cache` | **Adopt** | Reduces BFCache RAM |
| `disable-domain-reliability` | **Adopt** | Telemetry; not needed for automation |
| `disable-features` += `BackForwardCache,MediaRouter,AudioServiceOutOfProcess` | **Adopt (merge)** | One `disable-features` string; MUST retain `IsolateOrigins,site-per-process` from Supervisor; MUST NOT add `enable-features` |
| `js-flags=--max-old-space-size=320` | **Adopt (replace)** | Shipped value. Full-window median 1121.2 MiB |
| `js-flags` with `expose-gc,optimize-for-size,...` | **Reject** | 013 R2 brittle across Chrome versions |
| `js-flags=--max-old-space-size=250` | **Reject (measured)** | Stand-ins passed. Full-window median 1121.9 MiB, no idle win vs 320. Reverted (T030) |
| `disable-gpu` | **Reject** | 013 WebGL guardrail |
| `disable-software-rasterizer` | **Reject** | SwiftShader |
| `disable-translate` | **Skip** | Already in chromedp default `disable-features` |
| `no-zygote` / `single-process` | **Reject** | Stability / iframe risk on WSL2 |
| `disable-shared-workers` | **Reject (measured)** | Stand-ins passed. With heap 250, median 1122.1 MiB. No idle win. Reverted (T030) |

## R4: Bug -4 / Spribe — NEVER add (forbidden list)

From [assessment](../../../.specify/bugs/try-limit-of-app-config-exceeded-4/assessment.md) and `headlessUnset*` arrays — phase-2 MUST **not** pass these switches as enabled disables:

| Flag | Required state on headless |
|------|----------------------------|
| `disable-background-networking` | **Unset** (`false` override) |
| `disable-background-timer-throttling` | **Unset** |
| `disable-backgrounding-occluded-windows` | **Unset** |
| `disable-renderer-backgrounding` | **Unset** |
| `disable-remote-fonts` | **Unset** |
| `disable-reading-from-canvas` | **Unset** |
| `safebrowsing-disable-auto-update` | **Unset** |
| `enable-features=NetworkService,NetworkServiceInProcess` | **Unset** |
| Headless desktop `user-agent` | **Must remain** (do not remove -4 fix) |

**Rationale**: Re-adding Puppeteer-style background suppression or removing UA breaks Spribe config bootstrap (live and stand-in).

## R5: Operator confusion — limits vs flags (FR-007)

**Decision**: Document in [contracts/operator-memory-limits.md](./contracts/operator-memory-limits.md): `--memory-limit` only affects `status` WARNING columns; phase-2 flags reduce typical RSS.

## R6: CI vs baseline (SC-001)

**Decision**: Unit test `TestHeadlessPhase2AllocatorFlags` gates regressions; SC-001 percent improvement recorded in **Baseline Record (014)** below, not default CI.

## Baseline Record (014)

| Field | Value |
|-------|-------|
| host | WSL2 (reference; same as 013 Baseline Record) |
| chrome_binary | CHROME_PATH / PATH (same host as 013) |
| procedure | 013 R1 / spec 014 — idle headless `about:blank`, 30s discard, 2s × 150 samples |
| median_013_mib | 1125.3 same-day control (150 samples × 2s after 30s, 2026-11-23, 013 bundle only: `disk-cache-size=5242880`, `js-flags=384`, no phase-2 slice; min 964.3, max 1126.5). Imported earlier figure 1120.0 is superseded for this comparison. |
| median_014_mib | 1122.1 with heap 250 + `disable-shared-workers` (min 961.1, max 1134.1). Heap 320 alone was 1121.2. Shipped profile is heap 320 without `disable-shared-workers`. |
| sc001_pass | no — (1125.3−1122.1)/1125.3 = 0.3%, under 15%. Idle blank-page RSS on this host does not move with these flags. |
| js_flags_final | `320` |