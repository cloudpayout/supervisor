# Operator contract: MEMORY limits vs Chrome tuning (014)

**Audience**: Operators using `supervisor start` and `status`  
**Related**: [013 cli.md](../../013-optimizing-browsers/contracts/cli.md)

## `--memory-limit` is not a RAM cap

| Mechanism | What it does | What it does **not** do |
|-----------|----------------|-------------------------|
| `--memory-limit` / `--threshold` | Stores a policy on the instance; `status` shows `current/limit` MiB; NOTE `threshold exceeded` when RSS ≥ warning line | Stop Chrome, throttle RAM, or refuse `start` (v1) |
| Phase-2 Chrome flags (014) | Lower **typical** process RSS via launch options | Guarantee RSS stays below your limit (depends on page + Chrome) |

## Units

- Limits and `status` MEMORY use **MiB** (`600M`, `600MiB`).
- Decimal `600MB` / `600mb` are **rejected** at `start`.

## Examples

```text
# Warning policy only — Chrome may still report 1000+ MiB if the page/browser needs it
supervisor start --headless --memory-limit 600MiB

# More RAM savings attempt (014) — independent of limit
supervisor start --headless

# No lightweight extras (013 or 014)
supervisor start --headless --no-lightweight
```