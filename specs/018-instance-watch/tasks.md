---
description: "Task list for Instance Watch (018)"
---

# Tasks: Instance Watch (Local Live View)

**Input**: Design documents from `specs/018-instance-watch/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/cli.md](./contracts/cli.md), [contracts/watch-socket.md](./contracts/watch-socket.md), [quickstart.md](./quickstart.md)

**Tests**: Included. The constitution requires automated coverage of CLI behavior before it is complete, and the plan names test files under `tests/unit/` and `tests/integration/`. Default tests must not open a live casino and must not require a visible window. Write a story's tests so they fail before that story's implementation tasks.

**Organization**: By user story (US1 watch a running instance, US2 refuse watch, US3 control cadence, US4 screenshot while watching, US5 watch ends when the run ends).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no unfinished dependencies)
- **[Story]**: US1–US5

## Path Conventions

Single Go module at repository root: `internal/cli/`, `internal/worker/`, `internal/registry/`, `tests/unit/`, `tests/integration/`, `cmd/supervisor/`. Git branch `018-instance-watch`. Feature directory `specs/018-instance-watch`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Lock operator and socket contracts before editing capture code

- [x] T001 Review specs/018-instance-watch/contracts/cli.md, specs/018-instance-watch/contracts/watch-socket.md, and specs/018-instance-watch/research.md: `WATCH <fps>` on the existing capture socket; `CAPTURE` unchanged; loopback URL `http://127.0.0.1:<port>/`; fps **1–30** default **1**; one watch per instance; `instances.json` schema version stays **1**; interrupt or instance teardown exits **0** after a printed URL

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Command registration and socket dispatch shared by every story

**⚠️ CRITICAL**: No user story work begins until this phase is complete

- [x] T002 [P] Register `newWatchCmd()` from `internal/cli/root.go` and add `internal/cli/watch.go` with `Use: "watch <instance-id>"`, `Args: cobra.ExactArgs(1)`, trimmed-empty ID → `fmt.Errorf("instance ID required")` for `isUsageError` exit **2**, `--fps` flag default **1**, and validation that rejects non-integers and values outside **1–30** with `fmt.Errorf("usage: invalid --fps %q", …)` before any dial (FR-009; data-model.md cadence **1–30**)
- [x] T003 Refactor `internal/worker/capture.go` so the accept loop reads the first line and dispatches `CAPTURE` to the existing handler unchanged and `WATCH <fps>` to a new handler in `internal/worker/watch.go` (contracts/watch-socket.md; research R2). Any other first line → `ERR bad request\n`
- [x] T004 [P] Add loopback port selection in `internal/worker/watch.go` (or a small helper in the same package): try `127.0.0.1:8080`, then increment port up to **64** attempts; return the bound port (FR-014; research R3)

**Checkpoint**: `supervisor watch` with no ID exits **2**. `supervisor watch some-id --fps 0` exits **2**. `CAPTURE` behavior in unit tests still passes if present. User stories can start.

---

## Phase 3: User Story 1 — Watch a running instance in the local browser (Priority: P1) 🎯 MVP

**Goal**: `supervisor watch <instance-id>` prints `http://127.0.0.1:<port>/`, serves multipart JPEG at default **1** fps, and holds until interrupt; interrupt leaves the instance running (FR-001, FR-004, FR-005, FR-006, FR-011).

**Independent Test**: Start one headless instance, run `watch`, open the printed URL within 10 seconds, confirm multipart JPEG, interrupt with exit **0**, `status` still shows the same PIDs running.

### Tests for User Story 1

- [x] T005 [P] [US1] Add Chrome-free tests in `tests/unit/watch_socket_test.go`: fake worker handler accepts `WATCH 1\n` → `OK http://127.0.0.1:<port>/\n` then blocks; `GET /` returns `Content-Type: multipart/x-mixed-replace; boundary=frame` and at least one `image/jpeg` part; `CAPTURE\n` still returns PNG contract per `specs/003-screenshot-cli/contracts/capture-socket.md` on the same listener (watch-socket.md)
- [x] T006 [US1] Add `TestWatchHeadlessPrintsURL` in `tests/integration/watch_test.go` using `requireChrome`, `isolatedEnv`, and `supervisorBin`: `start --headless`, wait until `running`, record `owner_pid` and `browser_pid`, run `watch <id>` in a goroutine, read stdout URL within **10** seconds, `curl --no-buffer` the URL and assert multipart JPEG, send interrupt to watch, assert watch exits **0**, row still `running` with same PIDs (SC-001, SC-008; User Story 1)

### Implementation for User Story 1

- [x] T007 [US1] Implement watch serving in `internal/worker/watch.go`: on `WATCH <fps>`, if already watching return `ERR already watching\n`; bind loopback HTTP; `page.StartScreencast` JPEG on `taskCtx`; ack each `screencastFrame` in a **separate goroutine** (research R4, chromedp #920); emit at most `<fps>` JPEG parts per second via `multipart/x-mixed-replace; boundary=frame`; write `OK http://127.0.0.1:<port>/\n` then hold the Unix connection until client hang-up; on hang-up call `StopScreencast`, close HTTP listener, clear watching flag; do **not** cancel `taskCtx` (FR-005, FR-006)
- [x] T008 [US1] Wire `internal/worker/run.go` so `ServeCapture` passes the screencast closure into the dispatch from T003; on `taskCtx` done, ensure watch listener and HTTP server stop before `remove(reg, id)` (plan.md worker shutdown)
- [x] T009 [US1] In `internal/cli/watch.go`, `Store.View` only: not found / not running before dial (same rules as screenshot). Dial `Layout.CaptureSocket`, send `WATCH <fps>\n`, read `OK` line, print URL trimmed on stdout, block until connection closes or interrupt (reuse signal pattern from `internal/cli/logs.go` if applicable), exit **0** on interrupt while instance still running (research R6; FR-004, FR-011)
- [x] T010 [US1] Set `Short` and `Long` help on `internal/cli/watch.go` for required ID, `--fps` default **1** max **30**, loopback URL, interrupt ends watch (FR-015 partial; User Story 1 scenario 5)

**Checkpoint**: `go test ./tests/integration/ -run TestWatchHeadlessPrintsURL` passes when Chrome is installed. Socket unit test passes without Chrome. MVP is default-fps watch on headless blank page.

---

## Phase 4: User Story 2 — Refuse watch that cannot succeed (Priority: P1)

**Goal**: Missing/extra/blank ID, unknown ID, non-running row, bad worker response, and second concurrent watch fail clearly without a success URL (FR-002, FR-003, FR-007, FR-008, FR-013).

**Independent Test**: `watch` with no ID, `missing-id`, `starting` row, and a second `watch` on an active session; each fails per contracts/cli.md.

### Tests for User Story 2

- [x] T011 [P] [US2] Add tests in `tests/integration/watch_refuse_test.go` without `requireChrome`: no args, blank ID, extra arg → exit **2** no URL; `missing-id` → exit **1** `instance not found`; seeded non-`running` states → exit **1** `instance not running`; while one watch runs, second `watch` same id → exit **1** `already watching`, first watch continues (SC-004, SC-005; User Story 2)

### Implementation for User Story 2

- [x] T012 [US2] Ensure `internal/cli/watch.go` lookup and messages match `specs/018-instance-watch/contracts/cli.md` before dial: `watch failed: <id>: <reason>` exit **1** for worker `ERR`, short read, or dial failure; never print URL unless `OK` line parsed (FR-008; research R6)
- [x] T013 [US2] Enforce single watch in `internal/worker/watch.go`: global per-worker `watching` flag; second `WATCH` returns `ERR already watching\n` without starting HTTP (FR-013; data-model.md exclusivity)

**Checkpoint**: Refuse tests pass without Chrome. US1 watch still works.

---

## Phase 5: User Story 5 — Watch ends when the run ends (Priority: P1)

**Goal**: When the instance stops or worker shuts down, `watch` exits within **5** seconds and the HTTP stream stops (FR-010).

**Independent Test**: Start watch, `stop` the instance from another terminal; watch exits **0** within 5s and URL stops serving.

### Tests for User Story 5

- [x] T014 [US5] Add `TestWatchEndsWhenInstanceStops` in `tests/integration/watch_test.go` with `requireChrome`: start watch, `stop <id>` from test flow, assert watch process exits within **5** seconds with code **0**, subsequent curl to URL fails (SC-003; quickstart Scenario 4)

### Implementation for User Story 5

- [x] T015 [US5] In `internal/worker/watch.go` and `internal/worker/run.go`, close the Unix watch connection from the worker when `taskCtx` ends (stop, crash, entry teardown) so the CLI unblocks without requiring operator interrupt; tear down HTTP and screencast in that path (FR-010; User Story 5 scenarios 1–2)

**Checkpoint**: `TestWatchEndsWhenInstanceStops` passes when Chrome is installed.

---

## Phase 6: User Story 3 — Control update cadence (Priority: P2)

**Goal**: `--fps` **1–30** is passed to worker and throttles HTTP parts; static headless page may emit fewer parts without failure (FR-009; SC-006).

**Independent Test**: Default and `--fps 10` on a changing stand-in page (integration or `slow` tag); invalid fps usage errors.

### Tests for User Story 3

- [x] T016 [P] [US3] Add tests in `tests/unit/watch_fps_test.go` for CLI parsing: default **1**, **30** accepted, **0** and **31** and non-numeric rejected before dial (FR-009)
- [x] T017 [P] [US3] Add `TestWatchCadence` in `tests/integration/watch_test.go` with `//go:build slow` or stand-in page that animates: over **60** seconds average **0.8–1.2** parts/s at default and **8–12** at `--fps 10` (SC-006). Skip in default `go test ./...` if tagged `slow` (quickstart Scenario 2)

### Implementation for User Story 3

- [x] T018 [US3] Pass validated `--fps` from `internal/cli/watch.go` in the `WATCH <fps>\n` line; extend `Long` help with max **30** (FR-009, FR-015)
- [x] T019 [US3] In `internal/worker/watch.go`, throttle outbound multipart parts to `<fps>` per wall-clock second (keep-latest JPEG, drop older); do not treat sparse CDP events as failure on static pages (research R4; User Story 3 scenario 4)

**Checkpoint**: Unit fps tests pass. Slow cadence test passes when `-tags slow` and Chrome present.

---

## Phase 7: User Story 4 — Still capture while watching (Priority: P2)

**Goal**: `screenshot` succeeds while watch is active; watch continues; screenshot mutex does not block screencast loop indefinitely (FR-012; SC-007).

**Independent Test**: watch + screenshot same id; both succeed; watch still serves until interrupt.

### Tests for User Story 4

- [x] T020 [US4] Add `TestScreenshotWhileWatching` in `tests/integration/watch_screenshot_test.go` with `requireChrome`: active watch, run `screenshot <id>`, assert PNG exit **0**, watch still serving, interrupt watch exit **0**, instance still running until `stop` (SC-007; quickstart Scenario 5)

### Implementation for User Story 4

- [x] T021 [US4] In `internal/worker/watch.go` and `internal/worker/capture.go`, keep `CAPTURE` on the existing mutex and **10** second child context; ensure screencast handling does **not** hold that mutex across frame delivery (research R5; FR-012)

**Checkpoint**: `TestScreenshotWhileWatching` passes when Chrome is installed. Spec 003 screenshot tests still pass.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Root help, quickstart gate, full suite

- [x] T022 [P] Add `watch` to `internal/cli/root.go` root `Short`/`Long` sibling list and extend `tests/integration/root_help_test.go` want list with `watch` (FR-016)
- [x] T023 [P] Walk `specs/018-instance-watch/quickstart.md` Scenarios 1–6 against implemented behavior; headed watch stays manual
- [x] T024 Run `go test ./...` from repository root and fix failures in watch paths (constitution Delivery Workflow)
- [x] T025 CRITICAL: Cite spec FR/task IDs in comments on new behavior in `internal/cli/watch.go`, `internal/worker/watch.go`, and touched lines in `internal/worker/capture.go` / `internal/worker/run.go`, matching existing FR-comment style (Constitution I)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** → user stories → **Polish**
- **US1** after T002–T004. T007 before T008. T008 before T009.
- **US2** after T009 (same `watch.go`)
- **US5** after US1 worker paths (T007, T008)
- **US3** after US1 (fps on working stream)
- **US4** after US1 and stable `CAPTURE` mutex (T007, T021)

### User Story Dependencies

| Story | Depends on | Independent test |
|-------|------------|------------------|
| US1 | Phase 2 | Headless watch prints URL, JPEG stream, interrupt leaves PIDs |
| US2 | US1 CLI dial | Usage, not found, not running, second watch refused |
| US5 | US1 worker | `stop` ends watch within 5s |
| US3 | US1 | `--fps` validation and throttled parts |
| US4 | US1 | Screenshot during watch |

US2/US5/US3/US4 tests in different files can be written in parallel after Phase 2 where marked `[P]`.

### Parallel Opportunities

```text
After T001:
  T002 internal/cli/watch.go + root.go
  T004 port helper (same package as T007; start T004 before T007)

After Phase 2, tests in parallel:
  T005 tests/unit/watch_socket_test.go
  T011 tests/integration/watch_refuse_test.go
  T016 tests/unit/watch_fps_test.go

Implementation sequence:
  T003 → T007 → T008 → T009 → T010 (US1)
  T012, T013 (US2)
  T015 (US5)
  T018, T019 (US3)
  T021 (US4)
```

---

## Parallel Example: User Story 1

```text
T005 socket + MJPEG contract in tests/unit/watch_socket_test.go
T006 TestWatchHeadlessPrintsURL in tests/integration/watch_test.go
```

Write tests first so they fail. Then T007 → T008 → T009 → T010.

---

## Implementation Strategy

### MVP First (User Story 1)

1. T001 → T002, T003, T004
2. T005, T006, then T007 → T008 → T009 → T010
3. Validate quickstart Scenario 1
4. Stop here to demo local watch

### Incremental Delivery

1. US1: loopback watch at default fps (MVP)
2. US2: refuse paths and exclusivity
3. US5: watch follows instance teardown
4. US3: `--fps` up to 30
5. US4: screenshot while watching
6. Polish: root help, `go test ./...`, FR comments

### Suggested order

US1 (P1 MVP) → US2 (P1) → US5 (P1) → US3 (P2) → US4 (P2) → Polish

---

## Notes

- Do not change `start`, `status`, `stop`, `screenshot` contract except shared worker mutex behavior required for FR-012.
- Do not bump `schema_version`. Do not bind `0.0.0.0`.
- Total tasks: **25** (Setup 1, Foundational 3, US1 6, US2 3, US5 2, US3 4, US4 2, Polish 4).

## Phase 9: Convergence

- [x] T026 Add `TestWatchCadence` in `tests/integration/watch_cadence_test.go` with `//go:build slow`: changing stand-in page, assert **0.8–1.2** parts/s over **60** s at default fps and **8–12** at `--fps 10` per SC-006 and quickstart Scenario 2 (missing)
- [x] T027 Add `tests/integration/watch_help_test.go` asserting `watch --help` documents instance ID, `--fps` default **1** and max **30**, loopback URL, and interrupt behavior per FR-015 and quickstart Scenario 6 (partial)
- [x] T028 Treat non-integer `--fps` (e.g. `abc`) as a usage error with exit **2** and `invalid --fps` messaging per FR-009 and `contracts/cli.md`; add integration coverage alongside unit validation (partial)