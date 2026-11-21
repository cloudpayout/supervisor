# Contract: Session file (007 amendment)

**Feature**: `007-enter-aviator-onewin`  
**Amends**: [005 session file](../../005-start-command-syntax/contracts/session-file.md)

```json
{
  "casino": "onewin",
  "currency": "MXN",
  "demo": false,
  "locale": "es-MX",
  "cookies": [
    {
      "name": "session-id",
      "value": "secret",
      "domain": "1win.com",
      "path": "/"
    }
  ]
}
```

| Field | Rule |
|-------|------|
| `casino` | Unchanged from 005 |
| `cookies` | Unchanged from 005 |
| `currency` | Optional. If present: three-letter code `A–Z`, matched as substring of visible modal option text on live or stand-in |
| `demo` | Optional. Default **false** → real play. **true** → demo mode control |
| `locale` | Optional. If absent: unprefixed English play URL. If present: single segment (e.g. `es-MX`) prefixed in play URL path |

Omitting `currency`, `demo`, and `locale` yields defaults; documents valid under 004/005 remain valid.

Unknown top-level fields: parse error (unchanged).

Values such as `secret` are examples only and must not be committed.