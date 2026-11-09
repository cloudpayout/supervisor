# Research: Supervisor CLI (001-supervisor-cli)

**Date**: 2026-11-09 (updated)

## 1. CLI framework

**Decision**: Use `github.com/spf13/cobra` for `start`, `status`, `stop`, and hidden `worker run`.

**Rationale**: Standard Go CLI pattern; subcommands map 1:1 to spec; easy to add flags later without breaking v1.

**Alternatives considered**:

- `urfave/cli` — viable but less common in Go CLIs at scale.
- `flag` only — too little structure for nested worker command.

## 2. Long-lived chromedp session without blocking `start`

**Decision**: Detached **re-exec** of the same `supervisor` binary as `supervisor worker run --id <id>`.

**Rationale**: Matches clarification: CLI exits immediately; one owner process per instance holds chromedp context (FR-001a/b). Re-exec avoids a separate worker binary and keeps deployment one artifact.

**Alternatives considered**:

- Long-lived supervisor **daemon** with in-process contexts — rejected in clarify (not chosen); more moving parts for v1.
- Foreground `start` blocking in shell — rejected; prevents convenient concurrent starts from one terminal.
- Raw Chrome only without Go owner — loses chromedp session ownership and clean metrics attribution.

## 3. Cross-invocation instance registry

**Decision**: JSON file registry with **OS file lock** (`flock`) on the registry file (or sibling `.lock`).

**Rationale**: Satisfies FR-003, FR-012; human-debuggable; no database; works across terminals on one host.

**Alternatives considered**:

- SQLite — heavier; unnecessary for v1 instance counts.
- In-memory only in daemon — requires daemon; user chose file registry.
- Global mutex via singleton daemon — serializes starts unnecessarily.

**Locking pattern**: Short critical sections: allocate ID + append `starting` record; spawn worker outside lock; worker updates `running` + PIDs under lock.

## 4. Browser isolation (concurrent instances)

**Decision**: chromedp `ExecAllocator` with per-instance `UserDataDir` under `$XDG_DATA_HOME/supervisor/profiles/<instance_id>/`.

**Rationale**: FR-001c; separate Chrome processes and profiles; no cookie/storage leakage.

**Alternatives considered**:

- Shared profile, multiple targets — violates spec.
- Remote browser allocator — out of scope for local supervisor v1.

## 5. Resource metrics

**Decision**: `github.com/shirou/gopsutil/v4/process` to walk from `owner_pid` and `browser_pid`, aggregate CPU % and RSS for each instance row.

**Rationale**: FR-004a; portable across Linux/WSL; no custom `/proc` parsing in application code.

**Alternatives considered**:

- `ps` subprocess parsing — brittle, slower.
- Chrome DevTools Protocol only — does not include Go worker RSS/CPU.

**Unavailable metrics**: If any expected PID is missing, show `unavailable` for that field and schedule registry prune on next locked write.

## 6. `status` live view

**Decision**: Ticker every **2 seconds**; redraw table to stdout (ANSI clear screen when TTY, else newline-separated snapshots).

**Rationale**: FR-005, SC-002; simple implementation without TUI framework in v1.

**Alternatives considered**:

- `bubbletea` TUI — nicer UX but extra dependency and scope.
- One-shot `status` — violates FR-005.

## 7. Instance IDs

**Decision**: Opaque URL-safe random IDs (e.g. UUID v4 or `crypto/rand` base32), 8–16 chars display optional suffix for readability in plan/tasks.

**Rationale**: FR-002; copy/paste at stop prompt; no leaking sequence information.

## 8. State directory layout

**Decision**:

| Path | Purpose |
|------|---------|
| `$XDG_STATE_HOME/supervisor/instances.json` | Registry |
| `$XDG_STATE_HOME/supervisor/logs/<id>.log` | Worker stdout/stderr |
| `$XDG_DATA_HOME/supervisor/profiles/<id>/` | Chrome user data |

Fallback when XDG unset: `~/.local/state` and `~/.local/share` per XDG defaults.

## 9. Testing without casino sites

**Decision**: Integration tests use headless Chrome via chromedp; unit tests mock registry JSON and metrics with fake PID trees where possible.

**Rationale**: Constitution delivery workflow; aligns with FR-010 blank page scope.

## 10. Platform detach semantics

**Decision**: **Linux and WSL2 only** for v1 (FR-014). Detach via `Setsid` and ignore SIGHUP in worker; no macOS/Windows implementation in this feature slice.

**Rationale**: Clarified platform scope; avoids partial cross-platform `SysProcAttr` behavior.

## 11. Headed vs headless on `start`

**Decision**: Mutually exclusive Cobra bool flags `--headed` and `--headless`; **default headed** when neither is set; reject both set with usage exit `2` without registry write (FR-001d). Worker receives the mode and sets chromedp allocator `Headless` accordingly.

**Rationale**: Operator-visible control for manual vs automated runs; tests use `--headless` per constitution.

**Alternatives considered**:

- Headless-only default — rejected in clarification (default headed).
- Single `--headless` with default false — equivalent; two explicit flags chosen for clarity.

## 12. Stop and reconciliation semantics

**Decision**: `stop` unknown ID and already-stopped both exit `0` with distinct messages; empty prompt exit `0` (FR-006a, FR-007a/b). Reconciliation on status/stop: 60s `starting` timeout (FR-013); owner-dead/Chrome-alive → kill Chrome tree (FR-009a); Chrome-dead/owner-alive → kill owner (FR-009b).

**Rationale**: Operator-friendly CLI; clear scriptability without treating benign cases as errors.
