# Problem Definition: Observing a running browser instance over time

- **Slug**: real-time-streaming
- **Created**: 2026-11-25
- **Inputs used**: intake.md | research.md

## Problem Statement

Supervisor operators who supervise long-running or background browser instances cannot **continuously see** what the browser is showing through the product. They must rely on **discrete** visual checks, external tools, or a visible headed window—while **logs** and **status** often do not explain stalls or failures during monitoring (for example background game entry). That gap makes it hard to watch a run unfold in real time on the same machine where `start` was issued.

## Affected Users & Stakeholders

- **Users**: **Supervisor operators** (e.g. Oscar Cedano, trigger **monitoring**) — need sustained visual awareness of one running instance without issuing a separate capture command for every glance, for both headless and headed runs.
- **Stakeholders**: **Product / engineering owner** — stated the need once; research found no corroborating tickets or usage metrics in this repository (confidence: high for absence of other signals).

## Goals

- An operator on the **local machine** can **watch** what a **single** running instance is displaying for as long as the instance stays running, without a manual capture command on each update.
- **Headless** and **headed** runs are both in scope for the same monitoring need.
- **On-demand still captures** for evidence (today’s screenshot workflow) remain available and are not replaced by continuous observation.
- When the instance is **no longer running**, continuous observation **stops promptly** rather than leaving a stale or hanging view.
- The operator can choose how often the view updates, within an agreed **upper bound**; a **low default cadence** is acceptable even when the page changes rarely (including sparse updates on static headless views).

## Non-Goals

- Remote viewers, multiple simultaneous viewers, or streaming as a CI artifact.
- Replacing **screenshot** as the documented evidence command for success/failure review.
- Hiding or redacting sensitive page content on the local observation path (stakeholder accepts local exposure).
- Gallery, recording library, or multi-instance mosaic views (prior screenshot feature scope excluded video).
- In-browser or in-game automation beyond what existing worker flows already perform.

## Success Metrics

- **Continuous observation**: For a running test instance, an operator can maintain visual monitoring for **≥ 5 minutes** without entering a per-frame manual capture loop (baseline: repeated discrete captures or headed window only).
- **End of run**: When the instance leaves the running state (operator **stop**, crash, or supervised teardown such as failed background entry), observation **ends within a few seconds** (baseline: operator discovers via **status** polling or missing instance).
- **Cadence control**: Operator-selectable update rate from a **low default** (stakeholder: **1/s**) up to a **maximum of 30/s** (baseline: no product-supported cadence; ad-hoc shell loops only).
- **Parallel evidence**: While observation is active, the operator can still obtain a **one-off still capture** when needed (baseline: only discrete capture exists today).
- **Performance accountability**: Before or during specification, record **p50/p95** visual-update latency and approximate bytes per second at **1280×720** on headless stand-in pages, plus optional impact while entry automation runs (baseline: unmeasured in research; plan on ≤10s hard cap per update family from existing capture budgets until measured).

## Cost of Inaction

Operators keep polling **status**, tailing **logs**, running **screenshot** in a loop, or staring at a headed Chrome window during long or opaque phases (research cites background Aviator entry and thin CLI failure signals). Monitoring stays awkward, easy to miss short-lived UI states, and dependent on practices outside the Supervisor CLI contract.

## Open Questions

- **Resolved (stakeholder)**: Local operator only; single viewer; no authentication requirement for local observation; sensitive page content may appear on the stream.
- **Resolved (stakeholder)**: Default observation cadence **1/s**; configurable up to **30/s**; sparse frames on static headless pages at low cadence are acceptable for monitoring.
- **Resolved (stakeholder)**: Continuous observation and on-demand still capture **may run together** (not mutually exclusive).
- **Resolved (stakeholder)**: Observation stops **immediately** when the instance is no longer running (no indefinite hang).
- **Resolved (research, pre-specify)**: Hard latency and bandwidth numbers come from a **pre-implementation spike** on stand-in pages; until then use conservative planning bounds documented in research.
- **Resolved (stakeholder)**: The adress will be: `http://localhost:{port} and the port by default will be **8080** if it is not taken by another instance/program, in that case, another port will be taken.
- **Resolved (stakeholder)**: Yes, we the intake autor is sufficient to prioritize this over other Supervisor work.