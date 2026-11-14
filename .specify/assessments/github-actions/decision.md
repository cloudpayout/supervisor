# Decision: Record a test result on each change

- **Slug**: github-actions
- **Decided**: 2026-11-12
- **Verdict**: go
- **Artifacts reviewed**: intake.md, research.md, problem.md, concept.md

## Scorecard

| Criterion | Rating | Justification |
|-----------|--------|---------------|
| Problem validity | adequate | A push or pull request has no recorded test result (no `.github` history; public `main` workflows path returned HTTP 404), and a missing browser turns four default browser tests into skips that still pass. Only one person asked, and there is no history of broken builds. |
| Evidence strength | adequate | Research confidence is medium. The missing result, the skip behavior, Go 1.27.1 in `go.mod` and in the `actions/go-versions` manifest, and the constitution's test rules are cited. Duration, flake rate, and billing are assumptions or unmeasured. |
| Value vs. inaction | adequate | Doing nothing leaves the baseline at "no result" and still lets a Chrome-less run look successful. That is the cost Oscar Cedano named. How often people already run the suite locally was not measured. |
| Feasibility / appetite | unknown | A single hosted Go test check has public prior art, and Go 1.27.1 is listed stable in the fetched manifest. The recommended option also requires the suite to pass more than once, and that flake rate was never measured, so the weeks budget cannot be confirmed. |
| Strategic fit | strong | The constitution requires automated coverage of start, status, and stop, Go 1.27.1, no committed secrets, and default tests that do not touch a live casino. The recommended option runs that default set on Linux and adds no product behavior. |
| Risk posture | weak | Skips-as-passes, fixed 300ms and 500ms sleeps, fork permissions, and the stale intake sentences are identified. None of them is mitigated yet. The concept only sets a revisit if repeats need a broader test rewrite. |

## Verdict & Rationale

Go. Problem validity and evidence strength are both adequate, which clears the bar for specifying, and concept.md recommends Option C rather than an unshaped idea. Value beats inaction for the only named decision maker: a change should show whether the default tests passed, with browser tests actually run, and a failing or blocked check should hold the pull request. Strategic fit is strong. Feasibility is unknown and risk posture is weak; those scores are real. They do not drop the verdict to needs-clarification, because the unknowns are bounded assumptions for specification (how many repeats count as stable, whether the suite already agrees with itself) and the concept already says to revisit this decision if stability needs a broader test rewrite. They are not a reason to keep the repository without a test result.

## If go — Handoff to `/speckit-specify`

- **Problem**: People who change `github.com/cloudpayout/supervisor` cannot tell whether the Go tests succeeded after a push or a pull request, and a run without Chrome can report success while skipping the browser tests.
- **Chosen approach**: Option C — Repeat the check before it counts. On every push and pull request, any branch, including forks, run the default `go test ./...` set on `ubuntu-latest` with Go 1.27.1 and a browser present. Success means that same change passed on more than one execution and none of the four default browser tests were skipped. A failure, a skip, or a check that cannot run blocks the pull request.
- **In scope / out of scope**: In scope is that repeated default-suite result, test changes only where repeats disagree or a browser test would skip, and holding the pull request on failure or a blocked check. Out of scope: the `slow` ten-instance test, live casino or other external game checks, macOS and Windows, new supervisor behavior, secrets, a billing analysis, a self-hosted runner, and rewriting the product to make the suite faster.
- **Success metrics**: Every push and pull request on any branch has a recorded pass/fail for the default set (baseline: none). The four default `requireChrome` tests finish as pass or fail, not skip (baseline: they skip when Chrome is missing). The result uses Go 1.27.1 (baseline: unknown). The author can answer whether the tests succeeded from the change itself (baseline: they cannot).
- **Carried-forward open questions**:
  - How many repeated executions of the same change count as stable? The problem note does not give a number.
  - `problem.md` says the "no GitHub-hosted runner" sentence was removed and that `ubuntu-latest` is the choice. `intake.md` still contains that sentence and still limits the trigger to `main`. Specification follows the problem note.
  - Whether `ubuntu-latest` is a Linux environment that meets the v1 Linux/WSL2 limit. Runner docs were not fetched.
  - Whether fork pull requests can run the check with no secrets and no write access to the repository.
  - Duration of the default suite is unknown. If repeats cannot agree without a broader test rewrite, the weeks appetite is wrong and this decision is revisited.
