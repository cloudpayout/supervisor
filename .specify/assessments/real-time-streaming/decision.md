# Decision: Local visual monitoring for a running instance

- **Slug**: real-time-streaming
- **Decided**: 2026-11-25
- **Verdict**: go
- **Artifacts reviewed**: intake.md | research.md | problem.md | concept.md

## Scorecard

| Criterion | Rating | Justification |
|-----------|--------|---------------|
| Problem validity | adequate | Operators lack continuous visual observation inside the CLI during long/background runs; `logs`/`status` are thin and discrete `screenshot` is awkward for monitoring (`problem.md`, spec 007 observation patterns in `research.md`). Workarounds exist but do not meet stated goals. |
| Evidence strength | adequate | `research.md` overall confidence **medium**: codebase capture socket, chromedp/CDP screencast prior art, and stakeholder resolutions on format, cadence, ports, and lifecycle are cited; demand is a **single** intake author with no corroborating tickets—not **weak** because technical path and constraints are grounded, not **strong** because usage is unstated beyond one request. |
| Value vs. inaction | adequate | Monitoring during supervised automation (e.g. background entry) beats repeated captures or headed-only watching, but at default **1/s** value is incremental over shell loops (`research.md` evidence against—accepted by stakeholder sparse-frame rule). |
| Feasibility / appetite | adequate | **Option B** at **medium (weeks)** is credible on Linux/WSL2 with existing chromedp worker; pre-specify **spike** and screencast ack risks are explicit (`concept.md`, research gap resolutions). Option C is documented fallback. |
| Strategic fit | adequate | New operator CLI capability for **observing** running instances aligns with constitution II (CLI operator surface) and III (isolation and monitoring); does not expand casino/game scope. Reopens spec 003 “no video” boundary deliberately and within assessment non-goals. |
| Risk posture | adequate | Port collision, capture mutex contention, headless sparse frames, and loopback exposure are identified; mitigations include default low cadence, immediate stop when instance ends, spike before specify, bounded **30/s** cap, and parallel `screenshot` allowed with shared-resource awareness. |

## Verdict & Rationale

**Go** on **Option B — Worker-served loopback MJPEG watch**. Problem validity and value are **adequate** (real operator pain, acceptable urgency per stakeholder priority resolution in `problem.md`). Evidence is **adequate**, not **strong**: technical feasibility is supported by shipped capture infrastructure and CDP screencast art, while demand remains thin. Feasibility and risk are acceptable for a **medium** appetite with spike-gated performance assumptions. **Kill** or **needs-clarification** would mainly rest on rejecting spec-003 scope expansion or refusing to proceed without measured latency; stakeholder resolutions and concept boundaries address those without blocking specification.

## If go — Handoff to `/speckit-specify`

- **Problem**: Operators cannot continuously see what a running Supervisor browser instance is showing on the local machine without repeated discrete captures or external tools; monitoring long/background runs is awkward when logs and status are insufficient.
- **Chosen approach**: **Option B** — new local **watch** operator command; worker serves **loopback MJPEG** (CDP screencast → encode); CLI prints URL (**8080** default, next free port if taken); cadence **1/s** default, **--fps** up to **30**; observation ends on operator interrupt or when instance is no longer running; **`screenshot`** remains separate and may run concurrently; headless and headed both supported; sparse headless frames at low cadence accepted.
- **In scope / out of scope**: In scope: single local viewer, one instance per watch, viewport-sized stream (~1280×720 headless policy), lifecycle and integration tests on stand-in pages. Out of scope: remote/multi-viewer, CI artifacts, recordings gallery, replacing `screenshot`, auth on loopback, content redaction, multi-instance mosaic, new gameplay automation.
- **Success metrics**: Maintain watch **≥ 5 minutes** without per-glance manual capture; observation ends within **few seconds** of instance teardown; cadence **1–30/s**; still capture available while watching; record **p50/p95** latency and bytes/s in spike (planning bounds until measured).
- **Carried-forward open questions**:
  - Pre-implementation **spike** results: p95 frame latency, bytes/s, optional impact during entry automation (`problem.md`, `concept.md` assumptions).
  - **Per-instance port** allocation and printed URL when multiple instances run (stakeholder: 8080 with fallback—specify exact rules).
  - Screencast **frame ack** threading and shutdown when CLI disconnects vs worker continues.
  - Shared capture resources: scheduling between watch frames and `screenshot` without unacceptable stalls at default cadence.
  - Command naming and help text placement alongside existing `screenshot` / `logs` (not decided in assessment).