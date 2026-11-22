# Quickstart: Validate Headless Viewport (012)

Runnable checks after implementation. See [cli contract](./contracts/cli.md), [worker contract](./contracts/worker-viewport.md), and [data model](./data-model.md). Build and `XDG_*` setup match [001 quickstart](../001-supervisor-cli/quickstart.md).

## Prerequisites

- Built binary: `go build -o supervisor ./cmd/supervisor`
- Chrome/Chromium on `PATH` or `CHROME_PATH`
- Isolated state: `XDG_STATE_HOME` and `XDG_DATA_HOME` per 001

## Scenario 1 — Blank headless screenshot dimensions (P1, SC-001, SC-005)

1. Export isolated state dirs (see 001).
2. `./supervisor start --headless` → note `id`.
3. `./supervisor status` → instance `running`, `headless` true if shown.
4. `./supervisor screenshot "$id"` → note PNG path on stdout.
5. **Expect**: PNG decodes to width **1280**, height **720** (e.g. `identify` from ImageMagick, or a small Go/png decode script).
6. `./supervisor stop "$id"` → instance removed.

**Spec refs**: FR-001, FR-003, FR-006; User Story 1.

## Scenario 2 — Headed start unchanged (P2, SC-002)

1. `./supervisor start` (no `--headless`) → note `id`.
2. **Expect**: visible Chrome window; registry headed / not headless.
3. Optional: screenshot; dimensions are **not** required to be 1280×720.
4. `./supervisor stop "$id"`.

**Spec refs**: FR-002; User Story 2.

## Scenario 3 — Help mentions layout (P3, FR-007)

1. `./supervisor start --help`
2. **Expect**: text states headless uses **1280×720** layout (flag and/or long help).

**Spec refs**: User Story 3.

## Scenario 4 — Two headless instances (isolation)

1. Start two `--headless` instances (`id1`, `id2`).
2. Screenshot each; both PNGs 1280×720.
3. `./supervisor stop all` or stop each.

**Spec refs**: User Story 1 acceptance 3.

## Automated regression

```text
go test ./... -count=1
```

After implementation, includes `TestHeadlessViewportScreenshot` (or equivalent) in `tests/integration/headless_viewport_test.go` — skipped when Chrome unavailable, same as other integration tests.

**Spec refs**: FR-006, constitution default test run without live casino.