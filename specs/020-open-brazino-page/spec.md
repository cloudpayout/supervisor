# Feature Specification: Open the Brazino page

**Feature Branch**: `020-open-brazino-page`

**Feature Directory**: `specs/020-open-brazino-page`

**Created**: 2026-11-26

**Status**: Draft

**Input**: User description: "support Brazino casino, we should open the home page with an authenticated session". Extends `specs/004-open-onewin-page/` with a second casino; default tests reuse one shared local stand-in for both casinos.

The operator cannot put a `brazino` session on a page under Supervisor, because `start` without a supported `brazino` casino still opens a blank page or refuses the casino. The chosen approach is to open `brazino` with a session the operator already holds. The operator judges the live page privately. The recorded confirmation is a local stand-in, then `status`, `screenshot`, and `stop`. This feature uses the existing `screenshot` command to capture the page that `start` opened.

## Clarifications

### Session 2026-11-26

- Q: Which single home page URL should Supervisor register for the casino identifier `brazino` when `start` opens the site? → A: `https://www.brazino777.mx/`
- Q: How should this spec relate to `specs/004-open-onewin-page` for default acceptance testing and the local stand-in? → A: One shared stand-in — a single server/process in the default suite covers both `onewin` and `brazino`.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Open the Brazino home page with a session the operator already holds (Priority: P1)

An operator who already has a `brazino` session points `start` at that session and gets the casino's home page under Supervisor. They do not type the page address, and they do not paste the session document into the command. They judge the live page on their own machine. This feature does not record that judgment.

**Why this priority**: A blank page is the gap this feature closes. A screenshot of a blank page is not a picture of the casino.

**Independent Test**: Start a local stand-in page that shows a balance or account information only when the expected session cookie for that local address is present. Point `start` at a session file containing that cookie and confirm the stand-in shows the information. Repeat without the cookie and confirm the information is withheld. Do not open the live casino.

**Acceptance Scenarios**:

1. **Given** the operator has a session file for `brazino` and names that casino, **When** they run `start`, **Then** Supervisor applies every cookie in the file and opens the casino's home page, and `start` exits while the instance keeps running.
2. **Given** a local stand-in that shows a balance only when its expected cookie is present, **When** the ordinary confirmation runs `start` with a session file whose cookie is for that local address, **Then** the stand-in shows the balance.
3. **Given** the same local stand-in, **When** the ordinary confirmation runs `start` without that cookie, **Then** the stand-in does not show the balance.
4. **Given** the operator is looking at the live casino home page, **When** they decide whether the session is active, **Then** that decision stays on their machine and Supervisor stores no record of it.

---

### User Story 2 - Capture the page that start opened (Priority: P2)

An operator runs the existing `screenshot` command with the instance ID from `start` and receives one image of the page that instance is showing. For a `brazino` start, that page is the casino home page, not a blank page. The instance keeps running. In the ordinary confirmation, the image is of the local stand-in.

**Why this priority**: The screenshot command already captures a running instance. This feature is done only when that image is of the page the session opened.

**Independent Test**: After the stand-in start in User Story 1, run `screenshot` with that instance ID. Confirm one image of the stand-in page, a successful exit within 15 seconds, and a following `status` that still lists the same instance as running. When the cookie was present, the captured page shows the balance. When it was absent, the captured page does not. Do not open the live casino.

**Acceptance Scenarios**:

1. **Given** a `brazino` instance that `status` shows as running on the casino home page, **When** the operator runs `screenshot` with that instance ID and no output path, **Then** within 15 seconds the command writes one image of that page, prints that file's location as a single line, and exits successfully.
2. **Given** the ordinary confirmation opened the local stand-in with the expected cookie, **When** `screenshot` runs for that instance, **Then** the image shows the balance, and the instance is still running on that page.
3. **Given** the ordinary confirmation opened the local stand-in without the expected cookie, **When** `screenshot` runs for that instance, **Then** the image does not show the balance, and the instance is still running.
4. **Given** the operator chooses an output path that does not exist yet, **When** they run `screenshot` with that path, **Then** the image is written only at that path and no extra default file is created.
5. **Given** the operator runs `screenshot` with an unknown ID, or with an instance that is not running, **When** the command finishes, **Then** it fails, creates no image, and does not change the instance.

---

### User Story 3 - See the run and end it (Priority: P3)

An operator checks that the `brazino` instance is running and then ends it. When the instance ends, that run's browser data is gone, including after a failed start. `screenshot` does not end the run.

**Why this priority**: A session left on disk after the run is the harm this story prevents. `status` and `stop` are how the operator sees and ends the run.

**Independent Test**: Run `start` against the local stand-in, run `status` and confirm the instance is listed, run `screenshot` and confirm the instance is still listed, run `stop` for that instance, confirm `status` no longer lists it, and confirm that run's browser data is gone.

**Acceptance Scenarios**:

1. **Given** a `brazino` instance is running, **When** the operator runs `status`, **Then** that instance is listed as running.
2. **Given** a `brazino` instance is running, **When** the operator runs `stop` for it, **Then** the instance ends, `status` no longer lists it, and that run's browser data is removed.
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
- The operator names a casino other than `brazino`, or supplies `Brazino`, `brazino777`, `brazino777.bet.br`, or `www.brazino777.mx`. `start` refuses. A host or a page address is not a casino identifier, and the operator does not type the page address on `start`.
- The session file's `casino` field does not match `brazino`. `start` fails with a clear message.
- The operator passes the session document itself instead of a file path. `start` does not take the document that way.
- The ordinary confirmation uses a cookie for the local stand-in's address only. It does not open the live casino, including when a file meant for the live site is present on the machine.
- The operator asks to run a game or another casino on this slice. This feature does not start any game and does not add casinos beyond `brazino`.
- The operator starts headless and cannot see the live page. `start` still applies the session and opens the page. `screenshot` still captures that page. Judging the live page by eye requires a visible window.
- `screenshot` does not stop the instance and does not remove its browser data. `stop` does.
- A screenshot of an instance that is still starting fails because the instance is not running, and it creates no image.
- The image may show whatever the page is showing, including a balance. Supervisor does not redact it, store a gallery, or send it anywhere. An image of the live casino is the operator's file and is not added to the project.
- `stop` for an instance that is already gone keeps today's result for that command. Browser data belonging to an instance this feature has ended is still removed.
- A `start` with no casino still opens a blank page. `screenshot` of that instance remains the existing command's capture of that page.
- Existing `onewin` behavior from `specs/004-open-onewin-page` MUST remain unchanged.
- The default test run MUST NOT start a second stand-in process solely for `brazino`. The same local stand-in that gates `onewin` confirmation also gates `brazino` confirmation, with session files and casino identifiers distinguishing each run.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `start` MUST accept the casino identifier `brazino` and a path to a session file. It MUST NOT accept the session document as the argument text.
- **FR-002**: The home page address for `brazino` MUST be the address registered with that casino, `https://www.brazino777.mx/`. The operator MUST NOT have to type that address on each `start`.
- **FR-003**: `start` MUST apply every cookie in the session file before the page opens, and the page that opens MUST be that casino's home page. A blank page MUST NOT be the result of a successful `brazino` start.
- **FR-004**: A session file MUST be able to carry more than one cookie. Each cookie supplies its name, its value, and the attributes the browser needs in order to store it for the page.
- **FR-005**: Session values MUST be supplied only at runtime. They MUST NOT appear in the command text of `start`, of the background run, or of `screenshot`, in the run's recorded output, in the instance list, in the project, or in the screenshot file's name.
- **FR-006**: `status` MUST list a running `brazino` instance while that instance is running, using the same lifecycle the operator already has for other instances.
- **FR-007**: When an instance ends, including a failed `start` and a browser that exits outside `stop`, Supervisor MUST remove that run's browser data so the session material does not remain. `screenshot` MUST NOT end the instance and MUST NOT remove that data.
- **FR-008**: This feature MUST register exactly one additional casino, `brazino`. It MUST NOT start any game on this slice. It MUST NOT change which games `onewin` can run.
- **FR-009**: Casino identifiers MUST be one or more lowercase letters. `brazino` is the identifier this feature adds. `Brazino`, `brazino777`, `brazino777.bet.br`, and `www.brazino777.mx` MUST be refused.
- **FR-010**: The ordinary confirmation MUST use the same shared local stand-in as `specs/004-open-onewin-page` (one server/process in the default suite, not one per casino). That stand-in MUST be bound to the local machine, show a balance or account information only when the expected session cookie for that local address is present, and withhold it otherwise. The confirmation MUST run `start`, then `status`, then `screenshot`, then `stop`, and MUST NOT open the live casino. The screenshot MUST be of that stand-in page. The default test run MUST include a `brazino` confirmation on that shared stand-in alongside existing `onewin` confirmations.
- **FR-015**: Relationship to `specs/004-open-onewin-page` MUST be additive only: register `brazino` and its home page, extend session validation for the `brazino` casino field, and add default tests that exercise `brazino` on the shared stand-in. This feature MUST NOT amend, replace, or relocate the `onewin` spec. Catalog entries, operator commands, and `onewin` acceptance behavior from **004** MUST remain available unchanged after this feature lands.
- **FR-011**: `screenshot` MUST keep the behavior already specified for that command: one instance ID, one image of the page that instance is currently showing, a printed file location, success or failure within 15 seconds, no prompt, and an optional output path that is the only file written when the operator supplies one. For a `brazino` instance, the page in the image MUST be the page `start` opened.
- **FR-012**: Supervisor MUST NOT store the operator's judgment of the live page and MUST NOT add a screenshot of the live casino to the project. Completing the ordinary confirmation MUST NOT require a recorded live login.
- **FR-013**: Headed and headless `start` MUST keep their current meaning. The default remains a visible window so the operator can judge the live page. `screenshot` MUST be able to capture the page either way.
- **FR-014**: `session info` for a `brazino` start with no registered game in this slice MUST show `casino: brazino` and `game: null` while the instance row exists, consistent with constitution principle V (a casino may be supported before any of its games can be run).

### Key Entities

- **Casino**: The registered site `brazino`. It has one home-page address, `https://www.brazino777.mx/`. The address is not the identifier. No game is registered in this slice.
- **Session file**: A file the operator already holds, passed to `start` by its path. It contains one or more cookies and a `casino` field set to `brazino`. It is not copied into the project.
- **Cookie**: A name, a secret value, and the attributes required to store it for a page (where it applies, the path, when it expires, and whether it is secure, hidden from page scripts, same-site restricted, or partitioned). The value exists only at runtime.
- **Browser instance**: One supervised browser for one run, with its own browser data. That data is removed when the instance ends. `screenshot` leaves the instance running.
- **Screenshot**: One image of that instance's current page, produced by the existing `screenshot` command. It is an operator file, not a project record.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator with a valid `brazino` session file opens the casino home page in one `start`, without typing the session document or the page address. The baseline is refusal or a blank page and zero Brazino pages.
- **SC-002**: The ordinary confirmation shows the local balance when the expected local cookie is present and withholds it when the cookie is absent. That confirmation completes `start`, `status`, `screenshot`, and `stop` with zero requests to the live casino. The screenshot shows the same balance outcome as the page.
- **SC-003**: For a running `brazino` instance, `screenshot` returns within 15 seconds with one image of the page `start` opened, and the same instance is still running afterward.
- **SC-004**: After every ended or failed run in that confirmation, the run's browser data is gone, and the known test secret is absent from the project, the command text, the run's recorded output, the instance list, and the screenshot file's name.
- **SC-005**: The operator can judge the live home page on their own machine. The product stores no record of that judgment and no screenshot of the live casino.
- **SC-006**: All automated checks for `onewin` from `specs/004-open-onewin-page` still pass without modification to operator workflows for that casino.
- **SC-007**: The default test run completes both `onewin` and `brazino` ordinary confirmations against one shared local stand-in, with zero live requests to `1win.com` or `www.brazino777.mx`, and without starting more than one stand-in server/process for those confirmations.

## Assumptions

- The only stated demand is one operator extending support beyond `onewin`. No other operator or ticket volume was found.
- A file of cookies the operator already holds is enough to open the home page. This feature adds no sign-in step.
- The registered home page for `brazino` is `https://www.brazino777.mx/` (operator choice, 2026-11-26). Other regional domains such as `brazino777.bet.br` or `brazino777.com` are not opened by `start` in this slice unless a later spec amends the registered address.
- Whether the live site accepts a particular session file was not observed. If a real file is rejected, this feature stops for a new decision. It does not grow into a search for a working session.
- The operator's permission for an automated session is their own statement. The site's terms were not read. A terms conflict stops this feature.
- Browser data is removed after the browser for that instance has exited, so removal does not race a browser that is still writing the data.
- The local stand-in proves that a session cookie was applied and that `start`, `status`, `screenshot`, and `stop` finish. It proves nothing about the live Brazino site.
- Constitution 1.2.0 allows `brazino` to be supported with no runnable game in this slice.
- Linux and WSL2 remain the supported operator environments, as in the existing CLI features. Other operating systems stay out of scope.
- The `screenshot` command is the one already specified in `specs/003-screenshot-cli`. This feature does not rename it, change its arguments, or change its failure rules.
- A `start` that names no casino keeps today's blank page. Capturing that blank page stays the existing screenshot behavior.
- The default test run starts one shared local stand-in for casino page confirmations, as established in `specs/004-open-onewin-page`. This feature adds `brazino` coverage on that same stand-in; it does not introduce a Brazino-only stand-in or a second concurrent stand-in for the default suite.
- `specs/004-open-onewin-page` remains the normative source for `onewin` behavior, the stand-in confirmation shape, and secrets handling. This spec only adds `brazino` catalog and test obligations on top.