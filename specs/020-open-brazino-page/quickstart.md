# Quickstart: Open the Brazino page

Validation for `specs/020-open-brazino-page/spec.md`. Contracts: [cli.md](./contracts/cli.md), [session-file.md](./contracts/session-file.md). Model: [data-model.md](./data-model.md). Normative `onewin` baseline: `specs/004-open-onewin-page/quickstart.md`.

## Prerequisites

- Linux or WSL2
- Go 1.27.1
- Google Chrome on `PATH`
- Same shared stand-in helper as 004 (`tests/integration/onewin_helpers_test.go`)

## Default suite

From the repository root:

```text
go test -json ./...
```

### Brazino ordinary confirmation (new)

Mirrors 004's onewin flow; uses the **same** `newStandIn` helper type, not a second implementation.

1. Start `newStandIn` on `127.0.0.1` with ephemeral port.
2. Write a temp session file: `"casino":"brazino"`, cookie `domain` `127.0.0.1`, generated secret.
3. Set `SUPERVISOR_PAGE_URL=http://127.0.0.1:<port>/`.
4. `supervisor start --headless <file>` → exit `0`, instance id; stand-in shows `balance` when cookie matches.
5. `supervisor status` → id `running`.
6. `supervisor screenshot <id>` → success ≤ 15s, PNG pixel matches shown color; instance still running.
7. `supervisor stop <id>` → profile removed, id gone from status.
8. Repeat with wrong cookie → withheld color, no `balance` marker in screenshot pixel check.
9. Assert secret absent from worker argv, screenshot argv, worker log, registry, screenshot filename.

### Session info

After step 4, `supervisor session info <id>` → lines include `casino: brazino` and `game: null` (UTC date format per 019).

### Usage failures (no Chrome)

| Case | Result |
|------|--------|
| `{"casino":"Brazino",...}` | Non-zero, no instance |
| `{"casino":"brazino777",...}` | Non-zero, no instance |
| `{"casino":"onewin",...}` with wrong file only | Still valid for onewin, not brazino tests |
| Missing file | Non-zero, no profile |

### Onewin non-regression (SC-006)

All existing `TestOneWin*` and Aviator integration tests pass with no operator workflow change. Run the full suite; do not skip `tests/integration/onewin_*`.

### Live home page (external)

Not part of `go test ./...`.

Unset `SUPERVISOR_PAGE_URL`. Run `supervisor start <operator's brazino session file>` (headed). Browser should open `https://www.brazino777.mx/`. Operator judges session privately. Do not commit the file or live screenshots. Rejected live cookies stop the feature per spec assumptions.

## Done checklist

- [ ] `casino.Lookup("brazino")` returns Mexico home URL, empty game
- [ ] Worker opens home page for `brazino`, play URL unchanged for `onewin`
- [ ] New integration tests use shared `newStandIn`
- [ ] `go test ./...` green including all 004 onewin tests
- [ ] No request to `https://www.brazino777.mx/` in default integration tests