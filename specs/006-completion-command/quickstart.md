# Quickstart: Remove shell completion command

Validation for [spec.md](./spec.md). Contract: [contracts/cli.md](./contracts/cli.md). Catalog: [data-model.md](./data-model.md).

## Prerequisites

- Linux or WSL2
- Go 1.27.1 from `go.mod`
- Repository root

## Build

```text
go build -o /tmp/supervisor ./cmd/supervisor
```

## Root help (User Story 1)

```text
/tmp/supervisor help
```

**Expect**:

- Under “Available Commands”: `start`, `status`, `stop`, `screenshot`, `help`
- No line whose first token is `completion`
- No mention of installing shell completion

Repeat with `/tmp/supervisor` (no args) and `/tmp/supervisor --help`; same rules.

## Completion rejected (User Story 2)

```text
/tmp/supervisor completion
/tmp/supervisor completion bash
/tmp/supervisor help completion
```

**Expect** for `completion` and `completion bash`:

- Exit code `2`
- Empty standard output (no `# bash completion for supervisor` or similar script body)
- Standard error mentions `unknown command`

**Expect** for `help completion` (FR-004):

- Exit code `0`
- Empty standard output
- Standard error contains `Unknown help topic` (completion is not available)

## No regression (User Story 3)

```text
go test ./...
```

**Expect**: All packages pass, including existing `tests/integration` lifecycle and start-help tests without changes to operator workflows.

After implementation lands, also run the dedicated test (planned in [research.md](./research.md#r5-testing-strategy)):

```text
go test ./tests/integration/ -run RootHelp -v
```

(Exact test name may differ in `tasks.md`; assert help ban and completion exit code.)

## Manual spot check (optional)

```text
/tmp/supervisor start --help
```

**Expect**: Same start help as before 006 (positional path, `--headless`; no removed flags).