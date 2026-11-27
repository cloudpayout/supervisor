# Casino support — 004 vs 020 reference

## CLI evolution

**004 (historical quickstart)** documented:

```text
supervisor start --casino onewin --session-file <path> --headless
```

**Current operator path (005+ / 020):** casino id lives in the session file; `start` takes the file path only:

```text
supervisor start --headless <session-file>
```

When writing new specs or tests, use the session-file form and `"casino":"<id>"` in JSON.

## Catalog entries (current)

From `internal/casino/catalog.go`:

- `onewin` → Page `https://1win.com/`, Game `aviator` → worker uses `PlayURL(locale)` for game path.
- `brazino` → Page `https://www.brazino777.mx/`, Game `""` → worker uses registered home page only.

Refused identifier examples:

- onewin: `Onewin`, `1win.com`
- brazino: `Brazino`, `brazino777`, `brazino777.bet.br`, `www.brazino777.mx`

## Integration test helpers

| Helper        | Location                                      | Role                          |
|---------------|-----------------------------------------------|-------------------------------|
| `newStandIn`  | `tests/integration/onewin_helpers_test.go`    | Shared balance gate           |
| `startCasino` | `onewin_start_test.go` / helpers              | Start with session file       |
| `requireChrome`, `isolatedEnv`, `supervisorBin` | `harness_test.go` | Test harness                  |

Brazino tests mirror onewin file names with `brazino_` prefix; same stand-in type.

## Worker cookie → navigate sequence

1. Re-read session file in worker.
2. `Network.setCookie` for each cookie.
3. Navigate to `SUPERVISOR_PAGE_URL` if set (tests), else `StartURL(casino, locale)` with stand-in locale rewrite when applicable.

## Spec numbering convention

`specs/004-open-onewin-page` — first casino, established stand-in + screenshot in ordinary confirmation.

`specs/020-open-brazino-page` — additive second casino; FR-015, SC-006, SC-007 multi-casino suite rules.

New casinos should pick the next feature number and `open-<id>-page` slug consistent with the branch name in `spec.md`.