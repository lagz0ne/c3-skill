# C3 Architecture Navigator — VS Code Extension

Navigate C3 architecture documents directly from your editor. Cross-reference links, hover previews, tree view sidebar, and Ctrl+Click navigation for all `.c3/` files.

## Features

| Feature | Description |
|---------|-------------|
| **CodeLens** | Inline links above C3 IDs in YAML and markdown — click to open the architecture doc |
| **Hover** | Preview document title, goal, summary, and status on hover |
| **Go to Definition** | Ctrl+Click on any C3 ID (`c3-*`, `ref-*`, `adr-*`) to jump to its document |
| **Path Navigation** | Click file paths in `code-map.yaml` to open the source file or reveal folder |
| **Tree View** | Sidebar panel showing C3 architecture hierarchy with toggle between hierarchy and flat views |
| **Context Menu** | Right-click tree items for Show Dependencies, Show Dependents, Open in Code Map, Copy ID |

## Tree View

The **C3 Architecture** panel appears in the Explorer sidebar. Two modes toggled via toolbar button:

**Hierarchy mode (default):**

```
c3-0 · C3 Design                          [active]
├── c3-1 · Go CLI                          [active]
│   ├── Foundation
│   │   ├── c3-101 · Frontmatter Parser    [active]
│   │   ├── c3-102 · Walker                [active]
│   │   ├── c3-103 · Templates             [active]
│   │   ├── c3-104 · Wiring               [active]
│   │   └── c3-105 · Codemap Lib           [active]
│   └── Feature
│       ├── c3-110 · Init Cmd              [active]
│       ├── c3-111 · Add Cmd               [active]
│       ├── ...
│       └── c3-117 · Diff Cmd              [provisioned]
├── c3-2 · Claude Skill                    [active]
│   ├── c3-201 · Skill Router              [active]
│   └── c3-210 · Operation Refs            [active]
├── References
│   ├── ref-frontmatter-docs
│   ├── ref-embedded-templates
│   └── ref-cross-compiled-binary
└── ADRs
    ├── adr-00000000 · C3 Adoption
    └── adr-20260309 · Add Diff Cmd
```

**Flat mode:**

```
Containers (3)
Components (15)
References (3)
ADRs (2)
```

Click any item to open its `.md` document. Goal shown as tooltip on hover.

## Activation

The extension activates automatically when a workspace contains `.c3/code-map.yaml`.

## Install

### From GitHub Release (recommended)

1. Go to [Releases](https://github.com/Lagz0ne/c3-skill/releases/latest)
2. Download `c3-nav.vsix`
3. Install:

```bash
code --install-extension c3-nav.vsix --force
```

4. Reload VS Code

### From Source

```bash
cd vscode-c3-nav
npm install
npm run package
code --install-extension c3-nav.vsix --force
```

## Version

Managed independently in `package.json` (currently `0.3.0`).
