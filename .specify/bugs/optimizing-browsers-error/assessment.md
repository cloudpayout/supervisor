# Bug Assessment: CI default-checks fail on headless integration tests

- **Slug**: optimizing-browsers-error
- **Created**: 2026-11-24T23:10:00Z
- **Source**: https://github.com/cloudpayout/supervisor/actions/runs/45506868219/job/138844220127 (host: `github.com`, policy: **allowlisted**)
- **Verdict**: valid
- **Severity**: high

## Report (verbatim or summarized)

**URL trust**: User-supplied `https://github.com/cloudpayout/supervisor/actions/runs/45506868219/job/138844220127` — host `github.com`, fetched via `gh` (allowlisted).

**Run**: [Actions run 45506868219](https://github.com/cloudpayout/supervisor/actions/runs/45506868219) on branch `017-optimizing-browsers-4`, job `default-checks` (database id 138844220127), commit message `docs(017): align R2 matrix with operator argv amendment`, **conclusion: failure**.

**Failed step**: `Run the default checks once` — `go test -json ./...` — package `tests/integration` failed (~2.6s, many subtests).

**Representative error** (repeated across OneWin/Aviator, screenshot, start/stop, status, stop-all tests):

```text
start: exit status 1 supervisor: chrome-headless-shell is not available: chrome-headless-shell missing at .../bin/chrome-headless-shell (run scripts/fetch-chrome-headless-shell.sh): stat ...: no such file or directory
```

**Passing contrast**: Headed-path tests and `TestStartLaunchFailureLeavesRegistryUntouched` / `TestStartHeadlessFailsWithoutBundledShell` pass; profile ephemeral tests **skip** with `chrome-headless-shell missing` instead of failing.

## Symptom

GitHub Actions **default-checks** fails because most `tests/integration` cases call `supervisor start --headless`, which now **requires** a repo-local `bin/chrome-headless-shell`. CI installs **full** `google-chrome-stable` but does **not** run `scripts/fetch-chrome-headless-shell.sh`, so preflight rejects headless starts and tests fail.

**Expected**: `go test ./...` passes on CI after spec **017**, matching local workflow where the headless shell is provisioned (see [research R8](../../specs/017-optimizing-browsers-4/research.md)).

## Reproduction

1. Check out branch `017-optimizing-browsers-4` (or any tree with `PreflightHeadless` / bundled shell only).
2. Do **not** create `bin/chrome-headless-shell`.
3. Install Go and `google-chrome-stable` (as in `.github/workflows/default-checks.yml`).
4. Run `go test ./tests/integration/ -count=1`.
5. Observe failures on `TestStartStatusStop`, `TestOneWinStartAppliesCookie`, etc., with the preflight error above.

## Suspected Code Paths

- `.github/workflows/default-checks.yml` — installs Chrome only; no fetch step for headless shell (**primary CI gap**).
- `internal/cli/start.go` — `PreflightHeadless()` before registry write when `--headless` (correct per FR-006/007).
- `internal/browser/executable.go` — `HeadlessShellPath()` errors if `bin/chrome-headless-shell` absent.
- `tests/integration/headless_shell_require.go` — `requireHeadlessShell()` skips; only used by a subset of tests.
- `tests/integration/*` — majority of headless tests invoke CLI `start --headless` without calling `requireHeadlessShell()` first.

## Root Cause Hypothesis

Spec **017** made bundled **chrome-headless-shell** mandatory for headless starts (no PATH/`CHROME_PATH` fallback). The workflow from spec **003-record-test-results** was not updated to provision that artifact. CI therefore exercises the “missing shell” failure mode that product code intentionally returns, causing broad integration test failure. **Confidence: high.**

## Proposed Remediation

**Preferred**: Update `.github/workflows/default-checks.yml` to provision the headless shell before `go test`:

- After checkout (and optionally after Install Chrome), add a step:
  - `chmod +x scripts/fetch-chrome-headless-shell.sh`
  - `./scripts/fetch-chrome-headless-shell.sh`
  - `test -x bin/chrome-headless-shell` (or verify `--version`)
- Ensure job has `curl`/`unzip` (runner `ubuntu-latest` typically has them; script uses `curl` + `unzip`).

This matches [spec 017 research R8](../../specs/017-optimizing-browsers-4/research.md) and [contracts/headless-shell-binary.md](../../specs/017-optimizing-browsers-4/contracts/headless-shell-binary.md).

**Alternatives**:

- **Centralize skip**: Call `requireHeadlessShell(t)` at the start of every headless integration test — CI would **pass** but would **not** run headless coverage on CI (regression risk; contradicts R8 “mandatory prerequisite”).
- **CI env fallback**: Allow full Chrome on CI only — violates FR-007/FR-006 unless spec-amended.

**Files likely to change**:

- `.github/workflows/default-checks.yml`
- Optionally `README.md` or `specs/003-record-test-results/` if workflow is documented there

**Tests to add or update**:

- No new product tests required if workflow fetches shell.
- Optional: document in workflow comment linking to fetch script pin version.

## Risks & Considerations

- Fetch adds ~30–60s and network dependency to CI; pin version is already in script (`157.0.8103.78`).
- Google CfT CDN outage could flake CI — cache artifact or retry policy if needed later.
- After shell is present, some tests may still fail for allocator reasons (`single-process`, etc.) — re-run full suite post-fix.

## Open Questions

- Should `default-checks` keep installing full `google-chrome-stable` for headed integration tests, or rely on shell only? [NEEDS CLARIFICATION: current headed tests use `requireChrome()` / PATH — keep both.]