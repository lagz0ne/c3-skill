# VSCode C3 Document Navigation Design

**Date:** 2026-03-10
**Status:** Validated

## Goal

Extend the C3 Architecture Navigator VSCode extension to navigate any document in the `.c3/` folder — not just code-map.yaml. Two new capabilities:

1. **Markdown navigation** — Cross-reference navigation (CodeLens, Hover, Ctrl+Click) inside `.c3/**/*.md` files
2. **Tree view** — A sidebar panel showing the C3 architecture hierarchy with toggle between hierarchy and flat views

## Architecture Overview

Extend the existing modular structure. The current `DocMap` already indexes all `.c3/*.md` files with frontmatter — it's the foundation for both features.

```
Existing (unchanged):
  docMap.ts          — Already indexes IDs, titles, goals, summaries
  utils.ts           — Already has ID regex, path extraction, frontmatter parsing

New/Extended:
  codeLensProvider.ts   — Extend to handle .md files (frontmatter + tables)
  definitionProvider.ts — Extend to handle .md files (all C3 IDs + file paths)
  hoverProvider.ts      — Extend to handle .md files
  treeViewProvider.ts   — NEW: TreeDataProvider for sidebar panel
  extension.ts          — Register new providers and tree view
```

Key principle: `DocMap` is the single source of truth. Both features read from it, no new data layer needed.

## Markdown Navigation — Provider Changes

### Document selector

All three providers register for `.c3/**/*.md` in addition to existing `.c3/**/*.yaml`.

### CodeLens — Smart placement

- **Frontmatter block** (between `---` delimiters): CodeLens above `parent:`, `uses:`, and `via:` fields. Each referenced ID gets a `-> Title (filename)` lens.
- **Dependency/Code Reference tables**: CodeLens above rows containing C3 IDs or file paths. Detected by matching markdown table rows (`| ... |`) containing C3 ID patterns or backtick-wrapped file paths.
- **Body prose**: No CodeLens. Hover + Ctrl+Click only.

Principle: CodeLens for structured references, hover+click for prose references.

### Definition provider

- C3 IDs anywhere in the file -> jump to target `.c3/*.md` document
- File paths in backticks or quotes (e.g., `` `cli/internal/frontmatter/parse.go` ``) -> open the workspace file
- Works in frontmatter, tables, and body text uniformly

### Hover provider

- C3 IDs -> same rich card as YAML (id, title, goal, summary)
- File paths -> same path status display (exists/missing, file/folder)
- Works everywhere in the document

### Pattern matching additions (`utils.ts`)

- Detect frontmatter block boundaries (`---`)
- Detect markdown table rows for CodeLens placement decisions
- Extract file paths from backtick/quote wrappers in markdown context

## Tree View — Structure & Behavior

### Registration

New `TreeDataProvider` registered as a view in the Explorer sidebar. View ID: `c3Navigator`.

### Two toggle modes (toolbar button)

**Hierarchy mode (default):**

```
c3-0 - C3 Design                          [active]
+-- c3-1 - Go CLI                          [active]
|   +-- Foundation
|   |   +-- c3-101 - Frontmatter Parser    [active]
|   |   +-- c3-102 - Walker                [active]
|   |   +-- c3-103 - Templates             [active]
|   |   +-- c3-104 - Wiring               [active]
|   |   +-- c3-105 - Codemap Lib           [active]
|   +-- Feature
|       +-- c3-110 - Init Cmd              [active]
|       +-- ...
|       +-- c3-117 - Diff Cmd              [provisioned]
+-- c3-2 - Claude Skill                    [active]
|   +-- c3-201 - Skill Router              [active]
|   +-- c3-210 - Operation Refs            [active]
+-- References
|   +-- ref-frontmatter-docs
|   +-- ref-embedded-templates
|   +-- ref-cross-compiled-binary
+-- ADRs
    +-- adr-00000000 - C3 Adoption
    +-- adr-20260309 - Add Diff Cmd
```

**Flat mode:**

```
Containers (3)
Components (15)
References (3)
ADRs (2)
```

Each group lists items alphabetically by ID.

### Behavior

- Click any item -> opens the `.md` document
- Status badges: `[active]` / `[provisioned]` for components, parsed from frontmatter
- Goal shown as tooltip on hover (not inline, keeps tree clean)
- Tree auto-refreshes when `DocMap` fires `onDidRebuild` event

## Context Menu Actions

Right-click on any tree item:

| Action | Behavior |
|--------|----------|
| Show Dependencies | QuickPick list of components in `uses:` field. Select one -> opens doc. Disabled if empty. |
| Show Dependents | QuickPick list of components referencing this ID in their `uses:` or `via:` fields. Disabled if none. |
| Open in Code Map | Opens `code-map.yaml`, scrolls to line containing this component's ID. |
| Copy ID | Copies ID (e.g., `c3-101`) to clipboard. |

## DocMap Extensions

`DocEntry` interface needs additional fields from frontmatter:

| Field | Purpose |
|-------|---------|
| `type` | container / component / ref / adr — for tree grouping |
| `category` | foundation / feature — for hierarchy sub-grouping |
| `parent` | parent ID — for hierarchy tree building |
| `uses` | dependency list — for "Show Dependencies" |
| `via` | reverse dependency list — for "Show Dependents" |
| `status` | active / provisioned — for badge display |

## Files Changed

| File | Change |
|------|--------|
| `docMap.ts` | Extend `DocEntry` with `type`, `category`, `parent`, `uses`, `via`, `status` |
| `utils.ts` | Add frontmatter block detection, markdown table detection, backtick path extraction |
| `codeLensProvider.ts` | Add `.md` selector, smart placement logic (frontmatter + tables) |
| `definitionProvider.ts` | Add `.md` selector, backtick path support |
| `hoverProvider.ts` | Add `.md` selector |
| `treeViewProvider.ts` | **NEW** — TreeDataProvider with hierarchy/flat toggle, context menu commands |
| `extension.ts` | Register tree view, context menu commands, toggle command |
| `package.json` | Declare view container, view, commands, menus |

## Not Building (YAGNI)

- No search/filter bar — VSCode's built-in tree filtering is sufficient
- No drag-and-drop or editing from tree view — read-only navigation
- No dependency graph visualization — QuickPick lists are enough
- No breadcrumb integration — standard VSCode breadcrumbs work fine
- No custom editor or webview — standard tree view and text editor are sufficient
