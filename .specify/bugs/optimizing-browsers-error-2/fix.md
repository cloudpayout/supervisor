# Bug Fix: headed argv tests fail without a display

- **Slug**: optimizing-browsers-error-2
- **Fixed**: 2026-11-25T00:15:00Z
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

`chromeAllocatorArgs` now sets `DISPLAY=:1` for the duration of each chromedp probe so headed `VisibleWindow` allocation builds a command on headless Linux CI. The bogus `ExecPath` still prevents a real browser launch.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `internal/browser/chromedp_test.go` | modified | Dummy `DISPLAY` under a mutex around argv capture |

## Diff Highlights (optional)

```go
chromeProbeDisplayMu.Lock()
defer chromeProbeDisplayMu.Unlock()
prevDisplay, hadDisplay := os.LookupEnv("DISPLAY")
if err := os.Setenv("DISPLAY", ":1"); err != nil {
    t.Fatal(err)
}
defer func() { /* restore or Unsetenv */ }()
```

## Tests Added or Updated

- No new test. Existing `internal/browser` argv tests now pass when `DISPLAY` and `WAYLAND_DISPLAY` are unset.

## Local Verification

- Commands run: `env -u DISPLAY -u WAYLAND_DISPLAY go test ./internal/browser/ -count=1` → pass
- Commands run: `go test ./internal/browser/ -count=1` → pass (host display left intact)

## Deviations from Assessment

- Did not use `t.Setenv("DISPLAY", ":1")`. Every caller of `chromeAllocatorArgs` calls `t.Parallel()`, and `testing.T.Setenv` panics in parallel tests. A process-wide `os.Setenv` guarded by `chromeProbeDisplayMu` gives the same checkDisplay behavior and restores the previous value.

## Follow-ups

- Push and confirm `default-checks` on PR #16 is green.
- `/speckit-bug-test slug=optimizing-browsers-error-2`
