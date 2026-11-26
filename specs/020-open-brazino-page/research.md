# Research: Open the Brazino page

## Decision: Additive catalog entry with home-page navigation only

**Decision**: Register `brazino` with page `https://www.brazino777.mx/` and an empty `Game` field. Worker opens the registered **home page** for casinos with no game; `onewin` keeps opening the Aviator play URL via existing `PlayURL(locale)`.

**Rationale**: Spec FR-002 and clarifications fix the Mexico home URL. Today's `openTarget` calls `casino.PlayURL`, which is hard-wired to `onewin` play paths and would send `brazino` to the wrong URL. Splitting navigation by catalog entry preserves 004 behavior without duplicating worker logic.

**Alternatives considered**: Reusing `PlayURL` for all casinos was rejected because Brazino has no play path in this slice. Registering a fake game name was rejected (constitution: no game until explicitly supported).

## Decision: Session file carries `casino: "brazino"`

**Decision**: Reuse the session format from later features (005+): required top-level `casino` must match a supported `Lookup` id. `start` takes a positional file path, not `--casino` (004 contract flags were superseded by `specs/005-start-command-syntax`). Operator outcome matches spec FR-001: only `brazino` sessions open the Brazino page.

**Rationale**: `internal/session/file.go` already enforces `casino` and `casino.Lookup`. Extending `Lookup` is sufficient for parse-time refusal of unknown casinos.

**Alternatives considered**: CLI `--casino brazino` was rejected; current product uses file-only casino id. Duplicating casino on flag and file was rejected in 004 research.

## Decision: Reuse the shared Go stand-in helper

**Decision**: Integration tests call the existing `newStandIn` in `onewin_helpers_test.go` with `brazino` session files. No second stand-in package or Brazino-dedicated HTTP server implementation. SC-007 means one helper type for both casinos, not one long-lived process shared across parallel tests.

**Rationale**: Clarification session 2026-11-26 chose Option B. The helper already implements balance shown/withheld on cookie presence.

**Alternatives considered**: A Brazino-only copy of the stand-in was rejected (duplicate maintenance). Folding both casinos into 004's spec directory was rejected (separate feature id 020).

## Decision: Keep `SUPERVISOR_PAGE_URL` test seam

**Decision**: Unchanged from 004: when set, worker navigates to that URL (with optional locale path rewrite for stand-in tests). Unset, `brazino` navigates to `https://www.brazino777.mx/`.

**Rationale**: Constitution default tests must not hit live casinos. The seam is already implemented in `internal/worker/run.go`.

**Alternatives considered**: Public page override flag rejected (not an operator interface).

## Decision: `session info` game line stays `null`

**Decision**: No registry `game` field. `session_info` reads catalog `Game`; empty string → stdout `game: null`.

**Rationale**: Spec FR-014 and constitution 1.2.0. Code path already exists for casinos without a named game.

**Alternatives considered**: Omitting the game line was rejected (019 contract requires four lines).

## Decision: Non-regression on `onewin`

**Decision**: No edits to `onewin` catalog fields, play URL rules, or Aviator integration tests except any mechanical rename of shared test helpers. New code paths branch on casino id.

**Rationale**: Spec FR-015 and SC-006.

**Alternatives considered**: Refactoring 004 tests into generic `casino_test` in the same change was rejected as scope creep; optional follow-up after green `brazino` tests.

## Decision: Rejected live-site verification

**Decision**: If an operator's live cookie fails on `www.brazino777.mx`, stop for a new decision; do not expand scope to hunt sessions or add regional domains.

**Rationale**: Same boundary as 004 spec assumptions.

**Alternatives considered**: Default test against `.bet.br` rejected (clarification chose `.mx` only).