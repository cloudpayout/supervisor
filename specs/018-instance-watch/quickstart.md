# Quickstart: Instance Watch (018)

Runnable checks after implementation. See [cli contract](./contracts/cli.md), [watch socket](./contracts/watch-socket.md), and [data model](./data-model.md). Build and isolated `XDG` env match [001 quickstart](../001-supervisor-cli/quickstart.md).

## Prerequisites

- `go` 1.27.1 and Chrome or Chromium on `PATH` (or `CHROME_PATH`) for the live checks
- A fresh instance started by this binary. A worker built before this feature rejects `WATCH`

## Scenario 1 — Watch a headless instance (P1, SC-001, SC-008)

1. `./supervisor start --headless` and note the instance ID.
2. Wait until `./supervisor status` shows it `running`.
3. `./supervisor watch "$id"`.
4. **Expect**: Within 10 seconds, stdout is one line `http://127.0.0.1:<port>/` and the process is still running.
5. Open that URL on the same machine (browser or `curl --no-buffer`). **Expect**: multipart JPEG body, not an error page.
6. Interrupt watch (Ctrl+C).
7. **Expect**: Exit `0`. `status` still lists the same ID as `running` with the same processes.

## Scenario 2 — Cadence (P2, SC-006)

1. With a stand-in page that changes at least once per second, `./supervisor watch "$id"` (default fps).
2. **Expect**: About 0.8–1.2 image parts per second over 60 seconds.
3. Interrupt, then `./supervisor watch "$id" --fps 10`.
4. **Expect**: About 8–12 parts per second over 60 seconds on that changing page.
5. `./supervisor watch "$id" --fps 0` and `--fps 31`.
6. **Expect**: Exit `2`, stderr mentions invalid `--fps`, no URL.

## Scenario 3 — Refuse a bad watch (P1, SC-004, SC-005)

1. `./supervisor watch` with no ID, and `./supervisor watch some-id extra`.
2. **Expect**: Exit `2`. No URL.
3. `./supervisor watch missing-id`.
4. **Expect**: Exit `1`. stderr contains `instance not found`. No URL.
5. While an instance is `starting`, `./supervisor watch "$id"`.
6. **Expect**: Exit `1`. stderr contains `instance not running`.
7. Start watch, then in another terminal `./supervisor watch "$id"`.
8. **Expect**: Second command exit `1` with `already watching`. First watch still serves.

## Scenario 4 — Stop ends the watch (P1, SC-003)

1. Start watch and note the URL.
2. In another terminal, `./supervisor stop "$id"`.
3. **Expect**: The watch process exits `0` within 5 seconds. A new request to the URL fails (connection refused or reset).

## Scenario 5 — Screenshot while watching (P2, SC-007)

1. Start watch.
2. In another terminal, `./supervisor screenshot "$id"`.
3. **Expect**: Screenshot exit `0` and a PNG. Watch is still running and the URL still returns frames.
4. Interrupt watch. Instance still `running`. Then `./supervisor stop "$id"`.

## Scenario 6 — Help

1. `./supervisor watch --help` and `./supervisor --help`.
2. **Expect**: `watch`, instance ID, `--fps`, local URL, and interrupt behavior are described. Root help lists `watch`.

## Automated default

`go test ./...` covers argument parsing and port selection without Chrome. Integration watch tests skip when Chrome is absent and otherwise run Scenarios 1, 3, 4, and 5 against headless Chrome and a local stand-in. Scenario 2's 60-second rate check and any multi-minute soak use build tag `slow` or stay manual. Headed watch stays manual. No test opens a casino site.
