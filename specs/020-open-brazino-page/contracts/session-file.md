# Contract: Session file (Brazino)

**Feature**: `020-open-brazino-page`
**Extends**: `specs/004-open-onewin-page/contracts/session-file.md`

Brazino uses the same JSON document. The `casino` field must be `brazino`.

```json
{
  "casino": "brazino",
  "cookies": [
    {
      "name": "session-id",
      "value": "secret",
      "domain": "www.brazino777.mx",
      "path": "/",
      "secure": true,
      "httpOnly": true,
      "sameSite": "Lax"
    }
  ]
}
```

For the default test run, `domain` is `127.0.0.1` so cookies bind to the shared stand-in. The example domain above is illustrative for live files; do not commit real secrets.

| Rule | Detail |
|------|--------|
| `casino` | Required. Must be `brazino` for this feature. `casino.Lookup` must succeed. |
| `cookies` | At least one object. Same per-cookie rules as 004. |
| Optional `locale`, `currency`, `demo` | Parsed when present; ignored for navigation when `Game` is empty (Aviator prefs apply only to `onewin`). |

Unknown top-level or cookie fields fail parse. Cookie values must not appear in errors, registry, or worker logs.