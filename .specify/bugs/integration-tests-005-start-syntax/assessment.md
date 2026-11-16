# Bug Assessment: Integration tests fail after 005 start-command-syntax

- **Slug**: integration-tests-005-start-syntax
- **Created**: 2026-11-16T10:10:00-07:00
- **Source**: https://github.com/cloudpayout/supervisor/actions/runs/43845957789/job/133226057321?pr=7 (fetched via `gh run view` — host: `github.com`, policy: **allowlisted**)
- **Verdict**: valid
- **Severity**: medium

## Report (verbatim or summarized)

**User report**: Integration tests are failing after `specs/005-start-command-syntax/` was applied (PR #7).

**CI job**: GitHub Actions workflow `default-checks`, run `43845957789`, step `go test -json ./...`.

**Package outcome**: `github.com/cloudpayout/supervisor/tests/integration` — **FAIL** (22.202s). Unit tests passed.

**Failing tests** (only two):

1. `TestOneWinSecretStaysOffTheRecord` — `onewin_secrets_test.go:32: operator start argv missing the positional path: ""` (elapsed 0.00s)
2. `TestStartRecordsHeadless` — `start_flags_test.go:51: instance rdhjjjbgjtt2m not in registry` (elapsed 5.79s, consistent with `waitPresent`’s 5s poll window)

All other integration tests in that run passed, including headed `start <file>` coverage inside `TestOneWinStartAppliesCookie`.

## Symptom

After the 005 CLI change (`start [path]` with `--headless` only), CI `go test ./...` fails while most start/onewin integration scenarios still pass. Operators and workers appear functionally correct for headless paths; two tests added or extended for 005 behave incorrectly or flakily in CI.

**Expected**: Full integration suite green on `ubuntu-latest` with Chrome (default-checks workflow).

**Observed**: Two failures tied to new assertions (operator `/proc` argv capture and headed session start with multiple concurrent instances in one test).

## Reproduction

1. Check out the PR #7 branch (005 start-command-syntax applied).
2. Install Go per `go.mod` and Google Chrome as in `.github/workflows/default-checks.yml`.
3. From repo root: `go test ./...` (or `go test ./tests/integration/...`).
4. Failures may be intermittent locally; they reproduced on CI run `43845957789`.

Locally on WSL2 (same commit tree), `go test -run 'TestOneWinSecretStaysOffTheRecord|TestStartRecordsHeadless' ./tests/integration/` passed 5 consecutive runs — supports **flakiness / CI load** rather than deterministic product regression.

## Suspected Code Paths

- `tests/integration/onewin_secrets_test.go:22-36` — T019 extension: `exec.Command(bin, "start", file, "--headless")` with `Start()` then immediate `cmdline(t, startCmd.Process.Pid)` via `/proc/<pid>/cmdline`; fails when cmdline is empty.
- `tests/integration/start_flags_test.go:25-63` — T008: `TestStartRecordsHeadless` starts a **headed** blank instance and a **headless** blank instance (both left running via `t.Cleanup`), then runs headed `start <file>` and calls `waitPresent` without `SUPERVISOR_PAGE_URL` contention isolation beyond a fresh registry env.
- `tests/integration/screenshot_test.go:97-112` — `waitPresent` 5s timeout producing the `not in registry` message.
- `internal/cli/start.go:50-128` — Operator `start` (correct per spec): positional path, registry row before worker spawn, worker gets `--session-file` internally only.
- `internal/worker/run.go:52-62` — Worker `defer` removes registry row on any run failure after spawn; explains missing row if headed Chrome/worker exits quickly.

## Root Cause Hypothesis

Confidence: **medium**.

The production `start` implementation matches `specs/005-start-command-syntax/contracts/cli.md` (positional path, no operator `--session-file`). CI failures are **test-harness issues introduced with spec 005**, not incorrect CLI parsing:

1. **`TestOneWinSecretStaysOffTheRecord`**: Task T019 added a requirement that the operator process argv include the session file path and omit `--session-file`. Reading `/proc/<pid>/cmdline` immediately after `Start()` can race (empty cmdline, or PID no longer the `start` process) especially under parallel package pressure on CI. Failure in 0.00s with `""` strongly suggests a **read timing flake**, not that `start` stopped accepting the positional path (other tests use `start <path> --headless` successfully in the same job).

2. **`TestStartRecordsHeadless`**: The test keeps two browser workers alive, then starts a third **headed** session-backed instance and only checks registry **presence**, not `running`. If the worker fails (headed Chrome on CI, resource contention, or cookie/page open error), `worker.Run`’s deferred `remove()` deletes the row within seconds while `start` has already printed an id — matching the exact 5.79s `waitPresent` timeout. The same headed assertion passes in `TestOneWinStartAppliesCookie` because that case uses an isolated env with fewer concurrent browsers.

## Proposed Remediation

**Preferred**: Fix the integration tests; no operator CLI change required unless reproduction proves otherwise.

1. **`onewin_secrets_test.go`**: Replace one-shot `/proc` read with a short retry loop (e.g. poll up to 500ms until cmdline contains the temp session path or `Wait` completes). Alternatively capture argv before `start` returns by using `CombinedOutput` for the secret-value checks and only use `/proc` for worker/screenshot PIDs (which stay alive longer). Assert path as **basename or suffix** if temp dirs differ in normalization. Keep asserting no `--session-file` on operator argv and no cookie value in output.

2. **`start_flags_test.go`**: Restructure `TestStartRecordsHeadless` so headed session cases do not run while two other instances are still active — `stopID` after blank/headless assertions, or use separate `t.Run` subtests each with `isolatedEnv`. For headed `start <file>`, set `SUPERVISOR_PAGE_URL` to a stand-in (as in `onewin_start_test.go`) so worker open is reliable on CI. Prefer `waitRunning` or assert row still present after worker settles if the spec only requires registry field values, not headed window visibility.

3. Re-run `go test ./...` on CI and locally with `-count=10` on the two tests to confirm flake removal.

**Alternatives**:

- Drop operator `/proc` argv assertion from secrets test and rely on usage tests + worker-only `--session-file` (weaker FR-009 coverage).
- Mark headed registry tests manual-only per quickstart SC-003 (loses automated headed coverage).

**Files likely to change**:

- `tests/integration/onewin_secrets_test.go`
- `tests/integration/start_flags_test.go`
- Optionally extract shared `readCmdlineRetry` helper in `tests/integration/onewin_helpers_test.go` or `harness_test.go`

**Tests to add or update**:

- Stabilize existing tests; optional unit-style test that does not depend on `/proc` for operator short-lived processes.
- Stress run: `go test -count=20 -run TestOneWinSecretStaysOffTheRecord ./tests/integration/`

## Risks & Considerations

- Over-eager retries in `/proc` polling could mask a real regression where operator argv omits the path — cap retries and fail if `Wait` errors.
- Stopping instances mid-test increases runtime slightly but reduces CI flake from Chrome process limits.
- No API or registry schema changes; worker `--session-file` remains internal.

## Open Questions

- [NEEDS CLARIFICATION: Should FR-009 require `/proc` inspection of the short-lived operator `start` process, or is verifying worker spawn args via integration logs sufficient?]
- [NEEDS CLARIFICATION: On CI without a display, is headed `start <file>` expected to reach `running`, or is `waitPresent` on a `starting` row the intended contract for SC-003 automation?]