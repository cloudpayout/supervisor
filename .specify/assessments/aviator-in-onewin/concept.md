# Concept: Enter Aviator after onewin start

- **Slug**: aviator-in-onewin
- **Created**: 2026-11-21
- **Revised**: 2026-11-21
- **Recommended option**: Option B — Session-driven play URL, modal automation, and ready probe

## Options

### Option A — Direct play URL only (operator finishes the modal)

- **Sketch**: `start` still applies cookies but navigates only to the Aviator play URL (locale-prefixed or English default paths from session JSON). The operator manually chooses currency and **Jugar** / **Modo demo**. Supervisor may screenshot whatever is on screen, without a defined **Aviator ready** probe.
- **Appetite**: small (days)
- **Trade-offs**: Fastest slice; matches “no homepage hop.” Fails repeatability, one-minute ready bound, and failure-clarity goals. Leaves validation engineers without automated entry proof.
- **Rabbit holes**: Declaring victory because the play URL loaded while the Spribe iframe never became playable.

### Option B — Session-driven play URL, modal automation, and ready probe

- **Sketch**: After cookies are set, the worker opens the play URL derived from optional `locale` (English unprefixed URLs when omitted; e.g. `es-MX/...` when set). It drives the **outer-page** currency modal: optional `currency` (else first/default in UI); **fail-fast** if a requested currency is absent from the dropdown; optional `demo` (default `false` → real play, `true` → demo). No reload step before navigation. It waits until **Aviator ready**—modal cleared and Spribe launch iframe present with an inner ready marker (e.g. game logo), within ~one minute, using a frame-aware DOM probe. Operator confirms on live machines; `screenshot` remains the human evidence command. Default CI uses a **minimal local HTML** stand-in that reproduces modal + fake iframe-ready marker, not live `1win.com`.
- **Appetite**: medium (weeks) — mostly UI automation, ready detection, session shape, and stand-in; cookie navigation is already proven for the homepage.
- **Trade-offs**: Meets repeatability, time bound, fail-fast currency, and local-CI constraints from `problem.md`. Prior RobotCasino art supports outer-page modal + concurrent iframe wait pattern; Spanish copy and hashed selectors remain brittle. Ready is judged by probe, not PNG pixels alone; screenshot supplements operator review.
- **Rabbit holes**: In-round betting or script injection into Spribe. Live play URL in default `go test`. Hard-coding one locale. Silent fallback when currency missing. Proving financial outcomes. Porting entire RobotCasino JS stack.

### Option C — Do nothing (homepage-only)

- **Sketch**: Keep catalog `Game: aviator` and worker navigation to `https://1win.com/` only; operators document manual Aviator entry.
- **Appetite**: none
- **Trade-offs**: Zero automation risk; all problem goals and metrics stay unmet; engineering owner remains blocked on supervised game entry.
- **Rabbit holes**: Treating naming as running.

## Recommendation

**Option B.** Problem and research align on direct play URL, session JSON (`currency`, `demo`, `locale`), fail-fast currency, no reload, chromedp iframe ready within one minute, demo path, and minimal HTML stand-in for ordinary tests. Option B is the only option that satisfies repeatability and failure-clarity without pulling gameplay into scope. Option A is a fallback if appetite is cut to days; Option C only if policy withdraws live automation.

Evidence is **medium** overall: RobotCasino documents the flow; live `1win.com` was not fetched in research. Specification must rediscover selectors for English and `es-MX` and nail the ready probe contract (iframe-inner marker + operator `screenshot`).

## Out of Scope (for the recommended option)

- Bets, cash-out, auto-bet, and injected strategy scripts inside the game.
- Games other than Aviator; locales beyond what `locale` in session JSON selects.
- Login, deposits, KYC, cookie harvesting inside Supervisor.
- Live `1win.com` entry in the default CI job.
- Storing secrets, live session captures, or gameplay screenshots in the repo.
- Verifying RTP, balances, or winnings.

## Assumptions to Validate

- Extended session JSON (`currency`, `demo`, `locale`) remains backward-compatible when keys are omitted.
- Play URL templates: `https://1win.com/casino/play/v_spribe:aviator` and `https://1win.com/{locale}/casino/play/v_spribe:aviator`.
- Outer-page modal controls are automatable under chromedp for both English and Spanish labels.
- Spribe iframe `spribelaunch.com/aviator` and an inner ready element are detectable within the one-minute budget.
- Minimal HTML stand-in is sufficient for integration tests covering modal + ready probe without FastAPI parity to the full casino page.
- Selector drift on `1win.com` is maintainable at medium appetite; otherwise scope collapses toward Option A.