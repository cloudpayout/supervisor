# Operator flag matrix (phase 3 input)

**Source**: [spec.md](../spec.md) Input  
**Authoritative decisions**: [research.md](../research.md) §R2

| Operator proposal | Disposition | Rationale (short) |
|-------------------|-------------|-------------------|
| `--disable-gpu` | Reject | WebGL/SwiftShader stack |
| `--disable-webgl` | Reject | Conflicts with enabled WebGL |
| `--disable-3d-apis` | Reject | Game canvas |
| `--js-flags=--max-old-space-size=128 …` | Reject | FR-005 bundle |
| `--process-per-site` | Skip | 013 |
| `--disable-application-cache` | Skip | 013 |
| `--v8-cache-options=none` | Skip | 013 |
| `--single-process` | Reject | Stability |
| `--memory-pressure-off` | Adopt | Only safe net-new candidate |
| `--disable-background-networking` | Reject | Bug **-4** |

**Shipped profile**: 013 + 014 + **adopt** rows only.