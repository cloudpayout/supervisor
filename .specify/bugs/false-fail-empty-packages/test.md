# Bug Verification: Empty packages fail the default checks

- **Slug**: false-fail-empty-packages
- **Tested**: 2026-11-15
- **Assessment**: ./assessment.md
- **Fix**: ./fix.md
- **Result**: partial

## Summary

The JSON from the original failing job no longer fails the new check: eight package-level skips and zero named skips exit 0. The same file still fails the old `"Action":"skip"` grep, and a named skip still fails the new script. The Go unit tests and a fresh `go test -json ./...` were not run, and the GitHub Actions job was not re-run.

## Checks Performed

| Check | Command / Action | Result | Notes |
|-------|------------------|--------|-------|
| Reproduction (post-fix) | Replay `0_default-checks.txt` JSON (96 events, 8 package skips, 0 named skips) through `.github/scripts/default-checks-result.sh` with status 0 | pass | New script exits 0. The old `grep -q '"Action":"skip"'` still matches and would have failed the job. |
| Named skip still fails | Same script on a file that has a package skip plus `"Test":"TestParallelStartRegistry"` | pass | Exit 1 and `A check was skipped`. |
| Non-zero `go test` status | Same original JSON with status 1 | pass | Exit 1, and the script does not print `A check was skipped`. |
| New / updated tests | `go test ./tests/unit/ -count=1 -run 'TestPackageSkipDoesNotFailDefaultChecks\|TestTestSkipFailsDefaultChecks\|TestNonZeroGoTestStatusFailsDefaultChecks'` | not-run | Download of Go 1.27.1 was still running with no progress after about 3.5 minutes and was stopped. Local Go is 1.26.8. |
| Regression suite | `go test -json ./...` as the workflow runs it | not-run | Same toolchain download. The GitHub `default-checks` job was not re-run. |
| Lint / type-check | none | skipped | No project lint command was run. |

## Output Excerpts

```text
events 96
package_skips 8
named_skips 0
--- old grep ---
old_grep_would_fail
--- new script status 0 ---
exit=0
--- new script status 1 ---
exit=1
--- mixed package plus named skip ---
exit=1
A check was skipped
```

```text
go: downloading go1.27.1 (linux/amd64)
```

## Residual Risks

- The captured log is the original failing run, not a new `go test -json ./...` on this tree. A new package layout or a real `t.Skip` was not produced locally.
- The unit tests in `tests/unit/default_checks_result_test.go` were not executed by `go test`.
- The hosted job, including Chrome install and the full suite, was not re-run.
- `specs/003-record-test-results/research.md` and `specs/003-record-test-results/quickstart.md` still describe any `"Action":"skip"` as a failure.

## Recommendation

Hold. The original job's JSON no longer fails the check that failed it, and a named skip still fails that check. Close the bug after `go test ./tests/unit/` runs on Go 1.27.1 and the `default-checks` job is green on a tree that still has untested `cmd/` and `internal/` packages.
