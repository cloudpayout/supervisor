# Idea Intake: Real-time instance screen streaming

- **Slug**: real-time-streaming
- **Created**: 2026-11-24
- **Source**: pasted text; codebase pointer (`github.com/cloudpayout/supervisor` — CLI `start`, `status`, `stop`, `screenshot`, `logs` on isolated chromedp instances)
- **Type**: new-capability

## Idea (as captured)

> a new command to stream in real-time the screen of an instance

In this repository, operators already capture a **single** PNG per invocation via `supervisor screenshot <instance-id>` (15s budget, worker IPC). Logs are streamed line-by-line via `supervisor logs <instance-id>` with optional `--follow`. There is no CLI command today that continuously delivers live visual output from a running instance.

## Restated

Add a new `supervisor` CLI command that streams the display of a running browser instance to the operator in real time, rather than requiring repeated one-off screenshots.

## Origin & Context

- **Raised by**: Oscar Cedano <oscar.cedano@outlook.com>
- **Trigger**: Monitoring

## First-Glance Unknowns

- **Resolved**: MGPEG over HTTP. [NEEDS CLARIFICATION: What output format and channel — stdout video, MJPEG over HTTP, repeated frames, terminal UI, file pipe?]
- **Resolved**: Local environment. [NEEDS CLARIFICATION: Target audience and environment — local operator only, remote viewer, CI artifact?]
- **Resolved**: Frame rate may be configured using `--fps <integer>`, by default it should be `1 fps`.  [NEEDS CLARIFICATION: Performance and resource limits — frame rate, resolution, bandwidth, impact on instance/worker]
- **Resolved**: No relationship between `stream <instance-id>` and `screenshot <instance-id>`. [NEEDS CLARIFICATION: Relationship to existing `screenshot` — replace, complement, or share capture path?]
- **Resolved**: Same behavior for both. [NEEDS CLARIFICATION: Headless vs headed instances — same behavior for both?]
- **Resolved**: No, it would be local, there will be no multiple viewers. [NEEDS CLARIFICATION: Lifecycle — start/stop streaming, multiple viewers, behavior when instance stops]
- **Resolved**: Doesn't matter if sensitive page content is shown. [NEEDS CLARIFICATION: Security and privacy — sensitive page content in stream, authentication?]