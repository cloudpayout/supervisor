# Quickstart: Validate 008 Log Files

**Feature**: `008-log-files` | **Contract**: [contracts/cli.md](./contracts/cli.md)

## Prerequisites

- Linux or WSL2
- Go 1.27.1+
- Chrome available for headless `start`

## Build

```bash
go build -o supervisor ./cmd/supervisor
```

## 1. Structured log on start

```bash
export XDG_STATE_HOME="$(mktemp -d)"
./supervisor start --headless
# copy printed instance ID → $ID
./supervisor logs "$ID" | head
```

**Expect**: Lines match `TIMESTAMP ERROR|NOTICE|INFO message`; no usage error.

## 2. Level filter

```bash
./supervisor logs "$ID" --level error
```

**Expect**: Only `ERROR` structured lines; exit `0`.

## 3. Follow

```bash
./supervisor logs "$ID" --follow
# Ctrl+C after a few lines
```

**Expect**: Exit `0`; new worker lines appear within a few seconds.

## 4. Orphaned log after stop

```bash
./supervisor stop "$ID"
./supervisor logs "$ID" | head
```

**Expect**: Exit `0`; stderr mentions instance no longer tracked (if applicable).

```bash
./supervisor logs --recent
```

**Expect**: Row for `$ID` with UTC activity time and error summary or `no error recorded`.

## 5. Automated suite

```bash
go test ./...
```

**Expect**: New `logs_*` integration tests pass; existing start/status/stop tests unchanged.

## 6. Secrets (when using onewin session tests)

Re-run or extend `onewin_secrets_test` pattern: session cookie MUST NOT appear in `logs <id>` output.

## References

- Data layout: [data-model.md](./data-model.md)
- Research decisions: [research.md](./research.md)
- Spec: [spec.md](./spec.md)