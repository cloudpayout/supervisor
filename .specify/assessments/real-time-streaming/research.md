# Idea Research: Real-time instance screen streaming

- **Slug**: real-time-streaming
- **Created**: 2026-11-24
- **Evidence confidence (overall)**: medium

## Users & Demand

- One stated request from Oscar Cedano: a new command to stream an instance screen in real time, with trigger **monitoring** during supervised runs. Format **MJPEG over HTTP**, local operator only, default **1 fps** via `--fps`, no multi-viewer, same behavior for headless and headed, independent of `screenshot`, and no requirement to hide sensitive page content on the stream. — [source: `.specify/assessments/real-time-streaming/intake.md`] (confidence: high)
- No other demand signal in this repository: no issues, tickets, metrics, or prior assessment for live video/streaming. The want is stated once; it is not observed usage from an existing command. — [source: repository search of `.specify/assessments/`, `specs/`, and `internal/cli/`; only `screenshot` and `logs` exist today] (confidence: high for absence in this repo)
- Operators already use **`screenshot`** for human evidence during long or background flows. Feature 007 documents inferring Aviator entry failure via **private observation** (e.g. `screenshot` during the window or a headed browser), and assessment **aviator-in-onewin** treats `screenshot` as operator verification, not continuous monitoring. — [source: `specs/007-enter-aviator-onewin/spec.md` clarifications; `.specify/assessments/aviator-in-onewin/problem.md`, `decision.md`] (confidence: high that screenshot is the current visual tool; low that operators cannot meet monitoring needs without streaming)

## Prior Art

- **Internal — `screenshot` (spec 003)**: One PNG per CLI invocation; worker runs `chromedp.CaptureScreenshot` on the existing tab; CLI dials a per-instance Unix socket (`CAPTURE` → `OK <n>` + PNG). Captures are mutex-serialized; worker bound is 10s per capture; CLI budget 15s; max PNG 32 MiB. Plan explicitly scoped **“No gallery, video, or multi-instance capture.”** — [source: `specs/003-screenshot-cli/plan.md`, `contracts/capture-socket.md`, `internal/worker/capture.go`, `internal/worker/run.go`, `internal/cli/screenshot.go`] (confidence: high)
- **Internal — `logs --follow`**: Long-running stdout streaming with 500ms polling and signal-driven exit; precedent for a blocking CLI that follows live output, but over log files not browser pixels. — [source: `internal/cli/logs.go`; `specs/008-log-files/research.md`] (confidence: high)
- **Internal — no `stream` command**: Root registers `start`, `status`, `stop`, `screenshot`, `logs`, `worker` only. — [source: `internal/cli/root.go`] (confidence: high)
- **External — Chrome DevTools / chromedp**: Supervisor already depends on `github.com/chromedp/chromedp v0.23.0` and uses viewport capture, not screencast. CDP exposes `Page.startScreencast` / `Page.screencastFrame` with JPEG/PNG frames and mandatory `Page.screencastFrameAck`. chromedp documents `CaptureScreenshot` as viewport capture via `page.CaptureScreenshot`. — [source: `go.mod`; `https://raw.githubusercontent.com/chromedp/chromedp/master/screenshot.go`; `https://github.com/chromedp/cdproto/blob/master/page/page.go`] (confidence: high)
- **External — chromedp screencast example**: Official example records JPEG frames via `StartScreencast`, acknowledges each frame immediately, and notes **“A headless browser sends a frame only when the page changes, so a page that does not change gives few frames.”** It does not implement MJPEG HTTP or tie into a long-lived worker. — [source: `https://raw.githubusercontent.com/chromedp/examples/master/screencast/main.go`] (confidence: high)
- **External — screencast pitfalls**: chromedp issue #920: screencast stalls if `ScreencastFrameAck` runs on the listener thread; must ack in a separate goroutine. — [source: `https://github.com/chromedp/chromedp/issues/920`] (confidence: high)
- **MJPEG over HTTP**: Intake specifies MJPEG for local viewing. No MJPEG server or multipart writer exists in this repo. Mapping **1 fps** configurable `--fps` to either repeated `CaptureScreenshot` loops or CDP screencast with throttling is an implementation choice not yet specified. — [source: intake note; repo search — no matches for `mjpeg`, `Screencast`, `StartScreencast`] (confidence: high for absence; medium for feasibility of MJPEG as stated pattern — [ASSUMPTION] that standard `multipart/x-mixed-replace` over loopback HTTP is sufficient for a single local viewer)

## Market & Context

- **Cost of doing nothing**: Operators can run `screenshot` repeatedly, use `logs --follow`, run `status`, or watch a headed window. For background Aviator entry, product text already steers operators to timed screenshots or headed observation when CLI diagnostics are thin. — [source: `specs/007-enter-aviator-onewin/spec.md`; intake trigger “monitoring”] (confidence: high)
- **Alternatives outside this CLI**: OS-level screen capture of a headed Chrome window, remote debugging / DevTools, or external browser automation with built-in screencast — none are integrated into `supervisor` today. — [ASSUMPTION] (confidence: low)
- Headless instances use a fixed **1280×720** layout (spec 012); stream resolution likely follows viewport unless spec says otherwise. — [source: `specs/012-headless-mode-viewport/spec.md`] (confidence: high)

## Data & Constraints

- **Platform**: Module targets **Go 1.27.1**; v1 CLI support is **Linux and WSL2** per constitution and early specs. Local HTTP implies binding a loopback listener; port selection, collision with worker loopback (e.g. Aviator stand-in on `127.0.0.1`), and firewall policy are unspecified. — [source: `go.mod`; `specs/001-supervisor-cli/spec.md` FR-014; `specs/007-enter-aviator-onewin/spec.md` loopback stand-in] (confidence: high for platform; [NEEDS CLARIFICATION] for port/lifecycle)
- **Capture path contention**: Today one mutex serializes all capture-socket work per instance. Intake says **no relationship** between `stream` and `screenshot` but does not say whether they may run concurrently on the same tab. If both hit the same chromedp context without coordination, they can block or interleave. — [source: `internal/worker/capture.go`, `specs/003-screenshot-cli/contracts/capture-socket.md`; intake note] (confidence: high)
- **Headless frame rate**: At default **1 fps**, repeated PNG capture might approximate a slideshow; CDP screencast in headless may still emit sparse frames when the DOM is static (chromedp example), which can look like “frozen” video even at high `--fps`. — [source: chromedp screencast example; intake default 1 fps] (confidence: medium)
- **Performance unmeasured**: No benchmark for per-frame `CaptureScreenshot` latency at 1280×720 on this codebase, JPEG size at quality defaults, or CPU/memory impact while casino automations run. Screenshot budget allows up to 15s per single frame. — [source: `internal/cli/screenshot.go` `screenshotBudget`; `specs/003-screenshot-cli/plan.md`] (confidence: high that budgets exist; low on steady-state 1 fps load — unmeasured)
- **Security**: Intake accepts exposing sensitive page content on a **local** MJPEG endpoint with **no authentication**. Capture socket today is Unix-only `0600` under state dir; HTTP on loopback widens access to any local process that can open the URL. — [source: intake note; `specs/003-screenshot-cli/contracts/capture-socket.md`] (confidence: high)
- **Constitution**: New operator behavior should remain CLI-first; credentials must not appear in logs/registry/cmdline — streaming must not leak session material in URLs or multipart headers beyond what the page already shows. — [source: `.specify/memory/constitution.md` — not fully re-quoted here] (confidence: medium)

## Evidence Against the Idea

- **Thin demand**: One intake author and no corroborating tickets or usage data; screenshot plus headed observation already cover documented operator workflows. — [source: intake note; `specs/007-enter-aviator-onewin/spec.md`; assessments search] (confidence: high)
- **Prior scope rejection**: Spec 003 deliberately excluded video/streaming; this idea reopens that boundary and adds worker/CLI surface area beyond “one image per invocation.” — [source: `specs/003-screenshot-cli/plan.md` Scale/Scope] (confidence: high)
- **Headless “real-time” mismatch**: CDP screencast documentation in the chromedp example warns headless may not emit frames on a steady clock when the page is idle; **1 fps** default may not mean one new frame per second in practice. — [source: chromedp screencast example] (confidence: medium)
- **Implementation risk**: Screencast requires correct async ack handling (issue #920). MJPEG HTTP needs a long-lived server goroutine in CLI or worker, clean shutdown on Ctrl+C, and behavior when the instance stops mid-stream — none of which exist. Intake says stream and screenshot are unrelated, which risks duplicate capture mechanisms unless design merges them. — [source: issue #920; intake note; internal architecture] (confidence: medium)
- **Automation interference**: Extra capture load during iframe-heavy casino entry may contend with the same worker mutex and chromedp context used for navigation and modals. Impact on Aviator entry latency was not measured. — [ASSUMPTION; source for mutex: `internal/worker/capture.go`] (confidence: low)
- **Low default fps**: At **1 fps**, an operator may get little benefit over running `screenshot` every second in a shell loop, minus MJPEG browser viewing convenience. — [source: intake default; comparison to existing `screenshot` command — ASSUMPTION on operator workflow] (confidence: medium)

## Gaps & Open Questions

- **Resolved**: The port will be 8080; a worker wil bind the HTTP server. [NEEDS CLARIFICATION: Which URL/host/port does the CLI print for the MJPEG endpoint, and who binds the HTTP server — CLI process or worker?]
- **Resolved**: Use `Page.startScreencast` and encode it into MJPEG. [NEEDS CLARIFICATION: Implementation path — CDP `Page.startScreencast` vs periodic `CaptureScreenshot` frames encoded into MJPEG?]
- **Resolved**: Allowed. [NEEDS CLARIFICATION: Concurrent `screenshot` while `stream` is active — allowed, blocked, or shared queue?]
- **Resolved**: Immediate stream end. [NEEDS CLARIFICATION: Behavior when instance leaves `running` (stop, crash, failed Aviator entry) — immediate stream end, error page, or hang?]
- **Resolved**: Maximum 30fps and yes, it is okay the frames are sparse at default 1 fps for monitoring [NEEDS CLARIFICATION: Maximum `--fps` and whether headless sparse frames are acceptable at default 1 fps for “monitoring.”]
- **Resolved**: Defer hard numbers to a pre-implementation spike (p50/p95 capture latency and JPEG bytes at 1280×720 on headless stand-in pages, optional pass while entry automation runs). Until then, plan on ≤10s per-frame hard cap (spec 003 family), ~50–150 KiB/s at default 1 fps, and skip or drop frames if p95 capture nears 1s. Integration: new `stream` tests only; do not run background stream during existing `screenshot` tests; assert lifecycle and multipart protocol, not pixel timing; optional soak behind `slow` build tag.

## Sources

- `.specify/assessments/real-time-streaming/intake.md` (host: local, policy: internal artifact)
- `internal/cli/root.go`, `internal/cli/screenshot.go`, `internal/cli/logs.go`, `internal/worker/capture.go`, `internal/worker/run.go`, `go.mod` (host: local, policy: internal artifact)
- `specs/003-screenshot-cli/plan.md`, `specs/003-screenshot-cli/contracts/capture-socket.md`, `specs/007-enter-aviator-onewin/spec.md`, `specs/012-headless-mode-viewport/spec.md`, `specs/008-log-files/research.md` (host: local, policy: internal artifact)
- `https://raw.githubusercontent.com/chromedp/chromedp/master/screenshot.go` (host: raw.githubusercontent.com, policy: allowlisted)
- `https://raw.githubusercontent.com/chromedp/examples/master/screencast/main.go` (host: raw.githubusercontent.com, policy: allowlisted)
- `https://github.com/chromedp/cdproto/blob/master/page/page.go` (host: github.com, policy: allowlisted)
- `https://github.com/chromedp/chromedp/issues/920` (host: github.com, policy: allowlisted)