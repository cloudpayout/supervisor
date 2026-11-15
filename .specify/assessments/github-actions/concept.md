# Concept: A recorded test result on each change

- **Slug**: github-actions
- **Created**: 2026-11-12
- **Recommended option**: Option C — Repeat the check before it counts

## Options

### Option A — Leave it local

- **Sketch**: Nothing is added to the repository. After a push or a pull request, the author still finds out whether the tests passed only by running `go test ./...` themselves on Linux or WSL2, with Chrome available. A machine without Chrome can still report success while the four default browser tests skip. The pull request is not held when tests fail.
- **Appetite**: small
- **Trade-offs**: No new maintenance and no runner questions. It misses every goal: there is still no result on the change, browser skips still look like success, and the Go version of a local run is whoever's machine. The researched workspace could not run the suite at all.
- **Rabbit holes**: None in the repository. The cost sits with each author, and that cost was not measured.

### Option B — One check per change

- **Sketch**: Each push and each pull request, on any branch and from any fork, shows one pass/fail on the change itself. The check uses the GitHub-hosted `ubuntu-latest` label the problem note now chooses, Go 1.27.1, and a browser so the default suite's browser tests run instead of skipping. A skip, a failure, or a check that cannot run blocks the pull request. The suite runs once per change.
- **Appetite**: small (days). Uncertain: the suite has never been executed in research, so the first red result can turn those days into test-fixing work that this budget does not include.
- **Trade-offs**: Meets the visible result, the default `go test ./...` set, any branch, fork pull requests, Go 1.27.1, and "a skip is not a pass." Sacrifices the later decision that executions must be stable across repeats: one green run can be luck when tests sleep for a few hundred milliseconds around real Chrome processes. Public prior art for a single Go test check is the `actions/starter-workflows` `ci/go.yml` example, which runs `go test` once on `ubuntu-latest`; that example pins an older Go version and does not install a browser.
- **Rabbit holes**: Making Chrome present, and treating a skip as a failure, can require test changes rather than only a new check. Fork pull requests raise the question of what the check is allowed to do to the repository. A red first run on the hosted runner, especially around headless Chrome, can swallow the appetite. `intake.md` still says both `ubuntu-latest` and that no GitHub-hosted runner will be used; building on the old sentence would reopen the runner choice.

### Option C — Repeat the check before it counts

- **Sketch**: Same visible check as Option B: any branch, fork pull requests included, `ubuntu-latest`, Go 1.27.1, default suite only, browser tests actually executed. The difference the author sees is that success means the same change passed on more than one execution, and a failure, a browser-test skip, or a check that cannot run still blocks the pull request. One lucky pass does not answer "did the tests succeed?"
- **Appetite**: medium (weeks). Uncertain, and the evidence cannot pin this tighter. Adding the check is days of work, as in Option B. The weeks are the budget for the suite to pass repeatedly on the hosted runner. Flake rate and duration were never measured, so this could shrink to days if the suite already agrees with itself, or outgrow weeks if the fixed sleeps have to be reworked. That uncertainty is part of the appetite, not a hidden estimate.
- **Trade-offs**: This is the only option that matches the problem note's decision that test executions must be stable across repeats, and the stakeholder line that a pull request is blocked when a test fails or gets blocked. It costs more than one run per change, and it can force changes to the existing tests when repeats disagree. It still does not measure billing; the problem puts a cost estimate out of scope. The claim that GitHub's free tier makes cost irrelevant stays uncited.
- **Rabbit holes**: Defining "stable" as a number of repeats, then discovering the repeats disagree because of the 300ms and 500ms sleeps in the default integration tests. Fixing that disagreement by changing tests, Chrome startup, or process timing. Fork checks that need broader permissions than this repository's "no secrets" constraint. Treating "the check gets blocked" as a merge block without a clear rule for a check that never starts.

## Recommendation

Option C. The goals are a result on the change itself, for the default suite, on any branch, with browser tests actually run, under Go 1.27.1. The problem note adds three decisions Option B does not satisfy together: `ubuntu-latest` is the runner, fork pull requests are in scope, and the result must be stable across repeats. A pull request should be blocked when a test fails or the check is blocked. Option A leaves the baseline at "no result." Option B records a result and can still hide a flake behind one pass, which misses the stability decision and the metric that a skip is not a success only if that result is trustworthy on a second run.

The recommendation does not clear the bar as a large program. Demand is still one named person, and no duration or flake rate exists. The medium appetite is the ceiling for making the existing default suite agree with itself on the hosted runner. If the first repeats already pass, the work collapses toward Option B's days. If they do not, stopping to fix tests is the work, not widening into the slow ten-instance test or new product behavior.

## Out of Scope (for the recommended option)

- The `slow` ten-instance test, including as an extra repeated job.
- Live casino or other external game checks.
- macOS and Windows as the place the result is produced. v1 remains Linux and WSL2.
- New supervisor commands or session behavior.
- Secrets or credentials in the check.
- A billing or free-tier analysis.
- A self-hosted runner. The problem note says the "no GitHub-hosted runner" constraint was removed and that `ubuntu-latest` will be used.
- Rewriting the product so the suite is faster. Test changes are in bounds only where repeats disagree or browser tests skip.

## Assumptions to Validate

- The problem note is the current decision where it disagrees with `intake.md`. Intake still says "no Github-hosted runner will be used" and still limits the trigger to `main`. The recommendation assumes those two intake sentences are stale.
- `ubuntu-latest` is a Linux environment that satisfies the v1 Linux/WSL2 limit. Runner documentation on `docs.github.com` was not fetched.
- Go 1.27.1 can be used on that runner. The fetched `actions/go-versions` manifest lists `1.27.1` as stable. That a given runner image already has it installed was not checked.
- A browser can be present so `TestParallelStartRegistry`, `TestStartStatusStop`, `TestStartRegistersAndReachesRunning`, and `TestStatusReflectsConcurrentChanges` finish as pass or fail rather than skip.
- "Stable across repeats" means a small fixed number of executions of the same change, chosen during specification, all of them passing. The problem note does not give that number.
- "The PR will be blocked when a test fails or gets blocked" means both a failing test and a check that cannot complete hold the pull request.
- Fork pull requests can run this check without secrets and without write access to the repository.
- The default suite finishes in a bounded time. Duration is unknown.
- The suite can be made to agree with itself inside a weeks budget. If the first investigation shows the sleeps cannot pass twice without a broader test rewrite, the appetite is wrong and the decision should be revisited.
