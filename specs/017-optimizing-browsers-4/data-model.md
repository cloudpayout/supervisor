# Data Model: Headless Shell and Performance Flags (017)

**Spec**: [spec.md](./spec.md)

## Scope

No new registry fields or CLI flags (**FR-010**). Changes: headless **browser binary** selection, **015** profile restoration, optional **phase-4** Chrome flags, **ephemeral** profile lifecycle guarantees.

## Entities (unchanged shape)

| Entity | 017 change |
|--------|------------|
| `Instance` / registry row | `profile_dir` still recorded while alive; directory MUST NOT survive after teardown |
| `lightweight` | Unchanged semantics; `true` → 013 + 014 + 015 + **phase-4 adopt** on headless |
| `headless` registry flag | When true, worker uses **chrome-headless-shell**, not full Chrome |
| Memory policy / NOTE | Unchanged (**013**) |

## Binary selection

| Launch mode | Binary resolver | `CHROME_PATH` |
|-------------|-----------------|---------------|
| `start --headless` | Project `bin/chrome-headless-shell` | **Ignored** for selection |
| `start` (headed) | `Executable()` — `CHROME_PATH` then PATH | Honored |

## Allocator tiers (updated)

| Condition | Chrome extras |
|-----------|----------------|
| `headless && lightweight` | 013 + 014 + 015 + **017 phase-4** (if adopt non-empty); **no** 016 experimental |
| `headless && !lightweight` | No lightweight tiers; WebGL + SwiftShader on headless base |
| `headed && lightweight` | Headed-safe 013 only |
| `headed && !lightweight` | No lightweight extras |

## Ephemeral profile lifecycle

```text
start → Mkdir profile/<id> → running → stop / worker exit / reconcile → RemoveProfile(profile/<id>)
```

| State | Profile on disk |
|-------|-----------------|
| `starting` / `running` / `stopping` | MAY exist; MAY grow during run |
| After cleanup | MUST NOT exist (**FR-013**) |

## Baseline artifact

| Field | Location |
|-------|----------|
| 015 vs 017 medians, shell version, stand-ins, SC-006/007 | [research.md](./research.md) Baseline Record (017) |