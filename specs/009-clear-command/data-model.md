# Data Model: Clear Instance Registry Command

**Feature**: `009-clear-command` | **Date**: 2026-11-22

**Registry contract (unchanged)**: [001 registry schema](../../001-supervisor-cli/contracts/registry-schema.json)

## Scope

This feature does **not** add registry fields or change `schema_version`. It defines how the existing **instance registry document** may be reset to an empty list and how **confirmation** gates that write.

## Entity: Instance registry document

Same as 001 — top-level JSON at `{supervisor_state}/instances.json`.

| Field | Type | After successful `clear` |
|-------|------|---------------------------|
| `schema_version` | integer | `1` (current `registry.SchemaVersion`) |
| `instances` | array | `[]` (empty, not omitted) |

### Validation rules (post-clear)

1. Document MUST parse as valid JSON and match schema version `1`.
2. `instances` MUST be an empty array (not `null` after normalize).
3. No instance rows remain; `status` one-shot reports `No instances are running.`

### Unchanged persistence rules

- Writes MUST use exclusive flock via `registry.Store.Update` (same as other mutating commands).
- Profile dirs, worker logs, and capture files under state/data dirs are **not** modified by `clear`.

## Entity: Clear operation (ephemeral)

Logical operation, not persisted.

| Field | Source | Description |
|-------|--------|-------------|
| `instance_count` | `len(doc.Instances)` before prompt | Shown on confirmation prompt (FR-005) |
| `confirmed` | TTY prompt or `--yes` | Must be true to invoke `Update` |
| `stdin_tty` | `isTTY(os.Stdin)` | Selects prompt vs `--yes` requirement |

### State transition (registry content only)

```text
[Any valid Document with N ≥ 0 instances]
  └── clear + confirmed
        └── Document { schema_version: 1, instances: [] }
```

Processes referenced by previous rows may still run on the host; that state is **outside** the registry model (spec warning).

## Entity: Operator CLI catalog (delta)

Adds one operator-facing command name to the root catalog (see [contracts/cli.md](./contracts/cli.md)).

| Name | Mutates registry |
|------|-------------------|
| `clear` | Yes (on confirm / `--yes`) |

`start`, `status`, `stop`, `screenshot`, `logs`, and hidden `worker` behavior unchanged except observing an empty registry after clear.

## Relationships

```text
Operator
  └── supervisor clear [--yes]
        ├── View → Instance registry document (read count)
        ├── optional TTY prompt → Confirmation decision
        └── Update → Instance registry document (empty instances)
```