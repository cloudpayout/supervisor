---
description: "Task list for opening the onewin page and capturing it"
---

# Tasks: Open the onewin page

**Input**: Design documents from `specs/004-open-onewin-page/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: The specification requires the ordinary confirmation in the default test run (`start`, `status`, `screenshot`, `stop` against a local stand-in). Test tasks below are written first and are expected to fail until the implementation task for that story is done.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested on its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- Product code: `cmd/supervisor/`, `internal/cli/`, `internal/worker/`, `internal/browser/`, `internal/registry/`, `internal/casino/`, `internal/session/`
- Tests: `tests/unit/`, `tests/integration/` using `supervisorBin`, `isolatedEnv`, and `requireChrome` from `tests/integration/harness_test.go`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Keep the existing module. This feature adds no dependency.

- [X] T001 Leave `go.mod` on module `github.com/cloudpayout/supervisor` and Go 1.27.1 with no new requirement. The stand-in uses `net/http`. Do not add Python or FastAPI.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The catalog and the registry field every story reads. No user story starts until this phase is complete.

- [X] T002 [P] Add `internal/casino/catalog.go` with one entry. Identifier `onewin`. One or more lowercase ASCII letters. `Onewin` and `1win.com` are refused. Page `https://1win.com/`. The operator does not type it. Game `aviator`. Named. This feature has no command that starts it. One catalog entry. Any other identifier is a usage error. Game identifier `aviator`. One or more lowercase ASCII letters. Casino `onewin`. No runtime state. Nothing in the registry records a game run.
- [X] T003 [P] Add optional `casino` to `internal/registry/types.go`. Schema version stays `1`. The new field is optional, so existing rows still load. Optional string. `onewin` for this feature. Omitted on a blank `start`. Not a secret. The instance row has no cookie, no session file path, and no copy of the page address. A screenshot does not add a field and does not change `state`.

**Checkpoint**: Catalog lookup and a registry row that can name `onewin` without storing a secret.

---

## Phase 3: User Story 1 - Open the onewin page with a session the operator already holds (Priority: P1)

**Goal**: `start --casino onewin --session-file <path>` applies every cookie and opens the registered page. A blank `start` stays a blank page.

**Independent Test**: A local stand-in shows a balance only when the expected cookie for `127.0.0.1` is present. `start` against that page shows it, and a file without that cookie withholds it. The live casino is not opened.

### Tests for User Story 1

- [X] T004 [P] [US1] Add `tests/unit/catalog_test.go`. `onewin` resolves to page `https://1win.com/` and game `aviator`. `Onewin`, `1win.com`, and any other identifier are refused. No function in `internal/casino/catalog.go` starts a game.
- [X] T005 [P] [US1] Add `tests/unit/session_file_test.go` for the parser in `internal/session/file.go`. `cookies` is an array of length at least 1. Zero cookies fails. `name` is a non-empty string. `value` is a non-empty string. `domain` and `path` are strings passed through when present. `expires` is an RFC3339 timestamp and, when set, is the browser expiry. `maxAge` is integer seconds, used only when `expires` is absent, and the expiry is the time of the read plus `maxAge`. `secure`, `httpOnly`, and `partitioned` are booleans. When `partitioned` is true, the browser cookie is partitioned. `sameSite` is `Strict`, `Lax`, or `None`. Any other value fails the parse. Unknown fields fail the parse.
- [X] T006 [P] [US1] Add `tests/integration/onewin_usage_test.go` without Chrome. `start --casino Onewin`, `start --casino 1win.com`, `start --casino onewin` with no file, and `start --session-file` with no casino exit `2` with no instance id. A missing file and a file whose `cookies` length is 0 exit non-zero with no id and no profile.
- [X] T007 [P] [US1] Add `tests/integration/onewin_start_test.go` using `requireChrome`, `isolatedEnv`, and `supervisorBin`. Write it before `internal/cli/start.go` and `internal/worker/run.go` change. The stand-in listens on `127.0.0.1` and an ephemeral port. `GET /` includes the text `balance` and one full-viewport color only when the `Cookie` header contains the name and value generated for this run. Without that cookie it uses a different full-viewport color and omits the text `balance`. Set `SUPERVISOR_PAGE_URL` to that URL. `start --casino onewin --session-file <temp> --headless` exits `0`. The stand-in records that it showed `balance`. A second start without the expected cookie records that it withheld `balance`. A third file with two cookies delivers both names. No navigation uses `1win.com`. A `start --headless` with neither `--casino` nor `--session-file` follows `tests/integration/start_worker_test.go`: the command prints an id, and the registry row reaches `running` with `owner_pid` and `browser_pid` set and `casino` omitted. That start still uses `chromedp.Navigate("about:blank")` in `internal/worker/run.go`, which is how feature 001 opens the blank page.

### Implementation for User Story 1

- [X] T008 [US1] Implement `internal/session/file.go` to the rules tested in `tests/unit/session_file_test.go`. The file stays outside the repository. `start` receives its path. The document is not an argument and is not copied into the project, the registry, or the worker log. `value` is never written to the registry, the command line, the worker log, or an error string.
- [X] T009 [US1] In `internal/cli/start.go`, accept `--casino` and `--session-file` together. Neither flag keeps today's blank start. Exactly one of them exits `2` with no registry write. `--casino` must be `onewin`. Parse the file before preflight. On success, set the row's `casino` to `onewin` and spawn `worker run --id <id> [--headless] --session-file <path>`. Do not put the JSON document in the arguments. A failed spawn still rolls the row back and calls `RemoveProfile`.
- [X] T010 [US1] In `internal/cli/worker.go`, accept `--session-file` and pass the path to `internal/worker/run.go`. The flag is present only when `start` was given `--casino`.
- [X] T011 [US1] In `internal/worker/run.go`, when `casino` is `onewin`, read the session file, call `Network.setCookie` for every cookie, then navigate. The URL is `SUPERVISOR_PAGE_URL` when set, otherwise `https://1win.com/`. A set or navigation error removes the registry row and the profile. Logs and errors omit cookie values.

**Checkpoint**: `start` opens the stand-in with the session file, and a blank `start` is unchanged.

---

## Phase 4: User Story 2 - Capture the page that start opened (Priority: P2)

**Goal**: `screenshot <id>` writes one image of the page `start` opened and leaves the instance running.

**Independent Test**: After the stand-in start, `screenshot` exits `0` within 15 seconds. The shown page includes the text `balance` and the withheld page does not. A pixel is the automated check of which page was captured. `status` still lists the instance.

### Tests for User Story 2

- [X] T012 [US2] Add `tests/integration/onewin_screenshot_test.go` using the stand-in from `tests/integration/onewin_start_test.go`. The row must be `running`. Run `screenshot <id>`. Expect exit `0` within 15 seconds and one PNG. The shown page includes the text `balance`. The withheld page does not. A pixel from the PNG is the automated check: it matches the shown color when the cookie was present and the withheld color when it was not. The default name includes the instance ID and the capture time. It must not include a cookie value. The instance stays `running`. The profile stays until `stop`. Do not add a flag to `internal/cli/screenshot.go`. An output path, an unknown id, and a non-running id stay covered by `specs/003-screenshot-cli` and are not retested here.

**Checkpoint**: The existing screenshot command captures the opened page. `internal/cli/screenshot.go` has no new flag.

---

## Phase 5: User Story 3 - See the run and end it (Priority: P3)

**Goal**: `status` lists the running instance. `stop` ends it and removes the profile after the browser has exited. `screenshot` does not end it.

**Independent Test**: `start`, `status` lists the id, `screenshot` leaves it listed, `stop` removes the id and the profile directory.

### Tests for User Story 3

- [X] T013 [P] [US3] Add `tests/integration/onewin_stop_test.go`. After a headless `onewin` start, `status` lists the id as running. `screenshot <id>` leaves that row running. `stop <id>` exits `0`, a later `status` omits the id, and `profile_dir` is gone. A start that fails after the profile directory exists leaves no row and no directory.

### Implementation for User Story 3

- [X] T014 [P] [US3] In `internal/cli/stop.go`, call `browser.RemoveProfile` only after `proc.Terminate` and `proc.KillByProfile` have returned. The directory is removed after the browser process has exited.
- [X] T015 [P] [US3] In `internal/registry/reconcile.go`, call `browser.RemoveProfile` after `KillByProfile` inside `Apply`, so a browser that exits outside `stop` loses its profile too. `RemoveProfile` stays safe to call more than once.
- [X] T016 [US3] In `internal/worker/run.go`, on every exit cancel the browser allocator, wait until the browser PID is not alive, call `RemoveProfile`, then drop the registry row. Do this for a cookie or navigation error as well.

**Checkpoint**: Ending an instance removes the profile. `screenshot` does not.

---

## Phase 6: User Story 4 - Keep the session out of the project and the command (Priority: P4)

**Goal**: The generated cookie value is absent from command text, the worker log, the registry, the project, and the screenshot file name.

**Independent Test**: Run `start` and `screenshot` with a random value stored only in a temporary file. The value is absent from those five places. The session file path may appear on the `start` command line.

### Tests for User Story 4

- [X] T017 [US4] Add `tests/integration/onewin_secrets_test.go`. Generate the cookie value at runtime. After `start` and `screenshot`, read `/proc/<pid>/cmdline` for the worker and for the screenshot process, the worker log, the registry file, and the screenshot file name. The value is absent from all of them. The registry `casino` field is `onewin` and the row has no cookie, no session file path, and no page address. The PNG is not left under the repository.

### Implementation for User Story 4

- [X] T018 [US4] In `internal/cli/start.go` and `internal/worker/run.go`, make every error for a bad file or a failed `SetCookie` omit the cookie value. `value` is never written to the registry, the command line, the worker log, or an error string. The document is not an argument and is not copied into the project, the registry, or the worker log.

**Checkpoint**: A known test secret cannot be recovered from the command line, the log, the registry, or the image name.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: The default run matches the quickstart, and new behavior cites the spec.

- [X] T019 [P] Walk `specs/004-open-onewin-page/quickstart.md` against `specs/004-open-onewin-page/contracts/cli.md` and `specs/004-open-onewin-page/contracts/session-file.md`. Fix any mismatch in the tests under `tests/integration/`. The live-page section stays manual and stays out of `go test`.
- [X] T020 Run `go test ./...` from the repository root. The new confirmation is part of that run and does not open `1win.com`.
- [X] T021 Cite the spec scenario or task, in the existing FR-comment style, on the new behavior in `internal/casino/catalog.go`, `internal/session/file.go`, `internal/cli/start.go`, `internal/cli/worker.go`, `internal/worker/run.go`, `internal/cli/stop.go`, and `internal/registry/reconcile.go`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: T001 has no dependency.
- **Foundational (Phase 2)**: T002 and T003 depend on T001 and block every user story. They touch different files and can run together.
- **User Story 1 (Phase 3)**: Depends on Phase 2. T004, T005, T006, and T007 are the tests and can run together. T008 after T004, T005, and T006. T009 after T008. T010 after T009. T011 after T008 and T010. T007 is written before T009 and T011.
- **User Story 2 (Phase 4)**: Depends on User Story 1. T012 is the only task.
- **User Story 3 (Phase 5)**: Depends on User Story 1. T013 first. T014 and T015 can run together after T013. T016 after T011 and T013 because it edits `internal/worker/run.go`.
- **User Story 4 (Phase 6)**: Depends on User Stories 1 and 2. T017 first. T018 after T017, T009, and T011.
- **Polish (Phase 7)**: Depends on the stories being present. T020 after T019. T021 after the implementation tasks.

### User Story Dependencies

- **User Story 1 (P1)**: Starts after Phase 2. No dependency on the other stories.
- **User Story 2 (P2)**: Needs a running `onewin` page from User Story 1. It does not change `internal/cli/screenshot.go`.
- **User Story 3 (P3)**: Needs an instance from User Story 1. Profile removal is independent of the screenshot pixel check.
- **User Story 4 (P4)**: Needs `start` and `screenshot` so the secret check has a command line, a log, a registry row, and a file name to read.

### Parallel Opportunities

- T002 and T003 are different files.
- T004, T005, T006, and T007 are different test files.
- T014 and T015 are different files.
- T016 is not parallel with T011 or T018. All three edit `internal/worker/run.go`.

---

## Parallel Example: User Story 1

```text
T004 tests/unit/catalog_test.go
T005 tests/unit/session_file_test.go
T006 tests/integration/onewin_usage_test.go
T007 tests/integration/onewin_start_test.go
```

Those four tests are written before `internal/cli/start.go` and `internal/worker/run.go`. T008 (`internal/session/file.go`) follows the parser test. T009, T010, and T011 stay in that order because each passes the session path into the next process.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Finish Phase 1 and Phase 2.
2. Finish Phase 3.
3. Stop and run `tests/integration/onewin_start_test.go` and `tests/integration/onewin_usage_test.go`.
4. A later story can wait. The page open is the decision's outcome.

### Incremental Delivery

1. User Story 1 opens the page on the stand-in.
2. User Story 2 captures that page with the existing `screenshot` command.
3. User Story 3 removes the profile when the instance ends.
4. User Story 4 checks that the cookie value never lands in the command, the log, the registry, or the file name.
5. Phase 7 runs `go test ./...`.

### Parallel Team Strategy

After Phase 2, one person owns User Story 1 through T011. User Stories 2, 3, and 4 start only after T011. T014 and T015 can then proceed beside the screenshot test, as long as `internal/worker/run.go` has a single editor.

---

## Notes

- Checkbox, ID, optional `[P]`, story label on story tasks only, and a file path are required on every task.
- Quote the data-model rule in the task that implements it. Do not invent a second cookie schema.
- `internal/cli/screenshot.go` stays as specified in `specs/003-screenshot-cli`. An output path, an unknown id, and a non-running id stay covered by that feature's tests.
- The live `1win.com` check in `specs/004-open-onewin-page/quickstart.md` is manual, including the visible-window screenshot. It is not a task in `go test ./...`. The default test run stays headless.

---

## Phase 8: Convergence

- [X] T022 CRITICAL: In the onewin confirmation, run `supervisor status` after `start`, again after `screenshot`, and again after `stop`, for both the shown and withheld stand-in runs. The id is listed as running until `stop`, and a later `status` omits it. per Constitution Delivery Workflow (partial)
- [X] T023 CRITICAL: Cite the spec on the new behavior in `internal/cli/worker.go` (`--session-file`), `internal/cli/stop.go` (profile removal), and `internal/registry/reconcile.go` (profile removal). per Constitution I (partial)
