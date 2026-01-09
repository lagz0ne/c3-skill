---
id: ref-design-system
title: Design System
---

# Design System

## Goal

Establish the visual design patterns using DaisyUI v5 with custom lofi theme, tailwind-variants for composable styling, and consistent conventions for colors, typography, spacing, and interactive states.

## Conventions

| Rule | Why |
|------|-----|
| Use DaisyUI component classes as base | Consistent, accessible primitives |
| Define variants with `tailwind-variants` tv() | Type-safe, composable styling |
| Use lofi/lofi-dark themes only | Consistent brand appearance |
| Apply status colors via semantic names | success, warning, error, info, accent |
| Use Inter Variable for text, JetBrains Mono for numbers | Readable UI with aligned tabular data |
| Button sizes: xs/sm for compact, md default, lg prominent | Consistent sizing scale |
| Status indicators: 3px left border on list items | Visual workflow state at a glance |

## Theme Tokens

| Token | Light (lofi) | Dark (lofi-dark) |
|-------|--------------|------------------|
| base-100 | Pure white | Charcoal |
| primary | Dark blue | White |
| accent | Gold/Yellow | Gold/Yellow |
| success | Green | Green |
| warning | Orange | Orange |
| error | Red | Red |
| info | Blue | Blue |

## Component Variants

| Component | Variants | Usage |
|-----------|----------|-------|
| modal() | size (sm/md/lg/xl/full), actionsAlign | Dialog overlays |
| drawer() | side (left/right), width, actionsSticky | Slide-in panels |
| alert() | type (success/error/warning/info) | Inline notifications |
| button() | variant, size, loading | All buttons |
| input(), textarea(), select() | hasError | Form fields |
| listItem() | selected, status | List rows with state |

## Testing

| Convention | How to Test |
|------------|-------------|
| Theme switching | Toggle theme, verify all tokens update |
| Variant type safety | Use wrong variant name, verify compile error |
| Status colors | Render each status, verify correct color applied |
| Responsive behavior | Resize viewport, verify layout adapts |
| Dark mode contrast | Enable lofi-dark, verify WCAG AA contrast |

## References

- `apps/start/src/styles.css` - Theme tokens
- `apps/start/src/components/ui/` - Variant definitions
- [ref-visual-specs](./ref-visual-specs.md) - Typography scale, spacing tokens, wireframes
