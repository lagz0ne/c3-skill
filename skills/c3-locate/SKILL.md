---
name: c3-locate
description: Retrieve C3 documentation content by document or heading ID - supports exploration during scoping with ID-based navigation
---

# C3 Locate

## Overview

Retrieve content from `.c3/` documentation by document ID or heading ID. Supports the exploration phase of c3-design by enabling precise, ID-based content lookup.

**Core principle:** IDs are the primary navigation. Use document IDs (CTX-001, CON-002, COM-003) and heading IDs (#con-001-middleware) for precise retrieval.

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
c3-locate CTX-001
c3-locate CON-001-backend
c3-locate COM-002-auth-middleware
c3-locate ADR-003-caching-strategy
```

**Returns:**
- Frontmatter (id, title, summary)
- Overview section content
- List of heading IDs for deeper exploration

### By Heading ID

Retrieve specific section from any document:

```
c3-locate #con-001-middleware
c3-locate #com-002-configuration
c3-locate #adr-003-consequences
```

**Returns:**
- Heading title
- Heading summary (from HTML comment)
- Section content

### By Document + Heading

Retrieve specific section from specific document:

```
c3-locate CON-001 #con-001-middleware
c3-locate COM-002 #com-002-error-handling
```

**Returns:**
- Full section content with context

## Implementation

### Finding Documents by ID

```bash
# Document ID pattern: {TYPE}-{NNN}-{slug}
# TYPE: CTX, CON, COM, ADR

# Context documents
find .c3 -maxdepth 1 -name "CTX-*.md"

# Container documents
find .c3/containers -name "CON-*.md"

# Component documents
find .c3/components -name "COM-*.md"

# ADR documents
find .c3/adr -name "ADR-*.md"
```

### Extracting Frontmatter

```bash
# Extract frontmatter from document
awk '/^---$/,/^---$/ {print}' .c3/containers/CON-001-backend.md
```

### Finding Heading by ID

```bash
# Heading pattern: ## Title {#heading-id}
# Find heading and extract content until next heading

awk -v hid="con-001-middleware" '
  $0 ~ "{#" hid "}" {
    found = 1
    print
    next
  }
  found && /^## / { exit }
  found { print }
' .c3/containers/CON-001-backend.md
```

### Extracting Heading Summary

```bash
# Summary is in HTML comment after heading
awk -v hid="con-001-middleware" '
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
' .c3/containers/CON-001-backend.md
```

## During Exploration

Use c3-locate to investigate hypothesis:

```
Hypothesis: "This affects CON-001 middleware"

Exploration:
1. c3-locate CON-001
   → Get overview, see what components exist

2. c3-locate #con-001-middleware
   → Read middleware section details

3. c3-locate #con-001-components
   → See component organization

4. Discover: "Ah, this actually touches COM-002-auth-middleware"

5. c3-locate COM-002
   → Deeper exploration of that component
```

## ID Conventions

| Pattern | Level | Example |
|---------|-------|---------|
| `CTX-NNN-slug` | Context | CTX-001-system-overview |
| `CON-NNN-slug` | Container | CON-001-backend |
| `COM-NNN-slug` | Component | COM-002-auth-middleware |
| `ADR-NNN-slug` | Decision | ADR-003-caching-strategy |
| `#doc-nnn-section` | Heading | #con-001-middleware |

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
HYPOTHESIZE → "Affects CON-001"
     ↓
EXPLORE
  ├── c3-locate CON-001 (isolated)
  ├── c3-locate CTX-001 #ctx-001-containers (upstream)
  ├── c3-locate CON-002 (adjacent)
  └── c3-locate COM-001, COM-002 (downstream)
     ↓
DISCOVER → Refine or confirm hypothesis
```
