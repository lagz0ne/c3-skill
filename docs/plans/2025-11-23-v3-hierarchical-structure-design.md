# C3 Version 3: Hierarchical Structure Design

> Design document for migrating C3 from flat type folders to hierarchical container-as-folder structure.

## Problem Statement

Version 2 still has redundant type folders (`containers/`, `components/`) when the ID encoding already expresses the hierarchy.

## Solution: Containers as Folders

Containers become folders containing their components. Parent-child relationship is explicit via filesystem.

### Version 3 Structure

```
.c3/
├── README.md                      # Context (id: c3-0, c3-version: 3)
├── actors.md                      # Aux context (no ID, applies downward)
├── TOC.md                         # Auto-generated
├── adr/
│   └── adr-001-rest-api.md
├── scripts/
│   └── build-toc.sh
├── c3-1-backend/                  # Container folder
│   ├── README.md                  # Container doc (id: c3-1)
│   ├── auth-flows.md              # Aux (applies to container+children)
│   ├── c3-101-db-pool.md          # Component (id: c3-101)
│   └── c3-102-auth-service.md     # Component (id: c3-102)
└── c3-2-frontend/
    ├── README.md                  # Container doc (id: c3-2)
    ├── c3-201-api-client.md       # Component (id: c3-201)
    └── c3-202-state-manager.md    # Component (id: c3-202)
```

### ID & Naming Conventions

| Level | Folder/File | ID | Example |
|-------|-------------|-----|---------|
| Context | `.c3/README.md` | `c3-0` | `id: c3-0` |
| Container | `c3-{N}-{slug}/README.md` | `c3-{N}` | `id: c3-1` |
| Component | `c3-{N}{NN}-{slug}.md` | `c3-{N}{NN}` | `id: c3-101` |
| ADR | `adr/adr-{NNN}-{slug}.md` | `adr-{NNN}` | `id: adr-001` |

**Key rules:**
- IDs are numeric only (stable for linking)
- Slugs are descriptive (can be renamed)
- All lowercase
- Container number encoded in component ID (101 = container 1, component 01)

**Auxiliary docs:**
- No ID required
- Named descriptively: `actors.md`, `auth-flows.md`
- Placed where they apply downward

### Context Frontmatter

```yaml
---
id: c3-0
c3-version: 3
title: System Overview
summary: >
  Bird's-eye view of the system...
---
```

## Migration (v2 → v3)

### Transforms

1. **Convert containers to folders:**
   ```
   containers/C3-1-backend.md → c3-1-backend/README.md
   ```
   - Create folder from container name (lowercase)
   - Move container doc to README.md inside folder
   - Update `id: C3-1-backend` → `id: c3-1`

2. **Move components into container folders:**
   ```
   components/C3-101-db-pool.md → c3-1-backend/c3-101-db-pool.md
   ```
   - Lowercase filename
   - Update `id: C3-101-db-pool` → `id: c3-101`

3. **Update context:**
   - `id: context` → `id: c3-0`
   - Add `c3-version: 3` to frontmatter

4. **Remove VERSION file** (now in README.md frontmatter)

5. **Update internal links:**
   - `](./containers/C3-1-backend.md)` → `](./c3-1-backend/)`
   - `](./components/C3-101-*.md)` → `](./c3-1-backend/c3-101-*.md)`

6. **Lowercase ADRs:**
   - `ADR-001-*.md` → `adr-001-*.md`
   - `id: ADR-001-*` → `id: adr-001`

### Verification

- `.c3/README.md` contains `c3-version: 3`
- No `containers/` or `components/` directories
- Container folders match `c3-[0-9]-*/`
- Each container folder has `README.md`
- Components inside container folders match `c3-[0-9][0-9][0-9]-*.md`
- All IDs are numeric only (no slugs)

## Tooling Updates

### build-toc.sh

| v2 | v3 |
|----|-----|
| Scan `containers/*.md` | Scan `c3-[0-9]-*/README.md` |
| Scan `components/*.md` | Scan `c3-[0-9]-*/c3-[0-9]*.md` |
| Read VERSION file | Read `c3-version` from README.md |
| Uppercase patterns | Lowercase patterns |

### c3-locate

- `c3-0` → `.c3/README.md`
- `c3-1` → `.c3/c3-1-*/README.md`
- `c3-101` → `.c3/c3-1-*/c3-101-*.md`

### c3-adopt

- Create `c3-{N}-{slug}/` folders
- Put container doc as `README.md` inside folder
- Components go directly in container folder

### c3-migrate

- Add v2→v3 transforms
- Handle VERSION file → frontmatter migration
- Lowercase all names

## Benefits

- No redundant type folders (`containers/`, `components/`)
- Parent-child explicit via filesystem
- IDs stable (can rename slugs freely)
- Fewer files (no VERSION)
- Auxiliary docs scoped by location
