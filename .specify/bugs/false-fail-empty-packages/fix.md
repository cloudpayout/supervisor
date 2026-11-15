# Bug Fix: Empty packages fail the default checks

- **Slug**: false-fail-empty-packages
- **Fixed**: 2026-11-15
- **Assessment**: ./assessment.md
- **Status**: applied

## Summary

The default-checks job no longer treats a package with no test files as a skipped check. It still fails when `go test` exits non-zero or when a JSON line records `"Action":"skip"` for a named test.

## Changes

| File | Change | Notes |
|------|--------|-------|
| `.github/scripts/default-checks-result.sh` | added | Decides the job result from `test-results.json` and the `go test` status. |
| `.github/workflows/default-checks.yml` | modified | Calls that script instead of grepping every `"Action":"skip"`. |
| `tests/unit/default_checks_result_test.go` | added | Pins package skips, named skips, and a non-zero `go test` status. |
| `specs/003-record-test-results/plan.md` | modified | Testing rule matches the narrower skip check. |
| `specs/003-record-test-results/contracts/check-result.md` | modified | Job contract matches the narrower skip check. |

## Diff Highlights (optional)

The job step now keeps the `go test` status and hands the JSON to the script:

```bash
set -o pipefail
go test -json ./... | tee test-results.json
status=${PIPESTATUS[0]}
bash .github/scripts/default-checks-result.sh test-results.json "$status"
```

A line fails the job only when it contains both keys, in either order:

```bash
if grep -E '"Action":"skip".*"Test":|"Test":.*"Action":"skip"' "$results" >/dev/null; then
  echo "A check was skipped"
  exit 1
fi
exit "$status"
```

## Tests Added or Updated

- `tests/unit/default_checks_result_test.go::TestPackageSkipDoesNotFailDefaultChecks` — package-level `"Action":"skip"` with no `Test` field exits 0.
- `tests/unit/default_checks_result_test.go::TestTestSkipFailsDefaultChecks` — a named skip exits 1, including when `"Test"` appears before `"Action"`.
- `tests/unit/default_checks_result_test.go::TestNonZeroGoTestStatusFailsDefaultChecks` — status 1 with no skip line exits 1 and does not print `A check was skipped`.

## Local Verification

- `bash -n .github/scripts/default-checks-result.sh` → passed.
- The script, run on the same JSON fixtures as the unit tests: package skip → exit 0; named skip → exit 1; reversed keys → exit 1; `go test` status 1 with no skip line → exit 1 and no skip message.
- `go test ./tests/unit/ -run 'TestPackageSkipDoesNotFailDefaultChecks|TestTestSkipFailsDefaultChecks|TestNonZeroGoTestStatusFailsDefaultChecks'` → not run. This module requires Go 1.27.1. The local toolchain is Go 1.26.8, and downloading `go1.27.1` from `proxy.golang.org` timed out.

## Deviations from Assessment

- The decision lives in `.github/scripts/default-checks-result.sh`, which the workflow calls. The assessment named the workflow file. The script is the same rule, so the unit tests execute the code the job runs.
- The matcher accepts `"Test"` before or after `"Action"` on the same line. The assessment described today's `Time`, `Action`, `Package`, `Test` order and asked that a reorder not hide a named skip.

## Follow-ups

- `specs/003-record-test-results/research.md` and `specs/003-record-test-results/quickstart.md` still say any `"Action":"skip"` fails the job.
- Run `go test ./tests/unit/` once Go 1.27.1 is available.
- Re-run the `default-checks` job on a tree that still has untested `cmd/` and `internal/` packages, with Chrome installed, and confirm the job is green when every test passes.
