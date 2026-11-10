# Quickstart: Validate Status and Stop CLI Changes (002)

Runnable checks after implementation. See [cli contract](./contracts/cli.md) and [data model](./data-model.md). Base flows (build, XDG env, Chrome) match [001 quickstart](../001-supervisor-cli/quickstart.md).

## Prerequisites

- Built binary: `go build -o supervisor ./cmd/supervisor`
- Isolated state (recommended): `XDG_STATE_HOME` and `XDG_DATA_HOME` as in 001

## Scenario 1 — One-shot status (P1, SC-001)

1. Start one instance: `./supervisor start --headless`; note `id`.
2. Run `./supervisor status`.
3. **Expect**: Table with `id` printed **once**; shell prompt returns immediately (no Ctrl+C).
4. **Expect**: Terminal screen **not** cleared (scrollback/history still visible on TTY).
5. Run again with no instances (stop all); **expect** `No instances are running.` and immediate exit.

**Spec refs**: FR-001, FR-003; User Story 1.

## Scenario 2 — Follow mode (P2, SC-002)

1. With at least one instance running, run `./supervisor status -f` (or `--follow`).
2. **Expect**: Table updates ~every 2s; on TTY, screen clears between refreshes.
3. Press Ctrl+C.
4. **Expect**: Exit `0`.

**Spec refs**: FR-002; User Story 2.

## Scenario 3 — Stop with ID (P1, SC-003, SC-004)

1. Start instance; copy `id`.
2. Run `./supervisor stop "$id"`.
3. **Expect**: `stopped instance <id>` on stderr; exit `0`.
4. Run `./supervisor status`.
5. **Expect**: Instance absent.
6. Run `./supervisor stop` (no args).
7. **Expect**: Usage error; exit `2`; registry unchanged.
8. Run `./supervisor stop missing-id`.
9. **Expect**: `instance not found: missing-id`; exit `0`.

**Spec refs**: FR-004–FR-007; User Story 3.

## Scenario 4 — Help text (SC-005)

1. `./supervisor status --help` — documents default snapshot and `--follow` / `-f`.
2. `./supervisor stop --help` — documents `<instance-id>` argument; **no** stdin prompt.

**Spec refs**: FR-008.

## Automated regression

```bash
go test ./...
go test -tags=slow ./tests/integration/... -run 'StatusRefresh|StopLifecycle|StartRegisters'
```

**Expect**: Tests use `status -f` where loop behavior is asserted; `stop <id>` everywhere; no test relies on empty stdin stop abort.

**Spec refs**: FR-009.
