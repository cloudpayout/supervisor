# Quickstart: Open the onewin page

Validation for `specs/004-open-onewin-page/spec.md`. Contracts: [cli.md](./contracts/cli.md), [session-file.md](./contracts/session-file.md). Fields: [data-model.md](./data-model.md).

## Prerequisites

- Linux or WSL2
- Go 1.27.1, from `go.mod`
- Google Chrome on `PATH`, the same dependency the default checks job installs
- No Python. The stand-in is part of `go test`.

## Ordinary confirmation

Run from the repository root:

```text
go test -json ./...
```

The new test does all of the following. It is part of that default run.

1. Start a local HTTP server on `127.0.0.1` and an ephemeral port. `GET /` includes the marker `balance` only when the `Cookie` header contains the name and value generated for this run.
2. Write a session file in a temporary directory. The value is random for the run. The cookie's domain is `127.0.0.1`. The file is not under the repository.
3. Set `SUPERVISOR_PAGE_URL` to `http://127.0.0.1:<port>/`.
4. Run `supervisor start --casino onewin --session-file <temp> --headless`. Expect exit `0` and an instance id. The stand-in records that it showed `balance`.
5. Run `supervisor status` once. Expect that id in the list as running.
6. Run `supervisor screenshot <id>`. Expect exit `0` within 15 seconds and one PNG. A pixel from that PNG matches the stand-in's "shown" color. The file name does not contain the generated cookie value. `status` still lists the same id as running.
7. Run `supervisor stop <id>`. Expect exit `0`. Run `status` again. Expect the id to be gone. Expect the profile directory to be gone.
8. Repeat steps 4 through 7 with a session file that does not contain the expected cookie. Expect the stand-in to withhold `balance`, and expect a pixel from the screenshot to match the "withheld" color.
9. Read the worker command line, the screenshot command line, the worker log, the registry file, and the screenshot file name. Expect the generated value to be absent from all of them. The session file path may appear on the `start` command line.
10. Expect no navigation to `1win.com` in this test. The page URL is the stand-in. Do not leave the PNG in the repository.

A second case in the same suite sends two cookies and expects both names to arrive at the stand-in. Names are not secrets. The values are still absent from the command line, the log, and the registry.

## Usage failures

These do not need a browser:

| Command | Result |
|---------|--------|
| `start --casino Onewin --session-file <file>` | Exit `2`, no id |
| `start --casino 1win.com --session-file <file>` | Exit `2`, no id |
| `start --casino onewin` | Exit `2`, no id |
| `start --session-file <file>` | Exit `2`, no id |
| `start --casino onewin --session-file <missing>` | Non-zero, no id, no profile |
| `start --casino onewin --session-file <empty cookies>` | Non-zero, no id, no profile |

`start` with no casino flags still opens a blank page.

## Live page

Do not put this in `go test ./...`.

Leave `SUPERVISOR_PAGE_URL` unset. Do not pass `--headless`. Run `supervisor start --casino onewin --session-file <the operator's own file>` so the default visible window opens. The page is `https://1win.com/`. The operator looks at it. `supervisor screenshot <id>` writes an image of that visible page on the operator's machine. Supervisor stores no record of the look, and the image is not committed. The default `go test` run stays headless and does not perform this check. If the site rejects the file, stop. Do not commit the file, and do not change this feature to hunt for a working session or to start `aviator`.
