# CLI Contract: supervisor

**Version**: v1 (001-supervisor-cli)  
**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

## Commands (operator-facing)

### `supervisor start`

| Property | Value |
|----------|-------|
| Args | None |
| Flags | `--headed`, `--headless` (mutually exclusive; default **headed** when both omitted) |
| Success stdout | Instance `id` (single line or labeled line, stable for scripts) |
| Success exit | `0` |
| Usage error | Both `--headed` and `--headless` set → exit `2`; no registry entry (FR-001d) |
| Failure exit | Non-zero; **no stdout ID**; **no registry row** for synchronous failures (FR-011) |

Behavior: Validate flags; **preflight** Chrome/dirs (no registry write on failure); under flock append `starting`; detached `worker run` with display mode; if worker `Start` fails, **rollback** registry row and exit non-zero without printing ID; on successful `Start`, print ID and exit `0` without blocking (FR-001a). Async Chrome failure in worker removes the `starting` row in the worker process.

### `supervisor status`

| Property | Value |
|----------|-------|
| Args | None |
| Flags | None in v1 |
| Output | Table: `ID`, `STATE`, `CPU%`, `MEMORY`, optional `AGE` |
| States shown | `starting`, `running` (and reconciled-away rows per FR-004) |
| `starting` metrics | CPU and MEMORY columns show `unavailable` (FR-004b) |
| Refresh | ~2s until SIGINT (Ctrl+C); each cycle reads registry under lock (FR-005a) |
| Empty | Plain-language message: no instances (SC-005) |
| Exit | `0` after interrupt |

Metrics: For `running` rows, per instance owner + Chrome tree (FR-004a). Missing PIDs → `unavailable` and reconciliation.

### `supervisor stop`

| Property | Value |
|----------|-------|
| Args | None in v1 |
| Flags | None in v1 |
| Interaction | Prompt: instance ID (stdin) |
| Empty/cancel input | No termination; exit `0` (FR-006a) |
| Reconciliation | On non-empty ID, run registry reconcile (60s `starting`, FR-009a/b) under lock before lookup |
| Success (live stop) | Exit `0`; instance removed from registry (FR-007) |
| Unknown ID | Exit `0`; not-found message — distinct from already-stopped (FR-007b) |
| Already stopped | Exit `0`; already-stopped message; prune stale registry row (FR-007a) |

## Hidden command (implementer)

### `supervisor worker run`

Not listed in operator help (`Hidden: true` on Cobra command).

| Flag | Required | Description |
|------|----------|-------------|
| `--id` | yes | Instance ID from registry |

Optional internal flags (implementation): `--headed` / `--headless` as set by parent `start` (not operator-documented).

Behavior: Own chromedp session for `id`; update registry to `running` with PIDs; navigate `about:blank`; handle SIGTERM for graceful shutdown; remove or reconcile registry on exit; enforce profile-scoped Chrome teardown.

Invocation (internal only):

```text
supervisor worker run --id <instance_id> [--headed|--headless]
```

Detached spawn from `start` with new session / no controlling terminal.

## Exit codes (v1)

| Code | Meaning |
|------|---------|
| 0 | Success; stop abort; stop not-found; stop already-stopped |
| 1 | General error (start failure, registry/IO failure, stop failed on live instance) |
| 2 | Usage error (Cobra), including both `--headed` and `--headless` on `start` |

## Environment

| Variable | Effect |
|----------|--------|
| `XDG_STATE_HOME` | Registry and logs base |
| `XDG_DATA_HOME` | Chrome profiles base |
| `HOME` | Fallback when XDG unset |

## Platform

Official support: **Linux and WSL2** only (FR-014).

## Future compatibility

- Additional flags (`stop --id`, `status --once`) may be added without breaking positional behavior.
- `schema_version` in registry bumped on breaking JSON changes.
