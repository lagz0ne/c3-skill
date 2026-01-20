---
id: ref-design-system
c3-version: 3
title: Design System
---

# Design System

## Goal

Establish visual design conventions for consistent user interface across the application, including colors, typography, spacing, and component styling.

## Color Palette

### Primary Colors

| Token | Value | Usage |
|-------|-------|-------|
| primary-500 | #3B82F6 | Primary actions, links |
| primary-600 | #2563EB | Hover states |
| primary-700 | #1D4ED8 | Active/pressed |

### Semantic Colors

| Token | Value | Usage |
|-------|-------|-------|
| success | #22C55E | Positive feedback |
| warning | #F59E0B | Caution states |
| error | #EF4444 | Errors, destructive |
| info | #3B82F6 | Informational |

### Neutral Colors

| Token | Value | Usage |
|-------|-------|-------|
| gray-50 | #F9FAFB | Background, light |
| gray-100 | #F3F4F6 | Card background |
| gray-200 | #E5E7EB | Borders |
| gray-500 | #6B7280 | Secondary text |
| gray-900 | #111827 | Primary text |

### Canvas Colors

| Token | Value | Usage |
|-------|-------|-------|
| canvas-bg | #FAFBFC | Canvas background |
| canvas-grid | #E5E7EB | Grid lines |
| selection | #3B82F6 | Selection highlight |
| cursor-self | #3B82F6 | Own cursor |
| cursor-other | Dynamic | Collaborator cursors |

## Typography

### Font Stack

| Category | Font |
|----------|------|
| Sans | Inter, system-ui, sans-serif |
| Mono | JetBrains Mono, monospace |

### Scale

| Token | Size | Line Height | Usage |
|-------|------|-------------|-------|
| text-xs | 12px | 16px | Labels, captions |
| text-sm | 14px | 20px | Secondary text |
| text-base | 16px | 24px | Body text |
| text-lg | 18px | 28px | Subheadings |
| text-xl | 20px | 28px | Headings |
| text-2xl | 24px | 32px | Page titles |

### Weights

| Token | Weight | Usage |
|-------|--------|-------|
| normal | 400 | Body text |
| medium | 500 | Emphasis |
| semibold | 600 | Headings |
| bold | 700 | Strong emphasis |

## Spacing

### Scale

| Token | Value | Usage |
|-------|-------|-------|
| space-1 | 4px | Tight spacing |
| space-2 | 8px | Related elements |
| space-3 | 12px | Form gaps |
| space-4 | 16px | Card padding |
| space-6 | 24px | Section gaps |
| space-8 | 32px | Large sections |

### Layout

| Context | Spacing |
|---------|---------|
| Component internal | space-2 to space-4 |
| Between components | space-4 to space-6 |
| Page sections | space-6 to space-8 |

## Component Patterns

### Concept Node

| Property | Value |
|----------|-------|
| Border radius | 8px |
| Shadow (default) | sm |
| Shadow (hover) | md |
| Shadow (selected) | ring-2 primary |
| Min width | 120px |
| Max width | 320px |

### Buttons

| Variant | Style |
|---------|-------|
| Primary | Filled, primary-500 |
| Secondary | Outlined, gray-200 |
| Ghost | No border, hover bg |
| Danger | Filled, error |

### Inputs

| State | Style |
|-------|-------|
| Default | gray-200 border |
| Focus | primary-500 ring |
| Error | error ring |
| Disabled | gray-100 bg |

## Animation

### Duration

| Token | Value | Usage |
|-------|-------|-------|
| fast | 100ms | Micro-interactions |
| normal | 200ms | Standard transitions |
| slow | 300ms | Larger movements |

### Easing

| Token | Value | Usage |
|-------|-------|-------|
| ease-out | cubic-bezier(0, 0, 0.2, 1) | Enter |
| ease-in | cubic-bezier(0.4, 0, 1, 1) | Exit |
| ease-in-out | cubic-bezier(0.4, 0, 0.2, 1) | Move |

## Cited By

- c3-102 (Concept Node)
