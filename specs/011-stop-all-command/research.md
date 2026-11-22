# Research: Stop All Instances Command

**Feature**: `011-stop-all-command` | **Date**: 2026-11-22

## R1: Operator surface (keyword vs subcommand)

**Decision**: Keep `stop` as a single Cobra command with `cobra.ExactArgs(1)`. When the trimmed argument equals the reserved token `all` (case-sensitive), dispatch to bulk teardown; otherwise run existing single-ID stop (FR-008, FR-009).

**Rationale**: Matches operator wording `supervisor stop all` without restructuring the command tree. `logs clear` uses a subcommand because the parent accepts optional instance IDs; `stop` always requires exactly one token, so `all` fits the same arity as an opaque ID.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `stop` parent + `all` subcommand + `stop <id>` sibling | Forces Cobra reshaping; `stop <id>` today is the default `RunE`, not a child |
| Separate top-level `stop-all` command | Violates spec invocation and splits help |
| Case-insensitive `all` | Spec FR-009 reserves exact `all`; inconsistent with strict ID matching elsewhere |

## R2: Shared per-instance stop logic

**Decision**: Extract the body of `runStop` into an unexported helper (e.g. `stopInstance(id string) (outcome stopOutcome, err error)`) in `internal/cli/stop.go`. `runStop` branches on `id == stopAllToken` → `runStopAll`, else `stopInstance`. Bulk loop calls `stopInstance` only for IDs in the snapshot.

**Rationale**: FR-001/FR-004 require identical termination, profile removal, registry updates, and stderr lines per ID. Duplicating `store.Update` + `proc.Terminate` blocks would drift from 002/004 behavior.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Shell out to `supervisor stop <id>` N times | Breaks tests, slow, loses shared store |
| Move logic to `internal/registry` | Stop is operator lifecycle; stays in `cli` per 001 layout |

## R3: Stop target snapshot

**Decision**: First step of `runStopAll`: one `store.Update` that runs `registry.Apply`, then copies `doc.Instances` IDs into a `[]string` snapshot before releasing the lock. Iterate the snapshot in registry sort order (document order). Do not add IDs that appear in the registry only after the snapshot (concurrent `start`).

**Rationale**: Spec entity “Stop target set” and edge case on concurrent CLI. Each `stopInstance` still reconciles on its own `Update` as today.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Re-read full ID list each iteration | Could double-stop or miss spec snapshot rule |
| Stop only `running` | Spec includes `starting` (and live `stopping` if present) |

## R4: Empty registry messaging

**Decision**: When the snapshot length is zero after the initial `Apply`, print `No instances are running.` on **stderr** and return `nil` (exit `0`). Wording matches `status` one-shot empty (FR-003); stream differs because all other `stop` outcomes use stderr.

**Rationale**: Status prints empty to stdout for table UX; stop family already uses stderr for operational messages (`stopped instance`, `instance not found`, etc.).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Print to stdout | Inconsistent with single-ID stop |
| Silent success | Fails FR-003 and SC-003 |

## R5: Bulk summary and exit aggregation

**Decision**: After the loop, always print one summary line on stderr:

```text
stop all: stopped N, skipped M (already stopped or not found)
```

Increment **stopped** on successful live termination; **skipped** on not-found and already-stopped outcomes for snapshot IDs. If any `stopInstance` returns operational `err`, continue the loop, record failed IDs, set `failed` flag; final return `err` (exit `1`) if any failed (FR-005, FR-006). Success when no failures (including all skipped).

**Rationale**: FR-007; mirrors per-ID stderr lines plus aggregate for scripts.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| JSON summary | Out of scope; no spec requirement |
| Exit non-zero when any skipped | Contradicts FR-006 / already-stopped success semantics |

## R6: Interrupt during bulk stop

**Decision**: Register `signal.Notify` on `os.Interrupt` for the bulk loop only. On first SIGINT: set `interrupted`, break out without starting the next ID; allow in-flight `stopInstance` to finish; print `stop all interrupted: completed K of L` on stderr; return `errors.New(...)` so Cobra exits `1` (general error class). Registry state for completed IDs remains consistent (FR edge case).

**Rationale**: Spec edge case; align with general error for incomplete work; do not cancel `proc.Terminate` mid-flight without a separate spec.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Exit `0` on Ctrl+C (like status follow) | Bulk stop is destructive partial work; spec requires non-zero |
| Immediate `os.Exit` | Can skip registry row removal for in-progress ID |

## R7: Testing and docs

**Decision**: Add `tests/integration/stop_all_test.go` (empty registry, two instances, help mentions `all`). Grep-update quickstarts that say “stop all” informally to `supervisor stop all` (FR-011): `specs/001-supervisor-cli/quickstart.md`, `specs/002-status-stop-cli/quickstart.md`, and any integration helpers that stop multiple instances sequentially where bulk is intended.

**Rationale**: Constitution Delivery Workflow requires CLI stop coverage; SC-004.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Only unit tests | Lifecycle needs real registry + processes pattern from `stop_lifecycle_test.go` |
| Slow test with 5 instances | Optional tag; quickstart covers manual SC-001 |

## R8: Operational failure (FR-005)

**Decision**: Bulk stop treats an **operational failure** as non-nil `error` from `stopInstance` (registry I/O after live terminate, or any path single-ID `stop` exits `1`). Not-found and already-stopped return nil error; bulk increments `skipped` only. Loop continues; `RunE` returns error at end if any operational failure occurred.

**Rationale**: Aligns spec FR-005 with testable behavior without inventing new terminate failure modes in this feature.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Exit `1` on first failure | Violates FR-005 continue rule |
| Treat already-stopped as failure | Contradicts FR-004 / 002 exit `0` |