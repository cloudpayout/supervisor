# Internal Contract: Watch on the Capture Socket (018)

Not an operator interface. Extends [003 capture-socket](../../003-screenshot-cli/contracts/capture-socket.md). `CAPTURE` is unchanged.

**Transport**: existing Unix domain socket `{XDG state}/supervisor/sockets/<instance-id>.sock`

**Permissions**: parent directory `0700`, socket `0600`

## Request

One line, then the client waits:

```text
WATCH <fps>\n
```

`<fps>` is a decimal integer from 1 to 30. The CLI validates this before dialing. Any other first line that is not `CAPTURE` is an error response.

## Response

Success (connection stays open):

```text
OK http://127.0.0.1:<port>/\n
```

`<port>` is the port actually bound. The HTTP server is already accepting before this line is sent.

Failure (connection then closes):

```text
ERR <single-line reason>\n
```

Reasons include `already watching`, `bad request`, and a short capture/screencast failure string. No URL follows `ERR`.

## HTTP viewer

| Item | Rule |
|------|------|
| Bind | `127.0.0.1` only. First try port 8080, then 8081, … at most 64 tries |
| Method | `GET /` |
| Response | `200`, `Content-Type: multipart/x-mixed-replace; boundary=frame` |
| Part | `Content-Type: image/jpeg` then one JPEG, repeated |
| Rate | At most `<fps>` parts per second. Fewer parts when the page does not change is valid |
| Other paths | `404` |

## Worker rules

- Start screencast on the existing tab context. Do not navigate, resize, or close the target.
- Acknowledge each screencast frame on a goroutine that is not the CDP event callback.
- Stop screencast and close the HTTP listener when the Unix client disconnects or when the worker context is cancelled.
- Do not cancel the browser context from this path.
- `CAPTURE` during an active watch still returns one PNG under the existing mutex and 10-second child timeout.
- If the client disconnects before `OK` is written, do not leave the HTTP listener running.

## Client rules

- Dial only after the registry row is `running` and `--fps` is valid.
- Print the URL only after a full `OK` line.
- On interrupt, close the Unix connection and exit 0. Do not send a second command.
- Treat a closed connection after `OK` as exit 0 (instance ended).
- An absent socket, `ERR`, or a short read before `OK` is `watch failed` (exit 1).
