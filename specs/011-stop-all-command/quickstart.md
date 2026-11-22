# Quickstart: Validate Stop All (011)

Runnable checks after implementation. See [cli contract](./contracts/cli.md) and [data model](./data-model.md). Build and `XDG_*` setup match [001 quickstart](../001-supervisor-cli/quickstart.md).

## Prerequisites

- Built binary: `go build -o supervisor ./cmd/supervisor`
- Isolated state: set `XDG_STATE_HOME` and `XDG_DATA_HOME` per 001

## Scenario 1 — Bulk stop two instances (P1, SC-001)

1. `./supervisor start --headless` → note `id1`.
2. `./supervisor start --headless` → note `id2`.
3. `./supervisor status` → both IDs listed.
4. Run `./supervisor stop all`.
5. **Expect**: stderr contains `stopped instance <id1>` and `stopped instance <id2>` (order may match snapshot).
6. **Expect**: summary line `stop all: stopped 2, skipped 0 (already stopped or not found)`.
7. **Expect**: exit `0`.
8. `./supervisor status` → `No instances are running.` exit `0`.

**Spec refs**: FR-001, FR-004, FR-007; User Story 1.

## Scenario 2 — Empty registry (P2, SC-002, SC-003)

1. Ensure no instances (`./supervisor stop all` or a fresh state dir).
2. Run `./supervisor stop all`.
3. **Expect**: stderr `No instances are running.`; exit `0`; completes within a few seconds.

**Spec refs**: FR-003; User Story 2.

## Scenario 3 — Single-ID stop unchanged (P2, SC-005)

1. Start one instance; copy `id`.
2. `./supervisor stop "$id"` → same messages and exit as 002 quickstart Scenario 3.
3. `./supervisor stop --help` → documents `all` and `<instance-id>`; no stdin prompt.

**Spec refs**: FR-008, FR-010; User Story 3.

## Scenario 4 — Usage errors

1. `./supervisor stop` (no args) → usage exit `2`.
2. `./supervisor stop id1 id2` (if shell allows) → usage exit `2`.

**Spec refs**: Edge cases.

## Scenario 5 — Five instances (SC-001, manual or slow test)

1. Start five `--headless` instances.
2. Run `./supervisor stop all` once (no stdin).
3. **Expect**: All five absent from one-shot `status` within 60s; exit `0`.

Or: `go test ./tests/integration/... -tags=slow -run TestStopAllFiveInstances -v`.

## Scenario 6 — Interrupt (supplements TestStopAllSIGINT)

1. Start two instances; run `./supervisor status -f` in another terminal optional.
2. Run `./supervisor stop all`; send Ctrl+C mid-run.
3. **Expect**: Non-zero exit; interrupted summary on stderr; registry matches fully stopped IDs only.

## Automated regression

```bash
go test ./...
go test ./tests/integration/... -run 'StopAll'
```

**Expect**: New `stop_all` integration cases pass (including empty 10× loop and SIGINT); existing `StopLifecycle` unchanged.

## Teardown note (replaces informal “stop all”)

Use **`./supervisor stop all`** instead of stopping each ID manually after multi-instance tests or quickstart flows (FR-011). See updated [001](../001-supervisor-cli/quickstart.md) and [002](../002-status-stop-cli/quickstart.md) quickstarts.