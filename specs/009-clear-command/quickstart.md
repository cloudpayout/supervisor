# Quickstart: Clear instance registry

Validation for [spec.md](./spec.md). Contract: [contracts/cli.md](./contracts/cli.md). Data: [data-model.md](./data-model.md).

## Prerequisites

- Linux or WSL2
- Go 1.27.1 from `go.mod`
- Repository root

## Build

```text
go build -o /tmp/supervisor ./cmd/supervisor
```

Use an isolated state dir (same pattern as integration tests):

```text
export XDG_STATE_HOME=$(mktemp -d)
export XDG_DATA_HOME=$(mktemp -d)
```

## Root help (User Story 2)

```text
/tmp/supervisor help
```

**Expect**: Available commands include `clear` alongside `start`, `status`, `stop`, `screenshot`, `logs`, and `help`.

```text
/tmp/supervisor clear --help
```

**Expect**: Describes registry wipe, confirmation on a TTY, `--yes` for scripts, and that browsers/workers are not stopped.

## Non-interactive clear (User Story 1, scripts)

Seed a registry row (minimal JSON):

```text
mkdir -p "$XDG_STATE_HOME/supervisor"
printf '%s\n' '{"schema_version":1,"instances":[{"id":"test01","state":"stopped","created_at":"2026-01-01T00:00:00Z","profile_dir":"/tmp/p"}]}' \
  > "$XDG_STATE_HOME/supervisor/instances.json"
```

```text
/tmp/supervisor clear --yes
/tmp/supervisor status
```

**Expect**:

- `clear --yes` exits `0`
- `instances.json` contains `"instances": []` (valid JSON, schema version 1)
- `status` prints `No instances are running.` and exits `0`

## Non-TTY guard (User Story 3)

With the same seeded registry:

```text
echo yes | /tmp/supervisor clear
```

**Expect**:

- Exit `2`
- stderr explains `--yes` is required when stdin is not a terminal
- `instances.json` still contains `test01`

## Interactive decline (manual, User Story 1)

On a real terminal (stdin is a TTY), with seeded registry:

```text
/tmp/supervisor clear
```

When prompted, enter `n` or press Enter.

**Expect**: Exit `0`, registry unchanged, no “cleared” success claim.

Affirmative path: run again, answer `y` or `yes`.

**Expect**: Exit `0`, empty registry, empty `status`.

## Corrupt registry (edge case)

```text
printf 'not-json\n' > "$XDG_STATE_HOME/supervisor/instances.json"
/tmp/supervisor clear --yes
```

**Expect**: Exit `1`, file still `not-json`, stderr mentions decode/registry error.

## Automated suite

After implementation:

```text
go test ./... -count=1
go test ./tests/integration/ -run Clear -v
```

**Expect**: All pass; new integration tests cover `--yes`, non-TTY refusal, and corrupt registry.