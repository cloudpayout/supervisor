# Feature Specification: Screenshot CLI

**Feature Branch**: `feature/screenshot-command`

**Feature Directory**: `specs/003-screenshot-cli`

**Created**: 2026-11-15

**Status**: Draft

**Input**: User description: "Develop a new command to take screenshots of a running chrome instance. It will receive the instance ID as argument"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Capture a running instance (Priority: P1)

An operator who already started a browser instance needs a still image of the page that instance is showing, without stopping the run or taking over the terminal. They pass the instance ID that `start` printed and `status` lists.

**Why this priority**: A screenshot is the feature. Without a successful capture of a running instance, nothing else matters.

**Independent Test**: Start one instance, wait until status shows it as running, run the screenshot command with that ID, and confirm a new image file opens to that page, the command exits successfully, and a following status check still lists the same instance as running.

**Acceptance Scenarios**:

1. **Given** a browser instance that status shows as running, **When** the operator runs the screenshot command with that instance ID and no output path, **Then** within 15 seconds the system writes one PNG image of the page that instance is currently showing into the current directory, prints that file's location as a single line, and exits successfully.
2. **Given** a successful capture, **When** the operator checks status and the instance record, **Then** the instance is still running with the same owner and browser process as before the capture, it is still on the page it was showing (the blank page for an instance Supervisor started and has not navigated), and it was not stopped, navigated, resized, or replaced.
3. **Given** one running instance with a visible window and one running instance without a visible window, **When** the operator captures each by its own ID, **Then** both captures succeed and each image is of that instance's page. The windowless capture gates the default automated check. The visible-window capture is a manual check when a display is available.
4. **Given** two running instances on the blank page, **When** the operator captures one of them, **Then** that capture comes from that instance's own session, and the other instance keeps running with the same owner and browser process.
5. **Given** a running instance, **When** the operator captures it twice without choosing a path, **Then** two different files are created and the second capture does not replace the first.
6. **Given** the operator asks for command help, **When** help is shown, **Then** it describes the `screenshot` command, the required instance ID, and the default file location.

---

### User Story 2 - Refuse a capture that cannot succeed (Priority: P1)

An operator or script may pass a missing, unknown, or not-yet-running ID. The command must fail clearly and must not leave an image that looks like a successful capture.

**Why this priority**: A wrong or empty file is worse than no file. Scripts need a failure they can tell apart from success.

**Independent Test**: Run the command with no ID, with an unknown ID, and with an instance that is not running. Confirm each attempt fails, explains why, and creates no image file.

**Acceptance Scenarios**:

1. **Given** the command is run with no instance ID, with more than one ID, or with a blank ID, **When** the command runs, **Then** it does not prompt for input, does not create an image file, explains that an instance ID is required, and fails as a usage error.
2. **Given** an instance ID that status does not list, including an instance that was stopped and removed, **When** the operator runs the screenshot command with that ID, **Then** the command fails, reports that the instance was not found, creates no image file, and does not change any other instance.
3. **Given** a tracked row that status shows as not yet running, or any other tracked row whose state is not running, **When** the operator runs the screenshot command with that ID, **Then** the command fails, reports that the instance is not running, creates no image file, and does not stop or alter other instances. A removed row is scenario 2, not this one.
4. **Given** a capture that cannot be finished (the instance disappears during the capture, or the image cannot be saved), **When** the command ends, **Then** it fails, does not report success, and does not leave an incomplete file as the screenshot.

---

### User Story 3 - Choose where the image is saved (Priority: P2)

An operator or script needs the image at a chosen path instead of the default file in the current directory. Have in mind, the path can be relative/absolute.

**Why this priority**: The default file is enough to capture a page. Choosing a path matters once the image has to land in a known place for a person or a script.

**Independent Test**: Capture a running instance with an explicit new file path. Confirm the image is at that path, the printed location matches it, and an existing file at that path is left untouched when the command is refused.

**Acceptance Scenarios**:

1. **Given** a running instance and a file path that does not exist yet, **When** the operator runs the screenshot command with that instance ID and that output path, **Then** the image is written only at that path, the printed location is that path, and no extra default file is created.
2. **Given** the chosen path already exists, names a directory, or cannot be written, **When** the operator runs the screenshot command, **Then** the command fails, does not overwrite or replace anything already at that path, and does not report success.
3. **Given** the operator asks for screenshot help, **When** help is shown, **Then** it describes `--output` and `-o`, states that the path may be relative or absolute, and does not ask the operator to type an ID.

---

### Edge Cases

- The command never reads standard input and never asks the operator to type an ID.
- A whitespace-only ID is treated as a missing ID and fails as a usage error.
- Repeating a capture in the same second still produces a new file when the operator does not choose a path.
- Capturing one instance while another command stops it ends in one clear outcome: a complete image of the page from before the stop, or a failure with no image presented as success. The screenshot command does not itself stop the instance.
- Captures of two different running instances at the same time can both succeed, each with its own image.
- A page that is mostly blank still produces an openable image of that page, not an error in place of the file.
- The image may include whatever the page is showing, including sensitive information. The command does not redact, crop, or upload the image.
- The command returns the shell either with a finished image or with a failure. It does not hold the terminal open for monitoring.
- Platforms other than Linux and WSL2 are out of scope, consistent with the existing operator surface.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Supervisor MUST provide a `screenshot` command that takes one still image of a running browser instance identified by its instance ID.
- **FR-002**: The command MUST take the instance ID as its only positional argument and MUST NOT prompt on standard input.
- **FR-003**: When the ID is missing, blank, or not the sole positional argument, the command MUST fail as a usage error and MUST NOT create an image file.
- **FR-004**: On success, the command MUST write one PNG image of the page that instance is currently showing, MUST print that file's location as a single line, and MUST exit successfully.
- **FR-005**: A successful capture MUST leave the instance running and MUST NOT navigate, click, resize, stop, or otherwise change that instance in order to take the image.
- **FR-006**: When the operator does not choose a path, the command MUST create the image in the current directory under a new name that includes the instance ID and the capture time, and MUST NOT replace an existing file.
- **FR-007**: The command MUST accept an optional output path. When provided, the image MUST be written only at that path. When that path exists, is a directory, or cannot be written, the command MUST fail without overwriting existing content and without creating a default file.
- **FR-008**: The image MUST show only the requested instance's current page. It MUST NOT show another instance's page and MUST NOT be a picture of the operator's whole desktop.
- **FR-009**: A running instance MUST be capturable whether or not it has a visible window. The image is of the page the instance is showing.
- **FR-010**: When no tracked row exists for the ID, including an unknown ID and an instance that was stopped and removed, the command MUST report that the instance was not found. When a tracked row exists but its state is not running, including starting and a leftover stopped row, the command MUST report that the instance is not running. Both cases MUST NOT create an image file and MUST NOT change other instances.
- **FR-011**: When a capture cannot be completed, the command MUST fail, MUST NOT report success, and MUST NOT leave an incomplete file as the screenshot.
- **FR-012**: The command MUST finish with success or failure within 15 seconds for a single instance, rather than holding the terminal until the operator interrupts it.
- **FR-013**: Operator help MUST describe the `screenshot` command, the required instance ID, the optional output path, and the default file location.

### Key Entities

- **Instance**: An existing supervised browser session with a stable ID. Only an instance that is running can be captured. The ID is the same value `start` prints and `status` lists.
- **Screenshot**: One still PNG image of a single instance's current page, produced by one command invocation, stored as one file, and associated with that instance ID and the time it was taken.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In at least 95 of 100 trials with a running instance, the operator receives an openable PNG of that instance's current page and the shell returns within 15 seconds.
- **SC-002**: 100% of invocations that do not pass exactly one instance ID fail as a usage error and create no image file.
- **SC-003**: After every successful capture in a 20-trial check, the same instance still appears as running with the same owner and browser process, and it is still on the page it was showing before the capture.
- **SC-004**: 100% of attempts against an unknown ID or a stopped instance whose row was removed create no image file and report that it was not found. 100% of attempts against a tracked instance that is not running create no image file and report that it is not running.
- **SC-005**: Two consecutive captures of the same running instance, with no output path, produce two distinct files; the second does not replace the first.
- **SC-006**: When the operator supplies a new output path, the saved image is at that path, the printed location matches it, and no second default file is created.

## Assumptions

- The command name is `screenshot`. The instance ID is the only required argument, matching how `stop` already receives an ID.
- "Running" matches what `status` shows as running. An instance that is still starting is not capturable.
- The capture is the current view of that instance's page, not a stitched image of the entire scrollable page and not a photograph of the desktop or window frame.
- v1 starts each instance on one blank page and does not open more tabs. The captured page is that instance's worker tab. Choosing a different tab is out of scope until a later spec adds tabs.
- The default file name looks like `screenshot-<instance-id>-<YYYYMMDD-HHMMSS>.png` in the current directory. If that name is already taken, another unused name is chosen.
- The optional output path is given with `--output` or `-o`.
- A missing instance is a failure for this command. That differs from `stop`, where a missing ID can still count as finished, because a screenshot with no page is not a successful result.
- Usage errors stay distinct from capture failures, so a script can tell a bad invocation from an instance that could not be captured.
- The image is ordinary operator output. Supervisor does not store a gallery, redact the image, or send it anywhere.
- `start`, `status`, and `stop` keep their current behavior. Video, capturing every instance in one command, prompting for an ID, and support outside Linux and WSL2 are out of scope.
