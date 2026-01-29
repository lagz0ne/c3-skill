#!/usr/bin/env bun
/**
 * C3 Skills Build System
 *
 * Builds self-contained skills for multiple targets:
 * - claude-code: Claude Code plugin format
 * - opencode: OpenCode plugin format
 *
 * Each skill gets its references bundled alongside, making
 * skills work when installed directly (not via git clone).
 *
 * Usage:
 *   bun run scripts/build.ts              # Build all targets
 *   bun run scripts/build.ts --target=claude-code
 *   bun run scripts/build.ts --target=opencode
 */

import { $ } from "bun"
import { readdir, mkdir, readFile, writeFile, cp, rm } from "fs/promises"
import { join, dirname } from "path"
import { existsSync } from "fs"

const ROOT = import.meta.dir.replace("/scripts", "")

// ─────────────────────────────────────────────
// TARGET CONFIGURATION
// ─────────────────────────────────────────────

type Target = 'claude-code' | 'opencode' | 'codex';

interface TargetConfig {
  name: string;
  skillsDir: string;
  agentsDir: string | null;  // null = no agents support
  namespace: string;
  outputDir: string;
}

const TARGETS: Record<Target, TargetConfig> = {
  'claude-code': {
    name: 'claude-code',
    skillsDir: 'skills',
    agentsDir: 'agents',
    namespace: 'c3-skill',
    outputDir: 'dist/claude-code',
  },
  'opencode': {
    name: 'opencode',
    skillsDir: 'skill',
    agentsDir: '.opencode/agents',  // OpenCode expects agents here
    namespace: 'opencode-c3',
    outputDir: 'dist/opencode-c3',
  },
  'codex': {
    name: 'codex',
    skillsDir: 'skill',
    agentsDir: null,  // Codex doesn't support subagents
    namespace: 'codex-c3',
    outputDir: 'dist/codex-c3',
  },
};

// Parse CLI args
const args = process.argv.slice(2);
const targetArg = args.find(a => a.startsWith('--target='));
const selectedTarget = targetArg
  ? (targetArg.split('=')[1] as Target)
  : undefined;

// ─────────────────────────────────────────────
// REFERENCE EXTRACTION & BUNDLING
// ─────────────────────────────────────────────

// Reference patterns to match in skill content
const REFERENCE_PATTERNS = [
  /`\*\*\/references\/([^`]+\.md)`/g,
  /\*\*\/references\/([^\s\)]+\.md)/g,
];

function extractReferences(content: string): string[] {
  const refs = new Set<string>();
  for (const pattern of REFERENCE_PATTERNS) {
    const regex = new RegExp(pattern.source, pattern.flags);
    for (const match of content.matchAll(regex)) {
      refs.add(match[1]);
    }
  }
  return [...refs];
}

function rewriteReferences(content: string): string {
  let result = content;
  // Rewrite `**/references/X.md` to `references/X.md`
  result = result.replace(/`\*\*\/references\/([^`]+\.md)`/g, '`references/$1`');
  // Rewrite **/references/X.md (unquoted) to references/X.md
  result = result.replace(/\*\*\/references\/([^\s\)]+\.md)/g, 'references/$1');
  return result;
}

function rewriteNamespace(content: string, fromNs: string, toNs: string): string {
  if (fromNs === toNs) return content;
  // Only rewrite c3-skill: prefixes (subagent dispatch)
  return content.replace(new RegExp(`${fromNs}:`, 'g'), `${toNs}:`);
}

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

async function transformSkillsForTarget(config: TargetConfig, DIST: string): Promise<string[]> {
  const skillsDir = join(ROOT, "skills")
  const outDir = join(DIST, config.skillsDir)
  const transformed: string[] = []

  await mkdir(outDir, { recursive: true })

  const entries = await readdir(skillsDir, { withFileTypes: true })

  for (const entry of entries) {
    if (!entry.isDirectory()) continue

    const skillName = entry.name

    // Validate name
    if (!validateSkillName(skillName)) {
      console.warn(`  ⚠️  Skipping skill "${skillName}" - name doesn't match pattern`)
      continue
    }

    const srcPath = join(skillsDir, skillName, "SKILL.md")
    if (!existsSync(srcPath)) {
      console.warn(`  ⚠️  Skipping skill "${skillName}" - no SKILL.md found`)
      continue
    }

    const destDir = join(outDir, skillName)
    await mkdir(destDir, { recursive: true })

    // Read skill content
    let content = await readFile(srcPath, "utf-8")

    // Extract and bundle references
    const refs = extractReferences(content)
    if (refs.length > 0) {
      const refsDir = join(destDir, "references")
      await mkdir(refsDir, { recursive: true })

      for (const ref of refs) {
        const srcRef = join(ROOT, "references", ref)
        if (!existsSync(srcRef)) {
          console.warn(`  ⚠️  Skill "${skillName}" references missing file: ${ref}`)
          continue
        }
        // Handle nested paths
        const destRef = join(refsDir, ref)
        await mkdir(dirname(destRef), { recursive: true })
        await cp(srcRef, destRef)
      }
      console.log(`     → Bundled ${refs.length} reference(s)`)
    }

    // Rewrite reference paths in content
    content = rewriteReferences(content)

    // Rewrite namespace for non-Claude-Code targets
    if (config.namespace !== 'c3-skill') {
      content = rewriteNamespace(content, 'c3-skill', config.namespace)
    }

    // Write transformed skill
    await writeFile(join(destDir, "SKILL.md"), content)

    // Copy any additional files (except SKILL.md which we already wrote)
    const skillFiles = await readdir(join(skillsDir, skillName), { withFileTypes: true })
    for (const file of skillFiles) {
      if (file.name !== "SKILL.md") {
        await cp(
          join(skillsDir, skillName, file.name),
          join(destDir, file.name),
          { recursive: true }
        )
      }
    }

    transformed.push(skillName)
    console.log(`  ✓ Skill: ${skillName}`)
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

async function transformAgentsForTarget(config: TargetConfig, DIST: string): Promise<string[]> {
  const agentsDir = join(ROOT, "agents")
  const outDir = join(DIST, config.agentsDir)
  const transformed: string[] = []

  await mkdir(outDir, { recursive: true })

  const entries = await readdir(agentsDir)

  for (const file of entries) {
    if (!file.endsWith(".md")) continue

    const agentName = file.replace(".md", "")
    const srcPath = join(agentsDir, file)
    let content = await readFile(srcPath, "utf-8")

    // Rewrite namespace for sub-agent dispatch
    if (config.namespace !== 'c3-skill') {
      content = rewriteNamespace(content, 'c3-skill', config.namespace)
    }

    if (config.name === 'opencode') {
      // OpenCode needs different frontmatter format
      const { frontmatter, body } = parseFrontmatter(content)

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
        // Handle tools as array string like ["Read", "Glob"]
        const toolsMatch = frontmatter.tools.match(/\[([^\]]*)\]/)
        if (toolsMatch) {
          const toolsList = toolsMatch[1].split(",").map(t => t.trim().replace(/"/g, ''))
          const toolsObj: Record<string, boolean> = {}

          for (const tool of toolsList) {
            const mapped = TOOL_MAP[tool]
            if (mapped) {
              toolsObj[mapped] = true
            }
          }

          newFrontmatter.tools = toolsObj
        }
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
    } else {
      // Claude Code: keep as-is (just with namespace rewrite applied)
      await writeFile(join(outDir, file), content)
    }

    transformed.push(agentName)
    console.log(`  ✓ Agent: ${agentName}`)
  }

  return transformed
}

// ─────────────────────────────────────────────
// ASSET COPYING (templates, scripts, hooks, references)
// ─────────────────────────────────────────────

async function copyReferences(DIST: string): Promise<void> {
  const srcDir = join(ROOT, "references")
  const destDir = join(DIST, "references")

  if (!existsSync(srcDir)) {
    console.warn("  ⚠️  No references directory found")
    return
  }

  await cp(srcDir, destDir, { recursive: true })
  console.log("  ✓ References copied (global)")
}

async function copyTemplates(DIST: string): Promise<void> {
  const srcDir = join(ROOT, "templates")
  const destDir = join(DIST, "templates")

  if (!existsSync(srcDir)) return

  await cp(srcDir, destDir, { recursive: true })
  console.log("  ✓ Templates copied")
}

async function copyPluginScripts(DIST: string): Promise<void> {
  const srcDir = join(ROOT, "scripts")
  const destDir = join(DIST, "scripts")

  if (!existsSync(srcDir)) return

  // Only copy .sh files, not build scripts
  const entries = await readdir(srcDir)
  await mkdir(destDir, { recursive: true })
  for (const file of entries) {
    if (file.endsWith('.sh')) {
      await cp(join(srcDir, file), join(destDir, file))
    }
  }
  console.log("  ✓ Plugin scripts copied")
}

async function copyHooks(DIST: string): Promise<void> {
  const srcDir = join(ROOT, "hooks")
  const destDir = join(DIST, "hooks")

  if (!existsSync(srcDir)) return

  await cp(srcDir, destDir, { recursive: true })
  console.log("  ✓ Hooks copied")
}

async function copyCommands(DIST: string): Promise<void> {
  const srcDir = join(ROOT, "commands")
  const destDir = join(DIST, "commands")

  if (!existsSync(srcDir)) return

  await cp(srcDir, destDir, { recursive: true })
  console.log("  ✓ Commands copied")
}

// ─────────────────────────────────────────────
// PLUGIN COMPILATION (OpenCode)
// ─────────────────────────────────────────────

async function compilePlugin(DIST: string): Promise<void> {
  const srcPath = join(ROOT, "src/opencode/plugin.ts")

  if (!existsSync(srcPath)) {
    // Create minimal plugin.js if source doesn't exist
    await writeFile(join(DIST, "plugin.js"), "export default {};")
    console.log("  ✓ Plugin stub created")
    return
  }

  await $`bun build ${srcPath} --outfile=${join(DIST, "plugin.js")} --target=bun`
  console.log("  ✓ Plugin compiled")
}

// ─────────────────────────────────────────────
// MANIFEST GENERATION
// ─────────────────────────────────────────────

async function generateClaudePluginManifest(DIST: string): Promise<void> {
  const srcPath = join(ROOT, ".claude-plugin/plugin.json")
  const manifest = JSON.parse(await readFile(srcPath, "utf-8"))

  const destDir = join(DIST, ".claude-plugin")
  await mkdir(destDir, { recursive: true })
  await writeFile(join(destDir, "plugin.json"), JSON.stringify(manifest, null, 2))

  console.log("  ✓ .claude-plugin/plugin.json generated")
}

async function generatePackageJson(DIST: string, targetName: string): Promise<void> {
  const claudePluginPath = join(ROOT, ".claude-plugin/plugin.json")
  const claudePlugin = JSON.parse(await readFile(claudePluginPath, "utf-8"))

  const packageName = targetName === 'codex' ? 'codex-c3' : 'opencode-c3'
  const keywords = targetName === 'codex'
    ? ["codex", "skill", "c3", "architecture"]
    : ["opencode", "plugin", "c3", "architecture"]

  const pkg: Record<string, unknown> = {
    name: packageName,
    version: claudePlugin.version || "1.0.0",
    description: claudePlugin.description || "",
    main: "./plugin.js",
    type: "module",
    author: claudePlugin.author || {},
    license: claudePlugin.license || "MIT",
    keywords,
  }

  // Add peer dependencies only for OpenCode (not Codex)
  if (targetName === 'opencode') {
    pkg.peerDependencies = {
      "@opencode-ai/plugin": "*",
    }
  }

  await writeFile(join(DIST, "package.json"), JSON.stringify(pkg, null, 2))
  console.log("  ✓ package.json generated")
}

// ─────────────────────────────────────────────
// VERIFICATION
// ─────────────────────────────────────────────

async function verifyTarget(
  config: TargetConfig,
  skills: string[],
  agents: string[],
  DIST: string
): Promise<boolean> {
  const required: string[] = []

  if (config.name === 'claude-code') {
    required.push(
      ".claude-plugin/plugin.json",
      ...skills.map(s => `${config.skillsDir}/${s}/SKILL.md`),
      ...agents.map(a => `${config.agentsDir}/${a}.md`),
    )
  } else if (config.name === 'opencode') {
    required.push(
      "package.json",
      "plugin.js",
      "references",
      ...skills.map(s => `${config.skillsDir}/${s}/SKILL.md`),
    )
    // Add agents only if there are any
    if (config.agentsDir && agents.length > 0) {
      required.push(...agents.map(a => `${config.agentsDir}/${a}.md`))
    }
  } else if (config.name === 'codex') {
    // Codex: skills only, no agents
    required.push(
      "package.json",
      "plugin.js",
      "references",
      ...skills.map(s => `${config.skillsDir}/${s}/SKILL.md`),
    )
  }

  let allPresent = true

  for (const path of required) {
    const fullPath = join(DIST, path)
    if (!existsSync(fullPath)) {
      console.error(`  ❌ Missing: ${path}`)
      allPresent = false
    }
  }

  // Verify reference bundling for skills
  for (const skill of skills) {
    const skillPath = join(DIST, config.skillsDir, skill, "SKILL.md")
    const content = await readFile(skillPath, "utf-8")

    // Check no glob patterns remain
    if (content.includes("**/references/")) {
      console.error(`  ❌ Skill "${skill}" still has **/references/ patterns`)
      allPresent = false
    }
  }

  if (allPresent) {
    console.log("\n  ✅ Build verified")
  } else {
    console.error("\n  ❌ Build verification failed")
  }

  return allPresent
}

// ─────────────────────────────────────────────
// MAIN
// ─────────────────────────────────────────────

async function buildTarget(config: TargetConfig): Promise<boolean> {
  const DIST = join(ROOT, config.outputDir)

  console.log(`\n${'='.repeat(50)}`)
  console.log(`Building ${config.name}...`)
  console.log(`${'='.repeat(50)}\n`)

  // Clean dist for this target
  if (existsSync(DIST)) {
    await rm(DIST, { recursive: true })
  }
  await mkdir(DIST, { recursive: true })

  // Transform skills
  console.log("Skills:")
  const skills = await transformSkillsForTarget(config, DIST)

  // Transform agents (if target supports them)
  let agents: string[] = []
  if (config.agentsDir) {
    console.log("\nAgents:")
    agents = await transformAgentsForTarget(config, DIST)
  } else {
    console.log("\nAgents: (skipped - target doesn't support subagents)")
  }

  // Copy shared assets (templates, scripts, hooks, commands)
  console.log("\nAssets:")
  await copyTemplates(DIST)
  await copyPluginScripts(DIST)
  await copyHooks(DIST)
  await copyCommands(DIST)

  // Target-specific generation
  console.log("\nManifest:")
  if (config.name === 'opencode' || config.name === 'codex') {
    await copyReferences(DIST)
    await compilePlugin(DIST)
    await generatePackageJson(DIST, config.name)
  } else if (config.name === 'claude-code') {
    await generateClaudePluginManifest(DIST)
  }

  // Verify
  console.log("\nVerification:")
  const valid = await verifyTarget(config, skills, agents, DIST)

  console.log(`\n${config.name} output: ${DIST}`)

  return valid
}

async function main(): Promise<void> {
  console.log("C3 Skills Build System")
  console.log("=".repeat(50))

  let success = true

  if (selectedTarget) {
    // Build single target
    const config = TARGETS[selectedTarget]
    if (!config) {
      console.error(`Unknown target: ${selectedTarget}`)
      console.error(`Available: ${Object.keys(TARGETS).join(', ')}`)
      process.exit(1)
    }
    success = await buildTarget(config)
  } else {
    // Build all targets
    for (const target of Object.values(TARGETS)) {
      const targetSuccess = await buildTarget(target)
      success = success && targetSuccess
    }
  }

  console.log("\n" + "=".repeat(50))
  if (success) {
    console.log("✅ All builds completed successfully")
  } else {
    console.log("❌ Some builds failed")
    process.exit(1)
  }
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
