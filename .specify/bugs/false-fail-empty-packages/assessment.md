# Bug Assessment: Empty packages fail the default checks

- **Slug**: false-fail-empty-packages
- **Created**: 2026-11-15
- **Source**: https://github.com/cloudpayout/supervisor/actions/runs/43572876742/job/132380832868?pr=4
- **Verdict**: valid
- **Severity**: high

## URL fetch

- **URL**: `https://github.com/cloudpayout/supervisor/actions/runs/43572876742/job/132380832868?pr=4`
- **Host**: `github.com`
- **Policy**: allowlisted

The fetched page title is “Record one default test result on each change” for `cloudpayout/supervisor@a14392d`. The converted page did not include step logs. The job log for that same checkout (`a14392d5cccc4f8062403f857ca737d4b77221d6`, branch `003-record-test-results`) shows every real test passing and then `A check was skipped` followed by `Process completed with exit code 1`.

## Report (verbatim or summarized)

User report: packages that have no test files should not be counted, because those packages make the entire test suite fail.

Job log (checkout `a14392d`, workflow `.github/workflows/default-checks.yml`):

- `go test -json ./...` reported `ok` for `github.com/cloudpayout/supervisor/tests/unit` and `github.com/cloudpayout/supervisor/tests/integration`. No test event has `"Action":"skip"` together with a `Test` field.
- Eight packages emitted package-level skips with no `Test` field, each paired with `? … [no test files]`:
  - `github.com/cloudpayout/supervisor/cmd/supervisor`
  - `github.com/cloudpayout/supervisor/internal/browser`
  - `github.com/cloudpayout/supervisor/internal/cli`
  - `github.com/cloudpayout/supervisor/internal/instance`
  - `github.com/cloudpayout/supervisor/internal/metrics`
  - `github.com/cloudpayout/supervisor/internal/proc`
  - `github.com/cloudpayout/supervisor/internal/registry`
  - `github.com/cloudpayout/supervisor/internal/worker`
- The step then printed `A check was skipped` and exited 1.

## Symptom

The `default-checks` job fails after a passing `go test ./...` because packages with no `*_test.go` files produce `"Action":"skip"` in `go test -json` output. The job treats every such line as a skipped check. Expected: packages with no test files do not fail the suite. A green run stays green when every test that exists passed and no test was skipped.

## Reproduction

1. Push or open a pull request that runs `.github/workflows/default-checks.yml` on a tree that still has Go packages under `cmd/` and `internal/` with no `*_test.go` files (the current layout; tests live under `tests/`).
2. Let the job install Go and Chrome, then run `go test -json ./...`.
3. Observe that unit and integration tests pass, JSON still contains `"Action":"skip"` for the packages with no test files, and the step exits 1 with `A check was skipped`.

## Suspected Code Paths

- `.github/workflows/default-checks.yml:37-44` — the step runs `go test -json ./...` and fails when any line matches `"Action":"skip"`. That pattern matches package events for `[no test files]`, which have no `Test` field.
- `cmd/supervisor/main.go` and `internal/{browser,cli,instance,metrics,proc,registry,worker}/` — these packages have no `*_test.go` files, so `go test ./...` records them as skipped packages.
- `tests/integration/harness_test.go:79-83` — `requireChrome` calls `t.Skip("chrome not installed")`. That is a real test-level skip the job is meant to catch. It is not what failed this run (Chrome was installed and those tests passed).
- `specs/003-record-test-results/data-model.md` — `not_successful` is defined as a failed execution or a skip of one of the four browser checks (`TestParallelStartRegistry`, `TestStartStatusStop`, `TestStartRegistersAndReachesRunning`, `TestStatusReflectsConcurrentChanges`). `successful` requires those four to be `passed` or `failed`, not `skipped`.
- `specs/003-record-test-results/contracts/check-result.md` and `specs/003-record-test-results/plan.md` — both say any `"Action":"skip"` fails the job. That wording is wider than the data model and is what the workflow implements.

## Root Cause Hypothesis

`go test -json` emits `"Action":"skip"` for two different cases: a package with no test files (no `Test` field, output `[no test files]`), and a test that called `t.Skip` (a `Test` field is present). The workflow greps for the action alone, so the eight packages that have no tests fail the job even when every test passes. Confidence: high. The failing log, the workflow script, and the package layout all agree.

## Proposed Remediation

**Preferred**: In `.github/workflows/default-checks.yml`, fail the step only when a skip event names a test. Ignore package-level events that mean `[no test files]` (JSON objects with `"Action":"skip"` and no `Test` field). Keep failing the job when `go test` exits non-zero, and when a test is skipped (including `t.Skip("chrome not installed")` from `requireChrome`). `go test -json` emits `Time`, `Action`, `Package`, then `Test` on test skips, so a same-line match of `"Action":"skip"` and `"Test":` distinguishes the two cases.

**Alternatives**:

- Restrict the failure further to the four browser-check names in the data model. That matches `data-model.md` more tightly, and it would also ignore an unrelated `t.Skip` outside those four tests. The report only asks to ignore packages with no test files, so this is a wider behavior change.
- Drop `./...` and test only packages that contain tests. That also avoids the package events, but it changes which packages are compiled by the job and can hide a new package that should have been built.

**Files likely to change**:

- `.github/workflows/default-checks.yml`
- `specs/003-record-test-results/contracts/check-result.md` and `specs/003-record-test-results/plan.md`, if those documents stay the source of the “any `"Action":"skip"`” rule. They currently contradict the data model and the intended fix.

**Tests to add or update**:

- A check that a JSON line `{"Action":"skip","Package":"…"}` with no `Test` field (the `[no test files]` event) does not fail the job.
- A check that a JSON line with `"Action":"skip"` and a `Test` field still fails the job.
- A check that a non-zero `go test` status still fails the job when no skip line is present.
- Re-run `default-checks` on a tree with the current untested `cmd/` and `internal/` packages and confirm the job is green when Chrome is installed and every test passes.

## Risks & Considerations

- Narrowing the grep must not start treating `t.Skip` as success. The feature’s success rule is that a browser-dependent check skipped for a missing browser is not successful.
- `plan.md` and `contracts/check-result.md` currently require failing on any `"Action":"skip"`. Leaving them unchanged would make the workflow disagree with the written contract.
- Field order in `go test -json` is stable today (`Test` after `Action` on test events). A matcher that assumes that order should be noted so a future Go version that reorders fields does not silently stop detecting test skips.
- This job does not gate merge. The impact is a false `not_successful` result on every push and pull request, not a product or data failure.

## Open Questions

- Resolved: Only the four named browser checks should be allowed to use `t.Skip`. Any other test-level skip must still fail the job. Packages with no test files remain excluded.
