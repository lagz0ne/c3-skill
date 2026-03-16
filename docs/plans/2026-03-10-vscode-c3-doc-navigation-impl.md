# VSCode C3 Document Navigation — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend the C3 Architecture Navigator VSCode extension to navigate all `.c3/` documents (markdown cross-references + tree view sidebar).

**Architecture:** Extend existing DocMap/providers pattern. DocMap gains new frontmatter fields (type, category, parent, uses, via, status). All three providers gain `.c3/**/*.md` support. New TreeDataProvider for sidebar. No new data layer — DocMap remains single source of truth.

**Tech Stack:** TypeScript, VSCode Extension API, vitest (new, for unit testing pure functions)

**Worktree:** `/Users/cuongtran/Desktop/repo/c3-skill/.worktrees/feat-vscode-c3-doc-nav`
**Extension root:** `vscode-c3-nav/`

---

## Task 1: Set Up Test Framework

**Files:**
- Modify: `vscode-c3-nav/package.json`
- Create: `vscode-c3-nav/src/__tests__/utils.test.ts`
- Modify: `vscode-c3-nav/tsconfig.json`

**Step 1: Install vitest**

```bash
cd vscode-c3-nav && npm install -D vitest
```

**Step 2: Add test script to package.json**

In `package.json`, add to `"scripts"`:
```json
"test": "vitest run",
"test:watch": "vitest"
```

**Step 3: Create a smoke test**

Create `src/__tests__/utils.test.ts`:
```typescript
import { describe, it, expect } from "vitest";
import { extractIdFromFilename, stripGlobSuffix } from "../utils";

describe("extractIdFromFilename", () => {
  it("extracts c3 ID from component filename", () => {
    expect(extractIdFromFilename("c3-113-check-cmd.md")).toBe("c3-113");
  });

  it("extracts ref ID from ref filename", () => {
    expect(extractIdFromFilename("ref-frontmatter-docs.md")).toBe("ref-frontmatter-docs");
  });

  it("returns undefined for README.md", () => {
    expect(extractIdFromFilename("README.md")).toBeUndefined();
  });
});

describe("stripGlobSuffix", () => {
  it("strips /** suffix", () => {
    expect(stripGlobSuffix("backend/app/**")).toBe("backend/app");
  });

  it("keeps non-glob paths", () => {
    expect(stripGlobSuffix("cli/main.go")).toBe("cli/main.go");
  });
});
```

**Step 4: Configure vitest to handle vscode import**

The `utils.ts` file imports only `fs` (no vscode), so vitest works directly. But future test files may need a vscode mock. For now, no extra config needed — vitest resolves the pure-function tests natively.

**Step 5: Run tests**

```bash
cd vscode-c3-nav && npm test
```

Expected: 5 passing tests.

**Step 6: Commit**

```bash
git add vscode-c3-nav/package.json vscode-c3-nav/package-lock.json vscode-c3-nav/src/__tests__/utils.test.ts
git commit -m "test: add vitest and smoke tests for utils"
```

---

## Task 2: Extend DocEntry Interface and parseFrontmatter

**Files:**
- Modify: `vscode-c3-nav/src/utils.ts:6-11` (DocEntry interface)
- Modify: `vscode-c3-nav/src/utils.ts:17-49` (parseFrontmatter function)
- Modify: `vscode-c3-nav/src/__tests__/utils.test.ts`

**Step 1: Write failing tests for new frontmatter fields**

Add to `src/__tests__/utils.test.ts`:
```typescript
import { parseFrontmatter } from "../utils";
import { writeFileSync, mkdirSync, rmSync } from "fs";
import { join } from "path";
import { beforeEach, afterEach } from "vitest";

describe("parseFrontmatter", () => {
  const tmpDir = join(__dirname, "__tmp__");
  beforeEach(() => mkdirSync(tmpDir, { recursive: true }));
  afterEach(() => rmSync(tmpDir, { recursive: true, force: true }));

  it("parses component frontmatter with all fields", () => {
    const file = join(tmpDir, "c3-113-check-cmd.md");
    writeFileSync(file, `---
id: c3-113
title: check-cmd
type: component
category: feature
parent: c3-1
goal: Validate docs
status: active
uses: [c3-101, c3-102, c3-104]
---
# check-cmd`);

    const result = parseFrontmatter(file);
    expect(result.title).toBe("check-cmd");
    expect(result.type).toBe("component");
    expect(result.category).toBe("feature");
    expect(result.parent).toBe("c3-1");
    expect(result.status).toBe("active");
    expect(result.uses).toEqual(["c3-101", "c3-102", "c3-104"]);
  });

  it("parses ref frontmatter with via field", () => {
    const file = join(tmpDir, "ref-test.md");
    writeFileSync(file, `---
id: ref-test
title: Test Ref
type: ref
goal: A test ref
via: [c3-101, c3-103]
---
# Test`);

    const result = parseFrontmatter(file);
    expect(result.type).toBe("ref");
    expect(result.via).toEqual(["c3-101", "c3-103"]);
    expect(result.uses).toBeUndefined();
  });

  it("parses ADR frontmatter", () => {
    const file = join(tmpDir, "adr-test.md");
    writeFileSync(file, `---
id: adr-00000000-c3-adoption
title: C3 Adoption
type: adr
status: in-progress
---
# ADR`);

    const result = parseFrontmatter(file);
    expect(result.type).toBe("adr");
    expect(result.status).toBe("in-progress");
  });

  it("defaults status to active when not specified", () => {
    const file = join(tmpDir, "c3-101.md");
    writeFileSync(file, `---
id: c3-101
title: Frontmatter
type: component
category: foundation
parent: c3-1
goal: Parse frontmatter
---`);

    const result = parseFrontmatter(file);
    expect(result.status).toBe("active");
  });
});
```

**Step 2: Run tests to verify they fail**

```bash
cd vscode-c3-nav && npm test
```

Expected: FAIL — `type`, `category`, `parent`, `uses`, `via`, `status` are not parsed yet.

**Step 3: Extend DocEntry interface**

In `utils.ts`, replace the DocEntry interface (lines 6-11):
```typescript
export interface DocEntry {
  path: string;
  title?: string;
  goal?: string;
  summary?: string;
  type?: "container" | "component" | "ref" | "adr";
  category?: "foundation" | "feature";
  parent?: string;
  uses?: string[];
  via?: string[];
  status?: string;
}
```

**Step 4: Extend parseFrontmatter**

In `utils.ts`, update parseFrontmatter return type and add parsing for new fields. After the existing `summaryMatch` block (line 46), add:

```typescript
const typeMatch = fm.match(/^type:\s*(.+)$/m);
if (typeMatch) {
  result.type = stripYamlQuotes(typeMatch[1].trim()) as DocEntry["type"];
}

const categoryMatch = fm.match(/^category:\s*(.+)$/m);
if (categoryMatch) {
  result.category = stripYamlQuotes(categoryMatch[1].trim()) as DocEntry["category"];
}

const parentMatch = fm.match(/^parent:\s*(.+)$/m);
if (parentMatch) {
  result.parent = stripYamlQuotes(parentMatch[1].trim());
}

const usesMatch = fm.match(/^uses:\s*\[([^\]]*)\]$/m);
if (usesMatch) {
  result.uses = usesMatch[1].split(",").map((s) => s.trim()).filter(Boolean);
}

const viaMatch = fm.match(/^via:\s*\[([^\]]*)\]$/m);
if (viaMatch) {
  result.via = viaMatch[1].split(",").map((s) => s.trim()).filter(Boolean);
}

const statusMatch = fm.match(/^status:\s*(.+)$/m);
result.status = statusMatch ? stripYamlQuotes(statusMatch[1].trim()) : "active";
```

Also update the return type signature from `Pick<DocEntry, "title" | "goal" | "summary">` to `Omit<DocEntry, "path">`.

Update the `result` variable type to match: `const result: Omit<DocEntry, "path"> = {};`

**Step 5: Run tests**

```bash
cd vscode-c3-nav && npm test
```

Expected: All tests pass.

**Step 6: Verify TypeScript compiles**

```bash
cd vscode-c3-nav && npx tsc -p ./ --noEmit
```

Expected: Clean compile (no errors).

**Step 7: Commit**

```bash
git add vscode-c3-nav/src/utils.ts vscode-c3-nav/src/__tests__/utils.test.ts
git commit -m "feat: extend DocEntry with type, category, parent, uses, via, status"
```

---

## Task 3: Handle README.md and ADR Filenames in DocMap

**Files:**
- Modify: `vscode-c3-nav/src/utils.ts:64-76` (extractIdFromFilename)
- Modify: `vscode-c3-nav/src/docMap.ts:18-34` (build method)
- Modify: `vscode-c3-nav/src/__tests__/utils.test.ts`

**Context:** Currently `extractIdFromFilename("README.md")` returns `undefined`, so container docs (c3-0, c3-1, c3-2) are never indexed. ADR filenames like `adr-00000000-c3-adoption.md` are also not matched. We need both in the DocMap for the tree view.

**Step 1: Write failing tests for ADR filename extraction**

Add to `src/__tests__/utils.test.ts` in the `extractIdFromFilename` describe block:
```typescript
it("extracts ADR ID from filename", () => {
  expect(extractIdFromFilename("adr-00000000-c3-adoption.md")).toBe("adr-00000000-c3-adoption");
});

it("extracts ADR ID with date prefix", () => {
  expect(extractIdFromFilename("adr-20260309-add-diff-cmd.md")).toBe("adr-20260309-add-diff-cmd");
});
```

**Step 2: Run tests — verify they fail**

```bash
cd vscode-c3-nav && npm test
```

Expected: FAIL — ADR pattern not handled.

**Step 3: Add ADR pattern to extractIdFromFilename**

In `utils.ts`, add before the final `return undefined` in `extractIdFromFilename`:
```typescript
const adrMatch = filename.match(/^(adr-[\w-]+)\.md$/);
if (adrMatch) {
  return adrMatch[1];
}
```

**Step 4: Run tests — verify they pass**

```bash
cd vscode-c3-nav && npm test
```

Expected: All tests pass.

**Step 5: Handle README.md in DocMap.build()**

README.md files need special handling — the ID comes from frontmatter, not filename. In `docMap.ts`, modify the `build()` method. Replace lines 18-34:

```typescript
for (const file of files) {
  const filename = path.basename(file.fsPath);
  let id = extractIdFromFilename(filename);

  // README.md files store their ID in frontmatter (containers like c3-0, c3-1, c3-2)
  if (!id && filename === "README.md") {
    const frontmatter = parseFrontmatter(file.fsPath);
    if (frontmatter.type === "container" || (frontmatter as Record<string, unknown>).id) {
      // Read id directly from frontmatter
      id = this.readFrontmatterId(file.fsPath);
    }
    if (!id) {
      continue;
    }
    if (this.map.has(id)) {
      console.warn(`[C3 Nav] Duplicate ID "${id}" — keeping first match, skipping ${file.fsPath}`);
      continue;
    }
    this.map.set(id, { path: file.fsPath, ...frontmatter });
    continue;
  }

  if (!id) {
    continue;
  }

  if (this.map.has(id)) {
    console.warn(`[C3 Nav] Duplicate ID "${id}" — keeping first match, skipping ${file.fsPath}`);
    continue;
  }

  const frontmatter = parseFrontmatter(file.fsPath);
  this.map.set(id, { path: file.fsPath, ...frontmatter });
}
```

Add helper method to `DocMap` class:
```typescript
private readFrontmatterId(filePath: string): string | undefined {
  const fs = require("fs");
  let content: string;
  try {
    content = fs.readFileSync(filePath, "utf-8");
  } catch {
    return undefined;
  }
  const fmMatch = content.match(/^---\n([\s\S]*?)\n---/);
  if (!fmMatch) {
    return undefined;
  }
  const idMatch = fmMatch[1].match(/^id:\s*(.+)$/m);
  return idMatch ? idMatch[1].trim() : undefined;
}
```

**Step 6: Verify TypeScript compiles**

```bash
cd vscode-c3-nav && npx tsc -p ./ --noEmit
```

**Step 7: Commit**

```bash
git add vscode-c3-nav/src/utils.ts vscode-c3-nav/src/docMap.ts vscode-c3-nav/src/__tests__/utils.test.ts
git commit -m "feat: handle README.md and ADR filenames in DocMap"
```

---

## Task 4: Add Markdown Context Detection Utilities

**Files:**
- Modify: `vscode-c3-nav/src/utils.ts`
- Modify: `vscode-c3-nav/src/__tests__/utils.test.ts`

**Context:** CodeLens in markdown needs to know: "Is this line inside frontmatter? Is it a markdown table row?" So we can apply smart placement (CodeLens in frontmatter + tables, not in prose).

**Step 1: Write failing tests**

Add to `src/__tests__/utils.test.ts`:
```typescript
import { isInFrontmatter, isMarkdownTableRow, getBacktickPathAtPosition } from "../utils";

describe("isInFrontmatter", () => {
  const lines = [
    "---",           // 0
    "id: c3-113",    // 1
    "parent: c3-1",  // 2
    "uses: [c3-101]",// 3
    "---",           // 4
    "",              // 5
    "# Title",       // 6
    "Body text c3-101", // 7
  ];

  it("returns true for lines inside frontmatter", () => {
    expect(isInFrontmatter(lines, 1)).toBe(true);
    expect(isInFrontmatter(lines, 2)).toBe(true);
    expect(isInFrontmatter(lines, 3)).toBe(true);
  });

  it("returns false for frontmatter delimiters", () => {
    expect(isInFrontmatter(lines, 0)).toBe(false);
    expect(isInFrontmatter(lines, 4)).toBe(false);
  });

  it("returns false for lines outside frontmatter", () => {
    expect(isInFrontmatter(lines, 5)).toBe(false);
    expect(isInFrontmatter(lines, 6)).toBe(false);
    expect(isInFrontmatter(lines, 7)).toBe(false);
  });
});

describe("isMarkdownTableRow", () => {
  it("matches table data rows", () => {
    expect(isMarkdownTableRow("| IN (uses) | Entity graph | c3-102 |")).toBe(true);
  });

  it("does not match separator rows", () => {
    expect(isMarkdownTableRow("|-----------|------|---------|")).toBe(false);
  });

  it("does not match non-table lines", () => {
    expect(isMarkdownTableRow("Some text mentioning c3-101")).toBe(false);
  });
});

describe("getBacktickPathAtPosition", () => {
  it("extracts path from backtick-wrapped text", () => {
    const line = "| `cli/internal/frontmatter/parse.go` | Parses YAML |";
    const result = getBacktickPathAtPosition(line, 10);
    expect(result).toBeDefined();
    expect(result!.rawPath).toBe("cli/internal/frontmatter/parse.go");
    expect(result!.folderPath).toBe("cli/internal/frontmatter/parse.go");
  });

  it("returns undefined when position is outside backticks", () => {
    const line = "| `cli/main.go` | Main entry |";
    const result = getBacktickPathAtPosition(line, 25);
    expect(result).toBeUndefined();
  });

  it("strips glob suffix from backtick paths", () => {
    const line = "Maps to `backend-core/app/**` for coverage";
    const result = getBacktickPathAtPosition(line, 15);
    expect(result!.rawPath).toBe("backend-core/app/**");
    expect(result!.folderPath).toBe("backend-core/app");
  });
});
```

**Step 2: Run tests — verify they fail**

```bash
cd vscode-c3-nav && npm test
```

**Step 3: Implement the utilities**

Add to `utils.ts`:

```typescript
/**
 * Check if a line index is inside the YAML frontmatter block.
 * Frontmatter is between the first `---` (line 0) and the next `---`.
 * Returns true for content lines, false for delimiters and outside.
 */
export function isInFrontmatter(lines: string[], lineIndex: number): boolean {
  if (lines[0] !== "---") {
    return false;
  }
  // Find the closing ---
  for (let i = 1; i < lines.length; i++) {
    if (lines[i] === "---") {
      return lineIndex > 0 && lineIndex < i;
    }
  }
  return false;
}

/**
 * Check if a line is a markdown table data row (not a separator row).
 * Table rows start and end with | and contain non-dash content.
 */
export function isMarkdownTableRow(line: string): boolean {
  const trimmed = line.trim();
  if (!trimmed.startsWith("|") || !trimmed.endsWith("|")) {
    return false;
  }
  // Separator rows contain only |, -, :, and spaces
  return !/^\|[\s|:-]+\|$/.test(trimmed);
}

/**
 * Get a backtick-wrapped file path at a given position in a line.
 * Matches patterns like `cli/internal/frontmatter/parse.go`.
 * Returns path info with glob suffix stripped for navigation.
 */
export function getBacktickPathAtPosition(
  lineText: string,
  characterPos: number
): { rawPath: string; folderPath: string; start: number; end: number } | undefined {
  const regex = /`([^`]+\.[a-z]+[^`]*)`|`([^`]+\/[^`]+)`/g;
  let match: RegExpExecArray | null;

  while ((match = regex.exec(lineText)) !== null) {
    const pathValue = match[1] || match[2];
    const start = match.index + 1; // after opening backtick
    const end = start + pathValue.length;
    if (characterPos >= start && characterPos <= end) {
      return {
        rawPath: pathValue,
        folderPath: stripGlobSuffix(pathValue),
        start,
        end,
      };
    }
  }

  return undefined;
}
```

**Step 4: Run tests**

```bash
cd vscode-c3-nav && npm test
```

Expected: All tests pass.

**Step 5: Commit**

```bash
git add vscode-c3-nav/src/utils.ts vscode-c3-nav/src/__tests__/utils.test.ts
git commit -m "feat: add markdown context detection utilities"
```

---

## Task 5: Extend CodeLensProvider for Markdown

**Files:**
- Modify: `vscode-c3-nav/src/codeLensProvider.ts`

**Context:** Currently only processes YAML files. Add markdown smart placement: CodeLens in frontmatter fields (`parent`, `uses`, `via`) and markdown table rows. No CodeLens in prose paragraphs.

**Step 1: Update imports**

At the top of `codeLensProvider.ts`, add:
```typescript
import { C3_ID_PATTERN, getPathAtPosition, isInFrontmatter, isMarkdownTableRow, getBacktickPathAtPosition } from "./utils";
```

**Step 2: Add markdown detection**

In the `provideCodeLenses` method, after `const lenses: vscode.CodeLens[] = [];`, add:
```typescript
const isMarkdown = document.languageId === "markdown";
const allLines = isMarkdown
  ? Array.from({ length: document.lineCount }, (_, i) => document.lineAt(i).text)
  : [];
```

**Step 3: Add smart placement guard for C3 ID lenses**

Inside the line loop, before the C3 ID regex matching, add a guard:
```typescript
// In markdown: only show CodeLens in frontmatter and table rows
if (isMarkdown) {
  const inFrontmatter = isInFrontmatter(allLines, i);
  const inTable = isMarkdownTableRow(line.text);
  if (!inFrontmatter && !inTable) {
    // Still process for path lenses in tables, skip C3 ID CodeLens in prose
    continue;
  }
}
```

Wait — this would skip path lenses too. Better approach: restructure the loop to check context per-match-type. Here's the full replacement for `provideCodeLenses`:

```typescript
provideCodeLenses(document: vscode.TextDocument): vscode.CodeLens[] {
  const lenses: vscode.CodeLens[] = [];
  const isMarkdown = document.languageId === "markdown";
  const allLines = isMarkdown
    ? Array.from({ length: document.lineCount }, (_, i) => document.lineAt(i).text)
    : [];

  for (let i = 0; i < document.lineCount; i++) {
    const line = document.lineAt(i);
    const showIdLens = !isMarkdown || isInFrontmatter(allLines, i) || isMarkdownTableRow(line.text);
    const showPathLens = !isMarkdown || isInFrontmatter(allLines, i) || isMarkdownTableRow(line.text);

    // C3 ID lenses (frontmatter + tables in markdown, everywhere in YAML)
    if (showIdLens) {
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

    // Path lenses — YAML: quoted paths, Markdown: backtick paths in tables
    if (isMarkdown) {
      if (showPathLens) {
        const pathMatch = getBacktickPathAtPosition(line.text, line.text.indexOf("`") + 1);
        if (pathMatch) {
          const range = new vscode.Range(i, pathMatch.start, i, pathMatch.end);
          lenses.push(
            new vscode.CodeLens(range, {
              title: `→ Open: ${pathMatch.folderPath}`,
              command: "c3Nav.revealPath",
              arguments: [pathMatch.folderPath],
            })
          );
        }
      }
    } else {
      const pathMatch = getPathAtPosition(line.text, line.text.indexOf('"') + 1);
      if (pathMatch) {
        const range = new vscode.Range(i, pathMatch.start, i, pathMatch.end);
        lenses.push(
          new vscode.CodeLens(range, {
            title: `→ Open: ${pathMatch.folderPath}`,
            command: "c3Nav.revealPath",
            arguments: [pathMatch.folderPath],
          })
        );
      }
    }
  }

  return lenses;
}
```

**Step 4: Verify TypeScript compiles**

```bash
cd vscode-c3-nav && npx tsc -p ./ --noEmit
```

**Step 5: Commit**

```bash
git add vscode-c3-nav/src/codeLensProvider.ts
git commit -m "feat: extend CodeLensProvider for markdown smart placement"
```

---

## Task 6: Extend DefinitionProvider for Markdown

**Files:**
- Modify: `vscode-c3-nav/src/definitionProvider.ts`

**Context:** Add backtick path support for markdown files. C3 ID navigation already works (regex matches in any text). Paths in markdown use backticks instead of quotes.

**Step 1: Update imports**

```typescript
import { getIdAtPosition, getPathAtPosition, getBacktickPathAtPosition } from "./utils";
```

**Step 2: Add backtick path handling**

In `provideDefinition`, after the existing quoted path block, add a backtick path fallback:

```typescript
// Try backtick path (markdown files)
const backtickMatch = getBacktickPathAtPosition(line, position.character);
if (backtickMatch) {
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  if (workspaceFolder) {
    const absPath = path.join(workspaceFolder.uri.fsPath, backtickMatch.folderPath);
    return new vscode.Location(vscode.Uri.file(absPath), new vscode.Position(0, 0));
  }
}
```

**Step 3: Verify TypeScript compiles**

```bash
cd vscode-c3-nav && npx tsc -p ./ --noEmit
```

**Step 4: Commit**

```bash
git add vscode-c3-nav/src/definitionProvider.ts
git commit -m "feat: extend DefinitionProvider with backtick path support"
```

---

## Task 7: Extend HoverProvider for Markdown

**Files:**
- Modify: `vscode-c3-nav/src/hoverProvider.ts`

**Context:** Add backtick path hover support. C3 ID hover already works. Also add status badge to doc hover cards.

**Step 1: Update imports**

```typescript
import { getIdAtPosition, getPathAtPosition, getBacktickPathAtPosition } from "./utils";
```

**Step 2: Add backtick path hover**

In `provideHover`, after the existing `pathMatch` block, add:

```typescript
// Try backtick path (markdown files)
const backtickMatch = getBacktickPathAtPosition(line, position.character);
if (backtickMatch) {
  return this.buildPathHover(backtickMatch);
}
```

**Step 3: Add status to doc hover card**

In `buildDocHover`, after the goal block (line 49), add:

```typescript
if (entry.status && entry.status !== "active") {
  md.appendMarkdown(`*Status:* \`${entry.status}\`\n\n`);
}
```

Update the `entry` parameter type to include `status`:
```typescript
entry: { path: string; title?: string; goal?: string; summary?: string; status?: string }
```

**Step 4: Verify TypeScript compiles**

```bash
cd vscode-c3-nav && npx tsc -p ./ --noEmit
```

**Step 5: Commit**

```bash
git add vscode-c3-nav/src/hoverProvider.ts
git commit -m "feat: extend HoverProvider with backtick paths and status badge"
```

---

## Task 8: Register Markdown Document Selector

**Files:**
- Modify: `vscode-c3-nav/src/extension.ts`

**Context:** All three providers need to be registered for `.c3/**/*.md` in addition to `.c3/**/*.yaml`.

**Step 1: Add markdown selector**

In `extension.ts`, after the existing `YAML_IN_C3` constant (line 9-12), add:

```typescript
const MD_IN_C3: vscode.DocumentSelector = {
  scheme: "file",
  pattern: "**/.c3/**/*.md",
};
```

**Step 2: Register providers for both selectors**

Replace the three provider registrations (lines 30-32) with:

```typescript
vscode.languages.registerCodeLensProvider(YAML_IN_C3, codeLensProvider),
vscode.languages.registerCodeLensProvider(MD_IN_C3, codeLensProvider),
vscode.languages.registerDefinitionProvider(YAML_IN_C3, new C3DefinitionProvider(docMap)),
vscode.languages.registerDefinitionProvider(MD_IN_C3, new C3DefinitionProvider(docMap)),
vscode.languages.registerHoverProvider(YAML_IN_C3, new C3HoverProvider(docMap)),
vscode.languages.registerHoverProvider(MD_IN_C3, new C3HoverProvider(docMap)),
```

**Step 3: Verify TypeScript compiles**

```bash
cd vscode-c3-nav && npx tsc -p ./ --noEmit
```

**Step 4: Commit**

```bash
git add vscode-c3-nav/src/extension.ts
git commit -m "feat: register providers for markdown files in .c3/"
```

---

## Task 9: Create TreeViewProvider — Data Structure

**Files:**
- Create: `vscode-c3-nav/src/treeViewProvider.ts`

**Context:** Build the tree data provider that reads from DocMap and renders hierarchy or flat view. This task covers the data model and tree structure. Next task adds context menu actions.

**Step 1: Create treeViewProvider.ts**

```typescript
import * as vscode from "vscode";
import * as path from "path";
import { DocMap } from "./docMap";
import { DocEntry } from "./utils";

type ViewMode = "hierarchy" | "flat";

interface C3TreeItem {
  id: string;
  entry: DocEntry;
}

export class C3TreeViewProvider implements vscode.TreeDataProvider<string> {
  private _onDidChangeTreeData = new vscode.EventEmitter<string | undefined>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  private viewMode: ViewMode = "hierarchy";

  constructor(private docMap: DocMap) {}

  refresh(): void {
    this._onDidChangeTreeData.fire(undefined);
  }

  toggleViewMode(): void {
    this.viewMode = this.viewMode === "hierarchy" ? "flat" : "hierarchy";
    this.refresh();
  }

  getViewMode(): ViewMode {
    return this.viewMode;
  }

  getTreeItem(element: string): vscode.TreeItem {
    // Group headers (non-ID nodes)
    if (element.startsWith("__group:")) {
      const label = element.replace("__group:", "");
      const item = new vscode.TreeItem(label, vscode.TreeItemCollapsibleState.Expanded);
      item.contextValue = "group";
      return item;
    }

    // Category sub-headers
    if (element.startsWith("__category:")) {
      const parts = element.replace("__category:", "").split(":");
      const label = parts[1].charAt(0).toUpperCase() + parts[1].slice(1);
      const item = new vscode.TreeItem(label, vscode.TreeItemCollapsibleState.Expanded);
      item.contextValue = "category";
      return item;
    }

    const entry = this.docMap.get(element);
    if (!entry) {
      return new vscode.TreeItem(element);
    }

    const hasChildren = this.getChildIds(element).length > 0;
    const collapsible = hasChildren
      ? vscode.TreeItemCollapsibleState.Expanded
      : vscode.TreeItemCollapsibleState.None;

    const label = entry.title ? `${element} · ${entry.title}` : element;
    const item = new vscode.TreeItem(label, collapsible);

    // Status badge
    if (entry.status && entry.status !== "active") {
      item.description = `[${entry.status}]`;
    }

    // Goal as tooltip
    if (entry.goal) {
      item.tooltip = entry.goal;
    }

    // Click opens the document
    item.command = {
      command: "c3Nav.openDocument",
      title: "Open Document",
      arguments: [entry.path],
    };

    item.contextValue = entry.type || "document";

    return item;
  }

  getChildren(element?: string): string[] {
    if (!element) {
      return this.viewMode === "hierarchy" ? this.getHierarchyRoots() : this.getFlatGroups();
    }

    if (this.viewMode === "flat") {
      return this.getFlatGroupChildren(element);
    }

    return this.getHierarchyChildren(element);
  }

  // --- Hierarchy mode ---

  private getHierarchyRoots(): string[] {
    const roots: string[] = [];

    // Find root container (c3-0)
    for (const [id, entry] of this.docMap.entries()) {
      if (entry.type === "container" && !entry.parent) {
        roots.push(id);
      }
    }

    // Add References and ADRs groups
    if (this.hasEntriesOfType("ref")) {
      roots.push("__group:References");
    }
    if (this.hasEntriesOfType("adr")) {
      roots.push("__group:ADRs");
    }

    return roots;
  }

  private getHierarchyChildren(element: string): string[] {
    if (element.startsWith("__group:")) {
      const group = element.replace("__group:", "");
      if (group === "References") {
        return this.getEntriesOfType("ref");
      }
      if (group === "ADRs") {
        return this.getEntriesOfType("adr");
      }
      return [];
    }

    if (element.startsWith("__category:")) {
      const parts = element.replace("__category:", "").split(":");
      const parentId = parts[0];
      const category = parts[1];
      return this.getChildIds(parentId).filter((childId) => {
        const entry = this.docMap.get(childId);
        return entry?.category === category;
      });
    }

    // For containers: group children by category if they have categories
    const children = this.getChildIds(element);
    const categories = new Set<string>();
    for (const childId of children) {
      const entry = this.docMap.get(childId);
      if (entry?.category) {
        categories.add(entry.category);
      }
    }

    if (categories.size > 1) {
      // Group by category
      return Array.from(categories)
        .sort()
        .map((cat) => `__category:${element}:${cat}`);
    }

    // No categories or single category — list children directly
    return children;
  }

  private getChildIds(parentId: string): string[] {
    const children: string[] = [];
    for (const [id, entry] of this.docMap.entries()) {
      if (entry.parent === parentId && entry.type !== "ref" && entry.type !== "adr") {
        children.push(id);
      }
    }
    return children.sort();
  }

  // --- Flat mode ---

  private getFlatGroups(): string[] {
    const groups: string[] = [];
    if (this.hasEntriesOfType("container")) {
      groups.push("__group:Containers");
    }
    if (this.hasEntriesOfType("component")) {
      groups.push("__group:Components");
    }
    if (this.hasEntriesOfType("ref")) {
      groups.push("__group:References");
    }
    if (this.hasEntriesOfType("adr")) {
      groups.push("__group:ADRs");
    }
    return groups;
  }

  private getFlatGroupChildren(element: string): string[] {
    const group = element.replace("__group:", "");
    const typeMap: Record<string, string> = {
      Containers: "container",
      Components: "component",
      References: "ref",
      ADRs: "adr",
    };
    return this.getEntriesOfType(typeMap[group] || "");
  }

  // --- Helpers ---

  private hasEntriesOfType(type: string): boolean {
    for (const [, entry] of this.docMap.entries()) {
      if (entry.type === type) {
        return true;
      }
    }
    return false;
  }

  private getEntriesOfType(type: string): string[] {
    const ids: string[] = [];
    for (const [id, entry] of this.docMap.entries()) {
      if (entry.type === type) {
        ids.push(id);
      }
    }
    return ids.sort();
  }
}
```

**Step 2: Verify TypeScript compiles**

```bash
cd vscode-c3-nav && npx tsc -p ./ --noEmit
```

**Step 3: Commit**

```bash
git add vscode-c3-nav/src/treeViewProvider.ts
git commit -m "feat: create C3TreeViewProvider with hierarchy and flat modes"
```

---

## Task 10: Add Context Menu Commands

**Files:**
- Modify: `vscode-c3-nav/src/treeViewProvider.ts`
- Create: `vscode-c3-nav/src/treeCommands.ts`

**Context:** Right-click actions: Show Dependencies, Show Dependents, Open in Code Map, Copy ID.

**Step 1: Create treeCommands.ts**

```typescript
import * as vscode from "vscode";
import * as path from "path";
import { DocMap } from "./docMap";

export function registerTreeCommands(
  context: vscode.ExtensionContext,
  docMap: DocMap,
  workspaceFolder: vscode.WorkspaceFolder
): void {
  context.subscriptions.push(
    vscode.commands.registerCommand("c3Nav.showDependencies", async (id: string) => {
      const entry = docMap.get(id);
      if (!entry?.uses || entry.uses.length === 0) {
        vscode.window.showInformationMessage(`${id} has no dependencies.`);
        return;
      }

      const items = entry.uses
        .map((depId) => {
          const dep = docMap.get(depId);
          return {
            label: depId,
            description: dep?.title,
            detail: dep?.goal,
            depPath: dep?.path,
          };
        })
        .filter((item) => item.depPath);

      const selected = await vscode.window.showQuickPick(items, {
        placeHolder: `Dependencies of ${id}`,
      });

      if (selected?.depPath) {
        vscode.window.showTextDocument(vscode.Uri.file(selected.depPath), { preview: true });
      }
    }),

    vscode.commands.registerCommand("c3Nav.showDependents", async (id: string) => {
      const dependents: Array<{ id: string; title?: string; path: string }> = [];

      for (const [entryId, entry] of docMap.entries()) {
        if (entry.uses?.includes(id) || entry.via?.includes(id)) {
          dependents.push({ id: entryId, title: entry.title, path: entry.path });
        }
      }

      if (dependents.length === 0) {
        vscode.window.showInformationMessage(`No components depend on ${id}.`);
        return;
      }

      const items = dependents.map((dep) => ({
        label: dep.id,
        description: dep.title,
        depPath: dep.path,
      }));

      const selected = await vscode.window.showQuickPick(items, {
        placeHolder: `Dependents of ${id}`,
      });

      if (selected?.depPath) {
        vscode.window.showTextDocument(vscode.Uri.file(selected.depPath), { preview: true });
      }
    }),

    vscode.commands.registerCommand("c3Nav.openInCodeMap", async (id: string) => {
      const codeMapPath = path.join(workspaceFolder.uri.fsPath, ".c3", "code-map.yaml");
      const doc = await vscode.workspace.openTextDocument(codeMapPath);
      const editor = await vscode.window.showTextDocument(doc, { preview: true });

      // Find the line containing this ID
      for (let i = 0; i < doc.lineCount; i++) {
        if (doc.lineAt(i).text.includes(id)) {
          const range = new vscode.Range(i, 0, i, 0);
          editor.revealRange(range, vscode.TextEditorRevealType.InCenter);
          editor.selection = new vscode.Selection(range.start, range.start);
          break;
        }
      }
    }),

    vscode.commands.registerCommand("c3Nav.copyId", (id: string) => {
      vscode.env.clipboard.writeText(id);
      vscode.window.showInformationMessage(`Copied: ${id}`);
    })
  );
}
```

**Step 2: Verify TypeScript compiles**

```bash
cd vscode-c3-nav && npx tsc -p ./ --noEmit
```

**Step 3: Commit**

```bash
git add vscode-c3-nav/src/treeCommands.ts
git commit -m "feat: add tree context menu commands"
```

---

## Task 11: Update package.json — Views, Commands, Menus

**Files:**
- Modify: `vscode-c3-nav/package.json`

**Context:** Declare the tree view container, view, all new commands, and context menu bindings.

**Step 1: Update package.json contributes section**

Replace the entire `"contributes"` block:

```json
"contributes": {
  "views": {
    "explorer": [
      {
        "id": "c3Navigator",
        "name": "C3 Architecture",
        "when": "c3Nav.hasC3Folder"
      }
    ]
  },
  "commands": [
    {
      "command": "c3Nav.openDocument",
      "title": "C3: Open Architecture Document"
    },
    {
      "command": "c3Nav.revealPath",
      "title": "C3: Reveal Path"
    },
    {
      "command": "c3Nav.toggleViewMode",
      "title": "C3: Toggle Hierarchy/Flat View",
      "icon": "$(list-tree)"
    },
    {
      "command": "c3Nav.showDependencies",
      "title": "Show Dependencies"
    },
    {
      "command": "c3Nav.showDependents",
      "title": "Show Dependents"
    },
    {
      "command": "c3Nav.openInCodeMap",
      "title": "Open in Code Map"
    },
    {
      "command": "c3Nav.copyId",
      "title": "Copy ID"
    }
  ],
  "menus": {
    "view/title": [
      {
        "command": "c3Nav.toggleViewMode",
        "when": "view == c3Navigator",
        "group": "navigation"
      }
    ],
    "view/item/context": [
      {
        "command": "c3Nav.showDependencies",
        "when": "view == c3Navigator && viewItem =~ /component|container/",
        "group": "c3nav@1"
      },
      {
        "command": "c3Nav.showDependents",
        "when": "view == c3Navigator && viewItem =~ /component|ref/",
        "group": "c3nav@2"
      },
      {
        "command": "c3Nav.openInCodeMap",
        "when": "view == c3Navigator && viewItem =~ /component|container|ref/",
        "group": "c3nav@3"
      },
      {
        "command": "c3Nav.copyId",
        "when": "view == c3Navigator && viewItem != group && viewItem != category",
        "group": "c3nav@4"
      }
    ]
  }
}
```

**Step 2: Verify JSON is valid**

```bash
cd vscode-c3-nav && node -e "JSON.parse(require('fs').readFileSync('package.json','utf8'));console.log('Valid JSON')"
```

**Step 3: Commit**

```bash
git add vscode-c3-nav/package.json
git commit -m "feat: declare tree view, commands, and context menus in package.json"
```

---

## Task 12: Wire Everything in extension.ts

**Files:**
- Modify: `vscode-c3-nav/src/extension.ts`

**Context:** Register the tree view provider, toggle command, context commands, and set the `c3Nav.hasC3Folder` context for conditional view visibility.

**Step 1: Replace extension.ts**

```typescript
import * as vscode from "vscode";
import * as path from "path";
import * as fs from "fs";
import { DocMap } from "./docMap";
import { C3CodeLensProvider } from "./codeLensProvider";
import { C3DefinitionProvider } from "./definitionProvider";
import { C3HoverProvider } from "./hoverProvider";
import { C3TreeViewProvider } from "./treeViewProvider";
import { registerTreeCommands } from "./treeCommands";

const YAML_IN_C3: vscode.DocumentSelector = {
  scheme: "file",
  pattern: "**/.c3/**/*.yaml",
};

const MD_IN_C3: vscode.DocumentSelector = {
  scheme: "file",
  pattern: "**/.c3/**/*.md",
};

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  if (!workspaceFolder) {
    return;
  }

  // Set context for conditional view visibility
  vscode.commands.executeCommand("setContext", "c3Nav.hasC3Folder", true);

  const docMap = new DocMap();
  await docMap.build(workspaceFolder);

  // CodeLens
  const codeLensProvider = new C3CodeLensProvider(docMap);
  docMap.onDidRebuild(() => codeLensProvider.refresh());

  // Tree View
  const treeViewProvider = new C3TreeViewProvider(docMap);
  docMap.onDidRebuild(() => treeViewProvider.refresh());

  context.subscriptions.push(
    // File watching
    docMap.startWatching(workspaceFolder),

    // CodeLens for YAML and Markdown
    vscode.languages.registerCodeLensProvider(YAML_IN_C3, codeLensProvider),
    vscode.languages.registerCodeLensProvider(MD_IN_C3, codeLensProvider),

    // Definition for YAML and Markdown
    vscode.languages.registerDefinitionProvider(YAML_IN_C3, new C3DefinitionProvider(docMap)),
    vscode.languages.registerDefinitionProvider(MD_IN_C3, new C3DefinitionProvider(docMap)),

    // Hover for YAML and Markdown
    vscode.languages.registerHoverProvider(YAML_IN_C3, new C3HoverProvider(docMap)),
    vscode.languages.registerHoverProvider(MD_IN_C3, new C3HoverProvider(docMap)),

    // Tree view
    vscode.window.createTreeView("c3Navigator", {
      treeDataProvider: treeViewProvider,
      showCollapseAll: true,
    }),

    // Commands
    vscode.commands.registerCommand("c3Nav.openDocument", (filePath: string) => {
      const uri = vscode.Uri.file(filePath);
      vscode.window.showTextDocument(uri, { preview: true });
    }),
    vscode.commands.registerCommand("c3Nav.revealPath", async (relativePath: string) => {
      const absPath = path.join(workspaceFolder.uri.fsPath, relativePath);
      const uri = vscode.Uri.file(absPath);

      if (fs.existsSync(absPath) && fs.statSync(absPath).isDirectory()) {
        await vscode.commands.executeCommand("revealInExplorer", uri);
        await vscode.commands.executeCommand("list.expand");
      } else if (fs.existsSync(absPath)) {
        vscode.window.showTextDocument(uri, { preview: true });
      } else {
        vscode.window.showWarningMessage(`Path not found: ${relativePath}`);
      }
    }),
    vscode.commands.registerCommand("c3Nav.toggleViewMode", () => {
      treeViewProvider.toggleViewMode();
    })
  );

  // Register tree context menu commands
  registerTreeCommands(context, docMap, workspaceFolder);

  console.log("[C3 Nav] Extension activated");
}

export function deactivate(): void {
  // cleanup handled by disposables
}
```

**Step 2: Verify TypeScript compiles**

```bash
cd vscode-c3-nav && npx tsc -p ./ --noEmit
```

**Step 3: Run tests**

```bash
cd vscode-c3-nav && npm test
```

Expected: All tests still pass.

**Step 4: Commit**

```bash
git add vscode-c3-nav/src/extension.ts
git commit -m "feat: wire tree view, markdown providers, and all commands"
```

---

## Task 13: Build and Smoke Test

**Step 1: Full compile**

```bash
cd vscode-c3-nav && npx tsc -p ./
```

Expected: Clean compile, output in `out/`.

**Step 2: Package extension**

```bash
cd vscode-c3-nav && npm run package
```

Expected: Produces `c3-nav.vsix` without errors.

**Step 3: Run all tests**

```bash
cd vscode-c3-nav && npm test
```

Expected: All tests pass.

**Step 4: Bump version to 0.2.0**

In `package.json`, change `"version": "0.1.0"` to `"version": "0.2.0"`.

Update description:
```json
"description": "Navigate C3 architecture documents — cross-reference links, hover previews, and tree view for .c3/ files"
```

**Step 5: Final commit**

```bash
git add vscode-c3-nav/
git commit -m "feat(vscode-c3-nav): v0.2.0 — markdown navigation and tree view"
```

---

## Summary

| Task | What | Commit |
|------|------|--------|
| 1 | Test framework setup | `test: add vitest and smoke tests for utils` |
| 2 | Extend DocEntry + parseFrontmatter | `feat: extend DocEntry with type, category, parent, uses, via, status` |
| 3 | README.md + ADR filename handling | `feat: handle README.md and ADR filenames in DocMap` |
| 4 | Markdown context detection utils | `feat: add markdown context detection utilities` |
| 5 | CodeLensProvider for markdown | `feat: extend CodeLensProvider for markdown smart placement` |
| 6 | DefinitionProvider for markdown | `feat: extend DefinitionProvider with backtick path support` |
| 7 | HoverProvider for markdown | `feat: extend HoverProvider with backtick paths and status badge` |
| 8 | Register markdown document selector | `feat: register providers for markdown files in .c3/` |
| 9 | TreeViewProvider (data structure) | `feat: create C3TreeViewProvider with hierarchy and flat modes` |
| 10 | Context menu commands | `feat: add tree context menu commands` |
| 11 | package.json views/commands/menus | `feat: declare tree view, commands, and context menus in package.json` |
| 12 | Wire everything in extension.ts | `feat: wire tree view, markdown providers, and all commands` |
| 13 | Build, smoke test, version bump | `feat(vscode-c3-nav): v0.2.0 — markdown navigation and tree view` |
