# CLI Contract: start headless viewport (012 amendment)

**Version**: v6 (012-headless-mode-viewport)  
**Amends**: [005 start](../../005-start-command-syntax/contracts/cli.md)  
**Binary**: `supervisor`  
**Module**: `github.com/cloudpayout/supervisor`

## Delta summary

| Area | Feature 005 | Feature 012 |
|------|-------------|-------------|
| `--headless` semantics | Chrome without a window | Same, plus fixed **1280×720** page layout for that instance |
| New flags | — | None |
| `screenshot` | Current viewport PNG | Headless PNG dimensions **1280×720** when capture succeeds |
| Headed `start` | Visible window | Unchanged (no forced 1280×720) |

`status`, `stop`, `logs`, and `screenshot` argument shapes unchanged except screenshot image size for headless instances.

## `supervisor start --headless`

| Property | Value |
|----------|-------|
| Viewport | Width 1280, height 720 (layout pixels), applied for the life of the instance |
| Persistence | Not stored in registry; implied by `headless: true` on the row |
| Headed default | Omitting `--headless` does **not** apply this viewport |

## Help text (FR-007)

`start --help` MUST document that `--headless` runs without a window **and** uses a 1280×720 layout. Exact wording may appear in the `--headless` flag description and/or command `Long` string.

## `supervisor screenshot <instance-id>` (observable delta)

| Instance mode | PNG size (success) |
|---------------|-------------------|
| Headless | 1280×720 |
| Headed | Not specified by 012 (operator window size; unchanged obligation from 003) |

## Exit codes

No change to start or screenshot exit code table from 005/003.