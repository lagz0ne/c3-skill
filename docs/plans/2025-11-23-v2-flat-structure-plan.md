# V2 Flat Structure Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate C3 documentation from deep nested folders to flat structure (version 2).

**Architecture:** Remove nested subfolders within `components/`, use `README.md` as primary context with `id: context`. All skills and scripts updated to support both scanning and generating flat structure. Migration automated via `c3-migrate`.

**Tech Stack:** Bash scripts, Markdown documentation, Claude Code skills

---

## Task 1: Update MIGRATIONS.md with v2 Section

**Files:**
- Modify: `MIGRATIONS.md`

**Step 1: Read current MIGRATIONS.md**

Review existing format and version 1 documentation.

**Step 2: Add Version 2 section**

Append after the Version 1 section:

```markdown
## Version 2 (from 1)

### Changes
- Flatten components directory (remove container subfolders)
- Primary context becomes README.md with `id: context`
- Path: `components/{container}/C3-*.md` → `components/C3-*.md`
- Path: `CTX-system-overview.md` → `README.md`

### Transforms

**Move nested components to flat:**
- **Files:** `components/*/*.md`
- **Action:** Move to `components/` (parent directory)
- **Command:** `mv .c3/components/*/*.md .c3/components/ && rmdir .c3/components/*/`

**Rename primary context:**
- **Pattern:** `CTX-system-overview.md`
- **Replace:** `README.md`
- **Frontmatter:** Change `id: CTX-system-overview` → `id: context`

**Update internal links:**
- **Pattern:** `](./components/[^/]+/(C3-[0-9])`
- **Replace:** `](./components/$1`
- **Files:** `.c3/**/*.md`

**Update context links:**
- **Pattern:** `CTX-system-overview.md`
- **Replace:** `README.md`
- **Files:** `.c3/**/*.md`

### Verification
- `.c3/VERSION` contains `2`
- No subdirectories in `.c3/components/`
- `.c3/README.md` exists with `id: context`
- All component files match `.c3/components/C3-[0-9][0-9][0-9]-*.md`
- No broken internal links (grep for old paths returns empty)
```

**Step 3: Commit**

```bash
git add MIGRATIONS.md
git commit -m "docs: add v2 migration section for flat structure"
```

---

## Task 2: Update build-toc.sh for v2 Structure

**Files:**
- Modify: `.c3/scripts/build-toc.sh`

**Step 1: Update context scanning to include README.md**

Change the context section (around line 105-137) to:

```bash
# Context Level - check README.md first, then CTX-*.md
readme_file="$C3_ROOT/README.md"
ctx_files=$(find "$C3_ROOT" -maxdepth 1 -name "CTX-*.md" 2>/dev/null | sort)
has_readme=0

if [ -f "$readme_file" ]; then
    has_readme=1
fi

if [ "$has_readme" -eq 1 ] || [ -n "$ctx_files" ]; then
    echo "## Context Level" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"

    # Process README.md as primary context
    if [ "$has_readme" -eq 1 ]; then
        id=$(extract_frontmatter "$readme_file" "id")
        title=$(extract_frontmatter "$readme_file" "title")
        summary=$(extract_summary "$readme_file")

        echo "### [$id](./README.md) - $title" >> "$TEMP_FILE"
        echo "> $summary" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        echo "**Sections**:" >> "$TEMP_FILE"
        while IFS=$'\t' read -r heading_id heading_title; do
            heading_summary=$(extract_heading_summary "$readme_file" "$heading_id")
            if [ -n "$heading_summary" ]; then
                echo "- [$heading_title](#$heading_id) - $heading_summary" >> "$TEMP_FILE"
            else
                echo "- [$heading_title](#$heading_id)" >> "$TEMP_FILE"
            fi
        done < <(list_headings "$readme_file")
        echo "" >> "$TEMP_FILE"
        echo "---" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"
    fi

    # Process auxiliary CTX-*.md files
    for file in $ctx_files; do
        # ... existing CTX processing code ...
    done
fi
```

**Step 2: Update component scanning for flat structure**

Change the component section (around line 169-202) to support BOTH v1 (nested) and v2 (flat):

```bash
# Component Level - support both v1 (nested) and v2 (flat)
# Check for v2 flat structure first
flat_com_count=$(find "$C3_ROOT/components" -maxdepth 1 -name "C3-[0-9][0-9][0-9]-*.md" 2>/dev/null | wc -l || echo 0)

if [ "$flat_com_count" -gt 0 ]; then
    # V2: Flat structure
    echo "## Component Level" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"

    # Group by container number (first digit after C3-)
    for container_num in $(find "$C3_ROOT/components" -maxdepth 1 -name "C3-[0-9][0-9][0-9]-*.md" -exec basename {} \; | sed 's/C3-\([0-9]\).*/\1/' | sort -u); do
        echo "### Container $container_num Components" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        for file in $(find "$C3_ROOT/components" -maxdepth 1 -name "C3-${container_num}[0-9][0-9]-*.md" | sort); do
            id=$(extract_frontmatter "$file" "id")
            title=$(extract_frontmatter "$file" "title")
            summary=$(extract_summary "$file")

            echo "#### [$id](./components/${id}.md) - $title" >> "$TEMP_FILE"
            echo "> $summary" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"

            echo "**Sections**:" >> "$TEMP_FILE"
            while IFS=$'\t' read -r heading_id heading_title; do
                heading_summary=$(extract_heading_summary "$file" "$heading_id")
                if [ -n "$heading_summary" ]; then
                    echo "- [$heading_title](#$heading_id) - $heading_summary" >> "$TEMP_FILE"
                else
                    echo "- [$heading_title](#$heading_id)" >> "$TEMP_FILE"
                fi
            done < <(list_headings "$file")
            echo "" >> "$TEMP_FILE"
            echo "---" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"
        done
    done
elif [ "$com_count" -gt 0 ]; then
    # V1: Nested structure (existing code)
    # ... keep existing nested component code as fallback ...
fi
```

**Step 3: Update document counts**

Update the count section (around line 105-108):

```bash
# Count documents - support both README.md and CTX-*.md for context
readme_exists=0
[ -f "$C3_ROOT/README.md" ] && readme_exists=1
ctx_count=$(find "$C3_ROOT" -maxdepth 1 -name "CTX-*.md" 2>/dev/null | wc -l || echo 0)
ctx_total=$((readme_exists + ctx_count))

con_count=$(find "$C3_ROOT/containers" -name "C3-[0-9]-*.md" 2>/dev/null | wc -l || echo 0)

# Component count - check flat first, then nested
flat_com_count=$(find "$C3_ROOT/components" -maxdepth 1 -name "C3-[0-9][0-9][0-9]-*.md" 2>/dev/null | wc -l || echo 0)
nested_com_count=$(find "$C3_ROOT/components" -mindepth 2 -name "C3-[0-9][0-9][0-9]-*.md" 2>/dev/null | wc -l || echo 0)
com_count=$((flat_com_count + nested_com_count))

adr_count=$(find "$C3_ROOT/adr" -name "ADR-*.md" 2>/dev/null | wc -l || echo 0)
```

**Step 4: Commit**

```bash
git add .c3/scripts/build-toc.sh
git commit -m "feat: update build-toc.sh for v2 flat structure"
```

---

## Task 3: Update c3-adopt Skill for v2 Structure

**Files:**
- Modify: `skills/c3-adopt/SKILL.md`

**Step 1: Update scaffolding section (Phase 1)**

Change the directory creation (around line 66-67) from:
```bash
mkdir -p .c3/{containers,components,adr,scripts}
```

To v2 structure with explanation:
```markdown
### Create Scaffolding

If proceeding with new documentation:

```bash
# V2 flat structure - no nested component subfolders
mkdir -p .c3/{containers,components,adr,scripts}
```

> **Note:** V2 uses flat `components/` directory. Component files are named `C3-{container}{NN}-slug.md` where the container number is encoded in the filename, eliminating the need for nested folders.
```

**Step 2: Update index.md template**

Change the index.md template (around line 86-98) to reference README.md:

```markdown
Create `index.md`:
```markdown
---
layout: home
title: C3 Architecture Documentation
---

# C3 Architecture Documentation

- [Table of Contents](./TOC.md)
- [System Overview](./README.md)
```
```

**Step 3: Update Phase 2 context document output**

Change the delegation output (around line 170-176) to create README.md:

```markdown
### Delegate to c3-context-design

Once you have understanding:
> "I now understand your system context. I'll use the c3-context-design skill to create README.md (primary context)."

Use `c3-context-design` to create `README.md` with `id: context`:
- System overview
- Architecture diagram (from identified containers)
- Container list (high-level)
- Protocols section
- Cross-cutting concerns
- Deployment section
```

**Step 4: Update verification checklist**

Change the verification checklist (around line 292-307) to use v2 paths:

```markdown
### Verification Checklist

Present to user:

```markdown
## C3 Adoption Complete

### Created:
- [ ] `.c3/README.md` - System context (id: context)
- [ ] `.c3/containers/C3-1-*.md` - [Container 1]
- [ ] `.c3/containers/C3-2-*.md` - [Container 2]
- [ ] `.c3/components/C3-*.md` - [N] components (flat structure)
- [ ] `.c3/TOC.md` - Table of contents
- [ ] `.c3/VERSION` - Version file (contains "2")
- [ ] `.c3/scripts/build-toc.sh` - TOC generator
```
```

**Step 5: Update appendix build-toc.sh reference**

The appendix (line 374-445) contains embedded build-toc.sh code. Update it to match Task 2 changes or add note:

```markdown
> **Note:** The embedded script below is for reference. Always copy the current version from the plugin's `.c3/scripts/build-toc.sh` which supports both v1 and v2 structures.
```

**Step 6: Commit**

```bash
git add skills/c3-adopt/SKILL.md
git commit -m "feat: update c3-adopt for v2 flat structure"
```

---

## Task 4: Update c3-locate Skill for v2 Structure

**Files:**
- Modify: `skills/c3-locate/SKILL.md`

**Step 1: Update document finding patterns**

Change the implementation section (around line 69-89) to support both README.md and flat components:

```markdown
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
```

**Step 2: Update ID conventions table**

Update the ID conventions table (around line 162-172):

```markdown
## ID Conventions

| Pattern | Level | Example | File Path |
|---------|-------|---------|-----------|
| `context` | Primary Context | context | `.c3/README.md` |
| `CTX-slug` | Auxiliary Context | CTX-actors | `.c3/CTX-actors.md` |
| `C3-<C>-slug` | Container | C3-1-backend | `.c3/containers/C3-1-backend.md` |
| `C3-<C><NN>-slug` | Component | C3-102-auth | `.c3/components/C3-102-auth.md` |
| `ADR-###-slug` | Decision | ADR-003-cache | `.c3/adr/ADR-003-cache.md` |
```

**Step 3: Update usage examples**

Update examples to include `context` ID:

```markdown
### By Document ID

Retrieve document frontmatter and overview:

```
c3-locate context              # Primary context (README.md)
c3-locate CTX-actors           # Auxiliary context
c3-locate C3-1-backend
c3-locate C3-102-auth-middleware
c3-locate ADR-003-caching-strategy
```
```

**Step 4: Commit**

```bash
git add skills/c3-locate/SKILL.md
git commit -m "feat: update c3-locate for v2 flat structure"
```

---

## Task 5: Update c3-migrate Skill with v1→v2 Logic

**Files:**
- Modify: `skills/c3-migrate/SKILL.md`

**Step 1: Add v1→v2 specific transforms section**

Add after the "Common Patterns" section (around line 188):

```markdown
## V1 → V2 Migration Details

### Detecting V1 Structure

```bash
# V1 has nested component directories
nested_dirs=$(find .c3/components -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
if [ "$nested_dirs" -gt 0 ]; then
    echo "V1 structure detected: nested component directories"
fi
```

### Transform: Flatten Components

```bash
# Move all component files up one level
for container_dir in .c3/components/*/; do
    [ -d "$container_dir" ] || continue
    mv "$container_dir"*.md .c3/components/ 2>/dev/null || true
    rmdir "$container_dir" 2>/dev/null || true
done
```

### Transform: Rename Context

```bash
# Rename CTX-system-overview.md to README.md
if [ -f ".c3/CTX-system-overview.md" ]; then
    mv ".c3/CTX-system-overview.md" ".c3/README.md"
    # Update frontmatter id
    sed -i 's/^id: CTX-system-overview$/id: context/' ".c3/README.md"
fi
```

### Transform: Update Internal Links

```bash
# Update component links (remove container subfolder from path)
find .c3 -name "*.md" -exec sed -i \
    's|\](./components/[^/]*/\(C3-[0-9]\)|\](./components/\1|g' {} \;

# Update context links
find .c3 -name "*.md" -exec sed -i \
    's|CTX-system-overview\.md|README.md|g' {} \;
```

### V1→V2 Verification

After migration, verify:

```bash
# No nested component directories
[ $(find .c3/components -mindepth 1 -maxdepth 1 -type d | wc -l) -eq 0 ] || echo "FAIL: nested dirs remain"

# README.md exists with correct id
grep -q '^id: context$' .c3/README.md || echo "FAIL: README.md id not updated"

# VERSION updated
[ "$(cat .c3/VERSION)" = "2" ] || echo "FAIL: VERSION not updated"

# No broken component links
! grep -r '](./components/[^/]*/C3-' .c3/*.md .c3/**/*.md || echo "FAIL: old component links remain"
```
```

**Step 2: Commit**

```bash
git add skills/c3-migrate/SKILL.md
git commit -m "feat: add v1→v2 migration details to c3-migrate"
```

---

## Task 6: Update Plugin README.md

**Files:**
- Modify: `README.md`

**Step 1: Update documentation structure section**

Change the structure diagram (around line 98-112) to v2:

```markdown
## Documentation Structure

```
.c3/
├── README.md                     # Primary context (id: context)
├── CTX-actors.md                 # Auxiliary context (optional)
├── index.md                      # Navigation index
├── TOC.md                        # Auto-generated TOC
├── VERSION                       # Version number (e.g., "2")
├── containers/
│   └── C3-1-backend.md           # Container documents
├── components/
│   ├── C3-101-db-pool.md         # Component documents (flat!)
│   └── C3-102-auth-service.md    # No nested subfolders
├── adr/
│   └── ADR-001-rest-api.md       # Architecture decisions
└── scripts/
    └── build-toc.sh              # TOC generator
```

> **V2 Change:** Components are now flat in `components/` directory. The container number is encoded in the component ID (e.g., `C3-101` = container 1, component 01).
```

**Step 2: Update document conventions section**

Update the Unique IDs section (around line 118-124):

```markdown
### Unique IDs

Every document has a unique ID:

- **context**: Primary context (`.c3/README.md`)
- **CTX-slug**: Auxiliary context (e.g., `CTX-actors`)
- **C3-<C>-slug**: Container level (e.g., `C3-1-backend`; single digit container number)
- **C3-<C><NN>-slug**: Component level (e.g., `C3-101-db-pool`; container digit + 2-digit component number)
- **ADR-NNN-slug**: Architecture decisions (e.g., `ADR-001-rest-api`)
```

**Step 3: Update simplified frontmatter example**

Update the frontmatter example (around line 129-138) to show README.md:

```markdown
### Simplified Frontmatter

**For README.md (primary context):**
```yaml
---
id: context
title: System Overview
summary: >
  Bird's-eye view of the system, actors, and key interactions.
  Read this first to understand the overall architecture.
---
```

**For components:**
```yaml
---
id: C3-101-db-pool
title: Database Connection Pool Component
summary: >
  Explains PostgreSQL connection pooling strategy, configuration, and
  retry behavior. Read this to understand how the backend manages database
  connections efficiently and handles connection failures.
---
```
```

**Step 4: Commit**

```bash
git add README.md
git commit -m "docs: update README for v2 flat structure"
```

---

## Task 7: Bump VERSION to 2

**Files:**
- Modify: `VERSION`

**Step 1: Update VERSION file**

```bash
echo "2" > VERSION
```

**Step 2: Commit**

```bash
git add VERSION
git commit -m "chore: bump VERSION to 2"
```

---

## Task 8: Final Verification

**Step 1: Run verification checks**

```bash
# Check all files modified
git diff --stat origin/main..HEAD

# Verify VERSION is 2
cat VERSION

# Verify MIGRATIONS.md has v2 section
grep -q "Version 2" MIGRATIONS.md && echo "PASS: v2 in MIGRATIONS" || echo "FAIL"

# Verify build-toc.sh handles README.md
grep -q "README.md" .c3/scripts/build-toc.sh && echo "PASS: build-toc supports README" || echo "FAIL"

# Verify c3-adopt references README.md
grep -q "README.md" skills/c3-adopt/SKILL.md && echo "PASS: c3-adopt updated" || echo "FAIL"

# Verify c3-locate has context ID
grep -q 'id.*context' skills/c3-locate/SKILL.md && echo "PASS: c3-locate updated" || echo "FAIL"

# Verify c3-migrate has v1→v2
grep -q "V1.*V2" skills/c3-migrate/SKILL.md && echo "PASS: c3-migrate updated" || echo "FAIL"
```

**Step 2: Summary commit (if needed)**

If any fixes were needed, create a final cleanup commit.

---

## Summary

| Task | Files | Purpose |
|------|-------|---------|
| 1 | MIGRATIONS.md | Add v2 migration transforms |
| 2 | .c3/scripts/build-toc.sh | Support flat components + README.md |
| 3 | skills/c3-adopt/SKILL.md | Generate v2 structure |
| 4 | skills/c3-locate/SKILL.md | Find docs in v2 layout |
| 5 | skills/c3-migrate/SKILL.md | v1→v2 migration logic |
| 6 | README.md | Document v2 structure |
| 7 | VERSION | Bump to 2 |
| 8 | (verification) | Final checks |
