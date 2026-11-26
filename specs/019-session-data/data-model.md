# Data Model: Session Instance Info (019)

No registry schema change. `schema_version` stays 1. Session info is a read-only projection of one `Instance` row plus catalog lookup.

## Instance (read)

| Field | Used for `session info` | Display rule |
|-------|-------------------------|--------------|
| `id` | Positional argument | Must be one path segment before socket/capture rules apply elsewhere (same hygiene as screenshot: reject weird ids if product already rejects them) |
| `casino` | `casino` line | Identifier string, or literal `null` when empty |
| `created_at` | `date` line | UTC `YYYY-MM-DD HH:MM` (minute precision) |
| `state` | — | Not printed (use `status`) |
| `entry` | — | Not printed in v1 |
| Other fields | — | Never printed (no paths, no PIDs, no flags) |

Reconciliation (`registry.Apply`) may remove stale rows before lookup; if the id disappears, the command is **not found**.

## Session info record (ephemeral)

| Line | Source | Value |
|------|--------|-------|
| `casino` | `Instance.Casino` | Lowercase id or `null` |
| `game` | `casino.Lookup(Instance.Casino).Game` when lookup succeeds | Lowercase id or `null` |
| `strategy` | — | Always literal `null` (feature 019) |
| `date` | `Instance.CreatedAt` | UTC `YYYY-MM-DD HH:MM` |

## Catalog (read-only)

| Casino id | Game id (today) |
|-----------|-----------------|
| `onewin` | `aviator` |

Unsupported or empty casino → `game: null`.

## Operator I/O (ephemeral)

| Outcome | stdout | stderr |
|---------|--------|--------|
| Success | Four `label: value` lines only | Empty |
| Usage | Empty | Usage / `instance ID required` |
| Not found | Empty | `instance not found: <id>` |
| Registry I/O | Empty | `supervisor: …` |

## State transitions

```text
any tracked state ──session info id──► four stdout lines, empty stderr (success)
absent after reconcile ──session info id──► empty stdout, stderr not found
usage error ──► empty stdout, stderr usage
```

Session info does not mutate the registry except for side effects already implied by `registry.Apply` during reconcile (same as `status` one-shot).