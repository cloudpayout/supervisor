# Bug Assessment: default-checks fails internal/browser on headless CI runner

- **Slug**: optimizing-browsers-error-2
- **Created**: 2026-11-25T00:12:00Z
- **Source**: https://github.com/cloudpayout/supervisor/actions/runs/45513391668/job/138866553864?pr=16 (host: `github.com`, policy: **allowlisted**); user note: “still failing”
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

**URL trust**: User-supplied Actions job URL — host `github.com`, fetched via `gh` (allowlisted).

**Run**: [45513391668](https://github.com/cloudpayout/supervisor/actions/runs/45513391668) — `pull_request`, branch `017-optimizing-browsers-4`, head SHA `a3a9eb5`, PR #16, **conclusion: failure**.

**Job** `default-checks` (id **138866553864**):

| Step | Result |
|------|--------|
| Install Chrome | success |
| **Provision chrome-headless-shell** | **success** |
| Run the default checks once (`go test -json ./...`) | **failure** |

**Failing package**: `github.com/cloudpayout/supervisor/internal/browser` (~0.01s).  
**Passing**: `tests/integration` (~142s) and most other packages.

**Representative failures** (same message on multiple tests):

```text
chromedp_test.go:49: allocator did not build command
chromedp_test.go:122: allocator did not build command
chromedp_test.go:247: allocator did not build command
chromedp_test.go:349: allocator did not build command
chromedp_test.go:375: allocator did not build command
```

**Tests that still pass on this run** (single `chromeAllocatorArgs` call, headless-only): e.g. `TestHeadlessPhase4AllocatorFlags`, `TestHeadlessPhase3ForbiddenFlags`, `TestAllocatorNoLightweightOmitsFullTier`, `TestHeadlessShellPath`.

**Regression context**: The same `internal/browser` failures already appeared on run [45506879847](https://github.com/cloudpayout/supervisor/actions/runs/45506879847) before shell provisioning; integration failures there were a separate issue (slug `optimizing-browsers-error`, fixed in `0ddd77c`).

## Symptom

After chrome-headless-shell provisioning fixes integration, **default-checks still fails** because unit tests in `internal/browser` abort when exercising **headed** `AllocatorOptions` (`headless=false`, `lightweight=true`) on GitHub’s Linux runners, which have no `DISPLAY` or `WAYLAND_DISPLAY`.

**Expected**: `go test ./...` passes on CI (spec 003 default-checks).

## Reproduction

1. Linux host with `DISPLAY` and `WAYLAND_DISPLAY` unset (matches `ubuntu-latest` Actions).
2. Run:
   ```bash
   DISPLAY= WAYLAND_DISPLAY= go test ./internal/browser/ -count=1
   ```
3. Observe `TestAllocatorOptionsHeadlessWindowSize`, `TestLightweightAllocatorTiers`, etc. fail with `allocator did not build command`.

**Local confirmation** (this worktree, Linux):

```text
DISPLAY= WAYLAND_DISPLAY= go test ./internal/browser/ -run TestAllocatorOptionsHeadlessWindowSize
→ chromedp_test.go:49: allocator did not build command
```

## Suspected Code Paths

- `internal/browser/chromedp_test.go` — `chromeAllocatorArgs()` calls `chromedp.Do(ctx)`; fails when `ModifyCmdFunc` never runs (`args == nil`, lines 82–83).
- `internal/browser/chromedp.go` — `AllocatorOptions(..., headless=false, ...)` appends `chromedp.VisibleWindow` (lines 164–166).
- `github.com/chromedp/chromedp` v0.23.0 — `VisibleWindow` sets `ExecAllocator.visibleWindow`; on Linux `Allocate` returns `ErrNoDisplay` when `DISPLAY` and `WAYLAND_DISPLAY` are empty (`allocate_linux.go`).
- Failing line numbers match **headed** probe calls (e.g. `TestHeadlessPhase3AllocatorFlags` line 247: `headed := chromeAllocatorArgs(..., false, true)`).

## Root Cause Hypothesis

Argv-capture tests use `ExecPath("/do-not-run-chrome")`, but **headed lightweight** paths still enable chromedp **visible-window** mode. On Linux, chromedp refuses allocation before building `exec.Cmd` when no display is configured—even though only argv is needed. Headless-only probes succeed; multi-scenario tests fail on their **headed** branch. Developer machines (e.g. WSL with `WAYLAND_DISPLAY`) often pass. **Confidence: high** (reproduced with cleared display env).

## Proposed Remediation

**Preferred** (product/tests): Avoid requiring a real display for argv inspection.

- In `chromeAllocatorArgs`, set a dummy display for the probe only, e.g. `t.Setenv("DISPLAY", ":1")` before `chromedp.NewExecAllocator`, **or**
- Refactor to build argv without `chromedp.Do` / visible window (testable helper from `AllocatorOptions` flag logic).

The `DISPLAY=:1` approach is minimal: `checkDisplay` only tests env presence; launch still fails on `/do-not-run-chrome`, and `ModifyCmdFunc` records flags.

**Alternatives**:

- **CI xvfb**: `xvfb-run go test` in workflow — runner cost; does not fix bare `go test` on headless Linux.
- **Skip headed probes on `ErrNoDisplay`** — CI passes but loses headed flag coverage.

**Files likely to change**:

- `internal/browser/chromedp_test.go` (primary)
- Optionally `.github/workflows/default-checks.yml` if using xvfb-only mitigation

**Tests to add or update**:

- Existing tests should pass after the helper fix; optional comment that headed argv capture needs no physical display.

## Risks & Considerations

- Dummy `DISPLAY` is test-only; integration headed tests still need Chrome/xvfb if added later.
- Do not change headed **runtime** `AllocatorOptions`—only the test harness.
- `t.Setenv` with `t.Parallel()` is per-test isolated (Go 1.17+).

## Open Questions

- Minimal `t.Setenv("DISPLAY", …)` in `chromeAllocatorArgs` vs. argv refactor? [NEEDS CLARIFICATION: maintainability vs. minimal diff.]