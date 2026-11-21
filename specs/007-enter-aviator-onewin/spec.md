# Feature Specification: Enter Aviator on onewin

**Feature Branch**: `007-enter-aviator-onewin`

**Created**: 2026-11-21

**Status**: Draft

**Input**: User description: "use the handoff `.specify/assessments/aviator-in-onewin/decision.md` to develop this new feature."

**Assessment handoff**: Supervised `onewin` does not finish entering Aviator (play URL, currency modal, play or demo choice). Runs cannot reliably reach **Aviator ready** or prove it. The chosen approach: apply session cookies, open the Aviator play URL from optional `locale`, automate the outer-page currency modal, fail fast when a requested currency is missing, wait until **Aviator ready** within about one minute using a game-surface ready check, and use a minimal local stand-in in the ordinary confirmation. Operators verify live success with `screenshot`. In-round betting and live-casino default tests stay out of scope.

This feature **amends** feature 004: `onewin` `start` no longer stops at the casino main page when entering Aviator is the registered game behavior for this casino.

## Clarifications

### Session 2026-11-21

- Q: What observable condition must be true before Supervisor treats a run as **Aviator ready**? (FR-008) → A: The Spribe Aviator launch iframe is present and a registered inner game-ready element (for example the game logo) is visible inside that iframe.
- Q: When Aviator entry fails before **Aviator ready**, how should the `start` command finish? (FR-009) → A: `start` returns immediately after spawning the supervised browser and registering the instance; entry continues in the background. If entry fails before **Aviator ready**, the instance is removed and browser data is cleared even though `start` may already have exited successfully.
- Q: What format must the session file’s `currency` value use when matching the modal? (FR-004, FR-005) → A: Three-letter currency codes (e.g. `MXN`, `USD`) matched against visible modal options.
- Q: How should the ordinary confirmation load the minimal local Aviator stand-in page? (FR-012) → A: A loopback HTTP server (e.g. `http://127.0.0.1`) serves minimal HTML; the ordinary confirmation uses that origin, not `file://` or the live play URL.
- Q: After `start` has already exited successfully, how should an operator learn why background Aviator entry failed? (FR-010) → A: The instance disappears from `status` with no product-visible failure reason on the CLI; the operator infers failure from the missing instance or from private observation (e.g. `screenshot` during the window, headed browser).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reach Aviator ready after onewin start (Priority: P1)

An operator who already has an `onewin` session runs `start` with the session file path. Supervisor applies cookies, opens the Aviator play URL for the configured or default locale, completes the game-balance currency modal, activates real play or demo according to the session file, and waits until **Aviator ready** without operator clicks. The instance keeps running. The operator judges live success on their machine; the product does not store that judgment.

**Why this priority**: This closes the gap between naming `aviator` and actually reaching a playable game under Supervisor.

**Independent Test**: Against the local stand-in in User Story 3, run `start` with a session that names `onewin` and optional play preferences. Confirm the stand-in reaches its **game ready** marker without manual steps. Do not open the live casino in the ordinary confirmation.

**Acceptance Scenarios**:

1. **Given** a valid `onewin` session file, **When** the operator runs `start`, **Then** `start` exits after the instance is registered while entry continues in the background; cookies are applied, the Aviator play URL opens (not only the casino main page), the currency modal is settled, play or demo is activated per the session file, and the run reaches **Aviator ready** within one minute under typical conditions without further operator clicks.
2. **Given** the session file omits `currency`, **When** `start` completes entry, **Then** the first or default currency shown in the modal is used.
3. **Given** the session file sets `currency` to a code not offered in the modal, **When** `start` runs, **Then** it fails clearly, leaves no running instance, and removes browser data for the attempt.
4. **Given** the session file omits `demo` or sets `demo` to false, **When** entry runs, **Then** real-money play is activated (not demo mode).
5. **Given** the session file sets `demo` to true, **When** entry runs, **Then** demo mode is activated instead of real-money play.
6. **Given** the session file omits `locale`, **When** the play URL opens, **Then** it uses the English unprefixed play address registered for `aviator` on `onewin`.
7. **Given** the session file sets `locale` (for example `es-MX`), **When** the play URL opens, **Then** it uses the locale-prefixed play address for that value.

---

### User Story 2 - See failure when entry does not complete (Priority: P2)

An operator whose session is invalid, whose currency is wrong, or whose run times out receives a clear failure from `start` or from the running workflow, without a silent hang on the currency modal.

**Why this priority**: Failure clarity is a success metric from assessment; operators must not guess whether Aviator loaded.

**Independent Test**: Against the local stand-in, run `start` with a session that requests a currency the stand-in does not offer. Confirm fail-fast behavior with a clear message and no leftover instance.

**Acceptance Scenarios**:

1. **Given** a requested `currency` is absent from the modal, **When** `start` attempts entry, **Then** it fails fast with a message that identifies currency mismatch, and no instance remains running.
2. **Given** **Aviator ready** is not reached within one minute, **When** the wait elapses, **Then** the instance is removed from the registry and browser data is removed (even if `start` already exited successfully), without a dedicated CLI failure reason beyond the instance no longer appearing in `status`.
3. **Given** cookies do not authenticate the play page, **When** entry cannot proceed, **Then** failure is reported without claiming **Aviator ready**.

---

### User Story 3 - Recorded confirmation on a local Aviator stand-in (Priority: P3)

A validation engineer runs the ordinary confirmation without the live casino. The stand-in reproduces the currency modal and a fake **game ready** marker. The confirmation exercises `start`, `status`, `screenshot`, and `stop` through a successful Aviator entry on that stand-in.

**Why this priority**: Constitution requires default tests off the live casino; entry automation must still be provable in CI.

**Independent Test**: Run the ordinary confirmation only against the minimal local stand-in. Confirm zero requests to the live play URL and an image of the post-ready stand-in page.

**Acceptance Scenarios**:

1. **Given** the ordinary confirmation, **When** it runs `start` against the stand-in served at loopback HTTP (for example `127.0.0.1`), **Then** it does not open the live `onewin` play URL.
2. **Given** a session file valid for the stand-in, **When** the confirmation completes entry, **Then** the stand-in shows **game ready**, `status` lists the instance, `screenshot` captures that page within 15 seconds, and `stop` ends the run and removes browser data.
3. **Given** the confirmation, **When** it finishes, **Then** no session secret, no live-casino screenshot, and no gameplay capture are added to the project.

---

### User Story 4 - Capture the page after entry (Priority: P4)

An operator uses the existing `screenshot` command to capture what the instance shows after entry, for private verification on live machines.

**Why this priority**: Assessment pairs automated ready detection with operator visual confirmation.

**Independent Test**: After stand-in entry in User Story 3, run `screenshot` and confirm the image matches the post-ready stand-in, with the instance still running.

**Acceptance Scenarios**:

1. **Given** an instance that reached **Aviator ready**, **When** the operator runs `screenshot`, **Then** within 15 seconds they receive one image of the current page, printed path, success exit, and the instance remains running.
2. **Given** a live entry the operator considers successful, **When** they run `screenshot`, **Then** the image is their file only; Supervisor does not store or transmit it.

---

### User Story 5 - Session preferences without new command surface (Priority: P5)

An operator encodes optional play preferences in the same session JSON file they already pass to `start`, without pasting secrets on the command line.

**Why this priority**: Keeps the CLI surface from feature 005; extends the file contract only.

**Independent Test**: Parse acceptance for files with and without optional keys; confirm backward compatibility when keys are absent.

**Acceptance Scenarios**:

1. **Given** a session file that matches feature 004 and 005 rules and omits `currency`, `demo`, and `locale`, **When** `start` runs, **Then** entry uses defaults (default currency, real play, English unprefixed URLs).
2. **Given** unknown top-level fields in the session file, **When** `start` runs, **Then** behavior follows the session-file rules from prior features (unknown fields fail parse).

---

### Edge Cases

- Session file missing, unreadable, or wrong casino: `start` fails as today; no instance; browser data removed.
- Empty cookies: fail as today.
- Operator names a casino other than `onewin`: refused as today.
- `start` with no session file still opens a blank page; this feature does not enter Aviator.
- Headless `start`: entry still runs; operator eyeball verification on live site requires a headed window.
- Play URL UI changes language or layout: failures must be distinguishable from timeout where possible; maintenance is expected outside this spec.
- **Aviator ready** requires the launch iframe and inner game-ready element, not only a dismissed modal. A top-level screenshot alone may not prove inner game load; the iframe ready check is authoritative for automation success.
- This feature does not place bets, change stakes, or cash out.
- `stop` and profile removal rules from prior features still apply.
- `start` may exit before **Aviator ready**; an instance can disappear from `status` shortly after a successful-looking `start` when background entry fails, with no CLI message explaining why.
- Between `start` exit and **Aviator ready**, `status` may show the instance as entering; `screenshot` during that window may show the modal or partial load.
- Currency fail-fast and other entry errors remove the instance without a new CLI diagnostic; the operator uses headed observation or timing of `status`/`screenshot`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: For `onewin`, successful `start` with a session file MUST apply all cookies, then open the registered **Aviator play URL** for the effective locale, not only the casino main page.
- **FR-002**: The effective locale MUST come from optional `locale` in the session file when present; when absent, play URLs MUST use the English unprefixed addresses registered for `onewin` and `aviator`.
- **FR-003**: When `locale` is present, play URLs MUST use that locale as a path prefix in the form agreed for `onewin` (for example `es-MX` → `https://1win.com/es-MX/casino/play/v_spribe:aviator`).
- **FR-004**: Entry MUST automate the **outer-page** game-balance currency modal: open currency selection, choose `currency` from the session file when set (three-letter ISO 4217-style code such as `MXN` or `USD`, matched to visible modal options), otherwise accept the first or default option shown, then activate real play or demo per `demo`.
- **FR-005**: When `currency` is set to a three-letter code that is not offered in the modal, entry MUST fail fast with a clear currency error; it MUST NOT silently pick another currency.
- **FR-006**: When `demo` is absent or false, entry MUST activate real-money play; when `demo` is true, entry MUST activate demo mode.
- **FR-007**: Entry MUST NOT require a full-page reload after cookies are applied solely to satisfy this feature.
- **FR-008**: Supervisor MUST wait until **Aviator ready** or until one minute elapses. **Aviator ready** means: (1) the currency modal path is complete and real play or demo was activated per the session file, and (2) a Spribe Aviator launch iframe is present and a registered inner game-ready element (for example the game logo) is visible inside that iframe. Planning MAY name the exact selectors; it MUST NOT weaken this outcome.
- **FR-009**: `start` MUST return after registering the instance and spawning the supervised browser; Aviator entry MUST continue in the background without blocking `start` until **Aviator ready**. When entry reaches **Aviator ready**, the instance MUST remain running on the play page until `stop` or browser exit. When entry fails before **Aviator ready**, Supervisor MUST remove the instance from the registry and remove browser data for the attempt, even if `start` already exited successfully.
- **FR-009a**: After `start` exits, `status` MUST list the instance while it is still entering or **Aviator ready**. When background entry fails, the instance MUST disappear from `status`; this feature MUST NOT add a CLI-visible entry-failure reason or diagnostic command.
- **FR-010**: Internally, entry logic MUST treat failure causes separately (session, currency, controls, timeout, unexpected page) for maintenance and tests. The operator CLI MUST NOT surface those categories when the instance is removed after a post-`start` failure.
- **FR-011**: Session secrets MUST remain subject to FR-005 from feature 004: not in command text, registry, logs, or project files.
- **FR-012**: The ordinary confirmation MUST use a **minimal local HTML stand-in** served over **loopback HTTP** (for example `http://127.0.0.1`) that reproduces the currency modal and the same **Aviator ready** contract (including fake launch iframe and inner ready marker). The confirmation MUST run `start`, `status`, `screenshot`, and `stop`, and MUST NOT contact the live casino. The default test run MUST include this confirmation. Session cookies for the confirmation MUST target the stand-in’s loopback origin.
- **FR-013**: `screenshot` MUST keep existing behavior; for instances that reached **Aviator ready**, the captured page MUST be the page the instance is showing after entry.
- **FR-014**: This feature MUST run the registered game `aviator` for casino `onewin`. It MUST NOT add other games or casinos.
- **FR-015**: This feature MUST NOT automate in-round gameplay (bets, cash-out, auto-bet). Naming `aviator` in the catalog MUST align with entry automation after this feature ships.
- **FR-016**: Optional session fields `currency` (string, three-letter code), `demo` (boolean), and `locale` (string) MUST be supported. Omitting them MUST preserve backward compatibility with session files valid under features 004 and 005.
- **FR-017**: Live `onewin` entry MUST be testable by the operator manually; live play URL checks MUST NOT gate the default automated test run.

### Key Entities

- **Casino**: `onewin`, with registered main and play addresses per locale rules above, and game `aviator`.
- **Game**: `aviator`, entered by this feature after `start` with an `onewin` session.
- **Session file**: Extends the 004/005 JSON document with optional `currency`, `demo`, and `locale`. Cookies unchanged in meaning.
- **Play preferences**: The effective values of `currency` (three-letter code when set), `demo`, and `locale` for one run.
- **Aviator ready**: Automation success state: modal settled, play or demo activated, Spribe launch iframe present, and inner game-ready element visible, within the one-minute bound.
- **Local Aviator stand-in**: Minimal HTML served on loopback HTTP for the ordinary confirmation only; simulates modal + the same ready contract (fake launch iframe host and inner ready marker) without the live site.
- **Browser instance** and **screenshot**: Unchanged from feature 004 except the expected page content after successful `onewin` start.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With a known-good session, three consecutive supervised entries reach **Aviator ready** without manual browser steps (baseline: none today).
- **SC-002**: At least 90% of successful entries with typical connectivity reach **Aviator ready** within one minute of `start` completing its launch steps (baseline: unbounded / manual).
- **SC-003**: The ordinary confirmation completes `start`, `status`, `screenshot`, and `stop` on the local stand-in with zero live-casino requests and a screenshot showing post-ready stand-in content.
- **SC-004**: When `currency` requests a missing code, 100% of attempts in the stand-in test fail with an explicit currency error and leave no running instance.
- **SC-005**: After every failed or ended run in the ordinary confirmation, browser data is removed and session secrets do not appear in project artifacts, command text, or the instance list.
- **SC-006**: An operator on a live machine can use `screenshot` after a successful entry to obtain a private image within 15 seconds while the instance keeps running.

## Assumptions

- Assessment slug `aviator-in-onewin` and decision **go** (2026-11-21) are the authority for scope and out-of-scope items.
- Prior art documented an outer-page modal and an iframe-based ready signal; live `1win.com` DOM was not freshly observed in research. Planning may rediscover selectors for English and localized UIs.
- Legal permission to automate was stated by stakeholders; site terms were not read. A terms conflict stops implementation, not this spec.
- Default when `locale` is omitted: English unprefixed URLs `https://1win.com/` (catalog context) and play URL without locale prefix, per assessment resolution.
- Operator final judgment on live runs remains private; automated ready check is for Supervisor completion, not a warranty of account balance or winnings.
- Features 001–006 behaviors for blank start, CLI syntax, screenshot, status, and stop remain unless explicitly amended here.
- Linux and WSL2 remain the supported operator environments.