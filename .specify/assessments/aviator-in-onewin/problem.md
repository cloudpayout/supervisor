# Problem Definition: Reaching Aviator in-game state under Supervisor (onewin)

- **Slug**: aviator-in-onewin
- **Created**: 2026-11-16
- **Revised**: 2026-11-21
- **Inputs used**: intake.md | research.md

## Problem Statement

Operators who run Supervisor with an authenticated **onewin** session can reach the casino homepage but cannot reliably finish **entering** Aviator: the play URL, the game-balance currency choice, and activation of real play or demo must still be done by hand. The product already names `aviator` in the catalog while the worker only opens `https://1win.com/`, so supervised runs, screenshots, and integration checks cannot consistently show that Spribe Aviator is loaded and playable—not merely that the site opened.

## Affected Users & Stakeholders

- **Users**: **Supervisor operators** — session files and `start` / `screenshot` / `stop` should leave the browser in an Aviator-ready state without manual modal clicks in the normal case.
- **Users**: **Test or validation engineers** — need a repeatable “game is actually open” outcome and recorded checks that stay off the live casino in default CI (per prior casinos-support constraints in research).
- **Stakeholders**: **Engineering owner** — needs Aviator entry under Supervisor to explore strategies and workflow; demand is stated, not measured by usage data (research).
- **Stakeholders**: **Legal team** — stated that automations on `1win.com` are allowed (not independently verified against site terms in research).

## Goals

- After `start` with a valid onewin session, the instance reaches an agreed **Aviator ready** observable state (past the currency modal and with the Spribe client loaded) without operator clicks in the normal case.
- Operators can tell success from stall (modal still open, session invalid, currency missing from dropdown, timeout, wrong locale).
- `screenshot` and operator review support evidence that entry succeeded; live verification remains with the operator.
- Default automated proof continues to use a **local stand-in** (minimal HTML reproducing modal + fake ready state), not the live play URL on every CI run.

## Non-Goals

- In-round Aviator actions (bets, cash-out, auto-bet) or script injection into the game.
- Every 1win game or every market; Aviator entry only.
- Login, deposits, KYC, or obtaining session cookies inside Supervisor.
- Proving balances, RTP, or financial results.
- Mandatory live `1win.com` entry in the default `go test` job.

## Success Metrics

- **Repeatability**: 3 consecutive supervised runs with a known-good session reach **Aviator ready** without manual browser steps (baseline: none — research records 004/homepage-only behavior today).
- **Evidence**: After successful entry, capture shows post-launch game context—not homepage-only or currency-only modal—in 100% of successful runs (baseline: homepage norm). Research notes that outer PNG alone may not prove iframe-internal ready; ready detection may need an observable agreed at specify time.
- **Time-to-ready**: From `start` exit to **Aviator ready** ≤ 1 minute under typical conditions (baseline: unbounded / manual). Research cites prior art using iframe + inner `game-logo` as a ready candidate; live timing not re-measured in research.
- **Failure clarity**: Distinguish session, currency not in dropdown (fail-fast when configured currency missing), UI/locale mismatch, and timeout (baseline: operator discovers manually).

## Cost of Inaction

Operators keep validating Aviator with manual modal work after every supervised start. Screenshots remain weak evidence of “game running.” Strategy and workflow work stays blocked on a hand-driven entry path, and the gap left by casinos-support (“name `aviator`, do not run it”) persists.

## Open Questions

- **Resolved (product)**: Session JSON optional keys — `currency` (else first/default in modal), `demo` (default `false` → real play), `locale` (`locale/path` format). Direct navigation to play URL with cookies; no extra homepage hop. Operator confirms live success.
- **Resolved (product)**: When `locale` is omitted, use non-prefixed English URLs (`https://1win.com/` and `https://1win.com/casino/play/v_spribe:aviator`). When set (e.g. `es-MX`), use locale-prefixed paths as in intake.
- **Resolved (product)**: Missing `currency` in dropdown → fail-fast (not silent fallback).
- **Resolved (product)**: No page reload required after cookies before play URL navigation.
- **Resolved (stakeholder)**: Spanish `es-MX` UI (**Jugar** / **Modo demo**) treated as same structural flow as prior English/`Play` art; selectors must be rediscovered for specify.
- **Resolved (stakeholder)**: chromedp can wait for Spribe launch iframe and inner ready signal within the one-minute bound; **Modo demo** path is in scope when `demo: true` even though prior art only documented real **Play**.
- **Resolved (stakeholder)**: Local stand-in = minimal HTML for modal + fake iframe-ready marker.
- **Resolved (product)**: for success evidence, it should be probe by looking the iframe or an iframe-inner element, and the operator can verify the success by using the command `screenshot`. [NEEDS CLARIFICATION: Whether success evidence is outer `screenshot` only, an internal ready probe, or both — research flags iframe-inner ready may not appear in a top-level PNG.]
- **Resolved (stakeholder)**: Oscar Cedano <oscar.cedano@outlook.com> Product Engineer [NEEDS CLARIFICATION: Who raised the idea by name/role — intake still marks raiser unknown.]