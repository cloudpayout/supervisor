# Research: Open the onewin page

## Decision: Keep a blank `start` when no casino is named

**Decision**: `supervisor start` with neither `--casino` nor `--session-file` keeps today's behavior: one browser on `about:blank`. Naming `onewin` requires `--session-file`. Passing only one of those two flags is a usage error.

**Rationale**: Feature 001's start contract and its tests have no casino argument. This feature adds a casino path without taking the blank path away.

**Alternatives considered**: Making `--casino` required on every `start` was rejected because it would break the existing blank-page command.

## Decision: Session file path on the command, document only in the file

**Decision**: The operator passes `--session-file` with a filesystem path. `start` parses the file to reject a bad document before it registers an instance. The worker is spawned as `worker run --id <id> [--headless] --session-file <path>` and reads the file itself. The JSON document is never an argument. Cookie values are not written to the registry, the worker log, or error text.

**Rationale**: On Linux the command line is readable at `/proc/<pid>/cmdline`, and this program already reads other processes' command lines. A path is not the secret. The registry is the instance list the spec forbids secrets from. The worker log is the run's recorded output.

**Alternatives considered**: Passing the JSON on `start`'s arguments was rejected because the secret would be visible in the command line and in shell history. Copying the document into the registry or the state directory was rejected because that keeps the secret after the operator's file is the source. A pipe from parent to worker was rejected because the path is enough and adds a second channel to keep off the log.

## Decision: Cookies are set, then the registered page is opened

**Decision**: The worker calls chromedp's `Network.setCookie` once per cookie, then navigates. For `onewin` the page is `https://1win.com/` unless the test seam below overrides it. The pinned client (`chromedp` v0.23.0, `cdproto` v0.159.0) returns an empty result from `SetCookie`, so a successful call is not proof the browser stored the cookie. A call that returns an error fails the instance: the worker removes the registry row and the profile.

**Rationale**: The published chromedp cookie example sets cookies, then navigates, against a local server. The spec requires every cookie to be applied before the page opens. The empty result type was read from the pinned `cdproto` during assessment.

**Alternatives considered**: Navigating first and setting cookies after was rejected because the page would load without the session. Treating a successful RPC as proof that `1win.com` accepted the cookie was rejected because the result carries no success flag and the live site was not observed.

## Decision: Catalog holds the casino, the page, and the game name

**Decision**: One Go catalog entry: identifier `onewin`, page `https://1win.com/`, game `aviator`. Identifiers must match `^[a-z]+$`. `Onewin` and `1win.com` are usage errors. No flag starts `aviator`. The instance row may store `casino: "onewin"`. It must not store a URL copied from the operator, a cookie, or the session file path.

**Rationale**: Constitution 1.2.0 allows a casino to be supported before any of its games can be run, and still requires explicit support and lowercase identifiers. The page address is registered with the casino so the operator does not type it. The game name is data, not a command.

**Alternatives considered**: Taking the page URL as a flag was rejected because the spec says the address lives with the casino, and a free URL would let `start` open sites the product does not support.

## Decision: Test seam for the local page

**Decision**: When the environment variable `SUPERVISOR_PAGE_URL` is set, the worker opens that URL instead of the catalog page. Operator help does not document it. The ordinary test sets it to `http://127.0.0.1:<port>/`. Unset, `onewin` opens `https://1win.com/`.

**Rationale**: The ordinary confirmation must run `start` against a local stand-in and must not open the live casino. The worker is a separate process, so the test cannot pass the stand-in address through a function argument. The variable is not a supported operator interface.

**Alternatives considered**: A public `--page-url` flag was rejected because it would make an arbitrary page part of the operator contract. Hard-coding the stand-in as the `onewin` page was rejected because the registered page is `https://1win.com/`.

## Decision: The stand-in is a Go HTTP server in the test

**Decision**: The ordinary confirmation is a `net/http` server bound to `127.0.0.1` with an ephemeral port, started by the Go test. `GET /` returns a page that includes the marker `balance` only when the request's `Cookie` header contains the name and value the test generated for this run. Otherwise the marker is absent. The server records whether it showed or withheld the marker. It does not write the cookie value to the test log. The test generates the value at runtime into a temporary file and does not commit it.

**Rationale**: The constitution says the implementation language is Go. The default checks job installs Go and Chrome and runs `go test ./...`. A Python FastAPI server would add a second language and a job change the constitution does not allow. The spec left the tool to this plan and required the behavior: local page, balance only when the expected cookie for that host is present, then `status` and `stop`, no live casino.

**Alternatives considered**: FastAPI, as recorded in the assessment, was rejected because it is not Go. Driving a real `1win.com` login in the default test was rejected by the constitution and by the spec.

## Decision: Delete the profile only after the browser has exited

**Decision**: `browser.RemoveProfile` runs after `proc.Terminate` and `proc.KillByProfile` have returned, on every path that ends an instance: failed `start` (already implemented), live `stop`, reconciliation in `registry.Apply`, and worker exit. `RemoveProfile` stays idempotent. The worker cancels the browser allocator and waits until the browser PID is not alive before deleting the directory. Cookie values are not read back from the profile before deletion.

**Rationale**: `Terminate` already waits up to five seconds and then sends `SIGKILL`. Deleting while Chrome is still writing the profile was the open risk. The spec requires the session material to be gone when the instance ends, including a failed start and a browser that exits outside `stop`.

**Alternatives considered**: Deleting the profile at the start of `stop`, before the signals, was rejected because the browser would still be writing it. Leaving deletion only on failed `start`, which is today's code, was rejected because a successful run would keep the session on disk.

## Decision: Cookie attributes in the file

**Decision**: The file is a JSON object with a `cookies` array of one or more objects. Required fields are `name` and `value`, both non-empty strings. Optional fields are `domain`, `path`, `expires` (RFC3339), `maxAge` (integer seconds), `secure`, `httpOnly`, `sameSite` (`Strict`, `Lax`, or `None`), and `partitioned` (boolean). Unknown fields fail the parse. When `expires` is set it is the browser expiry. Otherwise `maxAge` is added to the time of the read. When neither is set, the cookie is a session cookie. `sameSite` values outside the three names fail the parse.

**Rationale**: The stakeholder named those fields. `Network.setCookie` in the pinned `cdproto` accepts domain, path, secure, HTTP-only, SameSite, expiry, and a partition key. An empty `cookies` array fails `start` before a browser is launched. The live site's required flags were not observed, so the file carries what the operator supplies rather than a guessed `1win.com` policy.

**Alternatives considered**: A map of name to value was rejected because it drops domain, expiry, and the other flags the browser uses to store a cookie. A `casino` field inside the file was rejected because the casino is already the `--casino` flag and a second copy can disagree.

## Decision: Reuse the existing screenshot command

**Decision**: `supervisor screenshot <id>` stays the command specified in `specs/003-screenshot-cli`. This feature does not change its arguments, its 15-second deadline, its output path, its failure text, or the capture socket. After an `onewin` start, the worker is already showing the opened page, so a capture of that running instance is a capture of that page. The ordinary test calls `screenshot` after `status` and before `stop`. The image file name and the screenshot command line must not contain the cookie value. The test does not commit the image.

**Rationale**: The command already writes one PNG of the instance's current page and leaves the instance running. The new requirement is which page is current, and that is the worker navigation in this feature.

**Alternatives considered**: A new capture flag that accepts a casino identifier was rejected because the operator already has the instance ID from `start`. Reading the word "balance" out of the PNG with a new dependency was rejected. The stand-in page uses one full-viewport color when the cookie matches and a different full-viewport color when it does not, and the test samples a pixel with the standard-library PNG decoder. The matching page also includes the text `balance` for a person looking at the image. The withheld page does not.
