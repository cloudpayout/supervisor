# Bug Fix: CI default-checks missing chrome-headless-shell

- **Slug**: optimizing-browsers-error
- **Fixed**: 2026-11-24T23:12:00Z
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

Default-checks now runs `scripts/fetch-chrome-headless-shell.sh` before `go test` so headless integration tests find `bin/chrome-headless-shell`. Full Chrome install is unchanged for headed tests.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `.github/workflows/default-checks.yml` | modified | New step after Install Chrome |

## Diff Highlights (optional)

```yaml
      - name: Provision chrome-headless-shell
        run: |
          sudo apt-get install -y unzip curl
          chmod +x scripts/fetch-chrome-headless-shell.sh
          ./scripts/fetch-chrome-headless-shell.sh
          test -x bin/chrome-headless-shell
          ./bin/chrome-headless-shell --version
```

## Tests Added or Updated

- None. Assessment: no new product tests; the workflow step is the regression lock.

## Local Verification

- Commands run: workflow YAML edit only. Did not re-run the full integration suite (needs a fresh CI runner and a network fetch of Chrome for Testing).
- Manual checks: `scripts/fetch-chrome-headless-shell.sh` exists and is the same script the new step invokes. Local `bin/chrome-headless-shell` was already provisioned earlier in this worktree.

## Deviations from Assessment

None. Preferred remediation applied. Full `google-chrome-stable` install kept for headed tests (open question in the assessment).

## Follow-ups

- Push and confirm the next `default-checks` run on `017-optimizing-browsers-4` is green.
- If tests still fail after the shell is present, triage allocator flags (`single-process`, `no-sandbox`) separately.
