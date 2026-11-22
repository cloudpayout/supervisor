# Data Model: Headless Viewport 1280×720

**Feature**: `012-headless-mode-viewport` | **Date**: 2026-11-22  
**Registry (unchanged)**: [001 registry schema](../001-supervisor-cli/contracts/registry-schema.json) and [001 data model](../001-supervisor-cli/data-model.md)

No `schema_version` bump. No new JSON fields on `Instance`.

## Persistent model

| Field | Change |
|-------|--------|
| `Instance.Headless` | Unchanged `*bool`; when non-nil and `true`, viewport policy applies at runtime |
| All other instance fields | Unchanged |

**Implied policy (not stored)**: `headless == true` → layout 1280×720 for that run.

## Compile-time constants (`internal/browser`)

| Name | Value | Use |
|------|-------|-----|
| `HeadlessViewportWidth` | `1280` | Allocator `WindowSize`, `EmulateViewport` |
| `HeadlessViewportHeight` | `720` | Allocator `WindowSize`, `EmulateViewport` |

## Runtime entities

### Headless layout policy (per run)

| Attribute | Rule |
|-----------|------|
| Applies when | Registry row has `headless: true` at worker start |
| Width | 1280 CSS/layout pixels |
| Height | 720 CSS/layout pixels |
| Timing | Before cookie set and first `Navigate` for the session |
| Headed | Policy inactive; allocator keeps `VisibleWindow`, no forced metrics |

### Layout observation (tests / validation)

| Signal | Expected when headless |
|--------|------------------------|
| PNG from `screenshot <id>` | Decoded bounds 1280×720 (normal capture success) |
| Registry | `headless: true` on row |

### Worker open action order (headless)

1. `EmulateViewport(1280, 720)`  
2. Cookie `SetCookie` actions (if session file)  
3. `Navigate(page)`  

Headless blank start: steps 1 and 3 only (`about:blank` target).

## State transitions

Unchanged from 001: `starting` → `running` → stop removes row. Viewport is fixed for entire worker lifetime; no resize events in v1.

## Relationships

```text
start --headless
  → registry.Instance.Headless = true
  → worker.Run reads Headless
  → browser.AllocatorOptions(..., headless=true) adds WindowSize
  → chromedp open chain prepends EmulateViewport
  → screenshot capture reads same tab metrics
```

## Non-goals (v1)

- Persisted viewport overrides in profile or registry  
- Per-instance dimension in `status` output  
- Mobile emulation or device scale factor changes  