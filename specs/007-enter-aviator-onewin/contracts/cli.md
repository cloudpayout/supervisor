# Contract: CLI (007 amendment)

**Feature**: `007-enter-aviator-onewin`  
**Amends**: [004 CLI](../../004-open-onewin-page/contracts/cli.md), [005 CLI](../../005-start-command-syntax/contracts/cli.md)

## `start`

Syntax unchanged from 005: `start <session-file> [--headless]`.

For `onewin` session files:

- `start` exits after instance registration and worker spawn (may exit before **Aviator ready**).
- Worker performs Aviator entry in the background.
- Exit code on spawn/parse failures unchanged; successful spawn still exits `0` even if entry later fails.

## `status`

For `onewin` instances with Aviator entry, output includes `entering` or `ready` when the field is set. No failure category on teardown.

## `screenshot`, `stop`

Unchanged. Capture reflects current page (modal, partial load, or post-ready).

## Environment

| Variable | Effect |
|----------|--------|
| `SUPERVISOR_PAGE_URL` | Overrides navigated URL for play/entry (integration tests only; undocumented operator interface) |

## External / live validation

Live `1win.com` entry is manual per [quickstart.md](../quickstart.md). Not part of `go test ./...`.