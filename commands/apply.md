---
description: Generate CLAUDE.md context files from C3 component Code References
argument-hint: [--dry-run]
---

# C3 Apply

Propagates C3 architecture context to code directories by generating CLAUDE.md files based on component Code References.

## Prerequisites

- Project must have `.c3/` directory with component documentation
- Components must have `## Code References` sections listing directories

## Behavior

1. **Scan components** - Read all `.c3/c3-*/c3-*.md` files
2. **Extract mappings** - Parse `## Code References` and `## Related Refs` sections
3. **Generate CLAUDE.md** - For each referenced directory:
   - Create CLAUDE.md if missing
   - Replace `<!-- c3-generated -->` block if exists
   - Preserve user content outside the block
4. **Report changes** - List files created/updated (does NOT commit)

## Dry Run

Use `--dry-run` to preview changes without writing files.

## Implementation

```bash
# Step 1: Find all components with Code References
for comp_file in .c3/c3-*/c3-*.md; do
  # Extract component ID and title from frontmatter
  comp_id=$(grep "^id:" "$comp_file" | head -1 | sed 's/id: *//')
  comp_title=$(grep "^title:" "$comp_file" | head -1 | sed 's/title: *//')

  # Extract Code References section (directories)
  # Look for paths like src/routes/auth/, lib/utils/

  # Extract Related Refs section
  # Look for ref-* patterns

  # For each directory, generate/update CLAUDE.md
done
```

## CLAUDE.md Template

```markdown
<!-- c3-generated: {component-id} -->
# {component-id}: {component-title}

Before modifying this code, read:
- Component: `.c3/{container}/{component-id}.md`
- Patterns: `{ref-1}`, `{ref-2}`

Full refs: `.c3/refs/ref-{name}.md`
<!-- end-c3-generated -->
```

## Block Replacement Logic

```
If CLAUDE.md exists in target directory:
  If contains <!-- c3-generated: -->:
    Replace content between markers (preserve rest of file)
  Else:
    Append c3-generated block at end of file
Else:
  Create new CLAUDE.md with c3-generated block only
```

## Example Output

```
/c3 apply

Scanning components...
Found 5 components with Code References

Generating CLAUDE.md files:
  [CREATE] src/routes/auth/CLAUDE.md (c3-201)
  [UPDATE] src/middleware/CLAUDE.md (c3-102)
  [SKIP]   src/utils/CLAUDE.md (already current)

Summary: 1 created, 1 updated, 1 skipped
Files are unstaged - review and commit when ready.
```

## Agent Instructions

When user runs `/c3 apply`:

1. **Check prerequisites**
   - Verify `.c3/` exists
   - If not, suggest `/c3` to onboard first

2. **Build component map**
   ```
   For each .c3/c3-*/c3-*.md:
     - Read file
     - Extract: id, title, container path
     - Parse ## Code References for directory paths
     - Parse ## Related Refs for ref-* patterns
     - Store: {directory → (component, refs[])}
   ```

3. **Process each directory**
   ```
   For each directory in map:
     - Check if CLAUDE.md exists
     - Generate expected block content
     - Compare with existing (if any)
     - Decide: CREATE, UPDATE, or SKIP
   ```

4. **Write files** (unless --dry-run)
   - Use Edit tool for updates (preserve user content)
   - Use Write tool for creates

5. **Report results**
   - List all changes
   - Remind user files are unstaged

## Related

- `/c3 audit` - Phase 10 checks CLAUDE.md presence/freshness
- Run audit first to see what's missing, then apply to fix
