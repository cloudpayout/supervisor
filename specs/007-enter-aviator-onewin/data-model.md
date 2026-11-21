# Data Model: Enter Aviator on onewin

**Feature**: `007-enter-aviator-onewin`

## Session file (amends 004/005)

Top-level JSON object.

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `casino` | string | yes | Must pass `casino.Lookup` (`onewin` only) |
| `cookies` | array | yes | Length ≥ 1; each cookie per 004 rules |
| `currency` | string | no | If present: exactly 3 chars, `[A-Z]{3}` |
| `demo` | boolean | no | Default effective value `false` |
| `locale` | string | no | If present: non-empty, no `/`, used as single path segment (e.g. `es-MX`) |

Unknown top-level keys: error at parse time.

**Play preferences** (derived at parse time): effective `currency` (optional), `demo` (default false), `locale` (optional).

Cookie values never stored in registry or logs.

## Casino catalog (amends `internal/casino`)

| Field | Meaning |
|-------|---------|
| `ID` | `onewin` |
| `Game` | `aviator` |
| `Page` | Main site URL (cookie default origin), `https://1win.com/` |
| Play URL | From `PlayURL(locale string) string`: empty locale → `https://1win.com/casino/play/v_spribe:aviator`; else `https://1win.com/{locale}/casino/play/v_spribe:aviator` |

`SUPERVISOR_PAGE_URL` env overrides the navigated URL for tests (unchanged seam).

## Registry instance (amends feature 001)

Existing fields unchanged. New optional field:

| Field | Type | Values | When set |
|-------|------|--------|----------|
| `entry` | string | `entering`, `ready` | `onewin` session starts only |

Transitions:

1. After worker begins Aviator entry: `entry` = `entering`.
2. After **Aviator ready**: `entry` = `ready`.
3. On entry failure before ready: instance row removed (no terminal failure field).
4. On `stop` or browser exit: row removed per existing rules.

No cookie or session preference values in the registry.

## Aviator ready (runtime state, not persisted)

Logical state when modal automation succeeded and iframe probe succeeded within timeout. Not a separate DB; implied by `entry=ready`.

## Entry failure causes (internal only)

Enumerated in worker for tests and maintenance: `session`, `currency`, `controls`, `timeout`, `unexpected`. Not exposed on CLI per FR-010.

## Local Aviator stand-in (test-only)

HTTP origin `http://127.0.0.1:<port>/`. DOM contract mirrors [contracts/aviator-entry.md](./contracts/aviator-entry.md) probe rules with `data-supervisor` hooks.