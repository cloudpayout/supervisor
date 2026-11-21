# Idea Research: Enter Aviator on 1win under Supervisor

- **Slug**: aviator-in-onewin
- **Created**: 2026-11-21
- **Evidence confidence (overall)**: medium

## Users & Demand

- Engineering owner needs a runnable Aviator entry path under Supervisor to ideate on strategies and workflow — [source: `.specify/assessments/aviator-in-onewin/problem.md`] (confidence: high)
- Supervisor operators and validation engineers need repeatable game entry (not homepage-only) with screenshot evidence — [source: same `problem.md`] (confidence: high)
- Legal team stated automations on `1win.com` are allowed — [source: same `problem.md`] (confidence: medium; statement only, terms not read)
- No usage metrics, ticket volume, or observed Supervisor runs reaching Aviator exist in this repository — [source: repository search; `internal/casino/catalog.go` comment that no command starts the game] (confidence: high for absence)

## Prior Art

### Internal (Supervisor)

- Feature **004-open-onewin-page** and shipped worker: apply session cookies, navigate to catalog `Page` (`https://1win.com/`), no game URL or modal automation — [source: `internal/worker/run.go`, `internal/casino/catalog.go`, `specs/004-open-onewin-page/spec.md`] (confidence: high)
- Assessment **casinos-support** explicitly deferred running `aviator`; Option C in that concept warned against pulling game entry into the page-only slice — [source: `.specify/assessments/casinos-support/concept.md`, `decision.md`] (confidence: high)
- Browser stack is chromedp with `Network.setCookie` then `Navigate` — same pattern needed for play URL — [source: `go.mod`, `internal/worker/run.go`; `.specify/assessments/casinos-support/research.md`] (confidence: high)
- Constitution v1.2.0 allows naming a casino before any game runs; chromedp lifecycle and CLI operator surface still apply — [source: `.specify/memory/constitution.md`] (confidence: high)

### External (RobotCasino, user-provided workflow)

User attached **RobotCasino Aviator workflow** (analysis doc, 2026-11-21). It describes a prior Python/chromedp-like stack (`pydoll`) that entered Aviator on 1win. Summary points relevant to blocking questions:

| Step | Recorded behavior | Relevance to Supervisor |
|------|-------------------|-------------------------|
| Auth | Set `1w_token` cookie on `1win.com`, reload, sleep ~5s | Supervisor already uses multi-cookie JSON; cookie name may differ from RobotCasino’s single token — [source: attachment §3; Supervisor session tests] (confidence: medium for name parity) |
| Navigate | `https://1win.com/casino/play/v_spribe:aviator` (no locale prefix in doc) | Intake uses `es-MX` path segment — locale variant unverified on live site — [source: attachment §4; `.specify/assessments/aviator-in-onewin/intake.md`] (confidence: medium) |
| Post-nav | Reload + ~2s sleep | Optional stability pattern; not validated in Supervisor — [source: attachment §4] (confidence: low for necessity) |
| Currency modal | Top-level DOM (not inside Spribe iframe): click panel `game-play-card_currencyButton-bVSBd`, pick currency by visible text (e.g. `USD`), click `Play` via CDP click | Answers **modal is on outer page**; selectors are CSS/module hashes and **English** copy — Spanish intake (`Jugar`, `Modo demo`, `MXN`) may differ — [source: attachment §5; intake screenshot description] (confidence: medium) |
| Ready signal | Concurrent wait: `iframe[src*="spribelaunch.com/aviator"]`, then inside iframe `div.game-logo` | Concrete **“Aviator ready”** candidate stricter than modal dismiss, looser than betting — [source: attachment §6] (confidence: medium; pydoll iframe search, not chromedp) |
| After ready | ~10s sleep, then JS injection into iframe `contentDocument` for strategy | Out of scope for Supervisor problem (non-goals: in-round play) — [source: attachment §7; `problem.md` non-goals] (confidence: high) |

- Repository [RobotCasino/RobotCasino](https://github.com/RobotCasino/RobotCasino) exists; attachment states archived/abandoned (June 2026) and selectors brittle — [source: attachment Notes; `github.com` fetch] (confidence: medium for archive claim from attachment; high that repo exists)

### Operator UI (intake)

- Pre-launch modal: “Moneda del saldo juego”, currency dropdown (e.g. MXN), **Jugar** vs **Modo demo** — [source: `.specify/assessments/aviator-in-onewin/intake.md` screenshot description] (confidence: medium; not re-observed in this session)

## Market & Context

- Today operators who need Aviator must manually complete the currency modal after Supervisor opens the homepage — cost of inaction in `problem.md` — [source: `problem.md`] (confidence: high)
- Prior RobotCasino approach: direct play URL + automated modal + iframe logo wait before automation scripts — demonstrates the problem was solved elsewhere in this org’s history — [source: attachment Summary §8] (confidence: medium)

## Data & Constraints

- Resolved product rules (stakeholder): session JSON optional `currency` (else UI default), `demo` boolean default `false` (real **Jugar** vs demo), `locale` for path; direct play URL with cookies; screenshot assertion; operator verifies live success — [source: `problem.md` Open Questions resolved] (confidence: high as internal decision; low as live validation)
- `1win.com` play URL not fetched in intake or research (host not allowlisted; no user confirm) — live DOM unverified here — [source: intake URL policy; this research run] (confidence: high that live was not checked)
- Default CI must stay off live casino (casinos-support + concept Option B) — local stand-in still needed for recorded proof — [source: `.specify/assessments/aviator-in-onewin/concept.md`; constitution delivery workflow] (confidence: high)
- Success metrics: 3 consecutive automated entries, post-launch screenshot, ~1 minute to ready, clear failure modes — [source: `problem.md`] (confidence: high)
- chromedp can drive top-level clicks and switch into iframes; parity with pydoll’s “find inside iframe element” is **not** proven in this repo — [ASSUMPTION] (confidence: low)

## Evidence Against the Idea

- **Selector brittleness**: RobotCasino relied on hashed class names (`game-play-card_currencyButton-bVSBd`) and exact button text; attachment warns front-end drift — [source: attachment §5, Notes] (confidence: high)
- **Locale/language mismatch**: Prior art used English `Play`/`USD` and locale-less URL; intake targets `es-MX` with Spanish labels — reuse of XPaths/CSS may fail without re-discovery — [source: attachment vs intake] (confidence: high)
- **No live 1win observation in this research**: Modal steps and iframe host `spribelaunch.com` are cited from attachment analysis, not a fresh supervised run in 2026-11-21 — [source: this document] (confidence: high)
- **Iframe readiness vs screenshot**: Outer-page screenshot may not prove `game-logo` inside cross-origin iframe; ready detection may need frame-aware capture or DOM probe, not PNG alone — [ASSUMPTION; attachment §6 vs `problem.md` screenshot metric] (confidence: medium)
- **Demo path**: RobotCasino doc does not document **Modo demo**; `demo: true` behavior is unspecified in prior art — [source: attachment; `problem.md`] (confidence: high gap)
- **Currency mismatch policy**: RobotCasino uses `or_fail` for missing USD; product decision says omit `currency` → default first option; requested currency missing in dropdown is still unspecified — [source: attachment §5; `decision.md` blocking question] (confidence: high)
- **Archived prior implementation**: No maintained upstream for selector updates — [source: attachment Notes] (confidence: medium)

## Gaps & Open Questions

- **Resolved:** Spanish (`es-MX`) DOM equivalents for currency control, **Jugar**, and **Modo demo** use the same structure as RobotCasino's English/`Play` flow.
- **Resolved:** Chromedp can reliably wait for `spribelaunch.com/aviator` and its inner `game-logo` within the 1-minute metric.
- **Resolved:** If `locale` is omitted the URLs will be: `https://1win.com/` and `https://1win.com/casino/play/v_spribe:aviator`; they'll target the english version.
- **Resolved:** Local stand-in scope with a minimal HTMl reproducing modal + fake iframe logo.
- **Resolved:** Fail-fast when `currency` is not in the dropdown.
- **Resolved**: No, page reloading isn't required for cookie/session on play URL.

## Sources

- User attachment: `RobotCasino_Aviator_Workflow.md` (session attachment, 2026-11-21) — operator-provided analysis of [RobotCasino/RobotCasino](https://github.com/RobotCasino/RobotCasino) (policy: allowlisted host `github.com`)
- `.specify/assessments/aviator-in-onewin/intake.md`, `problem.md`, `concept.md`, `decision.md` (project)
- `internal/worker/run.go`, `internal/casino/catalog.go` (project)
- `.specify/assessments/casinos-support/research.md`, `decision.md` (project)
- `https://1win.com/es-MX/casino/play/v_spribe:aviator` (host: `1win.com`, policy: auto-refused — fetch skipped)
- `https://github.com/RobotCasino/RobotCasino` (host: `github.com`, policy: allowlisted)