# Problem Definition: A onewin page under Supervisor

- **Slug**: casinos-support
- **Created**: 2026-11-15
- **Inputs used**: intake.md, research.md

## Problem Statement

The operator who runs Supervisor cannot put a `onewin` session on a casino page. Every run opens a blank page, so a session the operator already holds never becomes a page under Supervisor. The casino's named game is `aviator`. How that game works is a later assessment. This problem stops when the casino page is open under Supervisor and the operator can judge it on their own machine.

## Affected Users & Stakeholders

- **Users**: Oscar Cedano, the only person who stated the request. He cannot put his `onewin` session on a page under Supervisor. No other demand signal was found.
- **Stakeholders**: Oscar Cedano. He set the bounds: one casino, `onewin`; one game name, `aviator`, with the game's behavior deferred; no cookie value and no live session stored in this assessment; he judges the live page himself. His statement that automated sessions are allowed was not checked against the site's terms.
- **Stakeholders**: The constitution. It requires an explicitly supported casino and game, session cookies only at runtime and never committed, and automated coverage of start, status, and stop before that behavior is complete. The ordinary run stays off a live casino.

## Goals

- Supervisor applies the session the operator already holds and opens the `onewin` page. Today that run opens a blank page.
- The operator judges the live result on their own machine. That judgment is not a record this assessment stores.
- `aviator` is the supported game for that casino. Playing it is not required for this problem to be solved.
- Session cookie values are not committed, and they do not appear in process arguments, worker logs, or the instance registry.
- When the run ends, that run's session material does not remain in the browser profile. Today the profile directory stays on disk.

## Non-Goals

- A verified claim that `1win.com` showed a balance. Research did not observe that page, and no live session will be recorded here.
- How `aviator` works, and any later game behavior.
- Spaceman, Dino, and any casino other than `onewin`.
- A cookie value, or a record of a live `1win.com` session, stored in this assessment or the repository.
- A live-casino check inside the ordinary test run.

## Success Metrics

- A run opens the `onewin` page under Supervisor instead of a blank page. (baseline: a blank page; zero casino pages in the product)
- The operator judges that live page on their own machine. The judgment is not stored. (baseline: no casino page to judge; qualitative)
- No cookie value is stored in the assessment or the repository. (baseline: none stored)
- The ordinary confirmation of this behavior does not require a recorded live login. (baseline: no such confirmation exists)

## Cost of Inaction

Supervisor keeps opening only a blank page. The operator still cannot use an `onewin` session inside it, and still cannot judge that page under Supervisor. `aviator` stays unreachable, including the later assessment that would define how that game works.

## Open Questions

- None. The live-page result stays the operator's private judgment. Research closed that question by refusal, and named `aviator` as the only game, with its behavior deferred.
