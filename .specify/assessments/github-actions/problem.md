# Problem Definition: No test result after a push or pull request

- **Slug**: github-actions
- **Created**: 2026-11-12
- **Inputs used**: intake.md, research.md

## Problem Statement

People who change `github.com/cloudpayout/supervisor` cannot tell whether the project's Go tests succeeded after a push or a pull request, because the repository records no test result for those changes. The only documented check is a local test run, which this research could not execute in the workspace that was checked, and which reports success while skipping the browser tests when Chrome is absent.

## Affected Users & Stakeholders

- **Users**: Oscar Cedano, who asked to see whether tests succeeded after pushing and opening pull requests. The same request says the author does not matter, so any author of a push or pull request is in the affected set. No second requester, ticket, or usage metric was found. — [source: intake.md; research.md Users & Demand]
- **Stakeholders**: Oscar Cedano raised the request and is the only named decision maker. The PR will be blocked when a test fails or gets blocked.

## Goals

- After a push or a pull request, the author can tell whether the tests succeeded without relying on a local run.
- The result covers the default test set (`go test ./...`) on any branch. The later note in research gaps names that set, and names any branch, as the intended scope. — [source: research.md Gaps & Open Questions]
- Browser tests in that set actually run. A skip because Chrome is missing does not count as knowing they succeeded. The research gaps note says those tests should not be skipped. — [source: research.md Gaps & Open Questions; research.md Data & Constraints]
- The result is meaningful for this repository only if it comes from Go 1.27.1, the version `go.mod` declares and the constitution requires unless a spec amends it. — [source: research.md Data & Constraints]

## Non-Goals

- The `slow` ten-instance test. It is outside `go test ./...`, and the research gaps note limits scope to that default set. The feature tasks already treat it as optional when resources are tight. — [source: research.md Prior Art and Gaps]
- Live casino or other external game checks. The constitution says they must not gate the default test run, and no such test exists. — [source: research.md Prior Art and Data & Constraints]
- macOS and Windows as platforms for the result. v1 support is Linux and WSL2 only. — [source: research.md Data & Constraints, FR-014]
- New supervisor product behavior. The hurt is the missing result for the tests that already exist.
- Supplying secrets or credentials. The intake says the tests need none, and the constitution forbids committing them. — [source: intake.md; research.md Data & Constraints]
- A cost or billing estimate. The research gaps note says measuring cost is unnecessary. The sentence there about GitHub's free tier is not backed by a fetched billing source, so it is not treated as evidence. — [source: research.md Gaps & Open Questions; research.md Evidence Against]

## Success Metrics

- Share of pushes and pull requests that have a recorded pass/fail for the default test set, on any branch. (baseline: none — this clone has no `.github` history, and the public `main` workflows path returned HTTP 404)
- Of the browser tests that call `requireChrome` and are part of the default set (`TestParallelStartRegistry`, `TestStartStatusStop`, `TestStartRegistersAndReachesRunning`, `TestStatusReflectsConcurrentChanges`), the share that finish as pass or fail rather than skip. (baseline: they skip when Chrome is missing; a run with Chrome present was not measured)
- The recorded result is produced with Go 1.27.1. (baseline: unknown — the suite was not executed in research)
- Qualitative: the author can answer whether the tests succeeded from the change itself. (baseline: they cannot; the repository has no such signal)

## Cost of Inaction

Pushes and pull requests keep shipping with no in-repo test result. Anyone who wants the answer keeps depending on a local `go test ./...` on Linux or WSL2 with Chrome available. On the machine checked during research, Go was not on `PATH` and no Chrome binary was on `PATH`, so that local check could not be run. Where Chrome is missing, the four default browser tests skip and the run can still look successful. How often people already run the suite before they push was not measured. The only stated cost is the one Oscar Cedano named: not knowing whether the tests succeeded after a push or a pull request.

## Open Questions

- Yes, we will use `ubuntu-latest`, the intake document said that no Github-hosted runner would be used, but it was removed.
- Yes, pull requests from forks are in the scope.
- Test executions must be stable across repeats.
