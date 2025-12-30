# OpenCode Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add OpenCode compatibility to C3 plugin via Bun build script that transforms Claude Code format to OpenCode format and compiles TypeScript hooks.

**Architecture:** Claude Code remains source of truth. Build script transforms `skills/` → `dist/opencode-c3/skill/`, `agents/` → `dist/opencode-c3/agent/`, compiles `src/opencode/` → `dist/opencode-c3/plugin.js`, and generates `package.json` from `.claude-plugin/plugin.json`.

**Tech Stack:** Bun (build/runtime), TypeScript, @opencode-ai/plugin

---

## Task 1: Create package.json

**Files:**
- Create: `package.json`

**Step 1: Create root package.json**

```json
{
  "name": "c3-design",
  "version": "1.12.2",
  "private": true,
  "type": "module",
  "scripts": {
    "build:opencode": "bun run scripts/build-opencode.ts"
  },
  "devDependencies": {
    "bun-types": "latest",
    "@opencode-ai/plugin": "latest"
  }
}
```

**Step 2: Install dependencies**

Run: `bun install`
Expected: `bun.lockb` created, `node_modules/` populated

**Step 3: Commit**

```bash
git add package.json bun.lockb
git commit -m "chore: add package.json for OpenCode build"
```

---

## Task 2: Update .gitignore

**Files:**
- Modify: `.gitignore`

**Step 1: Append OpenCode-specific ignores**

Add to end of `.gitignore`:

```
# OpenCode build output
dist/
.opencode/
```

**Step 2: Commit**

```bash
git add .gitignore
git commit -m "chore: ignore OpenCode build outputs"
```

---

## Task 3: Create plugin hooks

**Files:**
- Create: `src/opencode/plugin.ts`

**Step 1: Create directory**

Run: `mkdir -p src/opencode`

**Step 2: Write plugin.ts**

```typescript
import type { Plugin } from "@opencode-ai/plugin"

export const C3Plugin: Plugin = async (ctx) => {
  const c3Path = `${ctx.worktree}/.c3`

  return {
    // ─────────────────────────────────────────────
    // TOOL GUARDS: Warn/block on sensitive C3 edits
    // ─────────────────────────────────────────────
    'tool.execute.before': async (input) => {
      const { tool, args } = input as { tool: string; args: Record<string, unknown> }

      // Warn on Context doc edits (high-impact)
      if (tool === 'edit' && args.file_path === `${c3Path}/README.md`) {
        console.warn("⚠️  Editing Context document - system-wide impact")
      }

      // Block deletion of C3 docs
      if (tool === 'bash' && typeof args.command === 'string') {
        if (/rm\s+(-rf?\s+)?.*\.c3/.test(args.command)) {
          throw new Error("🛑 Cannot delete C3 architecture documents")
        }
      }
    },

    // ─────────────────────────────────────────────
    // TOOL OBSERVERS: React after tool completion
    // ─────────────────────────────────────────────
    'tool.execute.after': async (input) => {
      const { tool, args } = input as { tool: string; args: Record<string, unknown> }

      // Log C3 doc modifications
      if (tool === 'write' && typeof args.file_path === 'string') {
        if (args.file_path.includes('.c3/')) {
          console.log(`📝 C3 doc written: ${args.file_path}`)
        }
      }
    },

    // ─────────────────────────────────────────────
    // FILE OBSERVER: Track architecture changes
    // ─────────────────────────────────────────────
    'file.edited': async ({ event }) => {
      const { path } = event as { path: string }

      // Track ADR changes
      if (path.includes('/adr-') && path.endsWith('.md')) {
        console.log(`📋 ADR modified: ${path}`)
      }

      // Track container changes
      if (/\.c3\/c3-\d+-/.test(path)) {
        console.log(`📦 Container doc modified: ${path}`)
      }
    },

    // ─────────────────────────────────────────────
    // SESSION INIT: Auto-detect C3 project
    // ─────────────────────────────────────────────
    'session.created': async () => {
      const file = Bun.file(`${c3Path}/README.md`)
      const hasC3 = await file.exists()
      if (hasC3) {
        console.log("🏗️  C3 architecture detected")
      }
    },

    // ─────────────────────────────────────────────
    // PERMISSION GATE: Protect critical operations
    // ─────────────────────────────────────────────
    'permission.ask': async (input) => {
      const { permission, context } = input as {
        permission: string
        context?: { path?: string }
      }

      // Auto-allow reads on C3 docs
      if (permission === 'read' && context?.path?.includes('.c3/')) {
        return { decision: 'allow' }
      }

      // Default: let user decide
      return { decision: 'ask' }
    },
  }
}

export default C3Plugin
```

**Step 3: Verify TypeScript compiles**

Run: `bun build src/opencode/plugin.ts --outdir=dist/test --target=bun`
Expected: No errors, `dist/test/plugin.js` created

**Step 4: Clean up test output**

Run: `rm -rf dist/test`

**Step 5: Commit**

```bash
git add src/opencode/plugin.ts
git commit -m "feat(opencode): add plugin hooks implementation"
```

---

## Task 4: Create build script

**Files:**
- Create: `scripts/build-opencode.ts`

**Step 1: Create directory**

Run: `mkdir -p scripts`

**Step 2: Write build script**

```typescript
#!/usr/bin/env bun
/**
 * Build script for OpenCode plugin
 *
 * Transforms Claude Code format to OpenCode format:
 * - skills/ → dist/opencode-c3/skill/
 * - agents/ → dist/opencode-c3/agent/
 * - src/opencode/ → dist/opencode-c3/plugin.js
 * - .claude-plugin/plugin.json → dist/opencode-c3/package.json
 */

import { $ } from "bun"
import { readdir, mkdir, readFile, writeFile, cp } from "fs/promises"
import { join, dirname } from "path"
import { existsSync } from "fs"

const ROOT = import.meta.dir.replace("/scripts", "")
const DIST = join(ROOT, "dist/opencode-c3")

// ─────────────────────────────────────────────
// UTILITIES
// ─────────────────────────────────────────────

function parseFrontmatter(content: string): { frontmatter: Record<string, string>; body: string } {
  const match = content.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/)
  if (!match) {
    return { frontmatter: {}, body: content }
  }

  const frontmatter: Record<string, string> = {}
  const lines = match[1].split("\n")

  for (const line of lines) {
    const colonIndex = line.indexOf(":")
    if (colonIndex > 0) {
      const key = line.slice(0, colonIndex).trim()
      let value = line.slice(colonIndex + 1).trim()
      // Handle multi-line values (simplified - just take first line)
      if (value.startsWith("|")) {
        value = ""
      }
      frontmatter[key] = value
    }
  }

  return { frontmatter, body: match[2] }
}

function validateSkillName(name: string): boolean {
  return /^[a-z0-9]+(-[a-z0-9]+)*$/.test(name)
}

// ─────────────────────────────────────────────
// SKILL TRANSFORMATION
// ─────────────────────────────────────────────

async function transformSkills(): Promise<string[]> {
  const skillsDir = join(ROOT, "skills")
  const outDir = join(DIST, "skill")
  const transformed: string[] = []

  await mkdir(outDir, { recursive: true })

  const entries = await readdir(skillsDir, { withFileTypes: true })

  for (const entry of entries) {
    if (!entry.isDirectory()) continue

    const skillName = entry.name

    // Validate name
    if (!validateSkillName(skillName)) {
      console.warn(`⚠️  Skipping skill "${skillName}" - name doesn't match OpenCode pattern`)
      continue
    }

    const srcPath = join(skillsDir, skillName, "SKILL.md")
    if (!existsSync(srcPath)) {
      console.warn(`⚠️  Skipping skill "${skillName}" - no SKILL.md found`)
      continue
    }

    const destDir = join(outDir, skillName)
    await mkdir(destDir, { recursive: true })

    // Copy SKILL.md as-is (frontmatter is already compatible)
    await cp(srcPath, join(destDir, "SKILL.md"))

    // Copy any additional files in the skill directory
    const skillFiles = await readdir(join(skillsDir, skillName))
    for (const file of skillFiles) {
      if (file !== "SKILL.md") {
        await cp(
          join(skillsDir, skillName, file),
          join(destDir, file),
          { recursive: true }
        )
      }
    }

    transformed.push(skillName)
    console.log(`✓ Skill: ${skillName}`)
  }

  return transformed
}

// ─────────────────────────────────────────────
// AGENT TRANSFORMATION
// ─────────────────────────────────────────────

const TOOL_MAP: Record<string, string> = {
  Glob: "glob",
  Grep: "grep",
  Read: "read",
  Edit: "edit",
  Write: "write",
  Bash: "bash",
  TodoWrite: "todowrite",
  TodoRead: "todoread",
  Skill: "skill",
  Task: "task",
  AskUserQuestion: "askuserquestion",
}

const MODEL_MAP: Record<string, string> = {
  opus: "anthropic/claude-opus-4-5",
  sonnet: "anthropic/claude-sonnet-4-5",
  haiku: "anthropic/claude-haiku-4-5",
}

async function transformAgents(): Promise<string[]> {
  const agentsDir = join(ROOT, "agents")
  const outDir = join(DIST, "agent")
  const transformed: string[] = []

  await mkdir(outDir, { recursive: true })

  const entries = await readdir(agentsDir)

  for (const file of entries) {
    if (!file.endsWith(".md")) continue

    const agentName = file.replace(".md", "")
    const srcPath = join(agentsDir, file)
    const content = await readFile(srcPath, "utf-8")

    const { frontmatter, body } = parseFrontmatter(content)

    // Transform frontmatter
    const newFrontmatter: Record<string, unknown> = {
      description: frontmatter.description || "",
      mode: "subagent",
    }

    // Transform model
    if (frontmatter.model) {
      newFrontmatter.model = MODEL_MAP[frontmatter.model] || frontmatter.model
    }

    // Transform tools list to object
    if (frontmatter.tools) {
      const toolsList = frontmatter.tools.split(",").map((t) => t.trim())
      const toolsObj: Record<string, boolean> = {}

      for (const tool of toolsList) {
        const mapped = TOOL_MAP[tool]
        if (mapped) {
          toolsObj[mapped] = true
        }
      }

      newFrontmatter.tools = toolsObj
    }

    // Build new content
    let newContent = "---\n"
    newContent += `description: ${newFrontmatter.description}\n`
    newContent += `mode: ${newFrontmatter.mode}\n`
    if (newFrontmatter.model) {
      newContent += `model: ${newFrontmatter.model}\n`
    }
    if (newFrontmatter.tools) {
      newContent += "tools:\n"
      for (const [tool, enabled] of Object.entries(newFrontmatter.tools as Record<string, boolean>)) {
        newContent += `  ${tool}: ${enabled}\n`
      }
    }
    newContent += "---\n"
    newContent += body

    await writeFile(join(outDir, file), newContent)

    transformed.push(agentName)
    console.log(`✓ Agent: ${agentName}`)
  }

  return transformed
}

// ─────────────────────────────────────────────
// PLUGIN COMPILATION
// ─────────────────────────────────────────────

async function compilePlugin(): Promise<void> {
  const srcPath = join(ROOT, "src/opencode/plugin.ts")

  if (!existsSync(srcPath)) {
    console.warn("⚠️  No plugin source found at src/opencode/plugin.ts")
    return
  }

  await $`bun build ${srcPath} --outfile=${join(DIST, "plugin.js")} --target=bun`

  console.log("✓ Plugin compiled")
}

// ─────────────────────────────────────────────
// PACKAGE.JSON GENERATION
// ─────────────────────────────────────────────

async function generatePackageJson(): Promise<void> {
  const claudePluginPath = join(ROOT, ".claude-plugin/plugin.json")
  const claudePlugin = JSON.parse(await readFile(claudePluginPath, "utf-8"))

  const pkg = {
    name: "opencode-c3",
    version: claudePlugin.version || "1.0.0",
    description: claudePlugin.description || "",
    main: "./plugin.js",
    type: "module",
    author: claudePlugin.author || {},
    license: claudePlugin.license || "MIT",
    keywords: ["opencode", "plugin", "c3", "architecture"],
    peerDependencies: {
      "@opencode-ai/plugin": "*",
    },
  }

  await writeFile(join(DIST, "package.json"), JSON.stringify(pkg, null, 2))

  console.log("✓ package.json generated")
}

// ─────────────────────────────────────────────
// VERIFICATION
// ─────────────────────────────────────────────

async function verify(skills: string[], agents: string[]): Promise<boolean> {
  const required = [
    "package.json",
    "plugin.js",
    ...skills.map((s) => `skill/${s}/SKILL.md`),
    ...agents.map((a) => `agent/${a}.md`),
  ]

  let allPresent = true

  for (const path of required) {
    const fullPath = join(DIST, path)
    if (!existsSync(fullPath)) {
      console.error(`❌ Missing: ${path}`)
      allPresent = false
    }
  }

  if (allPresent) {
    console.log("\n✅ Build verified")
  } else {
    console.error("\n❌ Build verification failed")
  }

  return allPresent
}

// ─────────────────────────────────────────────
// MAIN
// ─────────────────────────────────────────────

async function main(): Promise<void> {
  console.log("Building OpenCode plugin...\n")

  // Clean dist
  if (existsSync(DIST)) {
    await $`rm -rf ${DIST}`
  }
  await mkdir(DIST, { recursive: true })

  // Transform
  const skills = await transformSkills()
  const agents = await transformAgents()

  // Compile
  await compilePlugin()

  // Generate package.json
  await generatePackageJson()

  // Verify
  console.log("\nVerifying build...\n")
  const valid = await verify(skills, agents)

  if (!valid) {
    process.exit(1)
  }

  console.log(`\nOutput: ${DIST}`)
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
```

**Step 3: Make executable**

Run: `chmod +x scripts/build-opencode.ts`

**Step 4: Test build script**

Run: `bun run build:opencode`
Expected output:
```
Building OpenCode plugin...

✓ Skill: c3-implementation
✓ Skill: c3-structure
✓ Agent: c3
✓ Plugin compiled
✓ package.json generated

Verifying build...

✅ Build verified

Output: /path/to/c3-design/dist/opencode-c3
```

**Step 5: Verify output structure**

Run: `find dist/opencode-c3 -type f | sort`
Expected:
```
dist/opencode-c3/agent/c3.md
dist/opencode-c3/package.json
dist/opencode-c3/plugin.js
dist/opencode-c3/skill/c3-implementation/SKILL.md
dist/opencode-c3/skill/c3-structure/SKILL.md
```

**Step 6: Clean up dist**

Run: `rm -rf dist`

**Step 7: Commit**

```bash
git add scripts/build-opencode.ts
git commit -m "feat(opencode): add build script for OpenCode transformation"
```

---

## Task 5: Create CI workflow

**Files:**
- Create: `.github/workflows/publish-opencode.yml`

**Step 1: Create directory**

Run: `mkdir -p .github/workflows`

**Step 2: Write workflow file**

```yaml
name: Publish OpenCode Plugin

on:
  release:
    types: [published]
  workflow_dispatch:

jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest

      - name: Install dependencies
        run: bun install

      - name: Build OpenCode plugin
        run: bun run build:opencode

      - name: Setup npm auth
        run: |
          cd dist/opencode-c3
          echo "//registry.npmjs.org/:_authToken=\${NODE_AUTH_TOKEN}" > .npmrc

      - name: Publish to npm
        run: |
          cd dist/opencode-c3
          npm publish --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

**Step 3: Commit**

```bash
git add .github/workflows/publish-opencode.yml
git commit -m "ci: add OpenCode plugin publish workflow"
```

---

## Task 6: Final verification

**Step 1: Run full build**

Run: `bun run build:opencode`
Expected: Build completes successfully

**Step 2: Inspect agent transformation**

Run: `cat dist/opencode-c3/agent/c3.md | head -20`
Expected: Transformed frontmatter with `mode: subagent`, `tools:` as object

**Step 3: Inspect skill copy**

Run: `diff skills/c3-structure/SKILL.md dist/opencode-c3/skill/c3-structure/SKILL.md`
Expected: No diff (skill copied as-is)

**Step 4: Inspect package.json**

Run: `cat dist/opencode-c3/package.json`
Expected: Valid JSON with `name: "opencode-c3"`, `main: "./plugin.js"`

**Step 5: Test plugin loads (optional)**

Run: `bun dist/opencode-c3/plugin.js`
Expected: No errors (module exports properly)

**Step 6: Clean up**

Run: `rm -rf dist`

**Step 7: Final commit (if any uncommitted changes)**

```bash
git status
# If clean: done
# If changes: git add -A && git commit -m "chore: finalize OpenCode support"
```

---

## Summary

| Task | Files | Purpose |
|------|-------|---------|
| 1 | `package.json` | Root package with Bun deps |
| 2 | `.gitignore` | Ignore dist/, .opencode/ |
| 3 | `src/opencode/plugin.ts` | Hook implementations |
| 4 | `scripts/build-opencode.ts` | Build + transform + verify |
| 5 | `.github/workflows/publish-opencode.yml` | CI for npm publish |
| 6 | - | Final verification |

**Total: 6 tasks, ~20 steps**
