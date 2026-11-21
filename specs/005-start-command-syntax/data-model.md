# Data Model: Start command syntax

Entities for `specs/005-start-command-syntax/spec.md`. Feature 004 rules stay unless this feature overrides them.

## Session file

JSON on disk. Read when `start` receives a positional path. Not copied into the project, registry, or worker log.

| Field | Rule |
|-------|------|
| `casino` | Required string when a path is passed. Must pass `casino.Lookup` (`onewin` only today). Missing, empty, wrong case, URL-shaped values, and unknown casinos fail parse before an instance is registered. |
| `cookies` | Required array, length ≥ 1. Same cookie objects and validation as feature 004. |
| `currency`, `demo`, `locale` | Optional. Added by feature 007. Omitted keys keep the 004/005 document valid. Rules: [007 session-file contract](../007-enter-aviator-onewin/contracts/session-file.md). |

Shape for `onewin`:

```json
{
  "casino": "onewin",
  "cookies": [ ... ]
}
```

Unknown top-level fields are errors. Cookie secrets never stored in registry.

See [contracts/session-file.md](./contracts/session-file.md).

## Cookie

Unchanged from feature 004. One element of `cookies`.

## Casino

Unchanged catalog (`internal/casino`). Identifier from session file `casino` field, not from `start` argv. Registry `casino` optional string on instance row; omitted for blank start.

## Browser instance

Extends feature 004 row.

| Field | Rule |
|-------|------|
| `casino` | Set from session file when positional path present. Omitted when blank `start`. |
| `headless` | Optional pointer; set when operator passed `--headless` on `start`. |
| Other fields | Unchanged from 001/004. |

## CLI invocation (logical, not registry)

| Form | Page | Window |
|------|------|--------|
| `start` | `about:blank` | headed |
| `start --headless` | `about:blank` | headless |
| `start <path>` | catalog page for file's casino | headed |
| `start <path> --headless` | same | headless |
| `start --headless <path>` | same | headless |

Rejected: `--session-file`, `--casino`, `--headed`, `-h`, extra positionals, missing file on disk.

## State

```text
no path              → blank page, no casino on row
positional + valid file → parse casino + cookies → starting → running (casino page)
positional + missing/bad file → error, no row, no profile
usage error (bad flags) → error, no row, no profile
```

Ended paths unchanged from 004 (stop, worker exit, failed open).