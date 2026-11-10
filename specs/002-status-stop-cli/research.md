# Research: Status and Stop CLI Behavior (002)

**Date**: 2026-11-09

## R1 — Default status: one-shot vs follow flag

**Decision**: Add Cobra boolean flag `--follow` with shorthand `-f`, default `false`. When false, call `renderStatus` once and return. When true, run the existing ticker + signal loop.

**Rationale**: Matches FR-001/FR-002; minimal API surface (no `--once` redundant with default).

**Alternatives considered**:

- `status --once` with follow as default — rejected; spec explicitly inverts today’s default.
- Subcommand `status snapshot` — rejected; breaks muscle memory and scripts.

## R2 — TTY screen clear split

**Decision**: Parameterize `renderStatus(store, clearScreen bool)`. `clearScreen` is `false` for one-shot. For follow, `clearScreen = isTTY(os.Stdout)` before each render (preserves non-TTY piped follow from 001).

**Rationale**: Encodes clarification session 2026-11-09 without duplicating table rendering.

**Alternatives considered**:

- Separate `renderStatusSnapshot` duplicate — rejected (DRY, FR-003 same columns/logic).

## R3 — Stop positional argument validation

**Decision**: `cobra.ExactArgs(1)` on `stop`; ID from `args[0]` after `strings.TrimSpace`. Whitespace-only ID returns error from `RunE` (usage class).

**Rationale**: FR-004/FR-005/FR-006; Cobra handles arity; matches `start` stdout ID passthrough.

**Alternatives considered**:

- `stop --id` flag — deferred in 001 “future compatibility”; spec requires positional parameter.
- Optional ID with prompt fallback — rejected; FR-005 forbids stdin.

## R4 — Missing stop ID exit code

**Decision**: Rely on Cobra for missing/extra args → exit `2` (same as `start` headed+headless). Document removal of exit `0` “stop aborted”.

**Rationale**: Assumption in spec aligns with [001 cli contract](../001-supervisor-cli/contracts/cli.md) usage-error class.

**Alternatives considered**: Custom exit `1` — rejected; inconsistent with existing CLI.

## R5 — Integration test migration

**Decision**: `stopID` helper becomes `exec.Command(bin, "stop", id)`; `status_refresh_test` uses `status`, `-f`; add test that bare `status` exits without hang (timeout in test).

**Rationale**: FR-009; prevents regressions on SC-001 and SC-003.

**Alternatives considered**: Keep helper on stdin until phased removal — rejected; spec removes prompt entirely.

## R6 — Registry / reconciliation on one-shot status

**Decision**: One-shot path still calls `store.Update` + `registry.Apply` inside `renderStatus` (unchanged).

**Rationale**: FR-003 requires same reconciliation logic as follow cycles; one apply per invocation is sufficient for snapshot.

**Alternatives considered**: Skip reconcile on one-shot — rejected; stale rows would diverge from follow view.
