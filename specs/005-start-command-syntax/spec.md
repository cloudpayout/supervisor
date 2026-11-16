# Feature Specification: Start command syntax

**Feature Branch**: `005-start-command-syntax`

**Created**: 2026-11-16

**Status**: Draft

**Input**: Original request (superseded by clarifications below): mandatory `--session-file` and `casino` in the file. **As specified**: optional positional session path (or no path for a blank page); `casino` only in the file; no `--session-file`, `--casino`, or `--headed` on operator `start`; default visible window unless `--headless`. Example file: `{"casino": "onewin", "cookies": [...]}`.

An operator starts a supervised session by passing a session file path as the positional argument, or by passing no path. When a path is passed, the file names the casino and carries the cookies, and that casino's page opens. When no path is passed, a blank page opens. `start` does not take `--session-file`, a casino option, or `--headed`. Omitting `--headless` opens a visible window. Passing `--headless` runs with no window.

## Clarifications

### Session 2026-11-16

- Q: Which command lines should `supervisor start` accept as a successful start? → A: The session path is positional: `supervisor start <path>`, `supervisor start <path> --headless`, and `supervisor start --headless <path>`. An extra argument is a usage error. `--help` and `-h` are decided separately.
- Q: Should `supervisor start --headless <path>` succeed the same way as `supervisor start <path> --headless`? → A: Yes. `supervisor start --headless <path>` succeeds and has the same result as `supervisor start <path> --headless`.
- Q: Which `start` with no session file should open a blank page? → A: No path on the command opens a blank page (`supervisor start` and `supervisor start --headless`). A path that is not on disk fails and opens nothing.
- Q: Do `supervisor start --help` and `supervisor start -h` print help without opening a blank page? → A: Only `supervisor start --help` prints help and exits successfully. It does not start an instance, including when a path or `--headless` is also present. `supervisor start -h` is a usage error and opens no page.
- Q: Does `supervisor start --session-file=<path>` count as the same command as `supervisor start --session-file <path>`? → A: Remove the `--session-file` flag. The session file path is only the positional argument. Passing `--session-file` is a usage error and opens no page.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start from a session file that names the casino (Priority: P1)

An operator who already holds a session file runs `start` with that file's path as the positional argument. The file names the casino, for example `onewin`, and lists the cookies. The operator does not type the casino name on the command, and does not type the page address. A successful start applies every cookie and opens that casino's registered page. `start` exits while the instance keeps running.

**Why this priority**: The command the operator types is the change. The casino belongs with the session they already hold. A start with no path stays a blank page.

**Independent Test**: Write a session file whose casino is `onewin` and whose cookie is the one a local stand-in expects. Run `start <path> --headless`. Confirm the stand-in shows the balance, the command text contains no `--session-file` flag and no casino option, and the command prints an instance id. Repeat with a file that omits the expected cookie and confirm the stand-in withholds the balance. Do not open the live casino.

**Acceptance Scenarios**:

1. **Given** a session file whose casino is `onewin` and whose cookies are valid, **When** the operator runs `start <path>`, **Then** Supervisor applies every cookie, opens the casino's registered page, prints an instance id, and exits while the instance keeps running.
2. **Given** a local stand-in that shows a balance only when its expected cookie is present, **When** the ordinary confirmation runs `start <path> --headless` with a file that names `onewin` and includes that cookie, **Then** the stand-in shows the balance.
3. **Given** the same local stand-in, **When** the ordinary confirmation runs `start <path> --headless` with a file that names `onewin` and omits that cookie, **Then** the stand-in does not show the balance.
4. **Given** a valid session file, **When** the operator runs `start <path>`, **Then** the command passes that path as the positional argument and does not include `--session-file`, a casino option, or the session document.
5. **Given** the operator runs `start` with no path, **When** the command finishes, **Then** it opens a blank page, prints an instance id, and exits while the instance keeps running.
6. **Given** the operator runs `start --headless` with no path, **When** the command finishes, **Then** it opens a blank page with no window, prints an instance id, and exits while the instance keeps running.
7. **Given** the operator passes a path that is not on disk, **When** the command finishes, **Then** it fails, prints no instance id, opens no page, and leaves no instance and no browser data.
8. **Given** a valid session file, **When** the operator runs `start --headless <path>`, **Then** the outcome matches `start <path> --headless` for that same file.
9. **Given** any invocation that passes `--session-file`, adds an extra argument, or passes a flag other than `--headless` or `--help`, **When** the command finishes, **Then** it fails as a usage error, opens no page, and leaves no instance.

---

### User Story 2 - Open a visible window unless headless is requested (Priority: P2)

An operator who wants to look at the page runs `start` with a session file path and no display option. The instance opens in a visible window. An operator who wants no window adds `--headless`. `--headed` is not a `start` option. `--session-file` is not a `start` option.

**Why this priority**: The visible window is how the operator judges the live page. A separate headed option repeats the default and is withdrawn. The session path is the positional argument, so a session-file flag is withdrawn too.

**Independent Test**: Run `start <path>` with no `--headless` and confirm a visible window on the page the file's casino opens. Run `start <path> --headless` and confirm that run has no window. Run `start <path> --headed` and `start --session-file <path>` and confirm each is a usage error, with no instance id and no instance. The ordinary confirmation uses the windowless form and does not require the visible-window check.

**Acceptance Scenarios**:

1. **Given** a valid session file, **When** the operator runs `start <path>` and does not pass `--headless`, **Then** the instance opens in a visible window.
2. **Given** the operator runs `start` with no path and does not pass `--headless`, **When** the command finishes, **Then** the blank page opens in a visible window.
3. **Given** a valid session file, **When** the operator runs `start <path> --headless` or `start --headless <path>`, **Then** the instance runs with no window, and the page is still the one that casino opens. `start --headless <path>` matches `start <path> --headless` for the same file.
4. **Given** any arguments, **When** the operator passes `--headed` or `--session-file` on `start`, **Then** the command fails as a usage error, prints no instance id, and leaves no instance.
5. **Given** the operator runs `supervisor start --help`, including with a path or `--headless` also present, **When** the command finishes, **Then** it prints help and exits successfully, and it does not start an instance or open a page. The help lists the positional path as the way to pass the session file, states that omitting the path opens a blank page, lists `--headless` as the display option, and does not list `--session-file`, `--headed`, or a casino option.
6. **Given** the operator runs `supervisor start -h`, **When** the command finishes, **Then** it fails as a usage error, opens no page, and leaves no instance.

---

### User Story 3 - Keep the rest of the run (Priority: P3)

After a start that follows the new command, the operator can still list the instance, capture its page, and end it. Cookie values stay out of the command, the run's recorded output, the instance list, and the project. Ending the instance removes that run's browser data.

**Why this priority**: A shorter command is only useful if the run that follows still behaves as it does today.

**Independent Test**: After the headless stand-in start in User Story 1, run `status` and confirm the instance is listed, run `screenshot` and confirm one image of that page, then run `stop` and confirm the instance is gone and that run's browser data is gone. Confirm the known test cookie value is absent from the command text, the recorded output, the instance list, and the project.

**Acceptance Scenarios**:

1. **Given** a start from a valid session file has printed an instance id, **When** the operator runs `status`, **Then** that instance is listed as running.
2. **Given** that instance is running, **When** the operator runs `screenshot` with its id, **Then** the image is the page `start` opened, and the instance is still running.
3. **Given** that instance is running, **When** the operator runs `stop` for it, **Then** the instance ends and that run's browser data is removed.
4. **Given** a session file with a known test cookie value, **When** `start` and `screenshot` have run, **Then** that value is absent from the command text, the run's recorded output, the instance list, and the project. The session file path may appear on the command as the positional argument.

---

### Edge Cases

- The operator runs `start` with no path. A blank page opens, an instance id is printed, and the instance keeps running. `start --headless` with no path opens that blank page with no window.
- The operator runs `start <path>`, `start <path> --headless`, or `start --headless <path>` with a valid file. Each opens that casino's page. `start --headless <path>` matches `start <path> --headless`.
- The operator passes `--session-file`, including `--session-file <path>` and `--session-file=<path>`, or adds an extra argument. The command fails as a usage error, opens no page, and leaves no instance.
- The operator passes a casino option such as `--casino onewin`. The command fails as a usage error and leaves no instance. The casino name is read only from the session file.
- The operator passes `--headed`, including together with `--headless`. The command fails as a usage error and leaves no instance.
- The operator runs `supervisor start --help`. Help is printed, the command exits successfully, and no instance starts, including when a path or `--headless` is also present. `supervisor start -h` fails as a usage error and opens no page.
- The session file path is not on disk, or the file is unreadable or not a session document. `start` fails with a clear message, opens no page, leaves no instance, and leaves no browser data for the attempt.
- The session file omits `casino`, leaves it empty, or sets it to a value other than the registered identifier `onewin`. `Onewin` and `1win.com` are refused. `start` leaves no instance.
- The session file names `onewin` and contains no cookies, or contains a cookie with an empty name or an empty value. `start` fails and leaves no instance.
- The session file contains more than one cookie. All of them are applied before the page opens.
- The operator pastes the session document on the command instead of passing its path. `start` does not accept the document that way.
- The session file contains a field the operator was not told to supply. `start` rejects the file and leaves no instance.
- The ordinary confirmation uses a cookie for the local stand-in only. It does not open the live casino.
- The operator asks to run `aviator` or a casino other than `onewin`. This feature does not start them.
- A visible window is the run the operator gets by omitting `--headless`. Judging the live page by eye uses that window. `screenshot` still captures the page when the run is windowless.
- `start` fails after creating browser data for the attempt. No instance remains, and that browser data is removed.
- `status`, `stop`, and `screenshot` keep their current arguments and failure results. `screenshot` does not end the instance. `stop` does.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A session file path, when supplied, MUST be the single positional argument. The successful file command lines are `supervisor start <path>`, `supervisor start <path> --headless`, and `supervisor start --headless <path>`. `supervisor start --headless <path>` MUST match `supervisor start <path> --headless` for the same file. `supervisor start` and `supervisor start --headless` with no path MUST open a blank page, MUST print an instance id, and MUST exit while the instance keeps running. A path that is not on disk MUST fail, MUST print no instance id, MUST open no page, and MUST leave no instance and no browser data. Passing `--session-file` in any spelling, an extra argument, or any flag other than `--headless` or `--help` MUST fail as a usage error, MUST open no page, and MUST leave no instance. `supervisor start --help` MUST print help and exit successfully, and MUST NOT start an instance, including when a path or `--headless` is also present. `supervisor start -h` MUST fail as a usage error, MUST open no page, and MUST leave no instance.
- **FR-002**: When a session file is passed, the casino for that start MUST be the `casino` value in that file. A start with no path MUST NOT name a casino and MUST NOT read a session file. `start` MUST NOT offer a casino option or a `--session-file` option. Passing either MUST fail as a usage error and MUST leave no instance.
- **FR-003**: A session file MUST name the casino and MUST list the cookies. A file the operator supplies for `onewin` has the shape `{"casino": "onewin", "cookies": [...]}`. `start` MUST accept the file only by its positional path. It MUST NOT accept the document as argument text, and it MUST NOT accept a `--session-file` flag.
- **FR-004**: When the file names `onewin` and its cookies are valid, `start` MUST apply every cookie before the page opens, and the page MUST be the registered `onewin` page, `https://1win.com/`. The operator MUST NOT have to type that address. `start` MUST print an instance id and exit while the instance keeps running.
- **FR-005**: The casino value MUST be the registered identifier `onewin`, one or more lowercase letters. A missing casino, an empty casino, `Onewin`, `1win.com`, and any other value MUST be refused with no instance left behind.
- **FR-006**: `--headed` MUST NOT be a `start` option. Passing it MUST fail as a usage error and MUST leave no instance, including when `--headless` is also present. When the operator omits `--headless`, the instance MUST open in a visible window. When the operator passes `--headless`, the instance MUST run with no window.
- **FR-007**: `supervisor start --help` MUST print operator help for `start` and MUST exit successfully without starting an instance. That help MUST list the positional path as the way to pass the session file, MUST state that omitting the path opens a blank page, and MUST list `--headless` as the display option. Help MUST NOT list `--session-file`, `--headed`, `-h`, or a casino option. `supervisor start -h` MUST fail as a usage error and MUST open no page.
- **FR-008**: A session file with no cookies, or with a cookie that has an empty name or an empty value, MUST fail `start` with no instance and no browser data left behind. A file with more than one cookie MUST apply every cookie before the page opens.
- **FR-009**: Cookie values MUST be supplied only at runtime. They MUST NOT appear in the command text of `start`, of the background run, or of `screenshot`, in the run's recorded output, in the instance list, or in the project. The session file path MAY appear on the command as the positional argument. The casino name is not a secret.
- **FR-010**: `status`, `stop`, and `screenshot` MUST keep their current commands and results. `status` MUST list a running instance from this `start`. `screenshot` MUST capture the page that `start` opened and MUST leave the instance running. When an instance ends, including a failed `start` after browser data was created, Supervisor MUST remove that run's browser data.
- **FR-011**: The ordinary confirmation MUST use a local stand-in, a session file that contains `"casino": "onewin"` and the expected cookie, and the command `start <path> --headless`, then `status`, then `screenshot`, then `stop`. It MUST NOT open the live casino. The stand-in MUST show a balance only when the expected cookie is present, and MUST withhold it otherwise. The default test run MUST include this confirmation. The visible-window check MUST NOT gate the default test run.
- **FR-012**: This feature MUST NOT add a casino, MUST NOT start `aviator`, and MUST NOT change which cookie attributes a session file may carry beyond requiring `casino` in the same file.

### Key Entities

- **Session file**: A file the operator already holds. `start` receives its path as the single positional argument. It names one casino and lists one or more cookies. It is not copied into the project. Example contents: `{"casino": "onewin", "cookies": [...]}`. A start with no path does not read one.
- **Casino**: The registered site named in the session file. This feature accepts `onewin`. The page address is `https://1win.com/` and is not the identifier. The operator does not type the address or the identifier on `start`.
- **Cookie**: A name, a secret value, and the attributes already required to store it for a page. The value exists only at runtime.
- **Browser instance**: One supervised browser for one run. With no `--headless`, it has a visible window. With `--headless`, it has no window. Its browser data is removed when the instance ends.
- **Screenshot**: One image of that instance's current page, produced by the existing `screenshot` command.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator with a valid session file opens the casino page in one `start` that passes the file path as the positional argument and passes no `--session-file` flag and no casino name. In the ordinary confirmation, the stand-in shows the balance when the expected cookie is present and withholds it when the cookie is absent. That confirmation completes `start`, `status`, `screenshot`, and `stop` with zero requests to the live casino.
- **SC-002**: Each rejected `start` in this feature — `--session-file` in any spelling, an extra argument, a casino option, `--headed`, `-h`, a path that is not on disk, an unreadable file, a file with no casino or a casino other than `onewin`, or a file with no cookies — finishes with no instance id and zero instances left running.
- **SC-003**: A `start` that omits `--headless` shows a visible window, including a blank page when no path is given. A `start` that includes `--headless` shows no window, including `start --headless` with no path. `supervisor start --help` prints help and starts no instance. `supervisor start -h` finishes with no instance and no page. The help lists the positional path, the blank page for no path, and `--headless`, and lists neither `--session-file`, `--headed`, nor a casino option.
- **SC-004**: After the ordinary confirmation, the known test secret is absent from the command text, the run's recorded output, the instance list, and the project, and that run's browser data is gone after `stop`.
- **SC-005**: `supervisor start` and `supervisor start --headless` with no path each open a blank page and print an instance id. A path that is not on disk finishes with no instance id and no page.

## Assumptions

- "By default that mode will be chosen at least `--headless` is called" means the visible window is the default, and `--headless` is the only display option that selects a windowless run. `--headed` is withdrawn because it repeated the default.
- The session file path is optional and is only the positional argument. `--session-file` is not a `start` option. Passing it, with a space or an equals sign, is a usage error and opens no page. `supervisor start` and `supervisor start --headless` with no path open a blank page. A path that is not on disk fails and opens nothing.
- The casino option from the onewin page feature is withdrawn. The session file is the only place the operator names the casino. A command that still passes a casino option fails as a usage error.
- Cookie fields, and the rules that reject a bad cookie, stay those already specified for the onewin session file. This feature adds a required `casino` name beside `cookies`. Unknown fields are still rejected.
- The only registered casino remains `onewin`. Its page remains `https://1win.com/`. `aviator` stays named and is not started. A file that names any other casino is refused.
- `status`, `stop`, and `screenshot` keep the commands already specified. This feature changes the `start` command the operator types and where the casino name is read.
- The ordinary confirmation stays a local stand-in plus `--headless`, using `start <path> --headless`. Looking at a visible window on the live page stays the operator's own check and does not gate the default test run.
- Linux and WSL2 remain the supported operator environments. Other operating systems stay out of scope.
- A usage error prints a clear message, prints no instance id, and leaves no instance and no browser data. The message does not include a cookie value. `supervisor start -h` and any `start` that passes `--session-file` are usage errors. `supervisor start --help` prints help and does not start an instance.
