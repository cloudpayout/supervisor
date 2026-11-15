# Idea Research: Run Go tests in GitHub Actions

- **Slug**: github-actions
- **Created**: 2026-11-12
- **Evidence confidence (overall)**: medium

## Users & Demand

- One stated request, recorded in the intake note: Oscar Cedano wants to see whether this project's tests succeeded after a push or a pull request. The same note says both unit and integration tests should run, a browser is required, the Go version is 1.27.1, the workflow should use `ubuntu-latest`, no GitHub-hosted runner will be used, no secrets are required, and every push or pull request to `main` should run regardless of author. — [source: `.specify/assessments/github-actions/intake.md`] (confidence: high)
- No other demand signal was found in this repository: no issue archive, interview notes, usage metrics, or earlier assessment about CI. The want is stated once; it is not observed behavior from an existing pipeline. — [source: repository search of `.specify/` and `specs/`; `git log --all -- .github` in this clone is empty] (confidence: high for absence in this repo; low that this is the only person who wants it)

## Prior Art

- This repository has no GitHub Actions workflow. This clone's history has no `.github` path. The public repository `cloudpayout/supervisor` (default branch `main`) returned HTTP 200, and `https://github.com/cloudpayout/supervisor/tree/main/.github/workflows` returned HTTP 404 "File not found". — [source: local `git log`; pinned fetch of those two GitHub URLs on 2026-11-12, peer `140.82.114.4`, no redirects] (confidence: high)
- Accepted specs tell a developer to run tests locally with `go test ./...`. They do not describe a workflow file. Feature 001's quickstart also says default tests use headless or mocked Chrome and no live casino, and that any future game check should use `//go:build external` and stay out of the default run. No `//go:build external` test exists. — [source: `specs/001-supervisor-cli/quickstart.md`, `specs/002-status-stop-cli/plan.md`, `specs/002-status-stop-cli/quickstart.md`] (confidence: high)
- The feature task list already treats CI as a resource limit, without adding CI. Task T038 puts the ten-instance status test behind build tag `slow` so it can be skipped or shortened "in CI if resource-constrained". Task T033, still unchecked, says the manual ten-instance check is optional when that test is skipped in CI. — [source: `specs/001-supervisor-cli/tasks.md`] (confidence: high)
- GitHub's public starter workflow `actions/starter-workflows` `ci/go.yml` on `main` is a Go CI example: it triggers on `push` and `pull_request` to a `$default-branch` placeholder, uses `runs-on: ubuntu-latest`, checks out with `actions/checkout@v4`, sets up Go with `actions/setup-go@v4` and `go-version: '1.20'`, then runs `go build -v ./...` and `go test -v ./...`. A comment in that file links to GitHub's Go build-and-test docs; that linked page was not fetched. — [source: `https://github.com/actions/starter-workflows/blob/main/ci/go.yml`, pinned fetch, peer `140.82.114.4`] (confidence: high)
- The current `actions/setup-go` README says the action optionally downloads and caches a Go version and adds it to `PATH`. It documents `go-version-file` as a path to `go.mod` (or `go.work`, `.go-version`, or `.tool-versions`), and says `go-version` wins if both inputs are set. Examples on that page use `go-version` values `1.23` and `1.25`, with the note "The Go version to download (if necessary) and use." If the tool cache misses, it reads the manifest on `actions/go-versions` `main`, then falls back to a direct download. — [source: `https://github.com/actions/setup-go/blob/main/README.md`, pinned fetch, peer `140.82.114.4`] (confidence: high)
- That manifest lists `"version": "1.27.1"` with `"stable": true`, alongside `1.27.0` through `1.27.4`. — [source: `https://github.com/actions/go-versions/blob/main/versions-manifest.json`, pinned fetch, peer `140.82.114.4`] (confidence: high)

## Market & Context

- The documented way to run the suite today is a local `go test ./...` on Linux or WSL2, with Chrome or Chromium on `PATH` (or `CHROME_PATH`). — [source: `specs/001-supervisor-cli/quickstart.md`, `specs/001-supervisor-cli/research.md` decision 10, `specs/001-supervisor-cli/spec.md` FR-014] (confidence: high)
- In this workspace on 2026-11-12, `go` was not on `PATH` (`go: command not found`), and none of `google-chrome`, `chromium`, `chromium-browser`, or `chrome` was on `PATH`. A local run of the suite was not possible here. This is one machine, not a survey of developers. — [source: shell checks in this session] (confidence: high for this machine)
- Doing nothing leaves push and pull request with no in-repo automated test signal. There is no count of how often people already run `go test` before they push. — [source: absence of workflows above; frequency of local runs is ASSUMPTION] (confidence: high for the missing signal; low for how often local runs happen)

## Data & Constraints

- Module `github.com/cloudpayout/supervisor` declares `go 1.27.1`. The constitution says the implementation Go version MUST match `go.mod` (1.27.1 at ratification) unless a spec amends it. — [source: `go.mod`, `.specify/memory/constitution.md`] (confidence: high)
- The constitution also says automated tests MUST cover CLI start, status, and stop before that behavior is complete, and default tests MUST run without a live casino site. An external game check MUST NOT gate the default run. Credentials and secrets MUST NOT be committed. — [source: `.specify/memory/constitution.md`] (confidence: high)
- The tree has 16 `*_test.go` files and 17 `Test*` functions. `tests/integration/harness_test.go` and `tests/integration/exec_test.go` are helpers only. No test file contains `http://`, `https://`, `password`, `token`, `secret`, or a casino name. — [source: `tests/**/*_test.go`] (confidence: high)
- Five tests call `requireChrome` and skip with `t.Skip("chrome not installed")` when discovery fails: `TestParallelStartRegistry`, `TestStartStatusStop`, `TestStartRegistersAndReachesRunning`, `TestStatusReflectsConcurrentChanges`, and `TestStatusTenInstances`. Discovery matches the product: `CHROME_PATH`, else `google-chrome`, `google-chrome-stable`, `chromium`, `chromium-browser`, or `chrome` on `PATH`. — [source: `tests/integration/harness_test.go`, `internal/browser/chromedp.go`] (confidence: high)
- Browser-launching tests that were inspected pass `--headless`. The feature research decision says tests use headless Chrome and the default operator mode is headed when flags are omitted. — [source: `tests/integration/registry_concurrent_test.go`, `tests/integration/start_worker_test.go`, `tests/integration/status_concurrent_test.go`, `specs/001-supervisor-cli/research.md` decisions 9 and 11] (confidence: high)
- `TestStatusTenInstances` is behind `//go:build slow`, so `go test ./...` does not compile it. It starts ten `--headless` instances. — [source: `tests/integration/status_ten_instances_test.go`] (confidence: high)
- The integration harness builds `./cmd/supervisor` with `go build` during tests, so a `go` binary must be on `PATH` inside the test process, not only as the program that launched `go test`. — [source: `tests/integration/harness_test.go`] (confidence: high)
- v1 support is Linux and WSL2 only (FR-014). `internal/cli/start.go` sets `syscall.SysProcAttr{Setsid: true}` and has no OS build tag. Whether that file compiles on other `GOOS` values was not verified by running a build. — [source: `specs/001-supervisor-cli/spec.md` FR-014, `internal/cli/start.go`] (confidence: high for the requirement and the source line; the cross-compile result is unverified)
- The intake's runner lines conflict with each other: it names `ubuntu-latest` and also says no GitHub-hosted runner will be used. The starter workflow above uses the label `ubuntu-latest`. Runner-label documentation on `docs.github.com` was not fetched, because that host is outside the allowlist and the research rule forbids crawling links found inside a fetched page. — [source: intake note; starter workflow URL above] (confidence: high that the intake contains both statements; the meaning of the label is [NEEDS CLARIFICATION])
- Several integration tests sleep 100ms, 300ms, or 500ms while they drive real processes. Wall-clock duration and flake rate on a runner were not measured. `go` is absent in this workspace, so the suite was not executed during research. — [source: `tests/integration/registry_concurrent_test.go`, `start_worker_test.go`, `status_concurrent_test.go`, `status_refresh_test.go`, `status_ten_instances_test.go`; shell `go: command not found`] (confidence: high that those sleeps exist and that this session did not run the suite)

## Evidence Against the Idea

- Demand evidence is a single intake statement. There is no ticket volume, usage data, or history of broken builds to show the problem is recurring. — [source: intake note; empty `.github` history] (confidence: high)
- A green `go test` can hide the browser tests. `requireChrome` skips them when Chrome is missing, and the skip is not a failure. The intake says a browser is required and that the tests should actually run. — [source: `tests/integration/harness_test.go`; intake note] (confidence: high)
- "All Golang test files" includes the `slow` ten-instance test, which the task list already set aside for resource-constrained CI. That test starts ten headless Chrome processes. Whether it belongs in the default Actions run is unresolved, and no runtime or cost figure was measured. — [source: `tests/integration/status_ten_instances_test.go`, `specs/001-supervisor-cli/tasks.md` T038 and T033] (confidence: high for the tag and the task wording; cost is unmeasured)
- The intake asks for `ubuntu-latest` and also says no GitHub-hosted runner will be used. Those cannot be treated as one constraint until someone says which runner is actually available. v1 is Linux and WSL2 only, so a non-Linux runner would sit outside FR-014. No self-hosted runner inventory was provided. — [source: intake note; FR-014] (confidence: high)
- Pass, fail, duration, and flake rate are unknown. Timing sleeps around live Chrome processes are a plausible flake source on a shared runner, and that stability claim is an assumption until the suite is run somewhere that has Go and Chrome. — [ASSUMPTION on flake risk; source for the sleeps: integration tests listed above] (confidence: low)

## Gaps & Open Questions

- `ubuntu-latest`.
- default `go test ./...` set.
- No, they shouldn't skip those tests.
- Yes, execute the workflow for any branch
- It is not necesssary to measure the costs. With the Github's free tier we can execute the workflow as many times as we want without worrying about costs.

## Sources

- `https://github.com/actions/starter-workflows/blob/main/ci/go.yml` (host: github.com, policy: allowlisted; pinned to `140.82.114.4`; HTTP 200; no redirects)
- `https://github.com/actions/setup-go/blob/main/README.md` (host: github.com, policy: allowlisted; pinned to `140.82.114.4`; HTTP 200; no redirects)
- `https://github.com/actions/go-versions/blob/main/versions-manifest.json` (host: github.com, policy: allowlisted; pinned to `140.82.114.4`; HTTP 200; no redirects)
- `https://github.com/cloudpayout/supervisor` (host: github.com, policy: allowlisted; pinned to `140.82.114.4`; HTTP 200; no redirects; page payload default branch `main`)
- `https://github.com/cloudpayout/supervisor/tree/main/.github/workflows` (host: github.com, policy: allowlisted; pinned to `140.82.114.4`; HTTP 404; no redirects)
- `https://docs.github.com/en/actions/automating-builds-and-tests/building-and-testing-go` (host: docs.github.com, policy: not fetched; linked from the starter workflow; not on the allowlist; linked pages were not crawled)
