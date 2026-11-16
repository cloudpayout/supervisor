# Data Model: Operator CLI command catalog

**Feature**: `006-completion-command` | **Date**: 2026-11-16

This feature does not add registry rows, session files, or runtime state. It documents the **operator CLI command catalog** the binary may expose at the root.

## Entity: RootCommandCatalog

The set of named subcommands an operator may discover via root help and invoke as `supervisor <name>`.

| Field | Type | Rules |
|-------|------|-------|
| `binary_name` | string | Always `supervisor` |
| `operator_commands` | ordered set of names | MUST include `start`, `status`, `stop`, `screenshot` |
| `builtin_commands` | set of names | MUST include Cobra `help` |
| `excluded_commands` | set of names | MUST include `completion` (MUST NOT be registered or advertised) |
| `hidden_commands` | set of names | MUST include `worker`; MUST NOT appear in root help |

### Validation rules (FR-001 … FR-006)

1. `completion ∉ operator_commands ∪ builtin_commands ∪ hidden_commands` as an executable subcommand (not registered).
2. Root help text MUST enumerate exactly the operator commands + `help` (no `completion`, no `worker`).
3. Attempting `supervisor completion` or `supervisor completion <shell>` MUST NOT mutate catalog or emit scripts; outcome is unknown-command / usage error (exit 2).
4. Catalog for `start`, `status`, `stop`, `screenshot` unchanged from latest amended contracts (005 for `start`, 002 for `status`/`stop`, 003 for `screenshot`).

## Relationships

```text
Operator
  └── invokes supervisor binary
        └── RootCommandCatalog
              ├── operator_commands → lifecycle / capture (001–005 contracts)
              ├── excluded_commands → completion (removed by 006)
              └── hidden_commands → worker (internal only)
```

## State transitions

None. The catalog is static at build time; the only change is removal of the Cobra-injected `completion` entry.