---
id: c3-108
c3-version: 3
c3-seal: 2baa767963c2bc1504f4f334d31dbb386a3efd3a617added25fa23395f86bb2d
title: DevTools (DevShell)
type: component
category: foundation
parent: c3-1
goal: Development toolbar providing viewport controls, theme toggle, element picker, fullscreen mode, and log viewer panel
uses:
    - ref-ui-patterns
---

# DevTools (DevShell)

## Goal

Development toolbar providing viewport controls, theme toggle, element picker, fullscreen mode, and log viewer panel

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own DevTools (DevShell) behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep DevTools (DevShell) decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for DevTools (DevShell) so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before DevTools (DevShell) behavior is changed. | ref-ui-patterns |
| Inputs | Accept only the files, commands, data, or calls that belong to DevTools (DevShell) ownership. | ref-ui-patterns |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-ui-patterns |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-ui-patterns |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks DevTools (DevShell) to deliver its documented responsibility. | ref-ui-patterns |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-ui-patterns |
| Alternate paths | When a request falls outside DevTools (DevShell) ownership, hand it to the parent or sibling component. | ref-ui-patterns |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-ui-patterns |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-ui-patterns | ref | Governs DevTools (DevShell) behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| DevTools (DevShell) input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| DevTools (DevShell) output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |

## Architecture Details

## Dependencies

- `cn()` from `lib/utils` -- class merging
- `window.__devLogs` -- log entries populated by the frontend logging system

## Activation

1. Visit `/dev/inspect` -- sets `inspector=true` cookie (24h expiry)
2. Page redirects to `/` with full reload
3. `RootDocument` detects cookie and renders `<DevShell />` instead of normal app
4. DevShell loads the app in an `<iframe src="/">`

## Features

| Feature | Description |
| --- | --- |
| Viewport Simulation | Mobile (375x667), Tablet (768x1024), Desktop (100%). Constrained viewports show a browser chrome header |
| Theme Toggle | Light, Dark, System. Synced to both parent and iframe documentElement.classList |
| Fullscreen Mode | Hides toolbar and log panel for focused testing |
| Element Picker | Click any element in the iframe to copy React component ancestry, source location, and CSS selector to clipboard |
| Log Viewer | Polls window.__devLogs every 500ms. Filters: all/debug/info/warn/error. Max 100 entries |
| Chat Panel | Lazy-loaded DevChatPanel with /pending command support for dev-time interaction |

## Keyboard Shortcuts

| Key | Action |
| --- | --- |
| F | Toggle fullscreen |
| V | Cycle viewport: mobile -> tablet -> desktop |
| D | Toggle dark/light theme |
| S | Toggle element picker |
| ESC | Exit picker or fullscreen |

## Element Picker

Injects overlay + label elements into the iframe DOM. On click, builds clipboard text with:

1. React component ancestry (walks `__reactFiber$` up the tree, skipping Suspense/Fragment/etc.)
2. Debug source file:line (from `_debugSource`)
3. CSS selector (prefers `#id` > `[data-testid]` > generated path)

## Log Entry Structure

```typescript
interface DevLogEntry {
  timestamp: number
  level: 'debug' | 'info' | 'warn' | 'error'
  name: string         // Logger name, e.g. '[api]', '[sync]'
  msg: string
  data?: Record<string, unknown>
  sessionId: string
  user?: string
}
```

## Integration

DevShell is a self-contained HTML document (renders its own `<html>`, `<head>`, `<body>` with inline styles). It does not use Tailwind or the app's CSS -- all styling is scoped CSS variables with a `.shell` / `.shell.light-mode` toggle.

URL navigation in the iframe is reflected in the parent URL via `history.replaceState`.
