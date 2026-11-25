# Data Model: Instance Watch (018)

No registry schema change. `schema_version` stays 1. Watch state is process-local to the worker and the CLI.

## Instance (unchanged)

| Field | Role for watch |
|-------|----------------|
| `id` | Positional argument. Must be one path segment before it is joined into the socket path (same rule as screenshot). |
| `state` | Must be `running` before dial. Any other tracked state is "not running". No row is "not found". |
| `pid` / browser pid | Must be unchanged by a watch that the operator interrupts (SC-008). |

`watch` does not write the registry and does not reconcile.

## Watch session (ephemeral)

| Attribute | Rule |
|-----------|------|
| Instance | Exactly one running instance |
| Cadence | Integer 1–30 frames per second; default 1 |
| URL | `http://127.0.0.1:<port>/` where port is 8080 or the next free loopback port |
| Viewer | One HTTP client at a time is enough for v1; a second browser may connect but is not a product guarantee beyond "local processes can open the URL" |
| Exclusivity | At most one watch session per instance. A second `watch` fails |
| End | CLI interrupt closes the control connection and stops the stream; the instance stays `running`. Instance leaving `running` closes the control connection; the CLI exits within 5 seconds |
| Persistence | None. No file, no registry field, no recording |

## Live view

| Attribute | Rule |
|-----------|------|
| Subject | Current viewport of that instance's worker tab |
| Headless size | 1280×720 when the instance was started `--headless` (spec 012). Watch does not change viewport |
| Sensitivity | Pixels may include page secrets. No redaction, no upload, no commit |
| Static page | Fewer updates than `--fps` is valid |

## State transitions

```text
running, no watch ──watch──► running + watch session
running + watch ──interrupt──► running, no watch
running + watch ──stop / crash / entry teardown──► row removed or not running; watch ended
starting / stopping / stopped / absent ──watch──► refused, row unchanged
running + watch ──second watch──► refused, first session unchanged
running + watch ──screenshot──► running + watch, plus one PNG per spec 003
```

A watch failure before `OK` does not change instance state and does not leave an HTTP listener.
