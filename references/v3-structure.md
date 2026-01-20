# V3 Structure Reference

Single source of truth for C3 v3 documentation structure.

## Contents

- [Core Concept](#core-concept)
- [ID Patterns](#id-patterns)
- [File Paths](#file-paths)
- [Directory Layout](#directory-layout)
- [Required Frontmatter](#required-frontmatter)
- [Required Sections](#required-sections)
- [Validation Rules](#validation-rules)
- [Search Patterns](#search-patterns)

## Core Concept

**Context IS the README.** There is no separate "c3-0" file - the `.c3/README.md` file IS c3-0.

## ID Patterns

| Level | Pattern | File | Note |
|-------|---------|------|------|
| Context | `c3-0` | `.c3/README.md` | Context = README |
| Container | `c3-{N}` (N=1-9) | `.c3/c3-N-slug/README.md` | parent: c3-0 |
| Component | `c3-{N}{NN}` (NN=01-99) | `.c3/c3-N-slug/c3-NNN-slug.md` | parent: c3-N |
| ADR | `adr-{YYYYMMDD}-{slug}` | `.c3/adr/adr-*.md` | affects: [c3-N...] |
| Refs | `ref-{slug}` | `.c3/refs/ref-{slug}.md` | system-wide |

## File Paths

| Level | Path |
|-------|------|
| Context | `.c3/README.md` |
| Container | `.c3/c3-{N}-{slug}/README.md` |
| Component | `.c3/c3-{N}-{slug}/c3-{N}{NN}-{slug}.md` |
| ADR | `.c3/adr/adr-{YYYYMMDD}-{slug}.md` |
| Refs | `.c3/refs/ref-{slug}.md` |
| Settings | `.c3/settings.yaml` |
| TOC | `.c3/TOC.md` |

## Directory Layout

```
.c3/
├── README.md                        # Context (c3-0)
├── TOC.md                           # Table of contents
├── settings.yaml                    # Project settings
├── refs/                            # References folder
│   └── ref-{slug}.md               # Reference docs
├── c3-1-{slug}/                     # Container 1
│   ├── README.md                    # Container doc
│   ├── c3-101-{slug}.md             # Component
│   └── c3-102-{slug}.md             # Component
├── c3-2-{slug}/                     # Container 2
│   └── README.md
└── adr/                             # ADR folder
    └── adr-YYYYMMDD-{slug}.md
```

---

## Required Frontmatter

### Context (c3-0)

```yaml
---
id: c3-0                    # REQUIRED: always c3-0
c3-version: 3               # REQUIRED: always 3
title: {System Name}        # REQUIRED
summary: {One line}         # REQUIRED
---
```

### Container (c3-N)

```yaml
---
id: c3-{N}                  # REQUIRED: c3-1, c3-2, etc.
c3-version: 3               # REQUIRED
title: {Container Name}     # REQUIRED
type: container             # REQUIRED
parent: c3-0                # REQUIRED
summary: {One line}         # REQUIRED
---
```

### Component (c3-NNN)

```yaml
---
id: c3-{N}{NN}              # REQUIRED: c3-101, c3-102, etc.
c3-version: 3               # REQUIRED
title: {Component Name}     # REQUIRED
type: component             # REQUIRED
parent: c3-{N}              # REQUIRED: parent container
category: foundation | feature  # REQUIRED
summary: {One line}         # REQUIRED
---
```

### ADR

```yaml
---
id: adr-{YYYYMMDD}-{slug}   # REQUIRED
type: adr                   # REQUIRED
status: proposed            # REQUIRED: proposed|accepted|implemented
title: {Decision Title}     # REQUIRED
affects: [c3-1, c3-2]       # REQUIRED: affected containers
---
```

### Reference (ref-slug)

```yaml
---
id: ref-{slug}              # REQUIRED: ref-state-management, ref-logging, etc.
title: {Reference Name}     # REQUIRED
---
```

---

## Required Sections

### Context (c3-0)

| Section | Purpose |
|---------|---------|
| Overview | What the system does |
| Containers | Table: ID, Name, Purpose |
| Container Interactions | Mermaid diagram |
| External Actors | Who/what interacts with system |

### Container (c3-N)

| Section | Purpose |
|---------|---------|
| Technology Stack | Table: Layer, Technology, Purpose |
| Components | Table: ID, Name, Type, Responsibility, Status |
| Internal Structure | Mermaid diagram (optional but recommended) |

**Status Column Values:**
- *(empty)* - Conventions maturing, may document later
- `Documented` - Full component doc exists
- `Skip: {reason}` - Will never need a doc (e.g., `Skip: stdlib wrapper`)
- `Combined with c3-XXX` - Documented together with another component

### Component (c3-NNN)

| Section | Purpose |
|---------|---------|
| Goal | Why this component exists |
| (flexible) | Sections that serve the Goal |
| References | Code symbols, ref-* citations |

### ADR

| Section | Purpose |
|---------|---------|
| Context | WHY decision needed |
| Decision | WHAT was decided |
| Rationale | WHY this option |
| Consequences | Positive AND negative |
| Changes Across Layers | What docs update |
| Verification Checklist | How to confirm |
| Audit Record | Lifecycle tracking |

---

## Validation Rules

| Rule | Check |
|------|-------|
| All IDs lowercase | `c3-1` not `C3-1` |
| Container folders match ID | `c3-1-*` folder for `id: c3-1` |
| Component parent matches folder | `c3-101` in `c3-1-*/` folder |
| Slugs are lowercase-hyphenated | `api-backend` not `API_Backend` |
| c3-version is 3 | No other versions |
| Components encode parent | First digit = container number |

---

## Search Patterns

```bash
# Find context
cat .c3/README.md

# Find all containers
ls -d .c3/c3-[1-9]-*/

# Find components in container 1
ls .c3/c3-1-*/c3-1*.md

# Find all ADRs
ls .c3/adr/adr-*.md

# Validate IDs
grep -h "^id:" .c3/**/*.md
```
