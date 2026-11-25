# Bug Verification: default-checks fails internal/browser on headless CI runner

- **Slug**: optimizing-browsers-error-2
- **Tested**: 2026-11-25T00:16:00Z
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: partial

## Summary

The assessment reproduction (`DISPLAY` and `WAYLAND_DISPLAY` unset on Linux) no longer fails: headed argv probes pass under parallel load and the full `go test ./...` suite passes locally. The fix is not yet on the remote branch, so GitHub **default-checks** has not been re-run with this change.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | `env -u DISPLAY -u WAYLAND_DISPLAY go test ./internal/browser/ -count=1` | pass | Matches assessment steps 1–3 |
| Former CI failures | Same env; `-run` subset (`TestAllocatorOptionsHeadlessWindowSize`, `TestLightweightAllocatorTiers`, …) | pass | No `allocator did not build command` |
| Parallel stress | `env -u DISPLAY -u WAYLAND_DISPLAY go test ./internal/browser/ -parallel=16 -count=1` | pass | All package tests green |
| Regression suite | `go test ./... -count=1 -timeout 10m` | pass | `internal/browser` ~0.05s; integration ~149s |
| New / updated tests | N/A (helper change only per fix.md) | skipped | — |
| Lint / type-check | Covered by `go test` compile | pass | — |
| GitHub default-checks | Run after push on PR #16 | not-run | Fix uncommitted locally as of verification |

## Output Excerpts

```text
$ env -u DISPLAY -u WAYLAND_DISPLAY go test ./internal/browser/ -count=1 -parallel=16
--- PASS: TestAllocatorOptionsHeadlessWindowSize (0.01s)
--- PASS: TestLightweightAllocatorTiers (0.03s)
--- PASS: TestHeadlessPhase3AllocatorFlags (0.03s)
PASS
ok  	github.com/cloudpayout/supervisor/internal/browser	0.042s

$ go test ./... -count=1 -timeout 10m  (tail)
ok  	github.com/cloudpayout/supervisor/internal/browser	0.051s
ok  	github.com/cloudpayout/supervisor/tests/integration	148.990s
```

## Residual Risks

- `chromedp_test.go` change is local only; last Actions failure remains [run 45513391668](https://github.com/cloudpayout/supervisor/actions/runs/45513391668).
- Mutex + process-wide `DISPLAY` is safe for current callers but serializes argv probes (acceptable for unit runtime).

## Recommendation

Commit and push the `internal/browser/chromedp_test.go` fix (and optional `.specify/bugs/optimizing-browsers-error-2/` docs), then confirm **default-checks** green on PR #16. After a green run, treat the bug as closed end-to-end. If CI still fails, rerun `/speckit-bug-assess` with the new log.