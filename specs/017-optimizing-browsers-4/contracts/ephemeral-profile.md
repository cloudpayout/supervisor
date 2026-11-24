# Contract: ephemeral instance profiles (017)

**Spec**: [spec.md](../spec.md) **FR-012**, **FR-013**, **SC-006**, **SC-007**

## Rule

Each instance gets a **new** directory under the supervisor profiles layout at `start`. The directory exists only while the instance is active (starting → running → stopping). After normal teardown, it MUST be deleted.

## Cleanup paths (MUST call `browser.RemoveProfile`)

| Event | Location |
|-------|----------|
| Worker exits (normal or error) | `internal/worker/run.go` defer |
| `supervisor stop <id>` | `internal/cli/stop.go` |
| Failed `start` after profile create | `internal/cli/start.go` rollback |
| Dead instance reconcile | `internal/registry/reconcile.go` |

## Non-goals

- Wiping profile **during** a long `running` session (in-run growth is allowed; cache limits come from Chrome flags).
- Persisting cookies/storage across instance IDs (session files rehydrate cookies at start as today).

## Verification

- **SC-006**: Integration test start → stop → `os.Stat(profileDir)` is not found.
- **SC-007**: **N=10** consecutive start/stop headless; zero leftover profile dirs for those IDs when `status` is empty.
- **FR-012**: Integration test two sequential headless starts (after full stop of first) use **different** profile directory paths; second start must not reuse first instance's path.