#!/usr/bin/env bun
/**
 * C3 Skills Build System
 *
 * Builds self-contained Claude Code plugin from source.
 * Skills bundle their own references/ and templates/ subdirectories.
 *
 * Usage:
 *   bun run build
 */

import { readdir, mkdir, readFile, writeFile, cp, rm } from "fs/promises"
import { join } from "path"
import { existsSync } from "fs"

const ROOT = import.meta.dir.replace("/scripts", "")
const DIST = join(ROOT, "dist/claude-code")

// ─────────────────────────────────────────────
// UTILITIES
// ─────────────────────────────────────────────

function validateSkillName(name: string): boolean {
  return /^[a-z0-9]+(-[a-z0-9]+)*$/.test(name)
}

// ─────────────────────────────────────────────
// SKILLS
// ─────────────────────────────────────────────

async function buildSkills(): Promise<string[]> {
  const skillsDir = join(ROOT, "skills")
  const outDir = join(DIST, "skills")
  const built: string[] = []

  await mkdir(outDir, { recursive: true })

  const entries = await readdir(skillsDir, { withFileTypes: true })

  for (const entry of entries) {
    if (!entry.isDirectory()) continue

    const skillName = entry.name

    if (!validateSkillName(skillName)) {
      console.warn(`  ⚠️  Skipping skill "${skillName}" - name doesn't match pattern`)
      continue
    }

    const srcPath = join(skillsDir, skillName, "SKILL.md")
    if (!existsSync(srcPath)) {
      console.warn(`  ⚠️  Skipping skill "${skillName}" - no SKILL.md found`)
      continue
    }

    // Copy entire skill directory (SKILL.md + references/ + templates/)
    const destDir = join(outDir, skillName)
    await cp(join(skillsDir, skillName), destDir, { recursive: true })

    built.push(skillName)
    console.log(`  ✓ Skill: ${skillName}`)
  }

  return built
}

// ─────────────────────────────────────────────
// AGENTS
// ─────────────────────────────────────────────

async function buildAgents(): Promise<string[]> {
  const agentsDir = join(ROOT, "agents")
  const outDir = join(DIST, "agents")
  const built: string[] = []

  await mkdir(outDir, { recursive: true })

  const entries = await readdir(agentsDir)

  for (const file of entries) {
    if (!file.endsWith(".md")) continue

    await cp(join(agentsDir, file), join(outDir, file))

    const agentName = file.replace(".md", "")
    built.push(agentName)
    console.log(`  ✓ Agent: ${agentName}`)
  }

  return built
}

// ─────────────────────────────────────────────
// ASSETS
// ─────────────────────────────────────────────

async function copyTemplates(): Promise<void> {
  const srcDir = join(ROOT, "templates")
  if (!existsSync(srcDir)) return

  await cp(srcDir, join(DIST, "templates"), { recursive: true })
  console.log("  ✓ Templates copied")
}

// ─────────────────────────────────────────────
// MANIFEST
// ─────────────────────────────────────────────

async function generateManifest(): Promise<void> {
  const srcPath = join(ROOT, ".claude-plugin/plugin.json")
  const manifest = JSON.parse(await readFile(srcPath, "utf-8"))

  const destDir = join(DIST, ".claude-plugin")
  await mkdir(destDir, { recursive: true })
  await writeFile(join(destDir, "plugin.json"), JSON.stringify(manifest, null, 2))

  console.log("  ✓ .claude-plugin/plugin.json generated")
}

// ─────────────────────────────────────────────
// VERIFICATION
// ─────────────────────────────────────────────

async function verify(skills: string[], agents: string[]): Promise<boolean> {
  const required = [
    ".claude-plugin/plugin.json",
    ...skills.map(s => `skills/${s}/SKILL.md`),
    ...agents.map(a => `agents/${a}.md`),
  ]

  let allPresent = true

  for (const path of required) {
    if (!existsSync(join(DIST, path))) {
      console.error(`  ❌ Missing: ${path}`)
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

async function main(): Promise<void> {
  console.log("C3 Skills Build System")
  console.log("=".repeat(50))

  // Clean
  if (existsSync(DIST)) {
    await rm(DIST, { recursive: true })
  }
  await mkdir(DIST, { recursive: true })

  // Build
  console.log("\nSkills:")
  const skills = await buildSkills()

  console.log("\nAgents:")
  const agents = await buildAgents()

  console.log("\nAssets:")
  await copyTemplates()

  console.log("\nManifest:")
  await generateManifest()

  console.log("\nVerification:")
  const valid = await verify(skills, agents)

  console.log(`\nOutput: ${DIST}`)

  if (!valid) {
    process.exit(1)
  }

  console.log("\n" + "=".repeat(50))
  console.log("✅ Build completed successfully")
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
