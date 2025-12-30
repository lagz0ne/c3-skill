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
