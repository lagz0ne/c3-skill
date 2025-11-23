---
name: c3-toc
description: Review and update C3 Table of Contents - manage document tree structure, inspect downstream documents, and rebuild TOC
---

# C3 Table of Contents Management

## Overview

Manage the `.c3/TOC.md` file which serves as the navigation index for all C3 documentation. The TOC provides:
- Tree-like structure showing document relationships
- Entry point for hypothesis formation during scoping
- Quick reference for document discovery

**Announce at start:** "I'm using the c3-toc skill to manage the documentation index."

## TOC Structure

### Document Tree

The TOC reflects the physical organization of `.c3/`:

```
.c3/
├── TOC.md                           # This index
├── CTX-*.md                         # Context level (root)
├── containers/
│   ├── C3-1-backend.md              # Container 1
│   └── C3-2-frontend.md             # Container 2
├── components/
│   ├── backend/                     # Grouped by container
│   │   ├── C3-101-db-pool.md        # Container 1, Component 01
│   │   └── C3-102-auth-service.md   # Container 1, Component 02
│   └── frontend/
│       └── C3-201-ui-router.md      # Container 2, Component 01
└── adr/
    ├── ADR-001-*.md
    └── ADR-002-*.md
```

### Logical Hierarchy

```
CTX (Context)
 └── C3-<C> (Container, e.g., C3-1-backend)
      └── C3-<C><NN> (Component, e.g., C3-101-db-pool)

ADR (Decisions) - Cross-cutting, linked to affected levels
```

## TOC Responsibilities

### 1. Document Discovery

TOC is the first place to look when:
- Starting c3-design exploration
- Forming initial hypothesis
- Finding related documents

### 2. Downstream Inspection

TOC helps identify what documents exist downstream:

| From | Downstream |
|------|------------|
| Context | All Containers, all Components |
| Container | Components in that container's directory |
| Component | None (leaf level) |
| ADR | Documents it affects (via Related section) |

### 3. Relationship Tracking

Each TOC entry shows:
- Document ID and title
- Summary (why read this)
- Section headings with IDs
- Section summaries

## Reading TOC

```bash
# Read full TOC
cat .c3/TOC.md

# Quick count of documents
grep -c "^### \|^#### " .c3/TOC.md

# Find all document IDs (CTX, ADR, and C3 patterns)
grep -oP "(?<=\[)(CTX-[a-z][a-z0-9-]*|ADR-[0-9]{3}|C3-[0-9]+)[^\]]*(?=\])" .c3/TOC.md | sort -u
```

## Rebuilding TOC

When documents change, rebuild TOC:

```bash
.c3/scripts/build-toc.sh
```

> **Script missing?** Copy from the c3-skill plugin's `.c3/scripts/build-toc.sh` to your project.
> See `c3-adopt` skill for setup instructions.

### What the Script Does

1. **Scans directories** for CTX/C3-*/ADR files
2. **Extracts frontmatter** (id, title, summary, status)
3. **Lists headings** with their IDs and summaries
4. **Organizes by level** with proper nesting
5. **Generates stats** (document counts)

### When to Rebuild

Rebuild TOC after:
- Creating new documents
- Renaming documents
- Changing document titles or summaries
- Adding or renaming headings

### Manual Inspection Before Rebuild

Before rebuilding, inspect what changed:

```bash
# List all C3 documents
find .c3 -name "*.md" ! -name "TOC.md" | sort

# Compare with TOC entries
diff <(find .c3 -name "*.md" ! -name "TOC.md" -exec basename {} .md \; | sort) \
     <(grep -oP "(?<=\[)(CTX-[a-z][a-z0-9-]*|ADR-[0-9]{3}|C3-[0-9]+)[^\]]*(?=\])" .c3/TOC.md | sort -u)
```

## Build Script Reference

The build script (`.c3/scripts/build-toc.sh`) uses these key functions:

### Extract Frontmatter Field

```bash
extract_frontmatter() {
    local file=$1
    local field=$2
    awk -v field="$field" '
        /^---$/ { in_fm = !in_fm; next }
        in_fm && $0 ~ "^" field ":" {
            sub("^" field ": *", "")
            print
            exit
        }
    ' "$file"
}
```

### Extract Multi-line Summary

```bash
extract_summary() {
    local file=$1
    awk '
        /^---$/ { in_fm = !in_fm; next }
        in_fm && /^summary:/ {
            in_summary = 1
            sub(/^summary: *>? */, "")
            if (length($0) > 0) print
            next
        }
        in_fm && in_summary {
            if (/^[a-zA-Z-]+:/) { exit }
            if (length($0) > 0) print
        }
    ' "$file" | sed 's/^  //'
}
```

### List Headings with IDs

```bash
list_headings() {
    local file=$1
    awk '
        /^## .* \{#[a-z0-9-]+\}$/ {
            id_part = $0
            gsub(/.*\{#/, "", id_part)
            gsub(/\}.*/, "", id_part)
            heading_id = id_part

            title_part = $0
            gsub(/ \{#[^}]+\}$/, "", title_part)
            gsub(/^## /, "", title_part)

            print heading_id "\t" title_part
        }
    ' "$file"
}
```

### Extract Heading Summary

```bash
extract_heading_summary() {
    local file=$1
    local heading_id=$2
    awk -v hid="$heading_id" '
        $0 ~ "{#" hid "}" {
            getline
            if ($0 ~ /^<!--/) {
                in_comment = 1
                sub(/^<!-- */, "")
                if ($0 !~ /-->$/) {
                    print
                    next
                }
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
    ' "$file"
}
```

## Integration with c3-design

TOC is used in Phase 1 (Surface Understanding):

```
1. Read TOC for current state
2. Parse user request
3. Form initial hypothesis
```

The hypothesis is formed by matching request concepts to TOC entries, not by asking the user where to look.

## Common Operations

### Add New Document

1. Create document in appropriate directory
2. Ensure frontmatter has: id, title, summary
3. Add heading IDs to all `##` sections
4. Run `.c3/scripts/build-toc.sh`

### Verify TOC Accuracy

```bash
# Check for missing documents
find .c3 -name "*.md" ! -name "TOC.md" | while read f; do
    id=$(basename "$f" .md)
    if ! grep -q "$id" .c3/TOC.md; then
        echo "Missing from TOC: $id"
    fi
done

# Check for stale entries
grep -oP "(?<=\[)(CTX-[a-z][a-z0-9-]*|ADR-[0-9]{3}|C3-[0-9]+)[^\]]*(?=\])" .c3/TOC.md | while read id; do
    if ! find .c3 -name "${id}*.md" | grep -q .; then
        echo "Stale TOC entry: $id"
    fi
done
```

### Quick Stats

```bash
# From TOC footer
tail -3 .c3/TOC.md

# Or calculate fresh
echo "Contexts: $(find .c3 -maxdepth 1 -name 'CTX-*.md' | wc -l)"
echo "Containers: $(find .c3/containers -name 'C3-[0-9]-*.md' 2>/dev/null | wc -l)"
echo "Components: $(find .c3/components -name 'C3-[0-9][0-9][0-9]-*.md' 2>/dev/null | wc -l)"
echo "ADRs: $(find .c3/adr -name 'ADR-*.md' 2>/dev/null | wc -l)"
```

## Output for c3-design

When called during exploration, c3-toc provides:
- Current document inventory
- Document relationships (tree structure)
- Entry points for hypothesis validation
- Confirmation that TOC is up-to-date
