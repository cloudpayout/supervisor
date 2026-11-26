# Research: Session Instance Info (019)

**Date**: 2026-11-26

## R1 — Plain text lines, not JSON

**Decision**: Success stdout is exactly four lines `label: value` (one space after colon), fixed order, one trailing newline. No JSON, no headers, no tables.

**Rationale**: Clarification session 2026-11-26 in the spec. Matches operator habit for `logs` (line-oriented) without adopting `status` tabular layout for four fields.

**Alternatives considered**:

- JSON object — rejected per spec amendment.
- Single tab-separated row — rejected; spec names explicit labels.

## R2 — Game from casino catalog, not a new registry field

**Decision**: `game` is `casino.Lookup(row.Casino).Game` when lookup succeeds and `Game` is non-empty. Otherwise `game: null`. Do not add `game` to `registry.Instance` in this feature.

**Rationale**: Spec FR-007 allows association “when Supervisor has associated one with the instance.” For supported casinos, the catalog already binds `onewin` → `aviator` ([catalog.go](../../../internal/casino/catalog.go)). Future multi-game casinos can extend the catalog or a later spec can persist `game` on the row.

**Alternatives considered**:

- Always `null` until registry field ships — rejected; spec acceptance scenario requires `aviator` for onewin stand-in.
- Infer game from `Entry` (`entering`/`ready`) — rejected; blank casino with entry is impossible today, and entry is not a game identifier.

## R3 — Reconcile with `store.Update` + `registry.Apply`

**Decision**: Before `Find`, run `store.Update` and call `registry.Apply(doc, time.Now().UTC())` the same way as one-shot `status` and `stop` (FR-003).

**Rationale**: Spec requires reconciliation with instance-targeting commands. `logs` often uses `View` only; session info aligns with stop/status so stale rows match operator expectations.

**Alternatives considered**:

- `View` only (screenshot/watch pattern) — rejected; spec FR-003 names reconcile explicitly.
- Reconcile without persisting — rejected; `Update` is the established API that persists reap results.

## R4 — Any tracked state is readable

**Decision**: If the row exists after reconcile, print info regardless of `starting` | `running` | `stopping` | `stopped`. No `running` gate.

**Rationale**: Spec edge cases and User Story 1. Differs from `screenshot`/`watch`, which need a live worker.

**Alternatives considered**:

- Only `running` — rejected by spec edge case list.

## R5 — Timestamp format

**Decision**: `date` line uses UTC `YYYY-MM-DD HH:MM`: zero-padded date, space, zero-padded hour, colon, zero-padded minute. No seconds, no timezone suffix.

**Rationale**: Spec amendment 2026-11-26 (`{year}-{month}-{day} {hour}:{minute}`). Registration time remains `created_at` on the row, truncated to the minute in UTC. Go reference layout: `2006-01-02 15:04` on `CreatedAt.UTC()`.

**Alternatives considered**:

- RFC3339Nano (log lines) — rejected; operators asked for calendar layout `YYYY-MM-DD HH:MM` UTC.
- Local timezone — rejected; registry timestamps are UTC.

## R6 — Cobra grouping: `session` + `info`

**Decision**: Top-level `session` command with hidden-use parent (`session` with no subcommand prints help) and `info` subcommand with `Use: "info [instance-id]"` and `Args: cobra.ExactArgs(1)`.

**Rationale**: Spec FR-001 names `session info <instance-id>`. Matches `logs` + `logs clear` nesting pattern.

**Alternatives considered**:

- Flat `supervisor session-info` — rejected; spec spelling is two words / nested command.
- `session` as alias only — rejected; root help must list `session`.

## R7 — Exit codes and messages

**Decision**: Usage → exit `2` (`instance ID required`, Cobra arg count). Not found → exit `1`, error `instance not found: <id>`. Success → exit `0`; stdout only the four lines.

**Rationale**: Aligns with `screenshot` and `watch` refusal tests. `stop` not-found is exit `0` with stderr — spec says same *class* as stop/logs for unknown IDs; for a read failure, non-zero matches `screenshot`/`logs` log-missing behavior.

**Alternatives considered**:

- Exit `0` on not-found like `stop` — rejected; spec FR-004 requires non-zero and SC-004 requires non-zero for unknown ID.

## R8 — Stdout/stderr split on success and failure

**Decision**: Success: four lines on stdout only; stderr byte-empty. Failure (usage, not found, registry I/O): stdout byte-empty; errors on stderr via existing `cli.Execute` / `main` pattern (`SilenceErrors`, usage on stderr before exit 2).

**Rationale**: Clarification session 2026-11-26 option A. Integration tests can assert `len(stdout)==0` on failure paths, matching `screenshot_refuse_test` / `watch_refuse_test`.

**Alternatives considered**:

- Cobra usage on stdout for arg errors — rejected; breaks script parsers.
- Not-found on stdout — rejected; spec FR-005b.