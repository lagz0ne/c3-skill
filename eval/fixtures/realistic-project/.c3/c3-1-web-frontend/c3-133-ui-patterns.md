---
id: c3-133
c3-version: 3
title: UI Patterns
type: component
category: documentation
parent: c3-1
summary: Catalog of UI design patterns with implementation references
---

# UI Patterns

Documents reusable UI patterns discovered across the application. Each pattern has a unique ID (PTN-*), implementation reference, and usage locations.

## Uses

| Category | Component | For |
|----------|-----------|-----|
| Ref | ref-design-system | Styling primitives |
| Ref | ref-visual-specs | Typography, spacing, wireframes |
| Documentation | c3-131 Information Architecture | Region definitions |
| Documentation | c3-132 User Flows | Flow interactions |

## Conventions

| Rule | Why |
|------|-----|
| Prefix PTN- | Pattern identifier |
| Document structure | Visual ASCII diagram |
| Document API | Component props example |
| Document variants | Size/state variations |

## Pattern Catalog

| ID | Pattern | Implementation | Usage |
|----|---------|----------------|-------|
| PTN-01 | Master-Detail Layout | `MasterDetailLayout.tsx` | SCR-INV, SCR-PR, SCR-APR, SCR-PAY |
| PTN-02 | Collapsible Filter Panel | `PRFilterPanel.tsx` | SCR-INV, SCR-PR, SCR-APR |
| PTN-03 | Drawer (Side Sheet) | `Drawer.tsx` | All DLG-* dialogs |
| PTN-04 | Bulk Selection Mode | Screen state | SCR-INV, SCR-APR |
| PTN-05 | Grouped List View | List component | SCR-INV, SCR-PR |
| PTN-06 | Status Badge System | `badge` variant | All PNL-*-HDR |
| PTN-07 | Detail Panel Sections | Composition | All PNL-* |
| PTN-08 | Responsive Navigation | `AppSidebar.tsx` | GBL-SIDEBAR |
| PTN-09 | Form Patterns | `FormField` | All DLG-* |
| PTN-10 | Keyboard Shortcuts | Event handlers | Global |

## Pattern Relationships

| Pattern | Depends On | Used By |
|---------|------------|---------|
| PTN-01 Master-Detail | PTN-08 Navigation | PTN-02, PTN-04, PTN-05, PTN-07 |
| PTN-07 Detail Sections | PTN-01 | PTN-03, PTN-06 |
| PTN-03 Drawer | PTN-07 | PTN-09 |
| PTN-10 Shortcuts | - | PTN-02, PTN-03, PTN-04 |

## Behavior

| Trigger | Result |
|---------|--------|
| Screen load | PTN-01 renders with PTN-05 list |
| Item select | PTN-07 sections populate detail panel |
| Filter click | PTN-02 panel opens |
| "B" key | PTN-04 bulk mode toggles |
| Create click | PTN-03 drawer opens with PTN-09 form |

## Testing

| Scenario | Verifies |
|----------|----------|
| Pattern consistency | Same pattern renders identically across screens |
| Responsive behavior | PTN-01 adapts to mobile correctly |
| Keyboard shortcuts | PTN-10 triggers work globally |
| State management | PTN-04 selection persists correctly |

## References

- `apps/start/src/components/MasterDetailLayout.tsx` - PTN-01
- `apps/start/src/components/Drawer.tsx` - PTN-03
- `apps/start/src/components/PRFilterPanel.tsx` - PTN-02
- `apps/start/src/components/AppSidebar.tsx` - PTN-08
- `apps/start/src/components/ui/variants.ts` - PTN-06
