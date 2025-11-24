---
name: c3-locate
description: Retrieve C3 documentation content by document or heading ID - supports exploration during scoping with ID-based navigation
---

# C3 Locate

## Overview

Retrieve content from `.c3/` documentation by document ID or heading ID. Supports the exploration phase of c3-design by enabling precise, ID-based content lookup.

**Core principle:** IDs are the primary navigation. Use numeric document IDs (c3-0, c3-1, c3-101) and heading IDs (#c3-1-middleware) for precise retrieval.

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
c3-locate c3-0                 # Context (README.md)
c3-locate c3-1                 # Container 1 (c3-1-*/README.md)
c3-locate c3-101               # Component 01 in container 1
c3-locate adr-20251124-caching  # ADR document (date-based)
```

**Returns:**
- Frontmatter (id, title, summary)
- Overview section content
- List of heading IDs for deeper exploration

### By Heading ID

Retrieve specific section from any document:

```
c3-locate #c3-1-middleware
c3-locate #c3-101-configuration
c3-locate #adr-20251124-caching-consequences
```

**Returns:**
- Heading title
- Heading summary (from HTML comment)
- Section content

### By Document + Heading

Retrieve specific section from specific document:

```
c3-locate c3-1 #c3-1-middleware
c3-locate c3-101 #c3-101-error-handling
```

**Returns:**
- Full section content with context

## Implementation

### Finding Documents by ID

```bash
# Document ID patterns (v3 hierarchical):
# Context: c3-0 → .c3/README.md
# Container: c3-{N} → .c3/c3-{N}-*/README.md (folder with README)
# Component: c3-{N}{NN} → .c3/c3-{N}-*/c3-{N}{NN}-*.md (inside container folder)
# ADR: adr-{YYYYMMDD}-{slug} → .c3/adr/adr-{YYYYMMDD}-{slug}.md (date-based)

# Context (v3): c3-0 maps to README.md
if [ "$ID" = "c3-0" ]; then
    find .c3 -maxdepth 1 -name "README.md"
fi

# Container documents (v3): c3-{N} → folder/README.md
# e.g., c3-1 → .c3/c3-1-backend/README.md
if [[ "$ID" =~ ^c3-([0-9])$ ]]; then
    N="${BASH_REMATCH[1]}"
    find .c3 -maxdepth 2 -path ".c3/c3-${N}-*/README.md"
fi

# Component documents (v3): c3-{N}{NN} → inside container folder
# e.g., c3-101 → .c3/c3-1-*/c3-101-*.md
if [[ "$ID" =~ ^c3-([0-9])([0-9]{2})$ ]]; then
    N="${BASH_REMATCH[1]}"
    NN="${BASH_REMATCH[2]}"
    find .c3/c3-${N}-* -name "c3-${N}${NN}-*.md" 2>/dev/null
fi

# ADR documents (v3): date-based
# e.g., adr-20251124-caching → .c3/adr/adr-20251124-caching.md
find .c3/adr -name "adr-*.md"

# V2 fallback patterns (backward compatibility):
# Container: C3-{N}-slug.md in containers/
find .c3/containers -name "C3-[0-9]-*.md" 2>/dev/null
# Component: C3-{N}{NN}-slug.md in components/
find .c3/components -name "C3-[0-9][0-9][0-9]-*.md" 2>/dev/null
# ADR: ADR-###-slug.md (uppercase, legacy)
find .c3/adr -name "ADR-*.md" 2>/dev/null
```

### Extracting Frontmatter

```bash
# Extract frontmatter from document (v3 example)
awk '/^---$/,/^---$/ {print}' .c3/c3-1-backend/README.md

# V2 fallback example
awk '/^---$/,/^---$/ {print}' .c3/containers/C3-1-backend.md
```

### Finding Heading by ID

```bash
# Heading pattern: ## Title {#heading-id}
# Find heading and extract content until next heading

# V3 example (hierarchical structure)
awk -v hid="c3-1-middleware" '
  $0 ~ "{#" hid "}" {
    found = 1
    print
    next
  }
  found && /^## / { exit }
  found { print }
' .c3/c3-1-backend/README.md

# V2 fallback
awk -v hid="c3-1-middleware" '...' .c3/containers/C3-1-backend.md
```

### Extracting Heading Summary

```bash
# Summary is in HTML comment after heading

# V3 example
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
' .c3/c3-1-backend/README.md
```

## During Exploration

Use c3-locate to investigate hypothesis:

```
Hypothesis: "This affects c3-1 middleware"

Exploration:
1. c3-locate c3-1
   → Get overview, see what components exist

2. c3-locate #c3-1-middleware
   → Read middleware section details

3. c3-locate #c3-1-components
   → See component organization

4. Discover: "Ah, this actually touches c3-101-auth-middleware"

5. c3-locate c3-101
   → Deeper exploration of that component
```

## ID Conventions

### V3 Structure (Current)

| Pattern | Level | Example | File Path |
|---------|-------|---------|-----------|
| `c3-0` | Context | c3-0 | `.c3/README.md` |
| `c3-{N}` | Container | c3-1 | `.c3/c3-{N}-*/README.md` |
| `c3-{N}{NN}` | Component | c3-101 | `.c3/c3-{N}-*/c3-{N}{NN}-*.md` |
| `adr-{YYYYMMDD}-{slug}` | Decision | adr-20251124-caching | `.c3/adr/adr-{YYYYMMDD}-{slug}.md` |
| `#c3-xxx-section` | Heading | #c3-1-middleware | (within any doc) |

### V2 Fallback (Backward Compatibility)

| Pattern | Level | Example | File Path |
|---------|-------|---------|-----------|
| `context` | Primary Context | context | `.c3/README.md` |
| `CTX-slug` | Auxiliary Context | CTX-actors | `.c3/CTX-actors.md` |
| `C3-<C>-slug` | Container | C3-1-backend | `.c3/containers/C3-1-backend.md` |
| `C3-<C><NN>-slug` | Component | C3-102-auth | `.c3/components/C3-102-auth.md` |
| `ADR-###-slug` | Decision | ADR-003-cache | `.c3/adr/ADR-###-slug.md` |

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
HYPOTHESIZE → "Affects c3-1"
     ↓
EXPLORE
  ├── c3-locate c3-1 (isolated)
  ├── c3-locate c3-0 #c3-0-containers (upstream)
  ├── c3-locate c3-2 (adjacent)
  └── c3-locate c3-101, c3-102 (downstream)
     ↓
DISCOVER → Refine or confirm hypothesis
```
