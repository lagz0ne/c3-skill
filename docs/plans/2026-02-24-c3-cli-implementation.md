# C3 CLI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the MCP server with a pure file-based CLI that treats `.c3/` as the database

**Architecture:** CLI with pumped-fn context wiring. Core modules (frontmatter, walker, numbering, wiring) provide shared graph logic. Each command is a pumped-fn `flow`. No SQLite, no embeddings — pure file operations.

**Tech Stack:** TypeScript, @pumped-fn/lite, zod, yaml (npm package)

**Design doc:** `docs/plans/2026-02-24-c3-cli-design.md`

---

## Workstream Overview

```
WS-A: Core modules        (frontmatter, walker, numbering, wiring, config update)
WS-B: CLI framework        (entry point, help, output, arg parsing)
WS-C: Data commands         (list, read, trace, check, impact) — depends on A+B
WS-D: Action commands       (init, add, evolve, template) — depends on A+B
WS-E: Ref + sync commands   (ref add/usage/check/link, sync) — depends on A+B
WS-F: Build & cleanup       (package.json, build script, CI, remove MCP) — depends on all
```

**Parallelism:** A and B can run in parallel. C, D, E can run in parallel once A+B complete. F runs last.

---

## Task 0: Package Setup

**Files:**
- Modify: `/home/lagz0ne/c3-design/package.json`
- Modify: `/home/lagz0ne/c3-design/tsconfig.json`

**Step 1: Install new dependencies**

```bash
cd /home/lagz0ne/c3-design
bun add @pumped-fn/lite yaml
bun remove @modelcontextprotocol/sdk
```

**Step 2: Update package.json**

Update the name, bin, and scripts:

```json
{
  "name": "c3-cli",
  "version": "5.0.0",
  "type": "module",
  "bin": {
    "c3": "./dist/cli.js"
  },
  "files": ["dist/"],
  "scripts": {
    "build": "bun run scripts/build.ts",
    "build:cli": "bun build ./src/cli/index.ts --outfile ./dist/cli.js --target node",
    "dev": "bun run src/cli/index.ts",
    "check": "bunx @typescript/native-preview --noEmit",
    "check-refs": "bun run scripts/check-refs.ts",
    "fix-refs": "bun run scripts/fix-refs.ts"
  },
  "dependencies": {
    "@pumped-fn/lite": "^2.0.0",
    "yaml": "^2.7.0",
    "zod": "^4.3.6"
  },
  "devDependencies": {
    "bun-types": "latest"
  }
}
```

**Step 3: Commit**

```bash
git add package.json bun.lock
git commit -m "chore: replace MCP deps with CLI deps (pumped-fn, yaml)"
```

---

## Task 1: Core — frontmatter.ts

**Files:**
- Create: `src/core/frontmatter.ts`

**What it does:** Parse YAML frontmatter from markdown files. Extract structured metadata. This is the foundation everything else builds on.

**Implementation:**

```typescript
// src/core/frontmatter.ts
import { parse as parseYaml } from "yaml";
import { z } from "zod";

export const frontmatterSchema = z.object({
  id: z.string(),
  "c3-version": z.number().optional(),
  title: z.string().optional(),
  type: z.enum(["container", "component", "adr"]).optional(),
  category: z.string().optional(),
  parent: z.string().optional(),
  goal: z.string().optional(),
  summary: z.string().optional(),
  boundary: z.string().optional(),
  affects: z.array(z.string()).optional(),
  status: z.string().optional(),
  date: z.string().optional(),
  scope: z.array(z.string()).optional(),
  refs: z.array(z.string()).optional(),
}).passthrough();

export type Frontmatter = z.infer<typeof frontmatterSchema>;

export type DocType = "context" | "container" | "component" | "ref" | "adr";

export interface ParsedDoc {
  frontmatter: Frontmatter;
  body: string;
  path: string;
}

export function parseFrontmatter(content: string): { frontmatter: Frontmatter | null; body: string } {
  if (!content.startsWith("---\n")) return { frontmatter: null, body: content };

  const end = content.indexOf("\n---\n", 4);
  if (end === -1) return { frontmatter: null, body: content };

  const yamlStr = content.slice(4, end);
  const body = content.slice(end + 5);

  try {
    const parsed = parseYaml(yamlStr);
    const result = frontmatterSchema.safeParse(parsed);
    return { frontmatter: result.success ? result.data : null, body };
  } catch {
    return { frontmatter: null, body: content };
  }
}

export function classifyDoc(fm: Frontmatter): DocType | null {
  if (fm.id === "c3-0") return "context";
  if (fm.type === "container") return "container";
  if (fm.type === "component") return "component";
  if (fm.type === "adr" || fm.id.startsWith("adr-")) return "adr";
  if (fm.id.startsWith("ref-")) return "ref";
  return null;
}

export function deriveRelationships(fm: Frontmatter): string[] {
  const rels: string[] = [];
  if (fm.parent) rels.push(fm.parent);
  if (fm.affects) rels.push(...fm.affects);
  if (fm.refs) rels.push(...fm.refs);
  if (fm.scope) rels.push(...fm.scope);
  return rels;
}
```

**Commit:**

```bash
git add src/core/frontmatter.ts
git commit -m "feat(core): add frontmatter parser with YAML/zod validation"
```

---

## Task 2: Core — walker.ts

**Files:**
- Create: `src/core/walker.ts`

**What it does:** Walk the `.c3/` directory, parse all docs, build a typed relationship graph. This is the in-memory "database" that all commands use.

**Implementation:**

```typescript
// src/core/walker.ts
import * as fs from "fs";
import * as path from "path";
import { parseFrontmatter, classifyDoc, deriveRelationships, type Frontmatter, type DocType, type ParsedDoc } from "./frontmatter";

export interface C3Entity {
  id: string;
  type: DocType;
  title: string;
  slug: string;
  path: string;         // relative to .c3/
  frontmatter: Frontmatter;
  body: string;
  relationships: string[];  // IDs this entity references
}

export interface C3Graph {
  entities: Map<string, C3Entity>;
  byType: Map<DocType, C3Entity[]>;

  // Relationship queries
  children(parentId: string): C3Entity[];
  refsFor(entityId: string): C3Entity[];
  citedBy(refId: string): C3Entity[];
  forward(id: string): C3Entity[];        // what does this affect
  reverse(id: string): C3Entity[];        // what points to this
  transitive(id: string, depth: number): C3Entity[]; // blast radius
}

export async function walkC3Docs(c3Dir: string): Promise<ParsedDoc[]> {
  const docs: ParsedDoc[] = [];

  function walk(dir: string) {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        walk(fullPath);
      } else if (entry.name.endsWith(".md")) {
        const content = fs.readFileSync(fullPath, "utf-8");
        const { frontmatter, body } = parseFrontmatter(content);
        if (frontmatter) {
          docs.push({
            frontmatter,
            body,
            path: path.relative(c3Dir, fullPath),
          });
        }
      }
    }
  }

  walk(c3Dir);
  return docs;
}

function slugFromPath(filePath: string): string {
  const base = path.basename(filePath, ".md");
  // Strip ID prefix: c3-1-api -> api, c3-101-auth -> auth, ref-logging -> logging
  return base.replace(/^(c3-\d+-|c3-\d+|ref-|adr-\d+-|README)/, "") || base;
}

export function buildRelationshipGraph(docs: ParsedDoc[]): C3Graph {
  const entities = new Map<string, C3Entity>();
  const byType = new Map<DocType, C3Entity[]>();

  // Phase 1: Build entity map
  for (const doc of docs) {
    const type = classifyDoc(doc.frontmatter);
    if (!type) continue;

    const entity: C3Entity = {
      id: doc.frontmatter.id,
      type,
      title: doc.frontmatter.title || doc.frontmatter.id,
      slug: slugFromPath(doc.path),
      path: doc.path,
      frontmatter: doc.frontmatter,
      body: doc.body,
      relationships: deriveRelationships(doc.frontmatter),
    };

    entities.set(entity.id, entity);
    const list = byType.get(type) || [];
    list.push(entity);
    byType.set(type, list);
  }

  // Phase 2: Build graph query methods
  function children(parentId: string): C3Entity[] {
    return [...entities.values()].filter(e => e.frontmatter.parent === parentId);
  }

  function refsFor(entityId: string): C3Entity[] {
    const entity = entities.get(entityId);
    if (!entity) return [];
    return (entity.frontmatter.refs || [])
      .map(id => entities.get(id))
      .filter((e): e is C3Entity => !!e);
  }

  function citedBy(refId: string): C3Entity[] {
    return [...entities.values()].filter(e =>
      e.frontmatter.refs?.includes(refId) ||
      e.frontmatter.scope?.includes(refId)
    );
  }

  function forward(id: string): C3Entity[] {
    const entity = entities.get(id);
    if (!entity) return [];

    const result: C3Entity[] = [];
    // Direct children
    result.push(...children(id));
    // Entities in affects list
    if (entity.frontmatter.affects) {
      for (const affectedId of entity.frontmatter.affects) {
        const affected = entities.get(affectedId);
        if (affected) result.push(affected);
      }
    }
    // If this is a ref, find citers
    if (entity.type === "ref") {
      result.push(...citedBy(id));
    }
    return result;
  }

  function reverse(id: string): C3Entity[] {
    return [...entities.values()].filter(e =>
      e.relationships.includes(id) ||
      e.frontmatter.parent === id ||
      e.frontmatter.affects?.includes(id)
    );
  }

  function transitive(id: string, depth: number): C3Entity[] {
    const visited = new Set<string>([id]);
    const result: C3Entity[] = [];
    let frontier = [id];

    for (let d = 0; d < depth && frontier.length > 0; d++) {
      const nextFrontier: string[] = [];
      for (const currentId of frontier) {
        for (const entity of forward(currentId)) {
          if (!visited.has(entity.id)) {
            visited.add(entity.id);
            result.push(entity);
            nextFrontier.push(entity.id);
          }
        }
      }
      frontier = nextFrontier;
    }
    return result;
  }

  return { entities, byType, children, refsFor, citedBy, forward, reverse, transitive };
}
```

**Commit:**

```bash
git add src/core/walker.ts
git commit -m "feat(core): add doc walker and relationship graph builder"
```

---

## Task 3: Core — numbering.ts

**Files:**
- Create: `src/core/numbering.ts`

**What it does:** Auto-numbering for new entities. Scans existing IDs and picks the next available.

**Implementation:**

```typescript
// src/core/numbering.ts
import type { C3Graph } from "./walker";

export function nextContainerId(graph: C3Graph): number {
  const containers = graph.byType.get("container") || [];
  const nums = containers
    .map(c => parseInt(c.id.replace("c3-", ""), 10))
    .filter(n => !isNaN(n));
  return nums.length === 0 ? 1 : Math.max(...nums) + 1;
}

export function nextComponentId(graph: C3Graph, containerNum: number, feature: boolean): string {
  const prefix = `c3-${containerNum}`;
  const components = (graph.byType.get("component") || [])
    .filter(c => c.id.startsWith(prefix))
    .map(c => parseInt(c.id.replace(prefix, ""), 10))
    .filter(n => !isNaN(n));

  if (feature) {
    // Feature: 10+
    const featureNums = components.filter(n => n >= 10);
    const next = featureNums.length === 0 ? 10 : Math.max(...featureNums) + 1;
    return `c3-${containerNum}${String(next).padStart(2, "0")}`;
  } else {
    // Foundation: 01-09
    const foundationNums = components.filter(n => n >= 1 && n <= 9);
    const next = foundationNums.length === 0 ? 1 : Math.max(...foundationNums) + 1;
    if (next > 9) throw new Error(`Container c3-${containerNum} has no more foundation slots (01-09 full)`);
    return `c3-${containerNum}${String(next).padStart(2, "0")}`;
  }
}

export function nextAdrId(slug: string): string {
  const date = new Date().toISOString().slice(0, 10).replace(/-/g, "");
  return `adr-${date}-${slug}`;
}
```

**Commit:**

```bash
git add src/core/numbering.ts
git commit -m "feat(core): add auto-numbering for containers, components, ADRs"
```

---

## Task 4: Core — wiring.ts

**Files:**
- Create: `src/core/wiring.ts`

**What it does:** Update cross-links between entities when adding/linking. Modifies frontmatter in markdown files.

**Implementation:**

```typescript
// src/core/wiring.ts
import * as fs from "fs";
import * as path from "path";
import { stringify as stringifyYaml } from "yaml";
import { parseFrontmatter, type Frontmatter } from "./frontmatter";

export function updateFrontmatterField(filePath: string, field: string, value: unknown): void {
  const content = fs.readFileSync(filePath, "utf-8");
  const { frontmatter, body } = parseFrontmatter(content);
  if (!frontmatter) throw new Error(`No frontmatter in ${filePath}`);

  (frontmatter as Record<string, unknown>)[field] = value;
  const newContent = `---\n${stringifyYaml(frontmatter).trim()}\n---\n${body}`;
  fs.writeFileSync(filePath, newContent, "utf-8");
}

export function addToFrontmatterArray(filePath: string, field: string, value: string): void {
  const content = fs.readFileSync(filePath, "utf-8");
  const { frontmatter, body } = parseFrontmatter(content);
  if (!frontmatter) throw new Error(`No frontmatter in ${filePath}`);

  const arr = ((frontmatter as Record<string, unknown>)[field] as string[]) || [];
  if (!arr.includes(value)) {
    arr.push(value);
    (frontmatter as Record<string, unknown>)[field] = arr;
    const newContent = `---\n${stringifyYaml(frontmatter).trim()}\n---\n${body}`;
    fs.writeFileSync(filePath, newContent, "utf-8");
  }
}

export function addComponentToContainerTable(containerReadmePath: string, componentId: string, name: string, category: string, goal: string): void {
  let content = fs.readFileSync(containerReadmePath, "utf-8");
  // Find the Components table and append a row
  const tablePattern = /(\| ID \| Name \| Category \| Status \| Goal Contribution \|[\s\S]*?)(\n\n|\n##|\n---|\Z)/;
  const match = content.match(tablePattern);
  if (match) {
    const newRow = `| ${componentId} | ${name} | ${category} | active | ${goal} |\n`;
    content = content.replace(match[0], match[1] + newRow + match[2]);
    fs.writeFileSync(containerReadmePath, content, "utf-8");
  }
}

export function linkRefToComponent(c3Dir: string, refId: string, componentId: string, componentPath: string): void {
  // Update component's refs array in frontmatter
  const compFullPath = path.join(c3Dir, componentPath);
  addToFrontmatterArray(compFullPath, "refs", refId);
}
```

**Commit:**

```bash
git add src/core/wiring.ts
git commit -m "feat(core): add relationship wiring for cross-link updates"
```

---

## Task 5: Core — update config.ts

**Files:**
- Modify: `src/core/config.ts`

**What it does:** Replace `Bun.YAML.parse()` with the `yaml` npm package. Remove embedding config (no longer needed). Keep `findC3Dir` and `loadConfig`.

**Implementation — replace entire file:**

```typescript
// src/core/config.ts
import * as fs from "fs";
import * as path from "path";
import { parse as parseYaml } from "yaml";
import { z } from "zod";

const configSchema = z.object({
  // Minimal config — embedding removed, add future CLI config here
}).default({});

export type C3Config = z.infer<typeof configSchema>;

const DEFAULT_CONFIG: C3Config = configSchema.parse({});

export function loadConfig(c3Dir: string): C3Config {
  const configPath = path.join(c3Dir, "config.yaml");
  if (!fs.existsSync(configPath)) return DEFAULT_CONFIG;
  const raw = fs.readFileSync(configPath, "utf-8");
  const parsed = parseYaml(raw);
  const result = configSchema.safeParse(parsed);
  return result.success ? result.data : DEFAULT_CONFIG;
}

export function findC3Dir(startDir: string): string | null {
  let dir = path.resolve(startDir);
  while (true) {
    const candidate = path.join(dir, ".c3");
    if (fs.existsSync(candidate) && fs.statSync(candidate).isDirectory()) {
      return candidate;
    }
    const parent = path.dirname(dir);
    if (parent === dir) return null;
    dir = parent;
  }
}
```

**Commit:**

```bash
git add src/core/config.ts
git commit -m "refactor(core): replace Bun.YAML with yaml package, remove embedding config"
```

---

## Task 6: CLI Framework — context atoms and tags

**Files:**
- Create: `src/cli/context.ts`

**What it does:** Define the pumped-fn atoms, tags, and shared context that all commands use.

**Implementation:**

```typescript
// src/cli/context.ts
import { atom, tag, tags, createScope, type Lite } from "@pumped-fn/lite";
import { findC3Dir, loadConfig, type C3Config } from "../core/config";
import { walkC3Docs, buildRelationshipGraph, type C3Graph } from "../core/walker";

// Tags — ambient context per CLI invocation
export const c3DirTag = tag<string>({ label: "c3Dir" });

export interface CliOptions {
  command: string;
  args: string[];
  json: boolean;
  flat?: boolean;
  reverse?: boolean;
  depth?: number;
  docs?: boolean;
  code?: boolean;
  feature?: boolean;
  container?: string;
  refs?: string[];
  list?: boolean;
  c3Dir?: string;
  help: boolean;
  version: boolean;
}

export const optionsTag = tag<CliOptions>({ label: "options" });

// Atoms — cached per scope, built once
export const configAtom = atom({
  deps: { c3Dir: tags.required(c3DirTag) },
  factory: (_ctx, { c3Dir }) => loadConfig(c3Dir),
});

export const graphAtom = atom({
  deps: { c3Dir: tags.required(c3DirTag) },
  factory: async (_ctx, { c3Dir }) => {
    const docs = await walkC3Docs(c3Dir);
    return buildRelationshipGraph(docs);
  },
});
```

**Commit:**

```bash
git add src/cli/context.ts
git commit -m "feat(cli): add pumped-fn context atoms and tags"
```

---

## Task 7: CLI Framework — help.ts

**Files:**
- Create: `src/cli/help.ts`

**What it does:** Rich help text for global and per-command help. Modeled after agent-browser.

**Implementation:**

```typescript
// src/cli/help.ts

const GLOBAL_HELP = `
c3 — Architecture-aware toolkit for C3 projects

Usage: c3 <command> [options]

Data:
  list                   Topology view with relationships
  read <path>            Read a C3 doc (by ID or path)
  trace <name>           Follow relationship chains
  check                  Doc integrity + code coverage
  impact <name>          Blast radius of a change

Actions:
  init                   Scaffold .c3/ skeleton
  add <type> <slug>      Create entity with auto-numbering
  evolve <path>          Update a C3 doc
  template <type>        Emit doc template
  sync                   Generate guard skills

Refs:
  ref add <slug>         Create a reference pattern
  ref usage <id>         Find components citing a ref
  ref check [id]         Verify ref compliance
  ref link <ref> <comp>  Wire ref to component

Options:
  -h, --help             Show help (with command for details)
  -v, --version          Print version
  --json                 Machine-readable output
  --c3-dir <path>        Override .c3/ detection

Run 'c3 <command> --help' for details and examples.
`.trim();

const COMMAND_HELP: Record<string, string> = {
  list: `
Usage: c3 list [options]

Topology view of all C3 architecture docs with relationships

Options:
  --flat                 Simple file list (no topology)
  --json                 Output as JSON

Examples:
  c3 list
  c3 list --flat
  c3 list --json
`.trim(),

  read: `
Usage: c3 read <path> [options]

Read a C3 doc by ID or path (relative to .c3/)

Options:
  --json                 Output as JSON (frontmatter + body)

Examples:
  c3 read c3-101
  c3 read ref-logging
  c3 read c3-1-api/README
`.trim(),

  trace: `
Usage: c3 trace <name> [options]

Follow relationship chains from a C3 entity

Options:
  --reverse              Trace what points TO this entity
  --depth <n>            Max hops to follow (default: 1)
  --json                 Output as JSON

Examples:
  c3 trace c3-101
  c3 trace --reverse c3-210
  c3 trace ref-logging --depth 3
`.trim(),

  check: `
Usage: c3 check [options]

Doc integrity and code-to-doc coverage check

Options:
  --docs                 Doc integrity only
  --code                 Code coverage only
  --json                 Output as JSON

Examples:
  c3 check
  c3 check --docs
  c3 check --json
`.trim(),

  impact: `
Usage: c3 impact <name> [options]

Blast radius: transitive closure of what's affected if this entity changes

Options:
  --depth <n>            Max hops (default: 2)
  --json                 Output as JSON

Examples:
  c3 impact c3-101
  c3 impact c3-1
  c3 impact ref-logging --depth 3
`.trim(),

  init: `
Usage: c3 init

Scaffold a new .c3/ directory skeleton

Creates:
  .c3/config.yaml
  .c3/README.md              (context template)
  .c3/refs/                  (empty)
  .c3/adr/adr-00000000-c3-adoption.md

Examples:
  c3 init
`.trim(),

  add: `
Usage: c3 add <type> <slug> [options]

Create a new C3 entity with auto-numbering and relationship wiring

Types: container, component, ref, adr

Options:
  --container <id>       Parent container (required for component)
  --feature              Feature component (10+) instead of foundation (01-09)
  --refs <ids>           Comma-separated refs to link

Examples:
  c3 add container payments
  c3 add component auth-provider --container c3-3
  c3 add component checkout --container c3-3 --feature
  c3 add ref rate-limiting
  c3 add adr oauth-support
`.trim(),

  evolve: `
Usage: c3 evolve <path>

Update a C3 doc. Reads new content from stdin.

Examples:
  echo "new content" | c3 evolve c3-101
  cat updated.md | c3 evolve ref-logging
`.trim(),

  template: `
Usage: c3 template <type> [options]

Emit a doc template to stdout

Types: context, container, component, ref, adr

Options:
  --list                 Show available template types

Examples:
  c3 template component
  c3 template --list
`.trim(),

  sync: `
Usage: c3 sync

Generate guard skills from .c3/ component docs into .claude/skills/

Reads Code References from each component doc and generates
per-component Claude Code skills that trigger on matching file paths.

Examples:
  c3 sync
`.trim(),

  ref: `
Usage: c3 ref <subcommand>

Subcommands:
  add <slug>             Create a reference pattern doc
  usage <id>             Find components citing a ref
  check [id]             Verify ref compliance
  link <ref> <comp>      Wire ref to component

Examples:
  c3 ref add rate-limiting
  c3 ref usage ref-logging
  c3 ref check ref-logging
  c3 ref link ref-logging c3-301
`.trim(),
};

export function showHelp(command?: string): void {
  if (command && COMMAND_HELP[command]) {
    console.log(COMMAND_HELP[command]);
  } else {
    console.log(GLOBAL_HELP);
  }
}
```

**Commit:**

```bash
git add src/cli/help.ts
git commit -m "feat(cli): add rich help text for all commands"
```

---

## Task 8: CLI Framework — output.ts

**Files:**
- Create: `src/cli/output.ts`

**What it does:** Output formatting — plain text rendering and JSON output mode.

**Implementation:**

```typescript
// src/cli/output.ts
import type { C3Entity, C3Graph } from "../core/walker";
import type { DocType } from "../core/frontmatter";

export function renderTopology(graph: C3Graph): string {
  const lines: string[] = [];

  // Containers with their components
  const containers = graph.byType.get("container") || [];
  for (const container of containers.sort((a, b) => a.id.localeCompare(b.id))) {
    lines.push(`${container.id}-${container.slug} (container)`);
    const components = graph.children(container.id).sort((a, b) => a.id.localeCompare(b.id));
    for (let i = 0; i < components.length; i++) {
      const comp = components[i];
      const isLast = i === components.length - 1;
      const prefix = isLast ? "└── " : "├── ";
      const category = comp.frontmatter.category || (parseInt(comp.id.replace(/c3-\d+/, ""), 10) <= 9 ? "foundation" : "feature");
      const refs = graph.refsFor(comp.id).map(r => r.id).join(", ");
      const suffix = refs ? ` → ref: ${refs}` : "";
      lines.push(`${prefix}${comp.id}-${comp.slug} (${category})${suffix}`);
    }
    lines.push("");
  }

  // Cross-cutting refs
  const refs = graph.byType.get("ref") || [];
  if (refs.length > 0) {
    lines.push("Cross-cutting:");
    for (const ref of refs.sort((a, b) => a.id.localeCompare(b.id))) {
      const citers = graph.citedBy(ref.id).map(c => c.id).join(", ");
      lines.push(`  ${ref.id}${citers ? ` → used by: ${citers}` : ""}`);
    }
    lines.push("");
  }

  // ADRs
  const adrs = graph.byType.get("adr") || [];
  if (adrs.length > 0) {
    lines.push("ADRs:");
    for (const adr of adrs.sort((a, b) => a.id.localeCompare(b.id))) {
      const status = adr.frontmatter.status || "unknown";
      lines.push(`  ${adr.id}: ${adr.title} → status: ${status}`);
    }
    lines.push("");
  }

  return lines.join("\n").trim();
}

export function renderFlatList(graph: C3Graph): string {
  return [...graph.entities.values()]
    .sort((a, b) => a.path.localeCompare(b.path))
    .map(e => `${e.id}\t${e.type}\t${e.path}`)
    .join("\n");
}

export function renderTrace(entities: C3Entity[], rootId: string, reverse: boolean): string {
  if (entities.length === 0) return `No ${reverse ? "incoming" : "outgoing"} relationships for ${rootId}`;
  const direction = reverse ? "→ points to" : "← depends on";
  return entities.map(e => `${e.id} (${e.type}) ${direction} ${rootId}`).join("\n");
}

export function renderJson(data: unknown): string {
  return JSON.stringify(data, null, 2);
}
```

**Commit:**

```bash
git add src/cli/output.ts
git commit -m "feat(cli): add output formatters for topology, trace, JSON"
```

---

## Task 9: CLI Framework — entry point and arg parsing

**Files:**
- Create: `src/cli/index.ts`

**What it does:** CLI entry point. Parses args, sets up pumped-fn scope, dispatches to commands, handles errors with cleanup.

**Implementation:**

```typescript
#!/usr/bin/env node
// src/cli/index.ts
import { createScope } from "@pumped-fn/lite";
import { findC3Dir } from "../core/config";
import { c3DirTag, optionsTag, type CliOptions } from "./context";
import { showHelp } from "./help";

// Command imports (each is a pumped-fn flow)
import { listCommand } from "./commands/list";
import { readCommand } from "./commands/read";
import { traceCommand } from "./commands/trace";
import { checkCommand } from "./commands/check";
import { impactCommand } from "./commands/impact";
import { initCommand } from "./commands/init";
import { addCommand } from "./commands/add";
import { evolveCommand } from "./commands/evolve";
import { templateCommand } from "./commands/template";
import { syncCommand } from "./commands/sync";
import { refCommand } from "./commands/ref";

const VERSION = "5.0.0";

class C3Error extends Error {
  constructor(message: string, public hint?: string) {
    super(message);
    this.name = "C3Error";
  }
}

function parseArgs(argv: string[]): CliOptions {
  const args: string[] = [];
  let json = false, flat = false, reverse = false, help = false, version = false;
  let depth: number | undefined, container: string | undefined, c3Dir: string | undefined;
  let feature = false, docs = false, code = false, list = false;
  let refs: string[] | undefined;

  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === "--json") json = true;
    else if (arg === "--flat") flat = true;
    else if (arg === "--reverse") reverse = true;
    else if (arg === "--feature") feature = true;
    else if (arg === "--docs") docs = true;
    else if (arg === "--code") code = true;
    else if (arg === "--list") list = true;
    else if (arg === "-h" || arg === "--help") help = true;
    else if (arg === "-v" || arg === "--version") version = true;
    else if (arg === "--depth" && argv[i + 1]) depth = parseInt(argv[++i], 10);
    else if (arg === "--container" && argv[i + 1]) container = argv[++i];
    else if (arg === "--c3-dir" && argv[i + 1]) c3Dir = argv[++i];
    else if (arg === "--refs" && argv[i + 1]) refs = argv[++i].split(",");
    else args.push(arg);
  }

  return {
    command: args[0] || "",
    args: args.slice(1),
    json, flat, reverse, depth, container, c3Dir,
    feature, docs, code, list, refs, help, version,
  };
}

const commandMap: Record<string, (options: CliOptions, c3Dir: string) => Promise<void>> = {
  list: listCommand,
  read: readCommand,
  trace: traceCommand,
  check: checkCommand,
  impact: impactCommand,
  add: addCommand,
  evolve: evolveCommand,
  template: templateCommand,
  sync: syncCommand,
  ref: refCommand,
};

// Commands that don't need .c3/
const NO_C3_REQUIRED = new Set(["init", "template"]);

async function main() {
  const options = parseArgs(process.argv.slice(2));

  if (options.version) {
    console.log(VERSION);
    return;
  }

  if (options.help || !options.command) {
    showHelp(options.command || undefined);
    return;
  }

  // init is special — creates .c3/
  if (options.command === "init") {
    const { initCommand } = await import("./commands/init");
    await initCommand(options, process.cwd());
    return;
  }

  // template doesn't need .c3/ dir
  if (options.command === "template") {
    const { templateCommand } = await import("./commands/template");
    await templateCommand(options, "");
    return;
  }

  const handler = commandMap[options.command];
  if (!handler) {
    console.error(`error: unknown command '${options.command}'`);
    console.error(`hint: run 'c3 --help' to see available commands`);
    process.exitCode = 1;
    return;
  }

  const c3Dir = options.c3Dir ?? findC3Dir(process.cwd());
  if (!c3Dir) {
    console.error("error: No .c3/ directory found");
    console.error("hint: run 'c3 init' to create one, or use --c3-dir <path>");
    process.exitCode = 1;
    return;
  }

  const scope = createScope({
    tags: [c3DirTag(c3Dir), optionsTag(options)],
  });

  try {
    const ctx = scope.createContext();
    await handler(options, c3Dir);
    await ctx.close({ ok: true });
  } catch (err) {
    if (err instanceof C3Error) {
      console.error(`error: ${err.message}`);
      if (err.hint) console.error(`hint: ${err.hint}`);
      process.exitCode = 1;
    } else {
      console.error(`unexpected error: ${err}`);
      console.error(`Report: https://github.com/lagz0ne/c3-skill/issues`);
      process.exitCode = 2;
    }
  } finally {
    await scope.dispose();
  }
}

export { C3Error };
main();
```

**Note:** The command functions receive `options` and `c3Dir` directly rather than resolving from scope. The scope provides atom caching — commands that need the graph resolve `graphAtom` from the scope. This keeps commands simple while still benefiting from cached graph resolution.

**Commit:**

```bash
git add src/cli/index.ts
git commit -m "feat(cli): add entry point with arg parsing and scope lifecycle"
```

---

## Task 10-15: Commands (list, read, trace, check, impact, init, add, evolve, template, sync, ref)

Each command is a separate file in `src/cli/commands/`. Each follows the same pattern:

1. Receive `options` and `c3Dir`
2. Build graph from `walkC3Docs` + `buildRelationshipGraph` (or use simpler operations)
3. Execute the command logic
4. Render output (plain text or JSON)

**Files to create:**
- `src/cli/commands/list.ts`
- `src/cli/commands/read.ts`
- `src/cli/commands/trace.ts`
- `src/cli/commands/check.ts`
- `src/cli/commands/impact.ts`
- `src/cli/commands/init.ts`
- `src/cli/commands/add.ts`
- `src/cli/commands/evolve.ts`
- `src/cli/commands/template.ts`
- `src/cli/commands/sync.ts`
- `src/cli/commands/ref.ts`

These are detailed in the design doc. Each agent implementing a command should:

1. Read the design doc section for that command
2. Use the core modules (walker, frontmatter, numbering, wiring)
3. Use the output module for rendering
4. Handle `--json` flag for machine-readable output
5. Handle `--help` flag by deferring to help.ts

**Template files** are read from `${projectRoot}/templates/` (sibling to `.c3/`). For the npm package distribution, templates will be bundled into the dist. For now, read from the repo's `templates/` directory.

**Each command gets its own commit.**

---

## Task 16: Build & Cleanup

**Files:**
- Modify: `scripts/build.ts` — add CLI build step
- Delete: `src/mcp-server.ts`
- Delete: `src/core/embedding.ts`
- Delete: `src/core/vector-index.ts`
- Delete: `mcp.json`
- Delete: `install.sh`
- Modify: `src/index.ts` — update exports
- Modify: `.github/workflows/release.yml` — replace binary builds with npm publish

**Step 1: Remove MCP files**

```bash
rm src/mcp-server.ts src/core/embedding.ts src/core/vector-index.ts mcp.json install.sh
```

**Step 2: Update src/index.ts**

```typescript
// src/index.ts — library exports for programmatic use
export { loadConfig, findC3Dir, type C3Config } from "./core/config";
export { parseFrontmatter, classifyDoc, deriveRelationships, type Frontmatter, type DocType } from "./core/frontmatter";
export { walkC3Docs, buildRelationshipGraph, type C3Entity, type C3Graph } from "./core/walker";
export { nextContainerId, nextComponentId, nextAdrId } from "./core/numbering";
```

**Step 3: Update release.yml**

Replace the multi-platform binary build with npm publish. Remove the binary jobs, add:

```yaml
jobs:
  check-version:
    # ... (keep existing version check)

  publish:
    needs: check-version
    if: needs.check-version.outputs.should_release == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install
      - run: bun run build
      - run: npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

**Step 4: Commit**

```bash
git add -A
git commit -m "chore: remove MCP server, update build for CLI-only distribution"
```

---

## Task 17: Smoke Test

**Files:**
- Modify: `test-smoke.ts`

**What it does:** Quick smoke test that exercises the CLI commands against a real `.c3/` directory.

```typescript
// test-smoke.ts
import { execSync } from "child_process";

const CLI = "bun run src/cli/index.ts";
const TEST_DIR = process.env.C3_TEST_DIR || `${process.env.HOME}/dev/acountee-v5`;

function run(cmd: string): string {
  return execSync(`${CLI} --c3-dir ${TEST_DIR}/.c3 ${cmd}`, { encoding: "utf-8" });
}

console.log("=== c3 --version ===");
console.log(execSync(`${CLI} --version`, { encoding: "utf-8" }));

console.log("=== c3 list ===");
console.log(run("list"));

console.log("=== c3 list --flat ===");
console.log(run("list --flat"));

console.log("=== c3 list --json ===");
const json = JSON.parse(run("list --json"));
console.log(`Entities: ${Object.keys(json).length}`);

console.log("=== c3 check ===");
console.log(run("check"));

console.log("=== c3 template --list ===");
console.log(execSync(`${CLI} template --list`, { encoding: "utf-8" }));

console.log("\nSmoke test passed!");
```

**Commit:**

```bash
git add test-smoke.ts
git commit -m "test: update smoke test for CLI commands"
```

---

## Team Assignment

For parallel execution with a team of agents:

| Agent | Tasks | Depends On |
|---|---|---|
| **core-agent** | Task 0 (setup), Task 1-5 (core modules) | — |
| **cli-agent** | Task 6-9 (CLI framework: context, help, output, entry) | Task 0 |
| **data-commands-agent** | list, read, trace, check, impact commands | Tasks 1-9 |
| **action-commands-agent** | init, add, evolve, template, sync commands | Tasks 1-9 |
| **ref-build-agent** | ref subcommands, Task 16 (build/cleanup), Task 17 (smoke test) | Tasks 1-9 |

**Execution order:**
1. core-agent + cli-agent run in parallel
2. data-commands-agent + action-commands-agent + ref-build-agent run in parallel (after 1 completes)
