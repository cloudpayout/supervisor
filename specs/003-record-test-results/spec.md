# Feature Specification: Record One Test Result on Each Change

**Feature Branch**: `003-record-test-results`

**Created**: 2026-11-13

**Status**: Draft

**Input**: User description: "update the current specification to remove those gates, I do not care if changes whose tests failed are merged manually by the code owner."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See whether checks passed (Priority: P1)

An author pushes a change, or opens or updates a pull request, and can tell from that change whether the project's default automated checks passed. The checks run once for that revision. The author does not have to run them locally. This applies on any branch, and it applies to a pull request opened from a fork.

**Why this priority**: Today a change has no recorded result. One visible pass or fail removes that gap. A second run is not required.

**Independent Test**: Push a change on a branch that is not the default branch, and open a pull request from a fork. Each shows one recorded pass or fail for the default checks, without anyone running those checks locally.

**Acceptance Scenarios**:

1. **Given** a change pushed to any branch, **When** the author opens that change, **Then** it shows one recorded pass or fail for the default checks.
2. **Given** a pull request opened from a fork, **When** the author or a reviewer opens it, **Then** it shows the same kind of single recorded result as a pull request from a branch in this repository.
3. **Given** a push that is not yet part of a pull request, **When** the author inspects the push, **Then** a result is still recorded.
4. **Given** the optional ten-instance check and any live-casino check, **When** a result is recorded, **Then** the result does not depend on either of them.
5. **Given** a pull request from a fork, **When** the check runs, **Then** it uses no secrets and cannot change this repository.
6. **Given** a recorded result, **When** the author inspects it, **Then** it was produced with the language version the project constitution requires, on the hosted Linux environment selected for this feature.
7. **Given** a new revision of the same pull request, **When** the result is recorded, **Then** that one execution is of the latest revision, not of an older one.

---

### User Story 2 - Trust a successful result (Priority: P2)

A successful result means the default checks passed once for that revision, and every default check that needs a browser finished as a pass or a fail. A check that was skipped because a browser was unavailable is not a success.

**Why this priority**: One green result is useful when the checks actually ran. A skip hides the answer the author needs. The label is separate from the decision to merge.

**Independent Test**: Run the default checks once with a browser available, and once with it unavailable. Only the run where the checks passed and no browser-dependent check was skipped is recorded as success.

**Acceptance Scenarios**:

1. **Given** one execution of the default checks passed, and every browser-dependent default check finished as pass or fail, **When** the result is recorded, **Then** the change is marked successful.
2. **Given** that one execution failed, **When** the result is recorded, **Then** the change is not marked successful, and no further execution is required.
3. **Given** a browser-dependent default check was skipped, **When** the result is recorded, **Then** the change is not marked successful, even if the other checks passed.
4. **Given** the only way to make a browser-dependent check run would change what an operator does with the product, **When** the result is produced, **Then** operator-facing behavior stays the same and a second execution is not added.

---

### User Story 3 - Leave the merge to the code owner (Priority: P3)

The recorded result is information on the change. It does not decide the merge. A code owner who is allowed to merge may merge a pull request whose checks failed, whose browser-dependent check was skipped, or whose check did not finish.

**Why this priority**: Seeing the result is the need. Holding the merge, and requiring a separate approval when the check itself changes, are out of this feature. The code owner keeps the merge decision.

**Independent Test**: Open a pull request whose check failed, one whose browser-dependent check was skipped, and one whose check never completed. The failure and the skip stay visible as not successful. The check that never completed stays visible as did not run. A code owner merges each one. Open a pull request that changes the check itself. A code owner merges it without a separate approval created by this feature.

**Acceptance Scenarios**:

1. **Given** a pull request whose recorded result is a failure, **When** a code owner merges it, **Then** the merge completes and the failure stays visible on the change.
2. **Given** a pull request whose browser-dependent check was skipped, **When** a code owner merges it, **Then** the merge completes and the result stays not successful.
3. **Given** a pull request whose check did not finish, **When** a code owner merges it, **Then** the merge completes.
4. **Given** a pull request marked successful under User Story 2, **When** a code owner merges it, **Then** the result does not add a merge step.
5. **Given** a pull request that changes the check itself, **When** a code owner merges it, **Then** this feature does not add an approval that must happen first.

---

### Edge Cases

- A check that never starts, or that starts and does not finish, is treated as "did not run." The result is not successful. A code owner may still merge the pull request.
- The single execution fails. The change is not successful. A second execution is not required and must not be added. A code owner may still merge the pull request.
- A browser is missing. Browser-dependent default checks must not be recorded as skipped successes. A code owner may still merge the pull request.
- The optional ten-instance check, and any check that drives a live casino, are outside the result. Their absence does not fail the change, and their presence must not be required.
- A push to a branch other than the default branch still gets a result.
- A pull request from a fork gets a result without using stored secrets and without the check being able to change the repository.
- The author of the change is not the person who merges it. Both can see the same result.
- Making the browser-dependent checks run would require a change to what an operator does with the product. Work on this feature stops rather than making that change.
- A pull request changes the check itself. This feature does not add an approval of that change, and it does not keep the pull request unmerged.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Every push to any branch MUST produce one recorded pass or fail for the project's default automated checks.
- **FR-002**: Every pull request MUST produce that same single recorded result, including a pull request opened from a fork.
- **FR-003**: The recorded result MUST cover the default checks only. It MUST NOT require the optional ten-instance check, and it MUST NOT require a check that drives a live casino.
- **FR-004**: A change MUST be marked successful only when one execution of that revision has passed, and every default check that depends on a browser finished as pass or fail rather than skipped. A skipped browser-dependent default check MUST make the result not successful.
- **FR-005**: An author MUST be able to read the result on the change itself, without running the checks on their own machine.
- **FR-006**: The check MUST NOT require secrets or credentials, and MUST NOT store them.
- **FR-007**: A check for a fork pull request MUST NOT be able to change this repository.
- **FR-008**: The result MUST be produced with the language version the project constitution requires.
- **FR-009**: The result MUST be produced on the hosted Linux environment selected for this feature. macOS, Windows, and a privately operated machine are out of scope.
- **FR-010**: Existing checks MAY be adjusted only so that a browser-dependent default check runs instead of being skipped. Operator-facing product behavior MUST NOT change. A second execution MUST NOT be added.
- **FR-011**: The recorded result MUST leave the merge decision with the people who already have it. A code owner MUST be able to merge a pull request whose result is a failure, a skip, or did not run, and a pull request that changes the check itself, without this feature adding a hold or a required approval.

### Key Entities

- **Check result**: The recorded outcome for one revision of a change. It records the single execution's pass or fail, whether every browser-dependent default check finished rather than skipped, and the overall outcome (successful, not successful, or did not run). The outcome is visible on the change and does not decide the merge.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of pushes and pull requests, on any branch and including forks, show one recorded pass or fail for the default checks. Baseline: none do today.
- **SC-002**: On every result marked successful, 100% of the browser-dependent default checks finished as pass or fail, and none were skipped. Baseline: those checks are skipped when no browser is available, and the run can still look successful.
- **SC-003**: A change is marked successful after 1 passing execution of that revision in which no browser-dependent default check was skipped. A second execution is not required and is not performed.
- **SC-004**: A code owner can merge 100% of pull requests whose result is a failure, a skip, or did not run. The result remains visible. This feature adds no hold and no required approval, including when the pull request changes the check itself.
- **SC-005**: An author who can open the change can tell whether the checks succeeded without running them locally. Baseline: they cannot, because the change has no result.

## Assumptions

- One execution of the default checks is the success rule for this specification. The earlier assessment text that asked for more than one execution is not in force here.
- The hosted Linux environment is the one selected in `.specify/assessments/github-actions/decision.md` (the current hosted Ubuntu label). That label is Linux and satisfies the product's Linux and WSL2 support limit.
- The language version is the version required by `.specify/memory/constitution.md` at ratification, 1.27.1, unless a later spec amends the constitution.
- "Default checks" means the checks that run without the optional ten-instance marker. The browser-dependent ones in that set are the four checks that today skip when no browser is installed. The optional ten-instance check stays out of this result.
- A pull request from a fork can receive a result with no secrets and with no ability for the check to change the repository.
- A push that has no pull request still gets a result.
- "Did not run" means the check never started or did not finish a result. That state is not successful. It does not decide the merge.
- If the single execution cannot run the browser-dependent checks without changing operator-facing product behavior, this feature stops. It does not add a second execution, a billing analysis, or new operator-facing behavior.
- Earlier text in this feature that required a pull request to stay unmerged, or that required a reviewer to approve a change to the check, is not in force. A code owner may merge those changes manually. This feature records the result and leaves that decision alone.
