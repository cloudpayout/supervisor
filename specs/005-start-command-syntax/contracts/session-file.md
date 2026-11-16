# Contract: Session file (005 amendment)

**Feature**: `005-start-command-syntax`  
**Amends**: [004 session file](../../004-open-onewin-page/contracts/session-file.md)

The file is JSON. `start` passes only the filesystem path as the positional argument. The worker still reads the path from the hidden `--session-file` flag. Neither accepts the JSON document as an argument.

```json
{
  "casino": "onewin",
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

| Field | Rule |
|-------|------|
| `casino` | Required when a path is passed to `start`. Must be a registered identifier (`onewin`). `Onewin`, `1win.com`, and unknown ids fail. |
| `cookies` | Same as feature 004: length ≥ 1; cookie field rules unchanged. |

Unknown fields are an error. `name` and `value` on each cookie are required. The example value `secret` is not a real session and must not be committed.