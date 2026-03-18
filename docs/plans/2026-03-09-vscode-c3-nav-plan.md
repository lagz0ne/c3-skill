# VSCode C3 Navigator — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a VS Code extension that provides CodeLens, Ctrl+Click, and Hover navigation for C3 IDs (`c3-XXX`, `ref-XXX`) in YAML files within `.c3/` directories.

**Architecture:** The extension scans `.c3/**/*.md` on activation, builds an in-memory ID→document map by matching filenames against C3 ID patterns, and registers three VS Code providers (CodeLens, Definition, Hover) for YAML files in `.c3/`. A file watcher keeps the map current.

**Tech Stack:** TypeScript, VS Code Extension API, `@vscode/vsce` for packaging

---

### Task 1: Scaffold the extension project

**Files:**
- Create: `vscode-c3-nav/package.json`
- Create: `vscode-c3-nav/tsconfig.json`
- Create: `vscode-c3-nav/.vscodeignore`
- Create: `vscode-c3-nav/.gitignore`

**Step 1: Create `package.json`**

```json
{
  "name": "c3-nav",
  "displayName": "C3 Architecture Navigator",
  "description": "Navigate C3 architecture documents from code-map.yaml and other .c3/ YAML files",
  "version": "0.1.0",
  "publisher": "lagz0ne",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/Lagz0ne/c3-skill"
  },
  "engines": {
    "vscode": "^1.85.0"
  },
  "categories": ["Other"],
  "activationEvents": [
    "workspaceContains:.c3/code-map.yaml"
  ],
  "main": "./out/extension.js",
  "contributes": {
    "commands": [
      {
        "command": "c3Nav.openDocument",
        "title": "C3: Open Architecture Document"
      }
    ]
  },
  "scripts": {
    "compile": "tsc -p ./",
    "watch": "tsc -watch -p ./",
    "package": "vsce package --out c3-nav.vsix",
    "prepackage": "npm run compile"
  },
  "devDependencies": {
    "@types/vscode": "^1.85.0",
    "@types/node": "^20.0.0",
    "typescript": "^5.3.0",
    "@vscode/vsce": "^3.0.0"
  }
}
```

**Step 2: Create `tsconfig.json`**

```json
{
  "compilerOptions": {
    "module": "commonjs",
    "target": "ES2022",
    "outDir": "out",
    "rootDir": "src",
    "lib": ["ES2022"],
    "sourceMap": true,
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "exclude": ["node_modules", "out"]
}
```

**Step 3: Create `.vscodeignore`**

```
.gitignore
src/**
tsconfig.json
node_modules/**
*.vsix
```

**Step 4: Create `.gitignore`**

```
out/
node_modules/
*.vsix
```

**Step 5: Install dependencies**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npm install`
Expected: `node_modules/` created, no errors

**Step 6: Verify TypeScript compiles (empty project)**

Run: `mkdir -p src && touch src/extension.ts && npx tsc -p ./ --noEmit`
Expected: No errors (empty file is valid)

**Step 7: Commit**

```bash
git add vscode-c3-nav/package.json vscode-c3-nav/tsconfig.json vscode-c3-nav/.vscodeignore vscode-c3-nav/.gitignore
git commit -m "feat(vscode-c3-nav): scaffold extension project"
```

---

### Task 2: Implement shared utilities (`utils.ts`)

**Files:**
- Create: `vscode-c3-nav/src/utils.ts`

**Step 1: Create `utils.ts` with ID regex and frontmatter parser**

```typescript
import * as fs from "fs";

/** Matches c3-0 through c3-999 and ref-xxx-yyy patterns */
export const C3_ID_PATTERN = /\b(c3-\d{1,3}|ref-[a-z][\w-]*)\b/g;

/** Matches a single ID at a position (non-global for positional matching) */
export const C3_ID_PATTERN_SINGLE = /\b(c3-\d{1,3}|ref-[a-z][\w-]*)\b/;

export interface DocEntry {
  path: string;
  title?: string;
  goal?: string;
  summary?: string;
}

/**
 * Parse YAML frontmatter from a markdown file.
 * Extracts title, goal, and summary fields from the --- delimited block.
 */
export function parseFrontmatter(filePath: string): Pick<DocEntry, "title" | "goal" | "summary"> {
  let content: string;
  try {
    content = fs.readFileSync(filePath, "utf-8");
  } catch {
    return {};
  }

  const fmMatch = content.match(/^---\n([\s\S]*?)\n---/);
  if (!fmMatch) {
    return {};
  }

  const fm = fmMatch[1];
  const result: Pick<DocEntry, "title" | "goal" | "summary"> = {};

  const titleMatch = fm.match(/^title:\s*(.+)$/m);
  if (titleMatch) {
    result.title = titleMatch[1].trim();
  }

  const goalMatch = fm.match(/^goal:\s*(.+)$/m);
  if (goalMatch) {
    result.goal = goalMatch[1].trim();
  }

  const summaryMatch = fm.match(/^summary:\s*(.+)$/m);
  if (summaryMatch) {
    result.summary = summaryMatch[1].trim();
  }

  return result;
}

/**
 * Extract the C3/ref ID from a markdown filename.
 * e.g. "c3-101-titan-framework.md" → "c3-101"
 *      "ref-auth-patterns.md" → "ref-auth-patterns"
 * Returns undefined for non-matching filenames (e.g. "README.md").
 */
export function extractIdFromFilename(filename: string): string | undefined {
  const c3Match = filename.match(/^(c3-\d{1,3})-/);
  if (c3Match) {
    return c3Match[1];
  }

  const refMatch = filename.match(/^(ref-[a-z][\w-]*)\.md$/);
  if (refMatch) {
    return refMatch[1];
  }

  return undefined;
}

/**
 * Get the word (C3 ID) at a given position in a line of text.
 * Returns the matched ID and its start/end character positions, or undefined.
 */
export function getIdAtPosition(
  lineText: string,
  characterPos: number
): { id: string; start: number; end: number } | undefined {
  const regex = new RegExp(C3_ID_PATTERN.source, "g");
  let match: RegExpExecArray | null;

  while ((match = regex.exec(lineText)) !== null) {
    const start = match.index;
    const end = start + match[0].length;
    if (characterPos >= start && characterPos <= end) {
      return { id: match[0], start, end };
    }
  }

  return undefined;
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npx tsc -p ./ --noEmit`
Expected: No errors

**Step 3: Commit**

```bash
git add vscode-c3-nav/src/utils.ts
git commit -m "feat(vscode-c3-nav): add shared utils with ID regex and frontmatter parser"
```

---

### Task 3: Implement the document map (`docMap.ts`)

**Files:**
- Create: `vscode-c3-nav/src/docMap.ts`

**Step 1: Create `docMap.ts`**

```typescript
import * as vscode from "vscode";
import * as path from "path";
import { DocEntry, extractIdFromFilename, parseFrontmatter } from "./utils";

export class DocMap {
  private map = new Map<string, DocEntry>();
  private watcher: vscode.FileSystemWatcher | undefined;

  async build(workspaceFolder: vscode.WorkspaceFolder): Promise<void> {
    this.map.clear();

    const c3Dir = vscode.Uri.joinPath(workspaceFolder.uri, ".c3");
    const pattern = new vscode.RelativePattern(c3Dir, "**/*.md");
    const files = await vscode.workspace.findFiles(pattern);

    for (const file of files) {
      const filename = path.basename(file.fsPath);
      const id = extractIdFromFilename(filename);
      if (!id) {
        continue;
      }

      if (this.map.has(id)) {
        console.warn(`[C3 Nav] Duplicate ID "${id}" — keeping first match, skipping ${file.fsPath}`);
        continue;
      }

      const frontmatter = parseFrontmatter(file.fsPath);
      this.map.set(id, {
        path: file.fsPath,
        ...frontmatter,
      });
    }

    console.log(`[C3 Nav] Built document map with ${this.map.size} entries`);
  }

  get(id: string): DocEntry | undefined {
    return this.map.get(id);
  }

  entries(): IterableIterator<[string, DocEntry]> {
    return this.map.entries();
  }

  startWatching(workspaceFolder: vscode.WorkspaceFolder): vscode.Disposable {
    const c3Dir = vscode.Uri.joinPath(workspaceFolder.uri, ".c3");
    const pattern = new vscode.RelativePattern(c3Dir, "**/*.md");
    this.watcher = vscode.workspace.createFileSystemWatcher(pattern);

    const rebuild = () => this.build(workspaceFolder);
    this.watcher.onDidCreate(rebuild);
    this.watcher.onDidDelete(rebuild);
    this.watcher.onDidChange(rebuild);

    return this.watcher;
  }

  dispose(): void {
    this.watcher?.dispose();
  }
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npx tsc -p ./ --noEmit`
Expected: No errors

**Step 3: Commit**

```bash
git add vscode-c3-nav/src/docMap.ts
git commit -m "feat(vscode-c3-nav): add DocMap with file scanning and watcher"
```

---

### Task 4: Implement CodeLens provider (`codeLensProvider.ts`)

**Files:**
- Create: `vscode-c3-nav/src/codeLensProvider.ts`

**Step 1: Create `codeLensProvider.ts`**

```typescript
import * as vscode from "vscode";
import * as path from "path";
import { DocMap } from "./docMap";
import { C3_ID_PATTERN } from "./utils";

export class C3CodeLensProvider implements vscode.CodeLensProvider {
  private _onDidChangeCodeLenses = new vscode.EventEmitter<void>();
  readonly onDidChangeCodeLenses = this._onDidChangeCodeLenses.event;

  constructor(private docMap: DocMap) {}

  refresh(): void {
    this._onDidChangeCodeLenses.fire();
  }

  provideCodeLenses(document: vscode.TextDocument): vscode.CodeLens[] {
    const lenses: vscode.CodeLens[] = [];

    for (let i = 0; i < document.lineCount; i++) {
      const line = document.lineAt(i);
      const regex = new RegExp(C3_ID_PATTERN.source, "g");
      let match: RegExpExecArray | null;

      while ((match = regex.exec(line.text)) !== null) {
        const id = match[0];
        const entry = this.docMap.get(id);
        if (!entry) {
          continue;
        }

        const range = new vscode.Range(i, match.index, i, match.index + id.length);
        const filename = path.basename(entry.path);
        const title = entry.title ? `${entry.title} (${filename})` : filename;

        lenses.push(
          new vscode.CodeLens(range, {
            title: `→ ${title}`,
            command: "c3Nav.openDocument",
            arguments: [entry.path],
          })
        );
      }
    }

    return lenses;
  }
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npx tsc -p ./ --noEmit`
Expected: No errors

**Step 3: Commit**

```bash
git add vscode-c3-nav/src/codeLensProvider.ts
git commit -m "feat(vscode-c3-nav): add CodeLens provider for C3 IDs"
```

---

### Task 5: Implement Definition provider (`definitionProvider.ts`)

**Files:**
- Create: `vscode-c3-nav/src/definitionProvider.ts`

**Step 1: Create `definitionProvider.ts`**

```typescript
import * as vscode from "vscode";
import { DocMap } from "./docMap";
import { getIdAtPosition } from "./utils";

export class C3DefinitionProvider implements vscode.DefinitionProvider {
  constructor(private docMap: DocMap) {}

  provideDefinition(
    document: vscode.TextDocument,
    position: vscode.Position
  ): vscode.Definition | undefined {
    const line = document.lineAt(position.line).text;
    const match = getIdAtPosition(line, position.character);
    if (!match) {
      return undefined;
    }

    const entry = this.docMap.get(match.id);
    if (!entry) {
      return undefined;
    }

    return new vscode.Location(vscode.Uri.file(entry.path), new vscode.Position(0, 0));
  }
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npx tsc -p ./ --noEmit`
Expected: No errors

**Step 3: Commit**

```bash
git add vscode-c3-nav/src/definitionProvider.ts
git commit -m "feat(vscode-c3-nav): add Definition provider for Ctrl+Click navigation"
```

---

### Task 6: Implement Hover provider (`hoverProvider.ts`)

**Files:**
- Create: `vscode-c3-nav/src/hoverProvider.ts`

**Step 1: Create `hoverProvider.ts`**

```typescript
import * as vscode from "vscode";
import { DocMap } from "./docMap";
import { getIdAtPosition } from "./utils";

export class C3HoverProvider implements vscode.HoverProvider {
  constructor(private docMap: DocMap) {}

  provideHover(
    document: vscode.TextDocument,
    position: vscode.Position
  ): vscode.Hover | undefined {
    const line = document.lineAt(position.line).text;
    const match = getIdAtPosition(line, position.character);
    if (!match) {
      return undefined;
    }

    const entry = this.docMap.get(match.id);
    if (!entry) {
      return undefined;
    }

    const md = new vscode.MarkdownString("", true);
    md.isTrusted = true;

    if (entry.title) {
      md.appendMarkdown(`**${match.id}** — ${entry.title}\n\n`);
    } else {
      md.appendMarkdown(`**${match.id}**\n\n`);
    }

    if (entry.goal) {
      md.appendMarkdown(`*Goal:* ${entry.goal}\n\n`);
    }

    if (entry.summary) {
      md.appendMarkdown(`*Summary:* ${entry.summary}\n\n`);
    }

    const fileUri = vscode.Uri.file(entry.path);
    md.appendMarkdown(`[Open document](${fileUri})`);

    const range = new vscode.Range(
      position.line,
      match.start,
      position.line,
      match.end
    );

    return new vscode.Hover(md, range);
  }
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npx tsc -p ./ --noEmit`
Expected: No errors

**Step 3: Commit**

```bash
git add vscode-c3-nav/src/hoverProvider.ts
git commit -m "feat(vscode-c3-nav): add Hover provider with frontmatter preview"
```

---

### Task 7: Implement extension entry point (`extension.ts`)

**Files:**
- Create: `vscode-c3-nav/src/extension.ts`

**Step 1: Create `extension.ts`**

```typescript
import * as vscode from "vscode";
import { DocMap } from "./docMap";
import { C3CodeLensProvider } from "./codeLensProvider";
import { C3DefinitionProvider } from "./definitionProvider";
import { C3HoverProvider } from "./hoverProvider";

const YAML_IN_C3: vscode.DocumentSelector = {
  scheme: "file",
  pattern: "**/.c3/**/*.yaml",
};

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  if (!workspaceFolder) {
    return;
  }

  const docMap = new DocMap();
  await docMap.build(workspaceFolder);

  const codeLensProvider = new C3CodeLensProvider(docMap);

  context.subscriptions.push(
    docMap.startWatching(workspaceFolder),
    vscode.languages.registerCodeLensProvider(YAML_IN_C3, codeLensProvider),
    vscode.languages.registerDefinitionProvider(YAML_IN_C3, new C3DefinitionProvider(docMap)),
    vscode.languages.registerHoverProvider(YAML_IN_C3, new C3HoverProvider(docMap)),
    vscode.commands.registerCommand("c3Nav.openDocument", (filePath: string) => {
      const uri = vscode.Uri.file(filePath);
      vscode.window.showTextDocument(uri, { preview: true });
    })
  );

  // Refresh CodeLens when doc map rebuilds
  const watcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(vscode.Uri.joinPath(workspaceFolder.uri, ".c3"), "**/*.md")
  );
  watcher.onDidCreate(() => codeLensProvider.refresh());
  watcher.onDidDelete(() => codeLensProvider.refresh());
  watcher.onDidChange(() => codeLensProvider.refresh());
  context.subscriptions.push(watcher);

  console.log("[C3 Nav] Extension activated");
}

export function deactivate(): void {
  // cleanup handled by disposables
}
```

**Step 2: Verify full project compiles**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npx tsc -p ./`
Expected: No errors, `out/` directory created with `.js` files

**Step 3: Commit**

```bash
git add vscode-c3-nav/src/extension.ts
git commit -m "feat(vscode-c3-nav): add extension entry point wiring all providers"
```

---

### Task 8: Build and test the `.vsix` package

**Files:**
- None new — verify packaging works

**Step 1: Build the `.vsix`**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npm run package`
Expected: `c3-nav.vsix` created in `vscode-c3-nav/`

**Step 2: Verify the `.vsix` contents**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npx vsce ls`
Expected: Lists `package.json`, `out/*.js`, `out/*.js.map` — no `src/`, no `node_modules/`

**Step 3: Commit**

```bash
git add vscode-c3-nav/
git commit -m "feat(vscode-c3-nav): verify extension builds and packages"
```

---

### Task 9: Add install script and Claude command

**Files:**
- Create: `scripts/install-vscode-ext.sh`
- Create: `commands/c3-setup-vscode.md`

**Step 1: Create `scripts/install-vscode-ext.sh`**

```bash
#!/bin/bash
# Build and install the C3 Navigator VS Code extension
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
EXT_DIR="$SCRIPT_DIR/../vscode-c3-nav"

if [ ! -d "$EXT_DIR" ]; then
  echo "Error: vscode-c3-nav directory not found at $EXT_DIR"
  exit 1
fi

echo "Building C3 Navigator extension..."
cd "$EXT_DIR"
npm install --ignore-scripts
npm run package

VSIX=$(ls -1 "$EXT_DIR"/c3-nav.vsix 2>/dev/null | head -1)
if [ -z "$VSIX" ]; then
  echo "Error: .vsix file not found after build"
  exit 1
fi

echo "Installing extension..."
code --install-extension "$VSIX" --force

echo "C3 Navigator extension installed successfully."
echo "Reload VS Code to activate it."
```

**Step 2: Make it executable**

Run: `chmod +x /Users/cuongtran/Desktop/repo/c3-skill/scripts/install-vscode-ext.sh`

**Step 3: Create `commands/c3-setup-vscode.md`**

```markdown
---
description: Install the C3 Navigator VS Code extension for architecture doc navigation
---

Run the install script to build and install the C3 Navigator VS Code extension:

```bash
bash "$(dirname "$(find ~/.claude -path '*/c3-skill/scripts/install-vscode-ext.sh' 2>/dev/null | head -1)")/install-vscode-ext.sh"
```

After installation, reload VS Code. The extension activates in any workspace with a `.c3/code-map.yaml` file and provides:
- **CodeLens**: Clickable links above `c3-XXX` and `ref-XXX` IDs in `.c3/*.yaml` files
- **Ctrl+Click**: Navigate directly to the architecture document
- **Hover**: Preview document title, goal, and summary
```

**Step 4: Commit**

```bash
git add scripts/install-vscode-ext.sh commands/c3-setup-vscode.md
git commit -m "feat(vscode-c3-nav): add install script and Claude command"
```

---

### Task 10: Final verification and cleanup

**Files:**
- None new — end-to-end verification

**Step 1: Full clean build**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && rm -rf out node_modules && npm install && npm run compile`
Expected: Clean build with no errors

**Step 2: Package**

Run: `cd /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav && npm run package`
Expected: `c3-nav.vsix` created

**Step 3: Install and test in VS Code**

Run: `code --install-extension /Users/cuongtran/Desktop/repo/c3-skill/vscode-c3-nav/c3-nav.vsix --force`
Expected: Extension installs successfully

**Step 4: Manual test in VS Code**

Open `pvs-core-i-full` workspace in VS Code, open `.c3/code-map.yaml`:
1. Verify CodeLens appears above `c3-101:`, `c3-102:`, `ref-auth-patterns:` etc.
2. Verify Ctrl+Click on `c3-101` navigates to `.c3/c3-1-backend/c3-101-titan-framework.md`
3. Verify hovering over `c3-101` shows title "Titan Framework" and goal text

**Step 5: Final commit**

```bash
git add -A
git commit -m "feat(vscode-c3-nav): C3 Architecture Navigator extension v0.1.0"
```
