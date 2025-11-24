# Migration: 20251124-toc-fix-links

> From: `20251124-skill-zips`
> To: `20251124-toc-fix-links`

## Changes

- TOC section links now include full relative file paths (fixes dead anchor links)
- TOC build script moved to plugin `scripts/` directory (not per-project `.c3/scripts/`)
- Section links in TOC.md change from `#heading-id` to `./path/to/file.md#heading-id`
- c3-toc skill updated to reference plugin's script location

## Transforms

### Regenerate TOC

**Action:** Run the updated build-toc.sh script to regenerate TOC with correct links.

```bash
# The script is now in the plugin's scripts/ directory
# Run from project root
bash <path-to-plugin>/scripts/build-toc.sh
```

### Update c3-version

**File:** `.c3/README.md` (or context README)
**Pattern:** `^c3-version: 20251124-.*$`
**Replace:** `c3-version: 20251124-toc-fix-links`

## Verification

```bash
# TOC exists and was regenerated
test -f .c3/TOC.md

# Section links include file paths (not just anchors)
grep -q '](./.*\.md#' .c3/TOC.md && echo "Links include file paths"

# No orphan anchor-only section links
! grep -E '\- \[[^\]]+\]\(#[a-z0-9-]+\)' .c3/TOC.md && echo "No orphan anchors"

# c3-version updated
grep -q '^c3-version: 20251124-toc-fix-links$' .c3/README.md
```
