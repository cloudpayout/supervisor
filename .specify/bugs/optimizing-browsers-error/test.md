# Bug Verification: CI default-checks fail on headless integration tests

- **Slug**: optimizing-browsers-error
- **Tested**: 2026-11-25T00:30:00Z
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: partial

## Summary

With `bin/chrome-headless-shell` present, the full local suite including `tests/integration` passes; the preflight `chrome-headless-shell missing` failures from run 45506868219 do not reproduce. The workflow step in `.github/workflows/default-checks.yml` is present locally but was not executed on a fresh GitHub Actions runner in this validation.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | `go test ./tests/integration/ -count=1 -run 'TestStartStatusStop\|TestStartRegistersAndReachesRunning\|TestStartRecordsHeadless\|TestStartHeadlessFailsWithoutBundledShell'` with existing `bin/chrome-headless-shell` | pass | Subset mirrors former CI failures; no missing-shell preflight errors |
| Full integration (post-fix) | `go test ./... -count=1 -timeout 10m` | pass | `tests/integration` ~148s; all packages ok |
| New / updated tests | N/A (workflow-only fix per fix.md) | skipped | — |
| Regression suite | `go test ./... -count=1 -timeout 10m` | pass | Same run as full integration |
| Lint / type-check | `go test ./internal/browser/ -run Headless` | pass | Headless allocator unit tests |
| CI workflow step | GitHub Actions `default-checks` on PR #16 after push | not-run | Requires push + network fetch on `ubuntu-latest` |

## Output Excerpts

```text
$ test -x bin/chrome-headless-shell && echo SHELL_OK
SHELL_OK

$ go test ./tests/integration/ -count=1 -timeout 5m -run 'TestStartStatusStop|...'
ok  	github.com/cloudpayout/supervisor/tests/integration	5.704s

$ go test ./... -count=1 -timeout 10m  (tail)
ok  	github.com/cloudpayout/supervisor/tests/integration	148.315s
```

Workflow fix (verified in tree, not executed on Actions):

```yaml
      - name: Provision chrome-headless-shell
        run: |
          sudo apt-get install -y unzip curl
          ...
          ./scripts/fetch-chrome-headless-shell.sh
          test -x bin/chrome-headless-shell
```

## Residual Risks

- Next `default-checks` run on GitHub has not been observed green in this session.
- CfT CDN fetch could flake CI despite local success with a pre-provisioned shell.
- Unrelated allocator failures (`single-process`, `no-sandbox`) remain possible if they appear only on CI runners.

## Recommendation

Treat the missing-shell root cause as fixed for local parity: provision the shell, then run `go test ./...`. Push the workflow change (if not already on the remote branch) and confirm run **default-checks** on `017-optimizing-browsers-4` / PR #16 is green; upgrade to **verified** once Actions passes. If CI still fails after the fetch step, rerun `/speckit-bug-assess` with the new log.