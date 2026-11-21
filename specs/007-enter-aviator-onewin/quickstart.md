# Quickstart: Enter Aviator on onewin

**Feature**: `007-enter-aviator-onewin`

## Prerequisites

- Linux or WSL2, Go 1.27.1+, Chrome as in feature 001
- Built `supervisor` binary (`go build -o supervisor ./cmd/supervisor`)

## Ordinary confirmation (default `go test`)

```bash
go test ./...
```

Includes integration tests that:

1. Start loopback Aviator stand-in on `127.0.0.1`
2. Set `SUPERVISOR_PAGE_URL` to stand-in URL
3. Run `start` with session JSON (`casino`, cookies for `127.0.0.1`, optional `currency` / `demo` / `locale`)
4. Poll `status` until `ready` or instance gone (timeout case)
5. Run `screenshot`, assert post-ready marker
6. Run `stop`; assert profile removed and secrets absent from logs/registry

No request to live `1win.com`.

## Manual live check (operator, not CI)

1. Prepare session file with real cookies for `1win.com` and optional `currency`, `demo`, `locale` (see [contracts/session-file.md](./contracts/session-file.md)).
2. **Do not** set `SUPERVISOR_PAGE_URL`.
3. Run headed start: `start /path/to/session.json` (omit `--headless` unless you accept no eyeball check).
4. After `start` exits, watch `status` until it shows `ready` (or instance disappears on failure).
5. `screenshot <id>` within 15s for private verification.
6. `stop <id>`.

## Negative cases (stand-in)

| Case | Session | Expected |
|------|---------|----------|
| Bad currency | `"currency":"ZZZ"` | Instance removed from `status`; no CLI diagnostic |
| Timeout | Stand-in configured to never show `game-logo` | Instance removed within ~60s |

## Related docs

- [spec.md](./spec.md) — requirements
- [data-model.md](./data-model.md) — fields and registry `entry`
- [contracts/aviator-entry.md](./contracts/aviator-entry.md) — probe and modal contract