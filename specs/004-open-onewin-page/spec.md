# Feature Specification: Open the onewin page

**Feature Branch**: `004-open-onewin-page`

**Created**: 2026-11-15

**Status**: Draft

**Input**: User description: "use .specify/assessments/casinos-support/decision.md, take into account the new command `screenshot` to the spec too"

The decision's handoff: the operator cannot put an `onewin` session on a page under Supervisor, because `start` opens a blank page. The chosen approach is to open `onewin` with a session the operator already holds. The operator judges the live page privately. The recorded confirmation is a local stand-in, then `status`, `screenshot`, and `stop`. This feature also uses the existing `screenshot` command to capture the page that `start` opened.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Open the onewin page with a session the operator already holds (Priority: P1)

An operator who already has an `onewin` session points `start` at that session and gets the casino's main page under Supervisor. They do not type the page address, and they do not paste the session document into the command. They judge the live page on their own machine. This feature does not record that judgment.

**Why this priority**: A blank page is the gap the decision closes. A screenshot of a blank page is not a picture of the casino.

**Independent Test**: Start a local stand-in page that shows a balance or account information only when the expected session cookie for that local address is present. Point `start` at a session file containing that cookie and confirm the stand-in shows the information. Repeat without the cookie and confirm the information is withheld. Do not open the live casino.

**Acceptance Scenarios**:

1. **Given** the operator has a session file for `onewin` and names that casino, **When** they run `start`, **Then** Supervisor applies every cookie in the file and opens the casino's main page, and `start` exits while the instance keeps running.
2. **Given** a local stand-in that shows a balance only when its expected cookie is present, **When** the ordinary confirmation runs `start` with a session file whose cookie is for that local address, **Then** the stand-in shows the balance.
3. **Given** the same local stand-in, **When** the ordinary confirmation runs `start` without that cookie, **Then** the stand-in does not show the balance.
4. **Given** the operator is looking at the live casino page, **When** they decide whether the session is active, **Then** that decision stays on their machine and Supervisor stores no record of it.

---

### User Story 2 - Capture the page that start opened (Priority: P2)

An operator runs the existing `screenshot` command with the instance ID from `start` and receives one image of the page that instance is showing. For an `onewin` start, that page is the casino page, not a blank page. The instance keeps running. In the ordinary confirmation, the image is of the local stand-in.

**Why this priority**: The screenshot command already captures a running instance. This feature is done only when that image is of the page the session opened.

**Independent Test**: After the stand-in start in User Story 1, run `screenshot` with that instance ID. Confirm one image of the stand-in page, a successful exit within 15 seconds, and a following `status` that still lists the same instance as running. When the cookie was present, the captured page shows the balance. When it was absent, the captured page does not. Do not open the live casino.

**Acceptance Scenarios**:

1. **Given** an `onewin` instance that `status` shows as running on the casino page, **When** the operator runs `screenshot` with that instance ID and no output path, **Then** within 15 seconds the command writes one image of that page, prints that file's location as a single line, and exits successfully.
2. **Given** the ordinary confirmation opened the local stand-in with the expected cookie, **When** `screenshot` runs for that instance, **Then** the image shows the balance, and the instance is still running on that page.
3. **Given** the ordinary confirmation opened the local stand-in without the expected cookie, **When** `screenshot` runs for that instance, **Then** the image does not show the balance, and the instance is still running.
4. **Given** the operator chooses an output path that does not exist yet, **When** they run `screenshot` with that path, **Then** the image is written only at that path and no extra default file is created.
5. **Given** the operator runs `screenshot` with an unknown ID, or with an instance that is not running, **When** the command finishes, **Then** it fails, creates no image, and does not change the instance.

---

### User Story 3 - See the run and end it (Priority: P3)

An operator checks that the `onewin` instance is running and then ends it. When the instance ends, that run's browser data is gone, including after a failed start. `screenshot` does not end the run.

**Why this priority**: A session left on disk after the run is the harm this story prevents. `status` and `stop` are how the operator sees and ends the run.

**Independent Test**: Run `start` against the local stand-in, run `status` and confirm the instance is listed, run `screenshot` and confirm the instance is still listed, run `stop` for that instance, confirm `status` no longer lists it, and confirm that run's browser data is gone.

**Acceptance Scenarios**:

1. **Given** an `onewin` instance is running, **When** the operator runs `status`, **Then** that instance is listed as running.
2. **Given** an `onewin` instance is running, **When** the operator runs `stop` for it, **Then** the instance ends, `status` no longer lists it, and that run's browser data is removed.
3. **Given** `start` fails after creating browser data for the attempt, **When** the failure is reported, **Then** no instance remains and the browser data for that attempt is removed.
4. **Given** Supervisor treats an instance as finished because the browser exited outside `stop`, **When** that instance is cleared, **Then** its browser data is removed as well.

---

### User Story 4 - Keep the session out of the project and the command (Priority: P4)

An operator supplies the session as a file path. The session values never become part of the command text, the run's recorded output, the list of instances, the project, or the screenshot file's name.

**Why this priority**: The session is a secret the operator already holds. This feature uses it at runtime and does not keep a copy.

**Independent Test**: Run `start` and `screenshot` with a session file that contains a known test value for the local stand-in. Confirm that value is absent from the command text of `start`, of the background run, and of `screenshot`, from the run's recorded output, from the instance list, from the project, and from the image file's name. The session file path may appear.

**Acceptance Scenarios**:

1. **Given** a session file with a known test value, **When** the operator runs `start` by passing the file path, **Then** the session document is not part of the command text.
2. **Given** that run is in progress or has ended, **When** the run's recorded output and the instance list are inspected, **Then** the test value appears in neither place.
3. **Given** `screenshot` wrote an image, **When** the file name and the command text are inspected, **Then** the test value appears in neither place.
4. **Given** the feature is complete, **When** the project contents are inspected, **Then** no session value from a run, and no screenshot of the live casino, has been added to the project.

---

### Edge Cases

- The session file is missing, unreadable, or not a session document. `start` fails with a clear message, leaves no instance, and leaves no browser data for the attempt.
- The session file contains no cookies. `start` fails the same way.
- The session file contains more than one cookie. All of them are applied before the page opens.
- The operator names a casino other than `onewin`, or supplies `Onewin` or `1win.com`. `start` refuses. A host or a page address is not a casino identifier, and the operator does not type the page address on `start`.
- The operator passes the session document itself instead of a file path. `start` does not take the document that way.
- The ordinary confirmation uses a cookie for the local stand-in's address only. It does not open the live casino, including when a file meant for the live site is present on the machine.
- The operator asks to run `aviator`, Spaceman, Dino, or another casino. This feature does not start them.
- The operator starts headless and cannot see the live page. `start` still applies the session and opens the page. `screenshot` still captures that page. Judging the live page by eye requires a visible window.
- `screenshot` does not stop the instance and does not remove its browser data. `stop` does.
- A screenshot of an instance that is still starting fails because the instance is not running, and it creates no image.
- The image may show whatever the page is showing, including a balance. Supervisor does not redact it, store a gallery, or send it anywhere. An image of the live casino is the operator's file and is not added to the project.
- `stop` for an instance that is already gone keeps today's result for that command. Browser data belonging to an instance this feature has ended is still removed.
- A `start` with no casino still opens a blank page. `screenshot` of that instance remains the existing command's capture of that page.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `start` MUST accept the casino identifier `onewin` and a path to a session file. It MUST NOT accept the session document as the argument text.
- **FR-002**: The page address for `onewin` MUST be the address registered with that casino, `https://1win.com/`. The operator MUST NOT have to type that address on each `start`.
- **FR-003**: `start` MUST apply every cookie in the session file before the page opens, and the page that opens MUST be that casino's main page. A blank page MUST NOT be the result of a successful `onewin` start.
- **FR-004**: A session file MUST be able to carry more than one cookie. Each cookie supplies its name, its value, and the attributes the browser needs in order to store it for the page.
- **FR-005**: Session values MUST be supplied only at runtime. They MUST NOT appear in the command text of `start`, of the background run, or of `screenshot`, in the run's recorded output, in the instance list, in the project, or in the screenshot file's name.
- **FR-006**: `status` MUST list a running `onewin` instance while that instance is running, using the same lifecycle the operator already has for other instances.
- **FR-007**: When an instance ends, including a failed `start` and a browser that exits outside `stop`, Supervisor MUST remove that run's browser data so the session material does not remain. `screenshot` MUST NOT end the instance and MUST NOT remove that data.
- **FR-008**: This feature MUST register exactly one casino, `onewin`, and exactly one game for it, `aviator`. It MUST NOT start `aviator`, Spaceman, Dino, or any other casino.
- **FR-009**: Casino and game identifiers MUST be one or more lowercase letters. `onewin` and `aviator` are the identifiers this feature accepts. `Onewin` and `1win.com` MUST be refused.
- **FR-010**: The ordinary confirmation MUST start a local stand-in bound to the local machine, run `start`, then `status`, then `screenshot`, then `stop`, and MUST NOT open the live casino. The stand-in MUST show a balance or account information only when the expected session cookie for that local address is present, and MUST withhold it otherwise. The screenshot in that confirmation MUST be of that stand-in page. The default test run MUST include this confirmation.
- **FR-011**: `screenshot` MUST keep the behavior already specified for that command: one instance ID, one image of the page that instance is currently showing, a printed file location, success or failure within 15 seconds, no prompt, and an optional output path that is the only file written when the operator supplies one. For an `onewin` instance, the page in the image MUST be the page `start` opened.
- **FR-012**: Supervisor MUST NOT store the operator's judgment of the live page and MUST NOT add a screenshot of the live casino to the project. Completing the ordinary confirmation MUST NOT require a recorded live login.
- **FR-013**: Headed and headless `start` MUST keep their current meaning. The default remains a visible window so the operator can judge the live page. `screenshot` MUST be able to capture the page either way.

### Key Entities

- **Casino**: The registered site `onewin`. It has one main-page address, `https://1win.com/`, and one game, `aviator`. The address is not the identifier.
- **Game**: `aviator`, the game registered for `onewin`. This feature names it and does not start it.
- **Session file**: A file the operator already holds, passed to `start` by its path. It contains one or more cookies. It is not copied into the project.
- **Cookie**: A name, a secret value, and the attributes required to store it for a page (where it applies, the path, when it expires, and whether it is secure, hidden from page scripts, same-site restricted, or partitioned). The value exists only at runtime.
- **Browser instance**: One supervised browser for one run, with its own browser data. That data is removed when the instance ends. `screenshot` leaves the instance running.
- **Screenshot**: One image of that instance's current page, produced by the existing `screenshot` command. It is an operator file, not a project record.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator with a valid session file opens the `onewin` page in one `start`, without typing the session document or the page address. The baseline is a blank page and zero casino pages.
- **SC-002**: The ordinary confirmation shows the local balance when the expected local cookie is present and withholds it when the cookie is absent. That confirmation completes `start`, `status`, `screenshot`, and `stop` with zero requests to the live casino. The screenshot shows the same balance outcome as the page.
- **SC-003**: For a running `onewin` instance, `screenshot` returns within 15 seconds with one image of the page `start` opened, and the same instance is still running afterward.
- **SC-004**: After every ended or failed run in that confirmation, the run's browser data is gone, and the known test secret is absent from the project, the command text, the run's recorded output, the instance list, and the screenshot file's name.
- **SC-005**: The operator can judge the live page on their own machine. The product stores no record of that judgment and no screenshot of the live casino.

## Assumptions

- The only stated demand is one operator. No other operator or ticket volume was found.
- A file of cookies the operator already holds is enough to open the page. This feature adds no sign-in step.
- The registered page address for `onewin` is `https://1win.com/`, taken from the request that raised this feature. Whether the live site accepts a particular file was not observed. If a real file is rejected, this feature stops for a new decision. It does not grow into a search for a working session, and it does not start `aviator` to get past the rejection.
- The operator's permission for an automated session is their own statement. The site's terms were not read. A terms conflict stops this feature.
- Browser data is removed after the browser for that instance has exited, so removal does not race a browser that is still writing the data.
- The local stand-in proves that a session cookie was applied and that `start`, `status`, `screenshot`, and `stop` finish. It proves nothing about `1win.com`.
- The operator named a private on-page signal for their own look at the live site. That signal was not observed. It is not a product check and not a stored result.
- Constitution 1.2.0 allows `onewin` to be supported before `aviator` can be run. This feature names `aviator` and does not start it.
- Linux and WSL2 remain the supported operator environments, as in the existing CLI feature. Other operating systems stay out of scope.
- The `screenshot` command is the one already specified in `specs/003-screenshot-cli`. This feature does not rename it, change its arguments, or change its failure rules. It requires that the captured page be the page `start` opened.
- A `start` that names no casino keeps today's blank page. Capturing that blank page stays the existing screenshot behavior.
- The assessment recorded that the default test run starts the local stand-in itself. This specification requires the behavior of that stand-in. The plan chooses the tool.
