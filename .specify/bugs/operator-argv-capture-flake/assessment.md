# Bug Assessment: Operator start argv capture flakes on CI

- **Slug**: operator-argv-capture-flake
- **Created**: 2026-11-16T04:46:00-06:00
- **Source**: https://github.com/cloudpayout/supervisor/actions/runs/43850666374/job/133242037315?pr=7; user log `c:\Users\oscar\AppData\Local\Temp\0_default-checks (1).txt` (attachment in Cursor: `0_default-checks.txt` from run `43845957789` — older timestamp; treated as supplementary)
- **URL fetch**: host `github.com`, policy **allowlisted** (`gh run view 43850666374 --log-failed`)
- **Verdict**: valid
- **Severity**: medium

## Report (verbatim or summarized)

**User report**: PR #7 `default-checks` still failing after the prior integration-test fix.

**CI run 43850666374** (current):

- Workflow: `default-checks`, step `go test -json ./...`
- **Single failing test**: `TestOneWinSecretStaysOffTheRecord`
- Assertion: `onewin_secrets_test.go:33: operator start argv missing the positional path: ""`
- Elapsed: **0.51s** (matches the `500*time.Millisecond` poll budget in `waitCmdlineContains`)
- **`TestStartRecordsHeadless` — PASS** (1.90s); prior failure from run `43845957789` is no longer reproducing on this branch.

**Related context**: `.specify/bugs/integration-tests-005-start-syntax/` — first assessment/fix addressed registry contention (`start_flags_test.go`) and added `waitCmdlineContains`; `test.md` marked verification **partial** with the same secrets failure under stress.

**Attached log (`0_default-checks.txt`)**: Console export from an **earlier** job (`2026-11-16T10:07–10:08` timestamps) showing **both** `TestOneWinSecretStaysOffTheRecord` and `TestStartRecordsHeadless` failing. Current run `43850666374` (`10:38–10:39`) reflects **post-fix** state: only the secrets test remains red.

## Symptom

**Expected**: `go test ./...` passes on `ubuntu-latest` with Chrome.

**Observed**: Integration package fails because `TestOneWinSecretStaysOffTheRecord` exhausts a 500ms poll of `/proc/<operator-pid>/cmdline` without ever observing the session file path (final captured cmdline `""`). The operator CLI and worker spawn are otherwise healthy (same job passes headless/registry tests and `TestStartRecordsHeadless`).

## Reproduction

1. Branch with 005 start-command-syntax + `waitCmdlineContains` / `stopID` changes (PR #7).
2. Linux with Chrome (as in `.github/workflows/default-checks.yml`).
3. `go test ./tests/integration/ -run TestOneWinSecretStaysOffTheRecord`
4. May flake locally under `go test -count=20` (documented in `integration-tests-005-start-syntax/test.md`).
5. CI run `43850666374` fails deterministically on that test in the sampled job.

## Suspected Code Paths

- `tests/integration/onewin_secrets_test.go:23-37` — `Start()` + `waitCmdlineContains(..., 500ms)` before `Wait()`; failure at call site line 33.
- `tests/integration/onewin_secrets_test.go:98-114` — `waitCmdlineContains`: on successful open, requires non-empty cmdline containing path or basename; on repeated `ReadFile` errors after the operator exits, `last` stays `""` until deadline.
- `internal/cli/start.go:109-128` — Operator spawns worker with `--session-file` on **worker** argv only; prints id and returns quickly (short operator lifetime).
- `specs/005-start-command-syntax/spec.md` FR-009 — cookie values must not leak; session path **MAY** appear on operator argv as positional (not required for security).
- `tests/integration/onewin_usage_test.go` — already asserts operator rejects `--session-file` / legacy flags (exit 2).

## Root Cause Hypothesis

Confidence: **high**.

The failure is a **test harness defect**, not regressed 005 CLI behavior:

1. **`TestStartRecordsHeadless` is fixed** on CI — confirms the prior assessment’s contention hypothesis and the `stopID` remediation were correct.

2. **Operator `/proc` inspection is structurally unreliable** for `supervisor start <path> --headless`. The operator process is short-lived: it spawns the worker, prints the instance id, and exits. The test polls `/proc/<pid>/cmdline` for up to 500ms but often never observes a non-empty buffer containing the path. CI failure at **0.51s** indicates the full budget expired with `last == ""`, consistent with either (a) no successful read while the process was alive, or (b) `/proc` entries gone or unreadable after exit before a match. This matches local stress failures in `integration-tests-005-start-syntax/test.md`.

3. **The 500ms retry does not fix the race** because lengthening the window does not help when the only observable state for that PID is empty or missing cmdline after exit. The product still passes the path positionally (`exec.Command(bin, "start", file, "--headless")` in the test; worker spawn uses internal `--session-file` per `start.go`).

## Proposed Remediation

**Preferred**: Rewrite the FR-009 operator-argv portion of `TestOneWinSecretStaysOffTheRecord` to stop depending on `/proc` for the **short-lived operator** process.

1. **Keep** all secret-leak checks: worker, browser, worker log, registry, screenshot cmdline/filename (longer-lived or file-backed targets).

2. **Replace operator argv capture** with one or both of:
   - **Worker argv after `waitRunning`**: Assert `/proc/<owner-pid>/cmdline` contains `--session-file` and the session path (proves the positional path was honored end-to-end) and does **not** contain the cookie value. FR-009 forbids secrets on worker argv; internal `--session-file` on the worker is expected per `contracts/cli.md`.
   - **Retain operator “no `--session-file`” via usage tests** (`onewin_usage_test.go`) rather than racing `/proc` on a process that may live only milliseconds.

3. **Optional tightening** (if product proof of operator positional is still desired): Start a **background goroutine** immediately when `Start()` returns that reads cmdline with **no initial sleep**, stops when the path appears **or** when `Process.Signal(0)` fails / `Wait` completes; if `Wait` succeeds and the goroutine never saw the path, treat as flaky and fall back to worker argv assertion only. Do **not** rely on post-exit reads of the same PID.

4. **Remove** `waitCmdlineContains` if nothing else uses it, or narrow it to worker/screenshot PIDs only.

No change to `internal/cli/start.go` unless worker argv is wrong (evidence says it is not).

**Alternatives**:

- Instrument operator `start` with test-only logging (rejected: widens product surface).
- Drop T019 operator argv requirement entirely; rely on usage tests + worker argv (acceptable per FR-009 “MAY” on operator path).

**Files likely to change**:

- `tests/integration/onewin_secrets_test.go`

**Tests to add or update**:

- `TestOneWinSecretStaysOffTheRecord` — worker cmdline contains session path on `--session-file`, never cookie; drop or soften operator `/proc` assert.
- Re-run `go test -count=20 -run TestOneWinSecretStaysOffTheRecord ./tests/integration/` and full `./...` before merge.

## Risks & Considerations

- Moving argv proof to the worker checks **internal** `--session-file`, not the literal operator argv shape; usage tests still pin operator rejection of `--session-file`.
- Screenshot subtest still uses one-shot `cmdline` on a short process — may need the same goroutine pattern if it flakes later (out of scope unless CI reports it).

## Open Questions

- [NEEDS CLARIFICATION: Is strict “operator argv must show positional path” still required for T019/FR-009, or is worker spawn + usage tests sufficient given FR-009 “MAY” wording?]

## Prior work

- `.specify/bugs/integration-tests-005-start-syntax/` — original dual-test failure; partial fix and verification remain valid historical context; use **this slug** for the remaining CI blocker.