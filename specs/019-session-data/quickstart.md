# Quickstart: Session Instance Info (019)

<!-- Validated by tests/integration/session_info_test.go (seeded rows and refuse) and session_info_help_test.go. Chrome scenarios match TestSessionInfoBlankStart and TestSessionInfoOnewinStandIn. -->

Runnable checks after implementation. See [cli contract](./contracts/cli.md) and [data model](./data-model.md). Build and isolated `XDG` env match [001 quickstart](../001-supervisor-cli/quickstart.md).

## Prerequisites

- `go` 1.27.1
- Built `./supervisor` from this revision

## Scenario 1 — Blank start (P1, SC-001, SC-002)

1. `./supervisor start --headless` and capture stdout id.
2. `./supervisor session info "$id"`.
3. **Expect**: Exit `0`. Stdout exactly:
   ```text
   casino: null
   game: null
   strategy: null
   date: <YYYY-MM-DD HH:MM UTC>
   ```
4. Parse the `date` line: timestamp must be close to when `start` returned (same registration instant as stored on the row).

## Scenario 2 — Onewin stand-in (P1, FR-006–FR-007)

Use the same session file pattern as [007 quickstart](../007-enter-aviator-onewin/quickstart.md) against the local stand-in.

1. `./supervisor start <session.json>` and note id.
2. `./supervisor session info "$id"`.
3. **Expect**: Exit `0`.
   ```text
   casino: onewin
   game: aviator
   strategy: null
   date: <YYYY-MM-DD HH:MM UTC>
   ```

## Scenario 3 — Refuse bad input (P1, SC-004)

1. `./supervisor session info` and `./supervisor session info id extra`.
2. **Expect**: Exit `2`. Stdout completely empty; stderr has usage / `instance ID required`.
3. `./supervisor session info missing-id`.
4. **Expect**: Exit `1`. Stdout completely empty. Stderr contains `instance not found: missing-id`.

## Scenario 4 — Help (P2)

1. `./supervisor help` — **expect** `session` in the command list (not `worker`).
2. `./supervisor session info --help` — **expect** mention of `casino`, `game`, `strategy`, `date`, and plain-text example (not JSON).

## Scenario 5 — Lifecycle states (edge)

1. Start an instance; while still `starting`, run `session info` on its id.
2. **Expect**: Exit `0` and four lines (state is not printed).
3. After `stop`, if the row is gone, `session info` → not found.

## Automated tests

Default CI should include `tests/integration/session_info_test.go` and help assertions extending `root_help_test.go` (FR-011).