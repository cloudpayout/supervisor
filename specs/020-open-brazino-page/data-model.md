# Data Model: Open the Brazino page

Entities for `specs/020-open-brazino-page/spec.md`. Builds on `specs/004-open-onewin-page/data-model.md` without removing `onewin`.

## Casino

Registered in code. The session file's `casino` field must match.

| Identifier | Page | Game |
|------------|------|------|
| `onewin` | `https://1win.com/` (catalog; worker opens play URL per 004) | `aviator` |
| `brazino` | `https://www.brazino777.mx/` (worker opens this home page) | *(empty — no registered game)* |

| Field | Rule |
|-------|------|
| Identifier | `brazino`. Lowercase ASCII letters only. `Brazino`, `brazino777`, `brazino777.bet.br`, `www.brazino777.mx` are refused as casino names (not as page URLs in code). |
| Page | `https://www.brazino777.mx/`. Operator does not type it. |
| Game | Empty for `brazino`. Nothing in this feature starts a game. |

Any other identifier remains a usage error at `session.Parse` / `casino.Lookup`.

## Game

No new game entity. `brazino` has no catalog game. `session info` prints `game: null` for `brazino` instances.

## Session file

Same JSON shape as 005+ / 004 extensions. Extends allowed `casino` value to `brazino`.

| Field | Rule |
|-------|------|
| `casino` | Required. Must be `brazino` for this feature's flows. Must match supported `Lookup`. |
| `cookies` | Array, length ≥ 1. Same cookie rules as 004. |

See [contracts/session-file.md](./contracts/session-file.md). [004 session-file](../004-open-onewin-page/contracts/session-file.md) rules still apply to `onewin`.

## Cookie

Unchanged from 004. Values are secrets; never in registry, argv, worker log, or screenshot file name.

## Browser instance

Registry row unchanged schema. `casino` may be `brazino` or `onewin` (or omitted on blank start).

| Field | Rule |
|-------|------|
| `casino` | `brazino` when started with a Brazino session file. Not a secret. |
| Other fields | Same as 004 / 001. |

## Screenshot

Unchanged command semantics. For a `brazino` row in `running`, the image is the page the worker opened after cookies (home page or stand-in).

## State

```text
blank start                    → unchanged
brazino + valid session file   → starting → cookies → home page → running
running + screenshot           → still running (003 rules)
stop / worker exit / reap      → profile removed, row removed
file casino not brazin/onewin  → parse or lookup error
brazino file + wrong registry  → worker mismatch error
```

`onewin` transitions from 004 remain valid and must not regress.