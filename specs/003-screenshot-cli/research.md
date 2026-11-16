# Research: Screenshot CLI (003)

**Date**: 2026-11-15

## R1 — Capture on the worker's existing tab

**Decision**: The detached worker that already owns the instance runs `chromedp.Run` with `chromedp.CaptureScreenshot` on the same `taskCtx` it used for `Navigate("about:blank")`. The `screenshot` CLI does not open a browser and does not create a chromedp context.

**Rationale**: chromedp v0.23 documents that contexts which share a browser run in separate tabs (`Action.Run` in `chromedp@v0.23.0/action.go`). A second `NewContext` would add a tab and violate FR-005. `CaptureScreenshot` records the current viewport (`FromSurface` only, no clip, no `CaptureBeyondViewport`), which matches the spec: the current view, not a full-page stitch and not a desktop photo. Headless and headed instances both have that page target, so FR-009 does not need a window grab. The first `chromedp.Do(taskCtx, Navigate(...))` in `internal/worker/run.go` is what allocates the browser; a later `Run` on a child context does not.

**Alternatives considered**:

- CLI attaches with `remote.NewAllocator` and `chromedp.WithTargetID` — rejected. Default allocator options in v0.23 add `--remote-debugging-pipe` (`allocate.go`), so there is no DevTools port to attach to. Switching every instance to `remote.WebSocket` would publish a debugging port on the machine that holds casino credentials.
- `chromedp.FullScreenshot` — rejected. That captures beyond the viewport (FR assumption: current view only).
- `chromedp.Screenshot` of a CSS node — rejected. It scrolls clip math and is the wrong operator action.

## R2 — Ask the worker over a Unix socket

**Decision**: After the registry row is `running`, the worker listens on a Unix socket at `{XDG state}/supervisor/sockets/<id>.sock` (`registry.Layout.CaptureSocket`). The CLI dials that socket, sends one capture request, and receives PNG bytes or an error. The socket is mode `0600` in a `0700` directory. The worker removes the socket on shutdown. The instance ID must be a single path segment (`filepath.Base(id) == id`) before it is joined into that path.

**Rationale**: The socket path is derived from the ID, so `instances.json` stays schema version 1. Filesystem permissions limit the listener to the same user. One accept loop plus a mutex around `chromedp.Run` keeps concurrent captures of the same tab from racing (`action.go`: actions on one context share one tab). Different instances have different workers, so they capture in parallel (spec edge case).

**Alternatives considered**:

- Store the socket path on the registry row — rejected. The path is a pure function of the state dir and ID; a new field would bump `schema_version` for no operator change.
- SIGUSR1 plus a well-known output file — rejected. Signals cannot return a distinct not-found versus capture error, and they collide with the 15-second deadline.
- HTTP on `127.0.0.1` — rejected. A TCP port is easier to collide with and wider than a user-scoped socket.

## R3 — Do not reconcile during screenshot

**Decision**: `screenshot` reads the registry with `Store.View` only. It does not call `registry.Apply`.

**Rationale**: User Story 2 requires a failed capture to leave every other instance alone. `status` and `stop` remain the commands that reconcile. Lookup rules:

| Registry | Result |
|----------|--------|
| No row | `instance not found: <id>` (exit 1) |
| Row whose `state` is not `running` (`starting`, `stopping`, `stopped`) | `instance not running: <id>` (exit 1) |
| `running`, then the socket or capture fails | `screenshot failed: <id>: <reason>` (exit 1) |

A stopped instance whose row was already removed is not found. That matches "status does not list this ID" (User Story 2, scenario 2), including an instance stop has removed. "Not running" is reserved for a tracked row whose state is not `running`, including `starting`, `stopping`, and a leftover `stopped` row (scenario 3). v1 captures the worker's single blank-page tab. Another tab is out of scope.

**Alternatives considered**:

- Reconcile like `stop` before lookup — rejected. Reconciliation can terminate orphan Chrome for other IDs.
- Exit 0 for not found, matching `stop` — rejected. The spec assumption says a missing page is not a successful screenshot, and scripts must tell a usage error from a capture failure.

## R4 — Deadline that does not stop Chrome

**Decision**: The CLI cancels its own work after 15 seconds (FR-012). The worker runs `chromedp.Run` on `context.WithTimeout(taskCtx, 10s)`, not on `taskCtx` itself.

**Rationale**: chromedp v0.23 says the browser lives as long as the context of the first `Run`, and a timeout on that first context stops the browser. A child context limits one action. Ten seconds leaves the CLI time to publish the file or fail inside the 15-second command budget. Cancelling the child must not cancel `taskCtx` or the allocator context, and the worker must not navigate, resize, or close the tab.

**Alternatives considered**:

- Timeout on `taskCtx` — rejected. That is the context that owns the browser.
- No worker-side limit — rejected. A stuck DevTools call would pin the accept loop past the operator's deadline.

## R5 — Publish the PNG without replacing an existing file

**Decision**: Write the bytes to a temporary file in the destination directory with `O_EXCL`, then `os.Link` it to the final name, then remove the temporary file. `Link` fails if the final name exists, unlike `rename` on Linux. On any failure, remove the temporary file and do not print a path. Reject a final path that already exists, is a directory, or whose parent directory is missing, before asking the worker to capture.

**Rationale**: FR-007 forbids overwrite. FR-011 forbids an incomplete file presented as the screenshot. Checking the destination before capture avoids taking a picture that cannot be saved. The PNG is published only after the full byte count is on disk.

**Alternatives considered**:

- `os.Rename` onto the destination — rejected. Linux rename replaces an existing file.
- Write straight to the final name with `O_EXCL` — rejected. A crash mid-write leaves a partial PNG at the operator-visible path.
- Create missing parent directories — rejected. The operator named a file, not a directory tree to create. A missing parent is a write failure.

## R6 — Default name, relative path, and absolute path

**Decision**: `--output` / `-o` takes the operator's path as given. Relative paths are resolved against the current directory; absolute paths are used as absolute. Stdout prints that path string unchanged (trimmed) on success. With no flag, the file is `screenshot-<id>-<YYYYMMDD-HHMMSS>.png` in the current directory, using local time. If that name exists, try `-2`, `-3`, … before `.png`. Stdout then prints the relative file name. A blank `--output` value is a usage error.

**Rationale**: The spec's User Story 3 says the path may be relative or absolute and that the printed location is that path. The assumed default name is already in the spec. The numeric suffix covers two captures in the same second (FR-006) without overwrite.

**Alternatives considered**:

- Always print an absolute path — rejected. It would not equal a relative `--output` argument.
- UTC timestamps — rejected. The name is an operator label in the working directory; local time matches the clock they see.
- A random suffix on every file — rejected. The spec's name shape should stay readable; the suffix is only for collisions.

## R7 — Usage errors stay in the existing exit-2 class

**Decision**: `cobra.ExactArgs(1)`. A trimmed-empty ID returns `instance ID required`, which `isUsageError` already maps to exit 2. Extra arguments and unknown flags stay Cobra usage errors (exit 2). Not-found, not-running, and capture or write failures exit 1 via `cmd/supervisor` `main`. Success prints only the path on stdout and exits 0. Errors go to stderr.

**Rationale**: Same split as `stop` missing-ID (002) and `start` flag misuse (001), except not-found is exit 1 here (R3).

**Alternatives considered**:

- Exit 2 for not-found — rejected. The ID was supplied; the instance is an operational miss, not a bad invocation.
