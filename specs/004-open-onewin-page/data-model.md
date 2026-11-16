# Data Model: Open the onewin page

Entities for `specs/004-open-onewin-page/spec.md`. Field rules that the spec states are repeated here so implementation does not invent them.

## Casino

Registered in code. Not read from the operator's file.

| Field | Rule |
|-------|------|
| Identifier | `onewin`. One or more lowercase ASCII letters. `Onewin` and `1win.com` are refused. |
| Page | `https://1win.com/`. The operator does not type it. |
| Game | `aviator`. Named. This feature has no command that starts it. |

One catalog entry. Any other identifier is a usage error.

## Game

| Field | Rule |
|-------|------|
| Identifier | `aviator`. One or more lowercase ASCII letters. |
| Casino | `onewin`. |

No runtime state. Nothing in the registry records a game run.

## Session file

A file outside the repository. `start` receives its path. The document is not an argument and is not copied into the project, the registry, or the worker log.

| Field | Rule |
|-------|------|
| `cookies` | Array. Length at least 1. Zero cookies fails `start` with no instance and no profile left behind. |

See [contracts/session-file.md](./contracts/session-file.md).

## Cookie

One element of `cookies`. The value is a secret and exists only in the operator's file and in the browser while the instance runs.

| Field | Required | Rule |
|-------|----------|------|
| `name` | yes | Non-empty string. |
| `value` | yes | Non-empty string. Never written to the registry, the command line, the worker log, or an error string. |
| `domain` | no | String. Passed through when present. |
| `path` | no | String. Passed through when present. |
| `expires` | no | RFC3339 timestamp. When set, this is the browser expiry. |
| `maxAge` | no | Integer seconds. Used only when `expires` is absent. Expiry is the time of the read plus `maxAge`. |
| `secure` | no | Boolean. |
| `httpOnly` | no | Boolean. |
| `sameSite` | no | `Strict`, `Lax`, or `None`. Any other value fails the parse. |
| `partitioned` | no | Boolean. When true, the browser cookie is partitioned. |

Unknown fields fail the parse. Every cookie in the array is applied before navigation. A `SetCookie` error fails the instance.

## Browser instance

Extends the feature 001 instance. Schema version stays `1`. The new field is optional, so existing rows still load.

| Field | Rule |
|-------|------|
| `id` | Unchanged. Allocated by `start`. |
| `state` | Unchanged: `starting`, then `running`. Rows leave the registry when the instance ends. |
| `created_at` | Unchanged. |
| `profile_dir` | Unchanged path. The directory is removed after the browser process has exited. |
| `owner_pid`, `browser_pid`, `log_path`, `headless`, `last_seen_at` | Unchanged. |
| `casino` | Optional string. `onewin` for this feature. Omitted on a blank `start`. Not a secret. |

The instance row has no cookie, no session file path, and no copy of the page address. A screenshot does not add a field and does not change `state`.

## Screenshot

One file produced by the existing `screenshot` command. It is not a registry row and it is not committed.

| Field | Rule |
|-------|------|
| Instance | The ID `start` printed. The row must be `running`. |
| Page | Whatever that instance is showing. For `onewin`, that is the page opened after the cookies were set. |
| File | The path the existing command prints. The default name includes the instance ID and the capture time. It must not include a cookie value. |
| Effect | The instance stays `running`. The profile stays until `stop` or another end path. |

## State

```text
blank start (no casino)     → today's starting → running → ended
onewin + valid session file → starting → cookies set → page open → running
running + screenshot          → still running, one image file, profile still present
stop, worker exit, or reap    → processes exited, profile removed, row removed
bad casino or bad flags     → usage error, no row, no profile
bad or empty session file   → error, no row, no profile
cookie or navigation error  → row removed, profile removed
```

Ended includes a failed start after the profile directory was created, and a browser that exits outside `stop`.
