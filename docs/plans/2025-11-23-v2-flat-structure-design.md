# C3 Version 2: Flat Structure Design

> Design document for migrating C3 documentation from deep nested folders to a flat structure.

## Problem Statement

Version 1 structure has three pain points:

1. **File navigation** - Hard to find files in deep paths like `components/backend/C3-101-*.md`
2. **ID redundancy** - Numeric prefix `C3-101` already encodes container parentage, nested folders duplicate this
3. **TOC/tooling complexity** - Scripts need complex glob patterns for deep nesting

## Solution: Flatten to Type Folders Only

Remove nested subfolders within `containers/` and `components/`. Use README.md as primary context.

### Version 2 Structure

```
.c3/
├── README.md                     # Primary context (id: context)
├── CTX-actors.md                 # Auxiliary context (optional)
├── TOC.md                        # Auto-generated
├── VERSION                       # Contains "2"
├── containers/
│   ├── C3-1-backend.md          # Flat - no subfolders
│   └── C3-2-frontend.md
├── components/
│   ├── C3-101-db-pool.md        # All components flat
│   ├── C3-102-auth-service.md   # No more backend/ subfolder
│   └── C3-201-api-client.md     # No more frontend/ subfolder
├── adr/
│   └── ADR-001-rest-api.md
└── scripts/
    └── build-toc.sh
```

### Naming Conventions

| Type | Pattern | Example | Location |
|------|---------|---------|----------|
| Primary Context | `context` | `id: context` | `.c3/README.md` |
| Auxiliary Context | `CTX-{slug}` | `CTX-actors` | `.c3/CTX-actors.md` |
| Container | `C3-{N}-{slug}` | `C3-1-backend` | `.c3/containers/C3-1-backend.md` |
| Component | `C3-{NNN}-{slug}` | `C3-101-db-pool` | `.c3/components/C3-101-db-pool.md` |
| ADR | `ADR-{NNN}-{slug}` | `ADR-001-auth` | `.c3/adr/ADR-001-auth.md` |

### README.md Frontmatter

```yaml
---
id: context
title: System Overview
summary: >
  Bird's-eye view of the system, actors, and key interactions.
---
```

## Migration (v1 → v2)

### Automated Transforms

1. **Move nested components to flat:**
   ```
   components/backend/C3-101-*.md  →  components/C3-101-*.md
   components/frontend/C3-201-*.md →  components/C3-201-*.md
   ```

2. **Rename primary context:**
   ```
   CTX-system-overview.md  →  README.md
   ```
   Update frontmatter: `id: CTX-system-overview` → `id: context`

3. **Update internal links:**
   - `](./components/backend/C3-101-` → `](./components/C3-101-`
   - `](./CTX-system-overview.md)` → `](./README.md)`

4. **Bump VERSION:** `1` → `2`

5. **Regenerate TOC:** `.c3/scripts/build-toc.sh`

### Migration Verification

- [ ] No files in `components/*/` subdirectories
- [ ] `README.md` exists with `id: context`
- [ ] VERSION contains `2`
- [ ] No broken internal links
- [ ] TOC reflects new paths

## Tooling Updates

### build-toc.sh

| v1 | v2 |
|----|-----|
| Scans `components/*/*.md` | Scans `components/*.md` |
| Looks for `CTX-*.md` at root | Looks for `README.md` + `CTX-*.md` |

### c3-locate

- Path resolution simplified (no nested folder logic)
- `context` ID maps to `README.md`
- All other ID lookups unchanged

### c3-adopt

- Creates flat `components/` structure
- Generates `README.md` instead of `CTX-system-overview.md`
- Sets `id: context` in frontmatter

### c3-migrate

- Detect v1 vs v2 by checking for nested `components/*/` folders
- Execute transforms listed above
- Validate post-migration

## Files to Update

| File | Change |
|------|--------|
| `MIGRATIONS.md` | Add v2 section with transforms |
| `skills/c3-migrate/SKILL.md` | Add v1→v2 migration logic |
| `skills/c3-adopt/SKILL.md` | Generate v2 structure |
| `skills/c3-locate/SKILL.md` | Simplified path resolution |
| `scripts/build-toc.sh` | Flat glob patterns |
| `README.md` | Document v2 structure |

## Benefits

- ✓ Easier file navigation (flat listing, numeric sort)
- ✓ No ID/folder redundancy (hierarchy encoded in IDs only)
- ✓ Simpler TOC/tooling (single-level globs)
- ✓ Self-documenting `.c3/` folder (README.md at root)

## File Discovery Examples

```bash
# v2: Find all backend components (container 1)
ls .c3/components/C3-1*.md

# v2: Find all frontend components (container 2)
ls .c3/components/C3-2*.md

# v1 equivalent (more complex)
find .c3/components -mindepth 2 -name "C3-*.md"
```
