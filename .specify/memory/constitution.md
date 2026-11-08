<!--
Sync Impact Report
- Version change: unratified template → 1.0.0
- Modified principles:
  - [PRINCIPLE_1_NAME] → I. Spec Before Implementation
  - [PRINCIPLE_2_NAME] → II. CLI as the Operator Interface
  - [PRINCIPLE_3_NAME] → III. Single Shared Browser Session
  - [PRINCIPLE_4_NAME] → IV. Supervisor Owns the Run
  - [PRINCIPLE_5_NAME] → V. Named Games Only
- Added sections:
  - Technical Constraints
  - Delivery Workflow
- Removed sections: none
- Follow-up TODOs: none
- Sources: GitHub README (17fb175), go.mod, .specify/init-options.json.
  This worktree's README is still the stub "# Supervisor".
-->

# Cloudpayout Supervisor Constitution

## Core Principles

### I. Spec Before Implementation

Every behavior change MUST be recorded in a Spec Kit feature spec, plan, and
task list before implementation starts. Code that lands MUST cite the spec
scenario or task it satisfies. A change with no spec trace MUST NOT merge.

Rationale: Spec Kit was adopted as the delivery process before any application
code existed. Specs are the source of truth for what Supervisor does.

### II. CLI as the Operator Interface

Supervisor MUST be operated through its command-line interface. Starting
automations, observing their statuses, and stopping them MUST be possible from that
CLI. No other interface is the supported operator surface.

Rationale: Cloudpayout Supervisor is the official CLI for running game
automations.

### III. Multiple Browser Instances

Supervisor MUST run an isolated browser instance using `chromedp` to ensure 
isolation, privacy and monitoring about the resources being in real-time.

Rationale: Supervisor monitor automations which contains credentials and
it has to monitor the resources being expent in real-time.

### IV. Supervisor Owns the Run

Supervisor MUST start, monitor, and stop each automation run. 
Automations (sessions) MUST live in automation scripts, not in Supervisor's process
control. When Supervisor stops a run, script control of the session MUST end
with it.

Rationale: Supervisor is the monitor and lifecycle owner. Scripts control the
session only for the life of a run Supervisor started.

### V. Named Casinos and Games Only

Supervisor MUST start automations only for casinos and games it explicitly supports. 
Every casino has games that can be run.

Rationale: The product runs automations on registered games from casinos. Growth of that
set is an explicit decision, not an implied default.

## Technical Constraints

- The module path MUST remain `github.com/cloudpayout/supervisor`.
- Implementation language MUST be Go. The Go version MUST match `go.mod`
  (Go 1.27.1 at ratification) unless a spec amends it.
- Browser control MUST use chromedp.
- Feature directories MUST use sequential numbering, and the Spec Kit
  integration MUST remain cursor-agent, as set in `.specify/init-options.json`,
  unless a constitution amendment changes that.
- Credentials, session cookies, and other secrets MUST be supplied at runtime.
  They MUST NOT be committed to the repository.

## Delivery Workflow

- Feature work MUST follow Spec Kit in order: specify, plan, tasks, then
  implement. Clarify and checklist steps MUST be used when a spec is ambiguous
  or a review needs a feature-specific checklist.
- A review MUST block a change that violates a MUST principle in this
  constitution. The spec, plan, or tasks MUST change to comply. The principle
  MUST NOT be reinterpreted to fit the change.
- Automated tests MUST cover CLI start, status, and stop, and the session
  lifecycle those commands control, before that behavior is complete.
- Default tests MUST run without a live casino site. A check that drives a
  real game MUST be marked external and MUST NOT gate the default test run.

## Governance

This constitution supersedes informal practice, README wording, and examples
left in Spec Kit templates. Templates are read at runtime and MUST NOT be
edited to carry principle changes.

Amendments MUST be made in `.specify/memory/constitution.md` and MUST bump
**Version** using semantic versioning:

- MAJOR: a principle is removed or redefined so existing specs no longer comply.
- MINOR: a principle or section is added, or guidance is materially expanded.
- PATCH: wording is clarified without changing an obligation.

Each amendment MUST set **Last Amended** to the amendment date in
`YYYY-MM-DD`. **Ratified** stays the original adoption date. The Sync Impact
Report at the top of the file is review scratch and MUST be removed before the
amendment is committed.

Plans, tasks, and reviews MUST check compliance with this constitution. A MUST
violation is a blocking defect. Extra complexity beyond these principles MUST
be justified in the feature plan.

**Version**: 1.0.0 | **Ratified**: 2026-10-29 | **Last Amended**: 2026-10-29
