# Research: Headless Viewport 1280×720

**Feature**: `012-headless-mode-viewport` | **Date**: 2026-11-22

## R1: Where to apply viewport (allocator vs CDP)

**Decision**: Apply **both** when `headless` is true for the instance:

1. `chromedp.WindowSize(1280, 720)` in `browser.AllocatorOptions` (exec allocator flag for initial window geometry in headless Chrome).
2. `chromedp.EmulateViewport(1280, 720)` as the **first** chromedp action in the worker open sequence, **before** cookie injection and `Navigate` (FR-001).

**Rationale**: Headless Chrome’s default window metrics are smaller than 1280×720; allocator `window-size` alone does not always align layout/CSS metrics with screenshot capture. `EmulateViewport` sets device metrics override (layout width/height) that game DOM probes and `CaptureScreenshot` consume (003: current viewport capture). Running emulation before navigation satisfies edge cases (blank start, casino open, Aviator entry after `running`).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Allocator `WindowSize` only | Layout/CSS size can still differ; flaky iframe sizing |
| `EmulateViewport` only after `Navigate` | Violates FR-001 “before substantive navigation”; first paint may be wrong size |
| Only in `openActions` without allocator | Screenshot surface may not match 1280×720 on some Chrome builds |
| Env var / CLI override | Out of scope FR-004 for v1 |

## R2: Constants and package placement

**Decision**: Define `HeadlessViewportWidth = 1280` and `HeadlessViewportHeight = 720` in `internal/browser` (same package as `AllocatorOptions`). Worker imports these constants for `EmulateViewport`; no registry fields (FR-005).

**Rationale**: Single source of truth for tests and help text; keeps worker from magic numbers.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Constants only in `worker` | Allocator and worker would duplicate or cross-import awkwardly |
| `internal/registry` fields | Spec forbids persisted viewport for v1 |

## R3: Headed instances

**Decision**: When `headless` is false, do not append `WindowSize(1280,720)` to allocator options and do not run `EmulateViewport` in the worker open path (FR-002).

**Rationale**: Visible window uses `chromedp.VisibleWindow` today; operators may resize. Forcing 1280×720 would change headed debugging behavior.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Same viewport for headed for consistency | Spec User Story 2 explicitly excludes headed override |
| EmulateViewport only in headless but WindowSize always | Would resize headed windows |

## R4: Integration verification (no live casino)

**Decision**: Add `tests/integration/headless_viewport_test.go`: `requireChrome`, `isolatedEnv`, `start --headless` (blank), wait `running`, `screenshot <id>`, decode PNG with `image/png`, assert `Bounds().Dx() == 1280` and `Dy() == 720` (FR-006, SC-005). Reuse existing screenshot wait helpers where possible.

**Rationale**: Matches spec “screenshot metadata” path; no stand-in HTTP server required for dimensions. Default `go test ./...` stays off live casino.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `window.innerWidth` via stand-in page only | Extra server; screenshot already encodes viewport contract |
| Unit test mock chromedp | Does not validate real Chrome headless surface |

## R5: Operator documentation

**Decision**: Extend `start` flag description and/or `Long` help one sentence: headless runs use a 1280×720 layout (FR-007). No new flags.

**Rationale**: Spec User Story 3; matches 005 help patterns.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| New `supervisor help viewport` | Over-scoped for v1 |
| README only | Constitution: CLI is operator surface |

## R6: Capture path interaction

**Decision**: No change to capture socket protocol or `ServeCapture` mutex/timing. Viewport is set before `StateRunning` and before capture listener accepts requests, so existing `chromedp.CaptureScreenshot` on `taskCtx` inherits 1280×720 (FR-003).

**Rationale**: 003 forbids viewport emulation from capture child path; emulation belongs in worker startup only.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Re-emulate on each capture | Racey; violates 003 “do not call viewport emulation from capture path” spirit |

## R7: Aviator / layout-sensitive scripts

**Decision**: No separate Aviator code change in v1 unless integration proves remaining failure after viewport fix. Spec SC-003 allows documenting unrelated failures; larger layout should improve `internal/worker/aviator.go` rectangle thresholds that already require minimum sizes.

**Rationale**: Viewport increase is horizontal fix; Aviator logic unchanged unless tasks find a gap.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Lower Aviator size thresholds in same feature | Mixes layout policy with game heuristics |