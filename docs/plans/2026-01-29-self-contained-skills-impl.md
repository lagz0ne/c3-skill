# Self-Contained Skills Build System Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend the existing build script to bundle references into each skill, making skills self-contained for direct installs.

**Architecture:** Extend `scripts/build-opencode.ts` to support multiple targets (claude-code, opencode). Add reference extraction and bundling. Create CI workflow to push built output to `dist` branch.

**Tech Stack:** Bun, TypeScript, GitHub Actions

---

## Task 1: Extend Build Script with Reference Bundling

**Files:**
- Modify: `scripts/build-opencode.ts` → rename to `scripts/build.ts`
- Modify: `package.json` (update script name)

**Step 1: Rename script and update package.json**

```bash
mv scripts/build-opencode.ts scripts/build.ts
```

Update `package.json`:
```json
{
  "scripts": {
    "build": "bun run scripts/build.ts",
    "build:opencode": "bun run scripts/build.ts --target=opencode"
  }
}
```

**Step 2: Add reference extraction function**

Add to `scripts/build.ts` after the parseFrontmatter function:

```typescript
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
```

**Step 3: Update transformSkills to bundle references**

Replace the transformSkills function:

```typescript
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
      console.warn(`⚠️  Skipping skill "${skillName}" - name doesn't match pattern`)
      continue
    }

    const srcPath = join(skillsDir, skillName, "SKILL.md")
    if (!existsSync(srcPath)) {
      console.warn(`⚠️  Skipping skill "${skillName}" - no SKILL.md found`)
      continue
    }

    const destDir = join(outDir, skillName)
    await mkdir(destDir, { recursive: true })

    // Read and transform skill content
    let content = await readFile(srcPath, "utf-8")

    // Extract and bundle references
    const refs = extractReferences(content)
    if (refs.length > 0) {
      const refsDir = join(destDir, "references")
      await mkdir(refsDir, { recursive: true })

      for (const ref of refs) {
        const srcRef = join(ROOT, "references", ref)
        if (!existsSync(srcRef)) {
          console.warn(`⚠️  Skill "${skillName}" references missing file: ${ref}`)
          continue
        }
        // Handle nested paths
        const destRef = join(refsDir, ref)
        await mkdir(dirname(destRef), { recursive: true })
        await cp(srcRef, destRef)
      }
      console.log(`   → Bundled ${refs.length} reference(s)`)
    }

    // Rewrite reference paths in content
    content = rewriteReferences(content)

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
    console.log(`✓ Skill: ${skillName}`)
  }

  return transformed
}
```

**Step 4: Run build and verify**

```bash
bun run scripts/build.ts
```

Expected: Build completes, skills in `dist/opencode-c3/skill/*/` have `references/` folders with bundled files.

**Step 5: Verify reference bundling worked**

```bash
ls dist/opencode-c3/skill/c3-query/references/
```

Expected: `skill-harness.md`, `layer-navigation.md` present

```bash
grep "references/" dist/opencode-c3/skill/c3-query/SKILL.md | head -3
```

Expected: Paths show `references/X.md` not `**/references/X.md`

**Step 6: Commit**

```bash
git add scripts/build.ts package.json
git commit -m "feat(build): add reference bundling to skills

- Extract references from skill content using patterns
- Bundle referenced files into skill's references/ folder
- Rewrite paths from **/references/X.md to references/X.md
- Handle nested reference paths with mkdir"
```

---

## Task 2: Add Multi-Target Support

**Files:**
- Modify: `scripts/build.ts`

**Step 1: Add target configuration**

Add after the constants section:

```typescript
type Target = 'claude-code' | 'opencode';

interface TargetConfig {
  name: string;
  skillsDir: string;
  agentsDir: string;
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
    agentsDir: 'agent',
    namespace: 'opencode-c3',
    outputDir: 'dist/opencode-c3',
  },
};

// Parse CLI args
const args = process.argv.slice(2);
const targetArg = args.find(a => a.startsWith('--target='));
const selectedTarget = targetArg
  ? (targetArg.split('=')[1] as Target)
  : undefined;
```

**Step 2: Add namespace rewriting function**

```typescript
function rewriteNamespace(content: string, fromNs: string, toNs: string): string {
  if (fromNs === toNs) return content;
  // Only rewrite c3-skill: prefixes (subagent dispatch)
  return content.replace(new RegExp(`${fromNs}:`, 'g'), `${toNs}:`);
}
```

**Step 3: Refactor main to support targets**

Replace the main function:

```typescript
async function buildTarget(config: TargetConfig): Promise<void> {
  const DIST = join(ROOT, config.outputDir);

  console.log(`\n${'='.repeat(50)}`);
  console.log(`Building ${config.name}...`);
  console.log(`${'='.repeat(50)}\n`);

  // Clean dist for this target
  if (existsSync(DIST)) {
    await $`rm -rf ${DIST}`;
  }
  await mkdir(DIST, { recursive: true });

  // Transform skills
  const skills = await transformSkillsForTarget(config, DIST);

  // Transform agents
  const agents = await transformAgentsForTarget(config, DIST);

  // Target-specific generation
  if (config.name === 'opencode') {
    await copyReferences(DIST);
    await compilePlugin(DIST);
    await generatePackageJson(DIST);
  } else if (config.name === 'claude-code') {
    await generateClaudePluginManifest(DIST);
  }

  // Verify
  console.log("\nVerifying build...\n");
  await verifyTarget(config, skills, agents, DIST);

  console.log(`\n✅ ${config.name} build complete: ${DIST}`);
}

async function main(): Promise<void> {
  console.log("C3 Skills Build System\n");

  if (selectedTarget) {
    // Build single target
    const config = TARGETS[selectedTarget];
    if (!config) {
      console.error(`Unknown target: ${selectedTarget}`);
      console.error(`Available: ${Object.keys(TARGETS).join(', ')}`);
      process.exit(1);
    }
    await buildTarget(config);
  } else {
    // Build all targets
    for (const target of Object.values(TARGETS)) {
      await buildTarget(target);
    }
  }
}
```

**Step 4: Add Claude Code manifest generation**

```typescript
async function generateClaudePluginManifest(DIST: string): Promise<void> {
  const srcPath = join(ROOT, ".claude-plugin/plugin.json");
  const manifest = JSON.parse(await readFile(srcPath, "utf-8"));

  const destDir = join(DIST, ".claude-plugin");
  await mkdir(destDir, { recursive: true });
  await writeFile(join(destDir, "plugin.json"), JSON.stringify(manifest, null, 2));

  console.log("✓ .claude-plugin/plugin.json generated");
}
```

**Step 5: Update transform functions to accept config**

Rename and update `transformSkills` → `transformSkillsForTarget(config, DIST)` and `transformAgents` → `transformAgentsForTarget(config, DIST)`. Pass config for namespace rewriting.

**Step 6: Test multi-target build**

```bash
# Build all
bun run scripts/build.ts

# Build specific target
bun run scripts/build.ts --target=claude-code
```

Expected: Both `dist/claude-code/` and `dist/opencode-c3/` created with correct structure.

**Step 7: Commit**

```bash
git add scripts/build.ts
git commit -m "feat(build): add multi-target support

- Support --target=claude-code or --target=opencode
- Default builds all targets
- Claude Code target generates .claude-plugin/plugin.json
- OpenCode target generates package.json and compiles plugin.js
- Each target has own output directory"
```

---

## Task 3: Create CI Workflow

**Files:**
- Create: `.github/workflows/build-dist.yml`

**Step 1: Create workflow file**

```yaml
name: Build Distribution

on:
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Bun
        uses: oven-sh/setup-bun@v2

      - name: Install dependencies
        run: bun install

      - name: Build all targets
        run: bun run build

      - name: Verify builds
        run: |
          echo "=== Claude Code ==="
          ls -la dist/claude-code/
          ls -la dist/claude-code/.claude-plugin/
          ls -la dist/claude-code/skills/

          echo "=== OpenCode ==="
          ls -la dist/opencode-c3/
          ls -la dist/opencode-c3/skill/

      - name: Configure git
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"

      - name: Push to dist branch
        run: |
          # Create orphan branch
          git checkout --orphan dist-temp
          git rm -rf . || true

          # Move Claude Code output to root (enable dotglob for .claude-plugin)
          shopt -s dotglob
          mv dist/claude-code/* .

          # Keep OpenCode as subdirectory
          mv dist/opencode-c3 opencode

          # Cleanup
          rm -rf dist node_modules

          # Add README for dist branch
          echo "# C3 Skills Distribution" > README.md
          echo "" >> README.md
          echo "This branch contains the built distribution of C3 skills." >> README.md
          echo "" >> README.md
          echo "**For Claude Code:** Install directly from this branch" >> README.md
          echo "" >> README.md
          echo "**For OpenCode:** Use the \`opencode/\` subdirectory" >> README.md

          # Commit and push
          git add -A
          git commit -m "Build from ${GITHUB_SHA::7} [skip ci]"
          git push origin dist-temp:dist --force
```

**Step 2: Create .github/workflows directory if needed**

```bash
mkdir -p .github/workflows
```

**Step 3: Commit workflow**

```bash
git add .github/workflows/build-dist.yml
git commit -m "ci: add workflow to build and push to dist branch

- Triggers on push to main
- Builds all targets (claude-code, opencode)
- Pushes built output to dist branch
- Includes dotglob for .claude-plugin directory
- Adds dist branch README"
```

---

## Task 4: Test Complete Flow Locally

**Files:**
- None (testing only)

**Step 1: Run full build**

```bash
bun run build
```

**Step 2: Verify Claude Code structure**

```bash
# Check plugin manifest exists
cat dist/claude-code/.claude-plugin/plugin.json

# Check skills have bundled references
ls dist/claude-code/skills/c3-query/
ls dist/claude-code/skills/c3-query/references/

# Check paths are rewritten
grep "references/" dist/claude-code/skills/c3-query/SKILL.md
```

Expected:
- plugin.json present
- skills/c3-query/references/ has skill-harness.md, layer-navigation.md
- No `**/references/` patterns remain

**Step 3: Verify OpenCode structure**

```bash
# Check package.json
cat dist/opencode-c3/package.json

# Check skills have bundled references
ls dist/opencode-c3/skill/c3-query/
ls dist/opencode-c3/skill/c3-query/references/
```

**Step 4: Simulate dist branch content**

```bash
# Preview what dist branch would look like
mkdir -p /tmp/dist-preview
shopt -s dotglob
cp -r dist/claude-code/* /tmp/dist-preview/
cp -r dist/opencode-c3 /tmp/dist-preview/opencode
ls -la /tmp/dist-preview/
ls -la /tmp/dist-preview/.claude-plugin/
```

Expected: Root has skills/, agents/, .claude-plugin/ plus opencode/ subdirectory.

---

## Task 5: Push and Verify CI

**Files:**
- None (deployment)

**Step 1: Push to main**

```bash
git push origin main
```

**Step 2: Monitor GitHub Actions**

Go to repository → Actions → Watch "Build Distribution" workflow.

Expected: Workflow completes successfully.

**Step 3: Verify dist branch**

```bash
git fetch origin dist
git log origin/dist -1
```

Expected: Recent commit with "Build from xxxxxx"

**Step 4: Check dist branch contents**

```bash
git show origin/dist:skills/c3-query/SKILL.md | head -20
git show origin/dist:skills/c3-query/references/skill-harness.md | head -5
```

Expected: Self-contained skills with bundled references.

---

## Verification Checklist

After all tasks complete:

- [ ] `bun run build` completes without errors
- [ ] `dist/claude-code/skills/*/references/` contains bundled files
- [ ] `dist/claude-code/skills/*/SKILL.md` has rewritten paths (no `**/`)
- [ ] `dist/claude-code/.claude-plugin/plugin.json` exists
- [ ] `dist/opencode-c3/skill/*/references/` contains bundled files
- [ ] `dist/opencode-c3/package.json` exists
- [ ] CI workflow runs on push to main
- [ ] CI pushes to dist branch
- [ ] dist branch has correct root structure
