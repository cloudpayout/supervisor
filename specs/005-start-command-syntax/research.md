# Research: Start command syntax

**Feature**: `005-start-command-syntax`  
**Spec**: [spec.md](./spec.md)

## CLI surface (Cobra)

**Decision**: Operator `start` accepts zero or one positional path and one optional boolean flag `--headless`. Remove `--headed`, `--casino`, and `--session-file` from the operator command. Unknown flags remain Cobra usage errors (exit `2`).

**Rationale**: Matches FR-001 and clarifications. Fewer flags reduce coupling between casino name and command text.

**Alternatives considered**: Keep `--session-file` as hidden alias — rejected in clarification. Keep `--casino` when file lacks `casino` — rejected; casino is only in the file.

## Positional path and arity

**Decision**: `cobra.MaximumNArgs(1)` on `start`. Session path is `args[0]` when `len(args)==1`. Omitted path is a blank start. More than one positional argument is a usage error.

**Rationale**: Spec allows `start --headless <path>` with a single positional; Cobra parses flags before args.

**Alternatives considered**: Optional flag for path only — rejected in clarification.

## Display mode

**Decision**: Drop `--headed`. Default is headed (`useHeadless := headless` where `headless` defaults false). Only `--headless` selects windowless mode.

**Rationale**: FR-006; visible window is default without a redundant flag.

**Alternatives considered**: Mutual exclusion of headed/headless — withdrawn with `--headed`.

## Help: `--help` vs `-h`

**Decision**: On the `start` command only, clear the shorthand on the default help flag so `-h` is not accepted. `supervisor start --help` uses Cobra help and must not run `RunE` (no instance). `supervisor start -h` fails as an unknown shorthand flag (usage error, exit `2`).

**Rationale**: Clarification session 2026-11-16. Matches FR-007 and SC-003.

**Alternatives considered**: Accept both `-h` and `--help` — rejected. Route all help to `supervisor --help` — rejected.

## Session file document shape

**Decision**: Extend `internal/session` parsing with required top-level `casino` (string) beside `cookies`. `Parse` (or `ParseFile`) returns casino identifier and cookies. Unknown JSON fields stay errors (`DisallowUnknownFields`). Cookie rules unchanged from feature 004.

**Rationale**: FR-003, FR-005. Casino validation uses existing `casino.Lookup`.

**Alternatives considered**: Keep casino only on CLI — rejected. Infer casino from cookie domain — rejected (constitution: identifier not URL).

## Start pipeline

**Decision**:

1. If unknown/extra flags (including `--session-file`, `--casino`, `--headed`, `-h`): usage error before preflight.
2. If `len(args)==0`: blank start (no parse, no `casino` on registry row).
3. If `len(args)==1`: `os.Stat` or read via parse; missing path fails before registry (no id, no profile).
4. Parse file; `casino.Lookup`; preflight; registry append with `casino` set from file; spawn worker with hidden `--session-file <path>` unchanged.

**Rationale**: Reuses 004 worker spawn. Operator path may appear on `start` argv; JSON never does (FR-009).

**Alternatives considered**: Pass casino id to worker as a new flag — unnecessary; registry row already has `casino`.

## Worker

**Decision**: Hidden `supervisor worker run` keeps `--session-file` for the path. Worker continues to read cookies from the file and page from registry `casino` (set by `start`). Remove worker default that forces `onewin` when registry casino is empty but a file is present — casino must come from parse in `start` before registry write.

**Rationale**: Internal contract unchanged for worker argv; operator contract changes only on `start`.

**Alternatives considered**: Positional path on worker — out of scope; not operator-facing.

## Tests migration

**Decision**: Update `tests/integration/onewin_*` and related usage tests to use `start <path> --headless` and new rejection table (`--session-file`, `--casino`, `--headed`, `-h`). Blank start tests keep `start` and `start --headless` with no args.

**Rationale**: FR-011 ordinary confirmation command line change.

**Alternatives considered**: Dual support in tests for old flags — rejected; old flags must fail per spec.