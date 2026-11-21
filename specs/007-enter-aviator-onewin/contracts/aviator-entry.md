# Contract: Aviator entry and ready probe

**Feature**: `007-enter-aviator-onewin`

## Navigation target

After cookies, worker navigates to play URL from catalog + effective `locale`, unless `SUPERVISOR_PAGE_URL` is set (tests).

## Modal automation (top document)

1. Open currency selector (live: discover control; stand-in: `[data-supervisor-currency-open]`).
2. If `currency` set: click option containing that three-letter code; if none, fail entry (`currency`).
3. If `currency` omitted: click first/default option (stand-in: first `[data-supervisor-currency]`).
4. If `demo` effective **false**: activate real play (labels include `Play` or `Jugar`; stand-in: `[data-supervisor-play]`).
5. If `demo` effective **true**: activate demo (labels include `Modo demo` or `Demo`; stand-in: `[data-supervisor-demo]`).

No full-page reload after cookies (FR-007).

## Aviator ready

Within **60s** from start of entry automation:

- Launch iframe: `iframe[src*="spribelaunch.com/aviator"]` in top document.
- Inside that frame: visible `div.game-logo` (stand-in: same structure under `[data-supervisor-game-frame]`).

Success → registry `entry` = `ready`. Failure → teardown, row removed, no CLI reason string.

## Status

While `entry` is `entering` or `ready`, instance appears in `status` with that phase for `onewin`. Missing instance after a successful-looking `start` indicates entry failure (FR-010).