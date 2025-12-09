# Migration: 20251209-adr-status-separation

## Changes

### 1. ADR TOC Filtering
- Only ADRs with `status: implemented` appear in Table of Contents
- `proposed` and `accepted` ADRs are stored but not shown in TOC
- Quick Reference stats show implemented ADR count

### 2. Script Location
- `.c3/scripts/` directory is deprecated
- All projects should use the plugin's `scripts/build-toc.sh`
- Remove local script copies to ensure consistency

## Transforms

### Remove .c3/scripts/ directory

```bash
# Check if .c3/scripts/ exists
if [ -d ".c3/scripts" ]; then
    echo "Removing deprecated .c3/scripts/ directory..."
    rm -rf .c3/scripts/
    echo "Done. TOC script now lives in the c3-skill plugin."
fi
```

### Verify ADR status fields

```bash
# List ADRs and their status
find .c3/adr -name "*.md" 2>/dev/null | while read f; do
    status=$(awk '/^status:/ {print $2; exit}' "$f")
    echo "$(basename "$f"): status=$status"
done
```

No frontmatter changes required - existing `status` field values are compatible.

## Verification

After migration:

1. **No local scripts:**
   ```bash
   [ ! -d ".c3/scripts" ] && echo "OK: .c3/scripts/ removed"
   ```

2. **TOC shows only implemented ADRs:**
   ```bash
   # Rebuild TOC and check
   # (run plugin's build-toc.sh)
   grep -c "## Architecture Decisions" .c3/TOC.md
   # Should be 0 if no implemented ADRs, 1 if any exist
   ```

3. **Implemented ADRs appear:**
   ```bash
   # Count implemented ADRs
   grep -l "^status: implemented" .c3/adr/*.md 2>/dev/null | wc -l
   # This count should match ADRs shown in TOC
   ```

## Notes

- This is a behavioral change, not a structural change
- Existing ADRs continue to work
- ADRs with `status: proposed` or `status: accepted` are still accessible via `.c3/adr/` but won't clutter the TOC
- To make an ADR visible in TOC, use `c3-audit` to verify and transition status
