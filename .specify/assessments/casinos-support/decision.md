# Decision: A onewin page under Supervisor

- **Slug**: casinos-support
- **Decided**: 2026-11-15
- **Verdict**: go
- **Artifacts reviewed**: intake.md, research.md, problem.md, concept.md

## Scorecard

| Criterion | Rating | Justification |
|-----------|--------|---------------|
| Problem validity | adequate | `start` opens only a blank page, so the one named operator cannot put an `onewin` session on a page under Supervisor. `aviator` is named and its behavior is deferred. No other demand was found. |
| Evidence strength | adequate | Research confidence is medium. The blank page, the cookie-then-navigate call on a local server, the ban on committed secrets, and the rule that the ordinary run stays off a live casino are all in the notes. The live `1win.com` result is a non-goal, closed by refusal. The FastAPI server is the raiser's choice of local stand-in, not a measured result. |
| Value vs. inaction | adequate | Doing nothing leaves that operator on a blank page and leaves the later `aviator` assessment with no page to attach to. |
| Feasibility / appetite | adequate | Option B fits a days budget for a cookie file, the `onewin` page, profile removal, and a local check. A rejected live cookie is outside that budget and stops the work. Python install time on the default job was not measured. |
| Strategic fit | adequate | `onewin` and `aviator` use the lowercase identifier form. The ordinary check covers `start`, `status`, and `stop` and stays off a live casino. Secrets stay out of the repository. Principle V also says every casino has games that can be run, and this slice does not run `aviator`. |
| Risk posture | weak | The local check, the file path, and the refusal to store a cookie bound the recorded proof. Profile deletion while Chrome may still be writing, and a named game that does not run, are identified and not yet mitigated. |

## Verdict & Rationale

Go, on Option B. The earlier block was the claim of an authenticated live page. The problem and the concept now ask for the outcome the notes support: Supervisor applies the operator's session and opens the `onewin` page, the operator judges the live result privately, and the recorded proof is the local check. Problem validity and evidence strength are both adequate, and a recommended option exists. Evidence is not strong: nobody ran the FastAPI check, and nobody observed `1win.com`. Feasibility is adequate inside the days budget because a rejected live cookie ends the slice instead of expanding it. Risk posture stays weak and is carried forward. It does not drop the verdict, because the serious unknowns are bounds on specification, not a missing statement of the problem.

## If go — Handoff to `/speckit-specify`

- **Problem**: The operator cannot put an `onewin` session on a page under Supervisor, because `start` opens a blank page.
- **Chosen approach**: Option B — Open onewin with the supplied session. `start` reads a cookie file, opens the `onewin` page, the operator judges that page privately, the profile is deleted with the browser, `aviator` is named and not run, and the ordinary test is a local FastAPI server on `127.0.0.1` followed by `status` and `stop`.
- **In scope / out of scope**: In scope: one casino, `onewin`; the session file; the page open under Supervisor; secrets kept out of process arguments, worker logs, the instance registry, and the repository; profile removal when the instance ends; a local check of `start`, `status`, and `stop`. Out of scope: a verified `1win.com` balance, how `aviator` works, any run of that game, Spaceman, Dino, other casinos, a sign-in step, a stored cookie value or live session, and a live casino in the ordinary test.
- **Success metrics**: A run opens the `onewin` page instead of a blank page. The operator judges that page, and the judgment is not stored. No cookie value is stored. The ordinary confirmation does not require a recorded live login.
- **Carried-forward open questions**:
  - Principle V says every casino has games that can be run. This slice names `aviator` and does not run it. Specification must not reinterpret that MUST. If the slice cannot pass review on that sentence, it stops for a constitution change or the later game assessment.
  - Whether the profile directory can be deleted as soon as the browser instance ends, including while Chrome may still be writing it.
  - The operator's private signal `.adaptive-balance_number-GRGZd` was named and not observed. It is not a stored result and not a live-site test.
  - Automated sessions are the raiser's statement. The site's terms were not read.
