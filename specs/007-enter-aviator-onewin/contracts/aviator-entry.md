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

- Game iframe in the top document, matched by any of:
  - `iframe[src*="spribelaunch.com/aviator"]` (legacy launch URL)
  - `iframe[src*="aviaport.spribegaming.com"]` or `spribegaming.com`
  - `iframe[src*="/aviator"]` or `aviator.html` (bundled shell on 1win)
  - Fallback: largest visible non-tracking iframe after play activation
- Inside that frame (or via chromedp frame context when cross-origin): visible ready marker — any of:
  - `div.game-logo` (legacy / stand-in `?frame=1`)
  - `.header__game-logo`, `app-game`, or `app-header` (live Aviaport shell)
  - Stand-in live shell: `?frame=live` under `[data-supervisor-game-frame]`

Success → registry `entry` = `ready`. Failure → teardown, row removed, no CLI reason string.

Worker log error tokens include `game_config` when visible page or game iframe text matches Spribe’s **retry limit of app/client config exceeded** message (headless misconfiguration or provider failure — not a successful ready state).

## Headless Chrome (casino embeds)

When `headless` is true, the worker allocator uses **headless=new** with a fixed desktop **`--user-agent`** (`DefaultHeadlessUserAgent`, overridable via `SUPERVISOR_HEADLESS_USER_AGENT`), drops `enable-automation`, sets `disable-blink-features=AutomationControlled`, enables a SwiftShader WebGL stack, and **does not** pass chromedp’s background-suppression flags, `disable-remote-fonts`, `disable-reading-from-canvas`, `safebrowsing-disable-auto-update`, or Puppeteer’s `enable-features` pair. **Ready** must not succeed on iframe geometry alone; config-error text is probed in the top document, same-origin frames, and Spribe CDP targets before `entry ready`.

## Status

While `entry` is `entering` or `ready`, instance appears in `status` with that phase for `onewin`. Missing instance after a successful-looking `start` indicates entry failure (FR-010).