# C3 Architecture Navigator — VS Code Extension

Navigate C3 architecture documents directly from your editor. Provides CodeLens links, hover previews, Ctrl+Click navigation, and Go to Definition for C3 IDs (`c3-*`, `ref-*`, `adr-*`) found in `.c3/` YAML files and code-map entries.

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

## Features

| Feature | Description |
|---------|-------------|
| **CodeLens** | Inline links above C3 IDs — click to open the architecture doc |
| **Hover** | Preview frontmatter (id, type, status, refs) on hover |
| **Go to Definition** | Ctrl+Click on any C3 ID to jump to its `.c3/` document |
| **Path Navigation** | Click file paths in `code-map.yaml` to open the source file |

## Activation

The extension activates automatically when a workspace contains `.c3/code-map.yaml`.

## Version

Managed independently in `package.json` (currently `0.1.0`).
