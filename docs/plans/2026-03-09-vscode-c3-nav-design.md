# VSCode C3 Navigator Extension — Design

**Date:** 2026-03-09
**Status:** Approved
**Location:** `c3-skill/vscode-c3-nav/`

## Goal

Provide clickable navigation for C3 component IDs (`c3-XXX`) and ref IDs (`ref-XXX`) in YAML files within `.c3/`, linking them to their corresponding architecture documents.

## Activation

- **Trigger:** `workspaceContains:.c3/code-map.yaml`
- **On activation:**
  1. Scan `.c3/` recursively for all `.md` files
  2. Build an in-memory map: `{ "c3-101": ".c3/c3-1-backend/c3-101-titan-framework.md", ... }`
  3. Match by checking if the filename starts with the ID (e.g., `c3-101-titan-framework.md` matches key `c3-101`)
  4. Watch `.c3/` for file changes and rebuild the map automatically
- **File targeting:** Providers register for YAML files matching `**/.c3/**/*.yaml`

## Navigation Features

### 1. CodeLens (clickable links above each ID)

For each line matching `c3-XXX:` or `ref-XXX:` in a YAML file, show a CodeLens like `→ Open c3-101-titan-framework.md`. Clicking opens the corresponding `.md` document.

### 2. Ctrl+Click (Go to Definition)

Register a `DefinitionProvider` for YAML files. When the cursor is on a `c3-XXX` or `ref-XXX` token, Ctrl+Click navigates directly to the `.md` file (line 1).

### 3. Hover

When hovering over a `c3-XXX` or `ref-XXX` token, show a tooltip with:
- The document title (parsed from YAML frontmatter `title:` field)
- The file path as a clickable link
- The `goal` or `summary` from frontmatter if present

## Document Map & Pattern Matching

**ID detection regex:**
```
/\b(c3-\d{1,3}|ref-[a-z][\w-]*)\b/
```

**Document map building (`docMap.ts`):**
1. On activation, recursively glob `.c3/**/*.md`
2. For each `.md` file, extract the base ID from the filename:
   - `c3-101-titan-framework.md` → `c3-101`
   - `ref-auth-patterns.md` → `ref-auth-patterns`
3. Parse YAML frontmatter (between `---` markers) to extract `title`, `goal`, `summary` for hover tooltips
4. Store as `Map<string, { path, title?, goal?, summary? }>`

**File watcher:**
- Watch `.c3/**/*.md` with `vscode.workspace.createFileSystemWatcher`
- On create/delete/rename → rebuild the map

**Edge cases:**
- `README.md` files (no ID prefix) → skipped
- Duplicate IDs → first match wins, log warning
- Missing frontmatter → tooltip shows just the file path

## Project Structure

```
c3-skill/
  vscode-c3-nav/
    package.json              # Extension manifest
    tsconfig.json             # TypeScript config
    src/
      extension.ts            # Entry: activate(), register providers, build doc map
      docMap.ts               # Scan .c3/, build ID→path map, watch for changes
      codeLensProvider.ts     # CodeLens provider
      definitionProvider.ts   # Go-to-definition provider
      hoverProvider.ts        # Hover provider with frontmatter preview
      utils.ts                # Shared regex, frontmatter parsing
    .vscodeignore             # Exclude from .vsix
  scripts/
    install-vscode-ext.sh     # Build + install .vsix
  commands/
    c3-setup-vscode.md        # Claude command to trigger install
```

**Dependencies:**
- `vscode` engine API
- `js-yaml` for frontmatter parsing

## Distribution

**Install script (`scripts/install-vscode-ext.sh`):**
- Builds the `.vsix` if not present (`npm install && npm run package`)
- Installs via `code --install-extension`

**Claude command (`commands/c3-setup-vscode.md`):**
- Runs the install script
- Provides user feedback on success/failure

**Version:** managed in `vscode-c3-nav/package.json`, script always builds fresh.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Interaction model | CodeLens + Ctrl+Click + Hover | Full navigation experience |
| Scope | YAML files in `.c3/` | Focused, avoids noise |
| Doc resolution | File scan | Zero-config, adapts automatically |
| Language | TypeScript | VS Code standard, strong typing |
| Location | `c3-skill/vscode-c3-nav/` | Part of C3 plugin ecosystem |
| Distribution | Bundled `.vsix` + Claude command | Self-contained, easy install |
