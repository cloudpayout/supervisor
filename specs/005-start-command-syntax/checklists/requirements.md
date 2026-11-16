# Specification Quality Checklist: Start command syntax

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-16
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- Validation passed on the first pass (2026-11-16). Re-checked after clarify (2026-11-16): positional session path, optional blank `start`, casino in file only.
- Operator `start` accepts an optional positional path to the session file, or no path for a blank page. `--session-file` and `--casino` on the command are usage errors. `--headed` is withdrawn. A visible window is the default unless `--headless` is passed.
- Casino name is read only from the session file JSON (`casino` field), not from argv. Covered by FR-002, FR-003, and SC-001.
- Cookie attribute rules stay those of the onewin session file. This spec adds a required top-level `casino` name beside `cookies`.
