# Idea Intake: Casino support on start

- **Slug**: casinos-support
- **Created**: 2026-11-15
- **Source**: pasted text; attached CSV `Book_Hoja1_.csv` (columns `Casino`, `URL`, `Cookie name`); codebase pointer to Supervisor `start` (`specs/001-supervisor-cli`, constitution §V)
- **Type**: new-capability

## Idea (as captured)

> Add casino support to Supervisor. For now, it should load the session cookie and navigate to the main page. Attached you a CSV file; it has three columns: casino, the identifier; URL to open the casino page; and cookie, the mandatory cookie to initiate with an authenticated session. Have in mind that some casinos need more than one cookie to authenticate a session. Do not know how to address the issue about pass on information in CLI; but I'm sure that the command `start` will be used for it

Attached CSV `Book_Hoja1_.csv` (header and the only data row; no cookie values were in the file):

```text
Casino,URL,Cookie name
Onewin,https://1win.com/,session-id
```

Related codebase context (not part of the pasted ask): README describes Supervisor as the CLI that runs automations on casino games. Constitution §V requires automations only for explicitly supported casinos and games. `specs/001-supervisor-cli` FR-010 limits the current `start` slice to a blank page and treats casino or game selection as out of scope for that slice. Constitution technical constraints say credentials and session cookies are supplied at runtime and are not committed.

## Restated

The proposal is to add named-casino support to Supervisor so that `start` loads the session cookie (or cookies) for a casino and opens that casino's main page. A sample table names one casino, `Onewin`, with page `https://1win.com/` and cookie name `session-id`, and notes that some casinos need more than one cookie. How that information is passed on the CLI is left open, with `start` named as the command that will do it.

## Origin & Context

- **Raised by**: Oscar Cedano <oscar.cedano@outlook.com>
- **Trigger**: Resolved: We MUST follow the purpose of this project. README describes it.

## First-Glance Unknowns

- Resolved: `start` will receive the casino identifier, page URL, and cookie material from an argument. It can be a JSON string format (e.g '{"casino":"onewin", "cookies": {
    "session-id": "###"
}}') or a relative/absolute path to a JSON file (e.g ./user.json or /home/cedano/user.json).
- Resolved: CSV column `Cookie name` is the cookie's name, it should be set before opening the page URL. The Cookie's value will come from at runtime specifically from a JSON value.
- Resolved: For now, the JSON value should follow the following model: '{"casino": string, "cookies": {
    "cookie-1": "value",
    "cookie-2: "value",
    ...
}}'. Which and how many cookies will be defined by the casino.
- Resolved: The CSV file is an example. But **onewin.com** will be supported.
- Resolved: Before opening the URL, the authentication cookies should be set, in this way, the page will have an authenticated session.
- Resolved: For now, this feature will stop at an authenticated main page. Games and automations are out of scope.
