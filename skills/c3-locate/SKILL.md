---
name: c3-locate
description: Retrieve C3 documentation content by document or heading ID - supports exploration during scoping with ID-based navigation
---

# C3 Locate

## Overview

Retrieve content from `.c3/` documentation by document ID or heading ID. Supports the exploration phase of c3-design by enabling precise, ID-based content lookup.

**Core principle:** IDs are the primary navigation. Use document IDs (CTX-system-overview, C3-2, C3-201) and heading IDs (#c3-1-middleware) for precise retrieval.

## Quick Reference

| Input | Output |
|-------|--------|
| Document ID | Frontmatter + Overview section |
| Heading ID | Section content + heading summary |
| Document + Heading | Specific section from specific document |

## Usage

### By Document ID

Retrieve document frontmatter and overview:

```
c3-locate context              # Primary context (README.md)
c3-locate CTX-actors           # Auxiliary context
c3-locate C3-1-backend
c3-locate C3-102-auth-middleware
c3-locate ADR-003-caching-strategy
```

**Returns:**
- Frontmatter (id, title, summary)
- Overview section content
- List of heading IDs for deeper exploration

### By Heading ID

Retrieve specific section from any document:

```
c3-locate #c3-1-middleware
c3-locate #c3-102-configuration
c3-locate #adr-003-consequences
```

**Returns:**
- Heading title
- Heading summary (from HTML comment)
- Section content

### By Document + Heading

Retrieve specific section from specific document:

```
c3-locate C3-1 #c3-1-middleware
c3-locate C3-102 #c3-102-error-handling
```

**Returns:**
- Full section content with context

## Implementation

### Finding Documents by ID

```bash
# Document ID patterns:
# Context (v2): id "context" maps to README.md
# Context (v1/aux): CTX-slug (e.g., CTX-actors)
# Container: C3-<C>-slug where C is single digit (e.g., C3-1-backend)
# Component: C3-<C><NN>-slug where C is container, NN is 01-99 (e.g., C3-101-db-pool)
# ADR: ADR-###-slug (e.g., ADR-001-caching-strategy)

# Primary context (v2)
if [ "$ID" = "context" ]; then
    find .c3 -maxdepth 1 -name "README.md"
fi

# Auxiliary context documents
find .c3 -maxdepth 1 -name "CTX-*.md"

# Container documents (C3-<digit>-*.md)
find .c3/containers -name "C3-[0-9]-*.md"

# Component documents - check v2 flat first, then v1 nested
# V2 flat:
find .c3/components -maxdepth 1 -name "C3-[0-9][0-9][0-9]-*.md"
# V1 nested (fallback):
find .c3/components -mindepth 2 -name "C3-[0-9][0-9][0-9]-*.md"

# ADR documents
find .c3/adr -name "ADR-*.md"
```

### Extracting Frontmatter

```bash
# Extract frontmatter from document
awk '/^---$/,/^---$/ {print}' .c3/containers/C3-1-backend.md
```

### Finding Heading by ID

```bash
# Heading pattern: ## Title {#heading-id}
# Find heading and extract content until next heading

awk -v hid="c3-1-middleware" '
  $0 ~ "{#" hid "}" {
    found = 1
    print
    next
  }
  found && /^## / { exit }
  found { print }
' .c3/containers/C3-1-backend.md
```

### Extracting Heading Summary

```bash
# Summary is in HTML comment after heading
awk -v hid="c3-1-middleware" '
  $0 ~ "{#" hid "}" {
    getline
    if ($0 ~ /^<!--/) {
      in_comment = 1
      sub(/^<!-- */, "")
      if ($0 !~ /-->$/) { print; next }
    }
  }
  in_comment {
    if ($0 ~ /-->$/) {
      sub(/ *-->$/, "")
      if (length($0) > 0) print
      exit
    }
    print
  }
' .c3/containers/C3-1-backend.md
```

## During Exploration

Use c3-locate to investigate hypothesis:

```
Hypothesis: "This affects C3-1 middleware"

Exploration:
1. c3-locate C3-1
   → Get overview, see what components exist

2. c3-locate #c3-1-middleware
   → Read middleware section details

3. c3-locate #c3-1-components
   → See component organization

4. Discover: "Ah, this actually touches C3-102-auth-middleware"

5. c3-locate C3-102
   → Deeper exploration of that component
```

## ID Conventions

| Pattern | Level | Example | File Path |
|---------|-------|---------|-----------|
| `context` | Primary Context | context | `.c3/README.md` |
| `CTX-slug` | Auxiliary Context | CTX-actors | `.c3/CTX-actors.md` |
| `C3-<C>-slug` | Container | C3-1-backend | `.c3/containers/C3-1-backend.md` |
| `C3-<C><NN>-slug` | Component | C3-102-auth | `.c3/components/C3-102-auth.md` |
| `ADR-###-slug` | Decision | ADR-003-cache | `.c3/adr/ADR-003-cache.md` |
| `#ctx-section` | Context Heading | #ctx-architecture | (within context docs) |
| `#c3-xxx-section` | Container/Component Heading | #c3-1-middleware | (within C3 docs) |

## Fallback: Discovery Mode

When you don't know the ID yet (rare):

1. Read TOC first: `cat .c3/TOC.md`
2. Search for concept in TOC summaries
3. Identify relevant IDs
4. Switch to ID-based lookup

**Prefer ID-based lookup.** Discovery mode is only for initial orientation.

## Integration with c3-design

Called during EXPLORE phase of iterative scoping:

```
HYPOTHESIZE → "Affects C3-1"
     ↓
EXPLORE
  ├── c3-locate C3-1 (isolated)
  ├── c3-locate context #ctx-containers (upstream - v2)
  ├── c3-locate C3-2 (adjacent)
  └── c3-locate C3-101, C3-102 (downstream)
     ↓
DISCOVER → Refine or confirm hypothesis
```
