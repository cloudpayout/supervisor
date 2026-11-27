---
name: casino-support
description: Guides adding or extending casino support in Supervisor (catalog, session cookies, worker navigation, integration tests). Use when supporting a new casino, opening a casino home page with a session, mirroring onewin or brazino, or working from specs/004-open-onewin-page or specs/020-open-brazino-page.
---

# Casino support in Supervisor

## What “support” means

An operator opens a **registered casino page** under Supervisor using a **session file they already hold**. Supervisor applies cookies, navigates to the catalog URL (home or play, per casino), and keeps session secrets **only at runtime**. Default CI proves behavior with a **local HTTP stand-in** on `127.0.0.1`, not the live site.

Normative examples:

| Casino   | Spec directory                    | Home / start URL              | Game in catalog      |
|----------|-----------------------------------|-------------------------------|----------------------|
| `onewin` | `specs/004-open-onewin-page/`     | `https://1win.com/` (play via `PlayURL`) | `aviator` (named; page slice did not run it) |
| `brazino`| `specs/020-open-brazino-page/`    | `https://www.brazino777.mx/`  | none (`game: null` in `session info`) |

Assessment handoff: `.specify/assessments/casinos-support/decision.md` (Option B — open with session, local proof, live judgment private).

## Operator surface (unchanged across casinos)

1. `supervisor start <session-file>` — parse file, spawn worker, apply cookies, navigate.
2. `supervisor status` — instance listed while running.
3. `supervisor screenshot <id>` — image of the page `start` opened; instance stays running.
4. `supervisor stop <id>` — end instance; profile removed after browser exit.
5. `supervisor session info <id>` — `casino` and `game` from catalog (empty game → `game: null`).

Blank `start` (no session path) still opens `about:blank`.

**Test seam only (not operator API):** `SUPERVISOR_PAGE_URL` redirects navigation to the stand-in.

## Session file rules

- Path on disk; JSON document is **not** pasted into argv.
- Required: `casino` (lowercase catalog id), `cookies` (length ≥ 1).
- Cookie schema: see `specs/004-open-onewin-page/contracts/session-file.md` (and 020’s copy). Unknown JSON fields fail parse.
- `casino` in the file must match `casino.Lookup`. Hostnames and mixed case are **refused** as identifiers.

## Code touchpoints

```
internal/casino/catalog.go   # Lookup, StartURL, PlayURL (onewin game path)
internal/session/file.go     # parse + validate casino field
internal/worker/run.go       # set cookies, openTarget → StartURL or stand-in
internal/registry/types.go  # optional casino on instance row (not secret)
internal/cli/start.go        # spawn worker with --session-file path only
```

**Navigation rule (020):** `StartURL(casinoID, locale)` — if catalog `Game` is empty, open registered `Page`; if game is set (onewin), use play URL logic. Worker must not send `brazino` to an Aviator path.

**Secrets (FR-005):** Cookie values never in registry, worker log, error text, screenshot filename, or repo. Profile dir under `$XDG_DATA_HOME/supervisor/profiles/<id>/` removed when the instance ends (including failed start and external browser exit).

## Spec shape (copy from 004, extend like 020)

Each casino slice gets `specs/<NNN>-open-<name>-page/` with:

- `spec.md` — four user stories (P1 open, P2 screenshot, P3 status/stop, P4 secrets) plus edge cases.
- `plan.md` — catalog entry, worker behavior, constitution check, file list.
- `data-model.md`, `contracts/cli.md`, `contracts/session-file.md`, `quickstart.md`, `tasks.md`.

For a **second and later casino**, spec must be **additive only** (020 FR-015): do not change 004 artifacts; SC-006 requires all `onewin` tests still pass; **one shared** `newStandIn` helper (`tests/integration/onewin_helpers_test.go`), not a per-casino stand-in type.

Clarify up front:

- Registered **home URL** for the identifier (operator choice, recorded in clarifications).
- Whether a **game** is named in catalog without being runnable (constitution V).

## Implementation workflow

Follow task phases from the reference specs (tests often written before green implementation).

### Phase A — Catalog and navigation (blocking)

1. Add `Lookup` case: id `^[a-z]+$`, page URL, optional `Game`.
2. Ensure `StartURL` / worker `openTarget` use home vs play correctly.
3. Unit tests in `tests/unit/catalog_test.go` and `tests/unit/session_file_test.go`.

### Phase B — User story 1 (open page)

Integration tests mirroring `onewin_start_test.go` or `brazino_start_test.go`:

- `newStandIn` on ephemeral `127.0.0.1` port; shows marker `balance` only when expected cookie present.
- Temp session file; cookie `domain` = `127.0.0.1`.
- `SUPERVISOR_PAGE_URL` → stand-in URL.
- `start --headless <file>`; assert shown vs withheld balance.
- Usage tests without Chrome for bad ids / empty cookies / missing file.

### Phase C — User stories 2–4

| Story | Test file pattern              | Product change        |
|-------|--------------------------------|-----------------------|
| P2    | `<casino>_screenshot_test.go`  | Usually none (`screenshot` unchanged) |
| P3    | `<casino>_stop_test.go`        | Profile removal paths if not already covered |
| P4    | `<casino>_secrets_test.go`     | Audit argv, log, registry, PNG name |

Optional: `<casino>_session_info_test.go` when catalog has no game.

### Phase D — Polish

- Align `quickstart.md` with contracts.
- `go test ./...` — no navigation to live casino hosts (SC-007 for multi-casino).
- FR-style spec citations on new/changed production code.

## Ordinary confirmation checklist

Use for spec quickstart and when validating a new casino:

```
- [ ] Stand-in shows balance only with expected cookie
- [ ] start → status (running) → screenshot (≤15s, pixel matches) → status (still running) → stop → status (gone), profile gone
- [ ] Withheld cookie run: no balance, different pixel
- [ ] Generated secret absent from worker argv, screenshot argv, worker log, registry, screenshot filename
- [ ] Live casino URL never requested in default test run
```

Live page check stays **manual**, headed, outside `go test` (see each spec’s quickstart).

## Adding a third casino (template)

1. Run assessment / specify: problem, URL, in/out of scope, success metrics (copy decision.md handoff style).
2. Create `specs/<NNN>-open-<id>-page/` by cloning 020 structure; reference 004 for stand-in and secrets patterns.
3. Implement Phase A–D; prefix integration tests `<id>_*.go`.
4. Assert existing `TestOneWin*`, `brazino_*`, and Aviator tests unchanged unless a spec explicitly amends them.

## Non-goals (do not expand the slice)

- No sign-in automation, no hunting for working live sessions if cookies fail.
- No committing operator session files or live screenshots.
- No second stand-in implementation for default suite.
- No starting games unless a **separate** game feature spec says so (e.g. Aviator is its own track).

## Additional resources

- Detailed FR list and edge cases: `specs/004-open-onewin-page/spec.md`, `specs/020-open-brazino-page/spec.md`
- Task ordering and file paths: respective `tasks.md` and `plan.md`
- Implementation deltas (004 vs 020 CLI): [reference.md](reference.md)