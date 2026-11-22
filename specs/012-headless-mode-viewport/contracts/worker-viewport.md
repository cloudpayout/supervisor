# Worker Contract: headless viewport (internal)

**Feature**: `012-headless-mode-viewport`  
**Audience**: implementers; not an operator-facing API

## Allocator (`internal/browser/chromedp.go`)

When `AllocatorOptions(profileDir, headless)` is called with `headless == true`:

- Append `chromedp.WindowSize(HeadlessViewportWidth, HeadlessViewportHeight)`.
- Do **not** append `WindowSize` when `headless == false`.

Existing flags (`UserDataDir`, site isolation relaxations, `ExecPath`) unchanged.

## Worker open (`internal/worker/run.go`)

When the registry row indicates headless:

1. Before `openActions` run, include `chromedp.EmulateViewport(HeadlessViewportWidth, HeadlessViewportHeight)` as the first action in the `chromedp.Do` batch (or equivalent single `Do` with ordered actions).
2. Order MUST be: emulate viewport → set cookies (if any) → navigate.

When not headless:

- Do not call `EmulateViewport` in this path.

## Capture path (`internal/worker/run.go` capture callback)

- MUST NOT add viewport emulation or navigate on the screenshot child context (003 contract preserved).
- Assumes viewport was established at open; `CaptureScreenshot` returns 1280×720 for headless.

## Constants

`HeadlessViewportWidth` and `HeadlessViewportHeight` live in `internal/browser` and are the only authoritative dimension source for allocator, worker, tests, and help text.