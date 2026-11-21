# Idea Intake: Aviator game on 1win (onewin)

- **Slug**: aviator-in-onewin
- **Created**: 2026-11-16
- **Source**: pasted text; codebase: `internal/casino/catalog.go`, supervisor CLI (`start` / worker); URL (fetch skipped — see Idea section)
- **Type**: new-capability

## Idea (as captured)

> Add a new game to onewin named `aviator` whose URL is https://1win.com/es-MX/casino/play/v_spribe:aviator. It should open the game and select the currency and click the play button to the game.

**URL (sanitized)**: `https://1win.com/es-MX/casino/play/v_spribe:aviator`  
**Host**: `1win.com`  
**URL Trust Policy**: `[UNVERIFIED — fetch skipped: host not on safe list: 1win.com]`

**Related screenshot (operator-provided)**: Pre-launch modal on the Aviator play URL — label “Moneda del saldo juego” with currency dropdown (e.g. MXN), primary **“Jugar”** (Play) and secondary **“Modo demo”** (Demo mode).

**Repo context (read-only)**: The supervisor already registers casino `onewin` with `Page: https://1win.com/` and `Game: "aviator"` in `internal/casino/catalog.go`, with a comment that no command starts the game today. Prior spec **004-open-onewin-page** scoped opening the casino main page with a session, not navigating into a specific game or in-page play flow.

## Restated

Extend onewin support so automation opens the Spribe Aviator game at the given 1win play URL, chooses the game balance currency, and activates **Jugar** to enter the game—not only the casino homepage.

## Origin & Context

- **Raised by**: [NEEDS CLARIFICATION: stakeholder name or role]
- **Trigger**: User request with target URL and UI flow (currency + play); aligns with existing `onewin` / `aviator` catalog naming but not current launch behavior.

## First-Glance Unknowns

- [NEEDS CLARIFICATION: Which currency should be selected by default or from configuration (e.g. MXN vs operator choice)?]
- [NEEDS CLARIFICATION: Should demo mode (“Modo demo”) ever be used instead of real-money **Jugar**?]
- [NEEDS CLARIFICATION: Locale/path — is `es-MX` fixed or should locale follow session/account?]
- [NEEDS CLARIFICATION: Success criteria — loaded iframe, post-click game UI, screenshot assertion, or operator-only confirmation?]
- [NEEDS CLARIFICATION: Relationship to existing `start` workflow — new flags, automatic navigation after main page, or replace default `Page` URL?]
- [NEEDS CLARIFICATION: Legal/compliance and terms-of-use constraints for automating play on the live 1win site vs test stand-ins already used in integration tests.]