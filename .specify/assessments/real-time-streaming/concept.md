# Concept: Local visual monitoring for a running instance

- **Slug**: real-time-streaming
- **Created**: 2026-11-25
- **Recommended option**: Option B — Worker-served loopback MJPEG watch

## Options

### Option A — Do nothing (operator workarounds)

- **Sketch**: Leave Supervisor as-is. Operators monitor by tailing **logs**, polling **status**, repeating **screenshot** in a shell loop, watching a **headed** window, or using OS/browser tools outside the CLI contract. Document those patterns in quickstart or assessment notes only.
- **Appetite**: small (days) for documentation only; **none** for product change
- **Trade-offs**: No worker/CLI risk or spec-003 scope creep. All continuous-observation goals and success metrics stay unmet; monitoring remains ad hoc and easy to miss during background phases (problem cost of inaction).
- **Rabbit holes**: Treating a one-line wrapper script as “productized” without lifecycle when the instance stops, cadence limits, or parallel still capture.

### Option B — Worker-served loopback MJPEG watch

- **Sketch**: Add a **local watch** operator flow: for one **running** instance, the operator starts continuous observation from the CLI; the product prints a **loopback URL** (default port **8080**, another port if busy—per stakeholder). The operator opens that URL in a browser and sees a **multipart MJPEG** view of the instance viewport until they interrupt or the instance stops (**immediate** end). Update cadence defaults to **1/s**, configurable up to **30/s**; sparse frames on static headless pages are acceptable at low cadence. **Screenshot** remains a separate on-demand still capture and **may run while** observation is active. **Headless** and **headed** behave the same from the operator’s perspective.
- **Appetite**: **medium** (weeks) — new command, worker capture/screencast lifecycle, loopback HTTP, shutdown rules, integration tests; not a one-day tweak given mutex, ports, and spec 003 adjacency
- **Trade-offs**: Meets continuous local watch, cadence control, end-of-run stop, and parallel still-capture goals. Research/stakeholder direction favors **CDP screencast → MJPEG** on the **worker** (not extending the one-shot CAPTURE socket alone). Reopens the deliberate “no video” boundary from screenshot spec 003. Risks: port clashes with other loopback services (e.g. stand-ins), screencast **ack** and headless **change-driven** frames, capture contention with automation and **screenshot**, thin demand (single author).
- **Rabbit holes**: Multi-instance port matrix; authenticating loopback; remote viewers; DVR/recordings; full-frame PNG polling at high fps; rewriting screenshot; live casino in default tests; guaranteeing 30/s on loaded workers without spike data.

### Option C — CLI-assembled slideshow from existing still captures

- **Sketch**: A new CLI command dials the **existing** per-instance capture channel on a timer, receives **PNG** frames, and serves **MJPEG on loopback from the CLI process** while the command runs. The worker stays on the current one-shot capture contract; no long-lived screencast in the worker.
- **Appetite**: **medium** (weeks) — still a new command and HTTP stream, but worker changes may be smaller
- **Trade-offs**: Reuses familiar viewport semantics with **screenshot**. Can approximate **1/s** even when the page is static. Higher CPU/bandwidth if each tick is a full PNG; awkward overlap with **screenshot** on the same mutex (longer waits). Less aligned with stakeholder/research resolution to use **screencast on the worker** and bind HTTP there.
- **Rabbit holes**: Pushing high **--fps** through repeated full PNG captures; CLI dies but worker keeps screencast assumptions undefined; duplicating two capture philosophies forever.

## Recommendation

**Option B.** It is the only option that fully meets the problem goals (continuous local watch for **≥ 5 minutes**, prompt stop when the instance ends, cadence **1–30/s**, parallel still capture) inside the Supervisor CLI without outsourcing to shell loops. Stakeholder and research resolutions already point to **worker-bound loopback MJPEG**, **screencast-style** frames, **8080-with-fallback**, concurrent **screenshot** allowed, and acceptance of sparse headless updates at default cadence. Option A only if appetite is cut to zero; Option C is a fallback if screencast proves unstable on headless Chrome in the pre-specify spike.

Evidence and demand remain **medium / thin** (one stated requester). Proceed to decide only if that appetite is still acceptable.

## Out of Scope (for the recommended option)

- Remote or multi-viewer access, CI-published streams, recordings library.
- Replacing **screenshot** as evidence command or merging the two into one command.
- Content redaction on the local watch URL.
- Multi-instance mosaic or gallery.
- Gameplay automation beyond current worker scope.
- Authentication on loopback (stakeholder waived).

## Assumptions to Validate

- Pre-implementation **spike**: p95 frame latency and bytes/s at **1280×720** on stand-in pages; acceptable impact on entry automation when observation runs at default cadence.
- **CDP screencast** (+ correct frame **ack** behavior) is reliable for both headless and headed profiles on Linux/WSL2 targets.
- **Per-instance loopback port** selection (8080 default, fallback when taken) is deterministic and printable to the operator; behavior with **multiple running instances** is clear.
- **Screenshot** and observation can share capture resources without unacceptable stalls at **1/s** default (stakeholder allows both).
- Sparse updates at **1/s** on static headless pages still satisfy operator “monitoring” intent (stakeholder: yes).