# Research: Instance Watch (018)

**Date**: 2026-11-25

## R1 — Screencast on the worker tab, HTTP on the worker

**Decision**: The instance worker runs `page.StartScreencast` on the same `taskCtx` it already uses for navigation and `CaptureScreenshot`. It binds the MJPEG HTTP server itself and returns the URL to the CLI. The CLI does not open a browser and does not create a chromedp context.

**Rationale**: Assessment Option B and stakeholder resolutions put the listener on the worker. A second `chromedp.NewContext` would open another tab (spec 003 research R1). Headless and headed instances share that tab, so both modes use the same path (FR-009 of screenshot, watch FR-006).

**Alternatives considered**:

- CLI polls `CAPTURE` and builds MJPEG (concept Option C) — rejected. Stakeholder chose screencast on the worker, and a 1 Hz PNG poll duplicates the still-capture path under a long-lived CLI.
- CLI binds HTTP and the worker only pushes frames — rejected. The printed URL must survive only while watch is active, and the worker already owns Chrome lifetime.

## R2 — Same Unix socket, new verb

**Decision**: Extend the capture socket from spec 003. A connection's first line is either `CAPTURE` or `WATCH <fps>`. `CAPTURE` behavior is unchanged. `WATCH` replies `OK <url>` and holds the connection until the client hangs up or the worker shuts down.

**Rationale**: The socket path, permissions (`0600` / `0700`), and "not ready" retry already exist. A second socket would be another file to unlink on shutdown without an operator-visible benefit. Schema version stays 1.

**Alternatives considered**:

- Registry field for the watch URL — rejected. The URL is valid only while the CLI holds the connection; persisting it would lie after interrupt.
- A new TCP control channel — rejected. The Unix socket is already user-scoped; the only TCP surface should be the loopback viewer.

## R3 — Loopback port 8080 with fallback

**Decision**: `net.Listen("tcp", "127.0.0.1:8080")`, then `8081`, `8082`, … until success or a small ceiling (64 attempts). The URL printed is `http://127.0.0.1:<actual>/`. Never bind `0.0.0.0`.

**Rationale**: Stakeholder and spec FR-014. Each worker is its own process, so two watches are two listeners; the second fails `8080` and takes the next port. Binding only loopback matches "local viewer, no auth" without publishing the page on the LAN.

**Alternatives considered**:

- Fixed 8080 and fail if busy — rejected. Spec requires the next free port.
- Ephemeral `:0` only — rejected. Operators were promised 8080 when it is free.

## R4 — Throttle to `--fps`; ack off the listen goroutine

**Decision**: Request JPEG screencast frames, acknowledge each `screencastFrame` with `page.ScreencastFrameAck` in its own goroutine, and let the HTTP handler emit at most `fps` images per second (keep the latest JPEG, drop older ones). Default `fps` is 1. Values outside 1–30 are usage errors in the CLI before dialing.

**Rationale**: CDP `everyNthFrame` is not a wall-clock rate, and headless Chrome often emits a frame only when the page changes (chromedp screencast example). Wall-clock throttle meets SC-006 on a changing stand-in without failing a static page (User Story 3 scenario 4). chromedp issue 920 shows ack-on-the-listener-thread stalls the stream.

**Alternatives considered**:

- Timer plus `CaptureScreenshot` — rejected for v1 because the assessment locked screencast. Keep it as a fallback only if integration shows screencast never delivers a first frame within 10 seconds (SC-001).
- Unthrottled CDP rate — rejected. `--fps 30` would not be an operator cap.

## R5 — One watch; screenshot stays `CAPTURE`

**Decision**: A worker boolean `watching` guards `WATCH`. A second `WATCH` returns `ERR already watching` (exit 1). `CAPTURE` remains legal during watch. Both paths may touch the tab; screenshot keeps the existing mutex and 10-second child context. The screencast loop must not hold that mutex.

**Rationale**: FR-012 and FR-013. Serializing still capture avoids overlapping `chromedp.Run` on one target. Leaving screenshot on `CAPTURE` means spec 003 clients do not change.

**Alternatives considered**:

- Reject screenshot while watching — rejected. Spec requires parallel still capture.
- Share one in-flight queue that delays screenshot until the next frame slot — unnecessary at 1 fps if the mutex is only around the PNG call.

## R6 — Exit 0 when the session ends cleanly

**Decision**: After a URL has been printed, interrupt (CLI closes the socket, instance still running) and instance teardown (worker closes the socket) both exit 0. Failures before a URL, including not found, not running, `ERR`, and a dropped connection with no `OK` line, exit 1. Usage exits 2 via existing `isUsageError`.

**Rationale**: The operator asked to watch and then stopped watching, or the run ended and watch followed within 5 seconds (FR-010, FR-011). That is not a usage or lookup failure. Spec 003's "not found is exit 1" still applies when watch never starts.

**Alternatives considered**:

- Exit 1 when the instance stops during watch — rejected. Scripts that `stop` in another terminal would treat a correct follow-the-run as failure.
- Non-zero exit on Ctrl+C — rejected. `logs --follow` returns success on interrupt; watch should match that operator habit.

## R7 — No registry reconcile on watch

**Decision**: `watch` uses `Store.View` only. It does not call `registry.Apply`.

**Rationale**: Same as screenshot research R3. A failed watch must not stop other instances.
