# Research: Enter Aviator on onewin

**Feature**: `007-enter-aviator-onewin`  
**Date**: 2026-11-21  
**Inputs**: [spec.md](./spec.md), assessment `aviator-in-onewin`, RobotCasino workflow attachment

## Decision: Play URL and locale

**Decision**: Register play URL templates in the catalog. Default (no `locale`): `https://1win.com/casino/play/v_spribe:aviator`. With `locale` (e.g. `es-MX`): `https://1win.com/{locale}/casino/play/v_spribe:aviator`. Cookie `URL` / navigation still use the effective play URL origin for `onewin` starts; homepage URL remains for catalog metadata only.

**Rationale**: Matches assessment and clarify session; RobotCasino used locale-less URL; intake used `es-MX`.

**Alternatives considered**: Always homepage then in-page navigation (rejected: spec FR-001 direct play URL); locale from browser only (rejected: session `locale` field).

## Decision: Session file optional fields

**Decision**: Extend parsed session document with optional `currency` (string, exactly three ASCII letters A–Z), `demo` (boolean), `locale` (string, BCP47-like segment used only as path prefix, e.g. `es-MX`). Omitted keys use defaults: first modal currency, `demo` false, English unprefixed URL. Unknown top-level fields still fail parse (`DisallowUnknownFields`).

**Rationale**: Spec FR-016; backward compatible.

**Alternatives considered**: Separate play-preferences file (rejected: CLI surface).

## Decision: Worker entry lifecycle (async)

**Decision**: `start` returns after registry + browser spawn as today. Worker applies cookies, navigates to play URL, then runs **entry** in the foreground of the worker process (not blocking `start` RPC). Registry field `entry` is `entering` until **Aviator ready**, then `ready`. On failure, worker tears down browser, removes profile, removes registry row. No new CLI failure reason (clarify session Q5).

**Rationale**: Spec FR-009, FR-009a, FR-010.

**Alternatives considered**: Block `start` until ready (rejected in clarify).

## Decision: Aviator ready probe (chromedp)

**Decision**: After modal automation, wait up to 60s total for entry (spec one minute). Ready when: (1) modal path complete, and (2) an iframe whose `src` contains `spribelaunch.com/aviator` exists, and (3) inside that frame, an element matching `div.game-logo` is visible. Implementation uses chromedp frame targeting (`chromedp.Run` with iframe node). Stand-in fake iframe uses the same `src` substring and inner `div.game-logo`.

**Rationale**: Clarify session FR-008; RobotCasino `wait_for_aviator_iframe_to_load` pattern.

**Alternatives considered**: Modal dismiss only (rejected); screenshot pixel match (rejected).

## Decision: Currency modal automation (live + stand-in)

**Decision**: Automate on **top document** only. Open currency control, select option whose visible text includes the three-letter `currency` code (case-sensitive match on code substring), then click play control: real play label (`Play`, `Jugar`) or demo (`Modo demo`) per `demo`. RobotCasino hints (English, hashed class `game-play-card_currencyButton-bVSBd`, CDP click on play) are **starting points** for live tuning; stand-in uses stable `data-supervisor-*` attributes for CI.

**Rationale**: Research + RobotCasino attachment; Spanish labels required per spec locale paths.

**Alternatives considered**: XPath-only from RobotCasino without stand-in attributes (rejected for tests).

## Decision: Concurrent modal and iframe wait

**Decision**: Run currency modal automation and iframe-ready wait **overlapping** where safe (RobotCasino `TaskGroup`). If modal must complete before iframe appears on live site, implementation may serialize; tests use stand-in that allows parallel wait.

**Rationale**: RobotCasino workflow reduced time-to-ready.

**Alternatives considered**: Strictly serial only (fallback if live requires).

## Decision: Ordinary test stand-in

**Decision**: Go `net/http` on `127.0.0.1` (pattern from `tests/integration/onewin_helpers_test.go`). New **aviator stand-in** page: initial HTML shows currency modal with buttons labeled `USD`, `MXN`, etc.; play/demo buttons; hidden iframe with `src` containing `spribelaunch.com/aviator` and inner `div.game-logo` revealed after correct flow. `SUPERVISOR_PAGE_URL` overrides play URL to stand-in base. Tests assert entry, ready pixel/marker, currency fail-fast, timeout teardown, secrets.

**Rationale**: FR-012; constitution off live casino.

**Alternatives considered**: Reuse balance-only stand-in (rejected: does not exercise modal/iframe).

## Decision: Fail-fast currency

**Decision**: If `currency` set and no modal option text contains that code, entry aborts, instance removed. Tests assert registry gone; no CLI category string.

**Rationale**: FR-005; SC-004.

## Decision: No post-cookie reload

**Decision**: Worker does not call `Page.reload` after setting cookies (spec FR-007). RobotCasino used reload; not adopted.

**Rationale**: Clarify / problem resolution.

## Decision: Registry `entry` field

**Decision**: Optional string on instance: `entering` | `ready`. Omitted for blank starts and legacy rows. `status` output includes phase when `entering` or `ready` for `onewin` (human-readable, no secrets).

**Rationale**: FR-009a without failure diagnostics.

**Alternatives considered**: Separate command (rejected).