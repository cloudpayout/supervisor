# Bug Verification: Operator start argv capture flakes on CI

- **Slug**: operator-argv-capture-flake
- **Tested**: 2026-11-16T04:58:04-06:00
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: partial

## Summary

The CI failure mode (`operator start argv missing the positional path: ""` after a 500ms `/proc` poll) no longer occurs locally: `TestOneWinSecretStaysOffTheRecord` passed 20 consecutive iterations and the full integration and repo test suites passed. GitHub Actions `default-checks` on PR #7 was not re-run in this session, so merge confidence still depends on a green CI job.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | Stress the former failing assertion | pass | No `missing the positional path` failures; operator `/proc` code removed |
| New / updated tests | `go test -count=20 -timeout 15m -run TestOneWinSecretStaysOffTheRecord ./tests/integration/` | pass | `ok`, 27.025s |
| Regression suite (integration) | `go test -count=1 -timeout 10m ./tests/integration/...` | pass | 27.003s |
| Regression suite (repo) | `go test -count=1 -timeout 10m ./...` | pass | integration 27.913s, unit 0.112s |
| Lint / type-check | `go vet ./tests/integration/...` | pass | no output |
| CI (run 43850666374 class) | `default-checks` on `ubuntu-latest` | skipped | Not triggered from this environment |

## Output Excerpts

```text
ok  	github.com/cloudpayout/supervisor/tests/integration	27.025s
```

```text
ok  	github.com/cloudpayout/supervisor/tests/integration	27.913s
ok  	github.com/cloudpayout/supervisor/tests/unit	0.112s
```

Prior CI failure (pre-fix, run `43850666374`) for comparison:

```text
onewin_secrets_test.go:33: operator start argv missing the positional path: ""
--- FAIL: TestOneWinSecretStaysOffTheRecord (0.51s)
```

## Residual Risks

- Screenshot subtest still reads `/proc` on a short-lived `screenshot` process (one shot, not stressed here).
- Worker argv check assumes Linux `/proc`; matches existing integration harness.
- Operator positional path is not asserted on operator argv anymore; covered indirectly via worker spawn plus `onewin_usage_test.go` for legacy flags.

## Recommendation

Treat the fix as **ready for CI**: push or re-run `default-checks` on PR #7. If that job is green, close `operator-argv-capture-flake` and treat `integration-tests-005-start-syntax` as resolved for the remaining secrets failure. If CI fails again, reopen with the new log and `/speckit-bug-assess`.