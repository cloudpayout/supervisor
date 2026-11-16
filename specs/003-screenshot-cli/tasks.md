---
description: "Task list for screenshot CLI (003)"
---

# Tasks: Screenshot CLI

**Input**: Design documents from `specs/003-screenshot-cli/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/cli.md](./contracts/cli.md), [contracts/capture-socket.md](./contracts/capture-socket.md), [quickstart.md](./quickstart.md)

**Tests**: Included. The constitution requires automated coverage of CLI behavior before it is complete, and the plan names the test files under `tests/unit/` and `tests/integration/`. Default tests must not open a live casino and must not require a visible window. Write a story's tests so they fail before that story's implementation tasks.

**Organization**: By user story (US1 capture a running instance, US2 refuse a capture that cannot succeed, US3 choose the output path).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no unfinished dependencies)
- **[Story]**: US1, US2, or US3

## Path Conventions

Single Go module at repository root: `internal/cli/`, `internal/worker/`, `internal/registry/`, `tests/unit/`, `tests/integration/`, `cmd/supervisor/`. Git branch `feature/screenshot-command`. Feature directory `specs/003-screenshot-cli`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Lock the operator contract and the worker-capture decision before editing code

- [x] T001 Review specs/003-screenshot-cli/contracts/cli.md, specs/003-screenshot-cli/contracts/capture-socket.md, and specs/003-screenshot-cli/research.md: capture stays on the worker's single blank-page tab; a missing row is `instance not found`; a tracked non-running row is `instance not running`; usage errors exit 2; `instances.json` schema version stays 1

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Socket path and command entry shared by every story

**⚠️ CRITICAL**: No user story work begins until this phase is complete

- [x] T002 [P] Add `Layout.CaptureSocket(id string) string` in internal/registry/paths.go returning `filepath.Join(l.StateDir, "sockets", id+".sock")`. Callers must refuse an id that is not a single path segment (`filepath.Base(id) != id`) before joining, per data-model.md: "Must be a single path segment before it is used in a filesystem path (`filepath.Base(id) == id`)"
- [x] T003 [P] Register `newScreenshotCmd()` from internal/cli/root.go and add internal/cli/screenshot.go with `Use: "screenshot <instance-id>"`, `Args: cobra.ExactArgs(1)`, and a trimmed-empty ID returning `fmt.Errorf("instance ID required")` so internal/cli/root.go `isUsageError` exits 2. Do not read stdin. Data-model.md: "Required positional argument. Non-empty after trim" and "stdin is not read"

**Checkpoint**: `supervisor screenshot` with no ID exits 2. `supervisor screenshot some-id extra` exits 2. The socket path is a pure function of the state dir and ID. User stories can start.

---

## Phase 3: User Story 1 — Capture a running instance (Priority: P1) 🎯 MVP

**Goal**: `supervisor screenshot <instance-id>` writes one PNG of that running instance's current page into the current directory and leaves that same session running (FR-001, FR-004, FR-005, FR-006, FR-008, FR-009, FR-012).

**Independent Test**: Start one headless instance, wait until status shows `running`, run `screenshot` with that ID, and confirm one openable PNG, exit 0 within 15 seconds, and the same owner and browser process are still running on the blank page.

### Tests for User Story 1

- [x] T004 [P] [US1] Add a Chrome-free socket test in tests/unit/screenshot_test.go that dials `ServeCapture` with a fake PNG payload: request line is exactly `CAPTURE\n`; success is `OK <nbytes>\n` plus those bytes; the payload must start with `\x89PNG`; a payload over 32 MiB or any other first line returns `ERR <single-line reason>\n` with no payload (specs/003-screenshot-cli/contracts/capture-socket.md)
- [x] T005 [P] [US1] Add `TestScreenshotDefaultWritesPNG` in tests/integration/screenshot_test.go using `requireChrome`, `isolatedEnv`, and `supervisorBin`: `start --headless`, wait until the registry row is `running`, record `owner_pid` and `browser_pid`, run `screenshot <id>`, assert exit 0 within 15 seconds, stdout is one relative name `screenshot-<id>-<YYYYMMDD-HHMMSS>.png`, the file is a non-empty PNG, and the same row is still `running` with those same PIDs. Run a second capture and assert a different file whose creation does not change the first file's bytes or the PIDs (SC-001, SC-003, SC-005; User Story 1 scenario 2)
- [x] T006 [US1] Add `TestScreenshotTwoHeadlessInstances` in tests/integration/screenshot_test.go: start two `--headless` instances, wait until both are `running`, record both PID pairs, screenshot each ID, assert both commands exit 0 with a non-empty PNG, and both rows stay `running` with the same PIDs. Do not require the two blank-page PNGs to differ. Do not start a headed browser (User Story 1 scenarios 3 and 4; FR-008, FR-009)

### Implementation for User Story 1

- [x] T007 [US1] Implement `ServeCapture(ctx, socketPath, capture func(context.Context) ([]byte, error)) error` in internal/worker/capture.go. Create the parent directory at mode `0700` and the socket at mode `0600`. Read one `CAPTURE\n` line per connection, run `capture` under a mutex, and write either `OK <nbytes>\n` plus the PNG or `ERR <single-line reason>\n`. Cap the PNG at 32 MiB. Close the listener and remove the socket file when `ctx` is done (data-model.md: "Directory `0700`, socket `0600`"; "Created after the worker sets `state` to `running`. Removed when the worker shuts down.")
- [x] T008 [US1] After the registry update sets `StateRunning` in internal/worker/run.go, start `ServeCapture` with `chromedp.Run(context.WithTimeout(taskCtx, 10*time.Second), chromedp.CaptureScreenshot())` on the worker's existing tab. v1 has one page, the blank page from start; do not select another tab. The 10-second timeout must be a child of `taskCtx`, not a replacement for it, so the browser stays up. Do not call `Navigate`, click, viewport emulation, or `taskCancel` from this path. On `taskCtx` done, stop the listener and remove the socket before `remove(reg, id)` (research.md R1 and R4; FR-005, FR-009)
- [x] T009 [US1] In internal/cli/screenshot.go, `View` the registry and, only when `state` is `running`, dial `Layout.CaptureSocket` inside a 15-second command context. On `OK`, write a same-directory temporary file with `O_EXCL`, `os.Link` it to `screenshot-<id>-<YYYYMMDD-HHMMSS>.png` in the working directory using local time, and if that name exists try `-2`, `-3`, … before `.png`. Remove the temporary file after success and on every failure. Print only the relative file name on stdout and return nil (data-model.md: "Default name `screenshot-<instance-id>-<YYYYMMDD-HHMMSS>.png` in the current directory, local time. Collision suffix `-2`, `-3`, … before `.png`"; "Temporary file in the destination directory, then `os.Link` to the final name. The temporary file is removed on success and on failure"; FR-004, FR-006, FR-012)
- [x] T010 [US1] Set `Short` and `Long` help in internal/cli/screenshot.go so `screenshot --help` names the required instance ID and the default file location and name. Leave `--output` / `-o` help to T018 (User Story 1 scenario 6)

**Checkpoint**: `go test ./tests/integration/ -run 'TestScreenshotDefaultWritesPNG|TestScreenshotTwoHeadlessInstances'` passes when Chrome is installed, and the socket unit test passes without Chrome. MVP is a default-path capture. Visible-window capture stays in quickstart Scenario 2.

---

## Phase 4: User Story 2 — Refuse a capture that cannot succeed (Priority: P1)

**Goal**: A missing, extra, unknown, or non-running ID fails with a distinct message and creates no image. A capture that cannot finish does not leave a partial PNG or change the instance (FR-002, FR-003, FR-010, FR-011).

**Independent Test**: Run `screenshot` with no ID, with an extra argument, with `missing-id`, and with a tracked row whose state is not `running`. Each attempt fails, explains why, and creates no PNG.

### Tests for User Story 2

- [x] T011 [P] [US2] Add tests in tests/integration/screenshot_refuse_test.go that do not call `requireChrome`: no arguments, a blank ID, and `screenshot some-id extra` exit 2 with no new PNG; `screenshot missing-id` writes stderr `instance not found: missing-id` and exits 1; a seeded row with `state` `starting` and a seeded row with `state` `stopped` each write stderr `instance not running: <id>` and exit 1 with no PNG (contracts/cli.md; FR-003, FR-010; SC-002, SC-004)
- [x] T012 [P] [US2] Add a Chrome-free test in tests/unit/screenshot_publish_test.go for the publish helper in internal/cli/screenshot.go: when the final link fails or the write fails, the temporary file is removed, the final path is not created, and an existing destination's bytes are unchanged (FR-011; User Story 2 scenario 4)

### Implementation for User Story 2

- [x] T013 [US2] In internal/cli/screenshot.go, before any dial: no registry row returns exit class 1 with stderr `instance not found: <id>`; a row whose `state` is not `running` (`starting`, `stopping`, or `stopped`) returns exit class 1 with stderr `instance not running: <id>`. Use `Store.View` only. Do not call `registry.Apply`, do not write the registry, and do not create a file (data-model.md: "No registry row → Not found, no file" and "`state != running` → Not running, no file"; research.md R3; FR-010)
- [x] T014 [US2] Export the publish helper used by T012 from internal/cli/screenshot.go. A dial failure, `ERR` line, short read, payload that does not start with `\x89PNG`, or publish failure returns an error whose text is `screenshot failed: <id>: <reason>` so cmd/supervisor/main.go exits 1. Delete the temporary file. Do not print a path and do not change the registry row or the page (FR-011; contracts/cli.md)

**Checkpoint**: The refuse and failed-publish tests pass without Chrome. A successful US1 capture still leaves the same PIDs running.

---

## Phase 5: User Story 3 — Choose where the image is saved (Priority: P2)

**Goal**: `--output` / `-o` writes only at the given relative or absolute path, prints that path unchanged, and refuses to overwrite (FR-007, SC-006). Help describes that flag.

**Independent Test**: Capture with `-o shots/page.png` where the parent exists and the file does not. The image is only there, stdout is `shots/page.png`, and repeating the command leaves the existing file untouched and exits 1.

### Tests for User Story 3

- [x] T015 [P] [US3] Add tests in tests/unit/screenshot_output_test.go for the exported output helper in internal/cli/screenshot.go: a relative flag joins the working directory and the printed path is the trimmed flag value; an absolute flag stays absolute and is printed unchanged; an existing file, a directory, a missing parent, and a blank `--output` are rejected with no destination created (data-model.md: "Relative paths resolve against the current directory. Absolute paths are used as given"; "Destination exists, is a directory, or parent is missing → Fail before capture, no file created, existing content unchanged")
- [x] T016 [P] [US3] Add `TestScreenshotOutputPath` in tests/integration/screenshot_output_test.go using `requireChrome`: with a running headless instance, `-o` a new relative path prints that relative path, writes a PNG only there, and does not create a `screenshot-*.png` default file; `-o` a new absolute path prints that absolute path; repeating `-o` onto the existing file exits 1 and leaves its bytes unchanged (SC-006; quickstart Scenario 4)

### Implementation for User Story 3

- [x] T017 [US3] Add `--output` / `-o` in internal/cli/screenshot.go. A blank flag value is a usage error (exit 2). Resolve and reject the destination before dialing the worker. On success, `os.Link` only to that path, print the trimmed flag value on stdout, and do not also create a `screenshot-*.png` default file (FR-007; contracts/cli.md)
- [x] T018 [US3] Extend `Long` help in internal/cli/screenshot.go so `screenshot --help` documents `--output` and `-o`, states that the path may be relative or absolute, and does not describe a prompt (FR-013; User Story 3 scenario 3)

**Checkpoint**: Unit output tests and `TestScreenshotOutputPath` pass. Help names the ID, the default file, and `--output` / `-o`.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Full-suite gate and quickstart confirmation

- [x] T019 [P] Walk specs/003-screenshot-cli/quickstart.md Scenarios 1 through 5 and record any mismatch against contracts/cli.md. The visible-window half of Scenario 2 is manual. The default `go test` run stays headless and must not open a casino site
- [x] T020 Run `go test ./...` from the repository root and fix failures in the screenshot paths (constitution Delivery Workflow)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** → user stories → **Polish**
- **US1** starts after T002 and T003. T007 before T008. T007 and T003 before T009. T010 after T003, and after T009 if both edit help strings in internal/cli/screenshot.go
- **US2** implementation (T013, T014) edits internal/cli/screenshot.go after T009
- **US3** implementation (T017, T018) edits internal/cli/screenshot.go after T014
- **Polish** after US1–US3

### User Story Dependencies

| Story | Depends on | Independent test |
|-------|------------|------------------|
| US1 | T002, T003 | Headless `screenshot <id>` writes a PNG and leaves the same PIDs running. A second headless instance is unchanged |
| US2 | T003, T009 | No ID, an extra ID, an unknown ID, and a non-running row fail with no PNG. A failed publish leaves no final file |
| US3 | T009, T013 | `-o` relative and absolute paths; an existing destination is not replaced; help documents the flag |

US2 tests (T011, T012) and US3 tests (T015, T016) use different files from the US1 tests and can be written in parallel with T004, T005, and T006. Implementation of internal/cli/screenshot.go stays sequential.

### Parallel Opportunities

```text
After T001:
  T002 internal/registry/paths.go
  T003 internal/cli/root.go and internal/cli/screenshot.go

After Phase 2, test files in parallel:
  T004 tests/unit/screenshot_test.go
  T005 and T006 tests/integration/screenshot_test.go
  T011 tests/integration/screenshot_refuse_test.go
  T012 tests/unit/screenshot_publish_test.go
  T015 tests/unit/screenshot_output_test.go
  T016 tests/integration/screenshot_output_test.go

T005 and T006 share tests/integration/screenshot_test.go, so write them in one pass.
Implementation of screenshot.go stays sequential: T009 → T013 → T014 → T017 → T018
Worker files stay sequential: T007 → T008
```

---

## Parallel Example: User Story 1

```text
T004 socket contract test in tests/unit/screenshot_test.go
T005 and T006 headless PNG tests in tests/integration/screenshot_test.go
```

Write those tests first so they fail. Then T007 and T008 (worker), then T009 and T010 (CLI publish and help).

---

## Implementation Strategy

### MVP First (User Story 1)

1. T001 → T002 and T003
2. T004, T005, and T006, then T007 → T008 → T009 → T010
3. Validate with quickstart Scenario 1 and the headless half of Scenario 2
4. Stop here to demo a default-path screenshot

### Incremental Delivery

1. US1: capture a running instance to the current directory, including two headless instances
2. US2: usage, not-found, and not-running failures create no file, and a failed publish leaves no partial PNG
3. US3: `--output` / `-o` relative and absolute paths without overwrite, plus help for that flag
4. Polish: `go test ./...` and the quickstart

### Suggested order

US1 (P1, MVP) → US2 (P1, same command file) → US3 (P2) → Polish

---

## Notes

- Cite spec IDs in commit messages (for example FR-001, User Story 1).
- Do not change `start`, `status`, `stop`, Chrome allocator flags, or `schema_version`.
- Not-found is exit 1 for `screenshot`, including a stopped instance whose row was removed. A tracked `starting`, `stopping`, or `stopped` row is `instance not running`. `stop` not-found stays exit 0.
- Help for the required ID and the default file is T010. Help for `--output` / `-o` is T018.
- Total tasks: **20** (Setup 1, Foundational 2, US1 7, US2 4, US3 4, Polish 2).

---

## Phase 7: Convergence

- [x] T021 CRITICAL: Cite the spec scenario or task on each new screenshot behavior in internal/cli/screenshot.go, internal/worker/capture.go, internal/worker/run.go, and internal/registry/paths.go, using the existing FR-comment style, per Constitution I (missing)
- [x] T022 Keep the capture socket absent while the registry row is still starting: start ServeCapture only after that row is updated to running, remove the socket before the registry row is removed on shutdown, and retry the CLI dial inside the existing 15-second budget so a running instance still captures per plan: worker listen order (contradicts)
