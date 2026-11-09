# Quickstart: Validate Supervisor CLI (001-supervisor-cli)

End-to-end validation scenarios for operators and implementers. See [cli contract](./contracts/cli.md) and [data model](./data-model.md).

## Prerequisites

- Go 1.27.1+
- Chrome or Chromium available on `PATH` (chromedp default discovery)
- Linux or WSL2
- Build: `go build -o supervisor ./cmd/supervisor` (after implementation)

Optional: `export XDG_STATE_HOME=/tmp/supervisor-test-state` and `export XDG_DATA_HOME=/tmp/supervisor-test-data` for isolated runs.

## Scenario A — Start and background ownership (P1, SC-001)

1. Run `./supervisor start` (default headed) or `./supervisor start --headless`.
2. **Expect**: Shell prompt returns quickly; stdout prints a new instance `id`.
3. Run `./supervisor status` in another terminal (briefly).
4. **Expect**: Listed `id`; may show `starting` with CPU/RAM `unavailable`, then `running` with numeric metrics (FR-004b).
5. **Expect**: Chrome window (headed) or headless process with blank page (profile under data dir).
6. Run `./supervisor start --headed --headless`.
7. **Expect**: Usage error, exit `2`, no new registry row (FR-001d).

**Spec refs**: User Story 1; FR-001, FR-001a, FR-001b, FR-001d.

## Scenario B — Concurrent instances (SC-006, SC-007)

1. Run `./supervisor start` three times in one terminal without waiting.
2. **Expect**: Three distinct IDs printed.
3. Run `./supervisor status`.
4. **Expect**: Three rows within ~10s; metrics scoped per row (Story 2 scenario 1).
5. (Optional) Run two `start` commands in parallel from separate shells.
6. **Expect**: Both succeed with different IDs (FR-012).

**Spec refs**: FR-001c, FR-011, FR-012; clarifications session 2026-11-07.

## Scenario B2 — Start failure without phantom ID (FR-011, User Story 1 scenario 3)

1. Run with invalid Chrome: `CHROME_PATH=/nonexistent ./supervisor start --headless`.
2. **Expect**: Non-zero exit; clear error; **no** instance ID on stdout; `instances.json` unchanged (or no new row).
3. Run `./supervisor status`.
4. **Expect**: No new instance from the failed attempt.

**Spec refs**: FR-011; plan “start failure vs starting”.

## Scenario C — Live status refresh (SC-002)

1. With at least one instance running, run `./supervisor status`.
2. **Expect**: Table updates about every 2 seconds until Ctrl+C.
3. With multiple instances, confirm each row updates independently.
4. **(SC-002 load, optional)** Start ten headless instances; run `./supervisor status` for ~10s.
5. **Expect**: All ten IDs listed with CPU/RAM columns; each row refreshes at least once within a 2s window (or run `go test -tags=slow ./tests/integration/... -run TenInstances`).

**Spec refs**: FR-004, FR-004a, FR-005; SC-002.

## Scenario D — Stop by prompt (P3, SC-003)

1. Note an instance `id` from `start`.
2. Run `./supervisor stop`; enter `id` at prompt.
3. **Expect**: Success message; instance absent from `status` within ~5s; Chrome no longer running for that profile.
4. Run `./supervisor stop` with unknown ID.
5. **Expect**: Clear not-found message; exit `0`; other instances unaffected (FR-007b).
6. Stop the same ID again (or after manual kill).
7. **Expect**: Already-stopped message; exit `0`; stale row pruned (FR-007a).
8. Run `./supervisor stop` and submit empty input.
9. **Expect**: No termination; exit `0` (FR-006a).

**Spec refs**: FR-006, FR-006a, FR-007, FR-007a, FR-007b, FR-008.

## Scenario E — Empty status (SC-005)

1. Ensure no instances (stop all or fresh state dir).
2. Run `./supervisor status`.
3. **Expect**: Plain-language empty message, not an error exit.

## Scenario G — Failed start timeout (FR-013, SC-008)

1. (Test harness) Register a `starting` instance with an old `created_at` or block worker launch.
2. Run `./supervisor status` once after 60s+.
3. **Expect**: Instance removed from list; not shown as healthy `running`.

## Scenario H — Split-brain cleanup (FR-009a/b)

1. With a running instance, simulate owner crash with Chrome still running (or the reverse).
2. Run `./supervisor status`.
3. **Expect**: Row removed within one refresh; orphan Chrome or owner terminated.

## Scenario F — Registry concurrency (implementation)

1. Run integration test `TestParallelStartRegistry` (or manual parallel starts).
2. **Expect**: `instances.json` valid JSON; unique IDs; no torn writes under parallel `start`.

**Spec refs**: FR-012; [registry-schema.json](./contracts/registry-schema.json).

## Automated tests (constitution)

```bash
go test ./...
```

- Default tests: headless or mocked Chrome; no live casino.
- Tag external game tests with `//go:build external` if added later.

## Troubleshooting

| Symptom | Check |
|---------|--------|
| `start` fails immediately | Chrome installed; disk space; logs under `$XDG_STATE_HOME/supervisor/logs/` |
| Instance stuck `starting` | Worker log; reconcile timeout in worker/registry |
| Metrics `unavailable` | PID dead — status should prune on next cycle |
| Stale instance after kill | Run `status` to reconcile; verify flock not held by crashed process |
