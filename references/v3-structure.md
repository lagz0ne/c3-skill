# V3 File Structure Reference

Quick reference for C3 documentation file paths and ID patterns in v3 hierarchical structure.

## ID Patterns

| Level | ID Pattern | Example |
|-------|------------|---------|
| Context | `c3-0` | `c3-0` (single root) |
| Container | `c3-{N}` (N=1-9) | `c3-1`, `c3-2` |
| Component | `c3-{N}{NN}` (NN=01-99) | `c3-101`, `c3-215` |
| ADR | `adr-{YYYYMMDD}-{slug}` | `adr-20251124-database`, `adr-20250115-auth` |

## File Paths

| Level | Path Pattern | Example |
|-------|--------------|---------|
| Context | `.c3/README.md` | `.c3/README.md` |
| Container | `.c3/c3-{N}-{slug}/README.md` | `.c3/c3-1-backend/README.md` |
| Component | `.c3/c3-{N}-{slug}/c3-{N}{NN}-{slug}.md` | `.c3/c3-1-backend/c3-101-db-pool.md` |
| ADR | `.c3/adr/adr-{YYYYMMDD}-{slug}.md` | `.c3/adr/adr-20251124-database.md` |

## Directory Layout

```
.c3/
├── README.md              # Context (id: c3-0, c3-version: 3)
├── TOC.md                 # Table of contents
├── index.md               # VitePress index
├── platform/              # Platform docs (Context-level)
│   ├── deployment.md      # c3-0-deployment
│   ├── networking.md      # c3-0-networking
│   ├── secrets.md         # c3-0-secrets
│   └── ci-cd.md           # c3-0-cicd
├── c3-1-{slug}/           # Container 1 folder
│   ├── README.md          # Container doc (id: c3-1)
│   ├── c3-101-*.md        # Component 01 (id: c3-101)
│   └── c3-102-*.md        # Component 02 (id: c3-102)
├── c3-2-{slug}/           # Container 2 folder
│   └── README.md          # Container doc (id: c3-2)
├── adr/                   # Architecture Decision Records
│   └── adr-YYYYMMDD-*.md  # e.g., adr-20251124-database.md
└── scripts/
    └── build-toc.sh       # TOC generator
```

## Frontmatter

### Context (README.md)
```yaml
---
id: c3-0
c3-version: 3
title: System Overview
---
```

### Container
```yaml
---
id: c3-1
c3-version: 3
title: Backend API Container
---
```

### Component
```yaml
---
id: c3-101
c3-version: 3
title: Database Pool Component
---
```

### ADR
```yaml
---
id: adr-20250115-postgresql
title: Use PostgreSQL for persistence
status: accepted
date: 2025-01-15
---
```

## Anchor Format

All headings use lowercase IDs with level prefix:

| Level | Anchor Pattern | Example |
|-------|----------------|---------|
| Context | `{#c3-0-*}` | `{#c3-0-containers}` |
| Container | `{#c3-N-*}` | `{#c3-1-middleware}` |
| Component | `{#c3-NNN-*}` | `{#c3-101-config}` |
| ADR | `{#adr-YYYYMMDD-*}` | `{#adr-20251124-decision}` |

## Platform IDs

Platform documents use Context-level IDs with descriptive suffix:

| Document | ID |
|----------|-------|
| Deployment | c3-0-deployment |
| Networking | c3-0-networking |
| Secrets | c3-0-secrets |
| CI/CD | c3-0-cicd |

These are conceptually part of Context (c3-0), split for manageability.

## Key Rules

1. **All lowercase** - IDs, slugs, filenames are lowercase
2. **Numeric IDs are stable** - Slugs can change, numbers don't
3. **Components encode parent** - `c3-101` = container 1, component 01
4. **Containers own folders** - Each container is a directory
5. **Version in frontmatter** - `c3-version: 3` in context README.md

## Search Patterns (bash)

```bash
# Find context
ls .c3/README.md

# Find all containers
ls -d .c3/c3-[1-9]-*/

# Find all components in container 2
ls .c3/c3-2-*/c3-2*.md

# Find all ADRs
ls .c3/adr/adr-[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]-*.md

# Find ADRs by year
ls .c3/adr/adr-2025*.md
```
