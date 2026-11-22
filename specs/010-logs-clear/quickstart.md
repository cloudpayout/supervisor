# Quickstart: `logs clear`

Validation for [spec.md](./spec.md). Contract: [contracts/cli.md](./contracts/cli.md). Data: [data-model.md](./data-model.md).

## Prerequisites

- Linux or WSL2
- Go 1.27.1 from `go.mod`
- Repository root

## Build

```text
go build -o /tmp/supervisor ./cmd/supervisor
export XDG_STATE_HOME=$(mktemp -d)
export XDG_DATA_HOME=$(mktemp -d)
```

## Help (User Story 2)

```text
/tmp/supervisor logs --help
```

**Expect**: Subcommand `clear` listed.

```text
/tmp/supervisor logs clear --help
```

**Expect**: Mentions eligible `*.log` deletion, skip for `starting`/`running`/`stopping`, `--yes`, registry unchanged.

```text
/tmp/supervisor help
```

**Expect**: No top-level `clear` command (009 retired).

## Selective clear (User Story 1)

```text
mkdir -p "$XDG_STATE_HOME/supervisor/logs"
printf 'line\n' > "$XDG_STATE_HOME/supervisor/logs/orphan01.log"
printf 'line\n' > "$XDG_STATE_HOME/supervisor/logs/active01.log"
mkdir -p "$XDG_STATE_HOME/supervisor"
printf '%s\n' '{"schema_version":1,"instances":[{"id":"active01","state":"running","created_at":"2026-01-01T00:00:00Z","profile_dir":"/tmp/p"}]}' \
  > "$XDG_STATE_HOME/supervisor/instances.json"
```

```text
/tmp/supervisor logs clear --yes
ls "$XDG_STATE_HOME/supervisor/logs/"
```

**Expect**:

- Exit `0`
- `orphan01.log` absent
- `active01.log` still present
- `instances.json` still contains `active01`

## Non-TTY guard (User Story 3)

Re-seed orphan log only, then:

```text
echo yes | /tmp/supervisor logs clear
```

**Expect**: Exit `2`, orphan file still present, stderr mentions `--yes` or non-terminal.

## Decline (manual TTY)

On an interactive terminal, run `logs clear`, answer `n`.

**Expect**: Exit `0`, all log files unchanged.

## Automated suite

```text
go test ./... -count=1
go test ./tests/integration/ -run LogsClear -v
```

**Expect**: Pass; no `TestClearYes` for root `clear`.