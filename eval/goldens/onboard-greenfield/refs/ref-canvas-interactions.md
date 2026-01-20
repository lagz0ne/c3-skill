---
id: ref-canvas-interactions
title: Canvas Interaction Patterns
---

# Canvas Interaction Patterns

## Goal

Define consistent patterns for spatial canvas interactions including gesture handling, coordinate transformations, selection models, and viewport management for an intuitive user experience.

## Overview

Conventions for spatial canvas interactions and gesture handling.

## Coordinate Systems

| System | Use |
|--------|-----|
| Screen coordinates | Mouse/touch events |
| Canvas coordinates | Concept positions, stored |
| Viewport coordinates | What's visible |

### Transform Flow

```
Screen → (inverse viewport transform) → Canvas
Canvas → (viewport transform) → Screen
```

## Gesture Handling

### Primary Gestures

| Gesture | Action | Modifier |
|---------|--------|----------|
| Click | Select | - |
| Click + drag | Move selected | - |
| Click canvas | Deselect | - |
| Space + drag | Pan | - |
| Scroll wheel | Zoom | - |
| Shift + click | Multi-select | - |

### Touch Gestures

| Gesture | Action |
|---------|--------|
| Tap | Select |
| Long press | Context menu |
| Two-finger drag | Pan |
| Pinch | Zoom |
| Two-finger tap | Fit to view |

## Selection Model

| State | Visual | Behavior |
|-------|--------|----------|
| None | Default | Click to select |
| Single | Border highlight | Shows detail panel |
| Multi | Dashed borders | Group operations |
| All | All highlighted | Cmd+A |

## Viewport Management

### Zoom Levels

| Level | Use |
|-------|-----|
| 0.1 | Overview, many concepts |
| 0.5 | Scanning |
| 1.0 | Default, comfortable reading |
| 2.0 | Detail work |
| 4.0 | Maximum, fine positioning |

### Viewport Operations

| Operation | Behavior |
|-----------|----------|
| Fit all | Zoom to show all concepts with padding |
| Fit selection | Zoom to show selected concepts |
| Center on | Animate to center specific concept |
| Reset | Return to (0,0) at zoom 1.0 |

## Performance

| Technique | When |
|-----------|------|
| Viewport culling | Always - only render visible |
| Level of detail | Zoom < 0.5 - simplify nodes |
| Debounce updates | Position changes during drag |
| Batch renders | Multiple concept updates |

## Cited By

- c3-101 (Canvas Engine)
- c3-111 (Canvas Screen)
