---
id: ref-visual-specs
title: Visual Specs
---

# Visual Specs

## Goal

Document the visual specifications needed to generate new screens with consistent appearance: typography scale, spacing tokens, and pattern wireframes.

## Typography Scale

| Token | Size | Weight | Line Height | Usage |
|-------|------|--------|-------------|-------|
| `text-lg` | 1.125rem (18px) | 600 | 1.75 | Page titles, drawer headers |
| `text-base` | 1rem (16px) | 400 | 1.5 | Body text, descriptions |
| `text-sm` | 0.875rem (14px) | 400 | 1.25 | List items, form labels |
| `text-xs` | 0.75rem (12px) | 400 | 1 | Badges, timestamps, amounts |
| `text-[0.7rem]` | 0.7rem (11.2px) | 700 | 1 | Section headers (uppercase) |
| `text-[0.65rem]` | 0.65rem (10.4px) | 500 | 1 | Meta labels (uppercase) |

### Font Families

| Variable | Font | Usage |
|----------|------|-------|
| `--font-display` | Inter Variable | All UI text |
| `--font-mono` | JetBrains Mono | Amounts, IDs, codes |
| `--font-logo` | Orbit | Logo only |

## Spacing Scale

| Token | Value | Usage |
|-------|-------|-------|
| `space-0.5` | 2px | Tight gaps (badge padding) |
| `space-1` | 4px | Icon gaps, inline spacing |
| `space-2` | 8px | List item gaps, small padding |
| `space-3` | 12px | Section padding, list padding |
| `space-4` | 16px | Card padding, detail sections |
| `space-6` | 24px | Drawer body padding |
| `space-8` | 32px | Large section gaps |

### Component Heights

| Component | Height | Class |
|-----------|--------|-------|
| Button xs | 28px | `h-7` |
| Button sm | 32px | `h-8` |
| Button md | 36px | `h-9` |
| Button lg | 40px | `h-10` |
| Input | 36px | `h-9` |
| Footer bar | 56px | `min-h-[3.5rem]` |

## Status Badge Colors

| Status | Background | Text | Border | CSS Variable |
|--------|------------|------|--------|--------------|
| success | `bg-success/10` | `text-success` | `border-success/50` | `--success: 142 76% 36%` |
| warning | `bg-warning/10` | `text-warning` | `border-warning/50` | `--warning: 38 92% 50%` |
| error | `bg-destructive/10` | `text-destructive` | `border-destructive/50` | `--destructive: 0 84.2% 60.2%` |
| info | `bg-info/10` | `text-info` | `border-info/50` | `--info: 199 89% 48%` |
| neutral | `bg-muted` | `text-muted-foreground` | - | `--muted: 240 4.8% 95.9%` |

## Conventions

| Rule | Why |
|------|-----|
| Use typography tokens, not arbitrary sizes | Consistent hierarchy |
| Use spacing scale for padding/margins | Predictable rhythm |
| Match wireframe proportions | Visual consistency across screens |
| Apply status colors via semantic names | Theme-aware styling |

## Testing

| Scenario | Verifies |
|----------|----------|
| New screen matches typography scale | All text uses documented tokens |
| Spacing follows scale | No arbitrary values outside scale |
| Status colors applied correctly | Semantic names map to theme tokens |
| Wireframe proportions match | List width, panel flex ratios correct |

## References

- `apps/start/src/styles.css` - Token definitions
- `apps/start/src/components/ui/variants.ts` - Component variants
- [ref-design-system](./ref-design-system.md) - Token source
