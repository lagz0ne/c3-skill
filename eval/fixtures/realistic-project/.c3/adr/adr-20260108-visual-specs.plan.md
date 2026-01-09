# Plan: Add Visual Specs Component

**ADR:** adr-20260108-visual-specs
**Status:** Complete

## Pre-Execution Checklist

- [x] ADR accepted
- [ ] Extract typography values from styles.css
- [ ] Extract spacing values from existing components
- [ ] Document PTN-01 and PTN-03 wireframes

## Steps

### Step 1: Create c3-115 Visual Specs

Create `.c3/c3-1-web-frontend/c3-115-visual-specs.md` with:
- Frontmatter (id, type: component, category: auxiliary, parent: c3-1)
- Typography Scale table
- Spacing Tokens table
- PTN-01 Master-Detail wireframe
- PTN-03 Drawer wireframe
- Status Badge color mapping
- Component gallery references

### Step 2: Update c3-1 README

- Add c3-115 to Auxiliary table
- Update mermaid diagram to include c3-115 with linkages

### Step 3: Update c3-114 Design System

- Add c3-115 to References section

### Step 4: Update c3-133 UI Patterns

- Add c3-115 to Uses table

## Verification

```bash
# Check all files exist and have proper format
ls -la .c3/c3-1-web-frontend/c3-115-visual-specs.md
grep "c3-115" .c3/c3-1-web-frontend/README.md
grep "c3-115" .c3/c3-1-web-frontend/c3-114-design-system.md
grep "c3-115" .c3/c3-1-web-frontend/c3-133-ui-patterns.md
```
