# Contract: chrome-headless-shell binary (017)

**Package**: `internal/browser`, `internal/cli/start`  
**Spec**: [spec.md](../spec.md) **FR-005**–**FR-007**, **SC-002**

## Artifact

| Item | Value |
|------|--------|
| Path | `{repo}/bin/chrome-headless-shell` |
| Provision | `scripts/fetch-chrome-headless-shell.sh` |
| Version pin | **157.0.8103.78** (CfT chrome-headless-shell-linux64); [research.md](../research.md) R4 + Baseline Record |
| VCS | **Gitignored** (not committed) |

## Resolution rules

| Command | Binary | `CHROME_PATH` |
|---------|--------|---------------|
| `supervisor start --headless` | **Only** `bin/chrome-headless-shell` | **Ignored** (optional stderr notice if set) |
| `supervisor start` (headed) | `CHROME_PATH` if set, else PATH discovery | **Honored** |

## Preflight

1. Headless: file exists, executable, `--version` succeeds (or equivalent smoke).
2. Missing artifact: non-zero exit **before** registry write; message cites fetch script.
3. Headless MUST NOT fall back to `google-chrome` on PATH.

## Verification

- Integration test: headless start with `CHROME_PATH=/usr/bin/google-chrome` still launches bundled shell (argv or process line contains `chrome-headless-shell`).
- `start_launch_failure` style test: missing `bin/chrome-headless-shell` → clear error, no instance ID.