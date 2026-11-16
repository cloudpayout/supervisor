# Bug Verification: Integration tests fail after 005 start-command-syntax

- **Slug**: integration-tests-005-start-syntax
- **Tested**: 2026-11-16T04:32:27-06:00
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: partial

## Summary

The CI failure on `TestStartRecordsHeadless` (missing registry row under concurrent browsers) appears resolved: the test passes 20 consecutive iterations locally and the full integration package passes once. `TestOneWinSecretStaysOffTheRecord` still flakes under a 20× stress run when the operator `start` process exits before `/proc` cmdline contains the session path within the 500ms poll budget. A single full integration run and `go test ./...` both pass on this machine.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | Compare to CI run 43845957789 failures | skipped | GitHub Actions not re-run from this session |
| New / updated tests (stress) | `go test -count=20 -run 'TestOneWinSecretStaysOffTheRecord\|TestStartRecordsHeadless' ./tests/integration/` | fail | 1–2 iterations failed on secrets test; package exit non-zero |
| `TestStartRecordsHeadless` only (stress) | `go test -count=20 -run TestStartRecordsHeadless ./tests/integration/` | pass | 20/20 |
| Full integration package | `go test -count=1 -timeout 10m ./tests/integration/...` | pass | ~24s |
| Regression suite | `go test -count=1 -timeout 10m ./...` | pass | unit + integration |
| Lint / type-check | `go vet ./tests/integration/...` | pass | no issues |

## Output Excerpts

Stress run (both targeted tests, 20×):

```text
--- FAIL: TestOneWinSecretStaysOffTheRecord (0.50s)
    onewin_secrets_test.go:33: operator start argv missing the positional path: ""
FAIL	github.com/cloudpayout/supervisor/tests/integration	64.456s
```

`TestStartRecordsHeadless` 20×:

```text
ok  	github.com/cloudpayout/supervisor/tests/integration	44.792s
```

Full integration:

```text
ok  	github.com/cloudpayout/supervisor/tests/integration	24.136s
```

## Residual Risks

- Operator argv capture remains timing-sensitive: if `start` finishes in under ~500ms and cmdline never shows the path before the process exits, `waitCmdlineContains` can still fail (matches original CI failure shape with `""`).
- CI (`ubuntu-latest` + Chrome) was not re-executed here; local WSL2 may differ in scheduler load and display/headed behavior.
- `go test ./...` may cache a prior green integration result unless `-count=1` is used (verification used `-count=1` for a fresh integration run).

## Recommendation

Treat **`TestStartRecordsHeadless` as verified** for the contention/registry issue. **Hold closing the bug** until `TestOneWinSecretStaysOffTheRecord` is stabilized further (e.g. longer budget, read cmdline while `start` is still running with a done channel, or assert worker spawn args instead of short-lived operator `/proc`) or until PR CI `default-checks` is green. If CI passes but stress still flakes locally, downgrade priority of argv polling vs. re-open with `/speckit-bug-assess` citing this report. If CI fails again on the secrets test, reopen and adjust the fix before merge.