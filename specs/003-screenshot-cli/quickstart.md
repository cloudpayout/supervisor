# Quickstart: Validate Screenshot CLI (003)

Runnable checks after implementation. See [cli contract](./contracts/cli.md), [capture socket](./contracts/capture-socket.md), and [data model](./data-model.md). Build and isolated `XDG` env match [001 quickstart](../001-supervisor-cli/quickstart.md).

## Prerequisites

- Built binary: `go build -o supervisor ./cmd/supervisor`
- Chrome or Chromium on `PATH`, or `CHROME_PATH`
- Isolated state (recommended): `XDG_STATE_HOME` and `XDG_DATA_HOME` pointing at empty directories
- A fresh instance started by this binary. A worker built before this feature has no capture socket

## Scenario 1 — Capture a running instance (P1, SC-001, SC-003, SC-005)

1. `./supervisor start --headless`. Note the ID printed on stdout.
2. Poll `./supervisor status` until that ID shows `running`.
3. `./supervisor screenshot "$id"`.
4. **Expect**: Exit `0` within 15 seconds. Stdout is one relative name `screenshot-<id>-<timestamp>.png`.
5. **Expect**: That file exists in the current directory, is non-empty, and begins with the PNG signature.
6. Note the instance's owner and browser process before the capture, then `./supervisor status`.
7. **Expect**: The same ID is still `running` with those same processes, still on the blank page.
8. Run `./supervisor screenshot "$id"` again in the same second if you can, otherwise immediately.
9. **Expect**: A second file. The first file's bytes are unchanged.

**Spec refs**: FR-001, FR-004, FR-005, FR-006, FR-012; User Story 1.

## Scenario 2 — Two instances, with and without a window (P1, FR-008, FR-009)

The default automated check starts two `--headless` instances, captures each, and expects both to stay running with the same processes. The steps below add the visible-window half, which is manual when a display is available.

1. Start one headed instance and one `--headless` instance. Wait until both are `running`.
2. Capture each ID into its own `--output` file.
3. **Expect**: Both commands exit `0`. Each file is a PNG. `status` still lists both IDs as `running`, each with the same processes as before the capture.

**Spec refs**: User Story 1, scenarios 3 and 4.

## Scenario 3 — Refuse a bad capture (P1, SC-002, SC-004)

1. `./supervisor screenshot` with no ID, and `./supervisor screenshot some-id extra`.
2. **Expect**: Both exit `2`. No new PNG.
3. `./supervisor screenshot missing-id`.
4. **Expect**: stderr `instance not found: missing-id`. Exit `1`. No file.
5. While an instance is still `starting`, or with a seeded registry row whose state is `stopped` or `stopping`, `./supervisor screenshot "$id"`.
6. **Expect**: stderr `instance not running: <id>`. Exit `1`. No file.
7. Stop a running instance, then screenshot its old ID.
8. **Expect**: `instance not found` (the row is gone) and exit `1`. Other instances are unchanged aside from the stop you already performed.

**Spec refs**: FR-002, FR-003, FR-010, FR-011; User Story 2.

## Scenario 4 — Chosen path (P2, SC-006)

1. With a running instance: `./supervisor screenshot "$id" -o shots/page.png` where `shots/` exists and `page.png` does not.
2. **Expect**: Exit `0`. Stdout is `shots/page.png`. The file is there. No `screenshot-*.png` appeared in the current directory.
3. Repeat the same command.
4. **Expect**: Exit `1`. stderr contains `screenshot failed`. `shots/page.png` is unchanged.
5. `./supervisor screenshot "$id" -o /tmp/supervisor-shot-$$.png` with a path that does not exist.
6. **Expect**: Exit `0`. Stdout is that absolute path. The file there is a PNG.
7. `./supervisor screenshot "$id" -o shots` where `shots` is a directory.
8. **Expect**: Exit `1`. The directory is unchanged.

**Spec refs**: FR-007; User Story 3.

## Scenario 5 — Help (FR-013)

1. `./supervisor screenshot --help`.
2. **Expect**: Text names the required instance ID, `--output` and `-o`, relative and absolute paths, and the default file location. It does not describe a prompt.

## Automated regression

```bash
go test ./...
go test ./tests/integration/ -run Screenshot
```

**Expect**: Path, failed-publish, and exit-code tests pass without Chrome. Integration capture tests skip when Chrome is absent and otherwise perform Scenarios 1, 3, and 4 against headless Chrome, including two headless instances. The visible-window half of Scenario 2 stays manual. No test opens a casino site.
