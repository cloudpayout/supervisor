# Internal Contract: Capture Socket (003)

Not an operator interface. The `screenshot` command and the instance worker are the only peers.

**Transport**: Unix domain socket

**Path**: `{XDG state}/supervisor/sockets/<instance-id>.sock`

**Permissions**: parent directory `0700`, socket `0600`

**When it exists**: From the moment the worker has set the registry row to `running` until that worker shuts down. The worker deletes the socket file on the way out.

## Request

The client writes one line and nothing else:

```text
CAPTURE\n
```

Any other first line is an error response. The worker handles one request per connection, then closes.

## Response

Success:

```text
OK <nbytes>\n
<exactly nbytes of PNG>
```

`<nbytes>` is a decimal integer. The payload must start with the PNG signature `\x89PNG`. The worker refuses to send more than 32 MiB.

Failure:

```text
ERR <single-line reason>\n
```

No payload follows `ERR`.

## Worker rules

- Run `chromedp.CaptureScreenshot` on the worker's existing tab context.
- Bound that call with a 10-second child context. Do not cancel the context that owns the browser.
- Do not call navigate, click, viewport emulation, or target close.
- Serialize captures of the same instance with a mutex.
- If the client disconnects, discard the result. Do not change the page.

## Client rules

- Dial only after the registry row is `running` and the output path has been rejected or reserved as writable.
- Stop reading when the CLI's 15-second deadline fires.
- Publish with a temporary file plus `os.Link` only after the full PNG byte count is received. See [cli.md](./cli.md).
- An absent socket, a short read, a non-PNG payload, or `ERR` is `screenshot failed` (exit 1), not "not found" and not "not running".
