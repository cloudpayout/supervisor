---
description: "Task list for Supervisor CLI (start, status, stop)"
---

# Tasks: Supervisor CLI (start, status, stop)

**Input**: Design documents from `specs/001-supervisor-cli/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md  
**Spec revision**: 2026-11-09 (plan refresh: headed/headless, reconciliation, stop exits, `headless` registry field)

**Tests**: Included per constitution delivery workflow (CLI start/status/stop lifecycle must have automated coverage before behavior is complete).

**Organization**: Tasks grouped by user story for independent implementation and validation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Single Go module at repository root per plan.md: `cmd/supervisor/`, `internal/`, `tests/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Module layout and dependencies

- [X] T001 Create directory tree per plan in `cmd/supervisor/`, `internal/cli/`, `internal/registry/`, `internal/worker/`, `internal/browser/`, `internal/metrics/`, `internal/instance/`, `tests/unit/`, `tests/integration/`
- [X] T002 Add module dependencies in `go.mod`: `github.com/spf13/cobra`, `github.com/chromedp/chromedp`, `github.com/shirou/gopsutil/v4` (Go 1.27.1)
- [X] T003 [P] Create Cobra root command and `main` entry in `cmd/supervisor/main.go` and `internal/cli/root.go` (binary name `supervisor`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Registry, paths, IDs, browser helpers, hidden worker command shell — **blocks all user stories**

**⚠️ CRITICAL**: No user story work until this phase is complete

- [X] T004 Implement cryptographically random opaque instance ID generation in `internal/instance/id.go` (unique per FR-002; collision re-roll under lock per data-model)
- [X] T005 [P] Implement XDG path resolution (`XDG_STATE_HOME`, `XDG_DATA_HOME`, `HOME` fallback) in `internal/registry/paths.go` for registry, logs, and `profiles/<id>/`
- [X] T006 Implement `Instance` and registry document types in `internal/registry/types.go` matching `contracts/registry-schema.json` (`schema_version` const `1`; `state` enum `starting`|`running`|`stopping`|`stopped`; required fields `id`, `created_at`, `profile_dir`; `owner_pid`/`browser_pid` required when `running`; optional `headless` boolean when launched with `--headless`)
- [X] T007 Implement JSON registry load/save with `flock` exclusive/shared helpers in `internal/registry/registry.go` (short RMW for parallel `start` per FR-012)
- [X] T008 [P] Implement chromedp allocator helpers (isolated `UserDataDir`, `about:blank`, headed vs headless allocator option) and shared **Chrome preflight** (`CHROME_PATH`/discovery) used by `start` before registry write in `internal/browser/chromedp.go`
- [X] T009 Register hidden `worker run --id` command stub in `internal/cli/worker.go` and wire in `internal/cli/root.go` (`Hidden: true`)
- [X] T010 [P] Add unit tests for registry flock and JSON round-trip (including optional `headless` field) in `tests/unit/registry_test.go`

**Checkpoint**: Foundation ready — user story implementation can begin

---

## Phase 3: User Story 1 - Start a supervised browser instance (Priority: P1) 🎯 MVP

**Goal**: `supervisor start` registers an instance, spawns detached worker, prints ID, exits; worker holds chromedp on blank page with isolated profile (FR-001–FR-001d, FR-002, FR-003, FR-011, FR-012).

**Independent Test**: Run `supervisor start --headless` or default headed; receive ID; shell returns; instance appears as `starting` then `running` in registry; quickstart Scenario A.

### Tests for User Story 1

- [X] T011 [P] [US1] Add integration test for `start` registration and detached worker spawn in `tests/integration/start_worker_test.go` (temp `XDG_*` dirs; `--headless`; assert ID stdout and registry `starting`→`running`)
- [X] T034 [P] [US1] Add integration test that `start --headed --headless` exits with Cobra usage error (exit `2`) and leaves registry unchanged in `tests/integration/start_flags_test.go` (FR-001d)
- [X] T039 [P] [US1] Add integration test that `start` exits non-zero with a clear error, prints **no** ID on stdout, and leaves `instances.json` unchanged when preflight fails (e.g. invalid `CHROME_PATH` or missing binary) in `tests/integration/start_launch_failure_test.go` (FR-011, User Story 1 scenario 3; synchronous CLI failure per plan)

### Implementation for User Story 1

- [X] T012 [US1] Implement `start` in `internal/cli/start.go` with mutually exclusive `--headed` and `--headless` booleans per FR-001d (default **headed** when both omitted; both set → exit `2` without registry write; document in Cobra help)
- [X] T013 [US1] Implement start flow in `internal/cli/start.go` per plan **FR-011** ordering: preflight Chrome/dirs with **no** registry write on failure; then flock, allocate `id`, create `profile_dir`, append `starting` with `created_at` RFC3339 and `headless`; detached re-exec `worker run --id` with `Setsid` and log redirection; on worker `Start` failure, re-lock and **remove** the `starting` row (and empty profile dir), exit non-zero **without** printing ID; on successful `Start`, print ID and exit `0`
- [X] T014 [US1] Implement worker chromedp lifecycle in `internal/worker/run.go`: load record by `--id`, launch Chrome with profile and `headless` flag, navigate `about:blank`, flock update `owner_pid` and `browser_pid`, set `state` to `running`; on Chrome launch failure before `running`, flock-remove the `starting` row and exit non-zero (worker log per FR-008)
- [X] T015 [US1] Implement worker signal handling and registry cleanup on exit in `internal/worker/run.go` (SIGTERM tears down browser; remove or reconcile registry per FR-009)
- [X] T016 [US1] Implement reconciliation in `internal/registry/reconcile.go`: 60s `starting` timeout from `created_at` (FR-013); split-brain teardown per FR-009a (owner dead, Chrome alive → remove row, kill Chrome tree for profile) and FR-009b (Chrome dead, owner alive → remove row, kill owner)
- [X] T017 [US1] Print operator-readable start success (ID **only after** successful worker `Start`) and failure messages in `internal/cli/start.go` (FR-008, FR-011); synchronous failure → no ID on stdout and no durable registry row

**Checkpoint**: User Story 1 independently testable via `start` + registry inspection

---

## Phase 4: User Story 2 - Observe running instances and resource usage (Priority: P2)

**Goal**: `supervisor status` lists `starting` and `running` instances, refreshes ~2s, shows per-instance CPU/RAM or `unavailable`; reconciles orphans; reflects concurrent registry changes (FR-004–FR-005a).

**Independent Test**: With instances from US1, run `status`; see states, metrics, live refresh; empty message when none (SC-005); quickstart Scenarios B–C.

### Tests for User Story 2

- [X] T018 [P] [US2] Add unit tests for process-tree CPU/RSS aggregation with fake PIDs or fixtures in `tests/unit/metrics_test.go`
- [X] T019 [P] [US2] Add integration test for live `status` output fields and `starting`/`unavailable` metrics in `tests/integration/status_refresh_test.go`
- [X] T035 [P] [US2] Add integration test that `status` reflects concurrent `start`/`stop` within one refresh without duplicate/stale IDs in `tests/integration/status_concurrent_test.go` (FR-005a)
- [X] T038 [P] [US2] Add integration test that ten `--headless` instances each appear in `status` with CPU/RAM columns and at least one table refresh within 2s per row in `tests/integration/status_ten_instances_test.go` (SC-002; skip or shorten in CI if resource-constrained via build tag `slow`)

### Implementation for User Story 2

- [X] T020 [P] [US2] Implement owner + Chrome tree PID collection and CPU/RSS sum in `internal/metrics/gopsutil.go` (exclude other instances per FR-004a)
- [X] T021 [US2] Implement `status` command with 2s ticker table render in `internal/cli/status.go` (TTY clear vs plain output per research.md; locked registry read each refresh per FR-005a)
- [X] T022 [US2] List `starting` instances with CPU/RAM `unavailable` until `running` in `internal/cli/status.go` (FR-004b)
- [X] T023 [US2] Call `internal/registry/reconcile.go` on each refresh in `internal/cli/status.go` (dead PIDs, 60s timeout, FR-009a/b; visible within one refresh cycle)
- [X] T024 [US2] Implement empty-state plain-language message in `internal/cli/status.go` (SC-005, FR-008)

**Checkpoint**: User Stories 1 and 2 work independently

---

## Phase 5: User Story 3 - Stop a supervised instance (Priority: P3)

**Goal**: Interactive `stop` prompts for ID, terminates worker/Chrome tree, updates registry; distinct messages for not-found vs already-stopped; all non-abort outcomes exit `0` (FR-006–FR-008, FR-006a, FR-007a, FR-007b).

**Independent Test**: Start instance, `stop` with valid ID; instance gone from `status`; unknown ID exits `0` with not-found message; stale registry row exits `0` already-stopped; empty prompt exits `0`; quickstart Scenario D.

### Tests for User Story 3

- [X] T025 [P] [US3] Add integration test for stop success, not-found (exit `0`), already-stopped (exit `0`), empty-input abort (exit `0`), and that `stop` triggers global registry reconcile (stale `starting` >60s pruned) in `tests/integration/stop_lifecycle_test.go`

### Implementation for User Story 3

- [X] T026 [US3] Implement interactive ID prompt in `internal/cli/stop.go`; empty/cancel abort exits `0` with no termination (FR-006a, User Story 3 scenario 4)
- [X] T027 [US3] Implement stop by ID in `internal/cli/stop.go`: after prompt with non-empty ID, run `internal/registry/reconcile.go` under flock before lookup (FR-013, FR-009a/b; same entry points as plan); then flock lookup, SIGTERM process group for `owner_pid`/Chrome subtree, fallback SIGKILL timeout, registry removal (FR-006, FR-007, FR-009)
- [X] T028 [US3] Implement not-found (no registry row, exit `0`, distinct message) vs already-stopped (row exists, processes gone, exit `0`, prune row, distinct message) in `internal/cli/stop.go` (FR-007a, FR-007b)

**Checkpoint**: Full start → status → stop lifecycle complete

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Contract alignment, concurrency, docs, constitution test gate

- [X] T029 [P] Verify `specs/001-supervisor-cli/contracts/cli.md` and `contracts/registry-schema.json` match implemented behavior (`--headed`/`--headless`, exit codes, `headless` field); adjust docs only if drift found
- [X] T030 [P] Add parallel `start` registry safety integration test in `tests/integration/registry_concurrent_test.go` (FR-012, SC-007)
- [X] T031 Add end-to-end integration test covering start → status → stop in `tests/integration/start_status_stop_test.go` per constitution delivery workflow
- [X] T036 [P] Add integration test for FR-009a orphan teardown (dead owner, live Chrome) in `tests/integration/reconcile_orphans_test.go`
- [X] T037 [P] Add integration test for FR-009b (dead Chrome, live owner) in `tests/integration/reconcile_owner_orphan_test.go`
- [X] T032 [P] Confirm `specs/001-supervisor-cli/quickstart.md` Scenarios A–H and B2 (FR-011) match built CLI (flags, stop exits, timeout G, split-brain H)
- [ ] T033 Run quickstart Scenarios A–H manually or via scripted smoke against built `cmd/supervisor` binary on Linux/WSL2 (FR-014); include optional manual SC-002 check (10 running instances, `status` refresh ≥1/2s) when `status_ten_instances_test.go` is skipped in CI

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** → **Phase 2** → **Phase 3 (US1)** → **Phase 4 (US2)** → **Phase 5 (US3)** → **Phase 6**
- US2/US3 depend on foundational registry/worker types; US2 benefits from US1 for realistic integration tests but metrics unit tests can proceed after Phase 2

### User Story Dependencies

- **US1 (P1)**: After Phase 2 — MVP
- **US2 (P2)**: After Phase 2; full integration needs US1 worker PIDs
- **US3 (P3)**: After Phase 2; full integration needs US1 running instances

### Parallel Opportunities

- Phase 1: T003 parallel with T002 after T001
- Phase 2: T005, T008, T010 parallel with T004/T006 after T001 layout
- US1: T011, T034, T039 parallel after T012 stub
- US2: T018, T020, T035, T038 parallel; T019 after T021–T023
- US3: T025 after T028 stub
- Phase 6: T029, T030, T032, T036, T037 parallel

### Parallel Example: User Story 1

```bash
# After T013–T014 land, in parallel:
# T011 start_worker_test.go
# T034 start_flags_test.go
# T016 reconcile.go (shared with US2)
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Complete Phase 1–2  
2. Complete Phase 3 (US1)  
3. **STOP and VALIDATE**: `go test ./tests/integration/... -run Start` and quickstart Scenario A  

### Incremental Delivery

1. Foundation → US1 (start/worker) → demo ID + background browser  
2. US2 (`status`) → demo monitoring  
3. US3 (`stop`) → full lifecycle  
4. Phase 6 polish and concurrent registry hardening  

### Suggested MVP Scope

**User Story 1 only** — `start` + hidden `worker` + registry + `reconcile.go` (FR-009a/b and 60s timeout in T016).

---

## Notes

- Platform v1: Linux and WSL2 only (FR-014); use `Setsid` detach on Linux  
- No v1 instance cap (FR-011)  
- Default operator mode: headed; CI prefers `--headless` (FR-010)  
- Task traceability: cite spec IDs in commit messages when completing tasks
