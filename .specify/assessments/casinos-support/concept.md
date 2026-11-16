# Concept: A onewin page under Supervisor

- **Slug**: casinos-support
- **Created**: 2026-11-15
- **Recommended option**: Option B — Open onewin with the supplied session

## Options

### Option A — Leave start on a blank page

- **Sketch**: Supervisor stays as it is. `start` still opens a blank page. The operator keeps using a normal browser for an `onewin` session and judges that page there. `aviator` stays unnamed in the product.
- **Appetite**: small
- **Trade-offs**: No new secret handling and no new test dependency. It misses every goal. The operator still has no `onewin` page under Supervisor to judge, and the later `aviator` assessment has no page to attach to.
- **Rabbit holes**: None in the product. How often the operator already opens the site outside Supervisor was not measured.

### Option B — Open onewin with the supplied session

- **Sketch**: The operator points `start` at a file of session cookies and gets one supervised browser on `onewin`'s page. They judge that page on their own machine. No sign-in step is added. No cookie value and no live session are stored in the repository. Cookie values stay out of process arguments, worker logs, and the instance registry. When the browser instance ends, its profile directory is deleted with it. `onewin` is the only casino. `aviator` is the named game, and this option does not run it. The ordinary test starts a local FastAPI server on `127.0.0.1`. That server shows a balance only when the test cookie for that host is present. The check runs `start`, then `status` and `stop`, and does not open the live casino. The default checks job installs Python and FastAPI for that server.
- **Appetite**: small (days). Uncertain: the live page was not observed, so a rejected real cookie can turn those days into site trial and error that this budget does not include. The time to add Python and FastAPI to the default job was not measured.
- **Trade-offs**: This matches the restated goal. Supervisor applies the session and opens the casino page. The operator judges the live result, and that judgment is not stored. The local check is the recorded proof that a session cookie was applied and that `start`, `status`, and `stop` finish, without a live casino. It adds Python and FastAPI to a job that today installs only Go and Chrome. A local balance does not show what `1win.com` did. Naming `aviator` without running it leaves Principle V's phrase "games that can be run" for specification to square with a game that does not run yet.
- **Rabbit holes**: Treating the operator's private look as a failed test and then storing a cookie value or a live-session record to "prove" the page. Deleting the profile while Chrome is still writing it. Letting the balance selector `.adaptive-balance_number-GRGZd` become a live-site check. Starting `aviator` in this slice because the name is now on the casino.

### Option C — Open the page and run aviator

- **Sketch**: Same session and page as Option B, and the same local check. The difference the operator sees is that this slice also runs `aviator`. How the game works would be decided here instead of in a later assessment.
- **Appetite**: medium (weeks). Uncertain, and the evidence cannot pin this tighter. The page work is the same days as Option B. The weeks are the budget for a game nobody has specified. Research says that behavior is a later assessment.
- **Trade-offs**: This is the option that makes `aviator` runnable, which is closer to the README and to "every casino has games that can be run." It breaks the problem's bound that this slice stops when the page is open and that the game's behavior is deferred. It also puts a real game near the ordinary test run, which the constitution says must stay off a live casino and off a check that drives a real game.
- **Rabbit holes**: Defining `aviator` while the page work is the slice that was asked for. A game check that fails closed unless the live site is in the default run. Spaceman or Dino riding along because the README names them together.

## Recommendation

Option B. The goals are a `onewin` page under Supervisor instead of a blank page, a private judgment of that page, `aviator` named and not played, session values kept out of the repository and out of process arguments, logs, and the registry, and the profile gone when the run ends. The recorded proof is the local check. Option A leaves the blank page. Option C pulls the deferred game into this slice.

Demand is still one named person. The recommendation stays inside a small appetite only while opening the page with the supplied file is the work. If that file is rejected by the live site, the work stops for a new decision rather than storing a session or growing `aviator`.

## Out of Scope (for the recommended option)

- A verified claim that `1win.com` showed a balance.
- How `aviator` works, and any run of that game.
- Spaceman, Dino, and any casino other than `onewin`.
- A sign-in step. The session comes from cookies the operator already holds.
- A cookie value, or a record of a live `1win.com` session, stored in the repository.
- A live-casino check inside the ordinary test run. The ordinary check is the local FastAPI server.
- Passing the session document itself as a process argument.
- Leaving the browser profile on disk after the instance ends.

## Assumptions to Validate

- The operator can judge the live `onewin` page on their own machine. That judgment is not recorded. The selector `.adaptive-balance_number-GRGZd` was named and was not observed.
- A file of the cookies the operator already holds is enough to open the page. No login step is added.
- The profile directory can be deleted as soon as the browser instance ends, without leaving the session behind and without deleting it while the browser still needs it.
- The ordinary test can start a FastAPI server, send one cookie for `127.0.0.1`, read a balance, and then run `status` and `stop`. The default checks job grows a Python and FastAPI install. This check says nothing about `1win.com`.
- `onewin`'s page address is known to the product rather than typed on each start.
- Automated sessions on this casino are allowed. That is the raiser's statement. The site's terms were not read.
- Naming `aviator` without running it is enough for this slice. Principle V says every casino has games that can be run, and this concept does not run one.
