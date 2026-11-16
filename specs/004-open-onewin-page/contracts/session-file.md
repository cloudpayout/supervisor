# Contract: Session file

**Feature**: `004-open-onewin-page`

The file is JSON. `start` and the worker read it from the path in `--session-file`. They do not accept this document as an argument.

```json
{
  "cookies": [
    {
      "name": "session-id",
      "value": "secret",
      "domain": "1win.com",
      "path": "/",
      "expires": "2026-12-01T00:00:00Z",
      "maxAge": 3600,
      "secure": true,
      "httpOnly": true,
      "sameSite": "Lax",
      "partitioned": false
    }
  ]
}
```

Rules are in [data-model.md](../data-model.md). `name` and `value` are required. The other cookie fields may be omitted. `cookies` must contain at least one object. Unknown fields are an error.

The example value `secret` is not a real session and must not be committed as one. A test generates its own value at runtime.
