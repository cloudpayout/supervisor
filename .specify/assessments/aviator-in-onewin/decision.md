# Decision: Enter Aviator after onewin start

- **Slug**: aviator-in-onewin
- **Decided**: 2026-11-21
- **Verdict**: go
- **Artifacts reviewed**: intake.md | research.md | problem.md | concept.md

## Scorecard

| Criterion | Rating | Justification |
|-----------|--------|---------------|
| Problem validity | strong | Worker opens homepage only while catalog names `aviator`; operators cannot reach a repeatable **Aviator ready** state under Supervisor (`problem.md`, shipped `004`). |
| Evidence strength | adequate | `research.md` overall confidence **medium**: RobotCasino workflow documents play URL, outer-page modal, and iframe-ready signal; codebase proves cookie+navigate; stakeholder resolutions on session JSON, locale URLs, fail-fast currency, and stand-in. Live `1win.com` not fetched—same class of gap as casinos-support at **go**. |
| Value vs. inaction | strong | Manual modal every run; screenshots do not prove game loaded; engineering owner blocked on supervised entry (`problem.md` cost of inaction). |
| Feasibility / appetite | adequate | Option B scoped to **medium (weeks)** with homepage navigation already shipped; rabbit holes bounded (no in-round play, no live default CI). |
| Strategic fit | adequate | chromedp CLI lifecycle, session files, casinos-support handoff (local stand-in + live operator validation); closes deferred “name aviator, don’t run” gap. |
| Risk posture | adequate | Selector brittleness and locale/copy drift documented in research; mitigations include fail-fast currency, frame-aware ready probe, minimal HTML for CI, operator `screenshot` on live—not assumed solved in code yet. |

## Verdict & Rationale

**Go** on **Option B** (session-driven play URL, modal automation, ready probe). Earlier **needs-clarification** rested on missing `research.md`; that artifact now grounds modal-on-outer-page, iframe ready candidate, and product rules. Problem validity and value are **strong**; evidence is **adequate** (not **strong** because live DOM was not re-observed in this repo). Feasibility and strategic fit support specification inside medium appetite. Risk stays bounded by explicit out-of-scope and stand-in pattern—not a reason to stall at assessment.

## If go — Handoff to `/speckit-specify`

- **Problem**: Supervised `onewin` does not finish entering Aviator (play URL, currency modal, play/demo); runs cannot reliably reach **Aviator ready** or prove it.
- **Chosen approach**: **Option B** — cookies, direct play URL from optional `locale`, automate outer-page modal (`currency`, `demo`, fail-fast if currency missing), frame-aware ready probe (Spribe iframe + inner marker) within ~1 minute, operator verification via `screenshot`; minimal local HTML stand-in for default tests.
- **In scope / out of scope**: In scope: Aviator entry only; extended session JSON keys; ready probe + clear failures; local stand-in. Out of scope: bets/cash-out/script injection; other games; login/deposits; live `1win.com` in default `go test`; storing secrets or live captures in repo.
- **Success metrics**: 3 consecutive automated entries without manual steps; ready ≤ 1 minute; post-modal evidence (probe + operator screenshot); fail-fast on missing currency; distinguish session/timeout/UI failures.
- **Carried-forward open questions**:
  - Rediscover DOM selectors for English and `es-MX` (**Jugar** / **Modo demo**); prior art used English `Play` and hashed classes.
  - Specify exact **Aviator ready** probe contract (iframe URL pattern + inner element) and how worker signals failure to the operator.
  - Confirm minimal HTML stand-in suffices for integration coverage without full FastAPI casino page.
  - Legal automation stated but site terms not read in research—do not treat as compliance sign-off in spec.