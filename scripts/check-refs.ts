#!/usr/bin/env bun
/**
 * Checks that bundled reference/template files in each skill
 * are identical to the shared source-of-truth copies.
 *
 * Usage:
 *   bun run scripts/check-refs.ts          # Check only
 *   bun run scripts/check-refs.ts --fix    # Copy from shared to fix drifts
 */

import { readdir, readFile } from "fs/promises"
import { join } from "path"
import { existsSync } from "fs"
import { $ } from "bun"

const ROOT = import.meta.dir.replace("/scripts", "")
const SKILLS_DIR = join(ROOT, "skills")
const SHARED_REFS = join(ROOT, "references")
const SHARED_TEMPLATES = join(ROOT, "templates")

const fix = process.argv.includes("--fix")

interface Mismatch {
  skill: string
  file: string
  type: "references" | "templates"
  issue: "content-differs" | "missing-in-shared"
}

async function checkDir(
  skillName: string,
  subdir: "references" | "templates",
  sharedDir: string,
): Promise<Mismatch[]> {
  const bundledDir = join(SKILLS_DIR, skillName, subdir)
  if (!existsSync(bundledDir)) return []

  const files = await readdir(bundledDir)
  const mismatches: Mismatch[] = []

  for (const file of files) {
    const bundledPath = join(bundledDir, file)
    const sharedPath = join(sharedDir, file)

    if (!existsSync(sharedPath)) {
      mismatches.push({ skill: skillName, file, type: subdir, issue: "missing-in-shared" })
      continue
    }

    const bundled = await readFile(bundledPath, "utf-8")
    const shared = await readFile(sharedPath, "utf-8")

    if (bundled !== shared) {
      mismatches.push({ skill: skillName, file, type: subdir, issue: "content-differs" })

      if (fix) {
        await $`cp ${sharedPath} ${bundledPath}`
        console.log(`  Fixed: ${skillName}/${subdir}/${file}`)
      }
    }
  }

  return mismatches
}

async function main() {
  console.log(fix ? "Checking & fixing bundled references..." : "Checking bundled references...")
  console.log()

  const entries = await readdir(SKILLS_DIR, { withFileTypes: true })
  const allMismatches: Mismatch[] = []

  for (const entry of entries) {
    if (!entry.isDirectory()) continue
    const skillName = entry.name

    const refMismatches = await checkDir(skillName, "references", SHARED_REFS)
    const tplMismatches = await checkDir(skillName, "templates", SHARED_TEMPLATES)
    allMismatches.push(...refMismatches, ...tplMismatches)
  }

  if (allMismatches.length === 0) {
    console.log("All bundled files match their shared source. No drift detected.")
    return
  }

  if (!fix) {
    console.log(`Found ${allMismatches.length} mismatch(es):\n`)
    for (const m of allMismatches) {
      const icon = m.issue === "content-differs" ? "~" : "?"
      console.log(`  ${icon} skills/${m.skill}/${m.type}/${m.file} — ${m.issue}`)
    }
    console.log(`\nRun with --fix to copy from shared source.`)
    process.exit(1)
  } else {
    console.log(`\nFixed ${allMismatches.length} file(s).`)
  }
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
