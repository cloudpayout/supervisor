# Idea Intake: Run Go tests in GitHub Actions

- **Slug**: github-actions
- **Created**: 2026-11-12
- **Source**: pasted text; codebase pointer (this repository, module `github.com/cloudpayout/supervisor`)
- **Type**: new-capability

## Idea (as captured)

> Execute all Golang test files of this project to Github Actions

The request refers to this repository (`github.com/cloudpayout/supervisor`, Go 1.27.1). Go test files are present under `tests/unit/` and `tests/integration/` (`*_test.go`). No `.github/` workflow files were found in the repository at intake time.

## Restated

The proposal is to run all of this project's Go test files through GitHub Actions.

## Origin & Context

- **Raised by**: Oscar Cedano <oscar.cedano@outlook.com>
- **Trigger**: To see if the tests were sucessful after pushing and creating pull requests.

## First-Glance Unknowns

- push, pull request.
- Yes, both should be run, and yes, they need a browser.
- Go version 1.27.1, ubuntu-latest, and no Github-hosted runner will be used.
- No secrets or credentials are needed for the tests to succeed
- Every pull request or push to `main` branch, it should trigger the workflow. It won't matter the author, it should always run.
